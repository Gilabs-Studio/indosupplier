package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LeaveRequestRepository defines the interface for leave request data operations
type LeaveRequestRepository interface {
	// Basic CRUD
	Create(ctx context.Context, leaveRequest *models.LeaveRequest) error
	FindByID(ctx context.Context, id string) (*models.LeaveRequest, error)
	Update(ctx context.Context, leaveRequest *models.LeaveRequest) error
	Delete(ctx context.Context, id string) error

	// List with filters and pagination
	List(ctx context.Context, employeeID *string, status *models.LeaveStatus, startDate, endDate *time.Time, search *string, page, perPage int) ([]*models.LeaveRequest, int64, error)

	// Find by employee with pagination
	FindByEmployeeID(ctx context.Context, employeeID string, page, perPage int) ([]*models.LeaveRequest, int64, error)

	// Row-level locking for concurrent operations
	// WHY: Prevent race conditions during approval/rejection
	UpdateWithLock(ctx context.Context, id string, updateFn func(*models.LeaveRequest) error) error

	// Custom queries for business logic
	CountPendingRequests(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID *string) (int64, error)
	CalculateUsedLeaveDays(ctx context.Context, employeeID string, cutAnnualOnly bool) (int, error)
	FindOverlappingRequests(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID *string) ([]*models.LeaveRequest, error)

	// FindApprovedByDateForEmployees returns a map of employeeID → LeaveRequest for
	// employees who have an APPROVED leave request covering the given date.
	// Used by the auto-absent worker to skip employees on approved leave.
	FindApprovedByDateForEmployees(ctx context.Context, date time.Time, employeeIDs []string) (map[string]*models.LeaveRequest, error)
}

type leaveRequestRepository struct {
	db *gorm.DB
}

// NewLeaveRequestRepository creates a new instance of LeaveRequestRepository
func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &leaveRequestRepository{db: db}
}

// Create inserts a new leave request
func (r *leaveRequestRepository) Create(ctx context.Context, leaveRequest *models.LeaveRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return database.GetDB(ctx, r.db).Create(leaveRequest).Error
}

// FindByID retrieves a leave request by ID
func (r *leaveRequestRepository) FindByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var leaveRequest models.LeaveRequest

	// Note: Relations removed from model to prevent GORM circular dependency issues
	// Frontend/consumers should fetch related data via separate API calls if needed
	err := database.GetDB(ctx, r.db).
		Where("id = ?", id).
		First(&leaveRequest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("leave request not found")
		}
		return nil, err
	}

	return &leaveRequest, nil
}

// Update updates an existing leave request
func (r *leaveRequestRepository) Update(ctx context.Context, leaveRequest *models.LeaveRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Save(leaveRequest)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("leave request not found or no changes made")
	}

	return nil
}

// Delete soft deletes a leave request
// WHY: Soft delete for audit trail requirement
func (r *leaveRequestRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result := database.GetDB(ctx, r.db).Delete(&models.LeaveRequest{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("leave request not found")
	}

	return nil
}

// List retrieves leave requests with filters and pagination
// WHY: Enforce max 100 items per page to prevent memory issues
// WHY: Search functionality uses ILIKE for case-insensitive search across employee name, leave type, and reason
func (r *leaveRequestRepository) List(ctx context.Context, employeeID *string, status *models.LeaveStatus, startDate, endDate *time.Time, search *string, page, perPage int) ([]*models.LeaveRequest, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	var leaveRequests []*models.LeaveRequest
	var total int64

	// Use WithContext (not GetDB): search conditionally JOINs employees + leave_types which
	// both have tenant_id. GetDB's unqualified WHERE tenant_id=? causes PG ambiguity.
	query := r.db.WithContext(ctx).Model(&models.LeaveRequest{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Apply filters
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if startDate != nil {
		query = query.Where("start_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("end_date <= ?", *endDate)
	}

	// WHY: Search functionality requires joins with employees and leave_types tables
	// Search across employee name, leave type name, or reason (case-insensitive)
	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query = query.Joins("JOIN employees ON employees.id = leave_requests.employee_id").
			Joins("JOIN leave_types ON leave_types.id = leave_requests.leave_type_id").
			Where("employees.name ILIKE ? OR leave_types.name ILIKE ? OR leave_requests.reason ILIKE ?",
				searchPattern, searchPattern, searchPattern)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Note: Relations removed from model to prevent GORM circular dependency issues
	offset := (page - 1) * perPage
	err := query.
		Order("leave_requests.created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&leaveRequests).Error

	if err != nil {
		return nil, 0, err
	}

	return leaveRequests, total, nil
}

// FindByEmployeeID retrieves all leave requests for a specific employee with pagination
func (r *leaveRequestRepository) FindByEmployeeID(ctx context.Context, employeeID string, page, perPage int) ([]*models.LeaveRequest, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Enforce pagination limits
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	var leaveRequests []*models.LeaveRequest
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.LeaveRequest{}).Where("employee_id = ?", employeeID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Note: Relations removed from model to prevent GORM circular dependency issues
	offset := (page - 1) * perPage
	err := query.
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&leaveRequests).Error

	if err != nil {
		return nil, 0, err
	}

	return leaveRequests, total, nil
}

// UpdateWithLock updates a leave request with row-level locking (FOR UPDATE)
// WHY: Prevent race conditions during approval/rejection
func (r *leaveRequestRepository) UpdateWithLock(ctx context.Context, id string, updateFn func(*models.LeaveRequest) error) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var leaveRequest models.LeaveRequest

		// WHY: Clause Locking FOR UPDATE prevents concurrent approvals/rejections
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&leaveRequest).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("leave request not found")
			}
			return err
		}

		// Apply the update function
		if err := updateFn(&leaveRequest); err != nil {
			return err
		}

		// Save changes
		return tx.Save(&leaveRequest).Error
	})
}

