package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

type BankAccountMapper struct{}

func NewBankAccountMapper() *BankAccountMapper {
	return &BankAccountMapper{}
}

func (m *BankAccountMapper) ToResponse(model *models.BankAccount) *dto.BankAccountResponse {
	if model == nil {
		return nil
	}
	response := &dto.BankAccountResponse{
		ID:               model.ID,
		CompanyID:        model.CompanyID,
		Code:             model.Code,
		Name:             model.Name,
		AccountType:      model.AccountType,
		BankID:           model.BankID,
		AccountNumber:    model.AccountNumber,
		AccountHolder:    model.AccountHolder,
		CurrencyID:       model.CurrencyID,
		Currency:         model.Currency,
		ChartOfAccountID: model.ChartOfAccountID,
		VillageID:        model.VillageID,
		BankAddress:      model.BankAddress,
		BankPhone:        model.BankPhone,
		CountryCode:      model.CountryCode,
		BankBranchCode:   model.BankBranchCode,
		OpeningBalance:   model.OpeningBalance,
		CreatedBy:        model.CreatedBy,
		UpdatedBy:        model.UpdatedBy,
		IsActive:         model.IsActive,
		CreatedAt:        model.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        model.UpdatedAt.Format(time.RFC3339),
	}
	if model.CurrencyDetail != nil {
		response.CurrencyDetail = &dto.CurrencyResponse{
			ID:            model.CurrencyDetail.ID,
			Code:          model.CurrencyDetail.Code,
			Name:          model.CurrencyDetail.Name,
			Symbol:        model.CurrencyDetail.Symbol,
			DecimalPlaces: model.CurrencyDetail.DecimalPlaces,
			IsActive:      model.CurrencyDetail.IsActive,
			CreatedAt:     model.CurrencyDetail.CreatedAt,
			UpdatedAt:     model.CurrencyDetail.UpdatedAt,
		}
	}
	return response
}

func (m *BankAccountMapper) ToResponseList(items []models.BankAccount) []*dto.BankAccountResponse {
	out := make([]*dto.BankAccountResponse, 0, len(items))
	for i := range items {
		out = append(out, m.ToResponse(&items[i]))
	}
	return out
}
