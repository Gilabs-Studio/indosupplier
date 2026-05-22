package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	inventoryDto "github.com/gilabs/gims/api/internal/inventory/domain/dto"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	salesOrderRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/mapper"
	salesService "github.com/gilabs/gims/api/internal/sales/domain/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrDeliveryOrderNotFound           = errors.New("delivery order not found")
	ErrDeliveryOrderAlreadyExists      = errors.New("delivery order with this code already exists")
	ErrDeliveryOrderAlreadyPosted      = errors.New("delivery order has already been posted")
	ErrInvalidDeliveryStatusTransition = errors.New("invalid delivery status transition")
	ErrDeliveryProductNotFound         = errors.New("product not found in delivery")
	ErrInvalidDeliveryOrderStatus      = errors.New("cannot modify delivery order in current status")
	ErrDeliverySalesOrderNotFound      = errors.New("sales order not found for delivery")
	ErrInsufficientBatchStock          = errors.New("insufficient stock in selected batch")
	ErrBatchNotFound                   = errors.New("inventory batch not found")
)

const (
	errDeliveryWarehouseIDRequired = "warehouse_id is required"
	errDeliveryDBNil               = "db is nil"
)

// DeliveryOrderUsecase defines the interface for delivery order business logic
type DeliveryOrderUsecase interface {
	List(ctx context.Context, req *dto.ListDeliveryOrdersRequest) ([]dto.DeliveryOrderResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.DeliveryOrderResponse, error)
	ListItems(ctx context.Context, deliveryOrderID string, req *dto.ListDeliveryOrderItemsRequest) ([]dto.DeliveryOrderItemResponse, *utils.PaginationResult, error)
	Create(ctx context.Context, req *dto.CreateDeliveryOrderRequest, createdBy *string) (*dto.DeliveryOrderResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateDeliveryOrderRequest) (*dto.DeliveryOrderResponse, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, req *dto.UpdateDeliveryOrderStatusRequest, userID *string) (*dto.DeliveryOrderResponse, error)
	Ship(ctx context.Context, id string, req *dto.ShipDeliveryOrderRequest, userID *string) (*dto.DeliveryOrderResponse, error)
	Deliver(ctx context.Context, id string, req *dto.DeliverDeliveryOrderRequest, userID *string) (*dto.DeliveryOrderResponse, error)
	SelectBatches(ctx context.Context, req *dto.BatchSelectionRequest) (*dto.BatchSelectionResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
}

type deliveryOrderUsecase struct {
	db                *gorm.DB
	deliveryOrderRepo salesRepos.DeliveryOrderRepository
	salesOrderRepo    salesOrderRepos.SalesOrderRepository
	productRepo       productRepos.ProductRepository
	inventoryUC       inventoryUsecase.InventoryUsecase
	auditService      audit.AuditService
	salesJournalSvc   salesService.SalesJournalService
}

// NewDeliveryOrderUsecase creates a new DeliveryOrderUsecase
func NewDeliveryOrderUsecase(
	db *gorm.DB,
	deliveryOrderRepo salesRepos.DeliveryOrderRepository,
	salesOrderRepo salesOrderRepos.SalesOrderRepository,
	productRepo productRepos.ProductRepository,
	inventoryUC inventoryUsecase.InventoryUsecase,
	auditService audit.AuditService,
	salesJournalSvc salesService.SalesJournalService,
) DeliveryOrderUsecase {
	if salesJournalSvc == nil {
		salesJournalSvc = salesService.NewSalesJournalService(db, nil, nil)
	}

	return &deliveryOrderUsecase{
		db:                db,
		deliveryOrderRepo: deliveryOrderRepo,
		salesOrderRepo:    salesOrderRepo,
		productRepo:       productRepo,
		inventoryUC:       inventoryUC,
		auditService:      auditService,
		salesJournalSvc:   salesJournalSvc,
	}
}

func (u *deliveryOrderUsecase) List(ctx context.Context, req *dto.ListDeliveryOrdersRequest) ([]dto.DeliveryOrderResponse, *utils.PaginationResult, error) {
	deliveryOrders, total, err := u.deliveryOrderRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.DeliveryOrderResponse, len(deliveryOrders))
	for i := range deliveryOrders {
		responses[i] = mapper.ToDeliveryOrderResponse(&deliveryOrders[i])
	}

	// Calculate pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *deliveryOrderUsecase) ListItems(ctx context.Context, deliveryOrderID string, req *dto.ListDeliveryOrderItemsRequest) ([]dto.DeliveryOrderItemResponse, *utils.PaginationResult, error) {
	// Verify delivery order exists
	_, err := u.deliveryOrderRepo.FindByID(ctx, deliveryOrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrDeliveryOrderNotFound
		}
		return nil, nil, err
	}

	// Fetch paginated items
	items, total, err := u.deliveryOrderRepo.ListItems(ctx, deliveryOrderID, req)
	if err != nil {
		return nil, nil, err
	}

	// Map to response DTOs
	responses := make([]dto.DeliveryOrderItemResponse, len(items))
	for i := range items {
		responses[i] = mapper.ToDeliveryOrderItemResponse(&items[i])
	}

	// Calculate pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *deliveryOrderUsecase) GetByID(ctx context.Context, id string) (*dto.DeliveryOrderResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.DeliveryOrder{}, id, security.DefaultScopeQueryOptions()) {
		return nil, ErrDeliveryOrderNotFound
	}
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryOrderNotFound
		}
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(deliveryOrder)
	return &response, nil
}

