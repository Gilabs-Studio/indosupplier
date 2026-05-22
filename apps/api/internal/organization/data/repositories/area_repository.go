package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AreaRepository defines the interface for area data access
type AreaRepository interface {
	FindByID(ctx context.Context, id string) (*models.Area, error)
	FindByIDWithMembers(ctx context.Context, id string) (*models.Area, error)
	FindByName(ctx context.Context, name string) (*models.Area, error)
	FindAll(ctx context.Context) ([]models.Area, error)
	FindByIDs(ctx context.Context, ids []string) ([]models.Area, error)
	List(ctx context.Context, req *dto.ListAreasRequest) ([]models.Area, int64, error)
	Create(ctx context.Context, a *models.Area) error
	Update(ctx context.Context, a *models.Area) error
	Delete(ctx context.Context, id string) error
	// HasAssignedEmployees checks if any employees (supervisor or member) are assigned to the area.
	HasAssignedEmployees(ctx context.Context, areaID string) (bool, error)
}

type areaRepository struct {
	db *gorm.DB
}

// NewAreaRepository creates a new AreaRepository
func NewAreaRepository(db *gorm.DB) AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *areaRepository) FindByID(ctx context.Context, id string) (*models.Area, error) {
	var area models.Area
	err := r.getDB(ctx).Where("id = ?", id).First(&area).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (r *areaRepository) FindByIDWithMembers(ctx context.Context, id string) (*models.Area, error) {
	var area models.Area
	err := r.getDB(ctx).
		Preload("Manager").
		Preload("EmployeeAreas").
		Preload("EmployeeAreas.Employee").
		Preload("EmployeeAreas.Employee.Division").
		Preload("EmployeeAreas.Employee.JobPosition").
		Where("id = ?", id).
		First(&area).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (r *areaRepository) FindByName(ctx context.Context, name string) (*models.Area, error) {
	var area models.Area
	err := r.getDB(ctx).Where("name = ?", name).First(&area).Error
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (r *areaRepository) FindAll(ctx context.Context) ([]models.Area, error) {
	var areas []models.Area
	err := r.getDB(ctx).Find(&areas).Error
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (r *areaRepository) FindByIDs(ctx context.Context, ids []string) ([]models.Area, error) {
	var areas []models.Area
	err := r.getDB(ctx).Where("id IN ?", ids).Find(&areas).Error
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (r *areaRepository) List(ctx context.Context, req *dto.ListAreasRequest) ([]models.Area, int64, error) {
	var areas []models.Area
	var total int64

	query := r.getDB(ctx).Model(&models.Area{})

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		// Prefix search keeps queries index-friendly on large tables.
		search := "%" + s + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ? OR code ILIKE ? OR province ILIKE ?", search, search, search, search)
	}

	// Filter by province
	if province := strings.TrimSpace(req.Province); province != "" {
		query = query.Where("province ILIKE ?", province+"%")
	}

	// Filter by supervisor presence using subquery on employee_areas
	if req.HasSupervisor != nil {
		subquery := r.getDB(ctx).Table("employee_areas").
			Select("area_id").
			Where("is_supervisor = ? AND deleted_at IS NULL", true)
		if *req.HasSupervisor {
			query = query.Where("id IN (?)", subquery)
		} else {
			query = query.Where("id NOT IN (?)", subquery)
		}
	}

	// Filter by member presence using subquery on employee_areas
	if req.HasMembers != nil {
		subquery := r.getDB(ctx).Table("employee_areas").
			Select("area_id").
			Where("is_supervisor = ? AND deleted_at IS NULL", false)
		if *req.HasMembers {
			query = query.Where("id IN (?)", subquery)
		} else {
			query = query.Where("id NOT IN (?)", subquery)
		}
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
	sortBy := "updated_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}

	err := query.Order("is_active DESC").
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: sortBy},
			Desc:   req.SortDir != "asc",
		}).
		Offset(offset).
		Limit(perPage).
		Preload("Manager").
		Preload("EmployeeAreas.Employee").
		Find(&areas).Error
	if err != nil {
		return nil, 0, err
	}

	return areas, total, nil
}

func (r *areaRepository) Create(ctx context.Context, a *models.Area) error {
	return r.getDB(ctx).Create(a).Error
}

func (r *areaRepository) Update(ctx context.Context, a *models.Area) error {
	return r.getDB(ctx).Save(a).Error
}

func (r *areaRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.Area{}).Error
}

func (r *areaRepository) HasAssignedEmployees(ctx context.Context, areaID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&models.EmployeeArea{}).Where("area_id = ?", areaID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
