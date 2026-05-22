package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	invDto "github.com/gilabs/gims/api/internal/inventory/domain/dto"
	invUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/mapper"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrGoodsReceiptNotFound = errors.New("goods receipt not found")
	ErrGoodsReceiptConflict = errors.New("goods receipt conflict")
	ErrGoodsReceiptInvalid  = errors.New("invalid goods receipt request")
)

type GoodsReceiptUsecase interface {
	List(ctx context.Context, params repositories.GoodsReceiptListParams) ([]*dto.GoodsReceiptListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateGoodsReceiptRequest) (*dto.GoodsReceiptDetailResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateGoodsReceiptRequest) (*dto.GoodsReceiptDetailResponse, error)
	Delete(ctx context.Context, id string) error
	// Legacy confirm: DRAFT → CONFIRMED (kept for backward compatibility).
	Confirm(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	// New workflow: DRAFT → SUBMITTED → APPROVED → CLOSED.
	Submit(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	Approve(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	Reject(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	Close(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error)
	// ConvertToSupplierInvoice creates a draft Supplier Invoice from a CLOSED GR.
	ConvertToSupplierInvoice(ctx context.Context, id string) (*dto.GoodsReceiptConvertResponse, error)
	AddData(ctx context.Context) (*dto.GoodsReceiptAddResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.GoodsReceiptAuditTrailEntry, int64, error)
	TriggerJournalForReconciliation(ctx context.Context, gr *models.GoodsReceipt) error
}

type goodsReceiptUsecase struct {
	db                 *gorm.DB
	repo               repositories.GoodsReceiptRepository
	poRepo             repositories.PurchaseOrderRepository
	fiscalYearResolver *purchaseFiscalYearResolver
	mapper             *mapper.GoodsReceiptMapper
	auditService       audit.AuditService
	inventoryUC        invUsecase.InventoryUsecase
	journalUC          finUsecase.JournalEntryUsecase
	coaUC              finUsecase.ChartOfAccountUsecase
	assetUC            finUsecase.AssetUsecase
	engine             accounting.AccountingEngine
}

func NewGoodsReceiptUsecase(db *gorm.DB, repo repositories.GoodsReceiptRepository, poRepo repositories.PurchaseOrderRepository, auditService audit.AuditService, inventoryUC invUsecase.InventoryUsecase, journalUC finUsecase.JournalEntryUsecase, coaUC finUsecase.ChartOfAccountUsecase, assetUC finUsecase.AssetUsecase, engine accounting.AccountingEngine, fiscalYearRepo financeRepositories.FiscalYearRepository) GoodsReceiptUsecase {
	return &goodsReceiptUsecase{
		db:                 db,
		repo:               repo,
		poRepo:             poRepo,
		fiscalYearResolver: newPurchaseFiscalYearResolver(db, fiscalYearRepo),
		mapper:             mapper.NewGoodsReceiptMapper(),
		auditService:       auditService,
		inventoryUC:        inventoryUC,
		journalUC:          journalUC,
		coaUC:              coaUC,
		assetUC:            assetUC,
		engine:             engine,
	}
}

func (uc *goodsReceiptUsecase) List(ctx context.Context, params repositories.GoodsReceiptListParams) ([]*dto.GoodsReceiptListResponse, int64, error) {
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return uc.mapper.ToListResponseList(items), total, nil
}

func (uc *goodsReceiptUsecase) GetByID(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	// Fetch first (repo accepts UUID or code), then verify data-scope access using the real UUID.
	gr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}

	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.GoodsReceipt{}, gr.ID, security.PurchaseScopeQueryOptions()) {
		return nil, ErrGoodsReceiptNotFound
	}

	return uc.mapper.ToDetailResponse(gr), nil
}

func (uc *goodsReceiptUsecase) Create(ctx context.Context, req *dto.CreateGoodsReceiptRequest) (*dto.GoodsReceiptDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	warehouseID := strings.TrimSpace(req.WarehouseID)
	if warehouseID == "" {
		return nil, ErrGoodsReceiptInvalid
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var createdID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock by purchase order ID to prevent concurrent draft creation.
		lockKey := "goods_receipt_create:" + strings.TrimSpace(req.PurchaseOrderID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
			return err
		}

		var po models.PurchaseOrder
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&po, "id = ?", req.PurchaseOrderID).Error; err != nil {
			return err
		}
		if po.Status != models.PurchaseOrderStatusApproved {
			return ErrInvalidStatus
		}
		if err := validateActiveWarehouseID(ctx, tx, warehouseID); err != nil {
			return err
		}

		supplierID := ""
		if po.SupplierID != nil {
			supplierID = strings.TrimSpace(*po.SupplierID)
		}
		if supplierID == "" {
			return errors.New("purchase order supplier is empty")
		}

		code, err := getNextGoodsReceiptCodeLocked(ctx, tx, "GR")
		if err != nil {
			return err
		}

		var receiptDate *time.Time
		if req.ReceiptDate != nil && strings.TrimSpace(*req.ReceiptDate) != "" {
			parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ReceiptDate))
			if err != nil {
				return fmt.Errorf("invalid receipt_date format: %w", err)
			}
			receiptDate = &parsed
		}

		gr := &models.GoodsReceipt{
			Code:            code,
			PurchaseOrderID: po.ID,
			WarehouseID:     &warehouseID,
			SupplierID:      supplierID,
			ReceiptDate:     receiptDate,
			Notes:           req.Notes,
			ProofImageURL:   req.ProofImageURL,
			Status:          models.GoodsReceiptStatusDraft,
			CreatedBy:       actorID,
			Items:           make([]models.GoodsReceiptItem, 0, len(req.Items)),
		}
		gr.CompanyID = po.CompanyID
		gr.FiscalYearID = po.FiscalYearID
		if gr.CompanyID == nil || gr.FiscalYearID == nil {
			companyID, fiscalYearID, err := uc.fiscalYearResolver.Resolve(ctx, po.OrderDate)
			if err != nil {
				return err
			}
			gr.CompanyID = &companyID
			gr.FiscalYearID = &fiscalYearID
		}

		poItemByID := make(map[string]*models.PurchaseOrderItem, len(po.Items))
		for i := range po.Items {
			poIt := &po.Items[i]
			poItemByID[poIt.ID] = poIt
		}

		for _, it := range req.Items {
			poItemID := strings.TrimSpace(it.PurchaseOrderItemID)
			if poItemID == "" {
				return ErrGoodsReceiptConflict
			}
			poIt := poItemByID[poItemID]
			if poIt == nil {
				return ErrGoodsReceiptConflict
			}
			if strings.TrimSpace(it.ProductID) == "" || strings.TrimSpace(it.ProductID) != strings.TrimSpace(poIt.ProductID) {
				return ErrGoodsReceiptConflict
			}

			qty := math.Max(0, it.QuantityReceived)
			gr.Items = append(gr.Items, models.GoodsReceiptItem{
				PurchaseOrderItemID: poItemID,
				ProductID:           it.ProductID,
				QuantityReceived:    qty,
				Notes:               it.Notes,
			})
		}

		if err := snapshotGoodsReceipt(ctx, tx, gr, nil); err != nil {
			return err
		}

		if err := tx.Omit("Items").Create(gr).Error; err != nil {
			return err
		}
		if len(gr.Items) > 0 {
			for i := range gr.Items {
				gr.Items[i].GoodsReceiptID = gr.ID
				gr.Items[i].QuantityReceived = math.Max(0, gr.Items[i].QuantityReceived)
			}
			if err := tx.Create(&gr.Items).Error; err != nil {
				return err
			}
		}

		createdID = gr.ID
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}

	created, err := uc.repo.GetByID(ctx, createdID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "goods_receipt.create", created.ID, map[string]interface{}{
		"after": grAuditSnapshot(created),
	})

	return uc.mapper.ToDetailResponse(created), nil
}

