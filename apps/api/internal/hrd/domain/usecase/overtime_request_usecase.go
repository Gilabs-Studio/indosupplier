package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"gorm.io/gorm"
)

var (
	ErrOvertimeRequestNotFound     = errors.New("overtime request not found")
	ErrOvertimeAlreadyProcessed    = errors.New("overtime request already processed")
	ErrCannotModifyApprovedRequest = errors.New("cannot modify approved request")
	ErrUnauthorizedApproval        = errors.New("not authorized to approve this request")
)

// OvertimeRequestUsecase defines the interface for overtime request business logic
type OvertimeRequestUsecase interface {
	List(ctx context.Context, req *dto.ListOvertimeRequestsRequest) ([]dto.OvertimeRequestResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.OvertimeRequestResponse, error)
	GetPendingForManager(ctx context.Context, managerID string) ([]dto.OvertimeRequestResponse, error)
	Create(ctx context.Context, req *dto.CreateOvertimeRequestDTO, employeeID string) (*dto.OvertimeRequestResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateOvertimeRequestDTO) (*dto.OvertimeRequestResponse, error)
	Approve(ctx context.Context, id string, req *dto.ApproveOvertimeRequest, approverID string) (*dto.OvertimeRequestResponse, error)
	Reject(ctx context.Context, id string, req *dto.RejectOvertimeRequest, rejecterID string) (*dto.OvertimeRequestResponse, error)
	Cancel(ctx context.Context, id string, employeeID string) error
	Delete(ctx context.Context, id string) error
	CreateAutoDetectedOvertime(ctx context.Context, attendanceRecordID, employeeID string, overtimeMinutes int, date time.Time, startTime, endTime time.Time) (*models.OvertimeRequest, error)
	GetEmployeeMonthlySummary(ctx context.Context, employeeID string, year, month int) (*dto.OvertimeSummaryResponse, error)
	GetUnnotifiedPendingRequests(ctx context.Context) ([]dto.PendingOvertimeNotification, error)
	MarkAsNotified(ctx context.Context, ids []string) error
}

type overtimeRequestUsecase struct {
	repo         repositories.OvertimeRequestRepository
	employeeRepo orgRepos.EmployeeRepository
	mapper       *mapper.OvertimeRequestMapper
}

// NewOvertimeRequestUsecase creates a new OvertimeRequestUsecase
func NewOvertimeRequestUsecase(repo repositories.OvertimeRequestRepository, employeeRepo orgRepos.EmployeeRepository) OvertimeRequestUsecase {
	return &overtimeRequestUsecase{
		repo:         repo,
		employeeRepo: employeeRepo,
		mapper:       mapper.NewOvertimeRequestMapper(),
	}
}

func (u *overtimeRequestUsecase) List(ctx context.Context, req *dto.ListOvertimeRequestsRequest) ([]dto.OvertimeRequestResponse, *utils.PaginationResult, error) {
	requests, total, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := u.mapper.ToResponseList(requests)

	// Enrich with employee data
	employeeIDs := make([]string, 0, len(requests))
	for _, r := range requests {
		employeeIDs = append(employeeIDs, r.EmployeeID)
		// Also add approver/rejecter IDs if present
		if r.ApprovedBy != nil && *r.ApprovedBy != "" {
			employeeIDs = append(employeeIDs, *r.ApprovedBy)
		}
		if r.RejectedBy != nil && *r.RejectedBy != "" {
			employeeIDs = append(employeeIDs, *r.RejectedBy)
		}
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

func (u *overtimeRequestUsecase) GetByID(ctx context.Context, id string) (*dto.OvertimeRequestResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.OvertimeRequest{}, id, security.HRDScopeQueryOptions()) {
		return nil, ErrOvertimeRequestNotFound
	}

	or, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOvertimeRequestNotFound
		}
		return nil, err
	}

	resp := u.mapper.ToResponse(or)
	// Enrich with employee data (include approver and rejecter if present)
	employeeIDs := []string{or.EmployeeID}
	if or.ApprovedBy != nil && *or.ApprovedBy != "" {
		employeeIDs = append(employeeIDs, *or.ApprovedBy)
	}
	if or.RejectedBy != nil && *or.RejectedBy != "" {
		employeeIDs = append(employeeIDs, *or.RejectedBy)
	}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	u.mapper.EnrichResponse(resp, employeeMap)
	return resp, nil
}

