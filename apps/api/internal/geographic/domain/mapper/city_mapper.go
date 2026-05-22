package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
)

// ToCityResponse converts City model to CityResponse DTO
func ToCityResponse(m *models.City) *dto.CityResponse {
	if m == nil {
		return nil
	}
	resp := &dto.CityResponse{
		ID:         m.ID,
		ProvinceID: m.ProvinceID,
		Name:       m.Name,
		Code:       m.Code,
		Type:       m.Type,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
	if m.Province != nil {
		resp.Province = ToProvinceResponse(m.Province)
	}
	return resp
}

// ToCityResponses converts slice of City models to slice of CityResponse DTOs
func ToCityResponses(models []models.City) []dto.CityResponse {
	responses := make([]dto.CityResponse, len(models))
	for i, m := range models {
		responses[i] = *ToCityResponse(&m)
	}
	return responses
}

// CityFromCreateRequest creates City model from CreateCityRequest
func CityFromCreateRequest(req *dto.CreateCityRequest) *models.City {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	cityType := "city"
	if req.Type != "" {
		cityType = req.Type
	}
	return &models.City{
		ProvinceID: req.ProvinceID,
		Name:       req.Name,
		Code:       req.Code,
		Type:       cityType,
		IsActive:   isActive,
	}
}
