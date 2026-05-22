package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToLeadSourceResponse converts a LeadSource model to a response DTO
func ToLeadSourceResponse(m *models.LeadSource) dto.LeadSourceResponse {
	return dto.LeadSourceResponse{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		Order:       m.Order,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToLeadSourceResponseList converts a slice of LeadSource models to response DTOs
func ToLeadSourceResponseList(sources []models.LeadSource) []dto.LeadSourceResponse {
	responses := make([]dto.LeadSourceResponse, len(sources))
	for i, m := range sources {
		responses[i] = ToLeadSourceResponse(&m)
	}
	return responses
}
