package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
)

// RecruitmentApplicantRepository defines the interface for recruitment applicant data operations
type RecruitmentApplicantRepository interface {
	// Basic CRUD
	Create(ctx context.Context, applicant *models.RecruitmentApplicant) error
	FindByID(ctx context.Context, id string) (*models.RecruitmentApplicant, error)
	Update(ctx context.Context, applicant *models.RecruitmentApplicant) error
	Delete(ctx context.Context, id string) error

	// List with filters and pagination
	FindAll(ctx context.Context, page, perPage int, search, recruitmentRequestID, stageID, source string) ([]models.RecruitmentApplicant, int64, error)

	// Find by recruitment request
	FindByRecruitmentRequest(ctx context.Context, recruitmentRequestID string, page, perPage int) ([]models.RecruitmentApplicant, int64, error)

	// Find by stage (for Kanban board)
	FindByStage(ctx context.Context, stageID, recruitmentRequestID string, page, perPage int) ([]models.RecruitmentApplicant, int64, error)

	// Count by stage
	CountByStage(ctx context.Context, stageID, recruitmentRequestID string) (int64, error)

	// Move stage
	MoveStage(ctx context.Context, applicantID, newStageID string) error
}

// ApplicantStageRepository defines the interface for applicant stage operations
type ApplicantStageRepository interface {
	// CRUD
	Create(ctx context.Context, stage *models.ApplicantStage) error
	FindByID(ctx context.Context, id string) (*models.ApplicantStage, error)
	FindAll(ctx context.Context) ([]models.ApplicantStage, error)
	FindAllActive(ctx context.Context) ([]models.ApplicantStage, error)
	Update(ctx context.Context, stage *models.ApplicantStage) error
	Delete(ctx context.Context, id string) error

	// Seed default stages
	SeedDefaultStages(ctx context.Context) error
}

// ApplicantActivityRepository defines the interface for applicant activity operations
type ApplicantActivityRepository interface {
	// CRUD
	Create(ctx context.Context, activity *models.ApplicantActivity) error
	FindByID(ctx context.Context, id string) (*models.ApplicantActivity, error)
	FindByApplicant(ctx context.Context, applicantID string, page, perPage int) ([]models.ApplicantActivity, int64, error)
	Delete(ctx context.Context, id string) error
}
