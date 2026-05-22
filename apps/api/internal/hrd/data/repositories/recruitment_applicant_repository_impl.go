package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
)

// recruitmentApplicantRepositoryImpl implements RecruitmentApplicantRepository
type recruitmentApplicantRepositoryImpl struct {
	db *gorm.DB
}

// NewRecruitmentApplicantRepository creates a new instance of RecruitmentApplicantRepository
func NewRecruitmentApplicantRepository(db *gorm.DB) RecruitmentApplicantRepository {
	return &recruitmentApplicantRepositoryImpl{db: db}
}

func (r *recruitmentApplicantRepositoryImpl) Create(ctx context.Context, applicant *models.RecruitmentApplicant) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Create(applicant).Error; err != nil {
		return fmt.Errorf("failed to create applicant: %w", err)
	}
	return nil
}

func (r *recruitmentApplicantRepositoryImpl) FindByID(ctx context.Context, id string) (*models.RecruitmentApplicant, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var applicant models.RecruitmentApplicant
	if err := database.GetDB(ctx, r.db).Preload("Stage").Preload("Employee").Where("id = ?", id).First(&applicant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find applicant: %w", err)
	}
	return &applicant, nil
}

func (r *recruitmentApplicantRepositoryImpl) Update(ctx context.Context, applicant *models.RecruitmentApplicant) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Save(applicant).Error; err != nil {
		return fmt.Errorf("failed to update applicant: %w", err)
	}
	return nil
}

func (r *recruitmentApplicantRepositoryImpl) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.RecruitmentApplicant{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete applicant: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("applicant not found")
	}
	return nil
}

func (r *recruitmentApplicantRepositoryImpl) FindAll(ctx context.Context, page, perPage int, search, recruitmentRequestID, stageID, source string) ([]models.RecruitmentApplicant, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var applicants []models.RecruitmentApplicant
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.RecruitmentApplicant{})

	// Apply filters
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("full_name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	if recruitmentRequestID != "" {
		query = query.Where("recruitment_request_id = ?", recruitmentRequestID)
	}

	if stageID != "" {
		query = query.Where("stage_id = ?", stageID)
	}

	if source != "" {
		query = query.Where("source = ?", source)
	}

	// Count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count applicants: %w", err)
	}

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * perPage
	query = query.Preload("Stage").
		Offset(offset).Limit(perPage).
		Order("last_activity_at DESC")

	if err := query.Find(&applicants).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find applicants: %w", err)
	}

	return applicants, total, nil
}

func (r *recruitmentApplicantRepositoryImpl) FindByRecruitmentRequest(ctx context.Context, recruitmentRequestID string, page, perPage int) ([]models.RecruitmentApplicant, int64, error) {
	return r.FindAll(ctx, page, perPage, "", recruitmentRequestID, "", "")
}

func (r *recruitmentApplicantRepositoryImpl) FindByStage(ctx context.Context, stageID, recruitmentRequestID string, page, perPage int) ([]models.RecruitmentApplicant, int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var applicants []models.RecruitmentApplicant
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.RecruitmentApplicant{}).
		Where("stage_id = ?", stageID)

	if recruitmentRequestID != "" {
		query = query.Where("recruitment_request_id = ?", recruitmentRequestID)
	}

	// Count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count applicants: %w", err)
	}

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * perPage

	query = query.Preload("Stage").
		Offset(offset).Limit(perPage).
		Order("last_activity_at DESC")

	if err := query.Find(&applicants).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find applicants: %w", err)
	}

	return applicants, total, nil
}

