package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	coreMiddleware "github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	orgDTO "github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAttendanceNotFound = errors.New("attendance record not found")
	ErrAlreadyCheckedIn   = errors.New("already checked in for today")
	ErrNotCheckedIn       = errors.New("not checked in yet")
	ErrAlreadyCheckedOut  = errors.New("already checked out for today")
	ErrGPSRequired        = errors.New("GPS location is required")
	ErrOutsideGPSRadius   = errors.New("you are outside the allowed GPS radius")
	ErrNotWorkingDay      = errors.New("today is not a working day")
	ErrHolidayNoCheckIn   = errors.New("cannot check in on holiday")
	ErrHolidayNoCheckOut  = errors.New("cannot check out on holiday")
	ErrOffDayNoCheckOut   = errors.New("cannot check out on off day")
	ErrLateReasonRequired = errors.New("late reason is required when clocking in late")
	ErrPhotoRequired      = errors.New("photo proof is required for WFH and field work clock-in")
	ErrTooEarlyToCheckIn  = errors.New("cannot check in before scheduled start time")
)

// AttendanceRecordUsecase defines the interface for attendance record business logic
type AttendanceRecordUsecase interface {
	List(ctx context.Context, req *dto.ListAttendanceRecordsRequest) ([]dto.AttendanceRecordResponse, *utils.PaginationResult, error)
	// ListSelf returns attendance history for the authenticated employee, resolving userID → employeeID.
	ListSelf(ctx context.Context, req *dto.ListAttendanceRecordsRequest, userID string) ([]dto.AttendanceRecordResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.AttendanceRecordResponse, error)
	GetTodayAttendance(ctx context.Context, employeeID string) (*dto.TodayAttendanceResponse, error)
	ClockIn(ctx context.Context, employeeID string, req *dto.ClockInRequest) (*dto.AttendanceRecordResponse, error)
	ClockOut(ctx context.Context, employeeID string, req *dto.ClockOutRequest) (*dto.AttendanceRecordResponse, error)
	CreateManualEntry(ctx context.Context, req *dto.ManualAttendanceRequest, createdBy string) (*dto.AttendanceRecordResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateAttendanceRecordRequest) (*dto.AttendanceRecordResponse, error)
	Delete(ctx context.Context, id string) error
	GetMonthlyStats(ctx context.Context, req *dto.MonthlyReportRequest) ([]dto.MonthlyAttendanceStats, error)
	// GetSelfMonthlyStats returns monthly stats for the authenticated employee, resolving userID → employeeID.
	GetSelfMonthlyStats(ctx context.Context, req *dto.MonthlyReportRequest, userID string) ([]dto.MonthlyAttendanceStats, error)
	GetFormData(ctx context.Context) (*dto.AttendanceFormDataResponse, error)
	// GetEmployeeSchedule returns the work schedule assigned to an employee (via division or default fallback).
	GetEmployeeSchedule(ctx context.Context, employeeID string) (*dto.EmployeeScheduleResponse, error)
	// ProcessAutoAbsent creates ABSENT/LEAVE records for employees who didn't clock in on the given date.
	// companyID scopes the processing to a single company's timezone and holidays ("" = all employees, global holidays).
	ProcessAutoAbsent(ctx context.Context, date time.Time, companyID string) (*dto.AutoAbsentResult, error)
}

type AttendanceTodayPublisher interface {
	Publish(tenantID string, eventType string, payload map[string]interface{})
}

type attendanceRecordUsecase struct {
	attendanceRepo    repositories.AttendanceRecordRepository
	workScheduleRepo  repositories.WorkScheduleRepository
	holidayRepo       repositories.HolidayRepository
	leaveRequestRepo  repositories.LeaveRequestRepository
	employeeRepo      orgRepos.EmployeeRepository
	divisionRepo      orgRepos.DivisionRepository
	overtimeRequestUC OvertimeRequestUsecase
	todayPublisher    AttendanceTodayPublisher
	mapper            *mapper.AttendanceRecordMapper
	wsMapper          *mapper.WorkScheduleMapper
	holidayMapper     *mapper.HolidayMapper
}

// NewAttendanceRecordUsecase creates a new AttendanceRecordUsecase
func NewAttendanceRecordUsecase(
	attendanceRepo repositories.AttendanceRecordRepository,
	workScheduleRepo repositories.WorkScheduleRepository,
	holidayRepo repositories.HolidayRepository,
	leaveRequestRepo repositories.LeaveRequestRepository,
	employeeRepo orgRepos.EmployeeRepository,
	divisionRepo orgRepos.DivisionRepository,
	overtimeRequestUC OvertimeRequestUsecase,
) AttendanceRecordUsecase {
	return &attendanceRecordUsecase{
		attendanceRepo:    attendanceRepo,
		workScheduleRepo:  workScheduleRepo,
		holidayRepo:       holidayRepo,
		leaveRequestRepo:  leaveRequestRepo,
		employeeRepo:      employeeRepo,
		divisionRepo:      divisionRepo,
		overtimeRequestUC: overtimeRequestUC,
		mapper:            mapper.NewAttendanceRecordMapper(),
		wsMapper:          mapper.NewWorkScheduleMapper(),
		holidayMapper:     mapper.NewHolidayMapper(),
	}
}

