package usecase

// This file contains GetFormData implementations for all finance modules
// that need dropdown/form options on the frontend. Each module's GetFormData
// returns the specific data needed for its create/edit forms.

import (
	"context"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

// --- Shared helpers ---

func buildCOAOptions(coas []financeModels.ChartOfAccount) []dto.COAFormOption {
	opts := make([]dto.COAFormOption, 0, len(coas))
	for _, c := range coas {
		opts = append(opts, dto.COAFormOption{
			ID:   c.ID,
			Code: c.Code,
			Name: c.Name,
			Type: string(c.Type),
		})
	}
	return opts
}

func buildBankAccountOptions(banks []coreModels.BankAccount) []dto.BankAccountFormOption {
	opts := make([]dto.BankAccountFormOption, 0, len(banks))
	for _, b := range banks {
		opts = append(opts, dto.BankAccountFormOption{
			ID:            b.ID,
			AccountName:   b.Name,
			AccountNumber: b.AccountNumber,
			BankName:      b.AccountHolder,
			Currency:      b.Currency,
			COAId:         b.ChartOfAccountID,
		})
	}
	return opts
}

// --- Payment ---

func (uc *paymentUsecase) GetFormData(ctx context.Context) (*dto.PaymentFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	var banks []coreModels.BankAccount
	if err := uc.db.WithContext(ctx).Where("is_active = ?", true).Order("name").Find(&banks).Error; err != nil {
		return nil, err
	}

	return &dto.PaymentFormDataResponse{
		ChartOfAccounts: buildCOAOptions(coas),
		BankAccounts:    buildBankAccountOptions(banks),
	}, nil
}

// --- CashBankJournal ---

func (uc *cashBankJournalUsecase) GetFormData(ctx context.Context) (*dto.CashBankFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	var banks []coreModels.BankAccount
	if err := uc.db.WithContext(ctx).Where("is_active = ?", true).Order("name").Find(&banks).Error; err != nil {
		return nil, err
	}

	types := []dto.EnumOption{
		{Value: string(financeModels.CashBankTypeCashIn), Label: "Cash In"},
		{Value: string(financeModels.CashBankTypeCashOut), Label: "Cash Out"},
		{Value: string(financeModels.CashBankTypeTransfer), Label: "Transfer"},
	}

	return &dto.CashBankFormDataResponse{
		ChartOfAccounts: buildCOAOptions(coas),
		BankAccounts:    buildBankAccountOptions(banks),
		Types:           types,
	}, nil
}

// --- Budget ---

func (uc *budgetUsecase) GetFormData(ctx context.Context) (*dto.BudgetFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	return &dto.BudgetFormDataResponse{
		ChartOfAccounts: buildCOAOptions(coas),
	}, nil
}

// --- NonTradePayable ---

func (uc *nonTradePayableUsecase) GetFormData(ctx context.Context) (*dto.NonTradePayableFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	return &dto.NonTradePayableFormDataResponse{
		ChartOfAccounts: buildCOAOptions(coas),
	}, nil
}

// --- Asset ---

func (uc *assetUsecase) GetFormData(ctx context.Context) (*dto.AssetFormDataResponse, error) {
	cats, _, err := uc.catRepo.List(ctx, repositories.AssetCategoryListParams{Limit: 100})
	if err != nil {
		return nil, err
	}

	locs, _, err := uc.locRepo.List(ctx, repositories.AssetLocationListParams{Limit: 100})
	if err != nil {
		return nil, err
	}

	catOpts := make([]dto.AssetCategoryFormOption, 0, len(cats))
	for _, c := range cats {
		catOpts = append(catOpts, dto.AssetCategoryFormOption{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	locOpts := make([]dto.AssetLocationFormOption, 0, len(locs))
	for _, l := range locs {
		locOpts = append(locOpts, dto.AssetLocationFormOption{
			ID:   l.ID,
			Name: l.Name,
		})
	}

	return &dto.AssetFormDataResponse{
		Categories: catOpts,
		Locations:  locOpts,
	}, nil
}

// --- AssetCategory ---

func (uc *assetCategoryUsecase) GetFormData(ctx context.Context) (*dto.AssetCategoryFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	methods := []dto.EnumOption{
		{Value: string(financeModels.DepreciationMethodStraightLine), Label: "Straight Line (SL)"},
		{Value: string(financeModels.DepreciationMethodDecliningBalance), Label: "Declining Balance (DB)"},
	}

	return &dto.AssetCategoryFormDataResponse{
		ChartOfAccounts:     buildCOAOptions(coas),
		DepreciationMethods: methods,
	}, nil
}

// --- UpCountryCost ---

func (uc *upCountryCostUsecase) GetFormData(ctx context.Context) (*dto.UpCountryCostFormDataResponse, error) {
	costTypes := []dto.EnumOption{
		{Value: string(financeModels.CostTypeTransport), Label: "Transport"},
		{Value: string(financeModels.CostTypeAccommodation), Label: "Accommodation"},
		{Value: string(financeModels.CostTypeMeal), Label: "Meal"},
		{Value: string(financeModels.CostTypeFuel), Label: "Fuel"},
		{Value: string(financeModels.CostTypeOther), Label: "Other"},
	}

	return &dto.UpCountryCostFormDataResponse{
		CostTypes: costTypes,
	}, nil
}
