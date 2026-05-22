package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
)

// EmployeeEvaluationRepository defines the interface for employee evaluation data access
type EmployeeEvaluationRepository interface {
	FindAll(ctx context.Context, page, perPage int, search string, employeeID, evaluationGroupID, evaluationType string) ([]models.EmployeeEvaluation, int64, error)
	FindByID(ctx context.Context, id string) (*models.EmployeeEvaluation, error)
	FindByIDWithDetails(ctx context.Context, id string) (*models.EmployeeEvaluation, error)
	Create(ctx context.Context, evaluation *models.EmployeeEvaluation) error
	Update(ctx context.Context, evaluation *models.EmployeeEvaluation) error
	Delete(ctx context.Context, id string) error
	SaveCriteriaScores(ctx context.Context, evaluationID string, scores []models.EmployeeEvaluationCriteria) error
	DeleteCriteriaScores(ctx context.Context, evaluationID string) error
}
