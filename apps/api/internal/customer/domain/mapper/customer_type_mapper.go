package mapper

import (
	"github.com/gilabs/gims/api/internal/customer/data/models"
	"github.com/gilabs/gims/api/internal/customer/domain/dto"
)

// ToCustomerTypeResponse converts CustomerType model to response DTO
func ToCustomerTypeResponse(m *models.CustomerType) dto.CustomerTypeResponse {
	return dto.CustomerTypeResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToCustomerTypeResponseList converts a slice of CustomerType models to response DTOs
func ToCustomerTypeResponseList(models []models.CustomerType) []dto.CustomerTypeResponse {
	responses := make([]dto.CustomerTypeResponse, len(models))
	for i, m := range models {
		responses[i] = ToCustomerTypeResponse(&m)
	}
	return responses
}