func (uc *goodsReceiptUsecase) Update(ctx context.Context, id string, req *dto.UpdateGoodsReceiptRequest) (*dto.GoodsReceiptDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	warehouseID := strings.TrimSpace(req.WarehouseID)
	if warehouseID == "" {
		return nil, ErrGoodsReceiptInvalid
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	if existing.Status != models.GoodsReceiptStatusDraft {
		return nil, ErrGoodsReceiptConflict
	}
	if err := validateActiveWarehouseID(ctx, uc.db, warehouseID); err != nil {
		return nil, err
	}
	before := grAuditSnapshot(existing)

	receiptDate := existing.ReceiptDate
	if req.ReceiptDate != nil && strings.TrimSpace(*req.ReceiptDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ReceiptDate))
		if err != nil {
			return nil, fmt.Errorf("invalid receipt_date format: %w", err)
		}
		receiptDate = &parsed
	}

	gr := &models.GoodsReceipt{
		ID:              existing.ID,
		Code:            existing.Code,
		PurchaseOrderID: existing.PurchaseOrderID,
		WarehouseID:     &warehouseID,
		SupplierID:      existing.SupplierID,
		ReceiptDate:     receiptDate,
		Notes:           req.Notes,
		ProofImageURL:   req.ProofImageURL,
		Status:          existing.Status,
		CreatedBy:       existing.CreatedBy,
		Items:           make([]models.GoodsReceiptItem, 0, len(req.Items)),
	}

	po := existing.PurchaseOrder
	if po == nil {
		return nil, ErrGoodsReceiptConflict
	}
	poItemByID := make(map[string]*models.PurchaseOrderItem, len(po.Items))
	for i := range po.Items {
		poIt := &po.Items[i]
		poItemByID[poIt.ID] = poIt
	}

	for _, it := range req.Items {
		poItemID := strings.TrimSpace(it.PurchaseOrderItemID)
		if poItemID == "" {
			return nil, ErrGoodsReceiptConflict
		}
		poIt := poItemByID[poItemID]
		if poIt == nil {
			return nil, ErrGoodsReceiptConflict
		}
		if strings.TrimSpace(it.ProductID) == "" || strings.TrimSpace(it.ProductID) != strings.TrimSpace(poIt.ProductID) {
			return nil, ErrGoodsReceiptConflict
		}

		qty := math.Max(0, it.QuantityReceived)
		gr.Items = append(gr.Items, models.GoodsReceiptItem{
			PurchaseOrderItemID: poItemID,
			ProductID:           it.ProductID,
			QuantityReceived:    qty,
			Notes:               it.Notes,
		})
	}

	if err := snapshotGoodsReceipt(ctx, uc.db, gr, existing); err != nil {
		return nil, err
	}

	updated, err := uc.repo.Update(ctx, gr)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "goods_receipt.update", updated.ID, map[string]interface{}{
		"before": before,
		"after":  grAuditSnapshot(updated),
	})

	return uc.mapper.ToDetailResponse(updated), nil
}

func (uc *goodsReceiptUsecase) Delete(ctx context.Context, id string) error {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return errors.New("user not authenticated")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrGoodsReceiptNotFound
		}
		return err
	}
	if existing.Status != models.GoodsReceiptStatusDraft {
		return ErrGoodsReceiptConflict
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditService.Log(ctx, "goods_receipt.delete", id, map[string]interface{}{
		"before": grAuditSnapshot(existing),
	})
	return nil
}

