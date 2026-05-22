package usecase

import (
	"testing"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/stretchr/testify/require"
	dbtest "github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

func TestJournalEntry_ShouldRunValuationAndBeListedInValuationDomain(t *testing.T) {
	t.Parallel()

	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()
	var err error

	err = db.AutoMigrate(
		&coreModels.AuditLog{},
		&models.ChartOfAccount{},
		&models.JournalEntry{},
		&models.JournalLine{},
		&models.FinancialClosing{},
		&models.AccountingPeriod{},
		&models.FiscalYear{},
	)
	require.NoError(t, err)

	// Setup necessary data
	coa1 := models.ChartOfAccount{Code: "V001", Name: "Inventory Asset", Type: models.AccountTypeAsset, IsActive: true}
	coa2 := models.ChartOfAccount{Code: "V002", Name: "Valuation Expense", Type: models.AccountTypeExpense, IsActive: true}
	require.NoError(t, db.Create(&coa1).Error)
	require.NoError(t, db.Create(&coa2).Error)

	coaRepo := repositories.NewChartOfAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	journalMapper := mapper.NewJournalEntryMapper(mapper.NewChartOfAccountMapper())
	auditService := audit.NewAuditService(db)
	uc := NewJournalEntryUsecase(db, coaRepo, journalRepo, journalMapper, auditService)

	ctx := financeTestContext()

	// 1. Run Valuation (Skeleton)
	res, err := uc.RunValuation(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, models.JournalStatusPosted, res.Status)
	require.Equal(t, "INVENTORY_VALUATION", *res.ReferenceType)

	// 2. List with Valuation Domain Filter
	domain := "valuation"
	items, total, err := uc.List(ctx, &dto.ListJournalEntriesRequest{
		Domain:  &domain,
		Page:    1,
		PerPage: 10,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	require.Equal(t, "INVENTORY_VALUATION", *items[0].ReferenceType)

	// 3. Verify Trial Balance reflects the valuation entries
	trial, err := uc.TrialBalance(ctx, nil, nil)
	require.NoError(t, err)

	foundCoA1 := false
	for _, row := range trial.Rows {
		if row.Code == "V001" {
			foundCoA1 = true
			require.Equal(t, 100.0, row.DebitTotal)
		}
	}
	require.True(t, foundCoA1)
}
