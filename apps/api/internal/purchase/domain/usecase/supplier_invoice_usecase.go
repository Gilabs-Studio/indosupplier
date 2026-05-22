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
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/mapper"
	purchaseService "github.com/gilabs/gims/api/internal/purchase/domain/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrSupplierInvoiceNotFound         = errors.New("supplier invoice not found")
	ErrSupplierInvoiceConflict         = errors.New("supplier invoice conflict")
	ErrSupplierInvoiceInvalid          = errors.New("invalid supplier invoice")
	ErrSupplierInvoiceQuantityExceeded = errors.New("supplier invoice quantity exceeds received quantity")
	ErrPaymentTermsNotFound            = errors.New("payment terms not found")
)

type SupplierInvoiceUsecase interface {
	AddData(ctx context.Context) (*dto.SupplierInvoiceAddResponse, error)
	List(ctx context.Context, params repositories.SupplierInvoiceListParams) ([]*dto.SupplierInvoiceListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateSupplierInvoiceRequest) (*dto.SupplierInvoiceDetailResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSupplierInvoiceRequest) (*dto.SupplierInvoiceDetailResponse, error)
	Delete(ctx context.Context, id string) error
	Submit(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Approve(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Reject(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Cancel(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Pending(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	Reverse(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error)
	ReverseWithReason(ctx context.Context, id string, reason string) (*dto.SupplierInvoiceDetailResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.SupplierInvoiceAuditTrailEntry, int64, error)
	TriggerJournalForSupplierInvoice(ctx context.Context, si *models.SupplierInvoice) error
	PreviewJournal(ctx context.Context, req *dto.CreateSupplierInvoiceRequest) (*dto.SupplierInvoiceJournalPreviewResponse, error)
}

type supplierInvoiceUsecase struct {
	db                 *gorm.DB
	repo               repositories.SupplierInvoiceRepository
	poRepo             repositories.PurchaseOrderRepository
	grRepo             repositories.GoodsReceiptRepository
	auditService       audit.AuditService
	mapper             *mapper.SupplierInvoiceMapper
	journalUC          finUsecase.JournalEntryUsecase
	coaUC              finUsecase.ChartOfAccountUsecase
	engine             accounting.AccountingEngine
	purchaseJournalSvc purchaseService.PurchaseJournalService
}

func NewSupplierInvoiceUsecase(
	db *gorm.DB,
	repo repositories.SupplierInvoiceRepository,
	poRepo repositories.PurchaseOrderRepository,
	grRepo repositories.GoodsReceiptRepository,
	auditService audit.AuditService,
	journalUC finUsecase.JournalEntryUsecase,
	coaUC finUsecase.ChartOfAccountUsecase,
	engine accounting.AccountingEngine,
	purchaseJournalSvc ...purchaseService.PurchaseJournalService,
) SupplierInvoiceUsecase {
	uc := &supplierInvoiceUsecase{
		db:                 db,
		repo:               repo,
		poRepo:             poRepo,
		grRepo:             grRepo,
		auditService:       auditService,
		mapper:             mapper.NewSupplierInvoiceMapper(),
		journalUC:          journalUC,
		coaUC:              coaUC,
		engine:             engine,
		purchaseJournalSvc: purchaseService.NewPurchaseJournalService(db, journalUC, engine),
	}
	if len(purchaseJournalSvc) > 0 && purchaseJournalSvc[0] != nil {
		uc.purchaseJournalSvc = purchaseJournalSvc[0]
	}

	return uc
}

func (uc *supplierInvoiceUsecase) List(ctx context.Context, params repositories.SupplierInvoiceListParams) ([]*dto.SupplierInvoiceListResponse, int64, error) {
	params.Type = string(models.SupplierInvoiceTypeNormal)
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return uc.mapper.ToListResponseList(items), total, nil
}

func (uc *supplierInvoiceUsecase) GetByID(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.SupplierInvoice{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, ErrSupplierInvoiceNotFound
	}

	si, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	if si.Type != models.SupplierInvoiceTypeNormal {
		return nil, ErrSupplierInvoiceNotFound
	}
	return uc.mapper.ToDetailResponse(si), nil
}

func (uc *supplierInvoiceUsecase) AddData(ctx context.Context) (*dto.SupplierInvoiceAddResponse, error) {
	// Payment terms
	var paymentTerms []coreModels.PaymentTerms
	if err := database.GetDB(ctx, uc.db).
		Model(&coreModels.PaymentTerms{}).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&paymentTerms).Error; err != nil {
		return nil, err
	}

	ptRes := make([]dto.SupplierInvoiceAddPaymentTerms, 0, len(paymentTerms))
	for _, pt := range paymentTerms {
		ptRes = append(ptRes, dto.SupplierInvoiceAddPaymentTerms{ID: pt.ID, Name: pt.Name})
	}

	// Fetch APPROVED/PARTIAL/CLOSED Goods Receipts as the source for creating Supplier Invoices
	var grs []*models.GoodsReceipt
	if err := database.GetDB(ctx, uc.db).
		Model(&models.GoodsReceipt{}).
		Where("status IN ?", []string{string(models.GoodsReceiptStatusApproved), string(models.GoodsReceiptStatusPartial), string(models.GoodsReceiptStatusClosed)}).
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.PaymentTerms").
		Preload("Supplier").
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.PurchaseOrderItem").
		Order("created_at DESC").
		Limit(100).
		Find(&grs).Error; err != nil {
		return nil, err
	}

	// Build PO item price map for all POs involved
	poIDs := make(map[string]bool)
	for _, gr := range grs {
		poIDs[gr.PurchaseOrderID] = true
	}
	var allPOItems []models.PurchaseOrderItem
	poIDList := make([]string, 0, len(poIDs))
	for id := range poIDs {
		poIDList = append(poIDList, id)
	}
	if len(poIDList) > 0 {
		database.GetDB(ctx, uc.db).Where("purchase_order_id IN ?", poIDList).Find(&allPOItems)
	}
	priceMap := make(map[string]float64)
	qtyMap := make(map[string]float64)
	for _, p := range allPOItems {
		priceMap[p.ID] = p.Price
		qtyMap[p.ID] = p.Quantity
	}

	grIDs := make([]string, 0, len(grs))
	for _, gr := range grs {
		grIDs = append(grIDs, gr.ID)
	}
	type grQtySum struct {
		GoodsReceiptID string
		ProductID      string
		Qty            float64
	}
	var invoicedSums []grQtySum
	if len(grIDs) > 0 {
		invoicedQuery := uc.db.WithContext(ctx).
			Table("supplier_invoice_items")
		invoicedQuery, tenantErr := applyTenantJoinScope(ctx, invoicedQuery, "supplier_invoice_items.tenant_id", "supplier_invoices.tenant_id")
		if tenantErr != nil {
			return nil, tenantErr
		}

		if err := invoicedQuery.
			Select("supplier_invoices.goods_receipt_id, supplier_invoice_items.product_id, SUM(supplier_invoice_items.quantity) as qty").
			Joins("JOIN supplier_invoices ON supplier_invoices.id = supplier_invoice_items.supplier_invoice_id").
			Where("supplier_invoices.goods_receipt_id IN ? AND supplier_invoices.status NOT IN ?", grIDs, []string{
				string(models.SupplierInvoiceStatusRejected),
				string(models.SupplierInvoiceStatusCancelled),
			}).
			Where("supplier_invoices.deleted_at IS NULL").
			Group("supplier_invoices.goods_receipt_id, supplier_invoice_items.product_id").
			Scan(&invoicedSums).Error; err != nil {
			return nil, err
		}
	}
	grInvoicedMap := make(map[string]map[string]float64)
	for _, s := range invoicedSums {
		if grInvoicedMap[s.GoodsReceiptID] == nil {
			grInvoicedMap[s.GoodsReceiptID] = make(map[string]float64)
		}
		grInvoicedMap[s.GoodsReceiptID][s.ProductID] = s.Qty
	}

	grRes := make([]dto.SupplierInvoiceAddGoodsReceipt, 0, len(grs))
	for _, gr := range grs {
		items := make([]dto.SupplierInvoiceAddGoodsReceiptItem, 0, len(gr.Items))
		for _, it := range gr.Items {
			var prod *dto.SupplierInvoiceAddProductMini
			if it.Product != nil {
				prod = &dto.SupplierInvoiceAddProductMini{ID: it.Product.ID, Name: it.Product.Name, Code: it.Product.Code, ImageURL: it.Product.ImageURL}
			}
			price := priceMap[it.PurchaseOrderItemID]
			qtyInv := 0.0
			if grInvoicedMap[gr.ID] != nil {
				qtyInv = grInvoicedMap[gr.ID][it.ProductID]
			}
			qtyRem := it.QuantityReceived - qtyInv
			if qtyRem < 0 {
				qtyRem = 0
			}
			items = append(items, dto.SupplierInvoiceAddGoodsReceiptItem{
				ID:                  it.ID,
				PurchaseOrderItemID: it.PurchaseOrderItemID,
				Product:             prod,
				QuantityPO:          qtyMap[it.PurchaseOrderItemID],
				QuantityReceived:    it.QuantityReceived,
				QuantityInvoiced:    qtyInv,
				QuantityRemaining:   qtyRem,
				Price:               price,
				SubTotal:            qtyRem * price,
			})
		}

		var sup *dto.SupplierInvoiceAddSupplierMini
		if gr.Supplier != nil {
			sup = &dto.SupplierInvoiceAddSupplierMini{ID: gr.Supplier.ID, Name: gr.Supplier.Name}
		}

		var poMini *dto.SupplierInvoicePurchaseOrderMini
		if gr.PurchaseOrder != nil {
			poMini = &dto.SupplierInvoicePurchaseOrderMini{ID: gr.PurchaseOrder.ID, Code: gr.PurchaseOrder.Code}
		}

		var defaultPTID *string
		var defaultPTName *string
		if gr.PurchaseOrder != nil && gr.PurchaseOrder.PaymentTermsID != nil {
			defaultPTID = gr.PurchaseOrder.PaymentTermsID
			if gr.PurchaseOrder.PaymentTerms != nil {
				name := gr.PurchaseOrder.PaymentTerms.Name
				defaultPTName = &name
			} else if gr.PurchaseOrder.PaymentTermsNameSnapshot != "" {
				name := gr.PurchaseOrder.PaymentTermsNameSnapshot
				defaultPTName = &name
			}
		}

		addGR := dto.SupplierInvoiceAddGoodsReceipt{
			ID:                      gr.ID,
			Code:                    gr.Code,
			PurchaseOrder:           poMini,
			Supplier:                sup,
			ReceiptDate:             gr.ReceiptDate,
			Status:                  string(gr.Status),
			Items:                   items,
			DefaultPaymentTermsID:   defaultPTID,
			DefaultPaymentTermsName: defaultPTName,
		}

		// Attach latest DP invoice if exists for the PO behind this GR
		if dp, err := uc.repo.GetLatestDownPaymentByPO(ctx, gr.PurchaseOrderID); err == nil && dp != nil {
			addGR.InvoiceDP = &dto.SupplierInvoiceAddDownPaymentMini{
				ID:            dp.ID,
				PurchaseOrder: poMini,
				Code:          dp.Code,
				InvoiceNumber: dp.InvoiceNumber,
				InvoiceDate:   dp.InvoiceDate.Format("2006-01-02"),
				DueDate:       dp.DueDate.Format("2006-01-02"),
				Amount:        dp.Amount,
				PaidAmount:    dp.PaidAmount,
				Status:        string(dp.Status),
				Notes:         dp.Notes,
				CreatedAt:     dp.CreatedAt,
				UpdatedAt:     dp.UpdatedAt,
			}
		}

		grRes = append(grRes, addGR)
	}

	return &dto.SupplierInvoiceAddResponse{PaymentTerms: ptRes, GoodsReceipts: grRes}, nil
}

func (uc *supplierInvoiceUsecase) Create(ctx context.Context, req *dto.CreateSupplierInvoiceRequest) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var createdID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Fetch the Goods Receipt (source) with row-level lock
		var gr models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			Preload("Items.Product").
			First(&gr, "id = ?", req.GoodsReceiptID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrGoodsReceiptNotFound
			}
			return err
		}
		if gr.Status != models.GoodsReceiptStatusApproved && gr.Status != models.GoodsReceiptStatusPartial && gr.Status != models.GoodsReceiptStatusClosed {
			return ErrInvalidStatus
		}

		// Derive PO from GR and lock it
		var po models.PurchaseOrder
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			Preload("Items.Product").
			First(&po, "id = ?", gr.PurchaseOrderID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPurchaseOrderNotFound
			}
			return err
		}
		if po.SupplierID == nil || strings.TrimSpace(*po.SupplierID) == "" {
			return ErrSupplierInvoiceInvalid
		}

		var pt coreModels.PaymentTerms
		if err := tx.First(&pt, "id = ?", req.PaymentTermsID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPaymentTermsNotFound
			}
			return err
		}

		code, err := getNextSupplierInvoiceCodeLocked(tx, "SI")
		if err != nil {
			return err
		}

		items := make([]models.SupplierInvoiceItem, 0, len(req.Items))
		subTotal := 0.0

		poItemIDByProduct, err := uc.validateSIQuantity(ctx, tx, &gr, req.Items, "")
		if err != nil {
			return err
		}

		for _, it := range req.Items {
			disc := math.Max(0, math.Min(100, it.Discount))
			// Round each line item to 2 decimal places
			line := round2dp(it.Quantity * it.Price * (1 - disc/100))
			items = append(items, models.SupplierInvoiceItem{
				PurchaseOrderItemID: func() *string {
					id := poItemIDByProduct[it.ProductID]
					if strings.TrimSpace(id) == "" {
						return nil
					}
					return &id
				}(),
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
				Price:     it.Price,
				Discount:  disc,
				SubTotal:  line,
			})
			subTotal += line
		}

		taxRate := math.Max(0, math.Min(100, req.TaxRate))
		tax := round2dp(subTotal * taxRate / 100)
		// Calculate Gross Amount (before DP deduction)
		grossAmount := round2dp(subTotal + tax + math.Max(0, req.DeliveryCost) + math.Max(0, req.OtherCost))

		// Auto-apply paid Down Payments by tracing GR → PO → SIDP
		var dpAmount float64
		var dpInvoiceID *string
		var dpInvoices []models.SupplierInvoice
		if err := tx.Where("purchase_order_id = ? AND type = ? AND deleted_at IS NULL",
			po.ID, models.SupplierInvoiceTypeDownPayment).
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
		remainingAmount := math.Max(0, grossAmount-dpAmount)

		creatorID, _ := ctx.Value("user_id").(string)

		invoiceDateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.InvoiceDate))
		if err != nil {
			return errors.New("invalid invoice date format")
		}

		dueDateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate))
		if err != nil {
			return errors.New("invalid due date format")
		}

		grID := gr.ID
		si := models.SupplierInvoice{
			Type:                 models.SupplierInvoiceTypeNormal,
			PurchaseOrderID:      po.ID,
			GoodsReceiptID:       &grID,
			SupplierID:           *po.SupplierID,
			PaymentTermsID:       &pt.ID,
			Code:                 code,
			InvoiceNumber:        req.InvoiceNumber,
			InvoiceDate:          invoiceDateParsed,
			DueDate:              dueDateParsed,
			TaxRate:              req.TaxRate,
			TaxAmount:            tax,
			DeliveryCost:         req.DeliveryCost,
			OtherCost:            req.OtherCost,
			SubTotal:             subTotal,
			DownPaymentAmount:    dpAmount,
			Amount:               grossAmount,
			RemainingAmount:      remainingAmount,
			DownPaymentInvoiceID: dpInvoiceID,
			Status:               models.SupplierInvoiceStatusDraft,
			Notes:                req.Notes,
			CreatedBy:            creatorID,
			Items:                items,
		}
		si.CompanyID = po.CompanyID
		si.FiscalYearID = po.FiscalYearID
		if si.CompanyID == nil || si.FiscalYearID == nil {
			companyID, fiscalYearID, err := newPurchaseFiscalYearResolver(uc.db, financeRepositories.NewFiscalYearRepository(uc.db)).Resolve(ctx, req.InvoiceDate)
			if err != nil {
				return err
			}
			si.CompanyID = &companyID
			si.FiscalYearID = &fiscalYearID
		}

		if err := snapshotSupplierInvoice(ctx, tx, &si, nil); err != nil {
			return err
		}

		if err := tx.Create(&si).Error; err != nil {
			return err
		}

		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}
		createdID = si.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	out, err := uc.repo.GetByID(ctx, createdID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "supplier_invoice.create", out.ID, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) Update(ctx context.Context, id string, req *dto.UpdateSupplierInvoiceRequest) (*dto.SupplierInvoiceDetailResponse, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	if existing.Type != models.SupplierInvoiceTypeNormal {
		return nil, ErrSupplierInvoiceNotFound
	}
	if existing.Status != models.SupplierInvoiceStatusDraft {
		return nil, ErrSupplierInvoiceConflict
	}

	// Reuse create computation/validation; keep PO locked for safety.
	createReq := dto.CreateSupplierInvoiceRequest{
		GoodsReceiptID: req.GoodsReceiptID,
		PaymentTermsID: req.PaymentTermsID,
		InvoiceNumber:  req.InvoiceNumber,
		InvoiceDate:    req.InvoiceDate,
		DueDate:        req.DueDate,
		TaxRate:        req.TaxRate,
		DeliveryCost:   req.DeliveryCost,
		OtherCost:      req.OtherCost,
		Notes:          req.Notes,
		Items:          req.Items,
	}

	before := *existing
	res, err := uc.replaceDraft(ctx, id, &createReq)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "supplier_invoice.update", id, map[string]interface{}{"before": before, "after": res})
	return res, nil
}

