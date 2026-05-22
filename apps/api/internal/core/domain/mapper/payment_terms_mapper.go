package mapper

import (
	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

// ToPaymentTermsResponse converts PaymentTerms model to response DTO
func ToPaymentTermsResponse(m *models.PaymentTerms) dto.PaymentTermsResponse {
	return dto.PaymentTermsResponse{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		Days:        m.Days,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToPaymentTermsResponseList converts a slice of PaymentTerms models to response DTOs
func ToPaymentTermsResponseList(models []models.PaymentTerms) []dto.PaymentTermsResponse {
	responses := make([]dto.PaymentTermsResponse, len(models))
	for i, m := range models {
		responses[i] = ToPaymentTermsResponse(&m)
	}
	return responses
}
