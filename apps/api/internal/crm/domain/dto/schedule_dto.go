package dto

// CreateScheduleRequest represents the request to create a schedule
type CreateScheduleRequest struct {
	TaskID                *string `json:"task_id" binding:"omitempty,uuid"`
	EmployeeID            string  `json:"employee_id" binding:"required,uuid"`
	Title                 string  `json:"title" binding:"required,min=2,max=255"`
	Description           string  `json:"description"`
	ScheduledAt           string  `json:"scheduled_at" binding:"required"` // ISO 8601
	EndAt                 *string `json:"end_at"`                          // ISO 8601
	ReminderMinutesBefore *int    `json:"reminder_minutes_before" binding:"omitempty,min=0,max=10080"`
}

// UpdateScheduleRequest represents the request to update a schedule (all pointers)
type UpdateScheduleRequest struct {
	TaskID                *string `json:"task_id" binding:"omitempty,uuid"`
	EmployeeID            *string `json:"employee_id" binding:"omitempty,uuid"`
	Title                 *string `json:"title" binding:"omitempty,min=2,max=255"`
	Description           *string `json:"description"`
	ScheduledAt           *string `json:"scheduled_at"`
	EndAt                 *string `json:"end_at"`
	Status                *string `json:"status" binding:"omitempty,oneof=pending confirmed completed cancelled"`
	ReminderMinutesBefore *int    `json:"reminder_minutes_before" binding:"omitempty,min=0,max=10080"`
}

// ScheduleResponse represents the schedule data returned to client
type ScheduleResponse struct {
	ID                    string               `json:"id"`
	TaskID                *string              `json:"task_id"`
	Task                  *ScheduleTaskInfo    `json:"task,omitempty"`
	EmployeeID            string               `json:"employee_id"`
	Employee              *ScheduleEmployeeInfo `json:"employee,omitempty"`
	Title                 string               `json:"title"`
	Description           string               `json:"description"`
	ScheduledAt           string               `json:"scheduled_at"`
	EndAt                 *string              `json:"end_at"`
	Status                string               `json:"status"`
	ReminderMinutesBefore int                  `json:"reminder_minutes_before"`
	CreatedBy             *string              `json:"created_by"`
	CreatedAt             string               `json:"created_at"`
	UpdatedAt             string               `json:"updated_at"`
}

// ScheduleTaskInfo holds compact task info for schedule responses
type ScheduleTaskInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// ScheduleEmployeeInfo holds compact employee info for schedule responses
type ScheduleEmployeeInfo struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// ScheduleFormDataResponse returns dropdown options for schedule forms
type ScheduleFormDataResponse struct {
	Employees []ScheduleEmployeeOption `json:"employees"`
	Tasks     []ScheduleTaskOption     `json:"tasks"`
}

// ScheduleEmployeeOption for form dropdowns
type ScheduleEmployeeOption struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// ScheduleTaskOption for form dropdowns
type ScheduleTaskOption struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}
