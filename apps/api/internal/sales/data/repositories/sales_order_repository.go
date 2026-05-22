package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesOrderRepository defines the interface for sales order data access
type SalesOrderRepository interface {
	FindByID(ctx context.Context, id string) (*models.SalesOrder, error)
	FindByCode(ctx context.Context, code string) (*models.SalesOrder, error)
	List(ctx context.Context, req *dto.ListSalesOrdersRequest) ([]models.SalesOrder, int64, error)
	ListItems(ctx context.Context, orderID string, req *dto.ListSalesOrderItemsRequest) ([]models.SalesOrderItem, int64, error)
	Create(ctx context.Context, so *models.SalesOrder) error
	Update(ctx context.Context, so *models.SalesOrder) error
	Delete(ctx context.Context, id string) error
	GetNextOrderNumber(ctx context.Context, prefix string) (string, error)
	UpdateStatus(ctx context.Context, id string, status models.SalesOrderStatus, userID *string, reason *string) error
	ReserveStock(ctx context.Context, orderID string) error
	ReleaseStock(ctx context.Context, orderID string) error
	UpdateItemDeliveredQty(ctx context.Context, itemID string, qty float64) error
	UpdateItemInvoicedQty(ctx context.Context, itemID string, qty float64) error
}

type salesOrderRepository struct {
	db *gorm.DB
}

// NewSalesOrderRepository creates a new SalesOrderRepository
func NewSalesOrderRepository(db *gorm.DB) SalesOrderRepository {
	return &salesOrderRepository{db: db}
}

