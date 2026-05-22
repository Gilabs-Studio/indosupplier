package usecase

import (
	"context"
	"testing"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestValidateLines_ShouldRejectUnbalancedEntries(t *testing.T) {
	t.Parallel()

	_, _, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 100, Credit: 0},
		{ChartOfAccountID: "coa-2", Debit: 0, Credit: 90},
	})

	require.ErrorIs(t, err, ErrJournalUnbalanced)
}

func TestJournalReferenceTypesForDomain_ShouldMapKnownDomain(t *testing.T) {
	t.Parallel()

	domain := "purchase"
	types := journalReferenceTypesForDomain(&domain)

	require.Contains(t, types, "GOODS_RECEIPT")
	require.Contains(t, types, "SUPPLIER_INVOICE")
	require.Contains(t, types, "PURCHASE_PAYMENT")
}

func TestValidateLines_ShouldRejectSingleLine(t *testing.T) {
	t.Parallel()

	_, _, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 100, Credit: 0},
	})

	require.ErrorIs(t, err, ErrJournalInvalidLines)
}

func TestValidateLines_ShouldRejectLineWithBothDebitAndCredit(t *testing.T) {
	t.Parallel()

	_, _, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 100, Credit: 50},
		{ChartOfAccountID: "coa-2", Debit: 0, Credit: 50},
	})

	require.ErrorIs(t, err, ErrJournalInvalidLines)
}

func TestValidateLines_ShouldRejectLineWithZeroDebitAndCredit(t *testing.T) {
	t.Parallel()

	_, _, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 0, Credit: 0},
		{ChartOfAccountID: "coa-2", Debit: 100, Credit: 0},
	})

	require.ErrorIs(t, err, ErrJournalInvalidLines)
}

func TestValidateLines_ShouldPassBalancedEntry(t *testing.T) {
	t.Parallel()

	debit, credit, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 500, Credit: 0},
		{ChartOfAccountID: "coa-2", Debit: 0, Credit: 500},
	})

	require.NoError(t, err)
	require.InDelta(t, 500.0, debit, 0.001)
	require.InDelta(t, 500.0, credit, 0.001)
}

func TestValidateLines_ShouldPassMultiLineBalancedEntry(t *testing.T) {
	t.Parallel()

	debit, credit, err := validateLines([]dto.JournalLineRequest{
		{ChartOfAccountID: "coa-1", Debit: 300, Credit: 0},
		{ChartOfAccountID: "coa-2", Debit: 200, Credit: 0},
		{ChartOfAccountID: "coa-3", Debit: 0, Credit: 500},
	})

	require.NoError(t, err)
	require.InDelta(t, 500.0, debit, 0.001)
	require.InDelta(t, 500.0, credit, 0.001)
}

func TestJournalReferenceTypesForDomain_ShouldMapAdjustmentDomain(t *testing.T) {
	t.Parallel()

	domain := "adjustment"
	types := journalReferenceTypesForDomain(&domain)

	require.Contains(t, types, "MANUAL_ADJUSTMENT")
	require.Contains(t, types, "ADJUSTMENT")
	require.Contains(t, types, "CORRECTION")
}

func TestJournalReferenceTypesForDomain_ShouldMapValuationDomain(t *testing.T) {
	t.Parallel()

	domain := "valuation"
	types := journalReferenceTypesForDomain(&domain)

	require.Contains(t, types, "INVENTORY_VALUATION")
	require.Contains(t, types, "CURRENCY_REVALUATION")
	require.Contains(t, types, "COST_ADJUSTMENT")
}

func TestJournalReferenceTypesForDomain_ShouldReturnNilForNilDomain(t *testing.T) {
	t.Parallel()

	types := journalReferenceTypesForDomain(nil)

	require.Nil(t, types)
}

func TestJournalReferenceTypesForDomain_ShouldReturnNilForUnknownDomain(t *testing.T) {
	t.Parallel()

	domain := "unknown_xyz"
	types := journalReferenceTypesForDomain(&domain)

	require.Nil(t, types)
}

