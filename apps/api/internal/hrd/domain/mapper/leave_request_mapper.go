package mapper

import (
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
)

// LeaveRequestMapper handles conversion between LeaveRequest models and DTOs
type LeaveRequestMapper struct{}

// NewLeaveRequestMapper creates a new LeaveRequestMapper instance
func NewLeaveRequestMapper() *LeaveRequestMapper {
	return &LeaveRequestMapper{}
}

// ToModel converts CreateLeaveRequestDTO to LeaveRequest model
// WHY: Parses date strings and calculates TotalDays based on Duration
func (m *LeaveRequestMapper) ToModel(req *dto.CreateLeaveRequestDTO, totalDays float64, createdBy *string) (*models.LeaveRequest, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	leaveRequest := &models.LeaveRequest{
		EmployeeID:    req.EmployeeID,
		LeaveTypeID:   req.LeaveTypeID,
		StartDate:     startDate,
		EndDate:       endDate,
		Duration:      models.LeaveDuration(req.Duration),
		TotalDays:     totalDays,
		Reason:        req.Reason,
		Status:        models.LeaveStatusPending,
		AttachmentURL: req.AttachmentURL,
		CreatedBy:     createdBy,
	}

	return leaveRequest, nil
}

// ToResponseDTO converts LeaveRequest model to LeaveRequestResponseDTO (list view with employee name and leave type)
// WHY: List view should show names instead of IDs for better readability
func (m *LeaveRequestMapper) ToResponseDTO(model *models.LeaveRequest, employee *orgModels.Employee, leaveType *coreModels.LeaveType, employeeMap map[string]*orgModels.Employee) *dto.LeaveRequestResponseDTO {
	if model == nil {
		return nil
	}

	employeeName := "Unknown"
	if employee != nil {
		employeeName = employee.Name
	}

	leaveTypeName := "Unknown"
	if leaveType != nil {
		leaveTypeName = leaveType.Name
	}

	dto := &dto.LeaveRequestResponseDTO{
		ID:           model.ID,
		EmployeeName: employeeName,
		LeaveType:    leaveTypeName,
		StartDate:    model.StartDate.Format("2006-01-02"),
		EndDate:      model.EndDate.Format("2006-01-02"),
		Duration:     string(model.Duration),
		TotalDays:    model.TotalDays,
		Reason:       model.Reason,
		Status:       string(model.Status),
		CreatedAt:    model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    model.UpdatedAt.Format(time.RFC3339),
	}

	// Add rejection details if rejected
	if model.RejectedBy != nil {
		dto.RejectedBy = model.RejectedBy
		if rejecter, ok := employeeMap[*model.RejectedBy]; ok && rejecter != nil {
			dto.RejectedByName = rejecter.Name
		}
	}
	if model.RejectionNote != nil {
		dto.RejectionNote = model.RejectionNote
	}

	return dto
}

// ToDetailResponseDTO converts LeaveRequest model to LeaveRequestDetailResponseDTO with full details
// WHY: Detail view should include complete employee and leave type information
func (m *LeaveRequestMapper) ToDetailResponseDTO(model *models.LeaveRequest, employee *orgModels.Employee, leaveType *coreModels.LeaveType) *dto.LeaveRequestDetailResponseDTO {
	if model == nil {
		return nil
	}

	response := &dto.LeaveRequestDetailResponseDTO{
		ID:                 model.ID,
		StartDate:          model.StartDate.Format("2006-01-02"),
		EndDate:            model.EndDate.Format("2006-01-02"),
		Duration:           string(model.Duration),
		TotalDays:          model.TotalDays,
		Reason:             model.Reason,
		Status:             string(model.Status),
		AttachmentURL:      model.AttachmentURL,
		RejectionNote:      model.RejectionNote,
		RejectedBy:         model.RejectedBy,
		IsCarryOver:        model.IsCarryOver,
		RemainingCarryOver: model.RemainingCarryOver,
		CreatedAt:          model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          model.UpdatedAt.Format(time.RFC3339),
		CreatedBy:          model.CreatedBy,
		UpdatedBy:          model.UpdatedBy,
	}

	// Map employee details
	if employee != nil {
		phone := employee.Phone
		response.Employee = dto.EmployeeDetailDTO{
			ID:           employee.ID,
			Name:         employee.Name,
			Email:        employee.Email,
			Phone:        &phone,
			EmployeeCode: employee.EmployeeCode,
			// TODO: Add job position and division names from related entities
		}
	}

	// Map leave type details
	if leaveType != nil {
		response.LeaveType = dto.LeaveTypeDetailDTO{
			ID:               leaveType.ID,
			Name:             leaveType.Name,
			Code:             leaveType.Code,
			Description:      leaveType.Description,
			MaxDays:          leaveType.MaxDays,
			IsPaid:           leaveType.IsPaid,
			IsCutAnnualLeave: leaveType.IsCutAnnualLeave,
		}
	}

	// Format optional approval timestamp
	if model.ApprovedBy != nil {
		response.ApprovedBy = model.ApprovedBy
		if model.ApprovedAt != nil {
			approvedAt := model.ApprovedAt.Format(time.RFC3339)
			response.ApprovedAt = &approvedAt
		}
	}

	// Format optional carry-over expiry date
	if model.CarryOverExpiryDate != nil {
		expiryDate := model.CarryOverExpiryDate.Format("2006-01-02")
		response.CarryOverExpiryDate = &expiryDate
	}

	return response
}

