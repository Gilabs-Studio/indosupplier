package mapper

import (
	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
)

// ToSupplierTypeResponse converts SupplierType model to response DTO
func ToSupplierTypeResponse(m *models.SupplierType) dto.SupplierTypeResponse {
	return dto.SupplierTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToSupplierTypeResponseList converts a slice of SupplierType models to response DTOs
func ToSupplierTypeResponseList(models []models.SupplierType) []dto.SupplierTypeResponse {
	responses := make([]dto.SupplierTypeResponse, len(models))
	for i, m := range models {
		responses[i] = ToSupplierTypeResponse(&m)
	}
	return responses
}
