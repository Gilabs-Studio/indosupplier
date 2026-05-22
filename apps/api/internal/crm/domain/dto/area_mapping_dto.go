package dto

import "time"

// GetAreaMappingRequest contains query filters for area mapping.
type GetAreaMappingRequest struct {
	Month *int `form:"month"`
	Year  *int `form:"year"`
}

// AreaMappingLeadData represents a lead on the area map
type AreaMappingLeadData struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // "lead"
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Province  string  `json:"province"`
	City      string  `json:"city"`
	// Lead specific info
	LeadStatus   string  `json:"lead_status"`
	LeadScore    int     `json:"lead_score"`
	EstimatedVal float64 `json:"estimated_value"`
	AssignedTo   *string `json:"assigned_to,omitempty"`
	AssignedName *string `json:"assigned_name,omitempty"`
	// Activity metrics for intensity visualization
	ActivityCount  int        `json:"activity_count"`
	TaskCount      int        `json:"task_count"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	// Intensity score (0-100) for color intensity
	IntensityScore float64 `json:"intensity_score"`
}

// AreaMappingPipelineData represents a pipeline deal on the area map
type AreaMappingPipelineData struct {
	ID                string     `json:"id"`
	Code              string     `json:"code"`
	Title             string     `json:"title"`
	Type              string     `json:"type"` // "pipeline"
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	Province          string     `json:"province"`
	City              string     `json:"city"`
	PipelineStageID   string     `json:"pipeline_stage_id"`
	PipelineStageName string     `json:"pipeline_stage_name"`
	Status            string     `json:"status"`
	Value             float64    `json:"value"`
	Probability       int        `json:"probability"`
	ExpectedCloseDate *time.Time `json:"expected_close_date,omitempty"`
	AssignedTo        *string    `json:"assigned_to,omitempty"`
	AssignedName      *string    `json:"assigned_name,omitempty"`
	LeadID            *string    `json:"lead_id,omitempty"`
	LeadName          *string    `json:"lead_name,omitempty"`
	IntensityScore    float64    `json:"intensity_score"`
}

// AreaMappingItem represents either a lead or pipeline item (union type)
type AreaMappingItem struct {
	// Type: "lead" or "pipeline"
	Type     string                   `json:"type"`
	Lead     *AreaMappingLeadData     `json:"lead,omitempty"`
	Pipeline *AreaMappingPipelineData `json:"pipeline,omitempty"`
}

// AreaMappingSummary aggregated stats
type AreaMappingSummary struct {
	TotalLeads         int     `json:"total_leads"`
	TotalPipelines     int     `json:"total_pipelines"`
	TotalActivities    int     `json:"total_activities"`
	TotalPipelineValue float64 `json:"total_pipeline_value"`
	MaxIntensityScore  float64 `json:"max_intensity_score"`
	MinIntensityScore  float64 `json:"min_intensity_score"`
}

// AreaMappingClusterResponse contains aggregated cluster points per city.
type AreaMappingClusterResponse struct {
	City               string  `json:"city"`
	TotalPoints        int     `json:"total_points"`
	LeadCount          int     `json:"lead_count"`
	PipelineDealCount  int     `json:"pipeline_deal_count"`
	TotalPipelineValue float64 `json:"total_pipeline_value"`
	AvgIntensity       float64 `json:"avg_intensity"`
	MaxIntensity       float64 `json:"max_intensity"`
	CenterLat          float64 `json:"center_lat"`
	CenterLng          float64 `json:"center_lng"`
}

// AreaMappingFilterMeta contains the effective filter used by the response.
type AreaMappingFilterMeta struct {
	Month *int `json:"month,omitempty"`
	Year  *int `json:"year,omitempty"`
}

// AreaMappingResponse represents the full response for area mapping
type AreaMappingResponse struct {
	Items    []AreaMappingItem            `json:"items"`
	Clusters []AreaMappingClusterResponse `json:"clusters"`
	Summary  AreaMappingSummary           `json:"summary"`
	Filters  AreaMappingFilterMeta        `json:"filters"`
}
