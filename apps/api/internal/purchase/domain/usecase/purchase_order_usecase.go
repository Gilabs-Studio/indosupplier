package usecase

import (
	"context"
	"errors"
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
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/mapper"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPurchaseOrderNotFound = errors.New("purchase order not found")
	ErrPurchaseOrderConflict = errors.New("purchase order conflict")
	ErrSalesOrderNotFound    = errors.New("sales order not found")
)

type PurchaseOrderUsecase interface {
	List(ctx context.Context, params repositories.PurchaseOrderListParams) ([]*dto.PurchaseOrderListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error)
	Create(ctx context.Context, req *dto.CreatePurchaseOrderRequest) (*dto.PurchaseOrderDetailResponse, error)
	CreateFromPurchaseRequisition(ctx context.Context, purchaseRequisitionID string) (*dto.PurchaseOrderDetailResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdatePurchaseOrderRequest) (*dto.PurchaseOrderDetailResponse, error)
	Delete(ctx context.Context, id string) error
	Submit(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error)
	Approve(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error)
	Reject(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error)
	AddData(ctx context.Context) (*dto.PurchaseOrderAddResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.PurchaseOrderAuditTrailEntry, int64, error)
}

type purchaseOrderUsecase struct {
	db                 *gorm.DB
	repo               repositories.PurchaseOrderRepository
	prRepo             repositories.PurchaseRequisitionRepository
	fiscalYearResolver *purchaseFiscalYearResolver
	mapper             *mapper.PurchaseOrderMapper
	auditService       audit.AuditService
}

func NewPurchaseOrderUsecase(db *gorm.DB, repo repositories.PurchaseOrderRepository, prRepo repositories.PurchaseRequisitionRepository, auditService audit.AuditService, fiscalYearRepo financeRepositories.FiscalYearRepository) PurchaseOrderUsecase {
	return &purchaseOrderUsecase{
		db:                 db,
		repo:               repo,
		prRepo:             prRepo,
		fiscalYearResolver: newPurchaseFiscalYearResolver(db, fiscalYearRepo),
		mapper:             mapper.NewPurchaseOrderMapper(),
		auditService:       auditService,
	}
}

func (uc *purchaseOrderUsecase) List(ctx context.Context, params repositories.PurchaseOrderListParams) ([]*dto.PurchaseOrderListResponse, int64, error) {
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	for _, po := range items {
		if po == nil || po.Status != models.PurchaseOrderStatusApproved || len(po.Items) == 0 {
			continue
		}

		var totalOrdered, totalReceived float64
		for _, item := range po.Items {
			totalOrdered += item.Quantity
		}
		for _, gr := range po.GoodsReceipts {
			s := strings.ToUpper(strings.TrimSpace(string(gr.Status)))
			if s != string(models.GoodsReceiptStatusApproved) && s != string(models.GoodsReceiptStatusPartial) && s != string(models.GoodsReceiptStatusClosed) && s != string(models.GoodsReceiptStatusConfirmed) {
				continue
			}
			for _, grItem := range gr.Items {
				totalReceived += grItem.QuantityReceived
			}
		}
		if totalOrdered <= 0 || totalReceived+0.0001 < totalOrdered {
			continue
		}

		normalInvoices := make([]models.SupplierInvoice, 0, len(po.SupplierInvoices))
		for _, inv := range po.SupplierInvoices {
			if inv.Type == models.SupplierInvoiceTypeNormal {
				normalInvoices = append(normalInvoices, inv)
			}
		}

		hasInvoice := len(normalInvoices) > 0
		if !hasInvoice {
			continue
		}

		allSettled := true
		for _, inv := range normalInvoices {
			status := strings.ToUpper(strings.TrimSpace(string(inv.Status)))
			if status != string(models.SupplierInvoiceStatusPaid) && status != string(models.SupplierInvoiceStatusCancelled) && status != string(models.SupplierInvoiceStatusRejected) {
				allSettled = false
				break
			}
		}
		if !allSettled {
			continue
		}

		now := apptime.Now()
		_ = database.GetDB(ctx, uc.db).Model(&models.PurchaseOrder{}).
			Where("id = ?", po.ID).
			Updates(map[string]interface{}{
				"status":     models.PurchaseOrderStatusClosed,
				"closed_at":  &now,
				"updated_at": now,
			}).Error
		po.Status = models.PurchaseOrderStatusClosed
	}

	return uc.mapper.ToListResponseList(items), total, nil
}