func (u *deliveryOrderUsecase) Create(ctx context.Context, req *dto.CreateDeliveryOrderRequest, createdBy *string) (*dto.DeliveryOrderResponse, error) {
	warehouseID := strings.TrimSpace(req.WarehouseID)
	if warehouseID == "" {
		return nil, errors.New(errDeliveryWarehouseIDRequired)
	}
	req.WarehouseID = warehouseID

	// Verify sales order exists
	salesOrder, err := u.salesOrderRepo.FindByID(ctx, req.SalesOrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliverySalesOrderNotFound
		}
		return nil, err
	}

	// Auto-fill receiver info from sales order customer if not provided
	if req.ReceiverName == "" && salesOrder.CustomerName != "" {
		req.ReceiverName = salesOrder.CustomerName
	}
	if req.ReceiverPhone == "" && salesOrder.CustomerPhone != "" {
		req.ReceiverPhone = salesOrder.CustomerPhone
	}

	// Query pending delivery quantities from existing non-cancelled DOs
	pendingQtyMap, err := u.deliveryOrderRepo.GetPendingDeliveryQtyBySalesOrder(ctx, req.SalesOrderID)
	if err != nil {
		return nil, err
	}

	// Check if sales order is already fully allocated (delivered + pending DO qty >= ordered qty)
	isFullyAllocated := true
	for _, item := range salesOrder.Items {
		pendingQty := pendingQtyMap[item.ID]
		allocatedQty := item.DeliveredQuantity + pendingQty
		if item.Quantity > allocatedQty {
			isFullyAllocated = false
			break
		}
	}
	if len(salesOrder.Items) > 0 && isFullyAllocated {
		return nil, errors.New("sales order is already fully fulfilled — all items have been delivered or allocated to existing delivery orders")
	}

	// Validate products and batches
	for _, item := range req.Items {
		product, err := u.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrDeliveryProductNotFound
			}
			return nil, err
		}

		// Use product selling price if price not provided
		if item.Price == 0 {
			item.Price = product.SellingPrice
		}

		// Check for over-delivery (including pending DO quantities)
		if item.SalesOrderItemID != nil {
			var soItem *models.SalesOrderItem
			for _, soi := range salesOrder.Items {
				if soi.ID == *item.SalesOrderItemID {
					soItem = &soi
					break
				}
			}

			if soItem != nil {
				pendingQty := pendingQtyMap[soItem.ID]
				remaining := soItem.Quantity - soItem.DeliveredQuantity - pendingQty
				if item.Quantity > remaining {
					return nil, errors.New("cannot deliver more than remaining quantity (over-delivery)")
				}
			}
		}

		// Validate batch exists and has sufficient stock
		if item.InventoryBatchID == nil {
			return nil, errors.New("inventory_batch_id is required")
		}
		batchWarehouseID, batchCurrentQty, err := u.getBatchShipmentContext(ctx, *item.InventoryBatchID)
		if err != nil {
			return nil, err
		}
		if item.WarehouseID != nil && strings.TrimSpace(*item.WarehouseID) != "" && !strings.EqualFold(strings.TrimSpace(*item.WarehouseID), batchWarehouseID) {
			return nil, fmt.Errorf("inventory batch %s does not belong to warehouse %s", *item.InventoryBatchID, strings.TrimSpace(*item.WarehouseID))
		}
		if batchCurrentQty < item.Quantity {
			return nil, ErrInsufficientBatchStock
		}
	}

	if strings.TrimSpace(warehouseID) == "" {
		for _, item := range req.Items {
			if item.WarehouseID != nil && strings.TrimSpace(*item.WarehouseID) != "" {
				warehouseID = strings.TrimSpace(*item.WarehouseID)
				break
			}
		}
		if warehouseID == "" && len(req.Items) > 0 && req.Items[0].InventoryBatchID != nil {
			batchWarehouseID, _, err := u.getBatchShipmentContext(ctx, *req.Items[0].InventoryBatchID)
			if err != nil {
				return nil, err
			}
			warehouseID = batchWarehouseID
		}
	}
	req.WarehouseID = warehouseID

	// Generate delivery order number
	code, err := u.deliveryOrderRepo.GetNextDeliveryNumber(ctx, "DO")
	if err != nil {
		return nil, err
	}

	// Convert request to model
	deliveryOrder, err := mapper.ToDeliveryOrderModel(req, code, createdBy)
	if err != nil {
		return nil, err
	}

	// Check if this is a partial delivery
	deliveryOrder.IsPartialDelivery = u.isPartialDelivery(salesOrder, deliveryOrder)

	// Create delivery order and reserve batch stock (wrapped in transaction)
	err = database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		// Create delivery order
		if err := u.deliveryOrderRepo.Create(txCtx, deliveryOrder); err != nil {
			return err
		}

		// Reserve stock at batch level for each item
		for _, item := range deliveryOrder.Items {
			if item.InventoryBatchID != nil {
				if err := u.inventoryUC.ReserveBatchStock(txCtx, *item.InventoryBatchID, item.Quantity); err != nil {
					return err
				}
			}
		}

		// Sales order status no longer changes based on Delivery Order creation

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fetch created delivery order with relations
	created, err := u.deliveryOrderRepo.FindByID(ctx, deliveryOrder.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(created)
	logSalesAudit(u.auditService, ctx, "delivery_order.create", created.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":           created.Code,
			"status":         created.Status,
			"delivery_date":  created.DeliveryDate,
			"sales_order_id": created.SalesOrderID,
		},
	})
	return &response, nil
}

