package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// RecruitmentRequestUsecase defines the interface for recruitment request business logic
type RecruitmentRequestUsecase interface {
	// GetAll retrieves all recruitment requests with pagination and filters
	GetAll(ctx context.Context, page, perPage int, search string, status *string, divisionID, positionID *string, priority *string) ([]*dto.RecruitmentRequestResponse, *response.PaginationMeta, error)

	// GetByID retrieves a recruitment request by ID with enriched data
	GetByID(ctx context.Context, id string) (*dto.RecruitmentRequestResponse, error)

	// Create creates a new recruitment request
	Create(ctx context.Context, req *dto.CreateRecruitmentRequestDTO, userID string) (*dto.RecruitmentRequestResponse, error)

	// Update updates an existing recruitment request (DRAFT only)
	Update(ctx context.Context, id string, req *dto.UpdateRecruitmentRequestDTO, userID string) (*dto.RecruitmentRequestResponse, error)

	// Delete performs soft delete on a recruitment request (DRAFT only)
	Delete(ctx context.Context, id string) error

	// UpdateStatus transitions the recruitment request status
	UpdateStatus(ctx context.Context, id string, req *dto.UpdateRecruitmentStatusDTO, userID string) (*dto.RecruitmentRequestResponse, error)

	// UpdateFilledCount updates the number of filled positions
	UpdateFilledCount(ctx context.Context, id string, req *dto.UpdateFilledCountDTO) (*dto.RecruitmentRequestResponse, error)

	// GetFormData returns dropdown data for the recruitment form
	GetFormData(ctx context.Context) (*dto.RecruitmentFormDataResponse, error)
}
