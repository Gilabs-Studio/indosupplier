package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToContactResponse converts a Contact model to a response DTO
func ToContactResponse(m *models.Contact) dto.ContactResponse {
	resp := dto.ContactResponse{
		ID:            m.ID,
		CustomerID:    m.CustomerID,
		ContactRoleID: m.ContactRoleID,
		Name:          m.Name,
		Phone:         m.Phone,
		Email:         m.Email,
		Notes:         m.Notes,
		IsActive:      m.IsActive,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	if m.Customer != nil {
		resp.Customer = &dto.ContactCustomerInfo{
			ID:   m.Customer.ID,
			Code: m.Customer.Code,
			Name: m.Customer.Name,
		}
	}

	if m.ContactRole != nil {
		resp.ContactRole = &dto.ContactRoleInfo{
			ID:         m.ContactRole.ID,
			Name:       m.ContactRole.Name,
			Code:       m.ContactRole.Code,
			BadgeColor: m.ContactRole.BadgeColor,
		}
	}

	return resp
}

// ToContactResponseList converts a slice of Contact models to response DTOs
func ToContactResponseList(contacts []models.Contact) []dto.ContactResponse {
	responses := make([]dto.ContactResponse, len(contacts))
	for i, m := range contacts {
		responses[i] = ToContactResponse(&m)
	}
	return responses
}
