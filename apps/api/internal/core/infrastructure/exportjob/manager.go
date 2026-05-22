package exportjob

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/storage"
	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

var (
	ErrJobNotFound  = errors.New("export job not found")
	ErrJobForbidden = errors.New("forbidden")
	ErrJobNotReady  = errors.New("export job not completed")
)

type GeneratedFile struct {
	FileName    string
	ContentType string
	Bytes       []byte
}

type Generator func(ctx context.Context) (*GeneratedFile, error)

type Job struct {
	ID           string     `json:"id"`
	Status       Status     `json:"status"`
	Progress     int        `json:"progress"`
	FileName     string     `json:"file_name,omitempty"`
	ContentType  string     `json:"content_type,omitempty"`
	FileURL      string     `json:"file_url,omitempty"`
	DownloadPath string     `json:"download_path,omitempty"`
	Error        string     `json:"error,omitempty"`
	CreatedBy    string     `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`

	filePath string
}

func (j *Job) FilePath() string {
	if j == nil {
		return ""
	}
	return j.filePath
}

type Manager struct {
	mu          sync.RWMutex
	jobs        map[string]*Job
	cleanupTTL  time.Duration
	cleanupTick time.Duration
}

func NewManager() *Manager {
	manager := &Manager{
		jobs:        make(map[string]*Job),
		cleanupTTL:  7 * 24 * time.Hour,
		cleanupTick: time.Hour,
	}
	go manager.startCleanupWorker()
	return manager
}

var DefaultManager = NewManager()

func (m *Manager) Enqueue(createdBy string, generator Generator) *Job {
	now := apptime.Now()
	job := &Job{
		ID:        uuid.NewString(),
		Status:    StatusQueued,
		Progress:  0,
		CreatedBy: createdBy,
		CreatedAt: now,
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	go m.process(job.ID, generator)

	return cloneJob(job)
}

func (m *Manager) process(jobID string, generator Generator) {
	startedAt := apptime.Now()

	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return
	}
	job.Status = StatusProcessing
	job.Progress = 5
	job.StartedAt = &startedAt
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	file, err := generator(ctx)
	finishedAt := apptime.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok = m.jobs[jobID]
	if !ok {
		return
	}

	if err != nil {
		job.Status = StatusFailed
		job.Progress = 100
		job.Error = err.Error()
		job.FinishedAt = &finishedAt
		return
	}

	if file == nil || len(file.Bytes) == 0 {
		job.Status = StatusFailed
		job.Progress = 100
		job.Error = "empty export file"
		job.FinishedAt = &finishedAt
		return
	}

	path, fileURL, err := persistGeneratedFile(job.ID, file)
	if err != nil {
		job.Status = StatusFailed
		job.Progress = 100
		job.Error = err.Error()
		job.FinishedAt = &finishedAt
		return
	}

	job.Status = StatusCompleted
	job.Progress = 100
	job.FileName = file.FileName
	job.ContentType = file.ContentType
	job.filePath = path
	job.FileURL = fileURL
	job.DownloadPath = fmt.Sprintf("/api/v1/exports/jobs/%s/download", job.ID)
	job.FinishedAt = &finishedAt
}

func (m *Manager) Get(id, userID string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	if job.CreatedBy != "" && job.CreatedBy != userID {
		return nil, ErrJobForbidden
	}

	return cloneJob(job), nil
}

func (m *Manager) ResolveDownload(id, userID string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	if job.CreatedBy != "" && job.CreatedBy != userID {
		return nil, ErrJobForbidden
	}
	if job.Status != StatusCompleted || job.filePath == "" {
		return nil, ErrJobNotReady
	}

	return cloneJob(job), nil
}

func (m *Manager) SetProgress(jobID string, progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 99 {
		progress = 99
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return
	}
	if job.Status == StatusCompleted || job.Status == StatusFailed {
		return
	}

	job.Status = StatusProcessing
	if progress > job.Progress {
		job.Progress = progress
	}
}

func (m *Manager) startCleanupWorker() {
	ticker := time.NewTicker(m.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanupExpiredJobs()
	}
}

func (m *Manager) cleanupExpiredJobs() {
	cutoff := apptime.Now().Add(-m.cleanupTTL)
	m.cleanupExpiredJobRecords(cutoff)
}

func (m *Manager) cleanupExpiredJobRecords(cutoff time.Time) {

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, job := range m.jobs {
		finishedAt := job.FinishedAt
		if finishedAt == nil {
			continue
		}
		if finishedAt.After(cutoff) {
			continue
		}

		if job.filePath != "" {
			_ = storage.Delete(context.Background(), job.filePath)
		}
		delete(m.jobs, id)
	}
}

func persistGeneratedFile(jobID string, file *GeneratedFile) (string, string, error) {
	fileName := sanitizeFileName(file.FileName)
	key := fmt.Sprintf("exports/%s_%s", jobID, fileName)

	url, err := storage.Upload(context.Background(), key, file.Bytes, file.ContentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload export file: %w", err)
	}

	return key, url, nil
}

var fileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func sanitizeFileName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "export.bin"
	}
	cleaned := fileNameSanitizer.ReplaceAllString(trimmed, "_")
	if cleaned == "" {
		return "export.bin"
	}
	return cleaned
}

func cloneJob(job *Job) *Job {
	if job == nil {
		return nil
	}
	copy := *job
	return &copy
}