func (uc *purchaseOrderUsecase) GetByID(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.PurchaseOrder{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, ErrPurchaseOrderNotFound
	}

	po, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToDetailResponse(po)

	// Compute per-item confirmed received quantity for UI and accounting visibility.
	// Only APPROVED/PARTIAL/CLOSED/CONFIRMED GRs are considered as physically received.
	type receivedRow struct {
		PurchaseOrderItemID string  `gorm:"column:purchase_order_item_id"`
		TotalReceived       float64 `gorm:"column:total_received"`
	}
	var confirmedRows []receivedRow
	confirmedQuery := uc.db.WithContext(ctx).
		Table("goods_receipt_items")
	confirmedQuery, err = applyTenantJoinScope(ctx, confirmedQuery, "goods_receipt_items.tenant_id", "gr.tenant_id")
	if err != nil {
		return nil, err
	}

	if err := confirmedQuery.
		Select("goods_receipt_items.purchase_order_item_id, COALESCE(SUM(goods_receipt_items.quantity_received), 0) AS total_received").
		Joins("JOIN goods_receipts gr ON gr.id = goods_receipt_items.goods_receipt_id").
		Where("gr.purchase_order_id = ?", id).
		Where("UPPER(gr.status) IN ?", []string{
			string(models.GoodsReceiptStatusApproved),
			string(models.GoodsReceiptStatusPartial),
			string(models.GoodsReceiptStatusClosed),
			string(models.GoodsReceiptStatusConfirmed),
		}).
		Group("goods_receipt_items.purchase_order_item_id").
		Scan(&confirmedRows).Error; err != nil {
		return nil, err
	}
	confirmedByItemID := make(map[string]float64, len(confirmedRows))
	for _, r := range confirmedRows {
		confirmedByItemID[r.PurchaseOrderItemID] = r.TotalReceived
	}

	// Compute reservation from all active GRs (exclude REJECTED only) so draft/submitted
	// GRs still reserve quantity and prevent duplicate receiving.
	var reservedRows []receivedRow
	reservedQuery := uc.db.WithContext(ctx).
		Table("goods_receipt_items")
	reservedQuery, err = applyTenantJoinScope(ctx, reservedQuery, "goods_receipt_items.tenant_id", "gr.tenant_id")
	if err != nil {
		return nil, err
	}

	if err := reservedQuery.
		Select("goods_receipt_items.purchase_order_item_id, COALESCE(SUM(goods_receipt_items.quantity_received), 0) AS total_received").
		Joins("JOIN goods_receipts gr ON gr.id = goods_receipt_items.goods_receipt_id").
		Where("gr.purchase_order_id = ?", id).
		Where("UPPER(gr.status) <> ?", string(models.GoodsReceiptStatusRejected)).
		Group("goods_receipt_items.purchase_order_item_id").
		Scan(&reservedRows).Error; err != nil {
		return nil, err
	}
	reservedByItemID := make(map[string]float64, len(reservedRows))
	for _, r := range reservedRows {
		reservedByItemID[r.PurchaseOrderItemID] = r.TotalReceived
	}
	for i := range res.Items {
		res.Items[i].QuantityReceived = confirmedByItemID[res.Items[i].ID]
		remaining := res.Items[i].Quantity - reservedByItemID[res.Items[i].ID]
		if remaining < 0 {
			remaining = 0
		}
		res.Items[i].QuantityRemaining = remaining
	}

	return res, nil
}