func (uc *supplierInvoiceUsecase) replaceDraft(ctx context.Context, id string, req *dto.CreateSupplierInvoiceRequest) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var updatedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Status != models.SupplierInvoiceStatusDraft || si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceConflict
		}

		// Fetch the Goods Receipt (source) with row-level lock
		var gr models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&gr, "id = ?", req.GoodsReceiptID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrGoodsReceiptNotFound
			}
			return err
		}
		if gr.Status != models.GoodsReceiptStatusApproved && gr.Status != models.GoodsReceiptStatusPartial && gr.Status != models.GoodsReceiptStatusClosed {
			return ErrInvalidStatus
		}

		// Derive PO from GR
		var po models.PurchaseOrder
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&po, "id = ?", gr.PurchaseOrderID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPurchaseOrderNotFound
			}
			return err
		}
		if po.SupplierID == nil || strings.TrimSpace(*po.SupplierID) == "" {
			return ErrSupplierInvoiceInvalid
		}

		var pt coreModels.PaymentTerms
		if err := tx.First(&pt, "id = ?", req.PaymentTermsID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPaymentTermsNotFound
			}
			return err
		}

		poItemIDByProduct, err := uc.validateSIQuantity(ctx, tx, &gr, req.Items, id)
		if err != nil {
			return err
		}

		newItems := make([]models.SupplierInvoiceItem, 0, len(req.Items))
		subTotal := 0.0
		for _, it := range req.Items {
			disc := math.Max(0, math.Min(100, it.Discount))
			line := it.Quantity * it.Price * (1 - disc/100)
			poItemID := poItemIDByProduct[it.ProductID]
			var poItemPtr *string
			if strings.TrimSpace(poItemID) != "" {
				poItemPtr = &poItemID
			}

			newItems = append(newItems, models.SupplierInvoiceItem{PurchaseOrderItemID: poItemPtr, ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price, Discount: disc, SubTotal: line})
			subTotal += line
		}

		taxRate := math.Max(0, math.Min(100, req.TaxRate))
		tax := round2dp(subTotal * taxRate / 100)
		grossAmount := round2dp(subTotal + tax + math.Max(0, req.DeliveryCost) + math.Max(0, req.OtherCost))

		// Auto-apply paid Down Payments by tracing GR → PO → SIDP
		var dpAmount float64
		var dpInvoiceID *string
		var dpInvoices []models.SupplierInvoice
		if err := tx.Where("purchase_order_id = ? AND type = ? AND status = ? AND deleted_at IS NULL",
			po.ID, models.SupplierInvoiceTypeDownPayment, models.SupplierInvoiceStatusPaid).
			Find(&dpInvoices).Error; err == nil && len(dpInvoices) > 0 {
			for _, dp := range dpInvoices {
				dpAmount += dp.PaidAmount
				if dpInvoiceID == nil {
					id := dp.ID
					dpInvoiceID = &id
				}
			}
		}

		remainingAmount := math.Max(0, grossAmount-dpAmount)
		grID := gr.ID

		invoiceDateUpdateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.InvoiceDate))
		if err != nil {
			return errors.New("invalid invoice date format")
		}

		dueDateUpdateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate))
		if err != nil {
			return errors.New("invalid due date format")
		}

		updatedDraft := &models.SupplierInvoice{
			ID:                si.ID,
			Type:              si.Type,
			PurchaseOrderID:   po.ID,
			GoodsReceiptID:    &grID,
			SupplierID:        *po.SupplierID,
			PaymentTermsID:    &pt.ID,
			Code:              si.Code,
			InvoiceNumber:     req.InvoiceNumber,
			InvoiceDate:       invoiceDateUpdateParsed,
			DueDate:           dueDateUpdateParsed,
			TaxRate:           req.TaxRate,
			TaxAmount:         tax,
			DeliveryCost:      req.DeliveryCost,
			OtherCost:         req.OtherCost,
			SubTotal:          subTotal,
			DownPaymentAmount: dpAmount,
			Amount:            grossAmount,
			RemainingAmount:   math.Max(0, remainingAmount-si.PaidAmount),
			Status:            si.Status,
			Notes:             req.Notes,
			CreatedBy:         si.CreatedBy,
			Items:             newItems,
		}
		if err := snapshotSupplierInvoice(ctx, tx, updatedDraft, &si); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"purchase_order_id":           po.ID,
			"goods_receipt_id":            grID,
			"supplier_id":                 *po.SupplierID,
			"payment_terms_id":            pt.ID,
			"supplier_code_snapshot":      updatedDraft.SupplierCodeSnapshot,
			"supplier_name_snapshot":      updatedDraft.SupplierNameSnapshot,
			"payment_terms_name_snapshot": updatedDraft.PaymentTermsNameSnapshot,
			"payment_terms_days_snapshot": func() interface{} {
				if updatedDraft.PaymentTermsDaysSnapshot == nil {
					return nil
				}
				return *updatedDraft.PaymentTermsDaysSnapshot
			}(),
			"invoice_number":          req.InvoiceNumber,
			"invoice_date":            req.InvoiceDate,
			"due_date":                req.DueDate,
			"tax_rate":                req.TaxRate,
			"tax_amount":              tax,
			"delivery_cost":           req.DeliveryCost,
			"other_cost":              req.OtherCost,
			"sub_total":               subTotal,
			"down_payment_amount":     dpAmount,
			"amount":                  grossAmount,
			"remaining_amount":        math.Max(0, remainingAmount-si.PaidAmount),
			"down_payment_invoice_id": dpInvoiceID,
			"notes":                   req.Notes,
			"updated_at":              apptime.Now(),
		}
		if err := tx.Model(&si).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Where("supplier_invoice_id = ?", si.ID).Delete(&models.SupplierInvoiceItem{}).Error; err != nil {
			return err
		}
		for i := range newItems {
			newItems[i].SupplierInvoiceID = si.ID
			newItems[i].ProductCodeSnapshot = updatedDraft.Items[i].ProductCodeSnapshot
			newItems[i].ProductNameSnapshot = updatedDraft.Items[i].ProductNameSnapshot
		}
		if err := tx.Create(&newItems).Error; err != nil {
			return err
		}
		oldGRID := ""
		if si.GoodsReceiptID != nil {
			oldGRID = *si.GoodsReceiptID
		}

		if oldGRID != "" && oldGRID != gr.ID {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, oldGRID); err != nil {
				return err
			}
		}

		if err := uc.syncGoodsReceiptStatus(ctx, tx, gr.ID); err != nil {
			return err
		}

		updatedID = si.ID
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}

	loaded, err := uc.repo.GetByID(ctx, updatedID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	return uc.mapper.ToDetailResponse(loaded), nil
}

