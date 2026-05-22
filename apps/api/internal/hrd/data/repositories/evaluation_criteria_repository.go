package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
)

// EvaluationCriteriaRepository defines the interface for evaluation criteria data access
type EvaluationCriteriaRepository interface {
	FindByGroupID(ctx context.Context, groupID string) ([]models.EvaluationCriteria, error)
	FindByID(ctx context.Context, id string) (*models.EvaluationCriteria, error)
	Create(ctx context.Context, criteria *models.EvaluationCriteria) error
	Update(ctx context.Context, criteria *models.EvaluationCriteria) error
	Delete(ctx context.Context, id string) error
	GetTotalWeightByGroupID(ctx context.Context, groupID string, excludeID string) (float64, error)
}
