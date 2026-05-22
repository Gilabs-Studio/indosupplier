package dto

// Holiday DTOs

// CreateHolidayRequest represents the request to create a holiday
type CreateHolidayRequest struct {
	Date              string  `json:"date" binding:"required"`              // YYYY-MM-DD
	Name              string  `json:"name" binding:"required,max=200"`
	Description       string  `json:"description" binding:"max=500"`
	Type              string  `json:"type" binding:"required,oneof=NATIONAL COLLECTIVE COMPANY"`
	IsCollectiveLeave bool    `json:"is_collective_leave"`
	CutsAnnualLeave   bool    `json:"cuts_annual_leave"`
	IsRecurring       bool    `json:"is_recurring"`
	IsActive          bool    `json:"is_active"`
	CompanyID         *string `json:"company_id" binding:"omitempty,uuid"`
}

// UpdateHolidayRequest represents the request to update a holiday
type UpdateHolidayRequest struct {
	Date              *string `json:"date"`
	Name              *string `json:"name" binding:"omitempty,max=200"`
	Description       *string `json:"description" binding:"omitempty,max=500"`
	Type              *string `json:"type" binding:"omitempty,oneof=NATIONAL COLLECTIVE COMPANY"`
	IsCollectiveLeave *bool   `json:"is_collective_leave"`
	CutsAnnualLeave   *bool   `json:"cuts_annual_leave"`
	IsRecurring       *bool   `json:"is_recurring"`
	IsActive          *bool   `json:"is_active"`
	CompanyID         *string `json:"company_id" binding:"omitempty,uuid"`
}

// ListHolidaysRequest represents the request to list holidays
type ListHolidaysRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PerPage   int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search"`
	Year      int    `form:"year" binding:"omitempty,gte=2000,lte=2100"`
	Type      string `form:"type" binding:"omitempty,oneof=NATIONAL COLLECTIVE COMPANY"`
	DateFrom  string `form:"date_from"`
	DateTo    string `form:"date_to"`
	IsActive  *bool  `form:"is_active"`
	CompanyID string `form:"company_id" binding:"omitempty,uuid"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc ASC DESC"`
}

// ImportHolidaysRequest represents the request to import holidays from CSV
type ImportHolidaysRequest struct {
	Year      int    `json:"year" binding:"required,gte=2000,lte=2100"`
	Overwrite bool   `json:"overwrite"` // If true, delete existing holidays for the year
}

// HolidayCSVRow represents a row from CSV import
type HolidayCSVRow struct {
	Date              string `csv:"date"`
	Name              string `csv:"name"`
	Description       string `csv:"description"`
	Type              string `csv:"type"`
	IsCollectiveLeave string `csv:"is_collective_leave"` // "true" or "false"
	CutsAnnualLeave   string `csv:"cuts_annual_leave"`   // "true" or "false"
}

// HolidayResponse represents the response for a holiday
type HolidayResponse struct {
	ID                string  `json:"id"`
	Date              string  `json:"date"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Type              string  `json:"type"`
	Year              int     `json:"year"`
	IsCollectiveLeave bool    `json:"is_collective_leave"`
	CutsAnnualLeave   bool    `json:"cuts_annual_leave"`
	IsRecurring       bool    `json:"is_recurring"`
	IsActive          bool    `json:"is_active"`
	CompanyID         *string `json:"company_id"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// HolidayCalendarResponse represents holidays for calendar view
type HolidayCalendarResponse struct {
	Year     int                         `json:"year"`
	Holidays map[string][]HolidayResponse `json:"holidays"` // Key is "YYYY-MM"
}
