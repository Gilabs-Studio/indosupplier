package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
)

// EvaluationGroupRepository defines the interface for evaluation group data access
type EvaluationGroupRepository interface {
	FindAll(ctx context.Context, page, perPage int, search string, isActive *bool) ([]models.EvaluationGroup, int64, error)
	FindByID(ctx context.Context, id string) (*models.EvaluationGroup, error)
	FindByIDWithCriteria(ctx context.Context, id string) (*models.EvaluationGroup, error)
	Create(ctx context.Context, group *models.EvaluationGroup) error
	Update(ctx context.Context, group *models.EvaluationGroup) error
	Delete(ctx context.Context, id string) error
}
