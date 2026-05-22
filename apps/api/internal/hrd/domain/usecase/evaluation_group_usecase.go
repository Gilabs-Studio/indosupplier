package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// EvaluationGroupUsecase defines the interface for evaluation group business logic
type EvaluationGroupUsecase interface {
	// GetAll retrieves all evaluation groups with pagination and filters
	GetAll(ctx context.Context, page, perPage int, search string, isActive *bool) ([]*dto.EvaluationGroupResponse, *response.PaginationMeta, error)

	// GetByID retrieves an evaluation group by ID (with criteria)
	GetByID(ctx context.Context, id string) (*dto.EvaluationGroupResponse, error)

	// Create creates a new evaluation group
	Create(ctx context.Context, req *dto.CreateEvaluationGroupRequest) (*dto.EvaluationGroupResponse, error)

	// Update updates an existing evaluation group
	Update(ctx context.Context, id string, req *dto.UpdateEvaluationGroupRequest) (*dto.EvaluationGroupResponse, error)

	// Delete performs soft delete on an evaluation group
	Delete(ctx context.Context, id string) error

	// ListAuditTrail retrieves paginated audit trail rows for an evaluation group.
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.EvaluationAuditTrailEntry, int64, error)
}
