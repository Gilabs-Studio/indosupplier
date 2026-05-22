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

type SupplierInvoiceDownPaymentUsecase interface {
	AddData(ctx context.Context) (*dto.SupplierInvoiceDownPaymentAddResponse, error)
	List(ctx context.Context, params repositories.SupplierInvoiceListParams) ([]*dto.SupplierInvoiceDownPaymentListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateSupplierInvoiceDownPaymentRequest) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSupplierInvoiceDownPaymentRequest) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Delete(ctx context.Context, id string) error
	Pending(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Submit(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Approve(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Reject(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	Cancel(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.SupplierInvoiceAuditTrailEntry, int64, error)
	TriggerJournalForInvoiceDP(ctx context.Context, si *models.SupplierInvoice) error
}

type supplierInvoiceDownPaymentUsecase struct {
	db                 *gorm.DB
	repo               repositories.SupplierInvoiceRepository
	poRepo             repositories.PurchaseOrderRepository
	auditService       audit.AuditService
	mapper             *mapper.SupplierInvoiceMapper
	journalUC          finUsecase.JournalEntryUsecase
	coaUC              finUsecase.ChartOfAccountUsecase
	engine             accounting.AccountingEngine
	purchaseJournalSvc purchaseService.PurchaseJournalService
}

func NewSupplierInvoiceDownPaymentUsecase(db *gorm.DB, repo repositories.SupplierInvoiceRepository, poRepo repositories.PurchaseOrderRepository, auditService audit.AuditService, journalUC finUsecase.JournalEntryUsecase, coaUC finUsecase.ChartOfAccountUsecase, engine accounting.AccountingEngine, purchaseJournalSvc ...purchaseService.PurchaseJournalService) SupplierInvoiceDownPaymentUsecase {
	uc := &supplierInvoiceDownPaymentUsecase{db: db, repo: repo, poRepo: poRepo, auditService: auditService, mapper: mapper.NewSupplierInvoiceMapper(), journalUC: journalUC, coaUC: coaUC, engine: engine, purchaseJournalSvc: purchaseService.NewPurchaseJournalService(db, journalUC, engine)}
	if len(purchaseJournalSvc) > 0 && purchaseJournalSvc[0] != nil {
		uc.purchaseJournalSvc = purchaseJournalSvc[0]
	}

	return uc
}

func (uc *supplierInvoiceDownPaymentUsecase) AddData(ctx context.Context) (*dto.SupplierInvoiceDownPaymentAddResponse, error) {
	// Fetch APPROVED POs for DP creation.
	// DP can be created before GR/regular supplier invoice exists.
	var pos []*models.PurchaseOrder
	err := database.GetDB(ctx, uc.db).
		Where("status = ?", models.PurchaseOrderStatusApproved).
		Preload("Items").
		Preload("Items.Product").
		Preload("Supplier").
		Order("created_at DESC").
		Limit(100).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

	settledPOs := make(map[string]bool)
	hasActiveDPPOs := make(map[string]bool)
	if len(pos) > 0 {
		poIDs := make([]string, 0, len(pos))
		for _, po := range pos {
			poIDs = append(poIDs, po.ID)
		}

		type poSettlementRow struct {
			PurchaseOrderID   string `gorm:"column:purchase_order_id"`
			ActiveInvoiceCnt  int64  `gorm:"column:active_invoice_count"`
			UnsettledCount    int64  `gorm:"column:unsettled_count"`
		}

		rows := make([]poSettlementRow, 0)
		if err := database.GetDB(ctx, uc.db).
			Table("supplier_invoices").
			Select(`purchase_order_id,
				COUNT(*) AS active_invoice_count,
				SUM(CASE WHEN UPPER(status) IN ('PAID','CANCELLED','REJECTED') THEN 0 ELSE 1 END) AS unsettled_count`).
			Where("purchase_order_id IN ?", poIDs).
			Where("type = ?", models.SupplierInvoiceTypeNormal).
			Where("deleted_at IS NULL").
			Group("purchase_order_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}

		for _, row := range rows {
			if row.ActiveInvoiceCnt > 0 && row.UnsettledCount == 0 {
				settledPOs[row.PurchaseOrderID] = true
			}
		}

		type poDPRow struct {
			PurchaseOrderID string `gorm:"column:purchase_order_id"`
			ActiveDPCount   int64  `gorm:"column:active_dp_count"`
		}

		dpRows := make([]poDPRow, 0)
		if err := database.GetDB(ctx, uc.db).
			Table("supplier_invoices").
			Select("purchase_order_id, COUNT(*) AS active_dp_count").
			Where("purchase_order_id IN ?", poIDs).
			Where("type = ?", models.SupplierInvoiceTypeDownPayment).
			Where("status NOT IN ?", []models.SupplierInvoiceStatus{
				models.SupplierInvoiceStatusCancelled,
				models.SupplierInvoiceStatusRejected,
			}).
			Where("deleted_at IS NULL").
			Group("purchase_order_id").
			Scan(&dpRows).Error; err != nil {
			return nil, err
		}

		for _, row := range dpRows {
			if row.ActiveDPCount > 0 {
				hasActiveDPPOs[row.PurchaseOrderID] = true
			}
		}
	}

	poRes := make([]dto.SupplierInvoiceAddPurchaseOrder, 0, len(pos))
	for _, po := range pos {
		if settledPOs[po.ID] {
			continue
		}
		if hasActiveDPPOs[po.ID] {
			continue
		}
		items := make([]dto.SupplierInvoiceAddPurchaseOrderItem, 0, len(po.Items))
		for _, it := range po.Items {
			var prod *dto.SupplierInvoiceAddProductMini
			if it.Product != nil {
				prod = &dto.SupplierInvoiceAddProductMini{ID: it.Product.ID, Name: it.Product.Name, Code: it.Product.Code, ImageURL: it.Product.ImageURL}
			}
			items = append(items, dto.SupplierInvoiceAddPurchaseOrderItem{ID: it.ID, Product: prod, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal})
		}
		var sup *dto.SupplierInvoiceAddSupplierMini
		if po.Supplier != nil {
			sup = &dto.SupplierInvoiceAddSupplierMini{ID: po.Supplier.ID, Name: po.Supplier.Name}
		}
		poRes = append(poRes, dto.SupplierInvoiceAddPurchaseOrder{ID: po.ID, Supplier: sup, Code: po.Code, OrderDate: po.OrderDate, Status: string(po.Status), TotalAmount: po.TotalAmount, Items: items})
	}
	return &dto.SupplierInvoiceDownPaymentAddResponse{PurchaseOrders: poRes}, nil
}

func (uc *supplierInvoiceDownPaymentUsecase) List(ctx context.Context, params repositories.SupplierInvoiceListParams) ([]*dto.SupplierInvoiceDownPaymentListResponse, int64, error) {
	params.Type = string(models.SupplierInvoiceTypeDownPayment)
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// For each DP, if RegularInvoices is empty, try to find them by PO ID (fallback for older records or multiple DPs)
	for i := range items {
		if len(items[i].RegularInvoices) == 0 && items[i].PurchaseOrderID != "" {
			var regulars []models.SupplierInvoice
			if err := database.GetDB(ctx, uc.db).
				Where("purchase_order_id = ? AND type = ? AND deleted_at IS NULL", items[i].PurchaseOrderID, models.SupplierInvoiceTypeNormal).
				Find(&regulars).Error; err == nil {
				items[i].RegularInvoices = regulars
			}
		}
	}

	return uc.mapper.ToDownPaymentListResponseList(items), total, nil
}

func (uc *supplierInvoiceDownPaymentUsecase) GetByID(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
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
	if si.Type != models.SupplierInvoiceTypeDownPayment {
		return nil, ErrSupplierInvoiceNotFound
	}

	// Fallback to PO-based lookup if slice is empty
	if len(si.RegularInvoices) == 0 && si.PurchaseOrderID != "" {
		var regulars []models.SupplierInvoice
		if err := database.GetDB(ctx, uc.db).
			Where("purchase_order_id = ? AND type = ? AND deleted_at IS NULL", si.PurchaseOrderID, models.SupplierInvoiceTypeNormal).
			Find(&regulars).Error; err == nil {
			si.RegularInvoices = regulars
		}
	}

	return uc.mapper.ToDownPaymentDetailResponse(si), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Create(ctx context.Context, req *dto.CreateSupplierInvoiceDownPaymentRequest) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var out *models.SupplierInvoice
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po models.PurchaseOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&po, "id = ?", req.PurchaseOrderID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPurchaseOrderNotFound
			}
			return err
		}
		if po.Status != models.PurchaseOrderStatusApproved {
			return ErrInvalidStatus
		}
		if po.SupplierID == nil || strings.TrimSpace(*po.SupplierID) == "" {
			return ErrSupplierInvoiceInvalid
		}

		code, err := getNextSupplierInvoiceCodeLocked(tx, "SIDP")
		if err != nil {
			return err
		}
		invNo := fmt.Sprintf("SUP-DP-%s-%s", apptime.Now().Format("20060102"), strings.TrimPrefix(code, "SIDP-"))

		creatorID, _ := ctx.Value("user_id").(string)

		invoiceDateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.InvoiceDate))
		if err != nil {
			return errors.New("invalid invoice date format")
		}

		dueDateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate))
		if err != nil {
			return errors.New("invalid due date format")
		}

		si := models.SupplierInvoice{
			Type:            models.SupplierInvoiceTypeDownPayment,
			PurchaseOrderID: po.ID,
			SupplierID:      *po.SupplierID,
			Code:            code,
			InvoiceNumber:   invNo,
			InvoiceDate:     invoiceDateParsed,
			DueDate:         dueDateParsed,
			Amount:          req.Amount,
			Status:          models.SupplierInvoiceStatusDraft,
			Notes:           req.Notes,
			CreatedBy:       creatorID,
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

		var activeDPCount int64
		if err := tx.Model(&models.SupplierInvoice{}).
			Where("purchase_order_id = ?", po.ID).
			Where("type = ?", models.SupplierInvoiceTypeDownPayment).
			Where("status NOT IN ?", []models.SupplierInvoiceStatus{
				models.SupplierInvoiceStatusCancelled,
				models.SupplierInvoiceStatusRejected,
			}).
			Where("deleted_at IS NULL").
			Count(&activeDPCount).Error; err != nil {
			return err
		}
		if activeDPCount > 0 {
			return ErrSupplierInvoiceConflict
		}

		if err := tx.Create(&si).Error; err != nil {
			return err
		}

		// NOTE: Auto-create of regular SI has been removed.
		// Regular Supplier Invoices are now created from Goods Receipts, not from PO/DP.
		// The DP will be auto-applied when creating an SI from a GR that shares the same PO.

		// Fix reloading using the current transaction
		var s models.SupplierInvoice
		if err := tx.Preload("PurchaseOrder").
			Preload("PaymentTerms").
			Preload("DownPaymentInvoice").
			Preload("RegularInvoices").
			First(&s, "id = ?", si.ID).Error; err != nil {
			return err
		}
		out = &s
		return nil
	})
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice_dp.create", out.ID, map[string]interface{}{"after": out})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Update(ctx context.Context, id string, req *dto.UpdateSupplierInvoiceDownPaymentRequest) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	if existing.Type != models.SupplierInvoiceTypeDownPayment {
		return nil, ErrSupplierInvoiceNotFound
	}
	if existing.Status != models.SupplierInvoiceStatusDraft {
		return nil, ErrSupplierInvoiceConflict
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var out *models.SupplierInvoice
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Status != models.SupplierInvoiceStatusDraft || si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceConflict
		}

		var po models.PurchaseOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, "id = ?", req.PurchaseOrderID).Error; err != nil {
			return err
		}
		if po.Status != models.PurchaseOrderStatusApproved {
			return ErrInvalidStatus
		}
		if po.SupplierID == nil || strings.TrimSpace(*po.SupplierID) == "" {
			return ErrSupplierInvoiceInvalid
		}

		updates := map[string]interface{}{
			"purchase_order_id": po.ID,
			"supplier_id":       *po.SupplierID,
			"company_id":        po.CompanyID,
			"fiscal_year_id":    po.FiscalYearID,
			"invoice_date":      req.InvoiceDate,
			"due_date":          req.DueDate,
			"amount":            req.Amount,
			"notes":             req.Notes,
			"updated_at":        apptime.Now(),
		}
		if po.CompanyID == nil || po.FiscalYearID == nil {
			companyID, fiscalYearID, err := newPurchaseFiscalYearResolver(uc.db, financeRepositories.NewFiscalYearRepository(uc.db)).Resolve(ctx, req.InvoiceDate)
			if err != nil {
				return err
			}
			updates["company_id"] = &companyID
			updates["fiscal_year_id"] = &fiscalYearID
		}
		if err := tx.Model(&si).Updates(updates).Error; err != nil {
			return err
		}

		// Fix reloading using the current transaction
		var s models.SupplierInvoice
		if err := tx.Preload("PurchaseOrder").
			Preload("PaymentTerms").
			Preload("DownPaymentInvoice").
			Preload("RegularInvoices").
			First(&s, "id = ?", id).Error; err != nil {
			return err
		}
		out = &s
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice_dp.update", id, map[string]interface{}{"after": out})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrSupplierInvoiceNotFound
		}
		return err
	}
	if existing.Type != models.SupplierInvoiceTypeDownPayment {
		return ErrSupplierInvoiceNotFound
	}
	// Allow deletion of draft or unpaid invoices (aligned with Customer Invoice DP pattern)
	if existing.Status != models.SupplierInvoiceStatusDraft && existing.Status != models.SupplierInvoiceStatusUnpaid {
		return ErrSupplierInvoiceConflict
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditService.Log(ctx, "supplier_invoice_dp.delete", id, map[string]interface{}{"before": existing})
	return nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Pending(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var out *models.SupplierInvoice
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceNotFound
		}
		if si.Status != models.SupplierInvoiceStatusDraft {
			return ErrSupplierInvoiceConflict
		}
		if err := tx.Model(&si).Update("status", models.SupplierInvoiceStatusUnpaid).Error; err != nil {
			return err
		}

		// Fix reloading using the current transaction
		var s models.SupplierInvoice
		if err := tx.Preload("PurchaseOrder").
			Preload("PaymentTerms").
			Preload("DownPaymentInvoice").
			Preload("RegularInvoices").
			First(&s, "id = ?", id).Error; err != nil {
			return err
		}
		out = &s

		// Trigger Journal Entry for DP recognition (inside transaction)
		// Debit: Purchase Advances (11900) / Credit: AP (21000)
		txCtx := database.WithTx(ctx, tx)
		if err := uc.triggerDPJournalEntry(txCtx, &si); err != nil {
			fmt.Printf("⚠️ Failed to create journal entry for supplier invoice DP %s: %v\n", id, err)
			// Don't fail the pending operation if journal fails
		}

		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSupplierInvoiceNotFound
		}
		return nil, err
	}
	uc.auditService.Log(ctx, "supplier_invoice_dp.pending", id, map[string]interface{}{"after": out})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Submit(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		if si.Status != models.SupplierInvoiceStatusDraft {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		return tx.Model(&si).Updates(map[string]interface{}{
			"status":       models.SupplierInvoiceStatusSubmitted,
			"submitted_at": &now,
		}).Error
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
	uc.auditService.Log(ctx, "supplier_invoice_dp.submit", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "supplier_invoice_dp.approve",
		EntityType:     "supplier_invoice_dp",
		EntityID:       out.ID,
		Title:          "Supplier Invoice Down Payment Approval",
		Message:        "A supplier invoice down payment has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create supplier invoice DP notification: %v", err)
	}
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Approve(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		if si.Status != models.SupplierInvoiceStatusSubmitted {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		// Approve and immediately transition to UNPAID so it appears in the payment form
		if err := tx.Model(&si).Updates(map[string]interface{}{
			"status":           models.SupplierInvoiceStatusUnpaid,
			"approved_at":      &now,
			"remaining_amount": si.Amount,
		}).Error; err != nil {
			return err
		}

		// Trigger Journal Entry for DP recognition (Debit: Purchase Advances, Credit: AP)
		txCtx := database.WithTx(ctx, tx)
		if err := uc.triggerDPJournalEntry(txCtx, &si); err != nil {
			fmt.Printf("⚠️ Failed to create journal entry for supplier invoice DP %s: %v\n", id, err)
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
	uc.auditService.Log(ctx, "supplier_invoice_dp.approve", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Reject(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		if si.Status != models.SupplierInvoiceStatusSubmitted {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		return tx.Model(&si).Updates(map[string]interface{}{
			"status":      models.SupplierInvoiceStatusRejected,
			"rejected_at": &now,
		}).Error
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
	uc.auditService.Log(ctx, "supplier_invoice_dp.reject", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) Cancel(ctx context.Context, id string) (*dto.SupplierInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}
	var beforeStatus models.SupplierInvoiceStatus
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var si models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&si, "id = ?", id).Error; err != nil {
			return err
		}
		if si.Type != models.SupplierInvoiceTypeDownPayment {
			return ErrSupplierInvoiceNotFound
		}
		beforeStatus = si.Status
		allowed := si.Status == models.SupplierInvoiceStatusDraft ||
			si.Status == models.SupplierInvoiceStatusSubmitted ||
			si.Status == models.SupplierInvoiceStatusApproved ||
			si.Status == models.SupplierInvoiceStatusUnpaid
		if !allowed {
			return ErrSupplierInvoiceConflict
		}
		now := apptime.Now()
		return tx.Model(&si).Updates(map[string]interface{}{
			"status":       models.SupplierInvoiceStatusCancelled,
			"cancelled_at": &now,
		}).Error
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
	uc.auditService.Log(ctx, "supplier_invoice_dp.cancel", id, map[string]interface{}{
		"before_status": beforeStatus,
		"after_status":  out.Status,
		"before":        map[string]interface{}{"status": beforeStatus},
		"after":         map[string]interface{}{"status": out.Status},
	})
	return uc.mapper.ToDownPaymentDetailResponse(out), nil
}

func (uc *supplierInvoiceDownPaymentUsecase) TriggerJournalForInvoiceDP(ctx context.Context, si *models.SupplierInvoice) error {
	return uc.triggerDPJournalEntry(ctx, si)
}

func (uc *supplierInvoiceDownPaymentUsecase) triggerDPJournalEntry(ctx context.Context, si *models.SupplierInvoice) error {
	if si == nil || uc.journalUC == nil || uc.engine == nil {
		return nil
	}

	if si.Amount <= 0 {
		return nil
	}

	companyID, fiscalYearID, err := resolvePurchaseJournalScope(ctx, uc.db, si.CompanyID, si.FiscalYearID, si.InvoiceDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("failed to resolve supplier invoice DP journal scope: %w", err)
	}

	data := accounting.TransactionData{
		ReferenceType:   "SUPPLIER_INVOICE_DP",
		ReferenceID:     si.ID,
		CompanyID:       companyID,
		FiscalYearID:    fiscalYearID,
		EntryDate:       si.InvoiceDate.Format("2006-01-02"),
		Description:     fmt.Sprintf("Purchase Down Payment %s (%s)", si.InvoiceNumber, si.Code),
		TotalAmount:     si.Amount,
		DescriptionArgs: []interface{}{si.InvoiceNumber, si.Code},
	}

	if uc.purchaseJournalSvc != nil {
		if _, err := uc.purchaseJournalSvc.GenerateDPJournal(ctx, purchaseService.PurchaseJournalTxn{
			Profile: accounting.ProfileSupplierInvoiceDP,
			Data:    data,
		}); err != nil {
			return fmt.Errorf("failed to post supplier invoice DP journal: %w", err)
		}

		log.Printf("journal_observability event=trigger.success module=purchase_supplier_invoice_dp reference_id=%s", si.ID)
		return nil
	}

	req, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSupplierInvoiceDP, data)
	if err != nil {
		return fmt.Errorf("failed to generate supplier invoice DP journal: %w", err)
	}

	// Balance check
	var debitTotal, creditTotal float64
	for _, l := range req.Lines {
		debitTotal += l.Debit
		creditTotal += l.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.001 {
		return fmt.Errorf("generated supplier invoice DP journal is unbalanced: debit=%.2f credit=%.2f", debitTotal, creditTotal)
	}

	_, err = uc.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post supplier invoice DP journal: %w", err)
	}

	log.Printf("journal_observability event=trigger.success module=purchase_supplier_invoice_dp reference_id=%s", si.ID)
	return nil
}

func (uc *supplierInvoiceDownPaymentUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.SupplierInvoiceAuditTrailEntry, int64, error) {
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
		Where("audit_logs.permission_code LIKE ?", "supplier_invoice_dp.%")

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
