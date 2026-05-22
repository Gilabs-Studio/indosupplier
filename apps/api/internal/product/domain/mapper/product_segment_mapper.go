package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToProductSegmentResponse converts ProductSegment model to response DTO
func ToProductSegmentResponse(m *models.ProductSegment) dto.ProductSegmentResponse {
	return dto.ProductSegmentResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToProductSegmentResponseList converts a slice of ProductSegment models to response DTOs
func ToProductSegmentResponseList(models []models.ProductSegment) []dto.ProductSegmentResponse {
	responses := make([]dto.ProductSegmentResponse, len(models))
	for i, m := range models {
		responses[i] = ToProductSegmentResponse(&m)
	}
	return responses
}