// WithAttendanceTodayPublisher attaches an optional realtime publisher for attendance today updates.
func WithAttendanceTodayPublisher(uc AttendanceRecordUsecase, publisher AttendanceTodayPublisher) AttendanceRecordUsecase {
	impl, ok := uc.(*attendanceRecordUsecase)
	if !ok {
		return uc
	}
	impl.todayPublisher = publisher
	return impl
}

func (u *attendanceRecordUsecase) List(ctx context.Context, req *dto.ListAttendanceRecordsRequest) ([]dto.AttendanceRecordResponse, *utils.PaginationResult, error) {
	records, total, err := u.attendanceRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := u.mapper.ToResponseList(records)

	// Enrich with employee data (names, codes, divisions)
	employeeIDs := make([]string, 0, len(records))
	for _, r := range records {
		employeeIDs = append(employeeIDs, r.EmployeeID)
	}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	u.mapper.EnrichResponseList(responses, employeeMap)

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

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *attendanceRecordUsecase) GetByID(ctx context.Context, id string) (*dto.AttendanceRecordResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.AttendanceRecord{}, id, security.HRDScopeQueryOptions()) {
		return nil, ErrAttendanceNotFound
	}

	ar, err := u.attendanceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, err
	}

	resp := u.mapper.ToResponse(ar, ar.EmployeeID)
	// Enrich with employee data
	employeeMap := u.buildEmployeeMap(ctx, []string{ar.EmployeeID})
	u.mapper.EnrichResponse(resp, employeeMap)

	// Enrich with work schedule name
	if ar.WorkScheduleID != "" {
		ws, wsErr := u.workScheduleRepo.FindByID(ctx, ar.WorkScheduleID)
		if wsErr == nil && ws != nil {
			resp.WorkScheduleName = ws.Name
		}
	}

	// Enrich with approver name
	if ar.ApprovedBy != nil && *ar.ApprovedBy != "" {
		approverMap := u.buildEmployeeMap(ctx, []string{*ar.ApprovedBy})
		if approver, ok := approverMap[*ar.ApprovedBy]; ok {
			resp.ApprovedByName = approver.Name
		}
	}

	return resp, nil
}