func (u *deliveryOrderUsecase) Update(ctx context.Context, id string, req *dto.UpdateDeliveryOrderRequest) (*dto.DeliveryOrderResponse, error) {
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryOrderNotFound
		}
		return nil, err
	}

	// Check if delivery order can be modified
	if deliveryOrder.Status == models.DeliveryOrderStatusShipped ||
		deliveryOrder.Status == models.DeliveryOrderStatusDelivered ||
		deliveryOrder.Status == models.DeliveryOrderStatusCancelled {
		return nil, ErrInvalidDeliveryOrderStatus
	}

	beforeSnapshot := deliveryOrderAuditSnapshot(deliveryOrder)
	if req.WarehouseID != nil {
		trimmedWarehouseID := strings.TrimSpace(*req.WarehouseID)
		if trimmedWarehouseID == "" {
			return nil, errors.New(errDeliveryWarehouseIDRequired)
		}
		req.WarehouseID = &trimmedWarehouseID
	}

	// Validate products and batches if items are being updated
	if len(req.Items) > 0 {
		for _, item := range req.Items {
			product, err := u.productRepo.FindByID(ctx, item.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrProductNotFound
				}
				return nil, err
			}

			// Use product selling price if price not provided
			if item.Price == 0 {
				item.Price = product.SellingPrice
			}

			// Validate batch exists and has sufficient stock (only when batch ID is provided)
			if item.InventoryBatchID != nil {
				batchWarehouseID, batchCurrentQty, err := u.getBatchShipmentContext(ctx, *item.InventoryBatchID)
				if err != nil {
					return nil, err
				}
				if item.WarehouseID != nil && strings.TrimSpace(*item.WarehouseID) != "" && !strings.EqualFold(strings.TrimSpace(*item.WarehouseID), batchWarehouseID) {
					return nil, fmt.Errorf("inventory batch %s does not belong to warehouse %s", *item.InventoryBatchID, strings.TrimSpace(*item.WarehouseID))
				}
				if batchCurrentQty < item.Quantity {
					return nil, ErrInsufficientBatchStock
				}
			}
		}
	}

	// Release old batch reservations before applying new ones
	err = database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		// Release existing batch reservations
		for _, oldItem := range deliveryOrder.Items {
			if oldItem.InventoryBatchID != nil {
				if err := u.inventoryUC.ReleaseBatchStock(txCtx, *oldItem.InventoryBatchID, oldItem.Quantity); err != nil {
					return err
				}
			}
		}

		// Update model
		if err := mapper.UpdateDeliveryOrderModel(deliveryOrder, req); err != nil {
			return err
		}

		// Update delivery order
		if err := u.deliveryOrderRepo.Update(txCtx, deliveryOrder); err != nil {
			return err
		}

		// Reserve new batch stock
		for _, item := range deliveryOrder.Items {
			if item.InventoryBatchID != nil {
				if item.WarehouseID != nil && strings.TrimSpace(*item.WarehouseID) != "" {
					batchWarehouseID, _, err := u.getBatchShipmentContext(txCtx, *item.InventoryBatchID)
					if err != nil {
						return err
					}
					if !strings.EqualFold(strings.TrimSpace(*item.WarehouseID), batchWarehouseID) {
						return fmt.Errorf("inventory batch %s does not belong to warehouse %s", *item.InventoryBatchID, strings.TrimSpace(*item.WarehouseID))
					}
				}
				if err := u.inventoryUC.ReserveBatchStock(txCtx, *item.InventoryBatchID, item.Quantity); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fetch updated delivery order with relations
	updated, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(updated)
	logSalesAudit(u.auditService, ctx, "delivery_order.update", id, map[string]interface{}{
		"before": beforeSnapshot,
		"after":  deliveryOrderAuditSnapshot(updated),
	})
	return &response, nil
}

func (u *deliveryOrderUsecase) Delete(ctx context.Context, id string) error {
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeliveryOrderNotFound
		}
		return err
	}

	// Only allow deletion of draft delivery orders
	if deliveryOrder.Status != models.DeliveryOrderStatusDraft {
		return ErrInvalidDeliveryOrderStatus
	}

	// Release batch stock reservations and delete (wrapped in transaction)
	return database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		for _, item := range deliveryOrder.Items {
			if item.InventoryBatchID != nil {
				if err := u.inventoryUC.ReleaseBatchStock(txCtx, *item.InventoryBatchID, item.Quantity); err != nil {
					return err
				}
			}
			// Release product-level reservation as well
			if err := u.inventoryUC.ReleaseStock(txCtx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}

		return u.deliveryOrderRepo.Delete(txCtx, id)
	})
}

