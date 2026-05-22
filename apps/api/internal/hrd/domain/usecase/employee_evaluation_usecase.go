package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// EmployeeEvaluationUsecase defines the interface for employee evaluation business logic
type EmployeeEvaluationUsecase interface {
	// GetAll retrieves all employee evaluations with pagination and filters
	GetAll(ctx context.Context, page, perPage int, search, employeeID, evaluationGroupID, evaluationType string) ([]*dto.EmployeeEvaluationResponse, *response.PaginationMeta, error)

	// GetByID retrieves an employee evaluation by ID (with details)
	GetByID(ctx context.Context, id string) (*dto.EmployeeEvaluationResponse, error)

	// GetFormData retrieves form dropdown data (employees, groups, types)
	GetFormData(ctx context.Context) (*dto.EmployeeEvaluationFormDataResponse, error)

	// Create creates a new employee evaluation with criteria scores
	Create(ctx context.Context, req *dto.CreateEmployeeEvaluationRequest) (*dto.EmployeeEvaluationResponse, error)

	// Update updates an existing employee evaluation
	Update(ctx context.Context, id string, req *dto.UpdateEmployeeEvaluationRequest) (*dto.EmployeeEvaluationResponse, error)

	// Delete performs soft delete on an employee evaluation
	Delete(ctx context.Context, id string) error

	// ListAuditTrail retrieves paginated audit trail rows for an employee evaluation.
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.EvaluationAuditTrailEntry, int64, error)
}
