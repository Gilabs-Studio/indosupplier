package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
)

type evaluationCriteriaRepositoryImpl struct {
	db *gorm.DB
}

// NewEvaluationCriteriaRepository creates a new instance of EvaluationCriteriaRepository
func NewEvaluationCriteriaRepository(db *gorm.DB) EvaluationCriteriaRepository {
	return &evaluationCriteriaRepositoryImpl{db: db}
}

func (r *evaluationCriteriaRepositoryImpl) FindByGroupID(ctx context.Context, groupID string) ([]models.EvaluationCriteria, error) {
	var criteria []models.EvaluationCriteria
	if err := database.GetDB(ctx, r.db).
		Where("evaluation_group_id = ?", groupID).
		Order("sort_order ASC").
		Find(&criteria).Error; err != nil {
		return nil, fmt.Errorf("failed to find criteria by group ID: %w", err)
	}
	return criteria, nil
}

func (r *evaluationCriteriaRepositoryImpl) FindByID(ctx context.Context, id string) (*models.EvaluationCriteria, error) {
	var criteria models.EvaluationCriteria
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&criteria).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find evaluation criteria by ID: %w", err)
	}
	return &criteria, nil
}

func (r *evaluationCriteriaRepositoryImpl) Create(ctx context.Context, criteria *models.EvaluationCriteria) error {
	if err := database.GetDB(ctx, r.db).Create(criteria).Error; err != nil {
		return fmt.Errorf("failed to create evaluation criteria: %w", err)
	}
	return nil
}

func (r *evaluationCriteriaRepositoryImpl) Update(ctx context.Context, criteria *models.EvaluationCriteria) error {
	if err := database.GetDB(ctx, r.db).Save(criteria).Error; err != nil {
		return fmt.Errorf("failed to update evaluation criteria: %w", err)
	}
	return nil
}

func (r *evaluationCriteriaRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.EvaluationCriteria{}).Error; err != nil {
		return fmt.Errorf("failed to delete evaluation criteria: %w", err)
	}
	return nil
}

func (r *evaluationCriteriaRepositoryImpl) GetTotalWeightByGroupID(ctx context.Context, groupID string, excludeID string) (float64, error) {
	var totalWeight float64
	query := database.GetDB(ctx, r.db).
		Model(&models.EvaluationCriteria{}).
		Where("evaluation_group_id = ?", groupID)

	// Exclude a specific criteria (useful when updating to check remaining weight)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Select("COALESCE(SUM(weight), 0)").Scan(&totalWeight).Error; err != nil {
		return 0, fmt.Errorf("failed to get total weight: %w", err)
	}
	return totalWeight, nil
}
