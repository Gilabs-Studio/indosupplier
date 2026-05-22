package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToDivisionResponse converts Division model to DivisionResponse DTO
func ToDivisionResponse(m *models.Division) *dto.DivisionResponse {
	if m == nil {
		return nil
	}
	return &dto.DivisionResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToDivisionResponses converts slice of Division models to slice of DivisionResponse DTOs
func ToDivisionResponses(models []models.Division) []dto.DivisionResponse {
	responses := make([]dto.DivisionResponse, len(models))
	for i, m := range models {
		responses[i] = *ToDivisionResponse(&m)
	}
	return responses
}

// DivisionFromCreateRequest creates Division model from CreateDivisionRequest
func DivisionFromCreateRequest(req *dto.CreateDivisionRequest) *models.Division {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.Division{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}
}