func TestBuildOpeningBalanceLines_ShouldDebitAccountCreditEquity_ForAsset(t *testing.T) {
	t.Parallel()

	lines := buildOpeningBalanceLines("coa-asset", financeModels.AccountTypeAsset, 1000, "coa-equity")

	require.Len(t, lines, 2)
	require.Equal(t, "coa-asset", lines[0].ChartOfAccountID)
	require.Equal(t, 1000.0, lines[0].Debit)
	require.Equal(t, 0.0, lines[0].Credit)
	require.Equal(t, "coa-equity", lines[1].ChartOfAccountID)
	require.Equal(t, 0.0, lines[1].Debit)
	require.Equal(t, 1000.0, lines[1].Credit)
}

func TestBuildOpeningBalanceLines_ShouldDebitEquityCreditAccount_ForLiability(t *testing.T) {
	t.Parallel()

	lines := buildOpeningBalanceLines("coa-liability", financeModels.AccountTypeLiability, 2500, "coa-equity")

	require.Len(t, lines, 2)
	require.Equal(t, "coa-equity", lines[0].ChartOfAccountID)
	require.Equal(t, 2500.0, lines[0].Debit)
	require.Equal(t, 0.0, lines[0].Credit)
	require.Equal(t, "coa-liability", lines[1].ChartOfAccountID)
	require.Equal(t, 0.0, lines[1].Debit)
	require.Equal(t, 2500.0, lines[1].Credit)
}

type stubChartOfAccountRepository struct {
	items []financeModels.ChartOfAccount
}

func (s stubChartOfAccountRepository) Create(ctx context.Context, item *financeModels.ChartOfAccount) error {
	return nil
}

func (s stubChartOfAccountRepository) FindByID(ctx context.Context, id string) (*financeModels.ChartOfAccount, error) {
	return nil, gorm.ErrRecordNotFound
}

func (s stubChartOfAccountRepository) GetDB(ctx context.Context) *gorm.DB {
	return nil
}

func (s stubChartOfAccountRepository) FindAll(ctx context.Context, activeOnly bool) ([]financeModels.ChartOfAccount, error) {
	return s.items, nil
}

func (s stubChartOfAccountRepository) List(ctx context.Context, params repositories.ChartOfAccountListParams) ([]financeModels.ChartOfAccount, int64, error) {
	return s.items, int64(len(s.items)), nil
}

func (s stubChartOfAccountRepository) Update(ctx context.Context, item *financeModels.ChartOfAccount) error {
	return nil
}

func (s stubChartOfAccountRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (s stubChartOfAccountRepository) ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error) {
	return false, nil
}

func (s stubChartOfAccountRepository) FindByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	for _, item := range s.items {
		if item.Code == code {
			return &item, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s stubChartOfAccountRepository) GetByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	return s.FindByCode(ctx, code)
}

