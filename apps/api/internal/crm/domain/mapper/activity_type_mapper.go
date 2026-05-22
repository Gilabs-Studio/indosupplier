package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToActivityTypeResponse converts an ActivityType model to a response DTO
func ToActivityTypeResponse(m *models.ActivityType) dto.ActivityTypeResponse {
	return dto.ActivityTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		Icon:        m.Icon,
		BadgeColor:  m.BadgeColor,
		Order:       m.Order,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToActivityTypeResponseList converts a slice of ActivityType models to response DTOs
func ToActivityTypeResponseList(actTypes []models.ActivityType) []dto.ActivityTypeResponse {
	responses := make([]dto.ActivityTypeResponse, len(actTypes))
	for i, m := range actTypes {
		responses[i] = ToActivityTypeResponse(&m)
	}
	return responses
}
