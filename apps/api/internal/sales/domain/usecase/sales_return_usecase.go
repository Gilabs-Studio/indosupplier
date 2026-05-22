package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	invDto "github.com/gilabs/gims/api/internal/inventory/domain/dto"
	invUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"gorm.io/gorm"
)

var (
	ErrSalesReturnNotFound = errors.New("sales return not found")
	ErrSalesReturnInvalid  = errors.New("invalid sales return request")
)

const errDBNil = "db is nil"

const salesReturnWarehouseOverridePermission = "sales_return.warehouse_override"

type salesReturnCreateContext struct {
	DeliveryID  string
	WarehouseID string
	CustomerID  string
	InvoiceID   *string
	Action      models.SalesReturnAction
}

type SalesReturnUsecase interface {
	GetFormData(ctx context.Context) (*dto.SalesReturnFormDataResponse, error)
	List(ctx context.Context, params repositories.SalesReturnListParams) ([]*dto.SalesReturnResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.SalesReturnResponse, error)
	Create(ctx context.Context, req *dto.CreateSalesReturnRequest) (*dto.SalesReturnResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSalesReturnRequest) (*dto.SalesReturnResponse, error)
	UpdateStatus(ctx context.Context, id string, status string) (*dto.SalesReturnResponse, error)
	Delete(ctx context.Context, id string) error
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
	TriggerJournalForReturn(ctx context.Context, ret *models.SalesReturn) error
}

type salesReturnUsecase struct {
	db           *gorm.DB
	repo         repositories.SalesReturnRepository
	invUC        invUsecase.InventoryUsecase
	journalUC    finUsecase.JournalEntryUsecase
	coaUC        finUsecase.ChartOfAccountUsecase
	auditService audit.AuditService
	engine       accounting.AccountingEngine
}

func NewSalesReturnUsecase(
	db *gorm.DB,
	repo repositories.SalesReturnRepository,
	invUC invUsecase.InventoryUsecase,
	journalUC finUsecase.JournalEntryUsecase,
	coaUC finUsecase.ChartOfAccountUsecase,
	auditService audit.AuditService,
	engine accounting.AccountingEngine,
) SalesReturnUsecase {
	return &salesReturnUsecase{
		db:           db,
		repo:         repo,
		invUC:        invUC,
		journalUC:    journalUC,
		coaUC:        coaUC,
		auditService: auditService,
		engine:       engine,
	}
}

func (u *salesReturnUsecase) GetFormData(ctx context.Context) (*dto.SalesReturnFormDataResponse, error) {
	if u.db == nil {
		return nil, errors.New(errDBNil)
	}

	var warehouses []warehouseModels.Warehouse
	if err := database.GetDB(ctx, u.db).
		Model(&warehouseModels.Warehouse{}).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&warehouses).Error; err != nil {
		return nil, err
	}

	warehouseOptions := make([]dto.ReturnWarehouseOption, 0, len(warehouses))
	for _, wh := range warehouses {
		warehouseOptions = append(warehouseOptions, dto.ReturnWarehouseOption{
			ID:   wh.ID,
			Name: wh.Name,
		})
	}

	return &dto.SalesReturnFormDataResponse{
		Warehouses: warehouseOptions,
		ReturnReasons: []dto.ReturnOption{
			{Value: "DAMAGED", Label: "Damaged item"},
			{Value: "WRONG_ITEM", Label: "Wrong item"},
			{Value: "EXPIRED", Label: "Expired item"},
			{Value: "CUSTOMER_REQUEST", Label: "Customer request"},
			{Value: "OTHER", Label: "Other"},
		},
		ItemConditions: []dto.ReturnOption{
			{Value: "GOOD", Label: "Good"},
			{Value: "DAMAGED", Label: "Damaged"},
			{Value: "EXPIRED", Label: "Expired"},
			{Value: "OPENED", Label: "Opened"},
		},
		Actions: []dto.ReturnOption{
			{Value: string(models.SalesReturnActionRefund), Label: "Refund"},
			{Value: string(models.SalesReturnActionCreditNote), Label: "Credit Note"},
			{Value: string(models.SalesReturnActionReplacement), Label: "Replacement"},
		},
		RefundMethods: []dto.ReturnOption{
			{Value: "BANK_TRANSFER", Label: "Bank Transfer"},
			{Value: "CASH", Label: "Cash"},
			{Value: "WALLET", Label: "Wallet"},
		},
	}, nil
}

