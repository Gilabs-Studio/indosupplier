package mapper

import (
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
)

// ToProcurementTypeResponse converts ProcurementType model to response DTO
func ToProcurementTypeResponse(m *models.ProcurementType) dto.ProcurementTypeResponse {
	return dto.ProcurementTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToProcurementTypeResponseList converts a slice of ProcurementType models to response DTOs
func ToProcurementTypeResponseList(models []models.ProcurementType) []dto.ProcurementTypeResponse {
	responses := make([]dto.ProcurementTypeResponse, len(models))
	for i, m := range models {
		responses[i] = ToProcurementTypeResponse(&m)
	}
	return responses
}