func (u *deliveryOrderUsecase) UpdateStatus(ctx context.Context, id string, req *dto.UpdateDeliveryOrderStatusRequest, userID *string) (*dto.DeliveryOrderResponse, error) {
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryOrderNotFound
		}
		return nil, err
	}

	newStatus := models.DeliveryOrderStatus(req.Status)
	previousStatus := deliveryOrder.Status

	log.Printf("[Sales] UpdateStatus: ID=%s, currentStatus=%s, nextStatus=%s, warehouseID=%v", id, deliveryOrder.Status, newStatus, deliveryOrder.WarehouseID)

	// Validate status transition
	if !u.isValidStatusTransition(deliveryOrder.Status, newStatus) {
		return nil, ErrInvalidDeliveryStatusTransition
	}

	// For SHIPPED and DELIVERED, they must use specialized methods to ensure transactional logic runs
	if newStatus == models.DeliveryOrderStatusShipped {
		return nil, errors.New("cannot change status to SHIPPED via generic status update — please use the specialized /ship action")
	}
	if newStatus == models.DeliveryOrderStatusDelivered {
		return nil, errors.New("cannot change status to DELIVERED via generic status update — please use the specialized /deliver action")
	}

	var reason *string
	if req.CancellationReason != nil {
		reason = req.CancellationReason
	}

	// Release stock reservations when cancelling to prevent "trapped" inventory
	if newStatus == models.DeliveryOrderStatusCancelled {
		return u.cancelAndReleaseStock(ctx, deliveryOrder, userID, reason)
	}

	if newStatus == models.DeliveryOrderStatusApproved {
		err := database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
			txCtx := database.WithTx(ctx, tx)

			if err := u.postApprovedDeliveryStockAndJournal(txCtx, tx, deliveryOrder, userID); err != nil {
				return err
			}

			return u.deliveryOrderRepo.UpdateStatus(txCtx, id, newStatus, userID, reason)
		})
		if err != nil {
			return nil, err
		}

		updated, err := u.deliveryOrderRepo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		response := mapper.ToDeliveryOrderResponse(updated)
		logSalesAudit(u.auditService, ctx, "delivery_order.status_change", id, map[string]interface{}{
			"before_status": previousStatus,
			"after_status":  updated.Status,
			"reason":        req.CancellationReason,
		})
		return &response, nil
	}

	// Validation: Prepared status requires a warehouse
	if newStatus == models.DeliveryOrderStatusPrepared {
		if deliveryOrder.WarehouseID == nil || strings.TrimSpace(*deliveryOrder.WarehouseID) == "" {
			return nil, errors.New(errDeliveryWarehouseIDRequired)
		}
	}

	if err := u.deliveryOrderRepo.UpdateStatus(ctx, id, newStatus, userID, reason); err != nil {
		return nil, err
	}

	// Fetch updated delivery order
	updated, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Trigger approval notification when DO is sent for approval
	if newStatus == models.DeliveryOrderStatusSent {
		actorUserID := ""
		if userID != nil {
			actorUserID = *userID
		}
		if err := notificationService.CreateApprovalNotification(ctx, u.db, notificationService.ApprovalNotificationParams{
			PermissionCode: "delivery_order.approve",
			EntityType:     "delivery_order",
			EntityID:       updated.ID,
			Title:          "Delivery Order Approval",
			Message:        "A delivery order has been submitted and requires your approval.",
			ActorUserID:    actorUserID,
		}); err != nil {
			log.Printf("warning: failed to create delivery order notification: %v", err)
		}
	}

	response := mapper.ToDeliveryOrderResponse(updated)
	logSalesAudit(u.auditService, ctx, "delivery_order.status_change", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  updated.Status,
		"reason":        req.CancellationReason,
	})
	return &response, nil
}

