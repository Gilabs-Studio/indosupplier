package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToBusinessUnitResponse converts BusinessUnit model to BusinessUnitResponse DTO
func ToBusinessUnitResponse(m *models.BusinessUnit) *dto.BusinessUnitResponse {
	if m == nil {
		return nil
	}
	return &dto.BusinessUnitResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToBusinessUnitResponses converts slice of BusinessUnit models to slice of BusinessUnitResponse DTOs
func ToBusinessUnitResponses(models []models.BusinessUnit) []dto.BusinessUnitResponse {
	responses := make([]dto.BusinessUnitResponse, len(models))
	for i, m := range models {
		responses[i] = *ToBusinessUnitResponse(&m)
	}
	return responses
}

// BusinessUnitFromCreateRequest creates BusinessUnit model from CreateBusinessUnitRequest
func BusinessUnitFromCreateRequest(req *dto.CreateBusinessUnitRequest) *models.BusinessUnit {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.BusinessUnit{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}
}