func (u *attendanceRecordUsecase) GetTodayAttendance(ctx context.Context, employeeID string) (*dto.TodayAttendanceResponse, error) {
	employeeID = u.resolveEmployeeID(ctx, employeeID)
	today := apptime.NowForEmployee(employeeID)

	// Get today's attendance record
	ar, err := u.attendanceRepo.FindByEmployeeAndDate(ctx, employeeID, today)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Get work schedule (try division-based, fallback to default)
	ws, err := u.getScheduleForEmployee(ctx, employeeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Check if today is a holiday (company-scoped)
	companyID := u.resolveCompanyID(ctx, employeeID)
	isHoliday, holiday, err := u.holidayRepo.IsHolidayForCompany(ctx, today, companyID)
	if err != nil {
		return nil, err
	}

	empLoc := apptime.LocationForEmployee(employeeID)
	return u.mapper.ToTodayResponse(ar, ws, isHoliday, holiday, u.wsMapper, u.holidayMapper, empLoc), nil
}

func (u *attendanceRecordUsecase) ClockIn(ctx context.Context, employeeID string, req *dto.ClockInRequest) (*dto.AttendanceRecordResponse, error) {
	employeeID = u.resolveEmployeeID(ctx, employeeID)
	today := apptime.NowForEmployee(employeeID)

	// Check if already checked in
	existing, err := u.attendanceRepo.FindByEmployeeAndDate(ctx, employeeID, today)
	if err == nil && existing.CheckInTime != nil {
		return nil, ErrAlreadyCheckedIn
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Get work schedule (try division-based, fallback to default)
	ws, err := u.getScheduleForEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Check if today is a working day
	if !ws.IsWorkingDay(int(today.Weekday())) {
		// Allow WFH or field work on non-working days
		if req.CheckInType == string(models.CheckInTypeNormal) {
			return nil, ErrNotWorkingDay
		}
	}

	// Check if holiday (company-scoped)
	companyID := u.resolveCompanyID(ctx, employeeID)
	isHoliday, _, err := u.holidayRepo.IsHolidayForCompany(ctx, today, companyID)
	if err != nil {
		return nil, err
	}
	if isHoliday && req.CheckInType == string(models.CheckInTypeNormal) {
		return nil, ErrHolidayNoCheckIn
	}

	// Validate check-in time is not before scheduled start time
	// Employee can only clock in at or after their scheduled start time
	empLoc := apptime.LocationForEmployee(employeeID)
	now := time.Now().In(empLoc)

	// Determine the earliest allowed check-in time
	var earliestCheckInTime string
	if ws.IsFlexible && ws.FlexibleStartTime != "" {
		earliestCheckInTime = ws.FlexibleStartTime
	} else {
		earliestCheckInTime = ws.StartTime
	}

	// Parse earliest check-in time
	earliestTime, err := time.Parse("15:04", earliestCheckInTime)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule start time: %w", err)
	}

	earliestCheckInToday := time.Date(today.Year(), today.Month(), today.Day(),
		earliestTime.Hour(), earliestTime.Minute(), 0, 0, empLoc)

	// Check if current time is before earliest allowed check-in time
	if now.Before(earliestCheckInToday) {
		return nil, fmt.Errorf("TOO_EARLY_TO_CHECK_IN: Cannot check in before %s. Your scheduled start time is %s.", earliestCheckInTime, ws.StartTime)
	}

	// GPS validation
	if ws.RequireGPS && req.CheckInType == string(models.CheckInTypeNormal) {
		if req.Latitude == nil || req.Longitude == nil {
			return nil, ErrGPSRequired
		}

		// Calculate distance from office
		distance := u.calculateDistance(ws.OfficeLatitude, ws.OfficeLongitude, *req.Latitude, *req.Longitude)
		if distance > ws.GPSRadiusMeter {
			return nil, ErrOutsideGPSRadius
		}
	}

	// Calculate late minutes
	lateMinutes := 0

	// Parse schedule start time
	scheduleStart, _ := time.Parse("15:04", ws.StartTime)
	scheduleStartToday := time.Date(today.Year(), today.Month(), today.Day(),
		scheduleStart.Hour(), scheduleStart.Minute(), 0, 0, empLoc)

	// Add tolerance
	scheduleStartToday = scheduleStartToday.Add(time.Duration(ws.LateToleranceMinutes) * time.Minute)

	if now.After(scheduleStartToday) {
		lateMinutes = int(now.Sub(scheduleStartToday).Minutes())
	}

	// Determine status
	status := models.AttendanceStatusPresent
	if lateMinutes > 0 {
		status = models.AttendanceStatusLate
	}
	if req.CheckInType == string(models.CheckInTypeWFH) {
		status = models.AttendanceStatusWFH
	}

	// Validate late reason required for late NORMAL clock-in
	if lateMinutes > 0 && req.CheckInType == string(models.CheckInTypeNormal) {
		if len(req.LateReason) == 0 {
			return nil, ErrLateReasonRequired
		}
	}

	// Validate photo required for WFH / FIELD_WORK
	if req.CheckInType == string(models.CheckInTypeWFH) || req.CheckInType == string(models.CheckInTypeFieldWork) {
		if len(req.PhotoURL) == 0 {
			return nil, ErrPhotoRequired
		}
	}

	// Create or update attendance record
	ar := existing
	if ar == nil {
		ar = &models.AttendanceRecord{
			EmployeeID:     employeeID,
			Date:           today,
			WorkScheduleID: ws.ID,
		}
	}

	ar.CheckInTime = &now
	ar.CheckInType = models.CheckInType(req.CheckInType)
	ar.CheckInLatitude = req.Latitude
	ar.CheckInLongitude = req.Longitude
	ar.CheckInAddress = req.Address
	ar.CheckInNote = req.Note
	ar.LateMinutes = lateMinutes
	ar.LateReason = req.LateReason
	ar.PhotoURL = req.PhotoURL
	ar.Status = status

	if existing == nil {
		err = u.attendanceRepo.Create(ctx, ar)
	} else {
		err = u.attendanceRepo.Update(ctx, ar)
	}

	if err != nil {
		return nil, err
	}

	u.publishTodayUpdate(ctx, employeeID, "clock_in", ar, false)

	resp := u.mapper.ToResponse(ar, ar.EmployeeID)
	return resp, nil
}

func (u *attendanceRecordUsecase) ClockOut(ctx context.Context, employeeID string, req *dto.ClockOutRequest) (*dto.AttendanceRecordResponse, error) {
	employeeID = u.resolveEmployeeID(ctx, employeeID)
	today := apptime.NowForEmployee(employeeID)

	// Get today's attendance
	ar, err := u.attendanceRepo.FindByEmployeeAndDate(ctx, employeeID, today)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotCheckedIn
		}
		return nil, err
	}

	if ar.CheckInTime == nil {
		return nil, ErrNotCheckedIn
	}

	if ar.CheckOutTime != nil {
		return nil, ErrAlreadyCheckedOut
	}

	// Check if today is a holiday (block clock-out on holidays for normal check-in)
	if ar.CheckInType == models.CheckInTypeNormal {
		companyID := u.resolveCompanyID(ctx, employeeID)
		isHoliday, _, err := u.holidayRepo.IsHolidayForCompany(ctx, today, companyID)
		if err != nil {
			return nil, err
		}
		if isHoliday {
			return nil, ErrHolidayNoCheckOut
		}

		// Check if today is a working day via schedule
		ws, _ := u.getScheduleForEmployee(ctx, employeeID)
		if ws != nil && !ws.IsWorkingDay(int(today.Weekday())) {
			return nil, ErrOffDayNoCheckOut
		}
	}

	// Get work schedule
	ws, _ := u.workScheduleRepo.FindByID(ctx, ar.WorkScheduleID)

	empLoc := apptime.LocationForEmployee(employeeID)
	now := time.Now().In(empLoc)
	ar.CheckOutTime = &now
	ar.CheckOutLatitude = req.Latitude
	ar.CheckOutLongitude = req.Longitude
	ar.CheckOutAddress = req.Address
	ar.CheckOutNote = req.Note

	// Calculate working minutes
	ar.CalculateWorkingMinutes()

	var scheduleEndToday time.Time

	// Calculate early leave minutes
	if ws != nil {
		scheduleEnd, _ := time.Parse("15:04", ws.EndTime)
		scheduleEndToday = time.Date(today.Year(), today.Month(), today.Day(),
			scheduleEnd.Hour(), scheduleEnd.Minute(), 0, 0, empLoc)

		// Subtract tolerance
		scheduleEndToday = scheduleEndToday.Add(-time.Duration(ws.EarlyLeaveToleranceMinutes) * time.Minute)

		if now.Before(scheduleEndToday) {
			ar.EarlyLeaveMinutes = int(scheduleEndToday.Sub(now).Minutes())
		}

		// Calculate overtime (if worked beyond schedule end time + 30 mins buffer)
		if now.After(scheduleEndToday.Add(30 * time.Minute)) {
			ar.OvertimeMinutes = int(now.Sub(scheduleEndToday).Minutes()) - 30
		}

		// Calculate total break duration from all breaks
		totalBreakMinutes := 0
		for _, breakTime := range ws.Breaks {
			// Parse break start and end times
			breakStartHour := int(breakTime.StartTime[0]-'0')*10 + int(breakTime.StartTime[1]-'0')
			breakStartMin := int(breakTime.StartTime[3]-'0')*10 + int(breakTime.StartTime[4]-'0')
			breakEndHour := int(breakTime.EndTime[0]-'0')*10 + int(breakTime.EndTime[1]-'0')
			breakEndMin := int(breakTime.EndTime[3]-'0')*10 + int(breakTime.EndTime[4]-'0')

			breakStartMinutes := breakStartHour*60 + breakStartMin
			breakEndMinutes := breakEndHour*60 + breakEndMin

			breakDuration := breakEndMinutes - breakStartMinutes
			if breakDuration > 0 {
				totalBreakMinutes += breakDuration
			}
		}

		// Subtract break duration from working minutes
		ar.BreakMinutes = totalBreakMinutes
		ar.WorkingMinutes -= ar.BreakMinutes
		if ar.WorkingMinutes < 0 {
			ar.WorkingMinutes = 0
		}
	}

	if err := u.attendanceRepo.Update(ctx, ar); err != nil {
		return nil, err
	}

	// Auto-create overtime request if overtime detected
	if ar.OvertimeMinutes > 0 && u.overtimeRequestUC != nil && !scheduleEndToday.IsZero() {
		overtimeStartTime := scheduleEndToday.Add(30 * time.Minute)
		overtimeEndTime := now

		_, _ = u.overtimeRequestUC.CreateAutoDetectedOvertime(
			ctx,
			ar.ID,
			employeeID,
			ar.OvertimeMinutes,
			today,
			overtimeStartTime,
			overtimeEndTime,
		)
	}

	u.publishTodayUpdate(ctx, employeeID, "clock_out", ar, false)

	resp := u.mapper.ToResponse(ar, employeeID)
	return resp, nil
}