// cancelAndReleaseStock handles DO cancellation with proper stock release in a single transaction.
// This mirrors the stock release logic in Delete() but applies to status cancellation
// from draft, approved, or prepared states where batch/product reservations exist.
func (u *deliveryOrderUsecase) cancelAndReleaseStock(
	ctx context.Context,
	deliveryOrder *models.DeliveryOrder,
	userID *string,
	reason *string,
) (*dto.DeliveryOrderResponse, error) {
	err := database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		// Release batch-level and product-level stock reservations for each item
		for _, item := range deliveryOrder.Items {
			if item.InventoryBatchID != nil {
				if err := u.inventoryUC.ReleaseBatchStock(txCtx, *item.InventoryBatchID, item.Quantity); err != nil {
					return fmt.Errorf("failed to release batch stock for item %s: %w", item.ID, err)
				}
			}
			if err := u.inventoryUC.ReleaseStock(txCtx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("failed to release product stock for item %s: %w", item.ID, err)
			}
		}

		return u.deliveryOrderRepo.UpdateStatus(txCtx, deliveryOrder.ID, models.DeliveryOrderStatusCancelled, userID, reason)
	})
	if err != nil {
		return nil, err
	}

	updated, err := u.deliveryOrderRepo.FindByID(ctx, deliveryOrder.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(updated)
	logSalesAudit(u.auditService, ctx, "delivery_order.status_change", updated.ID, map[string]interface{}{
		"before_status": deliveryOrder.Status,
		"after_status":  updated.Status,
		"reason":        reason,
	})
	return &response, nil
}

func (u *deliveryOrderUsecase) Ship(ctx context.Context, id string, req *dto.ShipDeliveryOrderRequest, userID *string) (*dto.DeliveryOrderResponse, error) {
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryOrderNotFound
		}
		return nil, err
	}

	// Validate status
	if deliveryOrder.Status != models.DeliveryOrderStatusPrepared {
		return nil, ErrInvalidDeliveryStatusTransition
	}
	if deliveryOrder.WarehouseID == nil || strings.TrimSpace(*deliveryOrder.WarehouseID) == "" {
		return nil, errors.New(errDeliveryWarehouseIDRequired)
	}

	// Atomic transaction for all stock and status changes
	err = database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		log.Printf("[Sales] PROCESSING SHIP for DO %s (%s)", id, deliveryOrder.Code)

		if deliveryOrder.WarehouseID == nil {
			log.Printf("[Sales] SHIP FAILED: DO %s missing warehouse ID", deliveryOrder.Code)
			return fmt.Errorf("delivery order %s has no source warehouse assigned", deliveryOrder.Code)
		}

		if len(deliveryOrder.Items) == 0 {
			log.Printf("[Sales] SHIP FAILED: DO %s has no items", deliveryOrder.Code)
			return fmt.Errorf("delivery order %s has no items", deliveryOrder.Code)
		}

		// 1. Process stock changes, delivery qty updates, and COGS snapshots per item
		for i := range deliveryOrder.Items {
			item := &deliveryOrder.Items[i]
			isInventoryTracked := item.Product == nil || item.Product.IsInventoryTracked

			avgCostSnapshot := resolveDeliveryOrderItemCost(item)
			cogsAmount := 0.0

			if isInventoryTracked {
				if item.InventoryBatchID == nil || strings.TrimSpace(*item.InventoryBatchID) == "" {
					return fmt.Errorf("inventory batch is required for product %s", item.ProductID)
				}

				log.Printf("[Sales] Deducting stock for Product %s, Batch %s, Qty %.3f", item.ProductID, *item.InventoryBatchID, item.Quantity)

				if err := u.inventoryUC.ReleaseBatchStock(txCtx, *item.InventoryBatchID, item.Quantity); err != nil {
					return fmt.Errorf("release batch stock: %w", err)
				}

				if err := u.inventoryUC.ReleaseStock(txCtx, item.ProductID, item.Quantity); err != nil {
					return fmt.Errorf("release product stock: %w", err)
				}

				// Use sales order code as reference (not delivery order) so ledger shows SO code
				soCode := deliveryOrder.Code
				if deliveryOrder.SalesOrder != nil && strings.TrimSpace(deliveryOrder.SalesOrder.Code) != "" {
					soCode = deliveryOrder.SalesOrder.Code
				}

				if err := u.inventoryUC.DeductStock(txCtx, *item.InventoryBatchID, item.Quantity, "SO", soCode); err != nil {
					return fmt.Errorf("deduct stock: %w", err)
				}

				movementReq := &inventoryDto.StockMovementRequest{
					InventoryBatchID: *item.InventoryBatchID,
					ProductID:        item.ProductID,
					WarehouseID:      *deliveryOrder.WarehouseID,
					Type:             "OUT",
					Quantity:         item.Quantity,
					ReferenceType:    "SALES_ORDER",
					ReferenceID:      deliveryOrder.SalesOrderID,
					ReferenceNumber:  soCode,
					Description:      "Sales Order Fulfillment",
					Cost:             avgCostSnapshot,
					CreatedBy:        userID,
					SkipJournaling:   true,
				}

				if _, err := u.inventoryUC.CreateStockMovement(txCtx, movementReq); err != nil {
					return fmt.Errorf("create movement: %w", err)
				}

				cogsAmount = roundCurrency(avgCostSnapshot * item.Quantity)
			} else {
				// Keep reservation ledger clean even for non-stock products when previous flows reserved stock.
				_ = u.inventoryUC.ReleaseStock(txCtx, item.ProductID, item.Quantity)
			}

			item.AvgCostSnapshot = avgCostSnapshot
			item.COGSAmount = cogsAmount

			if err := tx.Model(&models.DeliveryOrderItem{}).
				Where("id = ?", item.ID).
				Updates(map[string]interface{}{
					"avg_cost_snapshot": roundCost(avgCostSnapshot),
					"cogs_amount":       cogsAmount,
				}).Error; err != nil {
				return fmt.Errorf("update delivery item costing: %w", err)
			}

			if item.SalesOrderItemID != nil {
				if err := u.salesOrderRepo.UpdateItemDeliveredQty(txCtx, *item.SalesOrderItemID, item.Quantity); err != nil {
					return fmt.Errorf("update sales order delivered qty: %w", err)
				}
			}
		}

		// 2. Post logistics journal (COGS and inventory release only)
		journalEntry, err := u.salesJournalSvc.GenerateSalesJournal(txCtx, deliveryOrder)
		if err != nil {
			return fmt.Errorf("post delivery sales journal: %w", err)
		}

		// Link journal entry to delivery order before shipping
		if journalEntry != nil && strings.TrimSpace(journalEntry.ID) != "" {
			if err := tx.Model(&models.DeliveryOrder{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{"journal_entry_id": journalEntry.ID, "is_posted": true}).Error; err != nil {
				return fmt.Errorf("link delivery order journal: %w", err)
			}
		}

		// 3. Mark DO as shipped only after stock and accounting completed
		if err := u.deliveryOrderRepo.Ship(txCtx, id, userID, req.TrackingNumber); err != nil {
			return fmt.Errorf("repo ship: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch updated delivery order
	updated, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(updated)
	logSalesAudit(u.auditService, ctx, "delivery_order.ship", id, map[string]interface{}{
		"after": map[string]interface{}{
			"status":          updated.Status,
			"tracking_number": updated.TrackingNumber,
			"shipped_at":      updated.ShippedAt,
		},
	})
	return &response, nil
}

func (u *deliveryOrderUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, errors.New(errDeliveryDBNil)
	}
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.DeliveryOrder{}, id, security.DefaultScopeQueryOptions()) {
		return nil, 0, ErrDeliveryOrderNotFound
	}

	return listAuditTrailEntries(ctx, u.db, id, "delivery_order.", page, perPage)
}

