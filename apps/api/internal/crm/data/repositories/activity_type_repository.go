package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActivityTypeRepository defines the interface for activity type data access
type ActivityTypeRepository interface {
	Create(ctx context.Context, actType *models.ActivityType) error
	FindByID(ctx context.Context, id string) (*models.ActivityType, error)
	List(ctx context.Context, params ListParams) ([]models.ActivityType, int64, error)
	Update(ctx context.Context, actType *models.ActivityType) error
	Delete(ctx context.Context, id string) error
	GetMaxOrder(ctx context.Context) (int, error)
}

type activityTypeRepository struct {
	db *gorm.DB
}

// NewActivityTypeRepository creates a new activity type repository
func NewActivityTypeRepository(db *gorm.DB) ActivityTypeRepository {
	return &activityTypeRepository{db: db}
}

func (r *activityTypeRepository) Create(ctx context.Context, actType *models.ActivityType) error {
	return database.GetDB(ctx, r.db).Create(actType).Error
}

func (r *activityTypeRepository) FindByID(ctx context.Context, id string) (*models.ActivityType, error) {
	var actType models.ActivityType
	err := database.GetDB(ctx, r.db).First(&actType, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &actType, nil
}

func (r *activityTypeRepository) List(ctx context.Context, params ListParams) ([]models.ActivityType, int64, error) {
	var actTypes []models.ActivityType
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ActivityType{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: params.SortBy},
			Desc:   params.SortDir == "desc",
		})
	} else {
		query = query.Order("\"order\" ASC, name ASC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&actTypes).Error; err != nil {
		return nil, 0, err
	}

	return actTypes, total, nil
}

func (r *activityTypeRepository) Update(ctx context.Context, actType *models.ActivityType) error {
	return database.GetDB(ctx, r.db).Save(actType).Error
}

func (r *activityTypeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ActivityType{}, "id = ?", id).Error
}

func (r *activityTypeRepository) GetMaxOrder(ctx context.Context) (int, error) {
	var maxOrder int
	err := database.GetDB(ctx, r.db).
		Model(&models.ActivityType{}).
		Select(`COALESCE(MAX("order"), 0)`).
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}
