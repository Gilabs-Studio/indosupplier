package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// RecruitmentApplicantUsecase defines the interface for recruitment applicant business logic
type RecruitmentApplicantUsecase interface {
	// GetAll retrieves all applicants with pagination and filters
	GetAll(ctx context.Context, params dto.ListApplicantsParams) ([]*dto.RecruitmentApplicantResponse, *response.PaginationMeta, error)

	// GetByID retrieves an applicant by ID with enriched data
	GetByID(ctx context.Context, id string) (*dto.RecruitmentApplicantResponse, error)

	// Create creates a new applicant
	Create(ctx context.Context, req *dto.CreateRecruitmentApplicantDTO, userID string) (*dto.RecruitmentApplicantResponse, error)

	// Update updates an existing applicant
	Update(ctx context.Context, id string, req *dto.UpdateRecruitmentApplicantDTO, userID string) (*dto.RecruitmentApplicantResponse, error)

	// Delete performs soft delete on an applicant
	Delete(ctx context.Context, id string) error

	// MoveStage moves an applicant to a different stage
	MoveStage(ctx context.Context, id string, req *dto.MoveApplicantStageDTO, userID string) (*dto.RecruitmentApplicantResponse, error)

	// GetByStage retrieves applicants grouped by stage (for Kanban board)
	GetByStage(ctx context.Context, params dto.ListApplicantsByStageParams) (map[string]*dto.ApplicantsByStageResponse, error)

	// GetStages retrieves all active stages
	GetStages(ctx context.Context) ([]*dto.ApplicantStageResponse, error)

	// GetActivities retrieves activity history for an applicant
	GetActivities(ctx context.Context, applicantID string, page, perPage int) ([]*dto.ApplicantActivityResponse, *response.PaginationMeta, error)

	// AddActivity adds an activity to an applicant's history
	AddActivity(ctx context.Context, applicantID string, req *dto.CreateApplicantActivityDTO, userID string) (*dto.ApplicantActivityResponse, error)

	// GetByRecruitmentRequest retrieves applicants for a specific recruitment request
	GetByRecruitmentRequest(ctx context.Context, recruitmentRequestID string, page, perPage int) ([]*dto.RecruitmentApplicantResponse, *response.PaginationMeta, error)

	// ConvertToEmployee converts an applicant to an employee
	ConvertToEmployee(ctx context.Context, applicantID string, req *dto.ConvertApplicantToEmployeeDTO, userID string) (*dto.RecruitmentApplicantResponse, error)

	// CanConvertToEmployee checks if an applicant can be converted to an employee
	CanConvertToEmployee(ctx context.Context, applicantID string) (bool, string, error)
}