func (uc *goodsReceiptUsecase) Confirm(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	// Use a transaction to prevent over-receiving on concurrent confirms.
	var out *models.GoodsReceipt
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var gr models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&gr, "id = ?", id).Error; err != nil {
			return err
		}
		if gr.Status != models.GoodsReceiptStatusDraft {
			return ErrGoodsReceiptConflict
		}

		// Lock purchase order as well.
		var po models.PurchaseOrder
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&po, "id = ?", gr.PurchaseOrderID).Error; err != nil {
			return err
		}
		if po.Status != models.PurchaseOrderStatusApproved {
			return ErrInvalidStatus
		}

		// Validate quantities do not exceed ordered qty (sum of confirmed GRs + this GR).
		for _, it := range gr.Items {
			ordered := 0.0
			for _, poIt := range po.Items {
				if poIt.ID == it.PurchaseOrderItemID {
					ordered = poIt.Quantity
					break
				}
			}

			var alreadyReceived float64
			alreadyReceivedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
				Table("goods_receipt_items")
			alreadyReceivedQuery, err := applyTenantJoinScope(ctx, alreadyReceivedQuery, "goods_receipt_items.tenant_id", "goods_receipts.tenant_id")
			if err != nil {
				return err
			}

			if err := alreadyReceivedQuery.
				Select("COALESCE(SUM(goods_receipt_items.quantity_received),0)").
				Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
				Where("goods_receipts.purchase_order_id = ?", po.ID).
				Where("goods_receipts.status = ?", models.GoodsReceiptStatusConfirmed).
				Where("goods_receipt_items.purchase_order_item_id = ?", it.PurchaseOrderItemID).
				Scan(&alreadyReceived).Error; err != nil {
				return err
			}

			receiving := math.Max(0, it.QuantityReceived)
			if alreadyReceived+receiving > ordered+0.0001 {
				return ErrGoodsReceiptConflict
			}
		}

		now := apptime.Now()
		if err := tx.Model(&gr).Updates(map[string]interface{}{
			"status":       models.GoodsReceiptStatusConfirmed,
			"receipt_date": now,
		}).Error; err != nil {
			return err
		}

		// Close PO if fully received.
		fullyReceived := true
		for _, poIt := range po.Items {
			var totalReceived float64
			totalReceivedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
				Table("goods_receipt_items")
			totalReceivedQuery, err := applyTenantJoinScope(ctx, totalReceivedQuery, "goods_receipt_items.tenant_id", "goods_receipts.tenant_id")
			if err != nil {
				return err
			}

			if err := totalReceivedQuery.
				Select("COALESCE(SUM(goods_receipt_items.quantity_received),0)").
				Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
				Where("goods_receipts.purchase_order_id = ?", po.ID).
				Where("goods_receipts.status = ?", models.GoodsReceiptStatusConfirmed).
				Where("goods_receipt_items.purchase_order_item_id = ?", poIt.ID).
				Scan(&totalReceived).Error; err != nil {
				return err
			}
			if totalReceived+0.0001 < poIt.Quantity {
				fullyReceived = false
				break
			}
		}
		if fullyReceived {
			// Only close PO if at least one NORMAL supplier invoice exists and all NORMAL
			// invoices are paid/cancelled/rejected. DP-only settlement must not close PO.
			var totalInvoiceCount3 int64
			tx.Model(&models.SupplierInvoice{}).
				Where("purchase_order_id = ?", po.ID).
				Where("type = ?", models.SupplierInvoiceTypeNormal).
				Count(&totalInvoiceCount3)
			var pendingInvoiceCount int64
			tx.Model(&models.SupplierInvoice{}).
				Where("purchase_order_id = ?", po.ID).
				Where("type = ?", models.SupplierInvoiceTypeNormal).
				Where("status NOT IN ?", []models.SupplierInvoiceStatus{
					models.SupplierInvoiceStatusPaid,
					models.SupplierInvoiceStatusCancelled,
					models.SupplierInvoiceStatusRejected,
				}).
				Count(&pendingInvoiceCount)
			if totalInvoiceCount3 > 0 && pendingInvoiceCount == 0 {
				_ = tx.Model(&po).Update("status", models.PurchaseOrderStatusClosed).Error
			}
		}

		loaded, err := uc.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		out = loaded

		// Trigger Inventory Update (Inside Transaction)
		txCtx := database.WithTx(ctx, tx)
		if err := uc.triggerStockUpdate(txCtx, out); err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		// Trigger Journal Entry via Inventory Module (Consolidated)
		if err := uc.inventoryUC.TriggerDocumentJournal(txCtx, tx, reference.RefTypeGoodsReceipt, out.ID); err != nil {
			return fmt.Errorf("failed to create journal: %w", err)
		}

		// Trigger Asset Creation (Inside Transaction)
		// Only for products categorized as "Device" (Asset)
		if err := uc.triggerAssetCreation(txCtx, out); err != nil {
			// Asset creation failure shouldn't necessarily block GR if not critical,
			// but for this ERP we want reliability.
			return fmt.Errorf("failed to create asset: %w", err)
		}

		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}

	uc.auditService.Log(ctx, "goods_receipt.confirm", id, map[string]interface{}{
		"after": grAuditSnapshot(out),
	})

	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *goodsReceiptUsecase) triggerStockUpdate(ctx context.Context, gr *models.GoodsReceipt) error {
	if gr == nil || gr.WarehouseID == nil || strings.TrimSpace(*gr.WarehouseID) == "" {
		return ErrGoodsReceiptInvalid
	}

	items := make([]invDto.ReceiveStockItem, 0, len(gr.Items))

	// Need to fetch Cost Price from PO Items?
	// GR Items only have Qty. PO Items have Price.
	// We need to query PO items to get price.
	// Optimally, we should load PO with GR.
	// In Confirm(), we loaded PO. Let's pass PO data or reload.
	// Since this is outside TX, we reload carefully or query join.

	// Quickest way: Query PurchaseOrderItems
	// Or use loaded 'out' from Confirm?
	// 'out' variable in Confirm has GR. 'gr.PurchaseOrder' might not be loaded deeply.

	// Let's implement strict fetching here.

	// Select Price from purchase_order_items
	// This assumes 1 PO Item -> 1 Product.
	// We need effective price (Price - Disc + Tax? No, usually Inventory Valuation uses Net Price before Tax, or Standard Cost).
	// Let's use Base Price for now.

	var poItems []models.PurchaseOrderItem
	if err := database.GetDB(ctx, uc.db).Where("purchase_order_id = ?", gr.PurchaseOrderID).Find(&poItems).Error; err != nil {
		return err
	}

	priceMap := make(map[string]float64)
	for _, pid := range poItems {
		priceMap[pid.ID] = pid.Price // Ignoring discount for HPP? Should we include discount?
		// HPP = (Price - Discount) normally.
		// Let's calculate Net Price if possible.
		// Model PurchaseOrderItem has Subtotal, TaxAmount, Total.
		// Let's use Price for simplicity as requested "ReceiveStockFromGR" usually takes unit cost.
	}

	warehouseID := strings.TrimSpace(*gr.WarehouseID)

	for _, it := range gr.Items {
		cost := priceMap[it.PurchaseOrderItemID]
		items = append(items, invDto.ReceiveStockItem{
			ProductID:   it.ProductID,
			Quantity:    it.QuantityReceived,
			CostPrice:   cost,
			BatchNumber: nil, // Auto generate
			ExpiryDate:  nil, // Not captured in GR currently
		})
	}

	notes := ""
	if gr.Notes != nil {
		notes = *gr.Notes
	}

	supplierName := ""
	if gr.Supplier != nil {
		supplierName = strings.TrimSpace(gr.Supplier.Name)
	}
	if supplierName == "" {
		supplierName = strings.TrimSpace(gr.SupplierNameSnapshot)
	}

	req := &invDto.ReceiveStockRequest{
		SourceID:     gr.ID,
		SourceNumber: gr.Code,
		SourceType:   "GR",
		Source:       supplierName,
		WarehouseID:  warehouseID,
		Items:        items,
		ReceivedAt:   apptime.Now(),
		ReceivedBy:   gr.CreatedBy,
		Notes:        notes,
	}

	return uc.inventoryUC.ReceiveStockFromGR(ctx, req)
}

