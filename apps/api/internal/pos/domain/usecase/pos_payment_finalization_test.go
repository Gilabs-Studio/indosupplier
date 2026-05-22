package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"gorm.io/gorm"
)

// TestPOSFinalizationAtomicRollbackOnInvoiceFailure verifies that if invoice creation fails,
// the entire finalization transaction (order status, SO, Invoice, Payment) is rolled back.
//
// Scenario:
// 1. Order in UNPAID status
// 2. POSPayment record created successfully
// 3. finalizeOrder() called, which starts atomic transaction
// 4. Order status → PAID (succeeds)
// 5. SalesOrder created (succeeds)
// 6. CustomerInvoice creation forced to fail (mocked repo)
// 7. Transaction rolls back → order status reverts to UNPAID, SO deleted, Invoice not created
//
// This test ensures the atomicity guarantee: if any ERP doc creation fails, the entire
// set of changes (order status + SO + Invoice + Payment) is rolled back.
func TestPOSFinalizationAtomicRollbackOnInvoiceFailure(t *testing.T) {
	// Note: This is a high-level test template. Actual implementation requires:
	// - A test database with POS + Sales schemas
	// - Mocked repository that fails on invoice creation
	// - Verification of transaction rollback via direct DB query
	//
	// Placeholder structure for future integration test suite:

	t.Run("order_status_reverts_when_invoice_creation_fails", func(t *testing.T) {
		// Given: POS order in UNPAID status
		// When: finalizeOrder() is called and invoice creation fails
		// Then: order status should remain UNPAID (not PAID)
		//       SalesOrder should not exist in DB
		//       Transaction should be rolled back atomically

		// Setup would require:
		// db := setupTestDB()
		// orderRepo := repositories.NewPosOrderRepository(db)
		// soRepo := salesRepos.NewSalesOrderRepository(db)
		// invoiceRepo := salesRepos.NewCustomerInvoiceRepository(db)
		// paymentRepo := repositories.NewPOSPaymentRepository(db)
		//
		// order := createTestPOSOrder()
		// payment := createTestPOSPayment(order.ID)
		//
		// // Mock invoice creation to fail
		// invoiceRepoMock := &failingInvoiceRepository{underlying: invoiceRepo}
		//
		// usecase := NewPOSPaymentUsecase(db, paymentRepo, orderRepo, ..., invoiceRepoMock, ...)
		// err := usecase.finalizeOrder(ctx, order, userID, payment)
		//
		// // Assert: error returned
		// require.Error(t, err)
		//
		// // Assert: order status still UNPAID (not PAID)
		// updatedOrder, _ := orderRepo.GetByID(ctx, order.ID)
		// require.Equal(t, models.PosOrderStatusUnpaid, updatedOrder.Status)
		//
		// // Assert: no SalesOrder created
		// salesOrders, _ := soRepo.List(ctx, &dto.ListSalesOrdersRequest{})
		// require.Empty(t, salesOrders)
		t.Skip("Integration test template; requires test database setup")
	})

	t.Run("partial_erp_sync_still_allows_manual_reconciliation", func(t *testing.T) {
		// Given: finalizeOrder() partially succeeds (SO created, Invoice fails)
		// When: transaction rolls back
		// Then: operator can retry finalization or create Invoice manually via ERP UI

		t.Skip("Integration test template; requires test database setup")
	})
}

// TestPOSFinalizationHappyPath verifies that successful finalization creates all ERP documents.
//
// Scenario:
// 1. POS order paid → POSPayment created
// 2. finalizeOrder() called
// 3. Order status → PAID
// 4. SalesOrder created
// 5. CustomerInvoice created
// 6. SalesPayment created
// 7. All changes committed (no rollback)
//
// This test ensures the happy path works as designed.
func TestPOSFinalizationHappyPath(t *testing.T) {
	t.Run("all_erp_documents_created_when_finalization_succeeds", func(t *testing.T) {
		// Given: POS order + payment ready for finalization
		// When: finalizeOrder() is called
		// Then: order status is PAID, SalesOrder created, Invoice created, Payment created

		// Setup would require:
		// db := setupTestDB()
		// usecase := NewPOSPaymentUsecase(db, ...)
		// order := createTestPOSOrder()
		// payment := createTestPOSPayment(order.ID)
		//
		// err := usecase.finalizeOrder(ctx, order, userID, payment)
		//
		// // Assert: no error
		// require.NoError(t, err)
		//
		// // Assert: order status is PAID
		// updatedOrder, _ := orderRepo.GetByID(ctx, order.ID)
		// require.Equal(t, models.PosOrderStatusPaid, updatedOrder.Status)
		//
		// // Assert: SalesOrder linked
		// require.NotNil(t, updatedOrder.SalesOrderID)
		// so, _ := soRepo.GetByID(ctx, *updatedOrder.SalesOrderID)
		// require.NotNil(t, so)
		//
		// // Assert: CustomerInvoice linked
		// require.NotNil(t, updatedOrder.CustomerInvoiceID)
		// invoice, _ := invoiceRepo.GetByID(ctx, *updatedOrder.CustomerInvoiceID)
		// require.NotNil(t, invoice)
		//
		// // Assert: SalesPayment created
		// payments, _ := paymentRepo.FindByInvoiceID(ctx, invoice.ID)
		// require.Len(t, payments, 1)
		t.Skip("Integration test template; requires test database setup")
	})
}

// failingInvoiceRepository is a mock repository that fails on Create to simulate failures.
// Used in atomic rollback tests.
type failingInvoiceRepository struct {
	underlying salesRepos.CustomerInvoiceRepository
}

func (r *failingInvoiceRepository) Create(ctx context.Context, invoice *models.CustomerInvoice) error {
	// Always fail to simulate a failure scenario
	return gorm.ErrInvalidData
}

// Delegate all other methods to underlying repository
func (r *failingInvoiceRepository) FindByID(ctx context.Context, id string) (*models.CustomerInvoice, error) {
	return r.underlying.FindByID(ctx, id)
}

func (r *failingInvoiceRepository) FindByCode(ctx context.Context, code string) (*models.CustomerInvoice, error) {
	return r.underlying.FindByCode(ctx, code)
}

func (r *failingInvoiceRepository) List(ctx context.Context, req *dto.ListCustomerInvoicesRequest) ([]models.CustomerInvoice, int64, error) {
	return r.underlying.List(ctx, req)
}

func (r *failingInvoiceRepository) ListItems(ctx context.Context, invoiceID string, req *dto.ListCustomerInvoiceItemsRequest) ([]models.CustomerInvoiceItem, int64, error) {
	return r.underlying.ListItems(ctx, invoiceID, req)
}

func (r *failingInvoiceRepository) Update(ctx context.Context, invoice *models.CustomerInvoice) error {
	return r.underlying.Update(ctx, invoice)
}

func (r *failingInvoiceRepository) UpdateStatus(ctx context.Context, id string, status models.CustomerInvoiceStatus, paidAmount *float64, paymentAt *time.Time, userID *string) error {
	return r.underlying.UpdateStatus(ctx, id, status, paidAmount, paymentAt, userID)
}

func (r *failingInvoiceRepository) GetNextInvoiceNumber(ctx context.Context, prefix string) (string, error) {
	return r.underlying.GetNextInvoiceNumber(ctx, prefix)
}

func (r *failingInvoiceRepository) Delete(ctx context.Context, id string) error {
	return r.underlying.Delete(ctx, id)
}