func (u *attendanceRecordUsecase) CreateManualEntry(ctx context.Context, req *dto.ManualAttendanceRequest, createdBy string) (*dto.AttendanceRecordResponse, error) {
	// Validate check-in/check-out are provided for statuses that require them
	statusesRequiringTimes := map[string]bool{"PRESENT": true, "LATE": true, "HALF_DAY": true, "WFH": true}
	if statusesRequiringTimes[req.Status] {
		if req.CheckInTime == nil || *req.CheckInTime == "" {
			return nil, errors.New("check_in_time is required for status " + req.Status)
		}
		if req.CheckOutTime == nil || *req.CheckOutTime == "" {
			return nil, errors.New("check_out_time is required for status " + req.Status)
		}
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	// Check if record already exists
	existing, err := u.attendanceRepo.FindByEmployeeAndDate(ctx, req.EmployeeID, date)
	if err == nil && existing != nil {
		return nil, errors.New("attendance record already exists for this date")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Get work schedule
	ws, err := u.workScheduleRepo.FindDefault(ctx)
	if err != nil {
		return nil, err
	}

	ar := &models.AttendanceRecord{
		EmployeeID:        req.EmployeeID,
		Date:              date,
		CheckInType:       models.CheckInType(req.CheckInType),
		Status:            models.AttendanceStatus(req.Status),
		Notes:             req.Notes,
		IsManualEntry:     true,
		ManualEntryReason: req.Reason,
		ApprovedBy:        &createdBy,
		WorkScheduleID:    ws.ID,
	}

	// Parse times if provided
	if req.CheckInTime != nil {
		checkInTime, err := time.Parse("15:04", *req.CheckInTime)
		if err == nil {
			t := time.Date(date.Year(), date.Month(), date.Day(),
				checkInTime.Hour(), checkInTime.Minute(), 0, 0, date.Location())
			ar.CheckInTime = &t
		}
	}

	if req.CheckOutTime != nil {
		checkOutTime, err := time.Parse("15:04", *req.CheckOutTime)
		if err == nil {
			t := time.Date(date.Year(), date.Month(), date.Day(),
				checkOutTime.Hour(), checkOutTime.Minute(), 0, 0, date.Location())
			ar.CheckOutTime = &t
		}
	}

	// Calculate working minutes if both times provided
	ar.CalculateWorkingMinutes()

	if err := u.attendanceRepo.Create(ctx, ar); err != nil {
		return nil, err
	}

	u.publishTodayUpdateIfToday(ctx, req.EmployeeID, date, "manual_create", ar, false)

	resp := u.mapper.ToResponse(ar, req.EmployeeID)
	return resp, nil
}

func (u *attendanceRecordUsecase) Update(ctx context.Context, id string, req *dto.UpdateAttendanceRecordRequest) (*dto.AttendanceRecordResponse, error) {
	ar, err := u.attendanceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, err
	}

	// Apply updates
	if req.CheckInTime != nil {
		checkInTime, err := time.Parse("15:04", *req.CheckInTime)
		if err == nil {
			t := time.Date(ar.Date.Year(), ar.Date.Month(), ar.Date.Day(),
				checkInTime.Hour(), checkInTime.Minute(), 0, 0, ar.Date.Location())
			ar.CheckInTime = &t
		}
	}

	if req.CheckOutTime != nil {
		checkOutTime, err := time.Parse("15:04", *req.CheckOutTime)
		if err == nil {
			t := time.Date(ar.Date.Year(), ar.Date.Month(), ar.Date.Day(),
				checkOutTime.Hour(), checkOutTime.Minute(), 0, 0, ar.Date.Location())
			ar.CheckOutTime = &t
		}
	}

	if req.CheckInType != nil {
		ar.CheckInType = models.CheckInType(*req.CheckInType)
	}

	if req.Status != nil {
		ar.Status = models.AttendanceStatus(*req.Status)
	}

	if req.Notes != nil {
		ar.Notes = *req.Notes
	}

	if req.ManualEntryReason != nil {
		ar.ManualEntryReason = *req.ManualEntryReason
		ar.IsManualEntry = true
	}

	// Recalculate working minutes
	ar.CalculateWorkingMinutes()

	if err := u.attendanceRepo.Update(ctx, ar); err != nil {
		return nil, err
	}

	u.publishTodayUpdateIfToday(ctx, ar.EmployeeID, ar.Date, "manual_update", ar, false)

	resp := u.mapper.ToResponse(ar, ar.EmployeeID)
	return resp, nil
}

func (u *attendanceRecordUsecase) Delete(ctx context.Context, id string) error {
	ar, err := u.attendanceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttendanceNotFound
		}
		return err
	}

	if err := u.attendanceRepo.Delete(ctx, id); err != nil {
		return err
	}

	u.publishTodayUpdateIfToday(ctx, ar.EmployeeID, ar.Date, "delete", ar, true)

	return nil
}