func (s stubChartOfAccountRepository) FindOpeningBalanceEquity(ctx context.Context) (*financeModels.ChartOfAccount, error) {
	for _, item := range s.items {
		if item.Code == "39999" {
			return &item, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s stubChartOfAccountRepository) HasChildren(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (s stubChartOfAccountRepository) HasJournalLines(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (s stubChartOfAccountRepository) IsUsedInJournal(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (s stubChartOfAccountRepository) UpdateIsPostable(ctx context.Context, id string, isPostable bool) error {
	return nil
}

func (s stubChartOfAccountRepository) RecalculateAllIsPostable(ctx context.Context) error {
	return nil
}

type stubFinanceReportRepository struct {
	balances       []repositories.GLAccountBalance
	lines          []financeModels.JournalLine
	netProfit      float64
	getNetProfitFn func(startDate, endDate time.Time) float64
}

func (s stubFinanceReportRepository) GetAccountBalances(context.Context, time.Time, time.Time, *string) ([]repositories.GLAccountBalance, error) {
	return s.balances, nil
}

func (s stubFinanceReportRepository) GetAccountBalancesByAccounts(_ context.Context, accountIDs []string, _ time.Time, _ time.Time, _ *string, _ *string) (map[string]repositories.GLAccountBalance, error) {
	result := make(map[string]repositories.GLAccountBalance, len(accountIDs))
	byID := make(map[string]repositories.GLAccountBalance, len(s.balances))
	for _, b := range s.balances {
		byID[b.ChartOfAccountID] = b
	}
	for _, id := range accountIDs {
		if item, ok := byID[id]; ok {
			result[id] = item
			continue
		}
		result[id] = repositories.GLAccountBalance{ChartOfAccountID: id}
	}
	return result, nil
}

func (s stubFinanceReportRepository) GetGLAccountTransactions(context.Context, string, time.Time, time.Time, *string, *string) ([]financeModels.JournalLine, error) {
	return s.lines, nil
}

func (s stubFinanceReportRepository) GetGLAccountTransactionsByAccounts(context.Context, []string, time.Time, time.Time, *string, *string) ([]financeModels.JournalLine, error) {
	return s.lines, nil
}

func (s stubFinanceReportRepository) GetNetProfit(_ context.Context, startDate, endDate time.Time, _ *string, _ *string) (float64, error) {
	if s.getNetProfitFn != nil {
		return s.getNetProfitFn(startDate, endDate), nil
	}
	return s.netProfit, nil
}

func TestFinanceReportUsecase_ShouldAggregateBalanceSheet_FromClosingBalances(t *testing.T) {
	t.Parallel()

	coaRepo := stubChartOfAccountRepository{items: []financeModels.ChartOfAccount{
		{ID: "asset-1", Code: "11100", Name: "Cash", Type: financeModels.AccountTypeAsset},
		{ID: "liab-1", Code: "21100", Name: "Accounts Payable", Type: financeModels.AccountTypeLiability},
		{ID: "equity-1", Code: "31100", Name: "Retained Earnings", Type: financeModels.AccountTypeEquity},
		{ID: "zero-1", Code: "11200", Name: "Unused Asset", Type: financeModels.AccountTypeAsset},
	}}
	reportRepo := stubFinanceReportRepository{balances: []repositories.GLAccountBalance{
		{ChartOfAccountID: "asset-1", ClosingBalance: 1500},
		{ChartOfAccountID: "liab-1", ClosingBalance: 500},
		{ChartOfAccountID: "equity-1", ClosingBalance: 1000},
		{ChartOfAccountID: "zero-1", ClosingBalance: 0},
	}}

	uc := NewFinanceReportUsecase(nil, coaRepo, reportRepo)
	res, err := uc.GetBalanceSheet(context.Background(), time.Now(), time.Now(), nil, nil, false)

	require.NoError(t, err)
	require.Len(t, res.Assets, 1)
	require.Len(t, res.Liabilities, 1)
	require.Len(t, res.Equities, 1)
	require.Equal(t, 1500.0, res.AssetTotal)
	require.Equal(t, 500.0, res.LiabilityTotal)
	require.Equal(t, 1000.0, res.EquityTotal)
	require.Equal(t, 0.0, res.CurrentYearProfit)
	require.Equal(t, 1000.0, res.EquityTotalFinal)
	require.Equal(t, 1500.0, res.LiabilityEquity)
}

func TestFinanceReportUsecase_ShouldCalculateProfitAndLoss_FromMovements(t *testing.T) {
	t.Parallel()

	coaRepo := stubChartOfAccountRepository{items: []financeModels.ChartOfAccount{
		{ID: "rev-1", Code: "4100", Name: "Sales Revenue", Type: financeModels.AccountTypeRevenue},
		{ID: "exp-1", Code: "6100", Name: "Operating Expense", Type: financeModels.AccountTypeExpense},
		{ID: "asset-1", Code: "11100", Name: "Cash", Type: financeModels.AccountTypeAsset},
	}}
	reportRepo := stubFinanceReportRepository{balances: []repositories.GLAccountBalance{
		{ChartOfAccountID: "rev-1", DebitTotal: 0, CreditTotal: 1500},
		{ChartOfAccountID: "exp-1", DebitTotal: 300, CreditTotal: 0},
		{ChartOfAccountID: "asset-1", DebitTotal: 1500, CreditTotal: 300},
	}}

	uc := NewFinanceReportUsecase(nil, coaRepo, reportRepo)
	res, err := uc.GetProfitAndLoss(context.Background(), time.Now(), time.Now(), nil, nil)

	require.NoError(t, err)
	require.Len(t, res.Revenues, 1)
	require.Len(t, res.Expenses, 1)
	require.Equal(t, 1500.0, res.RevenueTotal)
	require.Equal(t, 0.0, res.COGSTotal)
	require.Equal(t, 300.0, res.ExpenseTotal)
	require.Equal(t, 1500.0, res.GrossProfit)
	require.Equal(t, 1200.0, res.NetProfit)
	require.Equal(t, 100.0, res.GrossMargin)
	require.Equal(t, 80.0, res.NetMargin)
	require.Equal(t, 20.0, res.ExpenseRatio)
}

func TestFinanceReportUsecase_ShouldRespectIncludeZeroFlag_InBalanceSheet(t *testing.T) {
	t.Parallel()

	coaRepo := stubChartOfAccountRepository{items: []financeModels.ChartOfAccount{
		{ID: "asset-1", Code: "11100", Name: "Cash", Type: financeModels.AccountTypeAsset},
		{ID: "asset-2", Code: "11200", Name: "Inventory", Type: financeModels.AccountTypeAsset},
	}}
	reportRepo := stubFinanceReportRepository{balances: []repositories.GLAccountBalance{
		{ChartOfAccountID: "asset-1", ClosingBalance: 100},
		{ChartOfAccountID: "asset-2", ClosingBalance: 0},
	}}

	uc := NewFinanceReportUsecase(nil, coaRepo, reportRepo)
	hideZero, err := uc.GetBalanceSheet(context.Background(), time.Now(), time.Now(), nil, nil, false)
	require.NoError(t, err)
	require.Len(t, hideZero.Assets, 1)

	showZero, err := uc.GetBalanceSheet(context.Background(), time.Now(), time.Now(), nil, nil, true)
	require.NoError(t, err)
	require.Len(t, showZero.Assets, 2)
}

func TestFinanceReportUsecase_ShouldComposeFinalEquity_FromRetainedAndCurrentProfit(t *testing.T) {
	t.Parallel()

	endDate := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	coaRepo := stubChartOfAccountRepository{items: []financeModels.ChartOfAccount{
		{ID: "asset-1", Code: "11100", Name: "Cash", Type: financeModels.AccountTypeAsset},
		{ID: "liab-1", Code: "21100", Name: "Accounts Payable", Type: financeModels.AccountTypeLiability},
		{ID: "equity-1", Code: "31100", Name: "Capital", Type: financeModels.AccountTypeEquity},
	}}
	reportRepo := stubFinanceReportRepository{
		balances: []repositories.GLAccountBalance{
			{ChartOfAccountID: "asset-1", ClosingBalance: 1800},
			{ChartOfAccountID: "liab-1", ClosingBalance: 500},
			{ChartOfAccountID: "equity-1", ClosingBalance: 700},
		},
		getNetProfitFn: func(startDate, _ time.Time) float64 {
			if startDate.Equal(yearStart) {
				return 400 // current-year profit
			}
			return 200 // retained earnings
		},
	}

	uc := NewFinanceReportUsecase(nil, coaRepo, reportRepo)
	res, err := uc.GetBalanceSheet(context.Background(), yearStart, endDate, nil, nil, false)
	require.NoError(t, err)
	require.Equal(t, 700.0, res.EquityTotal)
	require.Equal(t, 200.0, res.RetainedEarnings)
	require.Equal(t, 400.0, res.CurrentYearProfit)
	require.Equal(t, 1300.0, res.EquityTotalFinal)
	require.Equal(t, 1800.0, res.AssetTotal)
	require.Equal(t, 1800.0, res.LiabilityEquity)
	require.True(t, res.IsBalanced)
}
