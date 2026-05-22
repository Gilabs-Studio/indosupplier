package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
)

// ToProvinceResponse converts Province model to ProvinceResponse DTO
func ToProvinceResponse(m *models.Province) *dto.ProvinceResponse {
	if m == nil {
		return nil
	}
	resp := &dto.ProvinceResponse{
		ID:        m.ID,
		CountryID: m.CountryID,
		Name:      m.Name,
		Code:      m.Code,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
	if m.Country != nil {
		resp.Country = ToCountryResponse(m.Country)
	}
	return resp
}

// ToProvinceResponses converts slice of Province models to slice of ProvinceResponse DTOs
func ToProvinceResponses(models []models.Province) []dto.ProvinceResponse {
	responses := make([]dto.ProvinceResponse, len(models))
	for i, m := range models {
		responses[i] = *ToProvinceResponse(&m)
	}
	return responses
}

// ProvinceFromCreateRequest creates Province model from CreateProvinceRequest
func ProvinceFromCreateRequest(req *dto.CreateProvinceRequest) *models.Province {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.Province{
		CountryID: req.CountryID,
		Name:      req.Name,
		Code:      req.Code,
		IsActive:  isActive,
	}
}
