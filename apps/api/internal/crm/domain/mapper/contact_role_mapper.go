package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToContactRoleResponse converts a ContactRole model to a response DTO
func ToContactRoleResponse(m *models.ContactRole) dto.ContactRoleResponse {
	return dto.ContactRoleResponse{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: m.Description,
		BadgeColor:  m.BadgeColor,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToContactRoleResponseList converts a slice of ContactRole models to response DTOs
func ToContactRoleResponseList(roles []models.ContactRole) []dto.ContactRoleResponse {
	responses := make([]dto.ContactRoleResponse, len(roles))
	for i, m := range roles {
		responses[i] = ToContactRoleResponse(&m)
	}
	return responses
}
