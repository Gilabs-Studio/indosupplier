package repositories

import (
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OvertimeRequestRepository defines the interface for overtime request data access
type OvertimeRequestRepository interface {
	FindByID(ctx context.Context, id string) (*models.OvertimeRequest, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) ([]models.OvertimeRequest, error)
	FindPendingByManager(ctx context.Context, managerID string) ([]models.OvertimeRequest, error)
	List(ctx context.Context, req *dto.ListOvertimeRequestsRequest) ([]models.OvertimeRequest, int64, error)
	Create(ctx context.Context, or *models.OvertimeRequest) error
	Update(ctx context.Context, or *models.OvertimeRequest) error
	Delete(ctx context.Context, id string) error
	Approve(ctx context.Context, id, approverID string, approvedMinutes int) error
	Reject(ctx context.Context, id, rejecterID, reason string) error
	Cancel(ctx context.Context, id string) error
	GetEmployeeMonthlyOvertime(ctx context.Context, employeeID string, year, month int) (int, error)
	GetUnnotifiedPendingRequests(ctx context.Context) ([]models.OvertimeRequest, error)
	MarkNotified(ctx context.Context, ids []string) error
}

type overtimeRequestRepository struct {
	db *gorm.DB
}

// NewOvertimeRequestRepository creates a new OvertimeRequestRepository
func NewOvertimeRequestRepository(db *gorm.DB) OvertimeRequestRepository {
	return &overtimeRequestRepository{db: db}
}

func (r *overtimeRequestRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *overtimeRequestRepository) FindByID(ctx context.Context, id string) (*models.OvertimeRequest, error) {
	var or models.OvertimeRequest
	err := r.getDB(ctx).Where("id = ?", id).First(&or).Error
	if err != nil {
		return nil, err
	}
	return &or, nil
}

func (r *overtimeRequestRepository) FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) ([]models.OvertimeRequest, error) {
	var requests []models.OvertimeRequest
	dateOnly := date.Format("2006-01-02")
	err := r.getDB(ctx).
		Where("employee_id = ? AND date = ?", employeeID, dateOnly).
		Order("start_time ASC").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *overtimeRequestRepository) FindPendingByManager(ctx context.Context, managerID string) ([]models.OvertimeRequest, error) {
	var requests []models.OvertimeRequest
	_ = managerID
	query := r.getDB(ctx).Model(&models.OvertimeRequest{})

	// Enforce scope on pending queue so approvers only see records inside their data scope.
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	err := query.
		Where("status = ?", models.OvertimeStatusPending).
		Order("created_at ASC").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *overtimeRequestRepository) List(ctx context.Context, req *dto.ListOvertimeRequestsRequest) ([]models.OvertimeRequest, int64, error) {
	var requests []models.OvertimeRequest
	var total int64

	query := r.getDB(ctx).Model(&models.OvertimeRequest{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Apply employee filter
	if req.EmployeeID != "" {
		query = query.Where("employee_id = ?", req.EmployeeID)
	}

	// Apply status filter
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Apply request type filter
	if req.RequestType != "" {
		query = query.Where("request_type = ?", req.RequestType)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("date <= ?", req.DateTo)
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
		case "date", "start_time", "status", "actual_minutes", "created_at":
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
	}).Offset(offset).Limit(perPage).Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *overtimeRequestRepository) Create(ctx context.Context, or *models.OvertimeRequest) error {
	return r.getDB(ctx).Create(or).Error
}

func (r *overtimeRequestRepository) Update(ctx context.Context, or *models.OvertimeRequest) error {
	return r.getDB(ctx).Save(or).Error
}

func (r *overtimeRequestRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.OvertimeRequest{}, "id = ?", id).Error
}

func (r *overtimeRequestRepository) Approve(ctx context.Context, id, approverID string, approvedMinutes int) error {
	now := apptime.Now()
	return r.getDB(ctx).Model(&models.OvertimeRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           models.OvertimeStatusApproved,
			"approved_by":      approverID,
			"approved_at":      now,
			"approved_minutes": approvedMinutes,
		}).Error
}

func (r *overtimeRequestRepository) Reject(ctx context.Context, id, rejecterID, reason string) error {
	now := apptime.Now()
	return r.getDB(ctx).Model(&models.OvertimeRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.OvertimeStatusRejected,
			"rejected_by":   rejecterID,
			"rejected_at":   now,
			"reject_reason": reason,
		}).Error
}

func (r *overtimeRequestRepository) Cancel(ctx context.Context, id string) error {
	return r.getDB(ctx).Model(&models.OvertimeRequest{}).
		Where("id = ?", id).
		Update("status", models.OvertimeStatusCanceled).Error
}

func (r *overtimeRequestRepository) GetEmployeeMonthlyOvertime(ctx context.Context, employeeID string, year, month int) (int, error) {
	// Get first and last day of month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, apptime.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	var totalMinutes int64
	err := r.getDB(ctx).Model(&models.OvertimeRequest{}).
		Where("employee_id = ? AND date >= ? AND date <= ? AND status = ?",
			employeeID, firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), models.OvertimeStatusApproved).
		Select("COALESCE(SUM(approved_minutes), 0)").
		Scan(&totalMinutes).Error
	if err != nil {
		return 0, err
	}

	return int(totalMinutes), nil
}

func (r *overtimeRequestRepository) GetUnnotifiedPendingRequests(ctx context.Context) ([]models.OvertimeRequest, error) {
	var requests []models.OvertimeRequest
	err := r.getDB(ctx).
		Where("status = ? AND is_manager_notified = ?", models.OvertimeStatusPending, false).
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *overtimeRequestRepository) MarkNotified(ctx context.Context, ids []string) error {
	now := apptime.Now()
	return r.getDB(ctx).Model(&models.OvertimeRequest{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"is_manager_notified":  true,
			"manager_notified_at": now,
		}).Error
}