func (u *attendanceRecordUsecase) GetMonthlyStats(ctx context.Context, req *dto.MonthlyReportRequest) ([]dto.MonthlyAttendanceStats, error) {
	// For now, just get stats for one employee
	// In production, this would handle division-level reports

	if req.EmployeeID == "" {
		return nil, errors.New("employee_id is required")
	}

	stats, err := u.attendanceRepo.GetEmployeeMonthlyStats(ctx, req.EmployeeID, req.Year, req.Month)
	if err != nil {
		return nil, err
	}

	// Calculate working days in month (using employee's timezone)
	empLoc := apptime.LocationForEmployee(req.EmployeeID)
	firstDay := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, empLoc)
	lastDay := firstDay.AddDate(0, 1, -1)
	workingDays := 0
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			workingDays++
		}
	}

	formattedStats := u.mapper.ToMonthlyStats(stats, workingDays)

	return []dto.MonthlyAttendanceStats{*formattedStats}, nil
}

// ListSelf returns the attendance history for the currently authenticated employee.
// It resolves the provided userID to the actual employee ID before querying.
func (u *attendanceRecordUsecase) ListSelf(ctx context.Context, req *dto.ListAttendanceRecordsRequest, userID string) ([]dto.AttendanceRecordResponse, *utils.PaginationResult, error) {
	// Resolve userID → employeeID so the DB filter matches the correct row.
	req.EmployeeID = u.resolveEmployeeID(ctx, userID)
	return u.List(ctx, req)
}

