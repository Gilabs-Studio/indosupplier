package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToLeadStatusResponse converts a LeadStatus model to a response DTO
func ToLeadStatusResponse(m *models.LeadStatus) dto.LeadStatusResponse {
	return dto.LeadStatusResponse{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		Score:       m.Score,
		Color:       m.Color,
		Order:       m.Order,
		IsActive:    m.IsActive,
		IsDefault:   m.IsDefault,
		IsConverted: m.IsConverted,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToLeadStatusResponseList converts a slice of LeadStatus models to response DTOs
func ToLeadStatusResponseList(statuses []models.LeadStatus) []dto.LeadStatusResponse {
	responses := make([]dto.LeadStatusResponse, len(statuses))
	for i, m := range statuses {
		responses[i] = ToLeadStatusResponse(&m)
	}
	return responses
}