func (u *salesReturnUsecase) List(ctx context.Context, params repositories.SalesReturnListParams) ([]*dto.SalesReturnResponse, int64, error) {
	rows, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*dto.SalesReturnResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSalesReturnRow(row))
	}

	return out, total, nil
}

func (u *salesReturnUsecase) GetByID(ctx context.Context, id string) (*dto.SalesReturnResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesReturn{}, id, security.SalesScopeQueryOptions()) {
		return nil, ErrSalesReturnNotFound
	}
	row, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSalesReturnNotFound
		}
		return nil, err
	}

	return mapSalesReturnRow(row), nil
}

func (u *salesReturnUsecase) Create(ctx context.Context, req *dto.CreateSalesReturnRequest) (*dto.SalesReturnResponse, error) {
	if req == nil || len(req.Items) == 0 {
		return nil, ErrSalesReturnInvalid
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if u.db == nil {
		return nil, errors.New(errDBNil)
	}

	createCtx, err := u.prepareCreateContext(ctx, req)
	if err != nil {
		return nil, err
	}

	now := apptime.Now()
	code := fmt.Sprintf("SR-%s", now.Format("20060102-150405"))

	row := &models.SalesReturn{
		Code:        code,
		InvoiceID:   createCtx.InvoiceID,
		DeliveryID:  &createCtx.DeliveryID,
		WarehouseID: createCtx.WarehouseID,
		CustomerID:  createCtx.CustomerID,
		Reason:      strings.TrimSpace(req.Reason),
		Action:      createCtx.Action,
		Status:      models.SalesReturnStatusDraft,
		Notes:       req.Notes,
		CreatedBy:   actorID,
	}

	items := make([]models.SalesReturnItem, 0, len(req.Items))
	totalAmount := 0.0
	for _, item := range req.Items {
		subtotal := item.Qty * item.UnitPrice
		totalAmount += subtotal

		items = append(items, models.SalesReturnItem{
			InvoiceItemID: normalizeOptionalString(item.InvoiceItemID),
			ProductID:     strings.TrimSpace(item.ProductID),
			UOMID:         normalizeOptionalString(item.UOMID),
			Condition:     strings.ToUpper(strings.TrimSpace(item.Condition)),
			Notes:         normalizeOptionalString(item.Notes),
			Quantity:      item.Qty,
			UnitPrice:     item.UnitPrice,
			Subtotal:      subtotal,
		})
	}
	row.TotalAmount = totalAmount
	row.Items = items

	if err := u.repo.Create(ctx, row); err != nil {
		return nil, err
	}

	created, err := u.repo.GetByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	logSalesAudit(u.auditService, ctx, "sales_return.create", created.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":         created.Code,
			"status":       created.Status,
			"action":       created.Action,
			"total_amount": created.TotalAmount,
			"delivery_id":  created.DeliveryID,
		},
	})

	return mapSalesReturnRow(created), nil
}

