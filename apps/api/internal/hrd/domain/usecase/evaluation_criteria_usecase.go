package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// EvaluationCriteriaUsecase defines the interface for evaluation criteria business logic
type EvaluationCriteriaUsecase interface {
	// GetByGroupID retrieves all criteria for a specific evaluation group
	GetByGroupID(ctx context.Context, groupID string) ([]*dto.EvaluationCriteriaResponse, error)

	// GetByID retrieves an evaluation criteria by ID
	GetByID(ctx context.Context, id string) (*dto.EvaluationCriteriaResponse, error)

	// Create creates a new evaluation criteria (validates weight sum <= 100%)
	Create(ctx context.Context, req *dto.CreateEvaluationCriteriaRequest) (*dto.EvaluationCriteriaResponse, error)

	// Update updates an existing evaluation criteria (validates weight sum <= 100%)
	Update(ctx context.Context, id string, req *dto.UpdateEvaluationCriteriaRequest) (*dto.EvaluationCriteriaResponse, error)

	// Delete performs soft delete on an evaluation criteria
	Delete(ctx context.Context, id string) error
}
