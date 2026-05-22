package mapper

import (
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToTaskResponse converts a Task model to TaskResponse
func ToTaskResponse(task *models.Task) dto.TaskResponse {
	resp := dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Type:        task.Type,
		Status:      task.Status,
		Priority:    task.Priority,
		AssignedTo:  task.AssignedTo,
		AssignedFrom: task.AssignedFrom,
		CustomerID:  task.CustomerID,
		ContactID:   task.ContactID,
		DealID:      task.DealID,
		LeadID:      task.LeadID,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05+07:00"),
	}

	if task.DueDate != nil {
		d := task.DueDate.Format("2006-01-02")
		resp.DueDate = &d
		// Determine overdue status
		if task.Status != string(models.TaskStatusCompleted) && task.Status != string(models.TaskStatusCancelled) {
			resp.IsOverdue = task.DueDate.Before(apptime.Now())
		}
	}

	if task.CompletedAt != nil {
		t := task.CompletedAt.Format("2006-01-02T15:04:05+07:00")
		resp.CompletedAt = &t
	}

	if task.AssignedEmployee != nil {
		resp.AssignedEmployee = &dto.TaskEmployeeInfo{
			ID:           task.AssignedEmployee.ID,
			EmployeeCode: task.AssignedEmployee.EmployeeCode,
			Name:         task.AssignedEmployee.Name,
		}
	}

	if task.AssignerEmployee != nil {
		resp.AssignerEmployee = &dto.TaskEmployeeInfo{
			ID:           task.AssignerEmployee.ID,
			EmployeeCode: task.AssignerEmployee.EmployeeCode,
			Name:         task.AssignerEmployee.Name,
		}
	}

	if task.Customer != nil {
		resp.Customer = &dto.TaskCustomerInfo{
			ID: task.Customer.ID, Code: task.Customer.Code, Name: task.Customer.Name,
		}
	}

	if task.Contact != nil {
		resp.Contact = &dto.TaskContactInfo{
			ID: task.Contact.ID, Name: task.Contact.Name, Email: task.Contact.Email,
		}
	}

	if task.Deal != nil {
		resp.Deal = &dto.TaskDealInfo{
			ID: task.Deal.ID, Code: task.Deal.Code, Name: task.Deal.Title,
		}
	}

	if task.Lead != nil {
		resp.Lead = &dto.TaskLeadInfo{
			ID: task.Lead.ID, Code: task.Lead.Code,
			FirstName: task.Lead.FirstName, LastName: task.Lead.LastName,
		}
	}

	if len(task.Reminders) > 0 {
		resp.Reminders = ToReminderResponseList(task.Reminders)
	}

	return resp
}

// ToTaskResponseList converts a slice of Task models to responses
func ToTaskResponseList(tasks []models.Task) []dto.TaskResponse {
	result := make([]dto.TaskResponse, 0, len(tasks))
	for i := range tasks {
		result = append(result, ToTaskResponse(&tasks[i]))
	}
	return result
}

// ToTaskSummaryResponse converts a Task model to a compact TaskSummaryResponse
func ToTaskSummaryResponse(task *models.Task) dto.TaskSummaryResponse {
	resp := dto.TaskSummaryResponse{
		ID:       task.ID,
		Title:    task.Title,
		Type:     task.Type,
		Status:   task.Status,
		Priority: task.Priority,
	}

	if task.DueDate != nil {
		d := task.DueDate.Format("2006-01-02")
		resp.DueDate = &d
		if task.Status != string(models.TaskStatusCompleted) && task.Status != string(models.TaskStatusCancelled) {
			resp.IsOverdue = task.DueDate.Before(apptime.Now())
		}
	}

	if task.AssignedEmployee != nil {
		resp.AssignedEmployee = &dto.TaskEmployeeInfo{
			ID:           task.AssignedEmployee.ID,
			EmployeeCode: task.AssignedEmployee.EmployeeCode,
			Name:         task.AssignedEmployee.Name,
		}
	}

	return resp
}

// ToTaskSummaryResponseList converts a slice of Task models to summary responses
func ToTaskSummaryResponseList(tasks []models.Task) []dto.TaskSummaryResponse {
	result := make([]dto.TaskSummaryResponse, 0, len(tasks))
	for i := range tasks {
		result = append(result, ToTaskSummaryResponse(&tasks[i]))
	}
	return result
}

// ToReminderResponse converts a Reminder model to ReminderResponse
func ToReminderResponse(reminder *models.Reminder) dto.ReminderResponse {
	resp := dto.ReminderResponse{
		ID:           reminder.ID,
		TaskID:       reminder.TaskID,
		RemindAt:     reminder.RemindAt.Format("2006-01-02T15:04:05+07:00"),
		ReminderType: reminder.ReminderType,
		IsSent:       reminder.IsSent,
		Message:      reminder.Message,
		CreatedBy:    reminder.CreatedBy,
		CreatedAt:    reminder.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
	}

	if reminder.SentAt != nil {
		t := reminder.SentAt.Format("2006-01-02T15:04:05+07:00")
		resp.SentAt = &t
	}

	return resp
}

// ToReminderResponseList converts a slice of Reminder models to responses
func ToReminderResponseList(reminders []models.Reminder) []dto.ReminderResponse {
	result := make([]dto.ReminderResponse, 0, len(reminders))
	for i := range reminders {
		result = append(result, ToReminderResponse(&reminders[i]))
	}
	return result
}