func (u *overtimeRequestUsecase) GetPendingForManager(ctx context.Context, managerID string) ([]dto.OvertimeRequestResponse, error) {
	requests, err := u.repo.FindPendingByManager(ctx, managerID)
	if err != nil {
		return nil, err
	}

	responses := u.mapper.ToResponseList(requests)
	// Enrich with employee data
	employeeIDs := make([]string, 0, len(requests))
	for _, r := range requests {
		employeeIDs = append(employeeIDs, r.EmployeeID)
	}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	u.mapper.EnrichResponseList(responses, employeeMap)
	return responses, nil
}

func (u *overtimeRequestUsecase) Create(ctx context.Context, req *dto.CreateOvertimeRequestDTO, employeeID string) (*dto.OvertimeRequestResponse, error) {
	or, err := u.mapper.ToModel(req, employeeID)
	if err != nil {
		return nil, err
	}

	// Set default overtime rate based on date (weekday vs weekend)
	if or.Date.Weekday() == time.Saturday || or.Date.Weekday() == time.Sunday {
		or.OvertimeRate = 2.0 // 2x for weekends
	} else {
		or.OvertimeRate = 1.5 // 1.5x for weekdays
	}

	if err := u.repo.Create(ctx, or); err != nil {
		return nil, err
	}

	// Create notification for approvers
	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, database.DB, notificationService.ApprovalNotificationParams{
		PermissionCode: "overtime.approve",
		EntityType:     "overtime",
		EntityID:       or.ID,
		Title:          "Overtime Request Approval",
		Message:        "An overtime request has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create overtime notification: %v", err)
	}

	resp := u.mapper.ToResponse(or)
	// Enrich with employee data
	employeeMap := u.buildEmployeeMap(ctx, []string{or.EmployeeID})
	u.mapper.EnrichResponse(resp, employeeMap)
	return resp, nil
}

func (u *overtimeRequestUsecase) Update(ctx context.Context, id string, req *dto.UpdateOvertimeRequestDTO) (*dto.OvertimeRequestResponse, error) {
	or, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOvertimeRequestNotFound
		}
		return nil, err
	}

	// Can only update pending requests
	if or.Status != models.OvertimeStatusPending {
		return nil, ErrCannotModifyApprovedRequest
	}

	if err := u.mapper.ApplyUpdate(or, req); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, or); err != nil {
		return nil, err
	}

	return u.mapper.ToResponse(or), nil
}

func (u *overtimeRequestUsecase) Approve(ctx context.Context, id string, req *dto.ApproveOvertimeRequest, approverID string) (*dto.OvertimeRequestResponse, error) {
	or, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOvertimeRequestNotFound
		}
		return nil, err
	}

	// Can only approve pending requests
	if or.Status != models.OvertimeStatusPending {
		return nil, ErrOvertimeAlreadyProcessed
	}

	if err := u.repo.Approve(ctx, id, approverID, req.ApprovedMinutes); err != nil {
		return nil, err
	}

	// Reload to get updated data
	or, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := u.mapper.ToResponse(or)
	// Enrich with employee data (include approver)
	employeeIDs := []string{or.EmployeeID, approverID}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	u.mapper.EnrichResponse(resp, employeeMap)
	return resp, nil
}

func (u *overtimeRequestUsecase) Reject(ctx context.Context, id string, req *dto.RejectOvertimeRequest, rejecterID string) (*dto.OvertimeRequestResponse, error) {
	or, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOvertimeRequestNotFound
		}
		return nil, err
	}

	// Can only reject pending requests
	if or.Status != models.OvertimeStatusPending {
		return nil, ErrOvertimeAlreadyProcessed
	}

	if err := u.repo.Reject(ctx, id, rejecterID, req.Reason); err != nil {
		return nil, err
	}

	// Reload to get updated data
	or, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := u.mapper.ToResponse(or)
	// Enrich with employee data (include rejecter)
	employeeIDs := []string{or.EmployeeID, rejecterID}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	u.mapper.EnrichResponse(resp, employeeMap)
	return resp, nil
}

func (u *overtimeRequestUsecase) Cancel(ctx context.Context, id string, employeeID string) error {
	or, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOvertimeRequestNotFound
		}
		return err
	}

	// Verify ownership
	if or.EmployeeID != employeeID {
		return errors.New("not authorized to cancel this request")
	}

	// Can only cancel pending requests
	if or.Status != models.OvertimeStatusPending {
		return ErrOvertimeAlreadyProcessed
	}

	return u.repo.Cancel(ctx, id)
}

