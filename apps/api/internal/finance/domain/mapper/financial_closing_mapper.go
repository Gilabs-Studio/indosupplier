package mapper

import (
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type FinancialClosingMapper struct{}

func NewFinancialClosingMapper() *FinancialClosingMapper {
	return &FinancialClosingMapper{}
}

func (m *FinancialClosingMapper) ToResponse(item *financeModels.FinancialClosing) dto.FinancialClosingResponse {
	if item == nil {
		return dto.FinancialClosingResponse{}
	}
	return dto.FinancialClosingResponse{
		ID: item.ID,
		PeriodEndDate: item.PeriodEndDate,
		Status: item.Status,
		Notes: item.Notes,
		ApprovedAt: item.ApprovedAt,
		ApprovedBy: item.ApprovedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
