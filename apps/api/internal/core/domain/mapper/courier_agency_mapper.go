package mapper

import (
	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

// ToCourierAgencyResponse converts CourierAgency model to response DTO
func ToCourierAgencyResponse(m *models.CourierAgency) dto.CourierAgencyResponse {
	return dto.CourierAgencyResponse{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		Phone:       m.Phone,
		Address:     m.Address,
		TrackingURL: m.TrackingURL,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToCourierAgencyResponseList converts a slice of CourierAgency models to response DTOs
func ToCourierAgencyResponseList(models []models.CourierAgency) []dto.CourierAgencyResponse {
	responses := make([]dto.CourierAgencyResponse, len(models))
	for i, m := range models {
		responses[i] = ToCourierAgencyResponse(&m)
	}
	return responses
}
