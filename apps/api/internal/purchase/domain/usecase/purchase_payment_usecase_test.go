package usecase

import (
	"context"
	"errors"
	"testing"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	financeDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/stretchr/testify/require"
)

type purchasePaymentSettingsStub struct {
	codes map[string]string
}

func (s purchasePaymentSettingsStub) GetCOACode(ctx context.Context, settingKey string) (string, error) {
	return "", errors.New("not used")
}

func (s purchasePaymentSettingsStub) GetCOAByKey(ctx context.Context, key string) (string, error) {
	if code, ok := s.codes[key]; ok {
		return code, nil
	}
	return "", errors.New("mapping not found")
}

func (s purchasePaymentSettingsStub) GetValue(ctx context.Context, settingKey string) (string, error) {
	return "", errors.New("not used")
}

func (s purchasePaymentSettingsStub) GetAll(ctx context.Context) ([]financeModels.FinanceSetting, error) {
	return nil, errors.New("not used")
}

func (s purchasePaymentSettingsStub) Upsert(ctx context.Context, key, value, description, category string) error {
	return errors.New("not used")
}

type purchasePaymentCOAStub struct {
	byCode map[string]*financeDto.ChartOfAccountResponse
}

func (c purchasePaymentCOAStub) Create(ctx context.Context, req *financeDto.CreateChartOfAccountRequest) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c purchasePaymentCOAStub) Update(ctx context.Context, id string, req *financeDto.UpdateChartOfAccountRequest) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c purchasePaymentCOAStub) Delete(ctx context.Context, id string) error {
	return errors.New("not used")
}

func (c purchasePaymentCOAStub) GetByID(ctx context.Context, id string) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c purchasePaymentCOAStub) List(ctx context.Context, req *financeDto.ListChartOfAccountsRequest) ([]financeDto.ChartOfAccountResponse, int64, error) {
	return nil, 0, errors.New("not used")
}

func (c purchasePaymentCOAStub) Tree(ctx context.Context, onlyActive bool) ([]financeDto.ChartOfAccountTreeNode, error) {
	return nil, errors.New("not used")
}

func (c purchasePaymentCOAStub) GetByCode(ctx context.Context, code string) (*financeDto.ChartOfAccountResponse, error) {
	if item, ok := c.byCode[code]; ok {
		return item, nil
	}
	return nil, errors.New("coa not found")
}

func (c purchasePaymentCOAStub) RecalculateAllIsPostable(ctx context.Context) error {
	return errors.New("not used")
}

func TestPurchasePaymentResolveTransactionCOAForJournal_CashUsesFinanceCashDefault(t *testing.T) {
	t.Parallel()

	uc := &purchasePaymentUsecase{
		settingsUC: purchasePaymentSettingsStub{codes: map[string]string{"finance.cash_default": "11101"}},
		coaUC: purchasePaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11101": &financeDto.ChartOfAccountResponse{ID: "coa-cash", Code: "11101", Name: "Cash"},
		}},
	}

	coaID, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.PurchasePayment{Method: models.PurchasePaymentMethodCash}, nil)

	require.NoError(t, err)
	require.Equal(t, "coa-cash", coaID)
}

func TestPurchasePaymentResolveTransactionCOAForJournal_BankUsesSelectedAccountCOA(t *testing.T) {
	t.Parallel()

	chartOfAccountID := "coa-bank"
	uc := &purchasePaymentUsecase{
		settingsUC: purchasePaymentSettingsStub{codes: map[string]string{"finance.bank_default": "11111"}},
		coaUC: purchasePaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11111": &financeDto.ChartOfAccountResponse{ID: "coa-default-bank", Code: "11111", Name: "Bank"},
		}},
	}

	coaID, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.PurchasePayment{Method: models.PurchasePaymentMethodBank}, &coreModels.BankAccount{ChartOfAccountID: &chartOfAccountID})

	require.NoError(t, err)
	require.Equal(t, "coa-bank", coaID)
}

func TestPurchasePaymentResolveTransactionCOAForJournal_BankRequiresLinkedAccountCOA(t *testing.T) {
	t.Parallel()

	uc := &purchasePaymentUsecase{
		settingsUC: purchasePaymentSettingsStub{codes: map[string]string{"finance.bank_default": "11111"}},
		coaUC: purchasePaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11111": &financeDto.ChartOfAccountResponse{ID: "coa-default-bank", Code: "11111", Name: "Bank"},
		}},
	}

	_, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.PurchasePayment{Method: models.PurchasePaymentMethodBank}, nil)

	require.Error(t, err)
}
