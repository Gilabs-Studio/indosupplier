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

func TestAdjustmentJournal_ShouldCreatePostAndReverse(t *testing.T) {
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
		&models.JournalReversal{},
	)
	require.NoError(t, err)

	coaCash := models.ChartOfAccount{Code: "11100", Name: "Cash", Type: models.AccountTypeAsset, IsActive: true}
	coaSales := models.ChartOfAccount{Code: "41000", Name: "Sales", Type: models.AccountTypeRevenue, IsActive: true}
	require.NoError(t, db.Create(&coaCash).Error)
	require.NoError(t, db.Create(&coaSales).Error)

	coaRepo := repositories.NewChartOfAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	journalMapper := mapper.NewJournalEntryMapper(mapper.NewChartOfAccountMapper())
	auditService := audit.NewAuditService(db)
	uc := NewJournalEntryUsecase(db, coaRepo, journalRepo, journalMapper, auditService)

	ctx := financeTestContext()

	createReq := &dto.CreateAdjustmentJournalRequest{
		EntryDate:   "2026-03-01",
		Description: "Adjustment entry",
		Lines: []dto.JournalLineRequest{
			{ChartOfAccountID: coaCash.ID, Debit: 500, Credit: 0, Memo: "adj debit cash"},
			{ChartOfAccountID: coaSales.ID, Debit: 0, Credit: 500, Memo: "adj credit sales"},
		},
	}

	created, err := uc.CreateAdjustmentJournal(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "MANUAL_ADJUSTMENT", *created.ReferenceType)

	updateReq := &dto.UpdateJournalEntryRequest{
		EntryDate:   "2026-03-02",
		Description: "Updated adjustment entry",
		Lines: []dto.JournalLineRequest{
			{ChartOfAccountID: coaCash.ID, Debit: 800, Credit: 0, Memo: "upd adj debit cash"},
			{ChartOfAccountID: coaSales.ID, Debit: 0, Credit: 800, Memo: "upd adj credit sales"},
		},
	}

	updated, err := uc.UpdateAdjustmentJournal(ctx, created.ID, updateReq)
	require.NoError(t, err)
	require.Equal(t, "Updated adjustment entry", updated.Description)
	require.Equal(t, "MANUAL_ADJUSTMENT", *updated.ReferenceType)
	require.Equal(t, float64(800), updated.Lines[0].Debit)

	posted, err := uc.PostAdjustmentJournal(ctx, updated.ID)
	require.NoError(t, err)
	require.Equal(t, models.JournalStatusPosted, posted.Status)

	reversal, err := uc.ReverseAdjustmentJournal(ctx, updated.ID)
	require.NoError(t, err)
	require.Equal(t, models.JournalStatusPosted, reversal.Status)

	var reversalMeta models.JournalReversal
	err = db.Where("original_journal_entry_id = ?", updated.ID).First(&reversalMeta).Error
	require.NoError(t, err)
	require.Equal(t, reversal.ID, reversalMeta.ReversalJournalEntryID)
}
