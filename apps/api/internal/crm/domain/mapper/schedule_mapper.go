package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToScheduleResponse converts a Schedule model to ScheduleResponse
func ToScheduleResponse(schedule *models.Schedule) dto.ScheduleResponse {
	resp := dto.ScheduleResponse{
		ID:                    schedule.ID,
		TaskID:                schedule.TaskID,
		EmployeeID:            schedule.EmployeeID,
		Title:                 schedule.Title,
		Description:           schedule.Description,
		ScheduledAt:           schedule.ScheduledAt.Format("2006-01-02T15:04:05+07:00"),
		Status:                schedule.Status,
		ReminderMinutesBefore: schedule.ReminderMinutesBefore,
		CreatedBy:             schedule.CreatedBy,
		CreatedAt:             schedule.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		UpdatedAt:             schedule.UpdatedAt.Format("2006-01-02T15:04:05+07:00"),
	}

	if schedule.EndAt != nil {
		t := schedule.EndAt.Format("2006-01-02T15:04:05+07:00")
		resp.EndAt = &t
	}

	if schedule.Task != nil {
		resp.Task = &dto.ScheduleTaskInfo{
			ID:       schedule.Task.ID,
			Title:    schedule.Task.Title,
			Status:   schedule.Task.Status,
			Priority: schedule.Task.Priority,
		}
	}

	if schedule.Employee != nil {
		resp.Employee = &dto.ScheduleEmployeeInfo{
			ID:           schedule.Employee.ID,
			EmployeeCode: schedule.Employee.EmployeeCode,
			Name:         schedule.Employee.Name,
		}
	}

	return resp
}

// ToScheduleResponseList converts a slice of Schedule models to responses
func ToScheduleResponseList(schedules []models.Schedule) []dto.ScheduleResponse {
	result := make([]dto.ScheduleResponse, 0, len(schedules))
	for i := range schedules {
		result = append(result, ToScheduleResponse(&schedules[i]))
	}
	return result
}
