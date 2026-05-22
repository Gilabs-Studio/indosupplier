package dto

// CreateSalesTargetRequest represents the request to create a sales target
type CreateSalesTargetRequest struct {
	EmployeeID  string                              `json:"employee_id" binding:"required"`
	Year        int                                 `json:"year" binding:"required,min=2020,max=2100"`
	TotalTarget float64                             `json:"total_target" binding:"gte=0"`
	Notes       string                              `json:"notes"`
	Months      []CreateMonthlySalesTargetRequest   `json:"months" binding:"required,len=12,dive"`
}

// CreateMonthlySalesTargetRequest represents a monthly target in the create request
type CreateMonthlySalesTargetRequest struct {
	Month        int     `json:"month" binding:"required,min=1,max=12"`
	TargetAmount float64 `json:"target_amount" binding:"gte=0"`
	Notes        string  `json:"notes"`
}

// UpdateSalesTargetRequest represents the request to update a sales target
type UpdateSalesTargetRequest struct {
	TotalTarget *float64                             `json:"total_target" binding:"omitempty,gte=0"`
	Notes       *string                              `json:"notes"`
	Months      *[]UpdateMonthlySalesTargetRequest   `json:"months" binding:"omitempty,len=12,dive"`
}

// UpdateMonthlySalesTargetRequest represents a monthly target in the update request
type UpdateMonthlySalesTargetRequest struct {
	Month        int     `json:"month" binding:"required,min=1,max=12"`
	TargetAmount float64 `json:"target_amount" binding:"gte=0"`
	Notes        string  `json:"notes"`
}

// ListSalesTargetsRequest represents the request to list sales targets
type ListSalesTargetsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PerPage    int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Year       *int   `form:"year"`
	EmployeeID string `form:"employee_id"`
	Search     string `form:"search"`
	SortBy     string `form:"sort_by"`
	SortDir    string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
	// Note: Scope is determined by permission middleware (RequirePermission),
	// not by query params. These fields are for reference/logging only.
	Scope   string `form:"scope"`   // Optional: reflect the resolved scope (OWN, AREA, DIVISION, ALL)
	ScopeID string `form:"scope_id"` // Optional: area_id or division_id for scoped queries
}

// ListAvailableSalesTargetEmployeesRequest represents request to list employees
// that do not yet have a sales target for a given year.
type ListAvailableSalesTargetEmployeesRequest struct {
	Year              int    `form:"year" binding:"required,min=2020,max=2100"`
	IncludeEmployeeID string `form:"include_employee_id"`
}

// SalesTargetResponse represents the response for a sales target
type SalesTargetResponse struct {
	ID                 string                       `json:"id"`
	EmployeeID         string                       `json:"employee_id"`
	Employee           *EmployeeResponse            `json:"employee,omitempty"`
	Year               int                          `json:"year"`
	TotalTarget        float64                      `json:"total_target"`
	TotalActual        float64                      `json:"total_actual"`
	AchievementPercent float64                      `json:"achievement_percent"`
	Notes              string                       `json:"notes"`

	MonthlyTargets []MonthlySalesTargetResponse `json:"monthly_targets,omitempty"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MonthlySalesTargetResponse represents the response for a monthly sales target
type MonthlySalesTargetResponse struct {
	ID                 string  `json:"id"`
	Month              int     `json:"month"`
	MonthName          string  `json:"month_name"`
	TargetAmount       float64 `json:"target_amount"`
	ActualAmount       float64 `json:"actual_amount"`
	AchievementPercent float64 `json:"achievement_percent"`
	Notes              string  `json:"notes"`
}