func (uc *goodsReceiptUsecase) AddData(ctx context.Context) (*dto.GoodsReceiptAddResponse, error) {
	// Eligible: purchase orders in APPROVED status
	items, _, err := uc.poRepo.List(ctx, repositories.PurchaseOrderListParams{
		Status:    string(models.PurchaseOrderStatusApproved),
		SortBy:    "created_at",
		SortDir:   "desc",
		Limit:     100,
		Offset:    0,
		WithItems: true,
	})
	if err != nil {
		return nil, err
	}

	if uc.db == nil || len(items) == 0 {
		return &dto.GoodsReceiptAddResponse{EligiblePurchaseOrders: []dto.GoodsReceiptPurchaseOrderOption{}}, nil
	}

	poIDs := make([]string, 0, len(items))
	for _, po := range items {
		poIDs = append(poIDs, po.ID)
	}

	// Compute total received qty per PO item from all active GRs (exclude REJECTED only).
	// This matches PO detail remaining-qty logic and prevents offering POs that are already
	// effectively consumed by in-progress GRs.
	type receivedRow struct {
		PurchaseOrderID     string  `gorm:"column:purchase_order_id"`
		PurchaseOrderItemID string  `gorm:"column:purchase_order_item_id"`
		TotalReceived       float64 `gorm:"column:total_received"`
	}
	receivedRows := make([]receivedRow, 0)
	receivedRowsQuery := uc.db.WithContext(ctx).
		Table("goods_receipt_items")
	receivedRowsQuery, err = applyTenantJoinScope(ctx, receivedRowsQuery, "goods_receipt_items.tenant_id", "gr.tenant_id")
	if err != nil {
		return nil, err
	}

	if err := receivedRowsQuery.
		Select("gr.purchase_order_id, goods_receipt_items.purchase_order_item_id, COALESCE(SUM(goods_receipt_items.quantity_received),0) AS total_received").
		Joins("JOIN goods_receipts gr ON gr.id = goods_receipt_items.goods_receipt_id").
		Where("gr.purchase_order_id IN ?", poIDs).
		Where("UPPER(gr.status) <> ?", string(models.GoodsReceiptStatusRejected)).
		Group("gr.purchase_order_id, goods_receipt_items.purchase_order_item_id").
		Scan(&receivedRows).Error; err != nil {
		return nil, err
	}
	// receivedMap[poID][itemID] = totalReceived
	receivedMap := make(map[string]map[string]float64, len(poIDs))
	for _, r := range receivedRows {
		if receivedMap[r.PurchaseOrderID] == nil {
			receivedMap[r.PurchaseOrderID] = make(map[string]float64)
		}
		receivedMap[r.PurchaseOrderID][r.PurchaseOrderItemID] += r.TotalReceived
	}

	res := make([]dto.GoodsReceiptPurchaseOrderOption, 0, len(items))
	for _, po := range items {
		// Skip POs that are 100% fulfilled by active GRs.
		if len(po.Items) > 0 {
			fullyFulfilled := true
			for _, poIt := range po.Items {
				received := receivedMap[po.ID][poIt.ID]
				if received+0.0001 < poIt.Quantity {
					fullyFulfilled = false
					break
				}
			}
			if fullyFulfilled {
				continue
			}
		}

		opt := dto.GoodsReceiptPurchaseOrderOption{ID: po.ID, Code: po.Code, Status: string(po.Status)}
		if po.Supplier != nil {
			opt.Supplier = &dto.GoodsReceiptSupplierMini{ID: po.Supplier.ID, Name: po.Supplier.Name}
		}
		res = append(res, opt)
	}

	return &dto.GoodsReceiptAddResponse{EligiblePurchaseOrders: res}, nil
}

