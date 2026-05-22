package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToProductBrandResponse converts ProductBrand model to response DTO
func ToProductBrandResponse(m *models.ProductBrand) dto.ProductBrandResponse {
	return dto.ProductBrandResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToProductBrandResponseList converts a slice of ProductBrand models to response DTOs
func ToProductBrandResponseList(models []models.ProductBrand) []dto.ProductBrandResponse {
	responses := make([]dto.ProductBrandResponse, len(models))
	for i, m := range models {
		responses[i] = ToProductBrandResponse(&m)
	}
	return responses
}
