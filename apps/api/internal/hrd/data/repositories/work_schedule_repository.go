package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WorkScheduleRepository defines the interface for work schedule data access
type WorkScheduleRepository interface {
	FindByID(ctx context.Context, id string) (*models.WorkSchedule, error)
	FindByDivisionID(ctx context.Context, divisionID string) (*models.WorkSchedule, error)
	FindDefault(ctx context.Context) (*models.WorkSchedule, error)
	List(ctx context.Context, req *dto.ListWorkSchedulesRequest) ([]models.WorkSchedule, int64, error)
	Create(ctx context.Context, ws *models.WorkSchedule) error
	Update(ctx context.Context, ws *models.WorkSchedule) error
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, id string) error
}

type workScheduleRepository struct {
	db *gorm.DB
}

// NewWorkScheduleRepository creates a new WorkScheduleRepository
func NewWorkScheduleRepository(db *gorm.DB) WorkScheduleRepository {
	return &workScheduleRepository{db: db}
}

func (r *workScheduleRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *workScheduleRepository) FindByID(ctx context.Context, id string) (*models.WorkSchedule, error) {
	var ws models.WorkSchedule
	err := r.getDB(ctx).Where("id = ?", id).First(&ws).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *workScheduleRepository) FindByDivisionID(ctx context.Context, divisionID string) (*models.WorkSchedule, error) {
	var ws models.WorkSchedule
	err := r.getDB(ctx).Where("division_id = ? AND is_active = ?", divisionID, true).First(&ws).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *workScheduleRepository) FindDefault(ctx context.Context) (*models.WorkSchedule, error) {
	var ws models.WorkSchedule
	err := r.getDB(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&ws).Error
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *workScheduleRepository) List(ctx context.Context, req *dto.ListWorkSchedulesRequest) ([]models.WorkSchedule, int64, error) {
	var schedules []models.WorkSchedule
	var total int64

	query := r.getDB(ctx).Model(&models.WorkSchedule{})

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Apply active filter
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Apply division filter
	if req.DivisionID != "" {
		query = query.Where("division_id = ?", req.DivisionID)
	}

	// Apply flexible filter
	if req.IsFlexible != nil {
		query = query.Where("is_flexible = ?", *req.IsFlexible)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	// Apply sorting
	sortField := "created_at"
	sortOrder := "DESC"
	if req.SortBy != "" {
		switch req.SortBy {
		case "name", "start_time", "end_time", "created_at", "updated_at":
			sortField = req.SortBy
		}
	}
	if req.SortOrder != "" && (req.SortOrder == "asc" || req.SortOrder == "ASC") {
		sortOrder = "ASC"
	}

	// Fetch data
	err := query.Order(clause.OrderByColumn{
		Column: clause.Column{Name: sortField},
		Desc:   sortOrder == "DESC",
	}).Offset(offset).Limit(perPage).Find(&schedules).Error
	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

func (r *workScheduleRepository) Create(ctx context.Context, ws *models.WorkSchedule) error {
	return r.getDB(ctx).Create(ws).Error
}

func (r *workScheduleRepository) Update(ctx context.Context, ws *models.WorkSchedule) error {
	return r.getDB(ctx).Save(ws).Error
}

func (r *workScheduleRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.WorkSchedule{}, "id = ?", id).Error
}

func (r *workScheduleRepository) SetDefault(ctx context.Context, id string) error {
	db := r.getDB(ctx)

	// First, unset all defaults
	if err := db.Model(&models.WorkSchedule{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}

	// Then set the new default and ensure it is active
	return db.Model(&models.WorkSchedule{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_default": true,
		"is_active":  true,
	}).Error
}
