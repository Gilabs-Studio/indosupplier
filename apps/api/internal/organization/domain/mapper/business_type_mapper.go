package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToBusinessTypeResponse converts BusinessType model to BusinessTypeResponse DTO
func ToBusinessTypeResponse(m *models.BusinessType) *dto.BusinessTypeResponse {
	if m == nil {
		return nil
	}
	return &dto.BusinessTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToBusinessTypeResponses converts slice of BusinessType models to slice of BusinessTypeResponse DTOs
func ToBusinessTypeResponses(models []models.BusinessType) []dto.BusinessTypeResponse {
	responses := make([]dto.BusinessTypeResponse, len(models))
	for i, m := range models {
		responses[i] = *ToBusinessTypeResponse(&m)
	}
	return responses
}

// BusinessTypeFromCreateRequest creates BusinessType model from CreateBusinessTypeRequest
func BusinessTypeFromCreateRequest(req *dto.CreateBusinessTypeRequest) *models.BusinessType {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.BusinessType{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}
}