// Submit transitions a GR from DRAFT to SUBMITTED.
func (uc *goodsReceiptUsecase) Submit(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	if existing.Status != models.GoodsReceiptStatusDraft {
		return nil, ErrGoodsReceiptConflict
	}
	if existing.WarehouseID == nil || strings.TrimSpace(*existing.WarehouseID) == "" {
		return nil, ErrGoodsReceiptInvalid
	}
	if err := validateActiveWarehouseID(ctx, uc.db, *existing.WarehouseID); err != nil {
		return nil, err
	}
	if err := uc.validateGoodsReceiptItemsAgainstPO(ctx, existing); err != nil {
		return nil, err
	}
	before := grAuditSnapshot(existing)

	now := apptime.Now()
	if err := database.GetDB(ctx, uc.db).Model(&models.GoodsReceipt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.GoodsReceiptStatusSubmitted,
			"submitted_at": now,
		}).Error; err != nil {
		return nil, err
	}

	updated, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "goods_receipt.submit", id, map[string]interface{}{
		"before": before,
		"after":  grAuditSnapshot(updated),
	})
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "goods_receipt.approve",
		EntityType:     "goods_receipt",
		EntityID:       updated.ID,
		Title:          "Goods Receipt Approval",
		Message:        "A goods receipt has been submitted and requires your approval.",
		ActorUserID:    actorID,
	}); err != nil {
		log.Printf("warning: failed to create goods receipt notification: %v", err)
	}
	return uc.mapper.ToDetailResponse(updated), nil
}

// Approve transitions a GR from SUBMITTED to APPROVED.
func (uc *goodsReceiptUsecase) Approve(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	if existing.Status != models.GoodsReceiptStatusSubmitted {
		return nil, ErrGoodsReceiptConflict
	}
	before := grAuditSnapshot(existing)

	now := apptime.Now()
	var out *models.GoodsReceipt
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update status to Approved
		if err := tx.Model(&models.GoodsReceipt{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":      models.GoodsReceiptStatusApproved,
				"approved_at": now,
			}).Error; err != nil {
			return err
		}

		// 2. Fetch loaded GR for side effects
		loaded, err := uc.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		out = loaded

		// 3. Trigger Stock & Financials (Atomic)
		txCtx := database.WithTx(ctx, tx)
		if err := uc.triggerStockUpdate(txCtx, out); err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}
		if err := uc.inventoryUC.TriggerDocumentJournal(txCtx, tx, reference.RefTypeGoodsReceipt, out.ID); err != nil {
			log.Printf("warning: failed to create GR journal on approve (non-blocking): %v", err)
		}
		if err := uc.triggerAssetCreation(txCtx, out); err != nil {
			log.Printf("warning: failed to create GR asset records on approve (non-blocking): %v", err)
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}

	uc.auditService.Log(ctx, "goods_receipt.approve", id, map[string]interface{}{
		"before": before,
		"after":  grAuditSnapshot(out),
	})
	return uc.mapper.ToDetailResponse(out), nil
}

// Reject transitions a GR from SUBMITTED to REJECTED.
func (uc *goodsReceiptUsecase) Reject(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	if existing.Status != models.GoodsReceiptStatusSubmitted {
		return nil, ErrGoodsReceiptConflict
	}
	before := grAuditSnapshot(existing)

	now := apptime.Now()
	if err := database.GetDB(ctx, uc.db).Model(&models.GoodsReceipt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      models.GoodsReceiptStatusRejected,
			"rejected_at": now,
		}).Error; err != nil {
		return nil, err
	}

	updated, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "goods_receipt.reject", id, map[string]interface{}{
		"before": before,
		"after":  grAuditSnapshot(updated),
	})
	return uc.mapper.ToDetailResponse(updated), nil
}

