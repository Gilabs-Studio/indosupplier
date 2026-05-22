package usecase

import (
	"sync"
	"testing"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"github.com/stretchr/testify/require"
	dbtest "github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

func TestPayment_ShouldApproveIdempotently_WhenConcurrentRequests(t *testing.T) {
	t.Parallel()

	db, cleanup := dbtest.OpenTestDB(t)
	defer cleanup()
	var err error

	err = db.AutoMigrate(
		&coreModels.AuditLog{},
		&coreModels.BankAccount{},
		&financeModels.ChartOfAccount{},
		&financeModels.JournalEntry{},
		&financeModels.JournalLine{},
		&financeModels.FinancialClosing{},
		&financeModels.Payment{},
		&financeModels.PaymentAllocation{},
		&financeModels.JournalReversal{},
	)
	require.NoError(t, err)

	coaCash := financeModels.ChartOfAccount{Code: "11100", Name: "Cash", Type: financeModels.AccountTypeAsset, IsActive: true}
	coaExpense := financeModels.ChartOfAccount{Code: "61000", Name: "Expense", Type: financeModels.AccountTypeExpense, IsActive: true}
	require.NoError(t, db.Create(&coaCash).Error)
	require.NoError(t, db.Create(&coaExpense).Error)

	bankAccount := coreModels.BankAccount{
		AccountHolder:    "BANK_01",
		Name:             "Main Bank",
		AccountNumber:    "123",
		ChartOfAccountID: &coaCash.ID,
		CreatedBy:        financeTestCompanyID,
		UpdatedBy:        financeTestCompanyID,
	}
	require.NoError(t, db.Create(&bankAccount).Error)

	coaRepo := repositories.NewChartOfAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)

	journalMapper := mapper.NewJournalEntryMapper(mapper.NewChartOfAccountMapper())
	paymentMapper := mapper.NewPaymentMapper(mapper.NewChartOfAccountMapper())

	auditService := audit.NewAuditService(db)
	journalUC := NewJournalEntryUsecase(db, coaRepo, journalRepo, journalMapper, auditService)
	paymentUC := NewPaymentUsecase(db, coaRepo, paymentRepo, journalUC, paymentMapper)

	ctx := financeTestContext()

	refType := reference.RefTypeSupplierInvoice
	refID := "00000000-0000-0000-0000-000000000001"

	req := &dto.CreatePaymentRequest{
		PaymentDate:   "2026-03-01",
		Description:   "Payment to Supplier",
		BankAccountID: bankAccount.ID,
		TotalAmount:   500,
		Allocations: []dto.PaymentAllocationRequest{
			{
				ChartOfAccountID: coaExpense.ID,
				Amount:           500,
				Memo:             "Expense",
				ReferenceType:    &refType,
				ReferenceID:      &refID,
			},
		},
	}

	created, err := paymentUC.Create(ctx, req)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	// Concurrent approve
	goroutines := 5
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, apErr := paymentUC.Approve(ctx, created.ID)
			mu.Lock()
			if apErr == nil {
				successCount++
			} else {
				errorCount++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// In SQLite concurrency, it might return db locked or simply pass gracefully due to idempotent check.
	// But mathematically, exactly one genuine insertion happens or identical idempotence returns success.
	// We want to ensure Journal is exactly 1.
	var journalCount int64
	db.Model(&financeModels.JournalEntry{}).Where("reference_id = ?", created.ID).Count(&journalCount)
	require.EqualValues(t, 1, journalCount, "Only ONE journal should be created regardless of concurrency")

	// Test Reverse
	reversed, err := paymentUC.Reverse(ctx, created.ID, "Mistake")
	require.NoError(t, err)
	require.Equal(t, financeModels.PaymentStatusReversed, reversed.Status)

	var journals []financeModels.JournalEntry
	db.Where("reference_id = ?", created.ID).Find(&journals)
	
	// Ensure posted journal is unchanged
	var posted *financeModels.JournalEntry
	for _, j := range journals {
		if j.Status == financeModels.JournalStatusPosted && j.ReferenceType != nil && *j.ReferenceType == reference.RefTypePayment {
			posted = &j
		}
	}
	require.NotNil(t, posted)

	var reversalMeta financeModels.JournalReversal
	err = db.Where("original_journal_entry_id = ?", posted.ID).First(&reversalMeta).Error
	require.NoError(t, err, "Reversal metadata must be inserted by ReverseWithReason")
}
