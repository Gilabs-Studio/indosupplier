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

// AttendanceRecordRepository defines the interface for attendance record data access
type AttendanceRecordRepository interface {
	FindByID(ctx context.Context, id string) (*models.AttendanceRecord, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*models.AttendanceRecord, error)
	FindByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]models.AttendanceRecord, error)
	List(ctx context.Context, req *dto.ListAttendanceRecordsRequest) ([]models.AttendanceRecord, int64, error)
	GetEmployeeMonthlyStats(ctx context.Context, employeeID string, year, month int) (*dto.MonthlyAttendanceStats, error)
	Create(ctx context.Context, ar *models.AttendanceRecord) error
	Update(ctx context.Context, ar *models.AttendanceRecord) error
	Delete(ctx context.Context, id string) error
	GetLateEmployeesForDate(ctx context.Context, date time.Time) ([]models.AttendanceRecord, error)
	GetAbsentEmployeesForDate(ctx context.Context, date time.Time, allEmployeeIDs []string) ([]string, error)
	DeleteByLeaveRequestID(ctx context.Context, leaveRequestID string) error
	CreateBatch(ctx context.Context, records []models.AttendanceRecord) error
}

type attendanceRecordRepository struct {
	db *gorm.DB
}

// NewAttendanceRecordRepository creates a new AttendanceRecordRepository
func NewAttendanceRecordRepository(db *gorm.DB) AttendanceRecordRepository {
	return &attendanceRecordRepository{db: db}
}

func (r *attendanceRecordRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *attendanceRecordRepository) FindByID(ctx context.Context, id string) (*models.AttendanceRecord, error) {
	var ar models.AttendanceRecord
	err := r.getDB(ctx).Where("id = ?", id).First(&ar).Error
	if err != nil {
		return nil, err
	}
	return &ar, nil
}

func (r *attendanceRecordRepository) FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*models.AttendanceRecord, error) {
	var ar models.AttendanceRecord
	dateOnly := date.Format("2006-01-02")
	err := r.getDB(ctx).Where("employee_id = ? AND date = ?", employeeID, dateOnly).First(&ar).Error
	if err != nil {
		return nil, err
	}
	return &ar, nil
}

func (r *attendanceRecordRepository) FindByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]models.AttendanceRecord, error) {
	var records []models.AttendanceRecord
	err := r.getDB(ctx).
		Where("employee_id = ? AND date >= ? AND date <= ?", employeeID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("date ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *attendanceRecordRepository) List(ctx context.Context, req *dto.ListAttendanceRecordsRequest) ([]models.AttendanceRecord, int64, error) {
	var records []models.AttendanceRecord
	var total int64

	query := r.getDB(ctx).Model(&models.AttendanceRecord{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Apply search filter (searches employee name and code via subquery)
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where(
			"employee_id IN (SELECT id FROM employees WHERE LOWER(name) LIKE LOWER(?) OR LOWER(employee_code) LIKE LOWER(?))",
			searchPattern, searchPattern,
		)
	}

	// Apply employee filter
	if req.EmployeeID != "" {
		query = query.Where("employee_id = ?", req.EmployeeID)
	}

	// Apply status filter
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Apply check-in type filter
	if req.CheckInType != "" {
		query = query.Where("check_in_type = ?", req.CheckInType)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("date <= ?", req.DateTo)
	}

	// Apply late filter
	if req.IsLate != nil && *req.IsLate {
		query = query.Where("late_minutes > 0")
	}

	// Apply early leave filter
	if req.IsEarlyLeave != nil && *req.IsEarlyLeave {
		query = query.Where("early_leave_minutes > 0")
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
	sortOrder := "DESC"
	if req.SortBy != "" {
		switch req.SortBy {
		case "date", "check_in_time", "check_out_time", "status", "working_minutes":
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
	}).Offset(offset).Limit(perPage).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *attendanceRecordRepository) GetEmployeeMonthlyStats(ctx context.Context, employeeID string, year, month int) (*dto.MonthlyAttendanceStats, error) {
	stats := &dto.MonthlyAttendanceStats{
		EmployeeID: employeeID,
		Year:       year,
		Month:      month,
	}

	// Get first and last day of month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, apptime.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	// Get all records for the month
	records, err := r.FindByEmployeeAndDateRange(ctx, employeeID, firstDay, lastDay)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		stats.TotalWorkingMinutes += record.WorkingMinutes
		stats.TotalOvertimeMinutes += record.OvertimeMinutes
		stats.TotalLateMinutes += record.LateMinutes
		stats.TotalEarlyLeaveMinutes += record.EarlyLeaveMinutes

		switch record.Status {
		case models.AttendanceStatusPresent, models.AttendanceStatusWFH:
			stats.PresentDays++
		case models.AttendanceStatusAbsent:
			stats.AbsentDays++
		case models.AttendanceStatusLate:
			stats.LateDays++
			stats.PresentDays++ // Late but still present
		case models.AttendanceStatusHalfDay:
			stats.HalfDays++
		case models.AttendanceStatusLeave:
			stats.LeaveDays++
		case models.AttendanceStatusHoliday:
			stats.HolidayDays++
		}
	}

	return stats, nil
}

func (r *attendanceRecordRepository) Create(ctx context.Context, ar *models.AttendanceRecord) error {
	return r.getDB(ctx).Create(ar).Error
}

func (r *attendanceRecordRepository) Update(ctx context.Context, ar *models.AttendanceRecord) error {
	return r.getDB(ctx).Save(ar).Error
}

func (r *attendanceRecordRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.AttendanceRecord{}, "id = ?", id).Error
}

func (r *attendanceRecordRepository) GetLateEmployeesForDate(ctx context.Context, date time.Time) ([]models.AttendanceRecord, error) {
	var records []models.AttendanceRecord
	dateOnly := date.Format("2006-01-02")
	err := r.getDB(ctx).
		Where("date = ? AND late_minutes > 0", dateOnly).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *attendanceRecordRepository) GetAbsentEmployeesForDate(ctx context.Context, date time.Time, allEmployeeIDs []string) ([]string, error) {
	dateOnly := date.Format("2006-01-02")

	// Get employees who have attendance record for this date
	var presentEmployeeIDs []string
	err := r.getDB(ctx).Model(&models.AttendanceRecord{}).
		Where("date = ?", dateOnly).
		Pluck("employee_id", &presentEmployeeIDs).Error
	if err != nil {
		return nil, err
	}

	// Find absent employees (those not in present list)
	presentMap := make(map[string]bool)
	for _, id := range presentEmployeeIDs {
		presentMap[id] = true
	}

	var absentEmployeeIDs []string
	for _, id := range allEmployeeIDs {
		if !presentMap[id] {
			absentEmployeeIDs = append(absentEmployeeIDs, id)
		}
	}

	return absentEmployeeIDs, nil
}

// DeleteByLeaveRequestID soft-deletes all attendance records linked to a specific leave request
func (r *attendanceRecordRepository) DeleteByLeaveRequestID(ctx context.Context, leaveRequestID string) error {
	return r.getDB(ctx).
		Where("leave_request_id = ?", leaveRequestID).
		Delete(&models.AttendanceRecord{}).Error
}

// CreateBatch bulk-inserts attendance records
func (r *attendanceRecordRepository) CreateBatch(ctx context.Context, records []models.AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.getDB(ctx).Create(&records).Error
}
