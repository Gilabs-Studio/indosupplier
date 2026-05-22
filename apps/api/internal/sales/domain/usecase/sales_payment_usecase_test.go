package usecase

import (
	"context"
	"errors"
	"testing"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	financeDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/stretchr/testify/require"
)

type salesPaymentSettingsStub struct {
	codes map[string]string
}

func (s salesPaymentSettingsStub) GetCOACode(ctx context.Context, settingKey string) (string, error) {
	return "", errors.New("not used")
}

func (s salesPaymentSettingsStub) GetCOAByKey(ctx context.Context, key string) (string, error) {
	if code, ok := s.codes[key]; ok {
		return code, nil
	}
	return "", errors.New("mapping not found")
}

func (s salesPaymentSettingsStub) GetValue(ctx context.Context, settingKey string) (string, error) {
	return "", errors.New("not used")
}

func (s salesPaymentSettingsStub) GetAll(ctx context.Context) ([]financeModels.FinanceSetting, error) {
	return nil, errors.New("not used")
}

func (s salesPaymentSettingsStub) Upsert(ctx context.Context, key, value, description, category string) error {
	return errors.New("not used")
}

type salesPaymentCOAStub struct {
	byCode map[string]*financeDto.ChartOfAccountResponse
}

func (c salesPaymentCOAStub) Create(ctx context.Context, req *financeDto.CreateChartOfAccountRequest) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c salesPaymentCOAStub) Update(ctx context.Context, id string, req *financeDto.UpdateChartOfAccountRequest) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c salesPaymentCOAStub) Delete(ctx context.Context, id string) error {
	return errors.New("not used")
}

func (c salesPaymentCOAStub) GetByID(ctx context.Context, id string) (*financeDto.ChartOfAccountResponse, error) {
	return nil, errors.New("not used")
}

func (c salesPaymentCOAStub) List(ctx context.Context, req *financeDto.ListChartOfAccountsRequest) ([]financeDto.ChartOfAccountResponse, int64, error) {
	return nil, 0, errors.New("not used")
}

func (c salesPaymentCOAStub) Tree(ctx context.Context, onlyActive bool) ([]financeDto.ChartOfAccountTreeNode, error) {
	return nil, errors.New("not used")
}

func (c salesPaymentCOAStub) GetByCode(ctx context.Context, code string) (*financeDto.ChartOfAccountResponse, error) {
	if item, ok := c.byCode[code]; ok {
		return item, nil
	}
	return nil, errors.New("coa not found")
}

func (c salesPaymentCOAStub) RecalculateAllIsPostable(ctx context.Context) error {
	return errors.New("not used")
}

func TestSalesPaymentResolveTransactionCOAForJournal_CashUsesFinanceCashDefault(t *testing.T) {
	t.Parallel()

	uc := &salesPaymentUsecase{
		settingsUC: salesPaymentSettingsStub{codes: map[string]string{"finance.cash_default": "11101"}},
		coaUC: salesPaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11101": &financeDto.ChartOfAccountResponse{ID: "coa-cash", Code: "11101", Name: "Cash"},
		}},
	}

	coaID, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.SalesPayment{Method: models.SalesPaymentMethodCash}, nil)

	require.NoError(t, err)
	require.Equal(t, "coa-cash", coaID)
}

func TestSalesPaymentResolveTransactionCOAForJournal_BankUsesSelectedAccountCOA(t *testing.T) {
	t.Parallel()

	chartOfAccountID := "coa-bank"
	uc := &salesPaymentUsecase{
		settingsUC: salesPaymentSettingsStub{codes: map[string]string{"finance.bank_default": "11111"}},
		coaUC: salesPaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11111": &financeDto.ChartOfAccountResponse{ID: "coa-default-bank", Code: "11111", Name: "Bank"},
		}},
	}

	coaID, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.SalesPayment{Method: models.SalesPaymentMethodBank}, &coreModels.BankAccount{ChartOfAccountID: &chartOfAccountID})

	require.NoError(t, err)
	require.Equal(t, "coa-bank", coaID)
}

func TestSalesPaymentResolveTransactionCOAForJournal_BankRequiresLinkedAccountCOA(t *testing.T) {
	t.Parallel()

	uc := &salesPaymentUsecase{
		settingsUC: salesPaymentSettingsStub{codes: map[string]string{"finance.bank_default": "11111"}},
		coaUC: salesPaymentCOAStub{byCode: map[string]*financeDto.ChartOfAccountResponse{
			"11111": &financeDto.ChartOfAccountResponse{ID: "coa-default-bank", Code: "11111", Name: "Bank"},
		}},
	}

	_, err := uc.resolveTransactionCOAForJournal(context.Background(), &models.SalesPayment{Method: models.SalesPaymentMethodBank}, nil)

	require.Error(t, err)
}

func TestValidateInvoiceStatusForConfirm_AllowsApprovedAndRejectsWaiting(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateInvoiceStatusForConfirm(models.CustomerInvoiceStatusApproved))
	require.NoError(t, validateInvoiceStatusForConfirm(models.CustomerInvoiceStatusOverdue))
	require.Error(t, validateInvoiceStatusForConfirm(models.CustomerInvoiceStatusWaitingPayment))
	require.Error(t, validateInvoiceStatusForConfirm(models.CustomerInvoiceStatusWaitingApproval))
}