func (u *salesReturnUsecase) Update(ctx context.Context, id string, req *dto.UpdateSalesReturnRequest) (*dto.SalesReturnResponse, error) {
	if req == nil || len(req.Items) == 0 {
		return nil, ErrSalesReturnInvalid
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSalesReturnNotFound
		}
		return nil, err
	}

	if existing.Status != models.SalesReturnStatusDraft {
		return nil, ErrSalesReturnInvalid
	}

	action, err := normalizeSalesReturnAction(req.Action)
	if err != nil {
		return nil, err
	}

	deliveryID := existing.DeliveryID
	if deliveryID == nil || strings.TrimSpace(*deliveryID) == "" {
		return nil, errors.New("delivery_id is required")
	}

	delivery, err := u.getDeliveryOrder(ctx, *deliveryID)
	if err != nil {
		return nil, err
	}

	warehouseID, err := u.resolveWarehouseForCreate(ctx, delivery, req.WarehouseID, *deliveryID)
	if err != nil {
		return nil, err
	}

	customerID := strings.TrimSpace(req.CustomerID)
	if customerID == "" {
		customerID = existing.CustomerID
	}

	if err := u.validateRequestedQtyForUpdate(ctx, *deliveryID, req.Items, existing.Items); err != nil {
		return nil, err
	}

	items := make([]models.SalesReturnItem, 0, len(req.Items))
	totalAmount := 0.0
	for _, item := range req.Items {
		subtotal := item.Qty * item.UnitPrice
		totalAmount += subtotal

		items = append(items, models.SalesReturnItem{
			InvoiceItemID: normalizeOptionalString(item.InvoiceItemID),
			ProductID:     strings.TrimSpace(item.ProductID),
			UOMID:         normalizeOptionalString(item.UOMID),
			Condition:     strings.ToUpper(strings.TrimSpace(item.Condition)),
			Notes:         normalizeOptionalString(item.Notes),
			Quantity:      item.Qty,
			UnitPrice:     item.UnitPrice,
			Subtotal:      subtotal,
		})
	}

	row := &models.SalesReturn{
		InvoiceID:   existing.InvoiceID,
		DeliveryID:  existing.DeliveryID,
		WarehouseID: warehouseID,
		CustomerID:  customerID,
		Reason:      strings.TrimSpace(req.Reason),
		Action:      action,
		Notes:       req.Notes,
		TotalAmount: totalAmount,
		Items:       items,
	}

	if err := u.repo.Update(ctx, id, row); err != nil {
		return nil, err
	}

	updated, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	logSalesAudit(u.auditService, ctx, "sales_return.update", updated.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":         updated.Code,
			"status":       updated.Status,
			"action":       updated.Action,
			"total_amount": updated.TotalAmount,
			"warehouse_id": updated.WarehouseID,
		},
	})

	return mapSalesReturnRow(updated), nil
}

func (u *salesReturnUsecase) UpdateStatus(ctx context.Context, id string, status string) (*dto.SalesReturnResponse, error) {
	row, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSalesReturnNotFound
		}
		return nil, err
	}

	nextStatus, err := normalizeSalesReturnStatus(status)
	if err != nil {
		return nil, ErrSalesReturnInvalid
	}
	previousStatus := row.Status

	if !canTransitionSalesReturnStatus(row.Status, nextStatus) {
		return nil, ErrSalesReturnInvalid
	}

	if err := u.repo.UpdateStatus(ctx, id, nextStatus); err != nil {
		return nil, err
	}

	if nextStatus == models.SalesReturnStatusProcessed {
		actorID, _ := ctx.Value("user_id").(string)
		actorID = strings.TrimSpace(actorID)
		if err := u.createStockMovementsFromRows(ctx, row.Items, row.WarehouseID, row.Code, actorID); err != nil {
			return nil, err
		}
		// Trigger journal entry
		if err := u.TriggerJournalForReturn(ctx, row); err != nil {
			fmt.Printf("⚠️ Failed to trigger journal for sales return %s: %v\n", id, err)
		}
	}

	updated, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	logSalesAudit(u.auditService, ctx, "sales_return.status_change", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  updated.Status,
	})

	return mapSalesReturnRow(updated), nil
}

func (u *salesReturnUsecase) Delete(ctx context.Context, id string) error {
	row, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrSalesReturnNotFound
		}
		return err
	}

	if row.Status != models.SalesReturnStatusDraft && row.Status != models.SalesReturnStatusRejected {
		return ErrSalesReturnInvalid
	}

	err = u.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	logSalesAudit(u.auditService, ctx, "sales_return.delete", id, map[string]interface{}{
		"before": map[string]interface{}{
			"code":         row.Code,
			"status":       row.Status,
			"action":       row.Action,
			"total_amount": row.TotalAmount,
		},
	})

	return nil
}