// GetSelfMonthlyStats returns monthly attendance statistics for the currently authenticated employee.
// It resolves the provided userID to the actual employee ID before querying.
func (u *attendanceRecordUsecase) GetSelfMonthlyStats(ctx context.Context, req *dto.MonthlyReportRequest, userID string) ([]dto.MonthlyAttendanceStats, error) {
	// Resolve userID → employeeID.
	req.EmployeeID = u.resolveEmployeeID(ctx, userID)
	return u.GetMonthlyStats(ctx, req)
}

// calculateDistance calculates distance between two GPS coordinates in meters using Haversine formula
func (u *attendanceRecordUsecase) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// resolveEmployeeID returns the actual employee ID for use in attendance records.
// Auth context may set only user_id; when the given id is a user_id (an employee exists with user_id = id),
// we return that employee's ID so that the attendance record stores the real employee_id and list enrichment finds the employee.
func (u *attendanceRecordUsecase) resolveEmployeeID(ctx context.Context, userIDOrEmployeeID string) string {
	emp, err := u.employeeRepo.FindByUserID(ctx, userIDOrEmployeeID)
	if err == nil && emp != nil {
		return emp.ID
	}
	return userIDOrEmployeeID
}

// resolveCompanyID returns the company ID for the given employee.
// Returns "" if the employee has no company assigned (falls back to global).
func (u *attendanceRecordUsecase) resolveCompanyID(ctx context.Context, employeeID string) string {
	emp, err := u.employeeRepo.FindByID(ctx, employeeID)
	if err != nil || emp == nil || emp.CompanyID == nil {
		return ""
	}
	return *emp.CompanyID
}

// buildEmployeeMap batch-fetches employees by IDs and builds a lookup map
func (u *attendanceRecordUsecase) buildEmployeeMap(ctx context.Context, ids []string) map[string]*orgModels.Employee {
	m := make(map[string]*orgModels.Employee)
	if len(ids) == 0 {
		return m
	}

	unique := make(map[string]bool)
	dedupIDs := make([]string, 0)
	for _, id := range ids {
		if !unique[id] {
			unique[id] = true
			dedupIDs = append(dedupIDs, id)
		}
	}

	employees, err := u.employeeRepo.FindByIDs(ctx, dedupIDs)
	if err != nil {
		return m
	}
	for i := range employees {
		m[employees[i].ID] = &employees[i]
	}
	// Fallback: some attendance records may have been stored with user_id in employee_id (e.g. before resolveEmployeeID existed).
	// Look up by user_id so list enrichment still shows employee name and division.
	for _, id := range dedupIDs {
		if _, ok := m[id]; ok {
			continue
		}
		emp, err := u.employeeRepo.FindByUserID(ctx, id)
		if err == nil && emp != nil {
			m[id] = emp
		}
	}
	return m
}

// getScheduleForEmployee returns the work schedule for an employee based on their division, falling back to default
func (u *attendanceRecordUsecase) getScheduleForEmployee(ctx context.Context, employeeID string) (*models.WorkSchedule, error) {
	// Try to find employee's division
	emp, err := u.employeeRepo.FindByID(ctx, employeeID)
	if err == nil && emp.DivisionID != nil && *emp.DivisionID != "" {
		ws, err := u.workScheduleRepo.FindByDivisionID(ctx, *emp.DivisionID)
		if err == nil {
			return ws, nil
		}
	}

	// Fallback to default schedule
	return u.workScheduleRepo.FindDefault(ctx)
}

// GetEmployeeSchedule returns the work schedule for a specific employee
func (u *attendanceRecordUsecase) GetEmployeeSchedule(ctx context.Context, employeeID string) (*dto.EmployeeScheduleResponse, error) {
	ws, err := u.getScheduleForEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("no work schedule found for employee: %w", err)
	}

	result := &dto.EmployeeScheduleResponse{
		ID:         ws.ID,
		Name:       ws.Name,
		StartTime:  ws.StartTime,
		EndTime:    ws.EndTime,
		IsFlexible: ws.IsFlexible,
	}
	if ws.FlexibleStartTime != "" {
		flexStart := ws.FlexibleStartTime
		result.FlexibleStartTime = &flexStart
	}
	if ws.FlexibleEndTime != "" {
		flexEnd := ws.FlexibleEndTime
		result.FlexibleEndTime = &flexEnd
	}

	return result, nil
}

