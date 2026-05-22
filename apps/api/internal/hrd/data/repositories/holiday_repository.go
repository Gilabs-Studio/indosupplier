package repositories

import (
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HolidayRepository defines the interface for holiday data access
type HolidayRepository interface {
	FindByID(ctx context.Context, id string) (*models.Holiday, error)
	FindByDate(ctx context.Context, date time.Time) (*models.Holiday, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Holiday, error)
	FindByYear(ctx context.Context, year int) ([]models.Holiday, error)
	List(ctx context.Context, req *dto.ListHolidaysRequest) ([]models.Holiday, int64, error)
	Create(ctx context.Context, h *models.Holiday) error
	CreateBatch(ctx context.Context, holidays []models.Holiday) error
	Update(ctx context.Context, h *models.Holiday) error
	Delete(ctx context.Context, id string) error
	IsHoliday(ctx context.Context, date time.Time) (bool, *models.Holiday, error)
	// Company-scoped variants: include global holidays (company_id IS NULL)
	// plus holidays for the specified company.
	IsHolidayForCompany(ctx context.Context, date time.Time, companyID string) (bool, *models.Holiday, error)
	FindByDateRangeForCompany(ctx context.Context, startDate, endDate time.Time, companyID string) ([]models.Holiday, error)
}

type holidayRepository struct {
	db *gorm.DB
}

// NewHolidayRepository creates a new HolidayRepository
func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepository{db: db}
}

func (r *holidayRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *holidayRepository) FindByID(ctx context.Context, id string) (*models.Holiday, error) {
	var h models.Holiday
	err := r.getDB(ctx).Where("id = ?", id).First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *holidayRepository) FindByDate(ctx context.Context, date time.Time) (*models.Holiday, error) {
	var h models.Holiday
	dateOnly := date.Format("2006-01-02")
	err := r.getDB(ctx).Where("date = ? AND is_active = ?", dateOnly, true).First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *holidayRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Holiday, error) {
	var holidays []models.Holiday
	err := r.getDB(ctx).
		Where("date >= ? AND date <= ? AND is_active = ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), true).
		Order("date ASC").
		Find(&holidays).Error
	if err != nil {
		return nil, err
	}
	return holidays, nil
}

func (r *holidayRepository) FindByYear(ctx context.Context, year int) ([]models.Holiday, error) {
	var holidays []models.Holiday
	err := r.getDB(ctx).
		Where("year = ? AND is_active = ?", year, true).
		Order("date ASC").
		Find(&holidays).Error
	if err != nil {
		return nil, err
	}
	return holidays, nil
}

func (r *holidayRepository) List(ctx context.Context, req *dto.ListHolidaysRequest) ([]models.Holiday, int64, error) {
	var holidays []models.Holiday
	var total int64

	query := r.getDB(ctx).Model(&models.Holiday{})

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Apply year filter
	if req.Year > 0 {
		query = query.Where("year = ?", req.Year)
	}

	// Apply type filter
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("date <= ?", req.DateTo)
	}

	// Apply active filter
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Apply company filter
	if req.CompanyID != "" {
		query = query.Where("(company_id IS NULL OR company_id = ?)", req.CompanyID)
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
	sortField := "date"
	sortOrder := "ASC"
	if req.SortBy != "" {
		switch req.SortBy {
		case "name", "date", "type", "year", "created_at":
			sortField = req.SortBy
		}
	}
	if req.SortOrder != "" && (req.SortOrder == "desc" || req.SortOrder == "DESC") {
		sortOrder = "DESC"
	}

	// Fetch data
	err := query.Order(clause.OrderByColumn{
		Column: clause.Column{Name: sortField},
		Desc:   sortOrder == "DESC",
	}).Offset(offset).Limit(perPage).Find(&holidays).Error
	if err != nil {
		return nil, 0, err
	}

	return holidays, total, nil
}

func (r *holidayRepository) Create(ctx context.Context, h *models.Holiday) error {
	return r.getDB(ctx).Create(h).Error
}

func (r *holidayRepository) CreateBatch(ctx context.Context, holidays []models.Holiday) error {
	return r.getDB(ctx).CreateInBatches(holidays, 100).Error
}

func (r *holidayRepository) Update(ctx context.Context, h *models.Holiday) error {
	return r.getDB(ctx).Save(h).Error
}

func (r *holidayRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.Holiday{}, "id = ?", id).Error
}

func (r *holidayRepository) IsHoliday(ctx context.Context, date time.Time) (bool, *models.Holiday, error) {
	holiday, err := r.FindByDate(ctx, date)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, holiday, nil
}

// IsHolidayForCompany checks whether the given date is a holiday that applies
// to the specified company. Matches global holidays (company_id IS NULL) OR
// holidays scoped to this company.
func (r *holidayRepository) IsHolidayForCompany(ctx context.Context, date time.Time, companyID string) (bool, *models.Holiday, error) {
	dateOnly := date.Format("2006-01-02")
	var h models.Holiday

	query := r.getDB(ctx).
		Where("date = ? AND is_active = ?", dateOnly, true)

	if companyID != "" {
		query = query.Where("(company_id IS NULL OR company_id = ?)", companyID)
	}

	tx := query.Limit(1).Find(&h)
	if tx.Error != nil {
		return false, nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return false, nil, nil
	}
	return true, &h, nil
}

// FindByDateRangeForCompany returns holidays in the date range that apply
// to the given company (global + company-specific).
func (r *holidayRepository) FindByDateRangeForCompany(ctx context.Context, startDate, endDate time.Time, companyID string) ([]models.Holiday, error) {
	var holidays []models.Holiday
	query := r.getDB(ctx).
		Where("date >= ? AND date <= ? AND is_active = ?",
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), true)

	if companyID != "" {
		query = query.Where("(company_id IS NULL OR company_id = ?)", companyID)
	}

	err := query.Order("date ASC").Find(&holidays).Error
	if err != nil {
		return nil, err
	}
	return holidays, nil
}