func (u *salesReturnUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, errors.New(errDBNil)
	}
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesReturn{}, id, security.SalesScopeQueryOptions()) {
		return nil, 0, ErrSalesReturnNotFound
	}

	return listAuditTrailEntries(ctx, u.db, id, "sales_return.", page, perPage)
}

func (u *salesReturnUsecase) prepareCreateContext(ctx context.Context, req *dto.CreateSalesReturnRequest) (*salesReturnCreateContext, error) {
	deliveryID, err := resolveSalesReturnDeliveryID(req)
	if err != nil {
		return nil, err
	}

	action, err := normalizeSalesReturnAction(req.Action)
	if err != nil {
		return nil, err
	}

	delivery, err := u.getDeliveryOrder(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	invoiceID, err := u.resolveInvoiceIDForDelivery(ctx, deliveryID, req.InvoiceID)
	if err != nil {
		return nil, err
	}

	warehouseID, err := u.resolveWarehouseForCreate(ctx, delivery, req.WarehouseID, deliveryID)
	if err != nil {
		return nil, err
	}

	customerID, err := resolveSalesReturnCustomerID(delivery, req.CustomerID)
	if err != nil {
		return nil, err
	}

	if err := u.validateRequestedQty(ctx, deliveryID, req.Items); err != nil {
		return nil, err
	}

	return &salesReturnCreateContext{
		DeliveryID:  deliveryID,
		WarehouseID: warehouseID,
		CustomerID:  customerID,
		InvoiceID:   invoiceID,
		Action:      action,
	}, nil
}

func resolveSalesReturnDeliveryID(req *dto.CreateSalesReturnRequest) (string, error) {
	if req == nil || req.DeliveryID == nil {
		return "", errors.New("delivery_id is required")
	}

	deliveryID := strings.TrimSpace(*req.DeliveryID)
	if deliveryID == "" {
		return "", errors.New("delivery_id is required")
	}

	return deliveryID, nil
}

func (u *salesReturnUsecase) resolveWarehouseForCreate(
	ctx context.Context,
	delivery *models.DeliveryOrder,
	requestedWarehouseID string,
	deliveryID string,
) (string, error) {
	warehouseID := strings.TrimSpace(requestedWarehouseID)
	if warehouseID == "" {
		if delivery.WarehouseID == nil || strings.TrimSpace(*delivery.WarehouseID) == "" {
			return "", errors.New("warehouse_id is required")
		}
		warehouseID = strings.TrimSpace(*delivery.WarehouseID)
	}

	if delivery.WarehouseID == nil {
		return warehouseID, nil
	}

	sourceWarehouseID := strings.TrimSpace(*delivery.WarehouseID)
	if sourceWarehouseID == "" || warehouseID == sourceWarehouseID {
		return warehouseID, nil
	}

	if !hasWarehouseOverridePermission(ctx, salesReturnWarehouseOverridePermission) {
		u.logWarehouseMismatchAttempt(ctx, deliveryID, sourceWarehouseID, warehouseID, "denied")
		return "", errors.New("warehouse_id must match delivery warehouse")
	}

	u.logWarehouseMismatchAttempt(ctx, deliveryID, sourceWarehouseID, warehouseID, "allowed")
	return warehouseID, nil
}

func resolveSalesReturnCustomerID(delivery *models.DeliveryOrder, requestedCustomerID string) (string, error) {
	customerID := strings.TrimSpace(requestedCustomerID)
	if customerID != "" {
		return customerID, nil
	}

	if delivery == nil || delivery.SalesOrder == nil || delivery.SalesOrder.CustomerID == nil {
		return "", errors.New("customer_id is required")
	}

	customerID = strings.TrimSpace(*delivery.SalesOrder.CustomerID)
	if customerID == "" {
		return "", errors.New("customer_id is required")
	}

	return customerID, nil
}

func (u *salesReturnUsecase) logWarehouseMismatchAttempt(
	ctx context.Context,
	deliveryID string,
	sourceWarehouseID string,
	requestedWarehouseID string,
	result string,
) {
	if u.auditService == nil {
		return
	}

	u.auditService.Log(ctx, "sales_return.warehouse_override", deliveryID, map[string]interface{}{
		"delivery_id":              deliveryID,
		"source_warehouse_id":      sourceWarehouseID,
		"requested_warehouse_id":   requestedWarehouseID,
		"required_permission_code": salesReturnWarehouseOverridePermission,
		"result":                   result,
	})
}

func hasWarehouseOverridePermission(ctx context.Context, permissionCode string) bool {
	if strings.EqualFold(strings.TrimSpace(getContextString(ctx, "user_role")), "admin") {
		return true
	}

	if permissions, ok := ctx.Value("user_permissions").(map[string]bool); ok {
		return permissions[permissionCode]
	}

	if scopedPermissions, ok := ctx.Value("user_permissions_scope").(map[string]string); ok {
		_, exists := scopedPermissions[permissionCode]
		return exists
	}

	return false
}

func getContextString(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(key).(string)
	return strings.TrimSpace(v)
}

func normalizeSalesReturnAction(raw string) (models.SalesReturnAction, error) {
	action := strings.ToUpper(strings.TrimSpace(raw))
	switch action {
	case string(models.SalesReturnActionRefund), string(models.SalesReturnActionCreditNote), string(models.SalesReturnActionReplacement):
		return models.SalesReturnAction(action), nil
	default:
		return "", ErrSalesReturnInvalid
	}
}

func normalizeSalesReturnStatus(raw string) (models.SalesReturnStatus, error) {
	status := strings.ToUpper(strings.TrimSpace(raw))
	switch status {
	case string(models.SalesReturnStatusDraft), string(models.SalesReturnStatusSubmitted), string(models.SalesReturnStatusProcessed), string(models.SalesReturnStatusRejected):
		return models.SalesReturnStatus(status), nil
	default:
		return "", ErrSalesReturnInvalid
	}
}

func canTransitionSalesReturnStatus(current, next models.SalesReturnStatus) bool {
	if current == next {
		return true
	}

	switch current {
	case models.SalesReturnStatusDraft:
		return next == models.SalesReturnStatusSubmitted || next == models.SalesReturnStatusRejected
	case models.SalesReturnStatusSubmitted:
		return next == models.SalesReturnStatusProcessed || next == models.SalesReturnStatusRejected
	case models.SalesReturnStatusRejected:
		return next == models.SalesReturnStatusDraft
	default:
		return false
	}
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (u *salesReturnUsecase) validateRequestedQty(
	ctx context.Context,
	deliveryID string,
	items []dto.CreateSalesReturnItemRequest,
) error {
	availableQtyByProduct, err := u.getAvailableDeliveryQtyByProduct(ctx, deliveryID)
	if err != nil {
		return err
	}

	requestQtyByProduct := make(map[string]float64)
	for _, item := range items {
		productID := strings.TrimSpace(item.ProductID)
		if _, ok := availableQtyByProduct[productID]; !ok {
			return errors.New("product not found in delivery order")
		}
		requestQtyByProduct[productID] += item.Qty
	}

	for productID, requestedQty := range requestQtyByProduct {
		availableQty := availableQtyByProduct[productID]
		if requestedQty <= 0 || requestedQty > availableQty+0.0001 {
			return errors.New("return quantity exceeds available quantity from delivery order")
		}
	}

	return nil
}

func (u *salesReturnUsecase) validateRequestedQtyForUpdate(
	ctx context.Context,
	deliveryID string,
	requestedItems []dto.UpdateSalesReturnItemRequest,
	existingItems []models.SalesReturnItem,
) error {
	availableQtyByProduct, err := u.getAvailableDeliveryQtyByProduct(ctx, deliveryID)
	if err != nil {
		return err
	}

	// Add back existing return quantities to the available pool
	for _, item := range existingItems {
		productID := strings.TrimSpace(item.ProductID)
		availableQtyByProduct[productID] += item.Quantity
	}

	requestQtyByProductUpdate := make(map[string]float64)
	for _, item := range requestedItems {
		productID := strings.TrimSpace(item.ProductID)
		if _, ok := availableQtyByProduct[productID]; !ok {
			return errors.New("product not found in delivery order")
		}
		requestQtyByProductUpdate[productID] += item.Qty
	}

	for productID, requestedQty := range requestQtyByProductUpdate {
		availableQty := availableQtyByProduct[productID]
		if requestedQty <= 0 || requestedQty > availableQty+0.0001 {
			return errors.New("return quantity exceeds available quantity from delivery order")
		}
	}

	return nil
}

func (u *salesReturnUsecase) createStockMovements(
	ctx context.Context,
	items []dto.CreateSalesReturnItemRequest,
	warehouseID string,
	referenceNumber string,
	actorID string,
) error {
	if u.invUC == nil {
		return nil
	}

	for _, item := range items {
		moveReq := &invDto.CreateManualMovementRequest{
			ProductID:       strings.TrimSpace(item.ProductID),
			WarehouseID:     warehouseID,
			Type:            "IN",
			Quantity:        item.Qty,
			ReferenceNumber: referenceNumber,
			Description:     "Sales return stock adjustment",
			CreatedBy:       actorID,
		}
		if err := u.invUC.CreateManualStockMovement(ctx, moveReq); err != nil {
			return err
		}
	}

	return nil
}

func (u *salesReturnUsecase) createStockMovementsFromRows(
	ctx context.Context,
	items []models.SalesReturnItem,
	warehouseID string,
	referenceNumber string,
	actorID string,
) error {
	if u.invUC == nil {
		return nil
	}

	for _, item := range items {
		moveReq := &invDto.CreateManualMovementRequest{
			ProductID:       strings.TrimSpace(item.ProductID),
			WarehouseID:     warehouseID,
			Type:            "IN",
			Quantity:        item.Quantity,
			ReferenceNumber: referenceNumber,
			Description:     "Sales return stock adjustment",
			CreatedBy:       actorID,
		}
		if err := u.invUC.CreateManualStockMovement(ctx, moveReq); err != nil {
			return err
		}
	}

	return nil
}

func (u *salesReturnUsecase) getDeliveryOrder(ctx context.Context, deliveryID string) (*models.DeliveryOrder, error) {
	if u.db == nil {
		return nil, errors.New(errDBNil)
	}

	var delivery models.DeliveryOrder
	if err := database.GetDB(ctx, u.db).
		Preload("SalesOrder").
		Preload("Items").
		First(&delivery, "id = ?", deliveryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("delivery order not found")
		}
		return nil, err
	}

	return &delivery, nil
}

func (u *salesReturnUsecase) resolveInvoiceIDForDelivery(ctx context.Context, deliveryID string, invoiceID *string) (*string, error) {
	trimmed := ""
	if invoiceID != nil {
		trimmed = strings.TrimSpace(*invoiceID)
	}
	if trimmed != "" {
		return &trimmed, nil
	}

	if u.db == nil {
		return nil, errors.New(errDBNil)
	}

	var invoice models.CustomerInvoice
	err := database.GetDB(ctx, u.db).
		Model(&models.CustomerInvoice{}).
		Where("delivery_order_id = ?", deliveryID).
		Order("created_at DESC").
		First(&invoice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	resolved := strings.TrimSpace(invoice.ID)
	if resolved == "" {
		return nil, nil
	}

	return &resolved, nil
}

func (u *salesReturnUsecase) getAvailableDeliveryQtyByProduct(ctx context.Context, deliveryID string) (map[string]float64, error) {
	if u.db == nil {
		return nil, errors.New(errDBNil)
	}

	type productQtyRow struct {
		ProductID string
		Qty       float64
	}

	sourceRows := make([]productQtyRow, 0)
	if err := database.GetDB(ctx, u.db).
		Model(&models.DeliveryOrderItem{}).
		Select("product_id, COALESCE(SUM(quantity), 0) AS qty").
		Where("delivery_order_id = ?", deliveryID).
		Group("product_id").
		Scan(&sourceRows).Error; err != nil {
		return nil, err
	}

	returnedRows := make([]productQtyRow, 0)
	returnedQuery := u.db.WithContext(ctx).
		Model(&models.SalesReturnItem{})
	returnedQuery, err := applyTenantJoinScope(ctx, returnedQuery, "sales_return_items.tenant_id", "sales_returns.tenant_id")
	if err != nil {
		return nil, err
	}

	if err := returnedQuery.
		Select("sales_return_items.product_id, COALESCE(SUM(sales_return_items.quantity), 0) AS qty").
		Joins("JOIN sales_returns ON sales_returns.id = sales_return_items.sales_return_id").
		Where("sales_returns.delivery_id = ?", deliveryID).
		Where("sales_returns.deleted_at IS NULL").
		Where("sales_returns.status <> ?", models.SalesReturnStatusRejected).
		Group("sales_return_items.product_id").
		Scan(&returnedRows).Error; err != nil {
		return nil, err
	}

	returnedQtyByProduct := make(map[string]float64)
	for _, row := range returnedRows {
		returnedQtyByProduct[row.ProductID] = row.Qty
	}

	availableByProduct := make(map[string]float64)
	for _, row := range sourceRows {
		availableByProduct[row.ProductID] = math.Max(0, row.Qty-returnedQtyByProduct[row.ProductID])
	}

	return availableByProduct, nil
}

func (u *salesReturnUsecase) TriggerJournalForReturn(ctx context.Context, ret *models.SalesReturn) error {
	if ret == nil || u.journalUC == nil || u.engine == nil {
		return nil
	}

	if ret.TotalAmount <= 0 {
		return nil
	}

	data := accounting.TransactionData{
		ReferenceType:   "SALES_RETURN",
		ReferenceID:     ret.ID,
		EntryDate:       apptime.Now().Format("2006-01-02"),
		Description:     fmt.Sprintf("Sales Return %s", ret.Code),
		TotalAmount:     ret.TotalAmount,
		SubTotal:        ret.TotalAmount, // Assuming no tax split on return for now
		DescriptionArgs: []interface{}{ret.Code},
	}

	req, err := u.engine.GenerateJournal(ctx, accounting.ProfileSalesReturn, data)
	if err != nil {
		return fmt.Errorf("failed to generate sales return journal: %w", err)
	}

	// Balance check
	var debitTotal, creditTotal float64
	for _, l := range req.Lines {
		debitTotal += l.Debit
		creditTotal += l.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.001 {
		return fmt.Errorf("generated sales return journal is unbalanced: debit=%.2f credit=%.2f", debitTotal, creditTotal)
	}

	req.IsSystemGenerated = true
	_, err = u.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post sales return journal: %w", err)
	}

	log.Printf("journal_observability event=trigger.success module=sales_return reference_id=%s", ret.ID)
	return nil
}

func mapSalesReturnRow(row *models.SalesReturn) *dto.SalesReturnResponse {
	items := make([]dto.SalesReturnItemResponse, 0, len(row.Items))
	for _, item := range row.Items {
		items = append(items, dto.SalesReturnItemResponse{
			ID:            item.ID,
			InvoiceItemID: item.InvoiceItemID,
			ProductID:     item.ProductID,
			UOMID:         item.UOMID,
			Condition:     item.Condition,
			Notes:         item.Notes,
			Qty:           item.Quantity,
			UnitPrice:     item.UnitPrice,
			Subtotal:      item.Subtotal,
		})
	}

	return &dto.SalesReturnResponse{
		ID:                row.ID,
		Code:              row.Code,
		InvoiceID:         row.InvoiceID,
		DeliveryID:        row.DeliveryID,
		WarehouseID:       row.WarehouseID,
		CustomerID:        row.CustomerID,
		Reason:            row.Reason,
		Action:            string(row.Action),
		Status:            string(row.Status),
		Notes:             row.Notes,
		TotalAmount:       row.TotalAmount,
		StockAdjustmentID: row.StockAdjustmentID,
		CreditNoteID:      row.CreditNoteID,
		Items:             items,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