// Close transitions a GR from APPROVED to CLOSED, triggering stock/journal/asset side effects.
func (uc *goodsReceiptUsecase) Close(ctx context.Context, id string) (*dto.GoodsReceiptDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var out *models.GoodsReceipt
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var gr models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&gr, "id = ?", id).Error; err != nil {
			return err
		}
		if gr.Status != models.GoodsReceiptStatusApproved {
			return ErrGoodsReceiptConflict
		}

		var po models.PurchaseOrder
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&po, "id = ?", gr.PurchaseOrderID).Error; err != nil {
			return err
		}

		// Validate quantities do not exceed ordered qty (sum of CLOSED GRs + this GR).
		for _, it := range gr.Items {
			ordered := 0.0
			for _, poIt := range po.Items {
				if poIt.ID == it.PurchaseOrderItemID {
					ordered = poIt.Quantity
					break
				}
			}

			var alreadyReceived float64
			if err := tx.
				Table("goods_receipt_items").
				Select("COALESCE(SUM(goods_receipt_items.quantity_received),0)").
				Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
				Where("goods_receipts.purchase_order_id = ?", po.ID).
				Where("goods_receipts.status IN ?", []string{
					string(models.GoodsReceiptStatusClosed),
				}).
				Where("goods_receipt_items.purchase_order_item_id = ?", it.PurchaseOrderItemID).
				Scan(&alreadyReceived).Error; err != nil {
				return err
			}

			receiving := math.Max(0, it.QuantityReceived)
			if alreadyReceived+receiving > ordered+0.0001 {
				return ErrGoodsReceiptConflict
			}
		}

		now := apptime.Now()
		if err := tx.Model(&gr).Updates(map[string]interface{}{
			"status":    models.GoodsReceiptStatusClosed,
			"closed_at": now,
		}).Error; err != nil {
			return err
		}

		// Close PO if all items are fully received.
		fullyReceived := true
		for _, poIt := range po.Items {
			var totalReceived float64
			if err := tx.
				Table("goods_receipt_items").
				Select("COALESCE(SUM(goods_receipt_items.quantity_received),0)").
				Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
				Where("goods_receipts.purchase_order_id = ?", po.ID).
				Where("goods_receipts.status IN ?", []string{
					string(models.GoodsReceiptStatusClosed),
				}).
				Where("goods_receipt_items.purchase_order_item_id = ?", poIt.ID).
				Scan(&totalReceived).Error; err != nil {
				return err
			}
			if totalReceived+0.0001 < poIt.Quantity {
				fullyReceived = false
				break
			}
		}
		if fullyReceived {
			// Only close PO if at least one NORMAL supplier invoice exists and all NORMAL
			// invoices are paid/cancelled/rejected. DP-only settlement must not close PO.
			var totalInvoiceCount int64
			tx.Model(&models.SupplierInvoice{}).
				Where("purchase_order_id = ?", po.ID).
				Where("type = ?", models.SupplierInvoiceTypeNormal).
				Count(&totalInvoiceCount)
			var pendingInvoiceCount int64
			tx.Model(&models.SupplierInvoice{}).
				Where("purchase_order_id = ?", po.ID).
				Where("type = ?", models.SupplierInvoiceTypeNormal).
				Where("status NOT IN ?", []models.SupplierInvoiceStatus{
					models.SupplierInvoiceStatusPaid,
					models.SupplierInvoiceStatusCancelled,
					models.SupplierInvoiceStatusRejected,
				}).
				Count(&pendingInvoiceCount)
			if totalInvoiceCount > 0 && pendingInvoiceCount == 0 {
				_ = tx.Model(&po).Update("status", models.PurchaseOrderStatusClosed).Error
			}
		}

		loaded, err := uc.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		out = loaded

		// Note: Stock & Journal are now triggered at APPROVE status per UI workflow.
		// Close just marks the document as terminal.

		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}

	uc.auditService.Log(ctx, "goods_receipt.close", id, map[string]interface{}{
		"after": grAuditSnapshot(out),
	})
	return uc.mapper.ToDetailResponse(out), nil
}

