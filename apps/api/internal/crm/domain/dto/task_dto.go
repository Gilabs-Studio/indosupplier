package dto

// CreateTaskRequest represents the request to create a task
type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"required,min=2,max=255"`
	Description string  `json:"description"`
	Type        string  `json:"type" binding:"omitempty,oneof=general call email meeting follow_up"`
	Priority    string  `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	DueDate     *string `json:"due_date"` // YYYY-MM-DD
	AssignedTo  *string `json:"assigned_to" binding:"omitempty,uuid"`
	CustomerID  *string `json:"customer_id" binding:"omitempty,uuid"`
	ContactID   *string `json:"contact_id" binding:"omitempty,uuid"`
	DealID      *string `json:"deal_id" binding:"omitempty,uuid"`
	LeadID      *string `json:"lead_id" binding:"omitempty,uuid"`
}

// UpdateTaskRequest represents the request to update a task (all pointers for partial update)
type UpdateTaskRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=2,max=255"`
	Description *string `json:"description"`
	Type        *string `json:"type" binding:"omitempty,oneof=general call email meeting follow_up"`
	Priority    *string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	Status      *string `json:"status" binding:"omitempty,oneof=pending in_progress completed cancelled"`
	DueDate     *string `json:"due_date"`
	AssignedTo  *string `json:"assigned_to" binding:"omitempty,uuid"`
	CustomerID  *string `json:"customer_id" binding:"omitempty,uuid"`
	ContactID   *string `json:"contact_id" binding:"omitempty,uuid"`
	DealID      *string `json:"deal_id" binding:"omitempty,uuid"`
	LeadID      *string `json:"lead_id" binding:"omitempty,uuid"`
}

// AssignTaskRequest represents the request to assign a task
type AssignTaskRequest struct {
	AssignedTo string `json:"assigned_to" binding:"required,uuid"`
}

// TaskResponse represents the task data returned to client
type TaskResponse struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	DueDate     *string           `json:"due_date"`
	CompletedAt *string           `json:"completed_at"`
	AssignedTo  *string           `json:"assigned_to"`
	AssignedFrom *string          `json:"assigned_from"`
	AssignedEmployee *TaskEmployeeInfo `json:"assigned_employee,omitempty"`
	AssignerEmployee *TaskEmployeeInfo `json:"assigner_employee,omitempty"`
	CustomerID  *string           `json:"customer_id"`
	Customer    *TaskCustomerInfo `json:"customer,omitempty"`
	ContactID   *string           `json:"contact_id"`
	Contact     *TaskContactInfo  `json:"contact,omitempty"`
	DealID      *string           `json:"deal_id"`
	Deal        *TaskDealInfo     `json:"deal,omitempty"`
	LeadID      *string           `json:"lead_id"`
	Lead        *TaskLeadInfo     `json:"lead,omitempty"`
	IsOverdue   bool              `json:"is_overdue"`
	CreatedBy   *string           `json:"created_by"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Reminders   []ReminderResponse `json:"reminders,omitempty"`
}

// TaskEmployeeInfo holds compact employee info for task responses
type TaskEmployeeInfo struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// TaskCustomerInfo holds compact customer info for task responses
type TaskCustomerInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// TaskContactInfo holds compact contact info for task responses
type TaskContactInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// TaskDealInfo holds compact deal info for task responses
type TaskDealInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// TaskLeadInfo holds compact lead info for task responses
type TaskLeadInfo struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// TaskFormDataResponse returns dropdown options for task forms
type TaskFormDataResponse struct {
	Employees []TaskEmployeeOption `json:"employees"`
	Deals     []TaskDealOption     `json:"deals"`
	Leads     []TaskLeadOption     `json:"leads"`
}

// TaskEmployeeOption represents an employee option for task forms
type TaskEmployeeOption struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// TaskDealOption represents a deal option for task forms
type TaskDealOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// TaskLeadOption represents a lead option for task forms
type TaskLeadOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// TaskSummaryResponse is a compact task representation for lead/deal detail responses
type TaskSummaryResponse struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Type             string            `json:"type"`
	Status           string            `json:"status"`
	Priority         string            `json:"priority"`
	DueDate          *string           `json:"due_date"`
	IsOverdue        bool              `json:"is_overdue"`
	AssignedEmployee *TaskEmployeeInfo `json:"assigned_employee,omitempty"`
}

// CreateReminderRequest represents the request to create a reminder
type CreateReminderRequest struct {
	RemindAt     string `json:"remind_at" binding:"required"` // ISO 8601
	ReminderType string `json:"reminder_type" binding:"omitempty,oneof=in_app email"`
	Message      string `json:"message"`
}

// UpdateReminderRequest represents the request to update a reminder
type UpdateReminderRequest struct {
	RemindAt     *string `json:"remind_at"`
	ReminderType *string `json:"reminder_type" binding:"omitempty,oneof=in_app email"`
	Message      *string `json:"message"`
}

// ReminderResponse represents the reminder data returned to client
type ReminderResponse struct {
	ID           string  `json:"id"`
	TaskID       string  `json:"task_id"`
	RemindAt     string  `json:"remind_at"`
	ReminderType string  `json:"reminder_type"`
	IsSent       bool    `json:"is_sent"`
	SentAt       *string `json:"sent_at"`
	Message      string  `json:"message"`
	CreatedBy    *string `json:"created_by"`
	CreatedAt    string  `json:"created_at"`
}