func (uc *supplierInvoiceUsecase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrSupplierInvoiceNotFound
		}
		return err
	}
	if existing.Type != models.SupplierInvoiceTypeNormal {
		return ErrSupplierInvoiceNotFound
	}
	// Allow deletion of draft invoices only (workflow: use Cancel for others)
	if existing.Status != models.SupplierInvoiceStatusDraft {
		return ErrSupplierInvoiceConflict
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditService.Log(ctx, "supplier_invoice.delete", id, map[string]interface{}{"before": existing})
	return nil
}

func (uc *supplierInvoiceUsecase) Submit(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		// Allow submitting from DRAFT state (DRAFT -> SUBMITTED)
		if si.Status != models.SupplierInvoiceStatusDraft {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		if err := tx.Model(&si).Updates(map[string]interface{}{
			"status":       models.SupplierInvoiceStatusSubmitted,
			"submitted_at": &now,
		}).Error; err != nil {
			return err
		}
		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	out, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice.submit", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "supplier_invoice.approve",
		EntityType:     "supplier_invoice",
		EntityID:       out.ID,
		Title:          "Supplier Invoice Approval",
		Message:        "A supplier invoice has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create supplier invoice notification: %v", err)
	}
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) Approve(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	var approvedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		if si.Status != models.SupplierInvoiceStatusSubmitted {
			return ErrSupplierInvoiceConflict
		}

		if err := uc.transitionSupplierInvoiceToPayable(ctx, tx, &si, true); err != nil {
			return err
		}

		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}

		approvedID = si.ID
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	out, err := uc.repo.GetByID(ctx, approvedID)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice.approve", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) transitionSupplierInvoiceToPayable(ctx context.Context, tx *gorm.DB, si *models.SupplierInvoice, setApprovedAt bool) error {
	// --- Three-Way Matching Validation ---
	// 1. Get total Qty Received for this PO
	type qtySum struct {
		ProductID string
		Qty       float64
	}

	var receivedSums []qtySum
	receivedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
		Table("goods_receipt_items")
	receivedQuery, err := applyTenantJoinScope(ctx, receivedQuery, "goods_receipt_items.tenant_id", "goods_receipts.tenant_id")
	if err != nil {
		return err
	}

	if err := receivedQuery.
		Select("product_id, SUM(quantity_received) as qty").
		Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
		Where("goods_receipts.purchase_order_id = ? AND goods_receipts.status IN ?", si.PurchaseOrderID, []string{
			string(models.GoodsReceiptStatusApproved),
			string(models.GoodsReceiptStatusPartial),
			string(models.GoodsReceiptStatusClosed),
		}).
		Group("product_id").Scan(&receivedSums).Error; err != nil {
		return err
	}

	receivedMap := make(map[string]float64)
	for _, r := range receivedSums {
		receivedMap[r.ProductID] = r.Qty
	}

	// 2. Get total Qty already Invoiced for this PO (excluding current one, draft, submitted, rejected, cancelled)
	var invoicedSums []qtySum
	invoicedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
		Table("supplier_invoice_items")
	invoicedQuery, err = applyTenantJoinScope(ctx, invoicedQuery, "supplier_invoice_items.tenant_id", "supplier_invoices.tenant_id")
	if err != nil {
		return err
	}

	if err := invoicedQuery.
		Select("product_id, SUM(quantity) as qty").
		Joins("JOIN supplier_invoices ON supplier_invoices.id = supplier_invoice_items.supplier_invoice_id").
		Where("supplier_invoices.purchase_order_id = ? AND supplier_invoices.status NOT IN ? AND supplier_invoices.id != ?",
			si.PurchaseOrderID, []models.SupplierInvoiceStatus{
				models.SupplierInvoiceStatusDraft,
				models.SupplierInvoiceStatusSubmitted,
				models.SupplierInvoiceStatusRejected,
				models.SupplierInvoiceStatusCancelled,
			}, si.ID).
		Group("product_id").Scan(&invoicedSums).Error; err != nil {
		return err
	}

	invoicedMap := make(map[string]float64)
	for _, i := range invoicedSums {
		invoicedMap[i.ProductID] = i.Qty
	}

	// 3. Compare: Current Invoicing Qty + Already Invoiced <= Total Received
	for _, it := range si.Items {
		received := receivedMap[it.ProductID]
		alreadyInvoiced := invoicedMap[it.ProductID]
		if it.Quantity+alreadyInvoiced > received+0.0001 {
			return fmt.Errorf("invoiced quantity for product %s exceeds received quantity (Received: %.2f, Invoiced: %.2f)", it.ProductID, received, alreadyInvoiced)
		}
	}

	var poItems []models.PurchaseOrderItem
	if err := tx.Where("purchase_order_id = ?", si.PurchaseOrderID).Find(&poItems).Error; err != nil {
		return err
	}
	poPriceByProduct := make(map[string]float64, len(poItems))
	for _, it := range poItems {
		poPriceByProduct[it.ProductID] = poNetUnitPrice(it)
	}
	si.GRIRVarianceTotal = calculateGRIRVarianceFromInvoiceItems(si.Items, poPriceByProduct)

	// --- Budget Guard ---
	extraCost := si.DeliveryCost + si.OtherCost
	if extraCost > 0 {
		// si.InvoiceDate is already time.Time, no need to parse
		expenseCOAID := ""
		if uc.engine != nil {
			if mappedCOAID, resolveErr := uc.engine.ResolveCOAID(ctx, "coa.purchase_expense"); resolveErr == nil {
				expenseCOAID = strings.TrimSpace(mappedCOAID)
			} else {
				// Fail loudly if mapping not configured: budget guard requires explicit COA mapping
				return fmt.Errorf("purchase_expense COA mapping not configured: %w. please configure coa.purchase_expense in system account mappings", resolveErr)
			}
		}

		if expenseCOAID != "" {
			if err := finUsecase.EnsureWithinBudget(ctx, tx, expenseCOAID, si.InvoiceDate, extraCost); err != nil {
				// Do not block supplier invoice approval when the mapped extra-cost account
				// is not budgeted for the period. Budget overrun remains enforced.
				if !finUsecase.IsBudgetAccountNotConfiguredForPeriod(err) {
					return err
				}
				log.Printf("supplier_invoice budget guard skipped: unbudgeted account for period, invoice_id=%s coa_id=%s", si.ID, expenseCOAID)
			}
		}
	}

	// Recalculate DP deduction before approving and ensure link is made
	var dpAmount float64
	var dpInvoiceID *string
	var dpInvoices []models.SupplierInvoice
	if err := tx.Where("purchase_order_id = ? AND type = ? AND deleted_at IS NULL",
		si.PurchaseOrderID, models.SupplierInvoiceTypeDownPayment).
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

	// Keep approval flow consistent with Customer Invoice flow:
	// SUBMITTED -> APPROVED -> UNPAID -> payment handling.
	// IMPORTANT: Only set to PAID if DP fully covers the amount.
	// DO NOT auto-close - SI must wait for payment confirmation before closing.
	status := models.SupplierInvoiceStatusUnpaid
	remainingAmount := math.Max(0, si.Amount-dpAmount)
	if remainingAmount <= 0.0001 {
		status = models.SupplierInvoiceStatusPaid
	}

	updates := map[string]interface{}{
		"status":                  status,
		"down_payment_amount":     dpAmount,
		"remaining_amount":        remainingAmount,
		"down_payment_invoice_id": dpInvoiceID,
	}
	if setApprovedAt {
		now := apptime.Now()
		updates["approved_at"] = &now
	}
	if status == models.SupplierInvoiceStatusPaid {
		now := apptime.Now()
		updates["payment_at"] = &now
	}

	if err := tx.Model(si).Updates(updates).Error; err != nil {
		return err
	}

	// Trigger Journal Entry (AP Recognition)
	// We do this INSIDE the transaction for atomicity.
	// Passing tx via context.
	txCtx := database.WithTx(ctx, tx)
	if err := uc.triggerJournalEntry(txCtx, si); err != nil {
		return err
	}

	return nil
}