func (uc *purchaseOrderUsecase) Create(ctx context.Context, req *dto.CreatePurchaseOrderRequest) (*dto.PurchaseOrderDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if req.PurchaseRequisitionID != nil && strings.TrimSpace(*req.PurchaseRequisitionID) != "" && req.SalesOrderID != nil && strings.TrimSpace(*req.SalesOrderID) != "" {
		return nil, ErrPurchaseOrderConflict
	}

	po := &models.PurchaseOrder{
		Code:                  "",
		SupplierID:            req.SupplierID,
		PaymentTermsID:        req.PaymentTermsID,
		BusinessUnitID:        req.BusinessUnitID,
		CreatedBy:             actorID,
		PurchaseRequisitionID: req.PurchaseRequisitionID,
		SalesOrderID:          req.SalesOrderID,
		OrderDate:             req.OrderDate,
		DueDate:               req.DueDate,
		Notes:                 req.Notes,
		Status:                models.PurchaseOrderStatusDraft,
		TaxRate:               clamp(req.TaxRate, 0, 100),
		DeliveryCost:          math.Max(0, req.DeliveryCost),
		OtherCost:             math.Max(0, req.OtherCost),
		Items:                 make([]models.PurchaseOrderItem, 0, len(req.Items)),
	}
	companyID, fiscalYearID, err := uc.fiscalYearResolver.Resolve(ctx, req.OrderDate)
	if err != nil {
		return nil, err
	}
	po.CompanyID = &companyID
	po.FiscalYearID = &fiscalYearID

	for _, it := range req.Items {
		discount := clamp(it.Discount, 0, 100)
		qty := math.Max(0, it.Quantity)
		price := math.Max(0, it.Price)
		subtotal := calcPOItemSubtotal(qty, price, discount)
		po.Items = append(po.Items, models.PurchaseOrderItem{
			ProductID: it.ProductID,
			Quantity:  qty,
			Price:     price,
			Discount:  discount,
			Subtotal:  subtotal,
			Notes:     it.Notes,
		})
	}

	sub, tax, total := calcPOTotals(po.Items, po.TaxRate, po.DeliveryCost, po.OtherCost)
	po.SubTotal = sub
	po.TaxAmount = tax
	po.TotalAmount = total

	// If create from PR: validate PR approved and not already converted to PO
	if po.PurchaseRequisitionID != nil && strings.TrimSpace(*po.PurchaseRequisitionID) != "" {
		existingPR, err := uc.prRepo.GetByID(ctx, *po.PurchaseRequisitionID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, ErrPurchaseRequisitionNotFound
			}
			return nil, err
		}
		if existingPR.Status != models.PurchaseRequisitionStatusApproved {
			return nil, ErrInvalidStatus
		}
		exists, err := uc.repo.ExistsByPurchaseRequisitionID(ctx, existingPR.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPurchaseOrderConflict
		}

		// Ensure PR -> PO integration keeps core procurement references even when frontend omits them.
		if po.SupplierID == nil || strings.TrimSpace(*po.SupplierID) == "" {
			po.SupplierID = existingPR.SupplierID
		}
		if po.PaymentTermsID == nil || strings.TrimSpace(*po.PaymentTermsID) == "" {
			po.PaymentTermsID = existingPR.PaymentTermsID
		}
		if po.BusinessUnitID == nil || strings.TrimSpace(*po.BusinessUnitID) == "" {
			po.BusinessUnitID = existingPR.BusinessUnitID
		}
	}

	// If create from Sales Order: validate SO exists and not already converted to PO
	if po.SalesOrderID != nil && strings.TrimSpace(*po.SalesOrderID) != "" {
		var soCount int64
		if err := database.GetDB(ctx, uc.db).
			Model(&salesModels.SalesOrder{}).
			Where("id = ?", *po.SalesOrderID).
			Count(&soCount).Error; err != nil {
			return nil, err
		}
		if soCount == 0 {
			return nil, ErrSalesOrderNotFound
		}
		exists, err := uc.repo.ExistsBySalesOrderID(ctx, *po.SalesOrderID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPurchaseOrderConflict
		}
	}

	if err := snapshotPurchaseOrderHeader(ctx, uc.db, po, nil); err != nil {
		return nil, err
	}
	if err := snapshotPurchaseOrderItems(ctx, uc.db, po, nil); err != nil {
		return nil, err
	}

	created, err := uc.repo.Create(ctx, po)
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.GetByID(ctx, created.ID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_order.create", created.ID, map[string]interface{}{
		"after": poAuditSnapshot(full),
	})

	// Side-effect: when created from PR, set PR to CONVERTED
	if full.PurchaseRequisitionID != nil && strings.TrimSpace(*full.PurchaseRequisitionID) != "" {
		_, _ = uc.prRepo.UpdateStatus(ctx, *full.PurchaseRequisitionID, models.PurchaseRequisitionStatusConverted)
		uc.auditService.Log(ctx, "purchase_requisition.convert", *full.PurchaseRequisitionID, map[string]interface{}{
			"after": map[string]interface{}{
				"id":     *full.PurchaseRequisitionID,
				"status": models.PurchaseRequisitionStatusConverted,
			},
		})
	}

	return uc.mapper.ToDetailResponse(full), nil
}

