package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToProductTypeResponse converts ProductType model to response DTO
func ToProductTypeResponse(m *models.ProductType) dto.ProductTypeResponse {
	return dto.ProductTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToProductTypeResponseList converts a slice of ProductType models to response DTOs
func ToProductTypeResponseList(models []models.ProductType) []dto.ProductTypeResponse {
	responses := make([]dto.ProductTypeResponse, len(models))
	for i, m := range models {
		responses[i] = ToProductTypeResponse(&m)
	}
	return responses
}