// GetFormData returns form data for attendance management (manual entry)
func (u *attendanceRecordUsecase) GetFormData(ctx context.Context) (*dto.AttendanceFormDataResponse, error) {
	// Get active employees
	employees, err := listScopedActiveEmployees(ctx, database.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}
	employeeOptions := make([]dto.EmployeeFormOption, 0, len(employees))
	for _, emp := range employees {
		parsedID, err := uuid.Parse(emp.ID)
		if err != nil {
			continue
		}
		employeeOptions = append(employeeOptions, dto.EmployeeFormOption{
			ID:           parsedID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		})
	}

	// Get active divisions
	divListReq := &orgDTO.ListDivisionsRequest{Page: 1, PerPage: 100}
	divisions, _, err := u.divisionRepo.List(ctx, divListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch divisions: %w", err)
	}
	divisionOptions := make([]dto.DivisionFormOption, 0, len(divisions))
	for _, div := range divisions {
		if div.IsActive {
			divisionOptions = append(divisionOptions, dto.DivisionFormOption{
				ID:   div.ID,
				Name: div.Name,
			})
		}
	}

	// Get active schedules
	scheduleReq := &dto.ListWorkSchedulesRequest{Page: 1, PerPage: 100}
	schedules, _, err := u.workScheduleRepo.List(ctx, scheduleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedules: %w", err)
	}
	scheduleOptions := make([]dto.AttendanceScheduleFormOption, 0, len(schedules))
	for _, s := range schedules {
		if s.IsActive {
			scheduleOptions = append(scheduleOptions, dto.AttendanceScheduleFormOption{
				ID:   s.ID,
				Name: s.Name,
			})
		}
	}

	checkInTypes := []dto.AttendanceCheckInTypeFormOption{
		{Value: "NORMAL", Label: "Normal (Office)"},
		{Value: "WFH", Label: "Work From Home"},
		{Value: "FIELD_WORK", Label: "Field Work"},
	}

	statuses := []dto.AttendanceStatusFormOption{
		{Value: "PRESENT", Label: "Present"},
		{Value: "ABSENT", Label: "Absent"},
		{Value: "LATE", Label: "Late"},
		{Value: "HALF_DAY", Label: "Half Day"},
		{Value: "LEAVE", Label: "Leave"},
		{Value: "WFH", Label: "Work From Home"},
		{Value: "OFF_DAY", Label: "Off Day"},
		{Value: "HOLIDAY", Label: "Holiday"},
	}

	return &dto.AttendanceFormDataResponse{
		Employees:    employeeOptions,
		Divisions:    divisionOptions,
		Schedules:    scheduleOptions,
		CheckInTypes: checkInTypes,
		Statuses:     statuses,
	}, nil
}

