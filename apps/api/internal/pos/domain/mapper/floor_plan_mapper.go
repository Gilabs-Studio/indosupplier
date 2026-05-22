package mapper

import (
	"encoding/json"
	"time"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

const timeFormat = time.RFC3339

// ToFloorPlanResponse converts model to response DTO
func ToFloorPlanResponse(m *models.FloorPlan) *dto.FloorPlanResponse {
	resp := &dto.FloorPlanResponse{
		ID:          m.ID,
		OutletID:    m.OutletID,
		CompanyID:   m.CompanyID,
		Name:        m.Name,
		FloorNumber: m.FloorNumber,
		Status:      m.Status,
		GridSize:    m.GridSize,
		SnapToGrid:  m.SnapToGrid,
		Width:       m.Width,
		Height:      m.Height,
		LayoutData:  json.RawMessage(m.LayoutData),
		Version:     m.Version,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt.Format(timeFormat),
		UpdatedAt:   m.UpdatedAt.Format(timeFormat),
	}

	if m.PublishedAt != nil {
		t := m.PublishedAt.Format(timeFormat)
		resp.PublishedAt = &t
	}
	resp.PublishedBy = m.PublishedBy

	return resp
}

// ToFloorPlanListResponse converts a slice of models to response DTOs
func ToFloorPlanListResponse(plans []models.FloorPlan) []dto.FloorPlanResponse {
	result := make([]dto.FloorPlanResponse, 0, len(plans))
	for i := range plans {
		result = append(result, *ToFloorPlanResponse(&plans[i]))
	}
	return result
}

// ToLayoutVersionResponse converts model to response DTO
func ToLayoutVersionResponse(m *models.LayoutVersion) *dto.LayoutVersionResponse {
	return &dto.LayoutVersionResponse{
		ID:              m.ID,
		FloorPlanID:     m.FloorPlanID,
		Version:         m.Version,
		LayoutData:      json.RawMessage(m.LayoutData),
		PublishedAt:     m.PublishedAt.Format(timeFormat),
		PublishedBy:     m.PublishedBy,
		PublishedByName: m.PublishedByName,
	}
}

// ToLayoutVersionListResponse converts a slice of models to response DTOs
func ToLayoutVersionListResponse(versions []models.LayoutVersion) []dto.LayoutVersionResponse {
	result := make([]dto.LayoutVersionResponse, 0, len(versions))
	for i := range versions {
		result = append(result, *ToLayoutVersionResponse(&versions[i]))
	}
	return result
}