func (u *deliveryOrderUsecase) Deliver(ctx context.Context, id string, req *dto.DeliverDeliveryOrderRequest, userID *string) (*dto.DeliveryOrderResponse, error) {
	deliveryOrder, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryOrderNotFound
		}
		return nil, err
	}

	// Validate status
	if deliveryOrder.Status != models.DeliveryOrderStatusShipped {
		return nil, ErrInvalidDeliveryStatusTransition
	}

	// Mark as delivered and close SO when now fully fulfilled + fully settled.
	err = database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := u.deliveryOrderRepo.Deliver(txCtx, id, userID, req.ReceiverSignature, req.ReceiverName); err != nil {
			return err
		}

		closeSalesOrderWhenSettledAndFulfilled(tx, deliveryOrder.SalesOrderID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fetch updated delivery order
	updated, err := u.deliveryOrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToDeliveryOrderResponse(updated)
	logSalesAudit(u.auditService, ctx, "delivery_order.deliver", id, map[string]interface{}{
		"after": map[string]interface{}{
			"status":        updated.Status,
			"delivered_at":  updated.DeliveredAt,
			"receiver_name": updated.ReceiverName,
		},
	})
	return &response, nil
}

// SelectBatches selects available batches using FIFO or FEFO method
func (u *deliveryOrderUsecase) SelectBatches(ctx context.Context, req *dto.BatchSelectionRequest) (*dto.BatchSelectionResponse, error) {
	// Validate product exists
	_, err := u.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryProductNotFound
		}
		return nil, err
	}

	// Fetch available batches from Inventory Module
	batches, err := u.inventoryUC.SelectBatches(ctx, req.ProductID, req.Quantity, req.Method)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	var responseBatches []dto.BatchInfo
	var totalAvailable float64

	for _, b := range batches {
		responseBatches = append(responseBatches, dto.BatchInfo{
			ID:           b.ID,
			BatchNumber:  b.BatchNumber,
			Quantity:     b.Quantity, // Current Quantity
			ExpiryDate:   b.ExpiredAt,
			ReceivedDate: b.ReceivedAt,
			Available:    float64(b.Quantity), // Simplified available
		})
		totalAvailable += float64(b.Quantity)
	}

	return &dto.BatchSelectionResponse{
		Batches:        responseBatches,
		TotalAvailable: totalAvailable,
	}, nil
}

