package dto

// BreakTime represents a single break period
type BreakTime struct {
	StartTime string `json:"start_time" binding:"required,len=5"` // HH:MM
	EndTime   string `json:"end_time" binding:"required,len=5"`   // HH:MM
}

// WorkSchedule DTOs

// CreateWorkScheduleRequest represents the request to create a work schedule
type CreateWorkScheduleRequest struct {
	Name                       string      `json:"name" binding:"required,max=100"`
	Description                string      `json:"description" binding:"max=255"`
	DivisionID                 *string     `json:"division_id" binding:"omitempty,uuid"`
	IsDefault                  bool        `json:"is_default"`
	IsActive                   bool        `json:"is_active"`
	StartTime                  string      `json:"start_time" binding:"required,len=5"` // HH:MM
	EndTime                    string      `json:"end_time" binding:"required,len=5"`   // HH:MM
	IsFlexible                 bool        `json:"is_flexible"`
	FlexibleStartTime          string      `json:"flexible_start_time" binding:"omitempty,len=5"`
	FlexibleEndTime            string      `json:"flexible_end_time" binding:"omitempty,len=5"`
	Breaks                     []BreakTime `json:"breaks" binding:"required,min=1,dive"`
	WorkingDays                int         `json:"working_days" binding:"required,gte=1,lte=127"` // Bitmask
	WorkingHoursPerDay         float64     `json:"working_hours_per_day" binding:"omitempty,gt=0,lte=24"`
	LateToleranceMinutes       int         `json:"late_tolerance_minutes" binding:"omitempty,gte=0"`
	EarlyLeaveToleranceMinutes int         `json:"early_leave_tolerance_minutes" binding:"omitempty,gte=0"`
	RequireGPS                 bool        `json:"require_gps"`
	GPSRadiusMeter             float64     `json:"gps_radius_meter" binding:"omitempty,gte=0"`
	OfficeLatitude             float64     `json:"office_latitude" binding:"omitempty"`
	OfficeLongitude            float64     `json:"office_longitude" binding:"omitempty"`
}

// UpdateWorkScheduleRequest represents the request to update a work schedule
type UpdateWorkScheduleRequest struct {
	Name                       *string      `json:"name" binding:"omitempty,max=100"`
	Description                *string      `json:"description" binding:"omitempty,max=255"`
	DivisionID                 *string      `json:"division_id" binding:"omitempty,uuid"`
	IsDefault                  *bool        `json:"is_default"`
	IsActive                   *bool        `json:"is_active"`
	StartTime                  *string      `json:"start_time" binding:"omitempty,len=5"`
	EndTime                    *string      `json:"end_time" binding:"omitempty,len=5"`
	IsFlexible                 *bool        `json:"is_flexible"`
	FlexibleStartTime          *string      `json:"flexible_start_time" binding:"omitempty,len=5"`
	FlexibleEndTime            *string      `json:"flexible_end_time" binding:"omitempty,len=5"`
	Breaks                     *[]BreakTime `json:"breaks" binding:"omitempty,min=1,dive"`
	WorkingDays                *int         `json:"working_days" binding:"omitempty,gte=1,lte=127"`
	WorkingHoursPerDay         *float64     `json:"working_hours_per_day" binding:"omitempty,gt=0,lte=24"`
	LateToleranceMinutes       *int         `json:"late_tolerance_minutes" binding:"omitempty,gte=0"`
	EarlyLeaveToleranceMinutes *int         `json:"early_leave_tolerance_minutes" binding:"omitempty,gte=0"`
	RequireGPS                 *bool        `json:"require_gps"`
	GPSRadiusMeter             *float64     `json:"gps_radius_meter" binding:"omitempty,gte=0"`
	OfficeLatitude             *float64     `json:"office_latitude"`
	OfficeLongitude            *float64     `json:"office_longitude"`
}

// ListWorkSchedulesRequest represents the request to list work schedules
type ListWorkSchedulesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PerPage    int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search"`
	DivisionID string `form:"division_id" binding:"omitempty,uuid"`
	IsActive   *bool  `form:"is_active"`
	IsFlexible *bool  `form:"is_flexible"`
	SortBy     string `form:"sort_by"`
	SortOrder  string `form:"sort_order" binding:"omitempty,oneof=asc desc ASC DESC"`
}

// WorkScheduleResponse represents the response for a work schedule
type WorkScheduleResponse struct {
	ID                         string      `json:"id"`
	Name                       string      `json:"name"`
	Description                string      `json:"description"`
	DivisionID                 *string     `json:"division_id"`
	DivisionName               string      `json:"division_name,omitempty"`
	IsDefault                  bool        `json:"is_default"`
	IsActive                   bool        `json:"is_active"`
	StartTime                  string      `json:"start_time"`
	EndTime                    string      `json:"end_time"`
	IsFlexible                 bool        `json:"is_flexible"`
	FlexibleStartTime          string      `json:"flexible_start_time"`
	FlexibleEndTime            string      `json:"flexible_end_time"`
	Breaks                     []BreakTime `json:"breaks"`
	WorkingDays                int         `json:"working_days"`
	WorkingDaysDisplay         []string    `json:"working_days_display"` // ["Mon", "Tue", ...]
	WorkingHoursPerDay         float64     `json:"working_hours_per_day"`
	LateToleranceMinutes       int         `json:"late_tolerance_minutes"`
	EarlyLeaveToleranceMinutes int         `json:"early_leave_tolerance_minutes"`
	RequireGPS                 bool        `json:"require_gps"`
	GPSRadiusMeter             float64     `json:"gps_radius_meter"`
	OfficeLatitude             float64     `json:"office_latitude"`
	OfficeLongitude            float64     `json:"office_longitude"`
	CreatedAt                  string      `json:"created_at"`
	UpdatedAt                  string      `json:"updated_at"`
}

// CompanyFormOption represents a company option for GPS coordinate picker
type CompanyFormOption struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// WorkScheduleFormDataResponse represents the form data for work schedule dropdowns
type WorkScheduleFormDataResponse struct {
	Divisions []DivisionFormOption `json:"divisions"`
	Companies []CompanyFormOption  `json:"companies"`
}
