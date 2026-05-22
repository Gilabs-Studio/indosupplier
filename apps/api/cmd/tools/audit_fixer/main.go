package main

import (
	"context"
	"flag"
	"log"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	// Finance
	finRepos "github.com/gilabs/gims/api/internal/finance/data/repositories"
	finAccounting "github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finSettings "github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	finMapper "github.com/gilabs/gims/api/internal/finance/domain/mapper"
	finService "github.com/gilabs/gims/api/internal/finance/domain/service"
	finUC "github.com/gilabs/gims/api/internal/finance/domain/usecase"

	// Purchase
	purchModels "github.com/gilabs/gims/api/internal/purchase/data/models"
	purchRepos "github.com/gilabs/gims/api/internal/purchase/data/repositories"
	// purchMapper "github.com/gilabs/gims/api/internal/purchase/domain/mapper"
	purchUC "github.com/gilabs/gims/api/internal/purchase/domain/usecase"

	// Sales
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	salesUC "github.com/gilabs/gims/api/internal/sales/domain/usecase"

	// Product
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"

	// Inventory
	invRepos "github.com/gilabs/gims/api/internal/inventory/data/repositories"
	invUC "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
)

type MockAuditService struct{}

func (m *MockAuditService) Log(ctx context.Context, action string, targetID string, metadata map[string]interface{}) {
}
func (m *MockAuditService) LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}) {
}
func (m *MockAuditService) LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{}) {
}
func (m *MockAuditService) LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{}) {
}

