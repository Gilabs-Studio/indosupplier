package mapper

import (
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

// WorkScheduleMapper handles mapping between WorkSchedule model and DTOs
type WorkScheduleMapper struct{}

// NewWorkScheduleMapper creates a new WorkScheduleMapper
func NewWorkScheduleMapper() *WorkScheduleMapper {
	return &WorkScheduleMapper{}
}

// ToModel converts CreateWorkScheduleRequest to WorkSchedule model
func (m *WorkScheduleMapper) ToModel(req *dto.CreateWorkScheduleRequest) *models.WorkSchedule {
	ws := &models.WorkSchedule{
		Name:                       req.Name,
		Description:                req.Description,
		DivisionID:                 req.DivisionID,
		IsDefault:                  req.IsDefault,
		IsActive:                   req.IsActive,
		StartTime:                  req.StartTime,
		EndTime:                    req.EndTime,
		IsFlexible:                 req.IsFlexible,
		FlexibleStartTime:          req.FlexibleStartTime,
		FlexibleEndTime:            req.FlexibleEndTime,
		Breaks:                     m.mapBreaksDTOToModel(req.Breaks),
		WorkingDays:                req.WorkingDays,
		LateToleranceMinutes:       req.LateToleranceMinutes,
		EarlyLeaveToleranceMinutes: req.EarlyLeaveToleranceMinutes,
		RequireGPS:                 req.RequireGPS,
		GPSRadiusMeter:             req.GPSRadiusMeter,
		OfficeLatitude:             req.OfficeLatitude,
		OfficeLongitude:            req.OfficeLongitude,
	}

	// Calculate working hours from start and end time
	ws.WorkingHoursPerDay = ws.CalculateWorkingHours()

	return ws
}

// mapBreaksDTOToModel converts DTO breaks to model breaks
func (m *WorkScheduleMapper) mapBreaksDTOToModel(breaks []dto.BreakTime) []models.BreakTime {
	result := make([]models.BreakTime, len(breaks))
	for i, b := range breaks {
		result[i] = models.BreakTime{
			StartTime: b.StartTime,
			EndTime:   b.EndTime,
		}
	}
	return result
}

// mapBreaksModelToDTO converts model breaks to DTO breaks
func (m *WorkScheduleMapper) mapBreaksModelToDTO(breaks []models.BreakTime) []dto.BreakTime {
	result := make([]dto.BreakTime, len(breaks))
	for i, b := range breaks {
		result[i] = dto.BreakTime{
			StartTime: b.StartTime,
			EndTime:   b.EndTime,
		}
	}
	return result
}

// ApplyUpdate applies UpdateWorkScheduleRequest to existing model
func (m *WorkScheduleMapper) ApplyUpdate(ws *models.WorkSchedule, req *dto.UpdateWorkScheduleRequest) {
	if req.Name != nil {
		ws.Name = *req.Name
	}
	if req.Description != nil {
		ws.Description = *req.Description
	}
	if req.DivisionID != nil {
		ws.DivisionID = req.DivisionID
	}
	if req.IsDefault != nil {
		ws.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		ws.IsActive = *req.IsActive
	}
	if req.StartTime != nil {
		ws.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		ws.EndTime = *req.EndTime
	}
	if req.IsFlexible != nil {
		ws.IsFlexible = *req.IsFlexible
	}
	if req.FlexibleStartTime != nil {
		ws.FlexibleStartTime = *req.FlexibleStartTime
	}
	if req.FlexibleEndTime != nil {
		ws.FlexibleEndTime = *req.FlexibleEndTime
	}
	if req.Breaks != nil {
		ws.Breaks = m.mapBreaksDTOToModel(*req.Breaks)
	}
	if req.WorkingDays != nil {
		ws.WorkingDays = *req.WorkingDays
	}
	if req.LateToleranceMinutes != nil {
		ws.LateToleranceMinutes = *req.LateToleranceMinutes
	}
	if req.EarlyLeaveToleranceMinutes != nil {
		ws.EarlyLeaveToleranceMinutes = *req.EarlyLeaveToleranceMinutes
	}
	if req.RequireGPS != nil {
		ws.RequireGPS = *req.RequireGPS
	}
	if req.GPSRadiusMeter != nil {
		ws.GPSRadiusMeter = *req.GPSRadiusMeter
	}
	if req.OfficeLatitude != nil {
		ws.OfficeLatitude = *req.OfficeLatitude
	}
	if req.OfficeLongitude != nil {
		ws.OfficeLongitude = *req.OfficeLongitude
	}

	// Recalculate working hours if start or end time changed
	if req.StartTime != nil || req.EndTime != nil {
		ws.WorkingHoursPerDay = ws.CalculateWorkingHours()
	}
}

// ToResponse converts WorkSchedule model to response DTO
func (m *WorkScheduleMapper) ToResponse(ws *models.WorkSchedule) *dto.WorkScheduleResponse {
	resp := &dto.WorkScheduleResponse{
		ID:                         ws.ID,
		Name:                       ws.Name,
		Description:                ws.Description,
		DivisionID:                 ws.DivisionID,
		IsDefault:                  ws.IsDefault,
		IsActive:                   ws.IsActive,
		StartTime:                  ws.StartTime,
		EndTime:                    ws.EndTime,
		IsFlexible:                 ws.IsFlexible,
		FlexibleStartTime:          ws.FlexibleStartTime,
		FlexibleEndTime:            ws.FlexibleEndTime,
		Breaks:                     m.mapBreaksModelToDTO(ws.Breaks),
		WorkingDays:                ws.WorkingDays,
		WorkingDaysDisplay:         m.bitmaskToWorkingDays(ws.WorkingDays),
		WorkingHoursPerDay:         ws.WorkingHoursPerDay,
		LateToleranceMinutes:       ws.LateToleranceMinutes,
		EarlyLeaveToleranceMinutes: ws.EarlyLeaveToleranceMinutes,
		RequireGPS:                 ws.RequireGPS,
		GPSRadiusMeter:             ws.GPSRadiusMeter,
		OfficeLatitude:             ws.OfficeLatitude,
		OfficeLongitude:            ws.OfficeLongitude,
		CreatedAt:                  ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                  ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	return resp
}

// ToResponseList converts a list of WorkSchedule models to response DTOs
func (m *WorkScheduleMapper) ToResponseList(schedules []models.WorkSchedule) []dto.WorkScheduleResponse {
	responses := make([]dto.WorkScheduleResponse, len(schedules))
	for i, ws := range schedules {
		responses[i] = *m.ToResponse(&ws)
	}
	return responses
}

// bitmaskToWorkingDays converts bitmask to array of day names
func (m *WorkScheduleMapper) bitmaskToWorkingDays(bitmask int) []string {
	days := []struct {
		bit  int
		name string
	}{
		{1, "Mon"},
		{2, "Tue"},
		{4, "Wed"},
		{8, "Thu"},
		{16, "Fri"},
		{32, "Sat"},
		{64, "Sun"},
	}

	var result []string
	for _, day := range days {
		if bitmask&day.bit != 0 {
			result = append(result, day.name)
		}
	}
	return result
}