func (r *recruitmentApplicantRepositoryImpl) CountByStage(ctx context.Context, stageID, recruitmentRequestID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var count int64
	query := database.GetDB(ctx, r.db).Model(&models.RecruitmentApplicant{}).
		Where("stage_id = ?", stageID)

	if recruitmentRequestID != "" {
		query = query.Where("recruitment_request_id = ?", recruitmentRequestID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count applicants: %w", err)
	}

	return count, nil
}

func (r *recruitmentApplicantRepositoryImpl) MoveStage(ctx context.Context, applicantID, newStageID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Model(&models.RecruitmentApplicant{}).
		Where("id = ?", applicantID).
		Update("stage_id", newStageID)

	if result.Error != nil {
		return fmt.Errorf("failed to move applicant stage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("applicant not found")
	}

	return nil
}

// applicantStageRepositoryImpl implements ApplicantStageRepository
type applicantStageRepositoryImpl struct {
	db *gorm.DB
}

// NewApplicantStageRepository creates a new instance of ApplicantStageRepository
func NewApplicantStageRepository(db *gorm.DB) ApplicantStageRepository {
	return &applicantStageRepositoryImpl{db: db}
}

func (r *applicantStageRepositoryImpl) Create(ctx context.Context, stage *models.ApplicantStage) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Create(stage).Error; err != nil {
		return fmt.Errorf("failed to create stage: %w", err)
	}
	return nil
}

func (r *applicantStageRepositoryImpl) FindByID(ctx context.Context, id string) (*models.ApplicantStage, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stage models.ApplicantStage
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&stage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find stage: %w", err)
	}
	return &stage, nil
}

func (r *applicantStageRepositoryImpl) FindAll(ctx context.Context) ([]models.ApplicantStage, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stages []models.ApplicantStage
	if err := database.GetDB(ctx, r.db).Order("\"order\"").Find(&stages).Error; err != nil {
		return nil, fmt.Errorf("failed to find stages: %w", err)
	}
	return stages, nil
}

func (r *applicantStageRepositoryImpl) FindAllActive(ctx context.Context) ([]models.ApplicantStage, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stages []models.ApplicantStage
	if err := database.GetDB(ctx, r.db).
		Where("is_active = ?", true).
		Order("\"order\"").
		Find(&stages).Error; err != nil {
		return nil, fmt.Errorf("failed to find stages: %w", err)
	}
	return stages, nil
}

func (r *applicantStageRepositoryImpl) Update(ctx context.Context, stage *models.ApplicantStage) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Save(stage).Error; err != nil {
		return fmt.Errorf("failed to update stage: %w", err)
	}
	return nil
}

func (r *applicantStageRepositoryImpl) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if stage has applicants
	var count int64
	if err := database.GetDB(ctx, r.db).Model(&models.RecruitmentApplicant{}).
		Where("stage_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check stage usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete stage with existing applicants")
	}

	result := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.ApplicantStage{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete stage: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("stage not found")
	}
	return nil
}

func (r *applicantStageRepositoryImpl) SeedDefaultStages(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if stages already exist
	var count int64
	if err := database.GetDB(ctx, r.db).Model(&models.ApplicantStage{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing stages: %w", err)
	}

	if count > 0 {
		return nil // Stages already seeded
	}

	// Create default stages
	stages := models.DefaultApplicantStages()
	for i := range stages {
		if err := database.GetDB(ctx, r.db).Create(&stages[i]).Error; err != nil {
			return fmt.Errorf("failed to create default stage: %w", err)
		}
	}

	return nil
}

// applicantActivityRepositoryImpl implements ApplicantActivityRepository
type applicantActivityRepositoryImpl struct {
	db *gorm.DB
}

// NewApplicantActivityRepository creates a new instance of ApplicantActivityRepository
func NewApplicantActivityRepository(db *gorm.DB) ApplicantActivityRepository {
	return &applicantActivityRepositoryImpl{db: db}
}

func (r *applicantActivityRepositoryImpl) Create(ctx context.Context, activity *models.ApplicantActivity) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.GetDB(ctx, r.db).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

func (r *applicantActivityRepositoryImpl) FindByID(ctx context.Context, id string) (*models.ApplicantActivity, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var activity models.ApplicantActivity
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find activity: %w", err)
	}
	return &activity, nil
}

func (r *applicantActivityRepositoryImpl) FindByApplicant(ctx context.Context, applicantID string, page, perPage int) ([]models.ApplicantActivity, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var activities []models.ApplicantActivity
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ApplicantActivity{}).
		Where("applicant_id = ?", applicantID)

	// Count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage).
		Order("created_at DESC")

	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find activities: %w", err)
	}

	return activities, total, nil
}

func (r *applicantActivityRepositoryImpl) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.ApplicantActivity{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete activity: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("activity not found")
	}
	return nil
}
