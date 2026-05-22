package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToUnitOfMeasureResponse converts UnitOfMeasure model to response DTO
func ToUnitOfMeasureResponse(m *models.UnitOfMeasure) dto.UnitOfMeasureResponse {
	return dto.UnitOfMeasureResponse{
		ID:          m.ID,
		Name:        m.Name,
		Symbol:      m.Symbol,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToUnitOfMeasureResponseList converts a slice of UnitOfMeasure models to response DTOs
func ToUnitOfMeasureResponseList(models []models.UnitOfMeasure) []dto.UnitOfMeasureResponse {
	responses := make([]dto.UnitOfMeasureResponse, len(models))
	for i, m := range models {
		responses[i] = ToUnitOfMeasureResponse(&m)
	}
	return responses
}