func (uc *purchaseOrderUsecase) CreateFromPurchaseRequisition(ctx context.Context, purchaseRequisitionID string) (*dto.PurchaseOrderDetailResponse, error) {
	purchaseRequisitionID = strings.TrimSpace(purchaseRequisitionID)
	if purchaseRequisitionID == "" {
		return nil, errors.New("purchase requisition id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var createdID string

	err := database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		var pr models.PurchaseRequisition
		if err := database.GetDB(ctx, tx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&pr, "id = ?", purchaseRequisitionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPurchaseRequisitionNotFound
			}
			return err
		}

		if pr.Status != models.PurchaseRequisitionStatusApproved {
			return ErrInvalidStatus
		}

		var count int64
		if err := database.GetDB(ctx, tx).Model(&models.PurchaseOrder{}).
			Where("purchase_requisition_id = ?", pr.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrPurchaseOrderConflict
		}

		po := &models.PurchaseOrder{
			Code:                  "",
			SupplierID:            pr.SupplierID,
			PaymentTermsID:        pr.PaymentTermsID,
			BusinessUnitID:        pr.BusinessUnitID,
			CreatedBy:             actorID,
			PurchaseRequisitionID: &pr.ID,
			SalesOrderID:          nil,
			OrderDate:             apptime.Now().Format("2006-01-02"),
			DueDate:               nil,
			Notes:                 pr.Notes,
			Status:                models.PurchaseOrderStatusDraft,
			TaxRate:               clamp(pr.TaxRate, 0, 100),
			DeliveryCost:          math.Max(0, pr.DeliveryCost),
			OtherCost:             math.Max(0, pr.OtherCost),
			Items:                 make([]models.PurchaseOrderItem, 0, len(pr.Items)),
		}
		companyID, fiscalYearID, err := uc.fiscalYearResolver.Resolve(ctx, po.OrderDate)
		if err != nil {
			return err
		}
		po.CompanyID = &companyID
		po.FiscalYearID = &fiscalYearID

		for _, it := range pr.Items {
			discount := clamp(it.Discount, 0, 100)
			qty := math.Max(0, it.Quantity)
			price := math.Max(0, it.PurchasePrice)
			subtotal := calcPOItemSubtotal(qty, price, discount)
			po.Items = append(po.Items, models.PurchaseOrderItem{
				ProductID: it.ProductID,
				Quantity:  qty,
				Price:     price,
				Discount:  discount,
				Subtotal:  subtotal,
				Notes:     it.Notes,
			})
		}

		sub, tax, total := calcPOTotals(po.Items, po.TaxRate, po.DeliveryCost, po.OtherCost)
		po.SubTotal = sub
		po.TaxAmount = tax
		po.TotalAmount = total

		if err := snapshotPurchaseOrderHeader(ctx, tx, po, nil); err != nil {
			return err
		}
		if err := snapshotPurchaseOrderItems(ctx, tx, po, nil); err != nil {
			return err
		}

		poRepoTx := repositories.NewPurchaseOrderRepository(tx)
		created, err := poRepoTx.Create(ctx, po)
		if err != nil {
			return err
		}
		createdID = created.ID

		prRepoTx := repositories.NewPurchaseRequisitionRepository(tx)
		if _, err := prRepoTx.UpdateStatus(ctx, pr.ID, models.PurchaseRequisitionStatusConverted, map[string]interface{}{
			"converted_to_purchase_order_id": createdID,
		}); err != nil {
			return err
		}

		full, err := poRepoTx.GetByID(ctx, createdID)
		if err == nil {
			uc.auditService.Log(ctx, "purchase_order.create", createdID, map[string]interface{}{
				"after": poAuditSnapshot(full),
			})
		}
		uc.auditService.Log(ctx, "purchase_requisition.convert", pr.ID, map[string]interface{}{
			"after": map[string]interface{}{
				"id":                pr.ID,
				"status":            models.PurchaseRequisitionStatusConverted,
				"purchase_order_id": createdID,
			},
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.GetByID(ctx, createdID)
	if err != nil {
		return nil, err
	}
	return uc.mapper.ToDetailResponse(full), nil
}

func (uc *purchaseOrderUsecase) Update(ctx context.Context, id string, req *dto.UpdatePurchaseOrderRequest) (*dto.PurchaseOrderDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	if existing.Status != models.PurchaseOrderStatusDraft {
		return nil, ErrPurchaseOrderConflict
	}
	before := poAuditSnapshot(existing)

	po := &models.PurchaseOrder{
		ID:                    existing.ID,
		Code:                  existing.Code,
		SupplierID:            req.SupplierID,
		PaymentTermsID:        req.PaymentTermsID,
		BusinessUnitID:        req.BusinessUnitID,
		CreatedBy:             existing.CreatedBy,
		PurchaseRequisitionID: existing.PurchaseRequisitionID,
		SalesOrderID:          existing.SalesOrderID,
		OrderDate:             req.OrderDate,
		DueDate:               req.DueDate,
		RevisionComment:       existing.RevisionComment,
		Notes:                 req.Notes,
		Status:                existing.Status,
		TaxRate:               clamp(req.TaxRate, 0, 100),
		DeliveryCost:          math.Max(0, req.DeliveryCost),
		OtherCost:             math.Max(0, req.OtherCost),
		Items:                 make([]models.PurchaseOrderItem, 0, len(req.Items)),
	}

	for _, it := range req.Items {
		discount := clamp(it.Discount, 0, 100)
		qty := math.Max(0, it.Quantity)
		price := math.Max(0, it.Price)
		subtotal := calcPOItemSubtotal(qty, price, discount)
		po.Items = append(po.Items, models.PurchaseOrderItem{
			ProductID: it.ProductID,
			Quantity:  qty,
			Price:     price,
			Discount:  discount,
			Subtotal:  subtotal,
			Notes:     it.Notes,
		})
	}

	sub, tax, total := calcPOTotals(po.Items, po.TaxRate, po.DeliveryCost, po.OtherCost)
	po.SubTotal = sub
	po.TaxAmount = tax
	po.TotalAmount = total

	if err := snapshotPurchaseOrderHeader(ctx, uc.db, po, existing); err != nil {
		return nil, err
	}
	if err := snapshotPurchaseOrderItems(ctx, uc.db, po, existing); err != nil {
		return nil, err
	}

	updated, err := uc.repo.Update(ctx, po)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_order.update", id, map[string]interface{}{
		"before": before,
		"after":  poAuditSnapshot(updated),
	})

	return uc.mapper.ToDetailResponse(updated), nil
}

func (uc *purchaseOrderUsecase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPurchaseOrderNotFound
		}
		return err
	}
	if existing.Status != models.PurchaseOrderStatusDraft {
		return ErrPurchaseOrderConflict
	}
	before := poAuditSnapshot(existing)
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditService.Log(ctx, "purchase_order.delete", id, map[string]interface{}{
		"before": before,
	})
	return nil
}

