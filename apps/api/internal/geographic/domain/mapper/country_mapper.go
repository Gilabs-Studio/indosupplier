package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
)

// ToCountryResponse converts Country model to CountryResponse DTO
func ToCountryResponse(m *models.Country) *dto.CountryResponse {
	if m == nil {
		return nil
	}
	return &dto.CountryResponse{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		PhoneCode: m.PhoneCode,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToCountryResponses converts slice of Country models to slice of CountryResponse DTOs
func ToCountryResponses(models []models.Country) []dto.CountryResponse {
	responses := make([]dto.CountryResponse, len(models))
	for i, m := range models {
		responses[i] = *ToCountryResponse(&m)
	}
	return responses
}

// CountryFromCreateRequest creates Country model from CreateCountryRequest
func CountryFromCreateRequest(req *dto.CreateCountryRequest) *models.Country {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.Country{
		Name:      req.Name,
		Code:      req.Code,
		PhoneCode: req.PhoneCode,
		IsActive:  isActive,
	}
}
