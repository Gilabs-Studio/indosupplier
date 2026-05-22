package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// CreateFloorPlanRequest for creating a new floor plan
type CreateFloorPlanRequest struct {
	OutletID    string `json:"outlet_id" binding:"omitempty,uuid"`
	CompanyID   string `json:"company_id,omitempty" binding:"omitempty,uuid"` // deprecated: use outlet_id
	Name        string `json:"name" binding:"required,max=200"`
	FloorNumber int    `json:"floor_number" binding:"required,min=1"`
	GridSize    *int   `json:"grid_size,omitempty"`
	SnapToGrid  *bool  `json:"snap_to_grid,omitempty"`
	Width       *int   `json:"width,omitempty"`
	Height      *int   `json:"height,omitempty"`
}

// UpdateFloorPlanRequest for updating a floor plan
type UpdateFloorPlanRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=200"`
	FloorNumber *int    `json:"floor_number,omitempty" binding:"omitempty,min=1"`
	GridSize    *int    `json:"grid_size,omitempty"`
	SnapToGrid  *bool   `json:"snap_to_grid,omitempty"`
	Width       *int    `json:"width,omitempty"`
	Height      *int    `json:"height,omitempty"`
}

// SaveLayoutDataRequest for saving canvas layout data
type SaveLayoutDataRequest struct {
	LayoutData json.RawMessage `json:"layout_data" binding:"required"`
}

// FloorPlanResponse returned to client
type FloorPlanResponse struct {
	ID          string          `json:"id"`
	OutletID    string          `json:"outlet_id"`
	CompanyID   *string         `json:"company_id,omitempty"` // deprecated: kept for backward compatibility
	Name        string          `json:"name"`
	FloorNumber int             `json:"floor_number"`
	Status      string          `json:"status"`
	GridSize    int             `json:"grid_size"`
	SnapToGrid  bool            `json:"snap_to_grid"`
	Width       int             `json:"width"`
	Height      int             `json:"height"`
	LayoutData  json.RawMessage `json:"layout_data"`
	Version     int             `json:"version"`
	PublishedAt *string         `json:"published_at"`
	PublishedBy *string         `json:"published_by"`
	CreatedBy   *string         `json:"created_by"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// LayoutVersionResponse returned to client
type LayoutVersionResponse struct {
	ID              string          `json:"id"`
	FloorPlanID     string          `json:"floor_plan_id"`
	Version         int             `json:"version"`
	LayoutData      json.RawMessage `json:"layout_data"`
	PublishedAt     string          `json:"published_at"`
	PublishedBy     string          `json:"published_by"`
	PublishedByName string          `json:"published_by_name"`
}

// OutletOption for form data select
type OutletOption struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	HasFloorPlan bool      `json:"has_floor_plan"`
}

// FloorPlanFormDataResponse for form dropdown options
type FloorPlanFormDataResponse struct {
	Outlets []OutletOption `json:"outlets"`
}

// ─── Table QR Token DTOs ─────────────────────────────────────────────────────

// GenerateTableTokenRequest requests creation (or rotation) of a QR token for a table.
type GenerateTableTokenRequest struct {
	TableLabel string `json:"table_label" binding:"required,max=50"`
}

// TableQRTokenResponse returned to the staff client.
type TableQRTokenResponse struct {
	ID            string `json:"id"`
	FloorPlanID   string `json:"floor_plan_id"`
	TableObjectID string `json:"table_object_id"`
	TableLabel    string `json:"table_label"`
	Token         string `json:"token"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
	// SelfOrderURL is the public URL customers use to scan and place orders.
	// Resolved dynamically from the request origin so it works in both dev and prod.
	SelfOrderURL string `json:"self_order_url,omitempty"`
	// QRBase64 is a data URI (data:image/png;base64,...) ready for <img src> or print.
	QRBase64 string `json:"qr_base64,omitempty"`
}