func (u *overtimeRequestUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOvertimeRequestNotFound
		}
		return err
	}

	return u.repo.Delete(ctx, id)
}

func (u *overtimeRequestUsecase) CreateAutoDetectedOvertime(ctx context.Context, attendanceRecordID, employeeID string, overtimeMinutes int, date time.Time, startTime, endTime time.Time) (*models.OvertimeRequest, error) {
	or := &models.OvertimeRequest{
		EmployeeID:         employeeID,
		Date:               date,
		RequestType:        models.OvertimeTypeAutoDetected,
		StartTime:          startTime,
		EndTime:            endTime,
		ActualMinutes:      overtimeMinutes,
		Reason:             "Auto-detected from clock out time",
		Status:             models.OvertimeStatusPending,
		AttendanceRecordID: &attendanceRecordID,
	}

	// Set overtime rate based on date
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		or.OvertimeRate = 2.0
	} else {
		or.OvertimeRate = 1.5
	}

	if err := u.repo.Create(ctx, or); err != nil {
		return nil, err
	}

	return or, nil
}

func (u *overtimeRequestUsecase) GetEmployeeMonthlySummary(ctx context.Context, employeeID string, year, month int) (*dto.OvertimeSummaryResponse, error) {
	// Get all requests for the month
	// WHY: Use per-employee timezone so "month" boundaries match the employee's local time
	empLoc := apptime.LocationForEmployee(employeeID)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, empLoc)
	lastDay := firstDay.AddDate(0, 1, -1)

	req := &dto.ListOvertimeRequestsRequest{
		EmployeeID: employeeID,
		DateFrom:   firstDay.Format("2006-01-02"),
		DateTo:     lastDay.Format("2006-01-02"),
		Page:       1,
		PerPage:    100,
	}

	requests, _, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	summary := &dto.OvertimeSummaryResponse{
		EmployeeID: employeeID,
		Year:       year,
		Month:      month,
	}

	for _, r := range requests {
		summary.TotalRequestedMinutes += r.ActualMinutes

		switch r.Status {
		case models.OvertimeStatusPending:
			summary.PendingRequests++
		case models.OvertimeStatusApproved:
			summary.ApprovedRequests++
			summary.TotalApprovedMinutes += r.ApprovedMinutes
		case models.OvertimeStatusRejected:
			summary.RejectedRequests++
			summary.TotalRejectedMinutes += r.ActualMinutes
		}
	}

	return summary, nil
}

func (u *overtimeRequestUsecase) GetUnnotifiedPendingRequests(ctx context.Context) ([]dto.PendingOvertimeNotification, error) {
	requests, err := u.repo.GetUnnotifiedPendingRequests(ctx)
	if err != nil {
		return nil, err
	}

	notifications := make([]dto.PendingOvertimeNotification, len(requests))
	// Enrich notifications with employee data
	employeeIDs := make([]string, 0, len(requests))
	for _, r := range requests {
		employeeIDs = append(employeeIDs, r.EmployeeID)
	}
	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	for i, r := range requests {
		resp := u.mapper.ToResponse(&r)
		u.mapper.EnrichResponse(resp, employeeMap)
		notification := dto.PendingOvertimeNotification{
			OvertimeRequest: *resp,
		}
		if emp, ok := employeeMap[r.EmployeeID]; ok {
			notification.EmployeeName = emp.Name
			if emp.Division != nil {
				notification.DivisionName = emp.Division.Name
			}
		}
		notifications[i] = notification
	}

	return notifications, nil
}

func (u *overtimeRequestUsecase) MarkAsNotified(ctx context.Context, ids []string) error {
	return u.repo.MarkNotified(ctx, ids)
}

// buildEmployeeMap batch-fetches employees by IDs and builds a lookup map
func (u *overtimeRequestUsecase) buildEmployeeMap(ctx context.Context, ids []string) map[string]*orgModels.Employee {
	m := make(map[string]*orgModels.Employee)
	if len(ids) == 0 {
		return m
	}

	unique := make(map[string]bool)
	dedupIDs := make([]string, 0)
	for _, id := range ids {
		if !unique[id] && id != "" {
			unique[id] = true
			dedupIDs = append(dedupIDs, id)
		}
	}

	if len(dedupIDs) == 0 {
		return m
	}

	employees, err := u.employeeRepo.FindByIDs(ctx, dedupIDs)
	if err != nil {
		return m
	}

	for i := range employees {
		emp := employees[i]
		m[emp.ID] = &emp
	}
	return m
}