// ConvertToSupplierInvoice creates a draft Supplier Invoice from a CLOSED Goods Receipt.
// Supplier data is taken from the GR (not the PO) to ensure accuracy.
func (uc *goodsReceiptUsecase) ConvertToSupplierInvoice(ctx context.Context, id string) (*dto.GoodsReceiptConvertResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	gr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	if gr.Status != models.GoodsReceiptStatusClosed {
		return nil, ErrGoodsReceiptConflict
	}

	// Build PO item price map for populating SI item prices.
	var poItems []models.PurchaseOrderItem
	if err := uc.db.WithContext(ctx).
		Where("purchase_order_id = ?", gr.PurchaseOrderID).
		Find(&poItems).Error; err != nil {
		return nil, err
	}
	priceMap := make(map[string]float64, len(poItems))
	for _, p := range poItems {
		priceMap[p.ID] = p.Price
	}

	var siID string
	today := apptime.Now()

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Generate unique SI code using advisory lock.
		code, err := getNextSupplierInvoiceCodeLockedInGR(ctx, tx, "SI")
		if err != nil {
			return fmt.Errorf("failed to generate SI code: %w", err)
		}

		supplierCode := ""
		supplierName := ""
		if gr.Supplier != nil {
			supplierCode = gr.Supplier.Code
			supplierName = gr.Supplier.Name
		} else if gr.SupplierCodeSnapshot != "" || gr.SupplierNameSnapshot != "" {
			supplierCode = gr.SupplierCodeSnapshot
			supplierName = gr.SupplierNameSnapshot
		}

		grID := gr.ID
		si := &models.SupplierInvoice{
			Type:                 models.SupplierInvoiceTypeNormal,
			PurchaseOrderID:      gr.PurchaseOrderID,
			GoodsReceiptID:       &grID,
			SupplierID:           gr.SupplierID,
			SupplierCodeSnapshot: supplierCode,
			SupplierNameSnapshot: supplierName,
			Code:                 code,
			// Placeholder invoice number; user can update in SI edit form.
			InvoiceNumber: "DRAFT-" + gr.Code,
			InvoiceDate:   today,
			DueDate:       today,
			Status:        models.SupplierInvoiceStatusDraft,
			CreatedBy:     actorID,
		}

		siItems := make([]models.SupplierInvoiceItem, 0, len(gr.Items))
		subTotal := 0.0
		for _, it := range gr.Items {
			price := priceMap[it.PurchaseOrderItemID]
			lineTotal := it.QuantityReceived * price
			subTotal += lineTotal

			productCodeSnap := ""
			productNameSnap := ""
			if it.Product != nil {
				productCodeSnap = it.Product.Code
				productNameSnap = it.Product.Name
			} else {
				productCodeSnap = it.ProductCodeSnapshot
				productNameSnap = it.ProductNameSnapshot
			}

			siItems = append(siItems, models.SupplierInvoiceItem{
				ProductID:           it.ProductID,
				PurchaseOrderItemID: &it.PurchaseOrderItemID,
				Quantity:            it.QuantityReceived,
				Price:               price,
				SubTotal:            lineTotal,
				ProductCodeSnapshot: productCodeSnap,
				ProductNameSnapshot: productNameSnap,
			})
		}
		si.SubTotal = subTotal
		si.Amount = subTotal
		si.RemainingAmount = subTotal

		// Auto-apply paid Down Payments by tracing GR → PO → SIDP
		var dpAmount float64
		var dpInvoiceID *string
		var dpInvoices []models.SupplierInvoice
		if err := tx.Where("purchase_order_id = ? AND type = ? AND deleted_at IS NULL",
			gr.PurchaseOrderID, models.SupplierInvoiceTypeDownPayment).
			Order("created_at DESC").
			Find(&dpInvoices).Error; err == nil && len(dpInvoices) > 0 {
			for _, dp := range dpInvoices {
				if dp.Status == models.SupplierInvoiceStatusPaid {
					dpAmount += dp.PaidAmount
				}
				if dpInvoiceID == nil {
					id := dp.ID
					dpInvoiceID = &id
				}
			}
		}
		if dpAmount > 0 || dpInvoiceID != nil {
			si.DownPaymentAmount = dpAmount
			si.DownPaymentInvoiceID = dpInvoiceID
			si.RemainingAmount = math.Max(0, subTotal-dpAmount)
		}

		si.Items = siItems

		if err := tx.Create(si).Error; err != nil {
			return fmt.Errorf("failed to create supplier invoice: %w", err)
		}
		siID = si.ID

		// Update GR with convert timestamp and link to SI
		now := apptime.Now()
		if err := tx.Model(&models.GoodsReceipt{}).Where("id = ?", id).Updates(map[string]interface{}{
			"converted_at":                     now,
			"converted_to_supplier_invoice_id": siID,
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "goods_receipt.convert", id, map[string]interface{}{
		"supplier_invoice_id": siID,
	})

	return &dto.GoodsReceiptConvertResponse{
		GoodsReceiptID:    id,
		SupplierInvoiceID: siID,
	}, nil
}

// getNextSupplierInvoiceCodeLockedInGR generates a unique SI code within a DB transaction.
func getNextSupplierInvoiceCodeLockedInGR(ctx context.Context, tx *gorm.DB, prefix string) (string, error) {
	now := database.GetDB(ctx, tx).NowFunc()
	dateStr := now.Format("20060102")
	codePrefix := prefix + "-" + dateStr + "-"

	lockKey := "supplier_invoice_code:" + dateStr
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
		return "", err
	}

	var last models.SupplierInvoice
	err := tx.WithContext(ctx).
		Unscoped().
		Model(&models.SupplierInvoice{}).
		Select("code").
		Where("code LIKE ?", codePrefix+"%").
		Order("code DESC").
		First(&last).Error

	seq := 1
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return "", err
		}
	} else if len(last.Code) >= len(codePrefix)+4 {
		lastSeqStr := last.Code[len(last.Code)-4:]
		var n int
		if _, convErr := fmt.Sscanf(strings.TrimSpace(lastSeqStr), "%d", &n); convErr == nil && n > 0 {
			seq = n + 1
		}
	}

	return fmt.Sprintf("%s%04d", codePrefix, seq), nil
}

