package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToPackagingResponse converts Packaging model to response DTO
func ToPackagingResponse(m *models.Packaging) dto.PackagingResponse {
	return dto.PackagingResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToPackagingResponseList converts a slice of Packaging models to response DTOs
func ToPackagingResponseList(models []models.Packaging) []dto.PackagingResponse {
	responses := make([]dto.PackagingResponse, len(models))
	for i, m := range models {
		responses[i] = ToPackagingResponse(&m)
	}
	return responses
}
