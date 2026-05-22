package dto

import "time"

type UpCountryCostEmployeeRequest struct {
	EmployeeID string `json:"employee_id" binding:"required,uuid"`
}

type UpCountryCostItemRequest struct {
	CostType    string  `json:"cost_type" binding:"required"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ExpenseDate string  `json:"expense_date"` // optional, YYYY-MM-DD
}

type CreateUpCountryCostRequest struct {
	Purpose   string                         `json:"purpose" binding:"required"`
	Location  string                         `json:"location"`
	StartDate string                         `json:"start_date" binding:"required"`
	EndDate   string                         `json:"end_date" binding:"required"`
	Notes     string                         `json:"notes"`
	Employees []UpCountryCostEmployeeRequest `json:"employees" binding:"required,min=1"`
	Items     []UpCountryCostItemRequest     `json:"items" binding:"required,min=1"`
}

type UpdateUpCountryCostRequest struct {
	Purpose   string                         `json:"purpose" binding:"required"`
	Location  string                         `json:"location"`
	StartDate string                         `json:"start_date" binding:"required"`
	EndDate   string                         `json:"end_date" binding:"required"`
	Notes     string                         `json:"notes"`
	Employees []UpCountryCostEmployeeRequest `json:"employees" binding:"required,min=1"`
	Items     []UpCountryCostItemRequest     `json:"items" binding:"required,min=1"`
}

type RejectUpCountryCostRequest struct {
	Comment string `json:"comment"`
}

type ListUpCountryCostsRequest struct {
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PerPage    int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string  `form:"search"`
	StartDate  *string `form:"start_date"`
	EndDate    *string `form:"end_date"`
	Status     *string `form:"status"`
	EmployeeID *string `form:"employee_id"`
	SortBy     string  `form:"sort_by"`
	SortDir    string  `form:"sort_dir"`
}

type UpCountryCostEmployeeResponse struct {
	ID         string `json:"id"`
	EmployeeID string `json:"employee_id"`
}

type UpCountryCostItemResponse struct {
	ID          string     `json:"id"`
	CostType    string     `json:"cost_type"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	ExpenseDate *time.Time `json:"expense_date"`
}

type UpCountryCostResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Purpose   string `json:"purpose"`
	Location  string `json:"location"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	Notes     string    `json:"notes"`

	Employees   []UpCountryCostEmployeeResponse `json:"employees"`
	Items       []UpCountryCostItemResponse     `json:"items"`
	TotalAmount float64                         `json:"total_amount"`

	// Submission
	SubmittedAt *time.Time `json:"submitted_at"`
	SubmittedBy *string    `json:"submitted_by"`

	// Manager approval
	ManagerApprovedAt *time.Time `json:"manager_approved_at"`
	ManagerApprovedBy *string    `json:"manager_approved_by"`
	ManagerComment    string     `json:"manager_comment"`

	// Finance approval
	FinanceApprovedAt *time.Time `json:"finance_approved_at"`
	FinanceApprovedBy *string    `json:"finance_approved_by"`

	// Payment
	PaidAt *time.Time `json:"paid_at"`
	PaidBy *string    `json:"paid_by"`

	CreatedBy *string   `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpCountryCostStatsResponse struct {
	TotalRequests   int64   `json:"total_requests"`
	PendingApproval int64   `json:"pending_approval"` // submitted + manager_approved
	Approved        int64   `json:"approved"`         // finance_approved + paid
	TotalAmount     float64 `json:"total_amount"`
}
