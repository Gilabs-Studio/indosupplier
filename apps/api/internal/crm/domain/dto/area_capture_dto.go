package dto

import "time"

// CreateAreaCaptureRequest represents the request to capture a GPS location
type CreateAreaCaptureRequest struct {
	VisitReportID *string  `json:"visit_report_id" binding:"omitempty"`
	CaptureType   string   `json:"capture_type" binding:"required,oneof=check_in check_out manual"`
	Latitude      float64  `json:"latitude" binding:"required"`
	Longitude     float64  `json:"longitude" binding:"required"`
	Address       string   `json:"address" binding:"omitempty,max=500"`
	Accuracy      float64  `json:"accuracy" binding:"omitempty"`
	AreaID        *string  `json:"area_id" binding:"omitempty"`
	CapturedAt    string   `json:"captured_at" binding:"omitempty"`
	CapturedBy    *string  `json:"captured_by" binding:"omitempty"`
}

// ListAreaCapturesRequest represents filters for listing captures
type ListAreaCapturesRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PerPage     int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	EmployeeID  string `form:"employee_id" binding:"omitempty"`
	AreaID      string `form:"area_id" binding:"omitempty"`
	CaptureType string `form:"capture_type" binding:"omitempty,oneof=check_in check_out manual"`
	DateFrom    string `form:"date_from" binding:"omitempty"`
	DateTo      string `form:"date_to" binding:"omitempty"`
}

// AreaCaptureResponse represents a single capture point response
type AreaCaptureResponse struct {
	ID            string   `json:"id"`
	VisitReportID *string  `json:"visit_report_id,omitempty"`
	CaptureType   string   `json:"capture_type"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	Address       string   `json:"address"`
	Accuracy      float64  `json:"accuracy"`
	AreaID        *string  `json:"area_id,omitempty"`
	CapturedAt    string   `json:"captured_at"`
	CapturedBy    *string  `json:"captured_by,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

// AreaCoverageResponse represents coverage analysis per area
type AreaCoverageResponse struct {
	AreaID       string  `json:"area_id"`
	AreaName     string  `json:"area_name"`
	AreaCode     string  `json:"area_code"`
	TotalVisits  int     `json:"total_visits"`
	UniquePoints int     `json:"unique_points"`
	LastVisitAt  *string `json:"last_visit_at,omitempty"`
}

// HeatmapPoint represents a single point for the heatmap
type HeatmapPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Intensity int     `json:"intensity"`
}

// HeatmapResponse represents heatmap data
type HeatmapResponse struct {
	Points []HeatmapPoint `json:"points"`
	MaxIntensity int      `json:"max_intensity"`
}

// AreaCaptureTimeFormat is the time format used for parsing capture timestamps
const AreaCaptureTimeFormat = time.RFC3339
