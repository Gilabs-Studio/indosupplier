package mapper

import (
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
)

// OvertimeRequestMapper handles mapping between OvertimeRequest model and DTOs
type OvertimeRequestMapper struct{}

// NewOvertimeRequestMapper creates a new OvertimeRequestMapper
func NewOvertimeRequestMapper() *OvertimeRequestMapper {
	return &OvertimeRequestMapper{}
}

// ToModel converts CreateOvertimeRequestDTO to OvertimeRequest model
func (m *OvertimeRequestMapper) ToModel(req *dto.CreateOvertimeRequestDTO, employeeID string) (*models.OvertimeRequest, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	// Parse times and combine with date
	startTime, err := m.parseTimeWithDate(date, req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := m.parseTimeWithDate(date, req.EndTime)
	if err != nil {
		return nil, err
	}

	// If end time is before start time, assume it's next day
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	or := &models.OvertimeRequest{
		EmployeeID:  employeeID,
		Date:        date,
		RequestType: models.OvertimeRequestType(req.RequestType),
		StartTime:   startTime,
		EndTime:     endTime,
		Reason:      req.Reason,
		Description: req.Description,
		TaskDetails: req.TaskDetails,
		Status:      models.OvertimeStatusPending,
	}

	or.CalculateActualMinutes()

	// For pre-approved, set planned minutes
	if req.RequestType == string(models.OvertimeTypePreApproved) {
		or.PlannedMinutes = or.ActualMinutes
	}

	return or, nil
}

// ApplyUpdate applies UpdateOvertimeRequestDTO to existing model
func (m *OvertimeRequestMapper) ApplyUpdate(or *models.OvertimeRequest, req *dto.UpdateOvertimeRequestDTO) error {
	if req.StartTime != nil {
		startTime, err := m.parseTimeWithDate(or.Date, *req.StartTime)
		if err != nil {
			return err
		}
		or.StartTime = startTime
	}
	if req.EndTime != nil {
		endTime, err := m.parseTimeWithDate(or.Date, *req.EndTime)
		if err != nil {
			return err
		}
		// If end time is before start time, assume it's next day
		if endTime.Before(or.StartTime) {
			endTime = endTime.AddDate(0, 0, 1)
		}
		or.EndTime = endTime
	}
	if req.Reason != nil {
		or.Reason = *req.Reason
	}
	if req.Description != nil {
		or.Description = *req.Description
	}
	if req.TaskDetails != nil {
		or.TaskDetails = *req.TaskDetails
	}

	// Recalculate actual minutes
	or.CalculateActualMinutes()

	return nil
}

// ToResponse converts OvertimeRequest model to response DTO
func (m *OvertimeRequestMapper) ToResponse(or *models.OvertimeRequest) *dto.OvertimeRequestResponse {
	resp := &dto.OvertimeRequestResponse{
		ID:                 or.ID,
		EmployeeID:         or.EmployeeID,
		Date:               or.Date.Format("2006-01-02"),
		RequestType:        string(or.RequestType),
		StartTime:          or.StartTime.Format("15:04"),
		EndTime:            or.EndTime.Format("15:04"),
		PlannedMinutes:     or.PlannedMinutes,
		PlannedHours:       m.formatMinutesToHours(or.PlannedMinutes),
		ActualMinutes:      or.ActualMinutes,
		ActualHours:        m.formatMinutesToHours(or.ActualMinutes),
		ApprovedMinutes:    or.ApprovedMinutes,
		ApprovedHours:      m.formatMinutesToHours(or.ApprovedMinutes),
		Reason:             or.Reason,
		Description:        or.Description,
		TaskDetails:        or.TaskDetails,
		Status:             string(or.Status),
		ApprovedBy:         or.ApprovedBy,
		RejectedBy:         or.RejectedBy,
		RejectReason:       or.RejectReason,
		AttendanceRecordID: or.AttendanceRecordID,
		OvertimeRate:       or.OvertimeRate,
		CompensationAmount: or.CompensationAmount,
		CreatedAt:          or.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          or.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if or.ApprovedAt != nil {
		approvedAtStr := or.ApprovedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ApprovedAt = &approvedAtStr
	}
	if or.RejectedAt != nil {
		rejectedAtStr := or.RejectedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.RejectedAt = &rejectedAtStr
	}

	return resp
}

// ToResponseList converts a list of OvertimeRequest models to response DTOs
func (m *OvertimeRequestMapper) ToResponseList(requests []models.OvertimeRequest) []dto.OvertimeRequestResponse {
	responses := make([]dto.OvertimeRequestResponse, len(requests))
	for i, or := range requests {
		responses[i] = *m.ToResponse(&or)
	}
	return responses
}

// parseTimeWithDate parses a time string and combines it with a date
// Supports both "HH:MM:SS" and "HH:MM" formats
func (m *OvertimeRequestMapper) parseTimeWithDate(date time.Time, timeStr string) (time.Time, error) {
	// Try parsing with seconds first (HH:MM:SS)
	t, err := time.Parse("15:04:05", timeStr)
	if err == nil {
		return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), t.Second(), 0, date.Location()), nil
	}

	// Fall back to parsing without seconds (HH:MM)
	t, err = time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %s (expected HH:MM or HH:MM:SS)", timeStr)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
}

// EnrichResponse enriches an overtime request response with employee data
func (m *OvertimeRequestMapper) EnrichResponse(resp *dto.OvertimeRequestResponse, employeeMap map[string]*orgModels.Employee) {
	if emp, ok := employeeMap[resp.EmployeeID]; ok {
		resp.EmployeeName = emp.Name
		resp.EmployeeCode = emp.EmployeeCode
		if emp.Division != nil {
			resp.DivisionName = emp.Division.Name
		}
	}
	// Enrich approver name if approved
	if resp.ApprovedBy != nil {
		if approver, ok := employeeMap[*resp.ApprovedBy]; ok {
			resp.ApprovedByName = approver.Name
		}
	}
	// Enrich rejecter name if rejected
	if resp.RejectedBy != nil {
		if rejecter, ok := employeeMap[*resp.RejectedBy]; ok {
			resp.RejectedByName = rejecter.Name
		}
	}
}

// EnrichResponseList enriches a list of overtime request responses with employee data
func (m *OvertimeRequestMapper) EnrichResponseList(responses []dto.OvertimeRequestResponse, employeeMap map[string]*orgModels.Employee) {
	for i := range responses {
		m.EnrichResponse(&responses[i], employeeMap)
	}
}

// formatMinutesToHours converts minutes to "Xh Ym" format
func (m *OvertimeRequestMapper) formatMinutesToHours(minutes int) string {
	if minutes <= 0 {
		return "0h 0m"
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}