func (uc *supplierInvoiceUsecase) Reject(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		if si.Status != models.SupplierInvoiceStatusSubmitted {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		if err := tx.Model(&si).Updates(map[string]interface{}{
			"status":      models.SupplierInvoiceStatusRejected,
			"rejected_at": &now,
		}).Error; err != nil {
			return err
		}
		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	out, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice.reject", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) Cancel(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		allowed := si.Status == models.SupplierInvoiceStatusDraft ||
			si.Status == models.SupplierInvoiceStatusSubmitted ||
			si.Status == models.SupplierInvoiceStatusApproved
		if !allowed {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		if err := tx.Model(&si).Updates(map[string]interface{}{
			"status":       models.SupplierInvoiceStatusCancelled,
			"cancelled_at": &now,
		}).Error; err != nil {
			return err
		}
		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	out, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice.cancel", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) Pending(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var pendingID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}
		if si.Status != models.SupplierInvoiceStatusApproved {
			return ErrSupplierInvoiceConflict
		}

		if err := uc.transitionSupplierInvoiceToPayable(ctx, tx, &si, false); err != nil {
			return err
		}

		pendingID = si.ID
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}

	out, err := uc.repo.GetByID(ctx, pendingID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}

	uc.auditService.Log(ctx, "supplier_invoice.pending", id, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) triggerJournalEntry(ctx context.Context, si *models.SupplierInvoice) error {
	if si == nil || uc.journalUC == nil || uc.engine == nil {
		return nil
	}

	companyID, fiscalYearID, err := resolvePurchaseJournalScope(ctx, uc.db, si.CompanyID, si.FiscalYearID, si.InvoiceDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("failed to resolve supplier invoice journal scope: %w", err)
	}

	data := accounting.TransactionData{
		ReferenceType:     reference.RefTypeSupplierInvoice,
		ReferenceID:       si.ID,
		CompanyID:         companyID,
		FiscalYearID:      fiscalYearID,
		EntryDate:         si.InvoiceDate.Format("2006-01-02"),
		Description:       fmt.Sprintf("Purchase Invoice %s (%s)", si.InvoiceNumber, si.Code),
		TotalAmount:       si.Amount,
		SubTotal:          si.SubTotal,
		TaxTotal:          si.TaxAmount,
		DepositTotal:      si.DownPaymentAmount,
		OtherTotal:        si.DeliveryCost + si.OtherCost,
		GRIRVarianceTotal: si.GRIRVarianceTotal,
		DescriptionArgs:   []interface{}{si.InvoiceNumber, si.SupplierNameSnapshot},
	}

	if uc.purchaseJournalSvc != nil {
		postedJournal, err := uc.purchaseJournalSvc.GeneratePurchaseJournal(ctx, purchaseService.PurchaseJournalTxn{
			Profile: accounting.ProfileSupplierInvoice,
			Data:    data,
		})
		if err != nil {
			return fmt.Errorf("failed to post supplier invoice journal: %w", err)
		}

		if postedJournal != nil && strings.TrimSpace(postedJournal.ID) != "" {
			if err := database.GetDB(ctx, uc.db).
				Model(&models.SupplierInvoice{}).
				Where("id = ?", si.ID).
				Updates(map[string]interface{}{
					"journal_entry_id": postedJournal.ID,
					"is_posted":        true,
				}).Error; err != nil {
				return fmt.Errorf("failed to link supplier invoice journal: %w", err)
			}
			si.JournalEntryID = &postedJournal.ID
			si.IsPosted = true
		}

		log.Printf("journal_observability event=trigger.success reference_type=%s reference_id=%s", reference.RefTypeSupplierInvoice, si.ID)
		return nil
	}

	req, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSupplierInvoice, data)
	if err != nil {
		return fmt.Errorf("failed to generate supplier invoice journal: %w", err)
	}

	// Double-check balance
	var debitTotal, creditTotal float64
	for _, l := range req.Lines {
		debitTotal += l.Debit
		creditTotal += l.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.001 {
		return fmt.Errorf("generated supplier invoice journal is unbalanced: debit=%.2f credit=%.2f", debitTotal, creditTotal)
	}

	postedJournal, err := uc.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post supplier invoice journal: %w", err)
	}

	if postedJournal != nil && strings.TrimSpace(postedJournal.ID) != "" {
		if err := database.GetDB(ctx, uc.db).
			Model(&models.SupplierInvoice{}).
			Where("id = ?", si.ID).
			Updates(map[string]interface{}{
				"journal_entry_id": postedJournal.ID,
				"is_posted":        true,
			}).Error; err != nil {
			return fmt.Errorf("failed to link supplier invoice journal: %w", err)
		}
		si.JournalEntryID = &postedJournal.ID
		si.IsPosted = true
	}

	log.Printf("journal_observability event=trigger.success reference_type=%s reference_id=%s", reference.RefTypeSupplierInvoice, si.ID)
	return nil
}

func (uc *supplierInvoiceUsecase) ReverseWithReason(ctx context.Context, id string, reason string) (*dto.SupplierInvoiceDetailResponse, error) {
	return uc.reverse(ctx, id, reason)
}

func (uc *supplierInvoiceUsecase) Reverse(ctx context.Context, id string) (*dto.SupplierInvoiceDetailResponse, error) {
	return uc.reverse(ctx, id, "Manual reversal")
}

func (uc *supplierInvoiceUsecase) reverse(ctx context.Context, id string, reason string) (*dto.SupplierInvoiceDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var revertedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}

		if si.Type != models.SupplierInvoiceTypeNormal {
			return ErrSupplierInvoiceNotFound
		}

		// Only Unpaid, Partial, or Paid (recognized) invoices can be reversed.
		// Pending/WaitingPayment too.
		isRecognized := si.Status == models.SupplierInvoiceStatusUnpaid ||
			si.Status == models.SupplierInvoiceStatusPartial ||
			si.Status == models.SupplierInvoiceStatusPaid ||
			si.Status == models.SupplierInvoiceStatusWaitingPayment

		if !isRecognized {
			return fmt.Errorf("only recognized invoices can be reversed (current status: %s)", si.Status)
		}

		now := apptime.Now()
		if err := tx.Model(&si).Updates(map[string]interface{}{
			"status":       models.SupplierInvoiceStatusReversed,
			"cancelled_at": &now, // Keep CancelledAt for generic time tracking or we could add ReversedAt
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		txCtx := database.WithTx(ctx, tx)
		if uc.purchaseJournalSvc != nil {
			if err := uc.purchaseJournalSvc.ReversePurchaseJournal(txCtx, reference.RefTypeSupplierInvoice, si.ID, reason); err != nil {
				return fmt.Errorf("failed to reverse journal: %w", err)
			}
		} else {
			// Find the journal for this invoice
			var journal financeModels.JournalEntry
			refType := reference.RefTypeSupplierInvoice
			err := tx.Where("reference_type = ? AND reference_id = ?", refType, si.ID).
				Where("status = ?", financeModels.JournalStatusPosted).
				First(&journal).Error

			if err == nil {
				// Trigger journal reversal
				// Exempt from closing guard if we are doing a reversal
				revCtx := finUsecase.WithReversalFlag(txCtx)

				if _, err := uc.journalUC.ReverseWithReason(revCtx, journal.ID, reason); err != nil {
					return fmt.Errorf("failed to reverse journal: %w", err)
				}
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
		}

		// Sync Goods Receipt status (to release the quantity for new invoicing)
		if si.GoodsReceiptID != nil {
			if err := uc.syncGoodsReceiptStatus(ctx, tx, *si.GoodsReceiptID); err != nil {
				return err
			}
		}

		revertedID = si.ID
		return nil
	})

	if err != nil {
		return nil, err
	}

	out, _ := uc.repo.GetByID(ctx, revertedID)
	uc.auditService.Log(ctx, "supplier_invoice.reverse", id, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *supplierInvoiceUsecase) TriggerJournalForSupplierInvoice(ctx context.Context, si *models.SupplierInvoice) error {
	return uc.triggerJournalEntry(ctx, si)
}

func (uc *supplierInvoiceUsecase) PreviewJournal(ctx context.Context, req *dto.CreateSupplierInvoiceRequest) (*dto.SupplierInvoiceJournalPreviewResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	if uc.engine == nil {
		return nil, errors.New("accounting engine is not configured")
	}

	var preview *dto.SupplierInvoiceJournalPreviewResponse
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var gr models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			Preload("Items.Product").
			First(&gr, "id = ?", req.GoodsReceiptID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrGoodsReceiptNotFound
			}
			return err
		}

		if gr.Status != models.GoodsReceiptStatusApproved && gr.Status != models.GoodsReceiptStatusPartial && gr.Status != models.GoodsReceiptStatusClosed {
			return ErrInvalidStatus
		}

		if _, err := uc.validateSIQuantity(ctx, tx, &gr, req.Items, ""); err != nil {
			return err
		}

		// Build PO item references for matching and valuation basis.
		var poItems []models.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", gr.PurchaseOrderID).Find(&poItems).Error; err != nil {
			return err
		}

		poQtyByProduct := make(map[string]float64, len(poItems))
		poPriceByProduct := make(map[string]float64, len(poItems))
		for _, it := range poItems {
			poQtyByProduct[it.ProductID] = it.Quantity
			poPriceByProduct[it.ProductID] = poNetUnitPrice(it)
		}

		type qtySum struct {
			ProductID string
			Qty       float64
		}

		var receivedSums []qtySum
		receivedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
			Table("goods_receipt_items")
		receivedQuery, err := applyTenantJoinScope(ctx, receivedQuery, "goods_receipt_items.tenant_id", "goods_receipts.tenant_id")
		if err != nil {
			return err
		}
		if err := receivedQuery.
			Select("product_id, SUM(quantity_received) as qty").
			Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
			Where("goods_receipts.purchase_order_id = ? AND goods_receipts.status IN ?", gr.PurchaseOrderID, []string{
				string(models.GoodsReceiptStatusApproved),
				string(models.GoodsReceiptStatusPartial),
				string(models.GoodsReceiptStatusClosed),
			}).
			Group("product_id").
			Scan(&receivedSums).Error; err != nil {
			return err
		}
		receivedByProduct := make(map[string]float64, len(receivedSums))
		for _, row := range receivedSums {
			receivedByProduct[row.ProductID] = row.Qty
		}

		var invoicedSums []qtySum
		invoicedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
			Table("supplier_invoice_items")
		invoicedQuery, err = applyTenantJoinScope(ctx, invoicedQuery, "supplier_invoice_items.tenant_id", "supplier_invoices.tenant_id")
		if err != nil {
			return err
		}
		if err := invoicedQuery.
			Select("supplier_invoice_items.product_id, SUM(supplier_invoice_items.quantity) as qty").
			Joins("JOIN supplier_invoices ON supplier_invoices.id = supplier_invoice_items.supplier_invoice_id").
			Where("supplier_invoices.purchase_order_id = ? AND supplier_invoices.status NOT IN ?", gr.PurchaseOrderID, []string{
				string(models.SupplierInvoiceStatusRejected),
				string(models.SupplierInvoiceStatusCancelled),
			}).
			Where("supplier_invoices.deleted_at IS NULL").
			Group("supplier_invoice_items.product_id").
			Scan(&invoicedSums).Error; err != nil {
			return err
		}
		invoicedByProduct := make(map[string]float64, len(invoicedSums))
		for _, row := range invoicedSums {
			invoicedByProduct[row.ProductID] = row.Qty
		}

		grItemByProduct := make(map[string]models.GoodsReceiptItem, len(gr.Items))
		for _, item := range gr.Items {
			grItemByProduct[item.ProductID] = item
		}

		requestQtyByProduct := make(map[string]float64)
		subtotal := 0.0
		for _, item := range req.Items {
			discount := math.Max(0, math.Min(100, item.Discount))
			line := round2dp(item.Quantity * item.Price * (1 - discount/100))
			subtotal += line
			requestQtyByProduct[item.ProductID] += item.Quantity
		}

		grirVarianceTotal := calculateGRIRVarianceFromCreateItems(req.Items, poPriceByProduct)

		matching := make([]dto.SupplierInvoiceThreeWayMatchingLine, 0, len(requestQtyByProduct))
		for productID, qtyBill := range requestQtyByProduct {
			grItem := grItemByProduct[productID]
			productName := ""
			productCode := ""
			if grItem.Product != nil {
				productName = strings.TrimSpace(grItem.Product.Name)
				productCode = strings.TrimSpace(grItem.Product.Code)
			}

			qtyPO := poQtyByProduct[productID]
			qtyGR := receivedByProduct[productID]
			qtyAlreadyBilled := invoicedByProduct[productID]
			qtyRemaining := qtyGR - qtyAlreadyBilled - qtyBill
			isValid := qtyRemaining >= -0.0001

			matching = append(matching, dto.SupplierInvoiceThreeWayMatchingLine{
				ProductID:             productID,
				ProductCode:           productCode,
				ProductName:           productName,
				QuantityPO:            qtyPO,
				QuantityGR:            qtyGR,
				QuantityAlreadyBilled: qtyAlreadyBilled,
				QuantityBill:          qtyBill,
				QuantityRemaining:     qtyRemaining,
				IsValid:               isValid,
			})
		}

		taxRate := math.Max(0, math.Min(100, req.TaxRate))
		taxAmount := round2dp(subtotal * taxRate / 100)
		totalAmount := round2dp(subtotal + taxAmount + math.Max(0, req.DeliveryCost) + math.Max(0, req.OtherCost))

		dpAmount := 0.0
		var dpInvoices []models.SupplierInvoice
		if err := tx.Where("purchase_order_id = ? AND type = ? AND status = ? AND deleted_at IS NULL",
			gr.PurchaseOrderID, models.SupplierInvoiceTypeDownPayment, models.SupplierInvoiceStatusPaid).
			Find(&dpInvoices).Error; err == nil {
			for _, dp := range dpInvoices {
				dpAmount += dp.PaidAmount
			}
		}

		companyID, fiscalYearID, err := resolvePurchaseJournalScope(ctx, uc.db, gr.CompanyID, gr.FiscalYearID, req.InvoiceDate)
		if err != nil {
			return err
		}

		journalReq, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSupplierInvoice, accounting.TransactionData{
			ReferenceType:     reference.RefTypeSupplierInvoice,
			ReferenceID:       "preview",
			CompanyID:         companyID,
			FiscalYearID:      fiscalYearID,
			EntryDate:         req.InvoiceDate,
			Description:       fmt.Sprintf("Supplier Invoice Preview %s", req.InvoiceNumber),
			TotalAmount:       totalAmount,
			SubTotal:          subtotal,
			TaxTotal:          taxAmount,
			DepositTotal:      dpAmount,
			OtherTotal:        math.Max(0, req.DeliveryCost) + math.Max(0, req.OtherCost),
			GRIRVarianceTotal: grirVarianceTotal,
			DescriptionArgs:   []interface{}{req.InvoiceNumber, gr.SupplierNameSnapshot},
		})
		if err != nil {
			return err
		}

		preview = &dto.SupplierInvoiceJournalPreviewResponse{
			ReferenceType:    reference.RefTypeSupplierInvoice,
			ReferenceID:      "preview",
			InvoiceDate:      req.InvoiceDate,
			InvoiceNumber:    req.InvoiceNumber,
			Subtotal:         subtotal,
			TaxAmount:        taxAmount,
			DownPayment:      dpAmount,
			TotalAmount:      totalAmount,
			IsBalanced:       isSupplierInvoicePreviewBalanced(journalReq),
			ThreeWayMatching: matching,
			Lines:            make([]dto.SupplierInvoiceJournalPreviewLine, 0, len(journalReq.Lines)),
		}

		for _, line := range journalReq.Lines {
			preview.Lines = append(preview.Lines, dto.SupplierInvoiceJournalPreviewLine{
				ChartOfAccountID: line.ChartOfAccountID,
				Debit:            line.Debit,
				Credit:           line.Credit,
				Memo:             line.Memo,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return preview, nil
}

func isSupplierInvoicePreviewBalanced(req *finDto.CreateJournalEntryRequest) bool {
	if req == nil {
		return false
	}

	var debitTotal float64
	var creditTotal float64
	for _, line := range req.Lines {
		debitTotal += line.Debit
		creditTotal += line.Credit
	}

	return math.Abs(debitTotal-creditTotal) <= 0.001
}

func calculateGRIRVarianceFromCreateItems(items []dto.CreateSupplierInvoiceItemRequest, poPriceByProduct map[string]float64) float64 {
	if len(items) == 0 || len(poPriceByProduct) == 0 {
		return 0
	}

	variance := 0.0
	for _, item := range items {
		poPrice, ok := poPriceByProduct[item.ProductID]
		if !ok || poPrice <= 0 {
			continue
		}
		discount := math.Max(0, math.Min(100, item.Discount))
		invoiceLineTotal := round2dp(item.Quantity * item.Price * (1 - discount/100))
		expectedLineTotal := round2dp(item.Quantity * poPrice)
		variance += invoiceLineTotal - expectedLineTotal
	}

	return round2dp(variance)
}

func calculateGRIRVarianceFromInvoiceItems(items []models.SupplierInvoiceItem, poPriceByProduct map[string]float64) float64 {
	if len(items) == 0 || len(poPriceByProduct) == 0 {
		return 0
	}

	variance := 0.0
	for _, item := range items {
		poPrice, ok := poPriceByProduct[item.ProductID]
		if !ok || poPrice <= 0 {
			continue
		}
		discount := math.Max(0, math.Min(100, item.Discount))
		invoiceLineTotal := round2dp(item.Quantity * item.Price * (1 - discount/100))
		expectedLineTotal := round2dp(item.Quantity * poPrice)
		variance += invoiceLineTotal - expectedLineTotal
	}

	return round2dp(variance)
}

func poNetUnitPrice(item models.PurchaseOrderItem) float64 {
	if item.Quantity <= 0 {
		return round2dp(item.Price)
	}
	return round2dp(item.Subtotal / item.Quantity)
}

func (uc *supplierInvoiceUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.SupplierInvoiceAuditTrailEntry, int64, error) {
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

	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.SupplierInvoice{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, 0, ErrSupplierInvoiceNotFound
	}

	tx := uc.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("audit_logs.permission_code LIKE ?", "supplier_invoice.%")

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

	entries := make([]dto.SupplierInvoiceAuditTrailEntry, 0, len(rows))
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

		entries = append(entries, dto.SupplierInvoiceAuditTrailEntry{
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

func (uc *supplierInvoiceUsecase) syncGoodsReceiptStatus(ctx context.Context, tx *gorm.DB, grID string) error {
	if grID == "" {
		return nil
	}
	var gr models.GoodsReceipt
	if err := tx.Preload("Items").First(&gr, "id = ?", grID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	// We only sync if GR is already Approved or Partial or Closed
	if gr.Status == models.GoodsReceiptStatusDraft || gr.Status == models.GoodsReceiptStatusSubmitted || gr.Status == models.GoodsReceiptStatusRejected {
		return nil
	}

	type qtySum struct {
		ProductID string
		Qty       float64
	}
	var invoicedSums []qtySum
	invoicedQuery := tx.Session(&gorm.Session{NewDB: true}).WithContext(ctx).
		Table("supplier_invoice_items")
	invoicedQuery, err := applyTenantJoinScope(ctx, invoicedQuery, "supplier_invoice_items.tenant_id", "supplier_invoices.tenant_id")
	if err != nil {
		return err
	}

	if err := invoicedQuery.
		Select("supplier_invoice_items.product_id, SUM(supplier_invoice_items.quantity) as qty").
		Joins("JOIN supplier_invoices ON supplier_invoices.id = supplier_invoice_items.supplier_invoice_id").
		Where("supplier_invoices.goods_receipt_id = ? AND supplier_invoices.status NOT IN ?",
			gr.ID, []string{
				string(models.SupplierInvoiceStatusRejected),
				string(models.SupplierInvoiceStatusCancelled),
			}).
		Where("supplier_invoices.deleted_at IS NULL").
		Group("supplier_invoice_items.product_id").
		Scan(&invoicedSums).Error; err != nil {
		return err
	}

	invoicedMap := make(map[string]float64)
	for _, s := range invoicedSums {
		invoicedMap[s.ProductID] = s.Qty
	}

	allDone := true
	hasAny := false
	for _, it := range gr.Items {
		inv := invoicedMap[it.ProductID]
		if inv > 0 {
			hasAny = true
		}
		if inv < it.QuantityReceived-0.0001 {
			allDone = false
		}
	}

	newStatus := models.GoodsReceiptStatusApproved
	if allDone && len(gr.Items) > 0 {
		newStatus = models.GoodsReceiptStatusClosed
	} else if hasAny {
		newStatus = models.GoodsReceiptStatusPartial
	}

	if gr.Status != newStatus {
		updates := map[string]interface{}{"status": newStatus}
		if newStatus == models.GoodsReceiptStatusClosed {
			now := apptime.Now()
			updates["closed_at"] = &now
		}
		return tx.Model(&gr).Updates(updates).Error
	}
	return nil
}
