package mapper

import (
	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

// ToSOSourceResponse converts SOSource model to response DTO
func ToSOSourceResponse(m *models.SOSource) dto.SOSourceResponse {
	return dto.SOSourceResponse{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToSOSourceResponseList converts a slice of SOSource models to response DTOs
func ToSOSourceResponseList(models []models.SOSource) []dto.SOSourceResponse {
	responses := make([]dto.SOSourceResponse, len(models))
	for i, m := range models {
		responses[i] = ToSOSourceResponse(&m)
	}
	return responses
}
