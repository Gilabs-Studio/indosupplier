package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"gorm.io/gorm"
)

// LeaveRequestUsecase defines business logic for leave request operations
type LeaveRequestUsecase interface {
	// CRUD operations
	Create(ctx context.Context, req *dto.CreateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	GetByID(ctx context.Context, id string, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	List(ctx context.Context, filters *dto.LeaveRequestListFilterDTO, currentUserID string) ([]*dto.LeaveRequestResponseDTO, int64, error)
	Update(ctx context.Context, id string, req *dto.UpdateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	Delete(ctx context.Context, id string, currentUserID string) error

	// Form data for dropdowns
	GetFormData(ctx context.Context, currentUserID string) (*dto.FormDataResponseDTO, error)

	// Balance calculation
	CalculateBalance(ctx context.Context, employeeID string, currentUserID string) (*dto.LeaveBalanceResponseDTO, error)

	// Approval workflow
	Approve(ctx context.Context, id string, req *dto.ApproveLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	Reject(ctx context.Context, id string, req *dto.RejectLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	Cancel(ctx context.Context, id string, req *dto.CancelLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	Reapprove(ctx context.Context, id string, req *dto.ApproveLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)

	// Self-service operations (employee owns their requests)
	CreateSelf(ctx context.Context, req *dto.CreateMyLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	ListSelf(ctx context.Context, filters *dto.LeaveRequestListFilterDTO, currentUserID string) ([]*dto.LeaveRequestResponseDTO, int64, error)
	GetSelfByID(ctx context.Context, id string, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	UpdateSelf(ctx context.Context, id string, req *dto.UpdateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	CancelSelf(ctx context.Context, id string, req *dto.CancelLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error)
	GetSelfBalance(ctx context.Context, currentUserID string) (*dto.LeaveBalanceResponseDTO, error)
	GetSelfFormData(ctx context.Context, currentUserID string) (*dto.FormDataResponseDTO, error)

	// Audit trail
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.LeaveRequestAuditTrailEntry, int64, error)
}

type leaveRequestUsecase struct {
	db                   *gorm.DB
	leaveRequestRepo     repositories.LeaveRequestRepository
	attendanceRecordRepo repositories.AttendanceRecordRepository
	employeeRepo         orgRepos.EmployeeRepository
	leaveTypeRepo        coreRepos.LeaveTypeRepository
	holidayRepo          repositories.HolidayRepository
	mapper               *mapper.LeaveRequestMapper
}

type leaveRequestAuditRow struct {
	ID             string    `gorm:"column:id"`
	ActorID        string    `gorm:"column:actor_id"`
	PermissionCode string    `gorm:"column:permission_code"`
	TargetID       string    `gorm:"column:target_id"`
	Action         string    `gorm:"column:action"`
	Metadata       string    `gorm:"column:metadata"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ActorEmail     *string   `gorm:"column:actor_email"`
	ActorName      *string   `gorm:"column:actor_name"`
}

// NewLeaveRequestUsecase creates a new instance of LeaveRequestUsecase
func NewLeaveRequestUsecase(
	db *gorm.DB,
	leaveRequestRepo repositories.LeaveRequestRepository,
	employeeRepo orgRepos.EmployeeRepository,
	leaveTypeRepo coreRepos.LeaveTypeRepository,
	holidayRepo repositories.HolidayRepository,
	attendanceRecordRepo repositories.AttendanceRecordRepository,
) LeaveRequestUsecase {
	return &leaveRequestUsecase{
		db:                   db,
		leaveRequestRepo:     leaveRequestRepo,
		attendanceRecordRepo: attendanceRecordRepo,
		employeeRepo:         employeeRepo,
		leaveTypeRepo:        leaveTypeRepo,
		holidayRepo:          holidayRepo,
		mapper:               mapper.NewLeaveRequestMapper(),
	}
}

func (u *leaveRequestUsecase) getCurrentEmployee(ctx context.Context, currentUserID string) (*orgModels.Employee, error) {
	employee, err := u.employeeRepo.FindByUserID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("FORBIDDEN: employee profile not found for current user")
	}
	return employee, nil
}

func (u *leaveRequestUsecase) ensureOwnership(ctx context.Context, leaveRequestID, currentUserID string) (*models.LeaveRequest, error) {
	employee, err := u.getCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	leaveRequest, err := u.leaveRequestRepo.FindByID(ctx, leaveRequestID)
	if err != nil {
		return nil, err
	}

	if leaveRequest.EmployeeID != employee.ID {
		return nil, fmt.Errorf("FORBIDDEN: you can only access your own leave requests")
	}

	return leaveRequest, nil
}

func (u *leaveRequestUsecase) CreateSelf(ctx context.Context, req *dto.CreateMyLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	employee, err := u.getCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	createReq := &dto.CreateLeaveRequestDTO{
		EmployeeID:    employee.ID,
		LeaveTypeID:   req.LeaveTypeID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Duration:      req.Duration,
		Reason:        req.Reason,
		AttachmentURL: req.AttachmentURL,
	}

	return u.Create(ctx, createReq, currentUserID)
}

func (u *leaveRequestUsecase) ListSelf(ctx context.Context, filters *dto.LeaveRequestListFilterDTO, currentUserID string) ([]*dto.LeaveRequestResponseDTO, int64, error) {
	employee, err := u.getCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, 0, err
	}

	if filters == nil {
		filters = &dto.LeaveRequestListFilterDTO{}
	}
	filters.EmployeeID = &employee.ID

	return u.List(ctx, filters, currentUserID)
}

func (u *leaveRequestUsecase) GetSelfByID(ctx context.Context, id string, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	if _, err := u.ensureOwnership(ctx, id, currentUserID); err != nil {
		return nil, err
	}
	return u.GetByID(ctx, id, currentUserID)
}

func (u *leaveRequestUsecase) UpdateSelf(ctx context.Context, id string, req *dto.UpdateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	if _, err := u.ensureOwnership(ctx, id, currentUserID); err != nil {
		return nil, err
	}
	return u.Update(ctx, id, req, currentUserID)
}

func (u *leaveRequestUsecase) CancelSelf(ctx context.Context, id string, req *dto.CancelLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	if _, err := u.ensureOwnership(ctx, id, currentUserID); err != nil {
		return nil, err
	}
	return u.Cancel(ctx, id, req, currentUserID)
}

func (u *leaveRequestUsecase) GetSelfBalance(ctx context.Context, currentUserID string) (*dto.LeaveBalanceResponseDTO, error) {
	employee, err := u.getCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	return u.CalculateBalance(ctx, employee.ID, currentUserID)
}

func (u *leaveRequestUsecase) GetSelfFormData(ctx context.Context, currentUserID string) (*dto.FormDataResponseDTO, error) {
	employee, err := u.getCurrentEmployee(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	leaveTypes, err := u.leaveTypeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leave types: %w", err)
	}

	usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, employee.ID, true)
	if err != nil {
		usedDays = 0
	}

	remainingBalance := float64(employee.TotalLeaveQuota - usedDays)
	if remainingBalance < 0 {
		remainingBalance = 0
	}

	formLeaveTypes := make([]dto.FormLeaveTypeDTO, 0, len(leaveTypes))
	for _, lt := range leaveTypes {
		formLeaveTypes = append(formLeaveTypes, dto.FormLeaveTypeDTO{
			ID:      lt.ID,
			Name:    lt.Name,
			Code:    lt.Code,
			MaxDays: lt.MaxDays,
		})
	}

	return &dto.FormDataResponseDTO{
		Employees: []dto.FormEmployeeDTO{
			{
				ID:               employee.ID,
				Name:             employee.Name,
				EmployeeCode:     employee.EmployeeCode,
				RemainingBalance: remainingBalance,
			},
		},
		LeaveTypes: formLeaveTypes,
	}, nil
}

// validateLeaveDates validates that leave dates are not in the past
// WHY: Employees can only request leave for today or future dates
func (u *leaveRequestUsecase) validateLeaveDates(startDate, endDate time.Time) error {
	// Get current date (truncate time to compare dates only)
	now := apptime.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Normalize input dates (truncate time)
	startDateNormalized := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDateNormalized := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	if startDateNormalized.Before(today) {
		return fmt.Errorf("INVALID_START_DATE: start_date cannot be in the past")
	}

	if endDateNormalized.Before(today) {
		return fmt.Errorf("INVALID_END_DATE: end_date cannot be in the past")
	}

	return nil
}

// Create creates a new leave request
// WHY: Only approvers/HR can create leave requests for employees (business rule change)
// WHY: Validates balance, calculates working days, prevents overlapping requests
func (u *leaveRequestUsecase) Create(ctx context.Context, req *dto.CreateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	// 1. Fetch employee to validate and get quota
	employee, err := u.employeeRepo.FindByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	// 2. Permission enforcement via router middleware (leave.create)
	// WHY: Business rule - employees cannot create their own leave requests, only HR/approvers can
	// Permission is validated at router level using middleware.RequirePermission("leave.create")

	// 3. Parse dates for validation
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("INVALID_DATE_FORMAT: start_date must be YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("INVALID_DATE_FORMAT: end_date must be YYYY-MM-DD")
	}

	// 4. Validate dates are not in the past
	if err := u.validateLeaveDates(startDate, endDate); err != nil {
		return nil, err
	}

	// 5. Calculate total days based on duration
	totalDays, err := u.calculateTotalDays(ctx, req.Duration, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. Check for overlapping requests
	// WHY: Prevent employee from having multiple leave requests for the same period
	overlapping, err := u.leaveRequestRepo.FindOverlappingRequests(ctx, req.EmployeeID, startDate, endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check overlapping requests: %w", err)
	}
	if len(overlapping) > 0 {
		return nil, fmt.Errorf("OVERLAPPING_LEAVE_REQUEST: Employee already has a leave request for these dates")
	}

	// 6. Fetch leave type to check if it cuts annual leave
	leaveType, err := u.leaveTypeRepo.FindByID(ctx, req.LeaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leave type: %w", err)
	}

	// 7. Validate balance if leave type cuts annual leave
	// WHY: Prevent submission if employee has insufficient leave balance
	if leaveType.IsCutAnnualLeave {
		usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, req.EmployeeID, true)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate used leave days: %w", err)
		}

		remainingBalance := employee.TotalLeaveQuota - usedDays
		if int(totalDays) > remainingBalance {
			return nil, fmt.Errorf("INSUFFICIENT_LEAVE_BALANCE: requested %.1f days but only %d days available", totalDays, remainingBalance)
		}
	}

	// 8. Convert DTO to model
	leaveRequest, err := u.mapper.ToModel(req, totalDays, &currentUserID)
	if err != nil {
		return nil, err
	}

	// 9. Save to database
	if err := u.leaveRequestRepo.Create(ctx, leaveRequest); err != nil {
		return nil, fmt.Errorf("failed to create leave request: %w", err)
	}

	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, u.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "leave_request.approve",
		EntityType:     "leave_request",
		EntityID:       leaveRequest.ID,
		Title:          "Leave Request Approval",
		Message:        "A leave request has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create leave request notification: %v", err)
	}

	// 10. Return response DTO with full details
	return u.mapper.ToDetailResponseDTO(leaveRequest, employee, leaveType), nil
}

// GetByID retrieves a leave request by ID
// WHY: Only approvers/HR can view leave request details (business rule change)
func (u *leaveRequestUsecase) GetByID(ctx context.Context, id string, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.LeaveRequest{}, id, security.HRDScopeQueryOptions()) {
		return nil, fmt.Errorf("leave request not found")
	}

	leaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Permission enforcement via router middleware (leave.read)
	// WHY: Business rule - employees cannot view leave request details, only HR/approvers can
	// Permission is validated at router level using middleware.RequirePermission("leave.read")

	// Fetch employee for detailed response
	employee, err := u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	// Fetch leave type for detailed response
	leaveType, err := u.leaveTypeRepo.FindByID(ctx, leaveRequest.LeaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leave type: %w", err)
	}

	return u.mapper.ToDetailResponseDTO(leaveRequest, employee, leaveType), nil
}

// List retrieves leave requests with filters and pagination
// WHY: Only approvers/HR can list leave requests (business rule change)
func (u *leaveRequestUsecase) List(ctx context.Context, filters *dto.LeaveRequestListFilterDTO, currentUserID string) ([]*dto.LeaveRequestResponseDTO, int64, error) {
	// Permission enforcement via router middleware (leave.read)
	// WHY: Business rule - employees cannot view any leave requests, only HR/approvers can
	// Permission is validated at router level using middleware.RequirePermission("leave.read")

	// Set default pagination
	// Set default pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}

	perPage := filters.PerPage
	if perPage < 1 {
		perPage = 20
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if filters.StartDate != nil {
		parsed, err := time.Parse("2006-01-02", *filters.StartDate)
		if err != nil {
			return nil, 0, fmt.Errorf("INVALID_DATE_FORMAT: start_date must be YYYY-MM-DD")
		}
		startDate = &parsed
	}

	if filters.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *filters.EndDate)
		if err != nil {
			return nil, 0, fmt.Errorf("INVALID_DATE_FORMAT: end_date must be YYYY-MM-DD")
		}
		endDate = &parsed
	}

	// Parse optional status filter (case-insensitive)
	var status *models.LeaveStatus
	if filters.Status != nil {
		// Convert to uppercase for case-insensitive matching
		statusUpper := strings.ToUpper(*filters.Status)

		// Validate status value
		validStatuses := map[string]models.LeaveStatus{
			"PENDING":   models.LeaveStatusPending,
			"APPROVED":  models.LeaveStatusApproved,
			"REJECTED":  models.LeaveStatusRejected,
			"CANCELLED": models.LeaveStatusCancelled,
		}

		if validStatus, ok := validStatuses[statusUpper]; ok {
			status = &validStatus
		} else {
			return nil, 0, fmt.Errorf("INVALID_STATUS: status must be one of: PENDING, APPROVED, REJECTED, CANCELLED (case-insensitive)")
		}
	}

	// Fetch from repository with all filters including search
	leaveRequests, total, err := u.leaveRequestRepo.List(ctx, filters.EmployeeID, status, startDate, endDate, filters.Search, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch leave requests: %w", err)
	}

	// Extract unique employee IDs and leave type IDs for batch fetching
	employeeIDs := make([]string, 0)
	leaveTypeIDs := make([]string, 0)
	employeeIDMap := make(map[string]bool)
	leaveTypeIDMap := make(map[string]bool)

	for _, req := range leaveRequests {
		if !employeeIDMap[req.EmployeeID] {
			employeeIDs = append(employeeIDs, req.EmployeeID)
			employeeIDMap[req.EmployeeID] = true
		}
		if !leaveTypeIDMap[req.LeaveTypeID] {
			leaveTypeIDs = append(leaveTypeIDs, req.LeaveTypeID)
			leaveTypeIDMap[req.LeaveTypeID] = true
		}
	}

	// Batch fetch employees and leave types
	employees, err := u.employeeRepo.FindByIDs(ctx, employeeIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch employees: %w", err)
	}

	leaveTypes, err := u.leaveTypeRepo.FindByIDs(ctx, leaveTypeIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch leave types: %w", err)
	}

	// Create maps for O(1) lookup
	employeeMap := make(map[string]*orgModels.Employee)
	for i := range employees {
		employeeMap[employees[i].ID] = &employees[i]
	}

	leaveTypeMap := make(map[string]*coreModels.LeaveType)
	for i := range leaveTypes {
		leaveTypeMap[leaveTypes[i].ID] = &leaveTypes[i]
	}

	return u.mapper.ToList(leaveRequests, employeeMap, leaveTypeMap), total, nil
}

// Update updates an existing leave request
// WHY: Only approvers/HR can update leave requests (business rule change)
func (u *leaveRequestUsecase) Update(ctx context.Context, id string, req *dto.UpdateLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	// 1. Permission enforcement via router middleware (leave.update)
	// WHY: Business rule - employees cannot update leave requests, only HR/approvers can
	// Permission is validated at router level using middleware.RequirePermission("leave.update")

	// 2. Fetch existing leave request
	leaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Fetch employee for detailed response
	employee, err := u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	// 4. Check if leave request is editable
	// WHY: Only PENDING or REJECTED leaves can be edited
	if !leaveRequest.IsEditable() {
		return nil, fmt.Errorf("INVALID_STATUS: only PENDING or REJECTED leave requests can be edited")
	}

	// 5. Recalculate total days if dates or duration changed
	var totalDays *float64
	if req.StartDate != nil || req.EndDate != nil || req.Duration != nil {
		// Parse dates
		startDate := leaveRequest.StartDate
		if req.StartDate != nil {
			parsed, err := time.Parse("2006-01-02", *req.StartDate)
			if err != nil {
				return nil, fmt.Errorf("INVALID_DATE_FORMAT: start_date must be YYYY-MM-DD")
			}
			startDate = parsed
		}

		endDate := leaveRequest.EndDate
		if req.EndDate != nil {
			parsed, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("INVALID_DATE_FORMAT: end_date must be YYYY-MM-DD")
			}
			endDate = parsed
		}

		// Validate dates are not in the past (only if dates are being changed)
		if req.StartDate != nil || req.EndDate != nil {
			if err := u.validateLeaveDates(startDate, endDate); err != nil {
				return nil, err
			}
		}

		duration := string(leaveRequest.Duration)
		if req.Duration != nil {
			duration = *req.Duration
		}

		calculated, err := u.calculateTotalDays(ctx, duration, startDate, endDate)
		if err != nil {
			return nil, err
		}
		totalDays = &calculated

		// Check for overlapping requests (excluding current one)
		overlapping, err := u.leaveRequestRepo.FindOverlappingRequests(ctx, leaveRequest.EmployeeID, startDate, endDate, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to check overlapping requests: %w", err)
		}
		if len(overlapping) > 0 {
			return nil, fmt.Errorf("OVERLAPPING_LEAVE_REQUEST: dates overlap with another leave request")
		}

		// Revalidate balance if leave type cuts annual leave
		if req.LeaveTypeID != nil {
			leaveType, err := u.leaveTypeRepo.FindByID(ctx, *req.LeaveTypeID)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch leave type: %w", err)
			}

			if leaveType.IsCutAnnualLeave {
				usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, leaveRequest.EmployeeID, true)
				if err != nil {
					return nil, fmt.Errorf("failed to calculate used leave days: %w", err)
				}

				remainingBalance := employee.TotalLeaveQuota - usedDays
				if int(*totalDays) > remainingBalance {
					return nil, fmt.Errorf("INSUFFICIENT_LEAVE_BALANCE: requested %.1f days but only %d days available", *totalDays, remainingBalance)
				}
			}
		}
	}

	// 6. Apply updates to model
	if err := u.mapper.ApplyUpdateDTO(leaveRequest, req, totalDays, &currentUserID); err != nil {
		return nil, err
	}

	// 7. Save changes
	if err := u.leaveRequestRepo.Update(ctx, leaveRequest); err != nil {
		return nil, fmt.Errorf("failed to update leave request: %w", err)
	}

	// 8. Fetch updated leave type for response
	leaveType, err := u.leaveTypeRepo.FindByID(ctx, leaveRequest.LeaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leave type: %w", err)
	}

	return u.mapper.ToDetailResponseDTO(leaveRequest, employee, leaveType), nil
}

// Delete soft deletes a leave request
// WHY: Only approvers/HR can delete leave requests (business rule change)
func (u *leaveRequestUsecase) Delete(ctx context.Context, id string, currentUserID string) error {
	// 1. Permission enforcement via router middleware (leave.delete)
	// WHY: Business rule - employees cannot delete leave requests, only HR/approvers can
	// Permission is validated at router level using middleware.RequirePermission("leave.delete")

	// 2. Fetch leave request
	leaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 3. Check if deletable (only PENDING or REJECTED)
	if !leaveRequest.IsEditable() {
		return fmt.Errorf("INVALID_STATUS: only PENDING or REJECTED leave requests can be deleted")
	}

	// 4. Soft delete
	return u.leaveRequestRepo.Delete(ctx, id)
}

// CalculateBalance calculates the leave balance for an employee
func (u *leaveRequestUsecase) CalculateBalance(ctx context.Context, employeeID string, currentUserID string) (*dto.LeaveBalanceResponseDTO, error) {
	// 1. Fetch employee
	employee, err := u.employeeRepo.FindByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	// 2. IDOR check: Only employee themselves or approvers can view balance
	if employee.UserID == nil || *employee.UserID != currentUserID {
		// Note: For HR/approvers with leave.read permission, this check should be bypassed
		// This is a legacy check that should be refactored to use permission middleware
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this employee's balance")
	}

	// 3. Calculate used days (only leave types that cut annual leave)
	usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, employeeID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate used leave days: %w", err)
	}

	// 4. Calculate pending days
	pendingDays := 0
	pendingRequests, _, err := u.leaveRequestRepo.List(ctx, &employeeID, &[]models.LeaveStatus{models.LeaveStatusPending}[0], nil, nil, nil, 1, 100)
	if err == nil {
		for _, req := range pendingRequests {
			pendingDays += int(req.TotalDays)
		}
	}

	// 5. Get carry-over balance (if any)
	// TODO: Implement carry-over logic from previous year
	carryOverBalance := 0.0
	var carryOverExpiry *time.Time

	// 6. Build response
	return u.mapper.ToBalanceDTO(
		employeeID,
		employee.TotalLeaveQuota,
		usedDays,
		pendingDays,
		carryOverBalance,
		carryOverExpiry,
	), nil
}

// calculateTotalDays calculates the total days for a leave request
// WHY: Returns inclusive calendar days (2-3 = 2 days, includes weekends)
// For HALF_DAY: 0.5, FULL_DAY: 1.0, MULTI_DAY: end - start + 1
func (u *leaveRequestUsecase) calculateTotalDays(ctx context.Context, duration string, startDate, endDate time.Time) (float64, error) {
	switch duration {
	case "HALF_DAY":
		return 0.5, nil
	case "FULL_DAY":
		return 1.0, nil
	case "MULTI_DAY":
		// Calculate inclusive calendar days (including weekends)
		days := endDate.Sub(startDate).Hours() / 24
		return days + 1, nil // +1 for inclusive count
	default:
		return 0, fmt.Errorf("INVALID_DURATION: must be FULL_DAY, HALF_DAY, or MULTI_DAY")
	}
}

// Approve approves a leave request
// WHY: Uses row-level locking (FOR UPDATE) to prevent race conditions
func (u *leaveRequestUsecase) Approve(ctx context.Context, id string, req *dto.ApproveLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	var employee *orgModels.Employee
	var leaveType *coreModels.LeaveType

	// Use UpdateWithLock to ensure row-level locking (FOR UPDATE)
	// WHY: Prevent concurrent approvals causing double-deduction of leave balance
	err := u.leaveRequestRepo.UpdateWithLock(ctx, id, func(leaveRequest *models.LeaveRequest) error {
		// 1. Check if leave request can be approved
		if !leaveRequest.CanBeApproved() {
			return fmt.Errorf("INVALID_STATUS: leave request cannot be approved (current status: %s)", leaveRequest.Status)
		}

		// 2. Permission enforcement via router middleware (leave.approve)
		// Router ensures only users with leave.approve permission can access this endpoint

		// 3. Fetch employee and leave type to revalidate balance
		var err error
		employee, err = u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
		if err != nil {
			return fmt.Errorf("failed to fetch employee: %w", err)
		}

		leaveType, err = u.leaveTypeRepo.FindByID(ctx, leaveRequest.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to fetch leave type: %w", err)
		}

		// 4. Revalidate balance if leave type cuts annual leave
		// WHY: Balance might have changed between submission and approval
		if leaveType.IsCutAnnualLeave {
			usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, leaveRequest.EmployeeID, true)
			if err != nil {
				return fmt.Errorf("failed to calculate used leave days: %w", err)
			}

			remainingBalance := employee.TotalLeaveQuota - usedDays
			if int(leaveRequest.TotalDays) > remainingBalance {
				return fmt.Errorf("INSUFFICIENT_LEAVE_BALANCE: employee has only %d days remaining", remainingBalance)
			}
		}

		// 5. Update status to APPROVED
		leaveRequest.Status = models.LeaveStatusApproved
		now := apptime.Now()
		leaveRequest.ApprovedAt = &now

		// Use approver ID from DTO if provided, otherwise current user
		if req.ApprovedBy != nil {
			leaveRequest.ApprovedBy = req.ApprovedBy
		} else {
			leaveRequest.ApprovedBy = &currentUserID
		}

		leaveRequest.UpdatedBy = &currentUserID

		// 6. TODO: Trigger notification to employee
		// notificationService.Send(employee.Email, "Leave Request Approved", ...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. Fetch updated leave request to return
	updatedLeaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated leave request: %w", err)
	}

	// 8. Create LEAVE attendance records for the approved leave period
	if err := u.createLeaveAttendanceRecords(ctx, updatedLeaveRequest); err != nil {
		// Log error but don't fail the approval
		fmt.Printf("WARNING: failed to create attendance records for leave %s: %v\n", id, err)
	}

	return u.mapper.ToDetailResponseDTO(updatedLeaveRequest, employee, leaveType), nil
}

// Reject rejects a leave request
// WHY: Uses row-level locking (FOR UPDATE) to prevent concurrent updates
func (u *leaveRequestUsecase) Reject(ctx context.Context, id string, req *dto.RejectLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	var employee *orgModels.Employee
	var leaveType *coreModels.LeaveType

	// Use UpdateWithLock to ensure row-level locking (FOR UPDATE)
	err := u.leaveRequestRepo.UpdateWithLock(ctx, id, func(leaveRequest *models.LeaveRequest) error {
		// 1. Check if leave request can be approved (same check for rejection)
		if !leaveRequest.CanBeApproved() {
			return fmt.Errorf("INVALID_STATUS: leave request cannot be rejected (current status: %s)", leaveRequest.Status)
		}

		// 2. Permission enforcement via router middleware (leave.approve)
		// Router ensures only users with leave.approve permission can access this endpoint

		// 3. Validate rejection note
		if req.RejectionNote == "" {
			return fmt.Errorf("VALIDATION_ERROR: rejection_note is required")
		}

		// 4. Fetch employee and leave type for response
		var err error
		employee, err = u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
		if err != nil {
			return fmt.Errorf("failed to fetch employee: %w", err)
		}

		leaveType, err = u.leaveTypeRepo.FindByID(ctx, leaveRequest.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to fetch leave type: %w", err)
		}

		// 5. Update status to REJECTED
		leaveRequest.Status = models.LeaveStatusRejected
		leaveRequest.RejectionNote = &req.RejectionNote

		// Use rejecter ID from DTO if provided, otherwise current user
		if req.RejectedBy != nil {
			leaveRequest.RejectedBy = req.RejectedBy
		} else {
			leaveRequest.RejectedBy = &currentUserID
		}

		leaveRequest.UpdatedBy = &currentUserID

		// 6. TODO: Trigger notification to employee
		// notificationService.Send(employee.Email, "Leave Request Rejected", ...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. Fetch updated leave request to return
	updatedLeaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated leave request: %w", err)
	}

	return u.mapper.ToDetailResponseDTO(updatedLeaveRequest, employee, leaveType), nil
}

// Cancel cancels a leave request
// WHY: Only approvers/HR can cancel leave requests - uses row-level locking to prevent race conditions
func (u *leaveRequestUsecase) Cancel(ctx context.Context, id string, req *dto.CancelLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	var employee *orgModels.Employee
	var leaveType *coreModels.LeaveType
	wasApproved := false

	// Use UpdateWithLock to ensure row-level locking (FOR UPDATE)
	err := u.leaveRequestRepo.UpdateWithLock(ctx, id, func(leaveRequest *models.LeaveRequest) error {
		// 1. Check if leave request can be cancelled
		// WHY: PENDING and APPROVED statuses can be cancelled
		if leaveRequest.Status != models.LeaveStatusApproved && leaveRequest.Status != models.LeaveStatusPending {
			return fmt.Errorf("INVALID_STATUS: only PENDING or APPROVED leave requests can be cancelled (current status: %s)", leaveRequest.Status)
		}

		// 1b. Check if cancel is performed before leave start date (APPROVED only)
		// WHY: PENDING requests can always be cancelled regardless of date.
		//      APPROVED requests can only be cancelled before the leave period starts.
		if leaveRequest.Status == models.LeaveStatusApproved {
			today := apptime.Now().Truncate(24 * time.Hour)
			startDate := leaveRequest.StartDate.Truncate(24 * time.Hour)
			if !today.Before(startDate) {
				return fmt.Errorf("INVALID_DATE: leave request can only be cancelled before the start date (start date: %s, today: %s)", startDate.Format("2006-01-02"), today.Format("2006-01-02"))
			}
		}

		// Track original status for conditional attendance record deletion
		wasApproved = leaveRequest.Status == models.LeaveStatusApproved

		// 2. Permission enforcement via router middleware (leave.approve)
		// WHY: Business rule - only HR/approvers can cancel leave requests
		// Permission is validated at router level using middleware.RequirePermission("leave.approve")

		// 3. Fetch employee and leave type for response
		var err error
		employee, err = u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
		if err != nil {
			return fmt.Errorf("failed to fetch employee: %w", err)
		}

		leaveType, err = u.leaveTypeRepo.FindByID(ctx, leaveRequest.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to fetch leave type: %w", err)
		}

		// 4. Update status to CANCELLED
		leaveRequest.Status = models.LeaveStatusCancelled

		// Store cancellation note if provided
		if req.CancellationNote != nil {
			leaveRequest.RejectionNote = req.CancellationNote // Reuse RejectionNote field for cancellation
		}

		// Use canceller ID from DTO if provided, otherwise current user
		if req.CancelledBy != nil {
			leaveRequest.RejectedBy = req.CancelledBy // Reuse RejectedBy field for canceller
		} else {
			leaveRequest.RejectedBy = &currentUserID
		}

		leaveRequest.UpdatedBy = &currentUserID

		// 5. TODO: Trigger notification to employee
		// notificationService.Send(employee.Email, "Leave Request Cancelled", ...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 6. Delete attendance records linked to this leave request (only for APPROVED → CANCELLED)
	// WHY: PENDING leave requests have no associated attendance records yet
	if wasApproved {
		if err := u.attendanceRecordRepo.DeleteByLeaveRequestID(ctx, id); err != nil {
			fmt.Printf("WARNING: failed to delete attendance records for cancelled leave %s: %v\n", id, err)
		}
	}

	// 7. Fetch updated leave request to return
	updatedLeaveRequest, err := u.leaveRequestRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated leave request: %w", err)
	}

	return u.mapper.ToDetailResponseDTO(updatedLeaveRequest, employee, leaveType), nil
}

// GetFormData returns data for form dropdowns (employees, leave types, and current user's balance)
func (u *leaveRequestUsecase) GetFormData(ctx context.Context, currentUserID string) (*dto.FormDataResponseDTO, error) {
	// 1. Fetch all active employees
	employees, err := listScopedActiveEmployees(ctx, u.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}

	// 2. Fetch all active leave types
	leaveTypes, err := u.leaveTypeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leave types: %w", err)
	}

	// 3. Map employees to form DTOs with their leave balance
	formEmployees := make([]dto.FormEmployeeDTO, 0, len(employees))
	for _, emp := range employees {
		// Calculate remaining balance for this employee
		usedDays, err := u.leaveRequestRepo.CalculateUsedLeaveDays(ctx, emp.ID, true)
		if err != nil {
			usedDays = 0 // Default to 0 if calculation fails
		}
		remainingBalance := float64(emp.TotalLeaveQuota) - float64(usedDays)
		if remainingBalance < 0 {
			remainingBalance = 0 // Ensure balance is never negative
		}

		formEmployees = append(formEmployees, dto.FormEmployeeDTO{
			ID:               emp.ID,
			Name:             emp.Name,
			EmployeeCode:     emp.EmployeeCode,
			RemainingBalance: remainingBalance,
		})
	}

	// 4. Map leave types to form DTOs
	formLeaveTypes := make([]dto.FormLeaveTypeDTO, 0, len(leaveTypes))
	for _, lt := range leaveTypes {
		formLeaveTypes = append(formLeaveTypes, dto.FormLeaveTypeDTO{
			ID:      lt.ID,
			Name:    lt.Name,
			Code:    lt.Code,
			MaxDays: lt.MaxDays,
		})
	}

	return &dto.FormDataResponseDTO{
		Employees:  formEmployees,
		LeaveTypes: formLeaveTypes,
	}, nil
}

// Reapprove re-approves a previously cancelled or rejected leave request
func (u *leaveRequestUsecase) Reapprove(ctx context.Context, id string, req *dto.ApproveLeaveRequestDTO, currentUserID string) (*dto.LeaveRequestDetailResponseDTO, error) {
	// Reuse the Approve method since CanBeApproved() now includes REJECTED and CANCELLED statuses
	return u.Approve(ctx, id, req, currentUserID)
}

// createLeaveAttendanceRecords creates LEAVE attendance records for each working day in the leave period
func (u *leaveRequestUsecase) createLeaveAttendanceRecords(ctx context.Context, leaveRequest *models.LeaveRequest) error {
	// First, delete any existing attendance records for this leave request (idempotent)
	if err := u.attendanceRecordRepo.DeleteByLeaveRequestID(ctx, leaveRequest.ID); err != nil {
		return fmt.Errorf("failed to clean up existing attendance records: %w", err)
	}

	// Resolve companyID for company-scoped holiday filtering
	companyID := ""
	emp, err := u.employeeRepo.FindByID(ctx, leaveRequest.EmployeeID)
	if err == nil && emp.CompanyID != nil {
		companyID = *emp.CompanyID
	}

	// Fetch holidays in the leave period (global + company-specific)
	holidays, err := u.holidayRepo.FindByDateRangeForCompany(ctx, leaveRequest.StartDate, leaveRequest.EndDate, companyID)
	if err != nil {
		return fmt.Errorf("failed to fetch holidays: %w", err)
	}

	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		holidayMap[h.Date.Format("2006-01-02")] = true
	}

	// Create attendance records for each working day in the leave period
	var records []models.AttendanceRecord
	currentDate := leaveRequest.StartDate
	leaveRequestID := leaveRequest.ID

	for !currentDate.After(leaveRequest.EndDate) {
		weekday := currentDate.Weekday()
		dateStr := currentDate.Format("2006-01-02")

		// Skip weekends and holidays
		if weekday != time.Saturday && weekday != time.Sunday && !holidayMap[dateStr] {
			// Check if an attendance record already exists for this date
			existing, _ := u.attendanceRecordRepo.FindByEmployeeAndDate(ctx, leaveRequest.EmployeeID, currentDate)
			if existing == nil {
				records = append(records, models.AttendanceRecord{
					EmployeeID:     leaveRequest.EmployeeID,
					Date:           currentDate,
					Status:         models.AttendanceStatusLeave,
					LeaveRequestID: &leaveRequestID,
					Notes:          fmt.Sprintf("Auto-created from approved leave request"),
					IsManualEntry:  true,
				})
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if len(records) > 0 {
		return u.attendanceRecordRepo.CreateBatch(ctx, records)
	}

	return nil
}

func (u *leaveRequestUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.LeaveRequestAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, fmt.Errorf("db is nil")
	}
	page, perPage = normalizeLeaveRequestAuditPagination(page, perPage)

	tx := u.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("audit_logs.permission_code LIKE ?", "leave_request.%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]leaveRequestAuditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return mapLeaveRequestAuditEntries(rows), total, nil
}

func normalizeLeaveRequestAuditPagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func mapLeaveRequestAuditEntries(rows []leaveRequestAuditRow) []dto.LeaveRequestAuditTrailEntry {
	entries := make([]dto.LeaveRequestAuditTrailEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, dto.LeaveRequestAuditTrailEntry{
			ID:             row.ID,
			Action:         row.Action,
			PermissionCode: row.PermissionCode,
			TargetID:       row.TargetID,
			Metadata:       parseLeaveRequestAuditMetadata(row.Metadata),
			User:           buildLeaveRequestAuditUser(row),
			CreatedAt:      row.CreatedAt,
		})
	}
	return entries
}

func parseLeaveRequestAuditMetadata(raw string) map[string]interface{} {
	meta := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &meta)
	}
	return meta
}

func buildLeaveRequestAuditUser(row leaveRequestAuditRow) *dto.LeaveRequestAuditTrailUser {
	if row.ActorID == "" {
		return nil
	}
	email := ""
	name := ""
	if row.ActorEmail != nil {
		email = *row.ActorEmail
	}
	if row.ActorName != nil {
		name = *row.ActorName
	}
	return &dto.LeaveRequestAuditTrailUser{ID: row.ActorID, Email: email, Name: name}
}
