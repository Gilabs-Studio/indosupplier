package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
)

type evaluationGroupRepositoryImpl struct {
	db *gorm.DB
}

// NewEvaluationGroupRepository creates a new instance of EvaluationGroupRepository
func NewEvaluationGroupRepository(db *gorm.DB) EvaluationGroupRepository {
	return &evaluationGroupRepositoryImpl{db: db}
}

func (r *evaluationGroupRepositoryImpl) FindAll(ctx context.Context, page, perPage int, search string, isActive *bool) ([]models.EvaluationGroup, int64, error) {
	var groups []models.EvaluationGroup
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.EvaluationGroup{})

	// Apply search filter (prefix search for GIN index)
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ?", searchPattern)
	}

	// Filter by is_active
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count evaluation groups: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Order by name ASC
	query = query.Order("name ASC")

	// Preload criteria to calculate total weight
	query = query.Preload("Criteria", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	})

	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find evaluation groups: %w", err)
	}

	return groups, total, nil
}

func (r *evaluationGroupRepositoryImpl) FindByID(ctx context.Context, id string) (*models.EvaluationGroup, error) {
	var group models.EvaluationGroup
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find evaluation group by ID: %w", err)
	}
	return &group, nil
}

func (r *evaluationGroupRepositoryImpl) FindByIDWithCriteria(ctx context.Context, id string) (*models.EvaluationGroup, error) {
	var group models.EvaluationGroup
	if err := database.GetDB(ctx, r.db).
		Preload("Criteria", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).
		First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find evaluation group with criteria: %w", err)
	}
	return &group, nil
}

func (r *evaluationGroupRepositoryImpl) Create(ctx context.Context, group *models.EvaluationGroup) error {
	if err := database.GetDB(ctx, r.db).Create(group).Error; err != nil {
		return fmt.Errorf("failed to create evaluation group: %w", err)
	}
	return nil
}

func (r *evaluationGroupRepositoryImpl) Update(ctx context.Context, group *models.EvaluationGroup) error {
	if err := database.GetDB(ctx, r.db).Save(group).Error; err != nil {
		return fmt.Errorf("failed to update evaluation group: %w", err)
	}
	return nil
}

func (r *evaluationGroupRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.EvaluationGroup{}).Error; err != nil {
		return fmt.Errorf("failed to delete evaluation group: %w", err)
	}
	return nil
}