func (uc *goodsReceiptUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.GoodsReceiptAuditTrailEntry, int64, error) {
	if uc.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.GoodsReceipt{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, 0, ErrGoodsReceiptNotFound
	}

	tx := uc.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("audit_logs.permission_code LIKE ?", "goods_receipt.%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type auditRow struct {
		ID             string    `gorm:"column:id"`
		ActorID        string    `gorm:"column:actor_id"`
		PermissionCode string    `gorm:"column:permission_code"`
		TargetID       string    `gorm:"column:target_id"`
		Action         string    `gorm:"column:action"`
		Metadata       string    `gorm:"column:metadata"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		ActorEmail     *string   `gorm:"column:actor_email"`
		ActorName      *string   `gorm:"column:actor_name"`
	}

	rows := make([]auditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	entries := make([]dto.GoodsReceiptAuditTrailEntry, 0, len(rows))
	refCache := make(map[string]string)
	for _, r := range rows {
		meta := parsePurchaseAuditMetadata(ctx, uc.db, r.Metadata, refCache)
		var usr *dto.AuditTrailUser
		if r.ActorID != "" {
			email := ""
			name := ""
			if r.ActorEmail != nil {
				email = *r.ActorEmail
			}
			if r.ActorName != nil {
				name = *r.ActorName
			}
			usr = &dto.AuditTrailUser{ID: r.ActorID, Email: email, Name: name}
		}
		entries = append(entries, dto.GoodsReceiptAuditTrailEntry{
			ID:             r.ID,
			Action:         r.Action,
			PermissionCode: r.PermissionCode,
			TargetID:       r.TargetID,
			Metadata:       meta,
			User:           usr,
			CreatedAt:      r.CreatedAt,
		})
	}

	return entries, total, nil
}

func grAuditSnapshot(gr *models.GoodsReceipt) map[string]interface{} {
	if gr == nil {
		return nil
	}
	items := make([]map[string]interface{}, 0, len(gr.Items))
	for _, it := range gr.Items {
		items = append(items, map[string]interface{}{
			"id":                     it.ID,
			"purchase_order_item_id": it.PurchaseOrderItemID,
			"product_id":             it.ProductID,
			"quantity_received":      it.QuantityReceived,
		})
	}
	return map[string]interface{}{
		"id":                gr.ID,
		"code":              gr.Code,
		"purchase_order_id": gr.PurchaseOrderID,
		"warehouse_id":      gr.WarehouseID,
		"supplier_id":       gr.SupplierID,
		"receipt_date":      gr.ReceiptDate,
		"notes":             gr.Notes,
		"proof_image_url":   gr.ProofImageURL,
		"status":            gr.Status,
		"created_by":        gr.CreatedBy,
		"items":             items,
	}
}

func validateActiveWarehouseID(ctx context.Context, db *gorm.DB, warehouseID string) error {
	warehouseID = strings.TrimSpace(warehouseID)
	if warehouseID == "" {
		return ErrGoodsReceiptInvalid
	}

	var count int64
	if err := database.GetDB(ctx, db).
		Model(&warehouseModels.Warehouse{}).
		Where("id = ?", warehouseID).
		Where("is_active = ?", true).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrGoodsReceiptInvalid
	}

	return nil
}

func (uc *goodsReceiptUsecase) validateGoodsReceiptItemsAgainstPO(ctx context.Context, gr *models.GoodsReceipt) error {
	if uc.db == nil {
		return errors.New("db is nil")
	}

	var po models.PurchaseOrder
	if err := database.GetDB(ctx, uc.db).
		Preload("Items").
		First(&po, "id = ?", gr.PurchaseOrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPurchaseOrderNotFound
		}
		return err
	}

	orderedByPOItemID := make(map[string]float64, len(po.Items))
	for _, poIt := range po.Items {
		orderedByPOItemID[poIt.ID] = poIt.Quantity
	}

	for _, it := range gr.Items {
		ordered, ok := orderedByPOItemID[it.PurchaseOrderItemID]
		if !ok {
			return ErrGoodsReceiptConflict
		}

		var alreadyReceived float64
		alreadyReceivedQuery := uc.db.WithContext(ctx).
			Table("goods_receipt_items")
		alreadyReceivedQuery, tenantErr := applyTenantJoinScope(ctx, alreadyReceivedQuery, "goods_receipt_items.tenant_id", "goods_receipts.tenant_id")
		if tenantErr != nil {
			return tenantErr
		}

		if err := alreadyReceivedQuery.
			Select("COALESCE(SUM(goods_receipt_items.quantity_received),0)").
			Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
			Where("goods_receipts.purchase_order_id = ?", po.ID).
			Where("UPPER(goods_receipts.status) <> ?", string(models.GoodsReceiptStatusRejected)).
			Where("goods_receipt_items.purchase_order_item_id = ?", it.PurchaseOrderItemID).
			Where("goods_receipts.id <> ?", gr.ID).
			Scan(&alreadyReceived).Error; err != nil {
			return err
		}

		receiving := math.Max(0, it.QuantityReceived)
		if alreadyReceived+receiving > ordered+0.0001 {
			return ErrGoodsReceiptConflict
		}
	}

	return nil
}

func getNextGoodsReceiptCodeLocked(ctx context.Context, tx *gorm.DB, prefix string) (string, error) {
	now := database.GetDB(ctx, tx).NowFunc()
	dateStr := now.Format("20060102")
	codePrefix := prefix + "-" + dateStr + "-"

	lockKey := "goods_receipt_code:" + dateStr
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
		return "", err
	}

	var last models.GoodsReceipt
	err := tx.WithContext(ctx).
		Unscoped().
		Model(&models.GoodsReceipt{}).
		Select("code").
		Where("code LIKE ?", codePrefix+"%").
		Order("code DESC").
		First(&last).Error

	seq := 1
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return "", err
		}
	} else if len(last.Code) >= len(codePrefix)+4 {
		lastSeqStr := last.Code[len(last.Code)-4:]
		var n int
		if _, convErr := fmt.Sscanf(strings.TrimSpace(lastSeqStr), "%d", &n); convErr == nil && n > 0 {
			seq = n + 1
		}
	}

	return fmt.Sprintf("%s%04d", codePrefix, seq), nil
}

func (uc *goodsReceiptUsecase) triggerAssetCreation(ctx context.Context, gr *models.GoodsReceipt) error {
	// 1. Get PO items to check product type
	var poItems []models.PurchaseOrderItem
	if err := database.GetDB(ctx, uc.db).Preload("Product.Type").Where("purchase_order_id = ?", gr.PurchaseOrderID).Find(&poItems).Error; err != nil {
		return err
	}

	for _, item := range gr.Items {
		var productPOItem *models.PurchaseOrderItem
		for i := range poItems {
			if poItems[i].ID == item.PurchaseOrderItemID {
				productPOItem = &poItems[i]
				break
			}
		}

		if productPOItem == nil || productPOItem.Product == nil || productPOItem.Product.Type == nil {
			continue
		}

		// Trigger only for "Device" type
		if strings.ToLower(productPOItem.Product.Type.Name) == "device" {
			acquisitionDate := apptime.Now().Format("2006-01-02")
			if gr.ReceiptDate != nil {
				acquisitionDate = gr.ReceiptDate.Format("2006-01-02")
			}

			req := &finDto.CreateAssetFromPurchaseRequest{
				Code:            fmt.Sprintf("AST-%s-%s", productPOItem.Product.Code, gr.Code),
				Name:            productPOItem.Product.Name,
				AcquisitionDate: acquisitionDate,
				AcquisitionCost: productPOItem.Price,
				ReferenceType:   "GOODS_RECEIPT",
				ReferenceID:     gr.ID,
			}
			if err := uc.assetUC.CreateFromPurchase(ctx, req); err != nil {
				return err
			}
		}
	}
	return nil
}
func (uc *goodsReceiptUsecase) TriggerJournalForReconciliation(ctx context.Context, gr *models.GoodsReceipt) error {
	return uc.inventoryUC.TriggerDocumentJournal(ctx, uc.db, reference.RefTypeGoodsReceipt, gr.ID)
}
