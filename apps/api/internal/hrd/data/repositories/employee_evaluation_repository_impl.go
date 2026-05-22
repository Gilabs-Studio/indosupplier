package repositories

import (
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
)

type employeeEvaluationRepositoryImpl struct {
	db *gorm.DB
}

// NewEmployeeEvaluationRepository creates a new instance of EmployeeEvaluationRepository
func NewEmployeeEvaluationRepository(db *gorm.DB) EmployeeEvaluationRepository {
	return &employeeEvaluationRepositoryImpl{db: db}
}

func (r *employeeEvaluationRepositoryImpl) FindAll(ctx context.Context, page, perPage int, search string, employeeID, evaluationGroupID, evaluationType string) ([]models.EmployeeEvaluation, int64, error) {
	var evaluations []models.EmployeeEvaluation
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.EmployeeEvaluation{})

	// Enforce HRD scope (OWN/DIVISION/AREA/OUTLET/ALL) before additional filters.
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Apply search filter on employee name, evaluation group name, and notes
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"employee_evaluations.notes ILIKE ? OR employee_evaluations.employee_id::text IN (SELECT id::text FROM employees WHERE name ILIKE ? AND deleted_at IS NULL) OR employee_evaluations.evaluation_group_id::text IN (SELECT id::text FROM evaluation_groups WHERE name ILIKE ? AND deleted_at IS NULL)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Filter by employee_id
	if employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}

	// Filter by evaluation_group_id
	if evaluationGroupID != "" {
		query = query.Where("evaluation_group_id = ?", evaluationGroupID)
	}

	// Filter by evaluation_type
	if evaluationType != "" {
		query = query.Where("evaluation_type = ?", evaluationType)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count employee evaluations: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Order by created_at DESC (newest first)
	query = query.Order("created_at DESC")

	// Preload EvaluationGroup for list display
	query = query.Preload("EvaluationGroup")

	if err := query.Find(&evaluations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find employee evaluations: %w", err)
	}

	return evaluations, total, nil
}

func (r *employeeEvaluationRepositoryImpl) FindByID(ctx context.Context, id string) (*models.EmployeeEvaluation, error) {
	var evaluation models.EmployeeEvaluation
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&evaluation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee evaluation by ID: %w", err)
	}
	return &evaluation, nil
}

func (r *employeeEvaluationRepositoryImpl) FindByIDWithDetails(ctx context.Context, id string) (*models.EmployeeEvaluation, error) {
	var evaluation models.EmployeeEvaluation
	if err := database.GetDB(ctx, r.db).
		Preload("EvaluationGroup").
		Preload("CriteriaScores").
		Where("id = ?", id).
		First(&evaluation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee evaluation with details: %w", err)
	}
	return &evaluation, nil
}

func (r *employeeEvaluationRepositoryImpl) Create(ctx context.Context, evaluation *models.EmployeeEvaluation) error {
	if err := database.GetDB(ctx, r.db).Create(evaluation).Error; err != nil {
		return fmt.Errorf("failed to create employee evaluation: %w", err)
	}
	return nil
}

func (r *employeeEvaluationRepositoryImpl) Update(ctx context.Context, evaluation *models.EmployeeEvaluation) error {
	if err := database.GetDB(ctx, r.db).Save(evaluation).Error; err != nil {
		return fmt.Errorf("failed to update employee evaluation: %w", err)
	}
	return nil
}

func (r *employeeEvaluationRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.EmployeeEvaluation{}).Error; err != nil {
		return fmt.Errorf("failed to delete employee evaluation: %w", err)
	}
	return nil
}

func (r *employeeEvaluationRepositoryImpl) SaveCriteriaScores(ctx context.Context, evaluationID string, scores []models.EmployeeEvaluationCriteria) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Delete existing scores for this evaluation
		if err := tx.Where("employee_evaluation_id = ?", evaluationID).
			Delete(&models.EmployeeEvaluationCriteria{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing criteria scores: %w", err)
		}

		// Insert new scores
		if len(scores) > 0 {
			if err := tx.Create(&scores).Error; err != nil {
				return fmt.Errorf("failed to create criteria scores: %w", err)
			}
		}

		return nil
	})
}

func (r *employeeEvaluationRepositoryImpl) DeleteCriteriaScores(ctx context.Context, evaluationID string) error {
	if err := database.GetDB(ctx, r.db).
		Where("employee_evaluation_id = ?", evaluationID).
		Delete(&models.EmployeeEvaluationCriteria{}).Error; err != nil {
		return fmt.Errorf("failed to delete criteria scores: %w", err)
	}
	return nil
}
