package mapper

import (
	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
)

// ToBankResponse converts Bank model to response DTO
func ToBankResponse(m *models.Bank) dto.BankResponse {
	return dto.BankResponse{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		SwiftCode: m.SwiftCode,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ToBankResponseList converts a slice of Bank models to response DTOs
func ToBankResponseList(models []models.Bank) []dto.BankResponse {
	responses := make([]dto.BankResponse, len(models))
	for i, m := range models {
		responses[i] = ToBankResponse(&m)
	}
	return responses
}
