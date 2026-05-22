package mapper

import (
	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

func ToCurrencyResponse(model *models.Currency) dto.CurrencyResponse {
	return dto.CurrencyResponse{
		ID:            model.ID,
		Code:          model.Code,
		Name:          model.Name,
		Symbol:        model.Symbol,
		DecimalPlaces: model.DecimalPlaces,
		IsActive:      model.IsActive,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

func ToCurrencyResponseList(models []models.Currency) []dto.CurrencyResponse {
	responses := make([]dto.CurrencyResponse, len(models))
	for i := range models {
		responses[i] = ToCurrencyResponse(&models[i])
	}
	return responses
}
