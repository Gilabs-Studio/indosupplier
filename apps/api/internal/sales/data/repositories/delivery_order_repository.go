package repositories

import (
	"context"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeliveryOrderRepository defines the interface for delivery order data access
type DeliveryOrderRepository interface {
	FindByID(ctx context.Context, id string) (*models.DeliveryOrder, error)
	FindByCode(ctx context.Context, code string) (*models.DeliveryOrder, error)
	List(ctx context.Context, req *dto.ListDeliveryOrdersRequest) ([]models.DeliveryOrder, int64, error)
	ListItems(ctx context.Context, deliveryOrderID string, req *dto.ListDeliveryOrderItemsRequest) ([]models.DeliveryOrderItem, int64, error)
	Create(ctx context.Context, do *models.DeliveryOrder) error
	Update(ctx context.Context, do *models.DeliveryOrder) error
	Delete(ctx context.Context, id string) error
	GetNextDeliveryNumber(ctx context.Context, prefix string) (string, error)
	UpdateStatus(ctx context.Context, id string, status models.DeliveryOrderStatus, userID *string, reason *string) error
	Ship(ctx context.Context, id string, userID *string, trackingNumber string) error
	Deliver(ctx context.Context, id string, userID *string, receiverSignature string, receiverName string) error
	GetPendingDeliveryQtyBySalesOrder(ctx context.Context, salesOrderID string) (map[string]float64, error)
}

type deliveryOrderRepository struct {
	db *gorm.DB
}

// NewDeliveryOrderRepository creates a new DeliveryOrderRepository
func NewDeliveryOrderRepository(db *gorm.DB) DeliveryOrderRepository {
	return &deliveryOrderRepository{db: db}
}

func (r *deliveryOrderRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *deliveryOrderRepository) FindByID(ctx context.Context, id string) (*models.DeliveryOrder, error) {
	var deliveryOrder models.DeliveryOrder
	err := r.getDB(ctx).
		Preload("Warehouse").
		Preload("SalesOrder").
		Preload("DeliveredBy").
		Preload("CourierAgency").
		Preload("Items.Product").
		Preload("Items.Warehouse").
		Preload("Items.SalesOrderItem").
		Preload("Items.InventoryBatch").
		Where("id = ?", id).
		First(&deliveryOrder).Error
	if err != nil {
		return nil, err
	}
	return &deliveryOrder, nil
}

func (r *deliveryOrderRepository) FindByCode(ctx context.Context, code string) (*models.DeliveryOrder, error) {
	var deliveryOrder models.DeliveryOrder
	err := r.getDB(ctx).
		Preload("Warehouse").
		Preload("SalesOrder").
		Preload("DeliveredBy").
		Preload("CourierAgency").
		Preload("Items.Product").
		Preload("Items.Warehouse").
		Preload("Items.SalesOrderItem").
		Preload("Items.InventoryBatch").
		Where("code = ?", code).
		First(&deliveryOrder).Error
	if err != nil {
		return nil, err
	}
	return &deliveryOrder, nil
}

func (r *deliveryOrderRepository) List(ctx context.Context, req *dto.ListDeliveryOrdersRequest) ([]models.DeliveryOrder, int64, error) {
	var deliveryOrders []models.DeliveryOrder
	var total int64
	var query *gorm.DB

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = r.db.WithContext(ctx).Model(&models.DeliveryOrder{}).
			Joins("LEFT JOIN sales_orders ON sales_orders.id = delivery_orders.sales_order_id").
			Joins("LEFT JOIN employees ON employees.id = sales_orders.sales_rep_id")

		// Apply tenant filter manually since we are using joins
		var err error
		query, err = applyTenantFilter(ctx, query, "delivery_orders.tenant_id", "sales_orders.tenant_id")
		if err != nil {
			return nil, 0, err
		}

		query = query.Where("sales_orders.customer_name ILIKE ? OR employees.name ILIKE ? OR delivery_orders.receiver_name ILIKE ? OR delivery_orders.code ILIKE ? OR delivery_orders.tracking_number ILIKE ? OR delivery_orders.notes ILIKE ?", search, search, search, search, search, search)
	} else {
		// Even without search, if we use getDB it might fail on later joins if added.
		// But let's be safe and use explicit prefix.
		query = r.db.WithContext(ctx).Model(&models.DeliveryOrder{})
		var err error
		query, err = applyTenantFilter(ctx, query, "delivery_orders.tenant_id")
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OwnerUserIDColumn: "delivery_orders.created_by",
	})

	// Apply status filter
	if req.Status != "" {
		query = query.Where("delivery_orders.status = ?", req.Status)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("delivery_orders.delivery_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("delivery_orders.delivery_date <= ?", req.DateTo)
	}

	// Apply sales order filter
	if req.SalesOrderID != "" {
		query = query.Where("delivery_orders.sales_order_id = ?", req.SalesOrderID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
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
	offset := (page - 1) * perPage

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"code":          "delivery_orders.code",
		"delivery_date": "delivery_orders.delivery_date",
		"status":        "delivery_orders.status",
		"created_at":    "delivery_orders.created_at",
		"updated_at":    "delivery_orders.updated_at",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(req.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "delivery_orders.delivery_date"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err := query.
		Preload("Warehouse").
		Preload("SalesOrder").
		Preload("DeliveredBy").
		Preload("CourierAgency").
		Limit(perPage).
		Offset(offset).
		Find(&deliveryOrders).Error
	if err != nil {
		return nil, 0, err
	}

	return deliveryOrders, total, nil
}

func (r *deliveryOrderRepository) Create(ctx context.Context, do *models.DeliveryOrder) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Create delivery order
		if err := tx.Create(do).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *deliveryOrderRepository) Update(ctx context.Context, do *models.DeliveryOrder) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update delivery order (exclude items to avoid conflict with manual management below)
		if err := tx.Omit("Items").Save(do).Error; err != nil {
			return err
		}

		// Delete existing items
		if err := tx.Where("delivery_order_id = ?", do.ID).Delete(&models.DeliveryOrderItem{}).Error; err != nil {
			return err
		}

		// Create new items
		if len(do.Items) > 0 {
			for i := range do.Items {
				do.Items[i].DeliveryOrderID = do.ID
				if err := tx.Create(&do.Items[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *deliveryOrderRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete items first (CASCADE should handle this, but explicit for safety)
		if err := tx.Where("delivery_order_id = ?", id).Delete(&models.DeliveryOrderItem{}).Error; err != nil {
			return err
		}

		// Delete delivery order
		return tx.Delete(&models.DeliveryOrder{}, "id = ?", id).Error
	})
}

func (r *deliveryOrderRepository) GetNextDeliveryNumber(ctx context.Context, prefix string) (string, error) {
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")
	prefixWithDate := prefix + "-" + dateStr

	var lastDeliveryOrder models.DeliveryOrder
	var sequence int

	// Find the last delivery order with the exact same prefix+date (including soft-deleted)
	err := r.getDB(ctx).
		Unscoped().
		Where("code LIKE ?", prefixWithDate+"%").
		Order("code DESC").
		First(&lastDeliveryOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No previous delivery order today, start from 1
			sequence = 1
		} else {
			return "", err
		}
	} else {
		// Extract sequence number from last code (format: PREFIX-YYYYMMDD-XXXX)
		parts := strings.Split(lastDeliveryOrder.Code, "-")
		if len(parts) >= 3 {
			// Sequence is the last part
			lastSeq, err := strconv.Atoi(parts[len(parts)-1])
			if err == nil {
				sequence = lastSeq + 1
			} else {
				sequence = 1
			}
		} else {
			sequence = 1
		}
	}

	// Format sequence with 4 digits
	code := prefixWithDate + "-" + formatSequence(sequence)

	return code, nil
}

func (r *deliveryOrderRepository) UpdateStatus(ctx context.Context, id string, status models.DeliveryOrderStatus, userID *string, reason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case models.DeliveryOrderStatusCancelled:
		updates["cancelled_by"] = userID
		updates["cancelled_at"] = database.GetDB(ctx, r.db).NowFunc()
		if reason != nil {
			updates["cancellation_reason"] = *reason
		}
	}

	return r.getDB(ctx).Model(&models.DeliveryOrder{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *deliveryOrderRepository) Ship(ctx context.Context, id string, userID *string, trackingNumber string) error {
	now := database.GetDB(ctx, r.db).NowFunc()
	updates := map[string]interface{}{
		"status":          models.DeliveryOrderStatusShipped,
		"shipped_by":      userID,
		"shipped_at":      now,
		"tracking_number": trackingNumber,
	}

	return r.getDB(ctx).Model(&models.DeliveryOrder{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *deliveryOrderRepository) Deliver(ctx context.Context, id string, userID *string, receiverSignature string, receiverName string) error {
	now := database.GetDB(ctx, r.db).NowFunc()
	updates := map[string]interface{}{
		"status":             models.DeliveryOrderStatusDelivered,
		"delivered_at":       now,
		"receiver_signature": receiverSignature,
		"receiver_name":      receiverName,
	}

	return r.getDB(ctx).Model(&models.DeliveryOrder{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListItems retrieves delivery order items with pagination
func (r *deliveryOrderRepository) ListItems(ctx context.Context, deliveryOrderID string, req *dto.ListDeliveryOrderItemsRequest) ([]models.DeliveryOrderItem, int64, error) {
	var items []models.DeliveryOrderItem
	var total int64

	// Set defaults
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

	// Count total items
	if err := r.getDB(ctx).Model(&models.DeliveryOrderItem{}).
		Where("delivery_order_id = ?", deliveryOrderID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated items with minimal preload (only product info)
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "image_url")
		}).
		Preload("SalesOrderItem").
		Preload("InventoryBatch").
		Where("delivery_order_id = ?", deliveryOrderID).
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetPendingDeliveryQtyBySalesOrder returns a map of sales_order_item_id -> total pending quantity
// from active (non-finalized) delivery orders for the given sales order.
func (r *deliveryOrderRepository) GetPendingDeliveryQtyBySalesOrder(ctx context.Context, salesOrderID string) (map[string]float64, error) {
	type result struct {
		SalesOrderItemID string  `json:"sales_order_item_id"`
		TotalQty         float64 `json:"total_qty"`
	}

	var results []result
	query := r.db.WithContext(ctx).Table("delivery_order_items doi")

	// Apply tenant filter with prefix
	var err error
	query, err = applyTenantFilter(ctx, query, "doi.tenant_id")
	if err != nil {
		return nil, err
	}

	err = query.
		Select("doi.sales_order_item_id, SUM(doi.quantity) as total_qty").
		Joins("JOIN delivery_orders dord ON dord.id = doi.delivery_order_id").
		Where("dord.sales_order_id = ?", salesOrderID).
		Where("UPPER(dord.status) NOT IN ?", []string{
			strings.ToUpper(string(models.DeliveryOrderStatusCancelled)),
			strings.ToUpper(string(models.DeliveryOrderStatusDelivered)),
			strings.ToUpper(string(models.DeliveryOrderStatusRejected)),
		}).
		Where("dord.deleted_at IS NULL").
		Where("doi.deleted_at IS NULL").
		Where("doi.sales_order_item_id IS NOT NULL").
		Group("doi.sales_order_item_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	pendingMap := make(map[string]float64, len(results))
	for _, r := range results {
		pendingMap[r.SalesOrderItemID] = r.TotalQty
	}
	return pendingMap, nil
}
