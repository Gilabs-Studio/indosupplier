package usecase

import (
	"context"
	"strings"
	"testing"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	dbtest "github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockSettingsService struct {
	codes map[string]string
}

func (m *mockSettingsService) GetCOACode(ctx context.Context, key string) (string, error) {
	return m.codes[key], nil
}
func (m *mockSettingsService) GetCOAByKey(ctx context.Context, key string) (string, error) {
	return m.codes[key], nil
}
func (m *mockSettingsService) GetValue(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (m *mockSettingsService) GetAll(ctx context.Context) ([]models.FinanceSetting, error) {
	return nil, nil
}
func (m *mockSettingsService) Upsert(ctx context.Context, key, value, desc, cat string) error {
	return nil
}

func TestCashBankJournal_ControlAccountRestriction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil && strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
		t.Skip("sqlite integration test skipped because CGO is disabled")
	}
	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()

	err = db.AutoMigrate(
		&models.ChartOfAccount{},
		&models.CashBankJournal{},
		&models.CashBankJournalLine{},
		&coreModels.BankAccount{},
		&models.FinanceSetting{},
		&models.AccountingPeriod{},
		&models.FiscalYear{},
	)
	require.NoError(t, err)

	// 1. Setup COAs
	coaCash := models.ChartOfAccount{Code: "11100", Name: "Cash", Type: models.AccountTypeAsset, IsActive: true}
	coaAR := models.ChartOfAccount{Code: "11200", Name: "Accounts Receivable", Type: models.AccountTypeAsset, IsActive: true}
	require.NoError(t, db.Create(&coaCash).Error)
	require.NoError(t, db.Create(&coaAR).Error)

	// 2. Setup Bank Account
	bankAccount := coreModels.BankAccount{Name: "Main Bank", AccountNumber: "123", IsActive: true}
	bankAccount.CreatedBy = financeTestCompanyID
	bankAccount.UpdatedBy = financeTestCompanyID
	require.NoError(t, db.Create(&bankAccount).Error)

	// 3. Setup Mock Settings
	mockSettings := &mockSettingsService{
		codes: map[string]string{
			models.SettingCOASalesReceivable: "11200", // Restricted AR Code
		},
	}

	// 4. Initialize UC
	coaRepo := repositories.NewChartOfAccountRepository(db)
	repo := repositories.NewCashBankJournalRepository(db)
	coaMapper := mapper.NewChartOfAccountMapper()
	uc := &cashBankJournalUsecase{
		db:              db,
		coaRepo:         coaRepo,
		repo:            repo,
		settingsService: mockSettings,
		mapper:          mapper.NewCashBankJournalMapper(coaMapper),
	}

	ctx := financeTestContext()

	t.Run("Should fail when using restricted AR account", func(t *testing.T) {
		req := &dto.CreateCashBankJournalRequest{
			TransactionDate: "2026-03-01",
			Type:            models.CashBankTypeCashIn,
			BankAccountID:   bankAccount.ID,
			Lines: []dto.CashBankJournalLineRequest{
				{ChartOfAccountID: coaAR.ID, Amount: 1000, Memo: "Payment for Invoice"},
			},
		}

		_, err := uc.Create(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "restricted: trade control accounts")
	})

	t.Run("Should succeed when using normal account", func(t *testing.T) {
		req := &dto.CreateCashBankJournalRequest{
			TransactionDate: "2026-03-01",
			Type:            models.CashBankTypeCashIn,
			BankAccountID:   bankAccount.ID,
			Lines: []dto.CashBankJournalLineRequest{
				{ChartOfAccountID: coaCash.ID, Amount: 1000, Memo: "Cash Capital"},
			},
		}

		res, err := uc.Create(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}
