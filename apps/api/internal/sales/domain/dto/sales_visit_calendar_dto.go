package dto

// GetCalendarSummaryRequest represents the request to get calendar summary
type GetCalendarSummaryRequest struct {
	DateFrom   string `form:"date_from" binding:"required"`
	DateTo     string `form:"date_to" binding:"required"`
	EmployeeID string `form:"employee_id" binding:"omitempty,uuid"`
	CompanyID  string `form:"company_id" binding:"omitempty,uuid"`
}

// CalendarPreviewItem represents a brief visit info for calendar card
type CalendarPreviewItem struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	ScheduledTime string `json:"scheduled_time"`
	CustomerName  string `json:"customer_name"`
	Status        string `json:"status"`
}

// CalendarDaySummary represents summary for a single day
type CalendarDaySummary struct {
	Date         string                `json:"date"`
	TotalCount   int                   `json:"total_count"`
	Planned      int                   `json:"planned"`
	InProgress   int                   `json:"in_progress"`
	Completed    int                   `json:"completed"`
	Cancelled    int                   `json:"cancelled"`
	PreviewItems []CalendarPreviewItem `json:"preview_items" gorm:"-"`
}

// CalendarSummaryResponse represents the response for calendar summary
type CalendarSummaryResponse struct {
	Summary []CalendarDaySummary `json:"summary"`
}