func (uc *purchaseOrderUsecase) Submit(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	if existing.Status != models.PurchaseOrderStatusDraft {
		return nil, ErrPurchaseOrderConflict
	}
	before := poAuditSnapshot(existing)
	now := apptime.Now()
	updated, err := uc.repo.UpdateStatusWithTimestamp(ctx, id, models.PurchaseOrderStatusSubmitted, map[string]interface{}{
		"submitted_at": now,
	})
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "purchase_order.submit", id, map[string]interface{}{
		"before": before,
		"after":  poAuditSnapshot(updated),
	})
	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "purchase_order.approve",
		EntityType:     "purchase_order",
		EntityID:       updated.ID,
		Title:          "Purchase Order Approval",
		Message:        "A purchase order has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create purchase order notification: %v", err)
	}
	return uc.mapper.ToDetailResponse(updated), nil
}

func (uc *purchaseOrderUsecase) Approve(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	if existing.Status != models.PurchaseOrderStatusSubmitted {
		return nil, ErrInvalidStatus
	}
	before := poAuditSnapshot(existing)
	now := apptime.Now()
	updated, err := uc.repo.UpdateStatusWithTimestamp(ctx, id, models.PurchaseOrderStatusApproved, map[string]interface{}{
		"approved_at": now,
	})
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "purchase_order.approve", id, map[string]interface{}{
		"before": before,
		"after":  poAuditSnapshot(updated),
	})
	return uc.mapper.ToDetailResponse(updated), nil
}

