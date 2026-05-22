package repositories

import (
	"context"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"

	"github.com/gilabs/indosupplier/api/internal/core/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LeaveTypeRepository defines the interface for leave type data access
type LeaveTypeRepository interface {
	Create(ctx context.Context, leaveType *models.LeaveType) error
	FindByID(ctx context.Context, id string) (*models.LeaveType, error)
	FindByIDs(ctx context.Context, ids []string) ([]models.LeaveType, error)
	FindAll(ctx context.Context) ([]models.LeaveType, error)
	List(ctx context.Context, params ListParams) ([]models.LeaveType, int64, error)
	Update(ctx context.Context, leaveType *models.LeaveType) error
	Delete(ctx context.Context, id string) error
}

type leaveTypeRepository struct {
	db *gorm.DB
}

// NewLeaveTypeRepository creates a new instance of LeaveTypeRepository
func NewLeaveTypeRepository(db *gorm.DB) LeaveTypeRepository {
	return &leaveTypeRepository{db: db}
}

func (r *leaveTypeRepository) Create(ctx context.Context, leaveType *models.LeaveType) error {
	return database.GetDB(ctx, r.db).Create(leaveType).Error
}

func (r *leaveTypeRepository) FindByID(ctx context.Context, id string) (*models.LeaveType, error) {
	var leaveType models.LeaveType
	err := database.GetDB(ctx, r.db).First(&leaveType, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &leaveType, nil
}

func (r *leaveTypeRepository) FindByIDs(ctx context.Context, ids []string) ([]models.LeaveType, error) {
	var leaveTypes []models.LeaveType
	err := database.GetDB(ctx, r.db).
		Where("id IN ?", ids).
		Find(&leaveTypes).Error
	if err != nil {
		return nil, err
	}
	return leaveTypes, nil
}

func (r *leaveTypeRepository) FindAll(ctx context.Context) ([]models.LeaveType, error) {
	var leaveTypes []models.LeaveType
	err := database.GetDB(ctx, r.db).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&leaveTypes).Error
	if err != nil {
		return nil, err
	}
	return leaveTypes, nil
}

func (r *leaveTypeRepository) List(ctx context.Context, params ListParams) ([]models.LeaveType, int64, error) {
	var leaveTypes []models.LeaveType
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.LeaveType{})

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
		query = query.Order("name ASC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&leaveTypes).Error; err != nil {
		return nil, 0, err
	}

	return leaveTypes, total, nil
}

func (r *leaveTypeRepository) Update(ctx context.Context, leaveType *models.LeaveType) error {
	return database.GetDB(ctx, r.db).Save(leaveType).Error
}

func (r *leaveTypeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.LeaveType{}, "id = ?", id).Error
}