// CountPendingRequests counts pending leave requests for an employee in a date range
// WHY: Validate overlapping leave requests to prevent conflicts
func (r *leaveRequestRepository) CountPendingRequests(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID *string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var count int64

	query := database.GetDB(ctx, r.db).Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where("status = ?", models.LeaveStatusPending).
		Where("(start_date BETWEEN ? AND ? OR end_date BETWEEN ? AND ? OR (start_date <= ? AND end_date >= ?))",
			startDate, endDate, startDate, endDate, startDate, endDate)

	// Exclude current request when updating
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// CalculateUsedLeaveDays calculates total approved leave days for an employee
// WHY: Used for leave balance calculation - only count approved requests where LeaveType.IsCutAnnualLeave = true
func (r *leaveRequestRepository) CalculateUsedLeaveDays(ctx context.Context, employeeID string, cutAnnualOnly bool) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var totalDays float64

	query := r.db.WithContext(ctx).Model(&models.LeaveRequest{}).
		Select("COALESCE(SUM(total_days), 0)").
		Where("employee_id = ?", employeeID).
		Where("status = ?", models.LeaveStatusApproved)

	// WHY: If cutAnnualOnly is true, we need to join with leave_types table
	// to filter only leave types that cut annual leave balance
	if cutAnnualOnly {
		query = query.Joins("JOIN leave_types ON leave_types.id = leave_requests.leave_type_id").
			Where("leave_types.is_cut_annual_leave = ?", true)
	}

	err := query.Scan(&totalDays).Error
	if err != nil {
		return 0, err
	}

	// Round to nearest integer for simplicity
	// WHY: Business rules typically work with full days for balance
	return int(totalDays + 0.5), nil
}

// FindOverlappingRequests finds leave requests that overlap with a given date range
// WHY: Prevent scheduling conflicts and validate leave request submissions
func (r *leaveRequestRepository) FindOverlappingRequests(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID *string) ([]*models.LeaveRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var leaveRequests []*models.LeaveRequest

	query := database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Where("status IN ?", []models.LeaveStatus{models.LeaveStatusPending, models.LeaveStatusApproved}).
		Where("(start_date BETWEEN ? AND ? OR end_date BETWEEN ? AND ? OR (start_date <= ? AND end_date >= ?))",
			startDate, endDate, startDate, endDate, startDate, endDate)

	// Exclude current request when updating
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Find(&leaveRequests).Error
	return leaveRequests, err
}

// FindApprovedByDateForEmployees returns a map of employeeID → *LeaveRequest for employees
// who have an APPROVED leave request covering the given date.
func (r *leaveRequestRepository) FindApprovedByDateForEmployees(ctx context.Context, date time.Time, employeeIDs []string) (map[string]*models.LeaveRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if len(employeeIDs) == 0 {
		return make(map[string]*models.LeaveRequest), nil
	}

	var leaveRequests []*models.LeaveRequest
	dateOnly := date.Format("2006-01-02")

	err := database.GetDB(ctx, r.db).
		Where("employee_id IN ?", employeeIDs).
		Where("status = ?", models.LeaveStatusApproved).
		Where("start_date <= ? AND end_date >= ?", dateOnly, dateOnly).
		Find(&leaveRequests).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.LeaveRequest, len(leaveRequests))
	for _, lr := range leaveRequests {
		result[lr.EmployeeID] = lr
	}

	return result, nil
}