func (uc *purchaseOrderUsecase) Reject(ctx context.Context, id string) (*dto.PurchaseOrderDetailResponse, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	if existing.Status != models.PurchaseOrderStatusSubmitted {
		return nil, ErrInvalidStatus
	}
	before := poAuditSnapshot(existing)
	updated, err := uc.repo.UpdateStatusWithTimestamp(ctx, id, models.PurchaseOrderStatusRejected, map[string]interface{}{
		"submitted_at": nil,
	})
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "purchase_order.reject", id, map[string]interface{}{
		"before": before,
		"after":  poAuditSnapshot(updated),
	})
	return uc.mapper.ToDetailResponse(updated), nil
}

func (uc *purchaseOrderUsecase) AddData(ctx context.Context) (*dto.PurchaseOrderAddResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var suppliers []supplierModels.Supplier
	if err := database.GetDB(ctx, uc.db).
		Model(&supplierModels.Supplier{}).
		Preload("Contacts").
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierIDs := make([]string, 0, len(suppliers))
	for _, s := range suppliers {
		supplierIDs = append(supplierIDs, s.ID)
	}

	var products []productModels.Product
	if len(supplierIDs) > 0 {
		if err := database.GetDB(ctx, uc.db).
			Model(&productModels.Product{}).
			Where("supplier_id IN ?", supplierIDs).
			Where("supplier_id IS NOT NULL").
			Where("is_active = ?", true).
			Where("is_approved = ?", true).
			Order("name ASC").
			Find(&products).Error; err != nil {
			return nil, err
		}
	}

	productsBySupplier := make(map[string][]productModels.Product)
	for _, p := range products {
		if p.SupplierID == nil || strings.TrimSpace(*p.SupplierID) == "" {
			continue
		}
		productsBySupplier[*p.SupplierID] = append(productsBySupplier[*p.SupplierID], p)
	}

	// Payment terms
	var paymentTerms []coreModels.PaymentTerms
	if err := database.GetDB(ctx, uc.db).
		Model(&coreModels.PaymentTerms{}).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&paymentTerms).Error; err != nil {
		return nil, err
	}

	// Business units
	var businessUnits []orgModels.BusinessUnit
	if err := database.GetDB(ctx, uc.db).
		Model(&orgModels.BusinessUnit{}).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&businessUnits).Error; err != nil {
		return nil, err
	}

	// Build response DTO (group products under suppliers for UI convenience)
	respSuppliers := make([]dto.PurchaseOrderAddSupplier, 0, len(suppliers))
	for _, s := range suppliers {
		prods := productsBySupplier[s.ID]
		respProducts := make([]dto.PurchaseOrderAddProduct, 0, len(prods))
		respPhones := make([]dto.PurchaseOrderAddSupplierContact, 0, len(s.Contacts))
		for _, p := range prods {
			respProducts = append(respProducts, dto.PurchaseOrderAddProduct{
				ID:         p.ID,
				Code:       p.Code,
				Name:       p.Name,
				Stock:      p.CurrentStock,
				CurrentHpp: p.CurrentHpp,
				SupplierID: p.SupplierID,
				IsActive:   p.IsActive,
				IsApproved: p.IsApproved,
			})
		}
		for _, ph := range s.Contacts {
			respPhones = append(respPhones, dto.PurchaseOrderAddSupplierContact{
				ID:          ph.ID,
				PhoneNumber: ph.Phone,
				Label:       ph.Position,
				IsPrimary:   ph.IsPrimary,
			})
		}
		respSuppliers = append(respSuppliers, dto.PurchaseOrderAddSupplier{
			ID:             s.ID,
			Code:           s.Code,
			Name:           s.Name,
			PaymentTermsID: s.PaymentTermsID,
			BusinessUnitID: s.BusinessUnitID,
			Contacts:       respPhones,
			Products:       respProducts,
		})
	}

	respPaymentTerms := make([]dto.PurchaseOrderAddPaymentTerms, 0, len(paymentTerms))
	for _, pt := range paymentTerms {
		respPaymentTerms = append(respPaymentTerms, dto.PurchaseOrderAddPaymentTerms{
			ID:   pt.ID,
			Code: pt.Code,
			Name: pt.Name,
			Days: pt.Days,
		})
	}

	respBusinessUnits := make([]dto.PurchaseOrderAddBusinessUnit, 0, len(businessUnits))
	for _, bu := range businessUnits {
		respBusinessUnits = append(respBusinessUnits, dto.PurchaseOrderAddBusinessUnit{
			ID:   bu.ID,
			Name: bu.Name,
		})
	}

	return &dto.PurchaseOrderAddResponse{
		Suppliers:     respSuppliers,
		PaymentTerms:  respPaymentTerms,
		BusinessUnits: respBusinessUnits,
	}, nil
}

