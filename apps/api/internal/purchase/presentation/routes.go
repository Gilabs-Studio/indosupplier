package presentation

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	invUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/usecase"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gilabs/gims/api/internal/purchase/presentation/router"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PurchaseDeps holds exported Purchase usecases for cross-module consumption.
type PurchaseDeps struct {
	OrderUC           usecase.PurchaseOrderUsecase
	RequisitionUC     usecase.PurchaseRequisitionUsecase
	GoodsReceiptUC    usecase.GoodsReceiptUsecase
	SupplierInvoiceUC usecase.SupplierInvoiceUsecase
}

func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB, jwtManager *jwt.JWTManager, permService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}, invUC invUsecase.InventoryUsecase, journalUC finUsecase.JournalEntryUsecase, coaUC finUsecase.ChartOfAccountUsecase, assetUC finUsecase.AssetUsecase, engine accounting.AccountingEngine, settingsUC financesettings.SettingsService, cashBankTransactionUC finUsecase.CashBankTransactionUsecase) *PurchaseDeps {
	_ = r

	prRepo := repositories.NewPurchaseRequisitionRepository(db)
	poRepo := repositories.NewPurchaseOrderRepository(db)
	grRepo := repositories.NewGoodsReceiptRepository(db)
	siRepo := repositories.NewSupplierInvoiceRepository(db)
	payRepo := repositories.NewPurchasePaymentRepository(db)
	returnRepo := repositories.NewPurchaseReturnRepository(db)
	recapRepo := repositories.NewPayableRecapRepository(db)
	auditService := audit.NewAuditService(db)
	fiscalYearRepo := financeRepositories.NewFiscalYearRepository(db)

	poUc := usecase.NewPurchaseOrderUsecase(db, poRepo, prRepo, auditService, fiscalYearRepo)
	poH := handler.NewPurchaseOrderHandler(poUc)
	poPrintH := handler.NewPurchaseOrderPrintHandler(poUc, db)

	prUc := usecase.NewPurchaseRequisitionUsecase(db, prRepo, auditService)
	prH := handler.NewPurchaseRequisitionHandler(prUc, poUc)
	prPrintH := handler.NewPurchaseRequisitionPrintHandler(prUc, db)

	grUc := usecase.NewGoodsReceiptUsecase(db, grRepo, poRepo, auditService, invUC, journalUC, coaUC, assetUC, engine, fiscalYearRepo)
	grH := handler.NewGoodsReceiptHandler(grUc)
	grPrintH := handler.NewGoodsReceiptPrintHandler(grUc, db)

	siUc := usecase.NewSupplierInvoiceUsecase(db, siRepo, poRepo, grRepo, auditService, journalUC, coaUC, engine)
	siH := handler.NewSupplierInvoiceHandler(siUc)
	siPrintH := handler.NewSupplierInvoicePrintHandler(siUc, db)

	siDpUc := usecase.NewSupplierInvoiceDownPaymentUsecase(db, siRepo, poRepo, auditService, journalUC, coaUC, engine)
	siDpH := handler.NewSupplierInvoiceDownPaymentHandler(siDpUc)
	siDpPrintH := handler.NewSupplierInvoiceDPPrintHandler(siDpUc, db)

	payUc := usecase.NewPurchasePaymentUsecase(db, payRepo, siRepo, auditService, journalUC, coaUC, engine, settingsUC, cashBankTransactionUC)
	payH := handler.NewPurchasePaymentHandler(payUc)
	payPrintH := handler.NewPurchasePaymentPrintHandler(payUc, db)

	returnUC := usecase.NewPurchaseReturnUsecase(db, returnRepo, invUC, journalUC, coaUC, auditService, engine)
	returnH := handler.NewPurchaseReturnHandler(returnUC)

	recapUc := usecase.NewPayableRecapUsecase(recapRepo)
	recapH := handler.NewPayableRecapHandler(recapUc)

	group := api.Group("/purchase")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.TenantGuard())
	group.Use(middleware.ScopeMiddleware(db))

	router.RegisterPurchaseRequisitionRoutes(group, prH, prPrintH)
	router.RegisterPurchaseOrderRoutes(group, poH, poPrintH)
	router.RegisterGoodsReceiptRoutes(group, grH, grPrintH)
	router.RegisterSupplierInvoiceRoutes(group, siH, siPrintH)
	router.RegisterSupplierInvoiceDownPaymentRoutes(group, siDpH, siDpPrintH)
	router.RegisterPurchasePaymentRoutes(group, payH, payPrintH)
	router.RegisterPurchaseReturnRoutes(group, returnH)
	router.RegisterPayableRecapRoutes(group, recapH)

	return &PurchaseDeps{
		OrderUC:           poUc,
		RequisitionUC:     prUc,
		GoodsReceiptUC:    grUc,
		SupplierInvoiceUC: siUc,
	}
}
