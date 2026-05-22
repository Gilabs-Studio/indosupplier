package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LeadStatusRepository defines the interface for lead status data access
type LeadStatusRepository interface {
	Create(ctx context.Context, status *models.LeadStatus) error
	FindByID(ctx context.Context, id string) (*models.LeadStatus, error)
	FindByCode(ctx context.Context, code string) (*models.LeadStatus, error)
	List(ctx context.Context, params ListParams) ([]models.LeadStatus, int64, error)
	Update(ctx context.Context, status *models.LeadStatus) error
	Delete(ctx context.Context, id string) error
	FindDefault(ctx context.Context) (*models.LeadStatus, error)
	FindConverted(ctx context.Context) (*models.LeadStatus, error)
	GetMaxOrder(ctx context.Context) (int, error)
}

type leadStatusRepository struct {
	db *gorm.DB
}

// NewLeadStatusRepository creates a new lead status repository
func NewLeadStatusRepository(db *gorm.DB) LeadStatusRepository {
	return &leadStatusRepository{db: db}
}

func (r *leadStatusRepository) Create(ctx context.Context, status *models.LeadStatus) error {
	return database.GetDB(ctx, r.db).Create(status).Error
}

func (r *leadStatusRepository) FindByID(ctx context.Context, id string) (*models.LeadStatus, error) {
	var status models.LeadStatus
	err := database.GetDB(ctx, r.db).First(&status, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *leadStatusRepository) FindByCode(ctx context.Context, code string) (*models.LeadStatus, error) {
	var status models.LeadStatus
	err := database.GetDB(ctx, r.db).Where("code = ?", code).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *leadStatusRepository) List(ctx context.Context, params ListParams) ([]models.LeadStatus, int64, error) {
	var statuses []models.LeadStatus
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.LeadStatus{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: params.SortBy}, Desc: params.SortDir == "desc"})
	} else {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}, Desc: false}).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: false})
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&statuses).Error; err != nil {
		return nil, 0, err
	}

	return statuses, total, nil
}

func (r *leadStatusRepository) Update(ctx context.Context, status *models.LeadStatus) error {
	return database.GetDB(ctx, r.db).Save(status).Error
}

func (r *leadStatusRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.LeadStatus{}, "id = ?", id).Error
}

func (r *leadStatusRepository) FindDefault(ctx context.Context) (*models.LeadStatus, error) {
	var status models.LeadStatus
	err := database.GetDB(ctx, r.db).Where("is_default = ?", true).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *leadStatusRepository) FindConverted(ctx context.Context) (*models.LeadStatus, error) {
	var status models.LeadStatus
	err := database.GetDB(ctx, r.db).Where("is_converted = ?", true).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *leadStatusRepository) GetMaxOrder(ctx context.Context) (int, error) {
	var maxOrder int
	err := database.GetDB(ctx, r.db).
		Model(&models.LeadStatus{}).
		Select(`COALESCE(MAX("order"), 0)`).
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}