func (uc *purchaseOrderUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.PurchaseOrderAuditTrailEntry, int64, error) {
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

	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.PurchaseOrder{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, 0, ErrPurchaseOrderNotFound
	}

	tx := uc.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("audit_logs.permission_code LIKE ?", "purchase_order.%")

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

	entries := make([]dto.PurchaseOrderAuditTrailEntry, 0, len(rows))
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
		entries = append(entries, dto.PurchaseOrderAuditTrailEntry{
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

func poAuditSnapshot(po *models.PurchaseOrder) map[string]interface{} {
	if po == nil {
		return nil
	}
	return map[string]interface{}{
		"id":                       po.ID,
		"code":                     po.Code,
		"status":                   po.Status,
		"supplier_id":              po.SupplierID,
		"payment_terms_id":         po.PaymentTermsID,
		"payment_terms_name":       po.PaymentTermsNameSnapshot,
		"business_unit_id":         po.BusinessUnitID,
		"business_unit_name":       po.BusinessUnitNameSnapshot,
		"created_by":               po.CreatedBy,
		"purchase_requisitions_id": po.PurchaseRequisitionID,
		"sales_order_id":           po.SalesOrderID,
		"order_date":               po.OrderDate,
		"due_date":                 po.DueDate,
		"tax_rate":                 po.TaxRate,
		"tax_amount":               po.TaxAmount,
		"delivery_cost":            po.DeliveryCost,
		"other_cost":               po.OtherCost,
		"sub_total":                po.SubTotal,
		"total_amount":             po.TotalAmount,
		"revision_comment":         po.RevisionComment,
		"items":                    poAuditItems(po.Items),
	}
}

func poAuditItems(items []models.PurchaseOrderItem) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"product_id":   item.ProductID,
			"product_code": item.ProductCodeSnapshot,
			"product_name": item.ProductNameSnapshot,
			"quantity":     item.Quantity,
			"price":        item.Price,
			"discount":     item.Discount,
		})
	}

	return out
}

func calcPOItemSubtotal(qty, price, discount float64) float64 {
	raw := qty * price
	if discount <= 0 {
		return roundTo2Decimals(raw)
	}
	return roundTo2Decimals(raw - (raw * (discount / 100)))
}

func calcPOTotals(items []models.PurchaseOrderItem, taxRate, deliveryCost, otherCost float64) (subTotal, taxAmount, total float64) {
	subTotal = 0
	for _, it := range items {
		subTotal += it.Subtotal
	}
	subTotal = math.Round(subTotal*100) / 100
	if taxRate > 0 {
		taxAmount = math.Round(subTotal*(taxRate/100)*100) / 100
	}
	// delivery/other are assumed pre-clamped
	total = math.Round((subTotal+taxAmount+deliveryCost+otherCost)*100) / 100
	return
}
