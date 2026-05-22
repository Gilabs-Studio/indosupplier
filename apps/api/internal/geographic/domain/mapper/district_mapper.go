package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
)

// ToDistrictResponse converts District model to DistrictResponse DTO
func ToDistrictResponse(m *models.District) *dto.DistrictResponse {
	if m == nil {
		return nil
	}
	resp := &dto.DistrictResponse{
		ID:        m.ID,
		CityID:    m.CityID,
		Name:      m.Name,
		Code:      m.Code,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
	if m.City != nil {
		resp.City = ToCityResponse(m.City)
	}
	return resp
}

// ToDistrictResponses converts slice of District models to slice of DistrictResponse DTOs
func ToDistrictResponses(models []models.District) []dto.DistrictResponse {
	responses := make([]dto.DistrictResponse, len(models))
	for i, m := range models {
		responses[i] = *ToDistrictResponse(&m)
	}
	return responses
}

// DistrictFromCreateRequest creates District model from CreateDistrictRequest
func DistrictFromCreateRequest(req *dto.CreateDistrictRequest) *models.District {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.District{
		CityID:   req.CityID,
		Name:     req.Name,
		Code:     req.Code,
		IsActive: isActive,
	}
}