// ProcessAutoAbsent creates ABSENT/LEAVE attendance records for employees who didn't clock in
// on the given date. It respects holidays, approved leaves, off-days, and existing records.
// When companyID is non-empty, only employees of that company are processed and holiday checks
// are company-scoped. When empty, all employees are processed with global holiday checks.
func (u *attendanceRecordUsecase) ProcessAutoAbsent(ctx context.Context, date time.Time, companyID string) (*dto.AutoAbsentResult, error) {
	empLoc := apptime.LocationForCompany(companyID)
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, empLoc)
	result := &dto.AutoAbsentResult{
		Date: dateOnly.Format("2006-01-02"),
	}

	// 1. Check if the date is a holiday — skip entirely (company-scoped)
	isHoliday, _, err := u.holidayRepo.IsHolidayForCompany(ctx, dateOnly, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check holiday: %w", err)
	}
	if isHoliday {
		result.HolidaySkipped = true
		return result, nil
	}

	// 2. Get all active employees
	employees, err := u.employeeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}
	result.TotalEmployees = len(employees)

	if len(employees) == 0 {
		return result, nil
	}

	// 3. Collect employee IDs for batch queries
	employeeIDs := make([]string, len(employees))
	employeeMap := make(map[string]*orgModels.Employee, len(employees))
	for i, emp := range employees {
		employeeIDs[i] = emp.ID
		e := emp // copy to avoid pointer reuse
		employeeMap[emp.ID] = &e
	}

	// 4. Batch: find which employees already have attendance records
	existingIDs, err := u.attendanceRepo.GetAbsentEmployeesForDate(ctx, dateOnly, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing records: %w", err)
	}
	// GetAbsentEmployeesForDate returns employees WITHOUT records, so we need the inverse
	// Actually, it returns IDs of absent employees (those NOT in attendance records).
	// We need to know who DOES have a record. Let's build a set of those who DON'T.
	absentSet := make(map[string]bool, len(existingIDs))
	for _, id := range existingIDs {
		absentSet[id] = true
	}

	// 5. Batch: find approved leaves for this date
	leaveMap, err := u.leaveRequestRepo.FindApprovedByDateForEmployees(ctx, dateOnly, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check leave requests: %w", err)
	}

	// 6. Cache work schedules by division to avoid repeated lookups
	scheduleCache := make(map[string]*models.WorkSchedule)
	var defaultSchedule *models.WorkSchedule

	getSchedule := func(emp *orgModels.Employee) *models.WorkSchedule {
		if emp.DivisionID != nil && *emp.DivisionID != "" {
			if ws, ok := scheduleCache[*emp.DivisionID]; ok {
				return ws
			}
			ws, err := u.workScheduleRepo.FindByDivisionID(ctx, *emp.DivisionID)
			if err == nil {
				scheduleCache[*emp.DivisionID] = ws
				return ws
			}
		}
		// Fallback to default
		if defaultSchedule == nil {
			ws, err := u.workScheduleRepo.FindDefault(ctx)
			if err == nil {
				defaultSchedule = ws
			}
		}
		return defaultSchedule
	}

	// 7. Process each employee
	for _, empID := range employeeIDs {
		// Skip if employee already has a record for this date
		if !absentSet[empID] {
			result.Skipped++
			continue
		}

		emp := employeeMap[empID]

		// Check work schedule — skip if not a working day
		ws := getSchedule(emp)
		if ws == nil {
			result.Skipped++
			continue
		}
		if !ws.IsWorkingDay(int(dateOnly.Weekday())) {
			result.Skipped++
			continue
		}

		// Check if employee has approved leave
		if lr, hasLeave := leaveMap[empID]; hasLeave {
			// Create LEAVE record
			leaveID := lr.ID
			ar := &models.AttendanceRecord{
				ID:             uuid.New().String(),
				EmployeeID:     empID,
				Date:           dateOnly,
				Status:         models.AttendanceStatusLeave,
				WorkScheduleID: ws.ID,
				LeaveRequestID: &leaveID,
				IsManualEntry:  true,
				Notes:          "Auto-generated: employee on approved leave",
			}
			if err := u.attendanceRepo.Create(ctx, ar); err != nil {
				result.Errors++
				continue
			}
			u.publishTodayUpdateIfToday(ctx, empID, dateOnly, "auto_absent_leave", ar, false)
			result.LeaveCreated++
			continue
		}

		// Create ABSENT record
		ar := &models.AttendanceRecord{
			ID:             uuid.New().String(),
			EmployeeID:     empID,
			Date:           dateOnly,
			Status:         models.AttendanceStatusAbsent,
			WorkScheduleID: ws.ID,
			IsManualEntry:  true,
			Notes:          "Auto-generated: no clock-in recorded",
		}
		if err := u.attendanceRepo.Create(ctx, ar); err != nil {
			result.Errors++
			continue
		}
		u.publishTodayUpdateIfToday(ctx, empID, dateOnly, "auto_absent", ar, false)
		result.AbsentCreated++
	}

	return result, nil
}

func (u *attendanceRecordUsecase) publishTodayUpdateIfToday(
	ctx context.Context,
	employeeID string,
	recordDate time.Time,
	trigger string,
	record *models.AttendanceRecord,
	deleted bool,
) {
	if !u.isTodayForEmployee(recordDate, employeeID) {
		return
	}
	u.publishTodayUpdate(ctx, employeeID, trigger, record, deleted)
}

func (u *attendanceRecordUsecase) publishTodayUpdate(
	ctx context.Context,
	employeeID string,
	trigger string,
	record *models.AttendanceRecord,
	deleted bool,
) {
	if u.todayPublisher == nil {
		return
	}

	tenantID := coreMiddleware.TenantFromContext(ctx)
	if tenantID == "" {
		return
	}

	payload := map[string]interface{}{
		"employee_id": employeeID,
		"trigger":     trigger,
		"deleted":     deleted,
	}

	if record != nil {
		payload["attendance_id"] = record.ID
		payload["date"] = record.Date.Format("2006-01-02")
	}

	u.todayPublisher.Publish(tenantID, "hrd.attendance.today_updated", payload)
}

func (u *attendanceRecordUsecase) isTodayForEmployee(recordDate time.Time, employeeID string) bool {
	loc := apptime.LocationForEmployee(employeeID)
	today := apptime.NowForEmployee(employeeID).In(loc)
	rec := recordDate.In(loc)

	return today.Year() == rec.Year() && today.YearDay() == rec.YearDay()
}