// isValidStatusTransition validates if status transition is allowed
func (u *deliveryOrderUsecase) isValidStatusTransition(current, new models.DeliveryOrderStatus) bool {
	validTransitions := map[models.DeliveryOrderStatus][]models.DeliveryOrderStatus{
		models.DeliveryOrderStatusDraft: {
			models.DeliveryOrderStatusSent,
			models.DeliveryOrderStatusCancelled,
		},
		models.DeliveryOrderStatusSent: {
			models.DeliveryOrderStatusApproved,
			models.DeliveryOrderStatusRejected,
		},
		models.DeliveryOrderStatusApproved: {
			models.DeliveryOrderStatusPrepared,
			models.DeliveryOrderStatusCancelled,
		},
		models.DeliveryOrderStatusRejected: {
			models.DeliveryOrderStatusDraft,
		},
		models.DeliveryOrderStatusPrepared: {
			models.DeliveryOrderStatusShipped,
			models.DeliveryOrderStatusCancelled,
		},
		models.DeliveryOrderStatusShipped: {
			models.DeliveryOrderStatusDelivered,
		},
		models.DeliveryOrderStatusDelivered: {
			// Cannot transition from delivered
		},
		models.DeliveryOrderStatusCancelled: {
			// Cannot transition from cancelled
		},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}

	return false
}

func deliveryOrderAuditSnapshot(deliveryOrder *models.DeliveryOrder) map[string]interface{} {
	if deliveryOrder == nil {
		return nil
	}

	return map[string]interface{}{
		"code":                deliveryOrder.Code,
		"status":              deliveryOrder.Status,
		"delivery_date":       deliveryOrder.DeliveryDate,
		"sales_order_id":      deliveryOrder.SalesOrderID,
		"warehouse_id":        deliveryOrder.WarehouseID,
		"delivered_by_id":     deliveryOrder.DeliveredByID,
		"courier_agency_id":   deliveryOrder.CourierAgencyID,
		"tracking_number":     deliveryOrder.TrackingNumber,
		"receiver_name":       deliveryOrder.ReceiverName,
		"receiver_phone":      deliveryOrder.ReceiverPhone,
		"delivery_address":    deliveryOrder.DeliveryAddress,
		"receiver_signature":  deliveryOrder.ReceiverSignature,
		"is_partial_delivery": deliveryOrder.IsPartialDelivery,
		"notes":               deliveryOrder.Notes,
		"items":               deliveryOrderAuditItems(deliveryOrder.Items),
	}
}

func deliveryOrderAuditItems(items []models.DeliveryOrderItem) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"id":                   item.ID,
			"sales_order_item_id":  item.SalesOrderItemID,
			"product_id":           item.ProductID,
			"inventory_batch_id":   item.InventoryBatchID,
			"quantity":             item.Quantity,
			"price":                item.Price,
			"subtotal":             item.Subtotal,
			"avg_cost_snapshot":    item.AvgCostSnapshot,
			"cogs_amount":          item.COGSAmount,
			"is_equipment":         item.IsEquipment,
			"installation_status":  item.InstallationStatus,
			"function_test_status": item.FunctionTestStatus,
			"installation_date":    item.InstallationDate,
			"function_test_date":   item.FunctionTestDate,
			"installation_notes":   item.InstallationNotes,
		})
	}

	return out
}

// isPartialDelivery checks if delivery order is a partial delivery
func (u *deliveryOrderUsecase) isPartialDelivery(salesOrder *models.SalesOrder, deliveryOrder *models.DeliveryOrder) bool {
	// Calculate total delivered quantity per product
	deliveredByProduct := make(map[string]float64)
	for _, item := range deliveryOrder.Items {
		deliveredByProduct[item.ProductID] += item.Quantity
	}

	// Check if any item is partially delivered
	for _, orderItem := range salesOrder.Items {
		deliveredQty := deliveredByProduct[orderItem.ProductID]
		if deliveredQty > 0 && deliveredQty < orderItem.Quantity {
			return true
		}
	}

	return false
}

func resolveDeliveryOrderItemCost(item *models.DeliveryOrderItem) float64 {
	if item == nil {
		return 0
	}

	if item.InventoryBatch != nil && item.InventoryBatch.CostPrice > 0 {
		return roundCost(item.InventoryBatch.CostPrice)
	}

	if item.Product != nil && item.Product.CurrentHpp > 0 {
		return roundCost(item.Product.CurrentHpp)
	}

	return 0
}