func (r *salesOrderRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *salesOrderRepository) FindByID(ctx context.Context, id string) (*models.SalesOrder, error) {
	var order models.SalesOrder
	err := r.getDB(ctx).
		Preload("Customer").
		Preload("SalesQuotation").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Preload("DeliveryArea").
		Preload("Items.Product").
		Preload("DeliveryOrders", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "delivery_date", "is_partial_delivery").Order("delivery_date desc")
		}).
		Preload("CustomerInvoices", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "invoice_date", "due_date", "amount", "paid_amount").Order("invoice_date desc")
		}).
		Where("id = ?", id).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *salesOrderRepository) FindByCode(ctx context.Context, code string) (*models.SalesOrder, error) {
	var order models.SalesOrder
	err := r.getDB(ctx).
		Preload("Customer").
		Preload("SalesQuotation").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Preload("DeliveryArea").
		Preload("Items.Product").
		Preload("DeliveryOrders", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "delivery_date", "is_partial_delivery").Order("delivery_date desc")
		}).
		Preload("CustomerInvoices", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "invoice_date", "due_date", "amount", "paid_amount").Order("invoice_date desc")
		}).
		Where("code = ?", code).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *salesOrderRepository) List(ctx context.Context, req *dto.ListSalesOrdersRequest) ([]models.SalesOrder, int64, error) {
	var orders []models.SalesOrder
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SalesOrder{})
	var err error
	query, err = applyTenantFilter(ctx, query, "sales_orders.tenant_id")
	if err != nil {
		return nil, 0, err
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.SalesScopeQueryOptions())

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = query.Joins("LEFT JOIN employees ON employees.id = sales_orders.sales_rep_id")
		query, err = applyTenantFilter(ctx, query, "employees.tenant_id")
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("sales_orders.customer_name ILIKE ? OR employees.name ILIKE ? OR sales_orders.code ILIKE ? OR sales_orders.notes ILIKE ?", search, search, search, search)
	}

	// Apply status filter
	if req.Status != "" {
		if strings.Contains(req.Status, ",") {
			query = query.Where("status IN ?", strings.Split(req.Status, ","))
		} else {
			query = query.Where("status = ?", req.Status)
		}
	}

	// Apply source type filter (e.g. "POS" for F&B POS orders)
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("order_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("order_date <= ?", req.DateTo)
	}

	// Apply sales rep filter
	if req.SalesRepID != "" {
		query = query.Where("sales_rep_id = ?", req.SalesRepID)
	}

	// Apply business unit filter
	if req.BusinessUnitID != "" {
		query = query.Where("business_unit_id = ?", req.BusinessUnitID)
	}

	// Apply quotation filter
	if req.SalesQuotationID != "" {
		query = query.Where("sales_quotation_id = ?", req.SalesQuotationID)
	}

	// Apply customer filter
	if req.CustomerID != "" {
		query = query.Where("customer_id = ?", req.CustomerID)
	}

	// Apply unfulfilled_only filter
	// Exclude SOs where ALL items have qty fully covered by delivered_quantity + pending DO allocations
	if req.UnfulfilledOnly {
		query = query.Where(`EXISTS (
			SELECT 1 FROM sales_order_items soi
			WHERE soi.sales_order_id = sales_orders.id
			AND soi.quantity > soi.delivered_quantity + COALESCE((
				SELECT SUM(doi.quantity) FROM delivery_order_items doi
				JOIN delivery_orders dord ON dord.id = doi.delivery_order_id
				WHERE doi.sales_order_item_id = soi.id
				AND dord.status != 'cancelled'
				AND dord.deleted_at IS NULL
				AND doi.deleted_at IS NULL
			), 0)
		)`)
	}

	// Exclude SOs that already have an active CIDP (#252)
	if req.ExcludeWithActiveCIDP {
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM customer_invoices ci
			WHERE ci.sales_order_id = sales_orders.id
			AND ci.type = 'down_payment'
			AND ci.status NOT IN ('CANCELLED', 'VOID')
			AND ci.deleted_at IS NULL
		)`)
	}

	// Exclude SOs that already have a paid CI (#243)
	if req.ExcludeWithPaidCI {
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM customer_invoices ci
			WHERE ci.sales_order_id = sales_orders.id
			AND ci.type IN ('regular', 'proforma')
			AND ci.status IN ('PAID')
			AND ci.deleted_at IS NULL
		)`)
	}

	// Exclude SOs that already have a non-cancelled regular/proforma invoice (#239)
	if req.AvailableForInvoice {
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM customer_invoices ci
			WHERE ci.sales_order_id = sales_orders.id
			AND ci.type IN ('regular', 'proforma')
			AND ci.status NOT IN ('CANCELLED', 'VOID')
			AND ci.deleted_at IS NULL
		)`)
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
		"code":          "code",
		"order_date":    "order_date",
		"customer_name": "customer_name",
		"total_amount":  "total_amount",
		"status":        "status",
		"created_at":    "created_at",
		"updated_at":    "updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(req.SortBy))]
	if sortBy == "" {
		sortBy = "order_date"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err = query.
		Preload("Customer").
		Preload("SalesQuotation").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Preload("DeliveryArea").
		Preload("Items.Product").
		Preload("DeliveryOrders", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "delivery_date", "is_partial_delivery").Order("delivery_date desc")
		}).
		Preload("CustomerInvoices", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "sales_order_id", "code", "status", "invoice_date", "due_date", "amount", "paid_amount").Order("invoice_date desc")
		}).
		Limit(perPage).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *salesOrderRepository) Create(ctx context.Context, so *models.SalesOrder) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Create order
		if err := tx.Create(so).Error; err != nil {
			return err
		}

		// Update Sales Quotation status if linked
		if so.SalesQuotationID != nil {
			var sq models.SalesQuotation
			if err := tx.First(&sq, "id = ?", so.SalesQuotationID).Error; err != nil {
				return err
			}

			updates := map[string]interface{}{
				"status":                      models.SalesQuotationStatusConverted,
				"converted_to_sales_order_id": so.ID,
				"converted_at":                apptime.Now(),
			}

			if err := tx.Model(&sq).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *salesOrderRepository) Update(ctx context.Context, so *models.SalesOrder) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order details
		if err := tx.Save(so).Error; err != nil {
			return err
		}

		// Handle Items
		// 1. Identify items to keep (update) and items to delete
		var keepIDs []string
		for _, item := range so.Items {
			if item.ID != "" {
				keepIDs = append(keepIDs, item.ID)
			}
		}

		// 2. Delete items that are not in the new list (Soft Delete)
		if len(keepIDs) > 0 {
			if err := tx.Where("sales_order_id = ? AND id NOT IN ?", so.ID, keepIDs).Delete(&models.SalesOrderItem{}).Error; err != nil {
				return err
			}
		} else {
			// If no items are kept/sent, delete all items for this order
			if err := tx.Where("sales_order_id = ?", so.ID).Delete(&models.SalesOrderItem{}).Error; err != nil {
				return err
			}
		}

		// 3. Upsert items (Update existing, Create new)
		if len(so.Items) > 0 {
			for i := range so.Items {
				so.Items[i].SalesOrderID = so.ID
				// Use Save to upsert (handles both case with ID and without ID)
				if err := tx.Save(&so.Items[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesOrderRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete items first (CASCADE should handle this, but explicit for safety)
		if err := tx.Where("sales_order_id = ?", id).Delete(&models.SalesOrderItem{}).Error; err != nil {
			return err
		}

		// Delete order
		return tx.Delete(&models.SalesOrder{}, "id = ?", id).Error
	})
}

func (r *salesOrderRepository) GetNextOrderNumber(ctx context.Context, prefix string) (string, error) {
	var lastOrder models.SalesOrder
	var sequence int

	// Find the last order with the same prefix
	err := r.getDB(ctx).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		First(&lastOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No previous order, start from 1
			sequence = 1
		} else {
			return "", err
		}
	} else {
		// Extract sequence number from last code
		var count int64
		r.getDB(ctx).Model(&models.SalesOrder{}).
			Where("code LIKE ?", prefix+"%").
			Count(&count)
		sequence = int(count) + 1
	}

	// Generate new code: PREFIX-YYYYMMDD-XXXX
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")

	// Format sequence with 4 digits
	code := prefix + "-" + dateStr + "-" + formatSequence(sequence)

	return code, nil
}

func (r *salesOrderRepository) UpdateStatus(ctx context.Context, id string, status models.SalesOrderStatus, userID *string, reason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case models.SalesOrderStatusApproved:
		updates["confirmed_by"] = userID
		updates["confirmed_at"] = database.GetDB(ctx, r.db).NowFunc()
	case models.SalesOrderStatusCancelled:
		updates["cancelled_by"] = userID
		updates["cancelled_at"] = database.GetDB(ctx, r.db).NowFunc()
		if reason != nil {
			updates["cancellation_reason"] = *reason
		}
	}

	return r.getDB(ctx).Model(&models.SalesOrder{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ReserveStock marks a sales order as having reserved stock.
// Product-level stock reservation is handled by InventoryUsecase.ReserveStock.
func (r *salesOrderRepository) ReserveStock(ctx context.Context, orderID string) error {
	return r.getDB(ctx).Model(&models.SalesOrder{}).
		Where("id = ?", orderID).
		Update("reserved_stock", true).Error
}

// ReleaseStock marks a sales order as no longer having reserved stock.
// Product-level stock release is handled by InventoryUsecase.ReleaseStock.
func (r *salesOrderRepository) ReleaseStock(ctx context.Context, orderID string) error {
	return r.getDB(ctx).Model(&models.SalesOrder{}).
		Where("id = ?", orderID).
		Update("reserved_stock", false).Error
}

// ListItems retrieves order items with pagination
func (r *salesOrderRepository) ListItems(ctx context.Context, orderID string, req *dto.ListSalesOrderItemsRequest) ([]models.SalesOrderItem, int64, error) {
	var items []models.SalesOrderItem
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
	if err := r.getDB(ctx).Model(&models.SalesOrderItem{}).
		Where("sales_order_id = ?", orderID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated items with minimal preload (only product info)
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "image_url")
		}).
		Where("sales_order_id = ?", orderID).
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateItemDeliveredQty updates the delivered quantity of a sales order item
func (r *salesOrderRepository) UpdateItemDeliveredQty(ctx context.Context, itemID string, qty float64) error {
	return r.getDB(ctx).Model(&models.SalesOrderItem{}).
		Where("id = ?", itemID).
		Update("delivered_quantity", gorm.Expr("COALESCE(delivered_quantity, 0) + ?", qty)).Error
}

func (r *salesOrderRepository) UpdateItemInvoicedQty(ctx context.Context, itemID string, qty float64) error {
	return r.getDB(ctx).Model(&models.SalesOrderItem{}).
		Where("id = ?", itemID).
		Update("invoiced_quantity", gorm.Expr("COALESCE(invoiced_quantity, 0) + ?", qty)).Error
}