// ToList converts a slice of LeaveRequest models to list view DTOs
// WHY: Batch conversion for pagination results
func (m *LeaveRequestMapper) ToList(models []*models.LeaveRequest, employees map[string]*orgModels.Employee, leaveTypes map[string]*coreModels.LeaveType) []*dto.LeaveRequestResponseDTO {
	if models == nil {
		return []*dto.LeaveRequestResponseDTO{}
	}

	dtos := make([]*dto.LeaveRequestResponseDTO, len(models))
	for i, model := range models {
		var employee *orgModels.Employee
		var leaveType *coreModels.LeaveType

		if employees != nil {
			employee = employees[model.EmployeeID]
		}
		if leaveTypes != nil {
			leaveType = leaveTypes[model.LeaveTypeID]
		}

		dtos[i] = m.ToResponseDTO(model, employee, leaveType, employees)
	}

	return dtos
}

// ApplyUpdateDTO applies UpdateLeaveRequestDTO to an existing model
// WHY: Only update fields that are provided (non-nil pointers)
func (m *LeaveRequestMapper) ApplyUpdateDTO(model *models.LeaveRequest, req *dto.UpdateLeaveRequestDTO, totalDays *float64, updatedBy *string) error {
	// Update leave type if provided
	if req.LeaveTypeID != nil {
		model.LeaveTypeID = *req.LeaveTypeID
	}

	// Update dates if provided
	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start_date format: %w", err)
		}
		model.StartDate = startDate
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end_date format: %w", err)
		}
		model.EndDate = endDate
	}

	// Update duration if provided
	if req.Duration != nil {
		model.Duration = models.LeaveDuration(*req.Duration)
	}

	// Update total days if recalculated
	if totalDays != nil {
		model.TotalDays = *totalDays
	}

	// Update reason if provided
	if req.Reason != nil {
		model.Reason = *req.Reason
	}

	// Update attachment URL if provided
	if req.AttachmentURL != nil {
		model.AttachmentURL = req.AttachmentURL
	}

	// Update audit field
	model.UpdatedBy = updatedBy

	return nil
}

// ToBalanceDTO converts balance calculation result to LeaveBalanceResponseDTO
// WHY: Provides clear breakdown of leave balance for employee
func (m *LeaveRequestMapper) ToBalanceDTO(employeeID string, totalQuota, usedDays, pendingDays int, carryOverBalance float64, carryOverExpiry *time.Time) *dto.LeaveBalanceResponseDTO {
	remainingBalance := totalQuota - usedDays
	totalAvailable := int(float64(remainingBalance) + carryOverBalance)

	response := &dto.LeaveBalanceResponseDTO{
		EmployeeID:          employeeID,
		TotalQuota:          totalQuota,
		UsedDays:            usedDays,
		RemainingBalance:    remainingBalance,
		CarryOverBalance:    carryOverBalance,
		TotalAvailableLeave: totalAvailable,
		PendingRequestsDays: pendingDays,
		CalculatedAt:        apptime.Now().Format(time.RFC3339),
	}

	if carryOverExpiry != nil {
		expiryDate := carryOverExpiry.Format("2006-01-02")
		response.CarryOverExpiryDate = &expiryDate
	}

	return response
}