func (u *deliveryOrderUsecase) postApprovedDeliveryStockAndJournal(ctx context.Context, tx *gorm.DB, deliveryOrder *models.DeliveryOrder, userID *string) error {
	if deliveryOrder == nil {
		return errors.New("delivery order is nil")
	}

	for i := range deliveryOrder.Items {
		item := &deliveryOrder.Items[i]
		isInventoryTracked := item.Product == nil || item.Product.IsInventoryTracked

		avgCostSnapshot := resolveDeliveryOrderItemCost(item)
		cogsAmount := 0.0

		if isInventoryTracked {
			if item.InventoryBatchID == nil || strings.TrimSpace(*item.InventoryBatchID) == "" {
				return fmt.Errorf("inventory batch is required for product %s", item.ProductID)
			}

			batchWarehouseID, batchCurrentQty, err := u.getBatchShipmentContext(ctx, *item.InventoryBatchID)
			if err != nil {
				return fmt.Errorf("resolve batch shipment context for product %s: %w", item.ProductID, err)
			}
			if batchCurrentQty < item.Quantity {
				return fmt.Errorf("validate batch stock for product %s: %w", item.ProductID, inventoryUsecase.ErrInsufficientBatchStock)
			}

			log.Printf("[Sales] Approving stock deduction for Product %s, Batch %s, Qty %.3f", item.ProductID, *item.InventoryBatchID, item.Quantity)

			if err := u.inventoryUC.ReleaseBatchStock(ctx, *item.InventoryBatchID, item.Quantity); err != nil {
				return fmt.Errorf("release batch stock: %w", err)
			}

			if err := u.inventoryUC.ReleaseStock(ctx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("release product stock: %w", err)
			}

			if err := u.inventoryUC.DeductStock(ctx, *item.InventoryBatchID, item.Quantity); err != nil {
				return fmt.Errorf("deduct stock: %w", err)
			}

			movementReq := &inventoryDto.StockMovementRequest{
				InventoryBatchID: *item.InventoryBatchID,
				ProductID:        item.ProductID,
				WarehouseID:      batchWarehouseID,
				Type:             "OUT",
				Quantity:         item.Quantity,
				ReferenceType:    reference.RefTypeDeliveryOrder,
				ReferenceID:      deliveryOrder.ID,
				ReferenceNumber:  deliveryOrder.Code,
				Description:      "Delivery Order Approval",
				Cost:             avgCostSnapshot,
				CreatedBy:        userID,
				SkipJournaling:   true,
			}

			if _, err := u.inventoryUC.CreateStockMovement(ctx, movementReq); err != nil {
				return fmt.Errorf("create movement: %w", err)
			}

			cogsAmount = roundCurrency(avgCostSnapshot * item.Quantity)
		} else {
			_ = u.inventoryUC.ReleaseStock(ctx, item.ProductID, item.Quantity)
		}

		item.AvgCostSnapshot = avgCostSnapshot
		item.COGSAmount = cogsAmount

		if err := tx.Model(&models.DeliveryOrderItem{}).
			Where("id = ?", item.ID).
			Updates(map[string]interface{}{
				"avg_cost_snapshot": roundCost(avgCostSnapshot),
				"cogs_amount":       cogsAmount,
			}).Error; err != nil {
			return fmt.Errorf("update delivery item costing: %w", err)
		}
	}

	journalEntry, err := u.salesJournalSvc.GenerateSalesJournal(ctx, deliveryOrder)
	if err != nil {
		return fmt.Errorf("post delivery sales journal: %w", err)
	}

	if journalEntry != nil && strings.TrimSpace(journalEntry.ID) != "" {
		if err := tx.Model(&models.DeliveryOrder{}).
			Where("id = ?", deliveryOrder.ID).
			Updates(map[string]interface{}{"journal_entry_id": journalEntry.ID, "is_posted": true}).Error; err != nil {
			return fmt.Errorf("link delivery order journal: %w", err)
		}
		deliveryOrder.JournalEntryID = &journalEntry.ID
		deliveryOrder.IsPosted = true
	}

	return nil
}

func (u *deliveryOrderUsecase) getBatchShipmentContext(ctx context.Context, batchID string) (string, float64, error) {
	trimmedBatchID := strings.TrimSpace(batchID)
	if trimmedBatchID == "" {
		return "", 0, ErrBatchNotFound
	}

	type batchRow struct {
		WarehouseID   string  `gorm:"column:warehouse_id"`
		CurrentQty    float64 `gorm:"column:current_quantity"`
		ReservedQty   float64 `gorm:"column:reserved_quantity"`
	}

	var row batchRow
	err := database.GetDB(ctx, u.db).
		Table("inventory_batches").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("warehouse_id, current_quantity, reserved_quantity").
		Where("id = ? AND deleted_at IS NULL", trimmedBatchID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, ErrBatchNotFound
		}
		return "", 0, err
	}

	warehouseID := strings.TrimSpace(row.WarehouseID)
	if warehouseID == "" {
		return "", 0, errors.New("batch warehouse is required")
	}

	return warehouseID, row.CurrentQty, nil
}

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}

func roundCost(value float64) float64 {
	return math.Round(value*1000000) / 1000000
}