func main() {
	targetType := flag.String("type", "", "Type of record (si, ci, sp, pp, gr)")
	targetCode := flag.String("code", "", "Code of the record")
	flag.Parse()

	if *targetType == "" || *targetCode == "" {
		log.Fatal("Usage: go run cmd/tools/audit_fixer/main.go -type [si|ci|sp|pp|gr] -code [CODE]")
	}

	// Init Config
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init DB
	if err := database.Connect(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db := database.DB

	ctx := context.Background()
	auditService := &MockAuditService{}

	// Init Finance Deps
	coaRepo := finRepos.NewChartOfAccountRepository(db)
	journalRepo := finRepos.NewJournalEntryRepository(db)
	coaMapper := finMapper.NewChartOfAccountMapper()
	journalMapper := finMapper.NewJournalEntryMapper(coaMapper)
	journalUC := finUC.NewJournalEntryUsecase(db, coaRepo, journalRepo, journalMapper, auditService)
	financeSettingRepo := finRepos.NewFinanceSettingRepository(db)
	settingsService := finSettings.NewSettingsService(financeSettingRepo)
	coaValidationSvc := finService.NewCOAValidationService(financeSettingRepo)
	engine := finAccounting.NewAccountingEngine(settingsService, coaRepo, coaValidationSvc)
	coaUC := finUC.NewChartOfAccountUsecase(db, coaRepo, coaMapper)

	// Additional Finance Deps (Assets)
	assetRepo := finRepos.NewAssetRepository(db)
	assetCatRepo := finRepos.NewAssetCategoryRepository(db)
	assetLocRepo := finRepos.NewAssetLocationRepository(db)
	// Correct AssetMapper nested mappers
	assetCatMapper := finMapper.NewAssetCategoryMapper()
	assetLocMapper := finMapper.NewAssetLocationMapper()
	assetMapper := finMapper.NewAssetMapper(assetCatMapper, assetLocMapper)

	assetAttachmentRepo := finRepos.NewAssetAttachmentRepository(db)
	assetAuditLogRepo := finRepos.NewAssetAuditLogRepository(db)
	assetAssignmentRepo := finRepos.NewAssetAssignmentRepository(db)

	accountingPeriodRepo := finRepos.NewAccountingPeriodRepository(db)
	assetUC := finUC.NewAssetUsecase(db, coaRepo, assetCatRepo, assetLocRepo, accountingPeriodRepo, assetRepo, assetMapper, assetAttachmentRepo, assetAuditLogRepo, assetAssignmentRepo, journalUC)

	// Init Inventory Deps
	invRepo := invRepos.NewInventoryRepository(db)
	inventoryUC := invUC.NewInventoryUsecase(db, invRepo, journalUC, engine)

	switch *targetType {
	case "si": // Supplier Invoice
		siRepo := purchRepos.NewSupplierInvoiceRepository(db)
		poRepo := purchRepos.NewPurchaseOrderRepository(db)
		grRepo := purchRepos.NewGoodsReceiptRepository(db)
		siUC := purchUC.NewSupplierInvoiceUsecase(db, siRepo, poRepo, grRepo, auditService, journalUC, coaUC, engine)

		var si purchModels.SupplierInvoice
		if err := db.Where("code = ?", *targetCode).First(&si).Error; err != nil {
			log.Fatalf("SI not found: %v", err)
		}

		log.Printf("Re-triggering journal for SI: %s (ID: %s)", si.Code, si.ID)
		if err := siUC.TriggerJournalForSupplierInvoice(ctx, &si); err != nil {
			log.Fatalf("Failed to trigger journal: %v", err)
		}
		log.Println("✅ Successfully triggered journal for SI")

	case "ci": // Customer Invoice
		ciRepo := salesRepos.NewCustomerInvoiceRepository(db)
		productRepo := productRepos.NewProductRepository(db)
		soRepo := salesRepos.NewSalesOrderRepository(db)
		ciUC := salesUC.NewCustomerInvoiceUsecase(db, ciRepo, productRepo, soRepo, journalUC, coaUC, auditService, engine, nil)

		var ci salesModels.CustomerInvoice
		if err := db.Where("code = ?", *targetCode).First(&ci).Error; err != nil {
			log.Fatalf("CI not found: %v", err)
		}

		log.Printf("Re-triggering journal for CI: %s (ID: %s)", ci.Code, ci.ID)
		if err := ciUC.TriggerJournalForInvoice(ctx, &ci); err != nil {
			log.Fatalf("Failed to trigger journal: %v", err)
		}
		log.Println("✅ Successfully triggered journal for CI")

	case "sp": // Sales Payment
		spRepo := salesRepos.NewSalesPaymentRepository(db)
		spUC := salesUC.NewSalesPaymentUsecase(db, spRepo, auditService, journalUC, coaUC, engine, settingsService)

		var sp salesModels.SalesPayment
		if err := db.Where("reference_number = ?", *targetCode).First(&sp).Error; err != nil {
			if err := db.Where("notes LIKE ?", "%"+*targetCode+"%").First(&sp).Error; err != nil {
				log.Fatalf("SP not found: %v", err)
			}
		}

		log.Printf("Re-triggering journal for SP (ID: %s)", sp.ID)
		if err := spUC.TriggerJournalForPayment(ctx, &sp); err != nil {
			log.Fatalf("Failed to trigger journal: %v", err)
		}
		log.Println("✅ Successfully triggered journal for SP")

	case "pp": // Purchase Payment
		ppRepo := purchRepos.NewPurchasePaymentRepository(db)
		siRepo := purchRepos.NewSupplierInvoiceRepository(db)
		ppUC := purchUC.NewPurchasePaymentUsecase(db, ppRepo, siRepo, auditService, journalUC, coaUC, engine, settingsService, nil)

		var pp purchModels.PurchasePayment
		if err := db.Where("reference_number = ?", *targetCode).First(&pp).Error; err != nil {
			log.Fatalf("PP not found: %v", err)
		}

		log.Printf("Re-triggering journal for PP (ID: %s)", pp.ID)
		if err := ppUC.TriggerJournalForPayment(ctx, &pp); err != nil {
			log.Fatalf("Failed to trigger journal: %v", err)
		}
		log.Println("✅ Successfully triggered journal for PP")

	case "gr": // Goods Receipt
		grRepo := purchRepos.NewGoodsReceiptRepository(db)
		poRepo := purchRepos.NewPurchaseOrderRepository(db)
		fiscalYearRepo := finRepos.NewFiscalYearRepository(db)
		grUC := purchUC.NewGoodsReceiptUsecase(db, grRepo, poRepo, auditService, inventoryUC, journalUC, coaUC, assetUC, engine, fiscalYearRepo)

		var gr purchModels.GoodsReceipt
		if err := db.Where("code = ?", *targetCode).First(&gr).Error; err != nil {
			log.Fatalf("GR not found: %v", err)
		}

		log.Printf("Re-triggering journal for GR: %s (ID: %s)", gr.Code, gr.ID)
		if err := grUC.TriggerJournalForReconciliation(ctx, &gr); err != nil {
			log.Fatalf("Failed to trigger journal: %v", err)
		}
		log.Println("✅ Successfully triggered journal for GR")

	default:
		log.Fatalf("Unsupported type: %s", *targetType)
	}
}
