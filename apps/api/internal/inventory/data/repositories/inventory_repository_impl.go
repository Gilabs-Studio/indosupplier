package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inventoryRepository struct {
	db *gorm.DB
}

func (r *inventoryRepository) DB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func NewInventoryRepository(db *gorm.DB) repository.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) GetStockList(ctx context.Context, req *dto.GetInventoryListRequest) ([]dto.InventoryStockItem, int64, error) {
	var items []dto.InventoryStockItem
	var total int64

	// Base query on Products to ensure we show products even with 0 stock (if desired)
	// OR query on InventoryBatches and group by Product + Warehouse.
	// Requirement: show stock inventory with warehouse stats.

	// Complex query to aggregate batches
	// We'll select from Products and left join aggregated Batches

	query := r.db.WithContext(ctx).Table("products p").
		Select(`
			p.id as product_id,
			p.code as product_code,
			p.name as product_name,
			p.image_url as product_image_url,
			pc.name as product_category,
			pb.name as product_brand,
			w.id as warehouse_id,
			w.name as warehouse_name,
			COALESCE(SUM(ib.current_quantity), 0) as on_hand,
			COALESCE(SUM(ib.reserved_quantity), 0) as reserved,
			COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) as available,
			COALESCE(SUM(CASE WHEN ib.expiry_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '30 days') AND ib.current_quantity > 0 THEN 1 ELSE 0 END), 0) > 0 as has_expiring_batches,
			p.min_stock,
			p.max_stock,
			u.name as uom_name,
			p.is_ingredient
		`).
		Joins("LEFT JOIN product_categories pc ON pc.id = p.category_id").
		Joins("LEFT JOIN product_brands pb ON pb.id = p.brand_id").
		Joins("LEFT JOIN units_of_measure u ON u.id = p.uom_id")

	// Apply tenant filter manually since we are using joins and custom Table() call
	var err error
	query, err = applyTenantFilter(ctx, query, "p.tenant_id")
	if err != nil {
		return nil, 0, err
	}

	if req.WarehouseID != "" {
		query = query.
			Joins("LEFT JOIN inventory_batches ib ON ib.product_id = p.id AND ib.deleted_at IS NULL AND ib.warehouse_id = ?", req.WarehouseID).
			Joins("LEFT JOIN warehouses w ON w.id = ?", req.WarehouseID)
		
		query, err = applyTenantFilter(ctx, query, "ib.tenant_id", "w.tenant_id")
		if err != nil {
			return nil, 0, err
		}
	} else {
		query = query.
			Joins("LEFT JOIN inventory_batches ib ON ib.product_id = p.id AND ib.deleted_at IS NULL").
			Joins("LEFT JOIN warehouses w ON w.id = ib.warehouse_id").
			Where("ib.id IS NOT NULL")
		
		query, err = applyTenantFilter(ctx, query, "ib.tenant_id", "w.tenant_id")
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply Product Filter
	if req.ProductID != "" {
		query = query.Where("p.id = ?", req.ProductID)
	}

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("p.name ILIKE ? OR p.code ILIKE ?", search, search)
	}

	// Apply ingredient filter
	if req.IsIngredient != nil {
		query = query.Where("p.is_ingredient = ?", *req.IsIngredient)
	}

	query = query.Group("p.id, p.code, p.name, p.image_url, pc.name, pb.name, w.id, w.name, p.min_stock, p.max_stock, u.name, p.is_ingredient")

	// --- Status / expiry HAVING filters ---
	// Resolve effective status filter: explicit Status takes precedence; LowStock is legacy shorthand
	effectiveStatus := req.Status
	if effectiveStatus == "" && req.LowStock {
		effectiveStatus = "low_stock"
	}

	available := "COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0)"

	switch effectiveStatus {
	case "out_of_stock":
		query = query.Having(available + " <= 0")
	case "low_stock":
		query = query.Having(available + " > 0 AND " + available + " <= p.min_stock")
	case "overstock":
		query = query.Having("p.max_stock > 0 AND " + available + " > p.max_stock")
	case "ok":
		query = query.Having(available + " > p.min_stock AND (p.max_stock = 0 OR " + available + " <= p.max_stock)")
	}

	if req.HasExpiring {
		query = query.Having("COALESCE(SUM(CASE WHEN ib.expiry_date BETWEEN CURRENT_DATE AND (CURRENT_DATE + INTERVAL '30 days') AND ib.current_quantity > 0 THEN 1 ELSE 0 END), 0) > 0")
	}

	if req.HasExpired {
		query = query.Having("COALESCE(SUM(CASE WHEN ib.expiry_date < CURRENT_DATE AND ib.current_quantity > 0 THEN 1 ELSE 0 END), 0) > 0")
	}

	// Count Total (wrapping group-by query in subquery for accuracy)
	if err := r.db.WithContext(ctx).Table("(?) as sub", query).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (req.Page - 1) * req.PerPage
	query = query.Limit(req.PerPage).Offset(offset)

	// Order: critical-first when filtering by low/OOS status; otherwise alphabetical
	if effectiveStatus == "low_stock" || effectiveStatus == "out_of_stock" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Raw: true, Name: "(" + available + " / NULLIF(p.max_stock, 0))"},
			Desc:   false,
		}).Order("p.name ASC")
	} else {
		query = query.Order("p.name ASC, w.name ASC")
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	// Calculate Status in Go (easier than SQL)
	for i := range items {
		if items[i].Available <= 0 {
			items[i].Status = "out_of_stock"
		} else if items[i].Available <= items[i].MinStock {
			items[i].Status = "low_stock"
		} else if items[i].MaxStock > 0 && items[i].Available > items[i].MaxStock {
			items[i].Status = "overstock"
		} else {
			items[i].Status = "ok"
		}
	}

	return items, total, nil
}

// GetTreeWarehouses returns a list of warehouses with stock summary
func (r *inventoryRepository) GetTreeWarehouses(ctx context.Context) ([]dto.GetInventoryTreeWarehousesResponse, error) {
	var result []dto.GetInventoryTreeWarehousesResponse

	// Query to get warehouses and aggregated stock stats
	// Status logic matches GetStockList:
	// Out of Stock: available <= 0
	// Low Stock: available <= min_stock AND available > 0
	// Overstock: available > max_stock AND max_stock > 0
	// OK: otherwise

	// We start from warehouses to ensure we list all active warehouses (or all relevant ones)
	// Then left join batches to calculate stats.

	// Note: status logic depends on Product metadata (min/max stock).
	// So we need to join products via batches.

	// This is a heavy query. In production, this might be a materialized view or cached.
	// For now, we compute on the fly.

	// Subquery to get product-warehouse availability first
	// (Same as GetStockList core query but grouped by product+warehouse)

	tenantID, scoped, err := tenantContext(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		WITH stock_levels AS (
			SELECT 
				w.id as warehouse_id,
				p.id as product_id,
				p.min_stock,
				p.max_stock,
				COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) as available
			FROM warehouses w
			JOIN inventory_batches ib ON ib.warehouse_id = w.id AND ib.deleted_at IS NULL
			JOIN products p ON p.id = ib.product_id
			WHERE w.deleted_at IS NULL
			GROUP BY w.id, p.id, p.min_stock, p.max_stock
		)
		SELECT 
			w.id,
			w.name,
			COUNT(sl.product_id) as total_items,
			COUNT(CASE 
				WHEN sl.available <= 0 THEN 1 
			END) as out_of_stock,
			COUNT(CASE 
				WHEN sl.available > 0 AND sl.available <= sl.min_stock THEN 1 
			END) as low,
			COUNT(CASE 
				WHEN sl.max_stock > 0 AND sl.available > sl.max_stock THEN 1 
			END) as overstock,
			COUNT(CASE 
				WHEN sl.available > sl.min_stock AND (sl.max_stock = 0 OR sl.available <= sl.max_stock) THEN 1 
			END) as ok
		FROM warehouses w
		LEFT JOIN stock_levels sl ON sl.warehouse_id = w.id
		WHERE w.deleted_at IS NULL
		GROUP BY w.id, w.name
		ORDER BY w.name ASC
	`

	args := make([]any, 0, 4)
	if scoped {
		query = `
		WITH stock_levels AS (
			SELECT 
				w.id as warehouse_id,
				p.id as product_id,
				p.min_stock,
				p.max_stock,
				COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) as available
			FROM warehouses w
			JOIN inventory_batches ib ON ib.warehouse_id = w.id AND ib.deleted_at IS NULL
			JOIN products p ON p.id = ib.product_id
			WHERE w.tenant_id = ? AND ib.tenant_id = ? AND p.tenant_id = ?
			  AND w.deleted_at IS NULL
			GROUP BY w.id, p.id, p.min_stock, p.max_stock
		)
		SELECT 
			w.id,
			w.name,
			COUNT(sl.product_id) as total_items,
			COUNT(CASE 
				WHEN sl.available <= 0 THEN 1 
			END) as out_of_stock,
			COUNT(CASE 
				WHEN sl.available > 0 AND sl.available <= sl.min_stock THEN 1 
			END) as low,
			COUNT(CASE 
				WHEN sl.max_stock > 0 AND sl.available > sl.max_stock THEN 1 
			END) as overstock,
			COUNT(CASE 
				WHEN sl.available > sl.min_stock AND (sl.max_stock = 0 OR sl.available <= sl.max_stock) THEN 1 
			END) as ok
		FROM warehouses w
		LEFT JOIN stock_levels sl ON sl.warehouse_id = w.id
		WHERE w.tenant_id = ? AND w.deleted_at IS NULL
		GROUP BY w.id, w.name
		ORDER BY w.name ASC
		`
		args = append(args, tenantID, tenantID, tenantID, tenantID)
	} else {
		query = `
		WITH stock_levels AS (
			SELECT 
				w.id as warehouse_id,
				p.id as product_id,
				p.min_stock,
				p.max_stock,
				COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) as available
			FROM warehouses w
			JOIN inventory_batches ib ON ib.warehouse_id = w.id AND ib.deleted_at IS NULL
			JOIN products p ON p.id = ib.product_id
			WHERE w.deleted_at IS NULL
			GROUP BY w.id, p.id, p.min_stock, p.max_stock
		)
		SELECT 
			w.id,
			w.name,
			COUNT(sl.product_id) as total_items,
			COUNT(CASE 
				WHEN sl.available <= 0 THEN 1 
			END) as out_of_stock,
			COUNT(CASE 
				WHEN sl.available > 0 AND sl.available <= sl.min_stock THEN 1 
			END) as low,
			COUNT(CASE 
				WHEN sl.max_stock > 0 AND sl.available > sl.max_stock THEN 1 
			END) as overstock,
			COUNT(CASE 
				WHEN sl.available > sl.min_stock AND (sl.max_stock = 0 OR sl.available <= sl.max_stock) THEN 1 
			END) as ok
		FROM warehouses w
		LEFT JOIN stock_levels sl ON sl.warehouse_id = w.id
		WHERE w.deleted_at IS NULL
		GROUP BY w.id, w.name
		ORDER BY w.name ASC
		`
	}

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item dto.GetInventoryTreeWarehousesResponse
		var total, oos, low, over, ok int

		if err := rows.Scan(&item.ID, &item.Name, &total, &oos, &low, &over, &ok); err != nil {
			return nil, err
		}

		item.Summary = dto.StockSummary{
			TotalItems: total,
			OutOfStock: oos,
			Low:        low,
			Overstock:  over,
			Ok:         ok,
		}
		result = append(result, item)
	}

	return result, nil
}

// GetTreeProducts returns products for a specific warehouse
func (r *inventoryRepository) GetTreeProducts(ctx context.Context, req *dto.GetInventoryTreeProductsRequest) ([]dto.InventoryStockItem, int64, dto.TreeProductsSummary, error) {
	var items []dto.InventoryStockItem
	var total int64
	var summary dto.TreeProductsSummary

	// Filter by WarehouseID IS REQUIRED and enforced by logic calling this
	query := r.db.WithContext(ctx).Table("products p").
		Select(`
			p.id as product_id,
			p.code as product_code,
			p.name as product_name,
			p.image_url as product_image_url,
			pc.name as product_category,
			pb.name as product_brand,
			w.id as warehouse_id,
			w.name as warehouse_name,
			COALESCE(SUM(ib.current_quantity), 0) as on_hand,
			COALESCE(SUM(ib.reserved_quantity), 0) as reserved,
			COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) as available,
			p.min_stock,
			p.max_stock,
			u.name as uom_name,
			p.is_ingredient
		`).
		Joins("LEFT JOIN product_categories pc ON pc.id = p.category_id").
		Joins("LEFT JOIN product_brands pb ON pb.id = p.brand_id").
		Joins("LEFT JOIN units_of_measure u ON u.id = p.uom_id").
		Joins("JOIN inventory_batches ib ON ib.product_id = p.id AND ib.deleted_at IS NULL"). // Inner join to filtering batches
		Joins("JOIN warehouses w ON w.id = ib.warehouse_id")

	// Apply tenant filter manually since we are using joins and custom Table() call
	var err error
	query, err = applyTenantFilter(ctx, query, "p.tenant_id", "ib.tenant_id", "w.tenant_id")
	if err != nil {
		return nil, 0, summary, err
	}

	// Apply Warehouse ID Filter
	query = query.Where("ib.warehouse_id = ?", req.WarehouseID)

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("p.name ILIKE ? OR p.code ILIKE ?", search, search)
	}

	// Apply ingredient filter
	if req.IsIngredient != nil {
		query = query.Where("p.is_ingredient = ?", *req.IsIngredient)
	}

	query = query.Group("p.id, p.code, p.name, p.image_url, pc.name, pb.name, w.id, w.name, p.min_stock, p.max_stock, u.name, p.is_ingredient")

	// Count Total
	if err := r.db.WithContext(ctx).Table("(?) as sub", query).Count(&total).Error; err != nil {
		return nil, 0, summary, err
	}

	// Calculate summary from full filtered dataset (not paginated page only)
	type treeSummaryRow struct {
		Ok         int `gorm:"column:ok"`
		Low        int `gorm:"column:low"`
		OutOfStock int `gorm:"column:out_of_stock"`
		Overstock  int `gorm:"column:overstock"`
	}

	var summaryRow treeSummaryRow
	if err := r.db.WithContext(ctx).Table("(?) as sub", query).
		Select(`
			COALESCE(SUM(CASE WHEN sub.available <= 0 THEN 1 ELSE 0 END), 0) AS out_of_stock,
			COALESCE(SUM(CASE WHEN sub.available > 0 AND sub.available <= sub.min_stock THEN 1 ELSE 0 END), 0) AS low,
			COALESCE(SUM(CASE WHEN sub.max_stock > 0 AND sub.available > sub.max_stock THEN 1 ELSE 0 END), 0) AS overstock,
			COALESCE(SUM(CASE WHEN sub.available > sub.min_stock AND (sub.max_stock <= 0 OR sub.available <= sub.max_stock) THEN 1 ELSE 0 END), 0) AS ok
		`).
		Scan(&summaryRow).Error; err != nil {
		return nil, 0, summary, err
	}

	summary = dto.TreeProductsSummary{
		TotalItems: int(total),
		Ok:         summaryRow.Ok,
		Low:        summaryRow.Low,
		OutOfStock: summaryRow.OutOfStock,
		Overstock:  summaryRow.Overstock,
	}

	// Pagination
	offset := (req.Page - 1) * req.PerPage
	query = query.Limit(req.PerPage).Offset(offset)

	// Order
	query = query.Order("p.name ASC")

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, summary, err
	}

	// Calculate Status
	for i := range items {
		if items[i].Available <= 0 {
			items[i].Status = "out_of_stock"
		} else if items[i].Available <= items[i].MinStock {
			items[i].Status = "low_stock"
		} else if items[i].MaxStock > 0 && items[i].Available > items[i].MaxStock {
			items[i].Status = "overstock"
		} else {
			items[i].Status = "ok"
		}
	}

	return items, total, summary, nil
}

// GetTreeBatches returns batches for a specific product and warehouse with pagination
func (r *inventoryRepository) GetTreeBatches(ctx context.Context, req *dto.GetInventoryTreeBatchesRequest) ([]dto.InventoryBatchItem, int64, error) {
	var items []dto.InventoryBatchItem
	var total int64

	// Default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	base := r.DB(ctx).Table("inventory_batches ib").
		Where("ib.deleted_at IS NULL").
		Where("ib.warehouse_id = ?", req.WarehouseID).
		Where("ib.product_id = ?", req.ProductID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PerPage
	query := r.DB(ctx).Table("inventory_batches ib").
		Select(`
			ib.id,
			ib.batch_number,
			ib.expiry_date,
			ib.current_quantity,
			ib.reserved_quantity,
			(ib.current_quantity - ib.reserved_quantity) as available
		`).
		Where("ib.deleted_at IS NULL").
		Where("ib.warehouse_id = ?", req.WarehouseID).
		Where("ib.product_id = ?", req.ProductID).
		Order("ib.created_at DESC, ib.id DESC").
		Limit(req.PerPage).
		Offset(offset)

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetInventoryMetrics returns high-level inventory summary for owner/admin dashboards
func (r *inventoryRepository) GetInventoryMetrics(ctx context.Context) (*dto.InventoryMetrics, error) {
	permissionScope, _ := ctx.Value("permission_scope").(string)
	permissionScope = strings.ToUpper(strings.TrimSpace(permissionScope))
	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	warehouseScoped := permissionScope == "WAREHOUSE"

	if warehouseScoped && len(warehouseIDs) == 0 {
		return &dto.InventoryMetrics{}, nil
	}

	// CTE to aggregate product-warehouse stock levels
	tenantID, scoped, err := tenantContext(ctx)
	if err != nil {
		return nil, err
	}
	// CTE to aggregate product-warehouse stock levels
	metricsQuery := `
		WITH stock_levels AS (
			SELECT 
				w.id  AS warehouse_id,
				p.id  AS product_id,
				p.min_stock,
				p.max_stock,
				COALESCE(SUM(ib.current_quantity), 0) AS on_hand,
				COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) AS available
			FROM products p
			JOIN inventory_batches ib ON ib.product_id = p.id AND ib.deleted_at IS NULL
			JOIN warehouses w ON w.id = ib.warehouse_id AND w.deleted_at IS NULL
			WHERE p.deleted_at IS NULL
	`
	metricsArgs := make([]any, 0, 3)
	if scoped {
		metricsQuery += "\n\t\t\tAND p.tenant_id = ? AND ib.tenant_id = ? AND w.tenant_id = ?"
		metricsArgs = append(metricsArgs, tenantID, tenantID, tenantID)
	}
	if warehouseScoped {
		metricsQuery += "\n\t\t\tAND w.id IN ?"
		metricsArgs = append(metricsArgs, warehouseIDs)
	}
	metricsQuery += `
			GROUP BY w.id, p.id, p.min_stock, p.max_stock
		)
		SELECT
			COUNT(*)                                                                     AS total_items,
			COUNT(DISTINCT product_id)                                                   AS total_products,
			COUNT(DISTINCT warehouse_id)                                                 AS total_warehouses,
			COALESCE(SUM(on_hand), 0)                                                    AS total_on_hand,
			COUNT(CASE WHEN available <= 0                                   THEN 1 END) AS out_of_stock_count,
			COUNT(CASE WHEN available > 0 AND available <= min_stock         THEN 1 END) AS low_stock_count,
			COUNT(CASE WHEN max_stock > 0 AND available > max_stock          THEN 1 END) AS overstock_count,
			COUNT(CASE WHEN available > min_stock AND (max_stock = 0 OR available <= max_stock) THEN 1 END) AS ok_count
		FROM stock_levels
	`

	type rawMetrics struct {
		TotalItems      int     `json:"total_items"`
		TotalProducts   int     `json:"total_products"`
		TotalWarehouses int     `json:"total_warehouses"`
		TotalOnHand     float64 `json:"total_on_hand"`
		OutOfStockCount int     `json:"out_of_stock_count"`
		LowStockCount   int     `json:"low_stock_count"`
		OverstockCount  int     `json:"overstock_count"`
		OkCount         int     `json:"ok_count"`
	}

	var raw rawMetrics
	if err := r.db.WithContext(ctx).Raw(metricsQuery, metricsArgs...).Scan(&raw).Error; err != nil {
		return nil, err
	}

	// Expiring and expired batches counts
	var expiringCount int64
	var expiredCount int64

	expiringQuery := r.DB(ctx).Table("inventory_batches").
		Where("deleted_at IS NULL AND current_quantity > 0 AND expiry_date >= CURRENT_DATE AND expiry_date <= CURRENT_DATE + INTERVAL '30 days'")
	if warehouseScoped {
		expiringQuery = expiringQuery.Where("warehouse_id IN ?", warehouseIDs)
	}
	expiringQuery.Count(&expiringCount)

	expiredQuery := r.DB(ctx).Table("inventory_batches").
		Where("deleted_at IS NULL AND current_quantity > 0 AND expiry_date < CURRENT_DATE")
	if warehouseScoped {
		expiredQuery = expiredQuery.Where("warehouse_id IN ?", warehouseIDs)
	}
	expiredQuery.Count(&expiredCount)

	return &dto.InventoryMetrics{
		TotalItems:           raw.TotalItems,
		TotalProducts:        raw.TotalProducts,
		TotalWarehouses:      raw.TotalWarehouses,
		TotalOnHand:          raw.TotalOnHand,
		OkCount:              raw.OkCount,
		LowStockCount:        raw.LowStockCount,
		OutOfStockCount:      raw.OutOfStockCount,
		OverstockCount:       raw.OverstockCount,
		ExpiringBatches30Day: int(expiringCount),
		ExpiredBatches:       int(expiredCount),
	}, nil
}

// Stock Management Implementations

// UpdateProductReservedStock updates the reserved stock counter on the Product
func (r *inventoryRepository) UpdateProductReservedStock(ctx context.Context, productID string, quantity float64) error {
	// Note: We are updating the Product table directly as per plan (Soft Reservation)
	// Even though GetStockList sums batch reserved_quantities, we might need to adjust that query
	// or ensure we distribute reservation to batches later.
	// HOWEVER, if the system design implies Product-level reservation BEFORE batch selection,
	// then we must have a ReservedStock field on Product.

	// Let's assume Product model has ReservedStock or we add it.
	// If it doesn't, we might fail here.
	// Based on typical GORM, we can use an expression.

	return r.DB(ctx).Table("products").Where("id = ?", productID).
		Update("reserved_stock", gorm.Expr("COALESCE(reserved_stock, 0) + ?", quantity)).Error
}

func (r *inventoryRepository) UpdateBatchQuantity(ctx context.Context, batchID string, quantity float64) error {
	// quantity is the change (delta). If negative, it deducts.
	return r.DB(ctx).Table("inventory_batches").Where("id = ?", batchID).
		Update("current_quantity", gorm.Expr("current_quantity + ?", quantity)).Error
}

func (r *inventoryRepository) GetBatchesByProduct(ctx context.Context, productID string) ([]dto.InventoryBatchItem, error) {
	var items []dto.InventoryBatchItem

	query := r.DB(ctx).Table("inventory_batches ib").
		Select(`
			ib.id,
			ib.batch_number,
			ib.expiry_date,
			ib.created_at as received_at,
			ib.current_quantity,
			ib.reserved_quantity,
			(ib.current_quantity - ib.reserved_quantity) as available
		`).
		Where("ib.deleted_at IS NULL").
		Where("ib.product_id = ?", productID).
		Where("ib.current_quantity > 0"). // Only available batches? Logic says "SelectBatches"
		Order("ib.created_at ASC")        // Default sort

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *inventoryRepository) CreateStockMovement(ctx context.Context, req *dto.StockMovementRequest) (string, error) {
	batchID := strings.TrimSpace(req.InventoryBatchID)
	var inventoryBatchID *string
	if batchID != "" {
		inventoryBatchID = &batchID
	}

	movement := models.StockMovement{
		InventoryBatchID: inventoryBatchID,
		ProductID:        req.ProductID,
		WarehouseID:      req.WarehouseID,
		MovementType:     models.StockMovementType(req.Type),
		RefType:          reference.Normalize(req.ReferenceType),
		RefID:            req.ReferenceID,
		RefNumber:        req.ReferenceNumber,
		Source:           strings.TrimSpace(req.Source),
		CreatedBy:        req.CreatedBy,
		Date:             apptime.Now(),
	}
	if movement.Source == "" {
		movement.Source = req.Description
	}

	// Costing: Ensure we use the latest Product HPP if not provided on the request
	// (usually it's on the request for GR, but for Adjust/DO we get it here)
	movement.Cost = req.Cost
	if movement.Cost == 0 {
		var productHPP float64
		r.DB(ctx).Table("products").Select("COALESCE(current_hpp, 0)").Where("id = ?", req.ProductID).Scan(&productHPP)
		movement.Cost = productHPP
	}

	switch {
	case req.Type == "IN":
		movement.QtyIn = req.Quantity
	case req.Type == "ADJUST" && req.Quantity > 0:
		// Surplus: physical count > system → stock increases
		movement.QtyIn = req.Quantity
	case req.Type == "ADJUST" && req.Quantity < 0:
		// Shortage: physical count < system → stock decreases
		movement.QtyOut = -req.Quantity // Store as positive value
	case req.Type == "TRANSFER" && req.MovementDirection == "IN":
		movement.QtyIn = req.Quantity
	case req.Type == "TRANSFER" && req.MovementDirection == "OUT":
		movement.QtyOut = req.Quantity
	default:
		// OUT or TRANSFER
		movement.QtyOut = req.Quantity
	}

	if err := r.DB(ctx).Create(&movement).Error; err != nil {
		return "", err
	}
	return movement.ID, nil
}

func (r *inventoryRepository) CreateBatch(ctx context.Context, item *dto.CreateBatchParams) (string, error) {
	batch := models.InventoryBatch{
		ProductID:       item.ProductID,
		WarehouseID:     item.WarehouseID,
		BatchNumber:     item.BatchNumber,
		ExpiryDate:      item.ExpiryDate,
		InitialQuantity: item.InitialQuantity,
		CurrentQuantity: item.InitialQuantity,
		CostPrice:       item.CostPrice,
		IsActive:        true,
	}

	if err := r.DB(ctx).Create(&batch).Error; err != nil {
		return "", err
	}
	return batch.ID, nil
}

func (r *inventoryRepository) UpdateProductAverageCost(ctx context.Context, productID string, newCost float64) error {
	return r.DB(ctx).Table("products").Where("id = ?", productID).Update("current_hpp", newCost).Error
}

func (r *inventoryRepository) UpdateProductStock(ctx context.Context, productID string, delta float64) error {
	return r.DB(ctx).Table("products").Where("id = ?", productID).
		Update("current_stock", gorm.Expr("COALESCE(current_stock, 0) + ?", delta)).Error
}

func (r *inventoryRepository) GetProductCostInfo(ctx context.Context, productID string) (float64, float64, error) {
	type costRes struct {
		CurrentHpp   float64
		CurrentStock float64
	}
	var res costRes
	if err := r.DB(ctx).Table("products").
		Select("COALESCE(current_hpp, 0) as current_hpp, COALESCE(current_stock, 0) as current_stock").
		Where("id = ?", productID).
		Scan(&res).Error; err != nil {
		return 0, 0, err
	}
	return res.CurrentHpp, res.CurrentStock, nil
}

// GetBatchByID fetches a single batch with computed available quantity
func (r *inventoryRepository) GetBatchByID(ctx context.Context, batchID string) (*dto.InventoryBatchDetail, error) {
	var item dto.InventoryBatchDetail

	err := r.DB(ctx).Table("inventory_batches ib").
		Select(`
			ib.id,
			ib.product_id,
			ib.warehouse_id,
			ib.batch_number,
			ib.expiry_date,
			ib.current_quantity,
			ib.reserved_quantity,
			ib.cost_price,
			(ib.current_quantity - ib.reserved_quantity) as available,
			ib.is_active
		`).
		Where("ib.deleted_at IS NULL").
		Where("ib.id = ?", batchID).
		First(&item).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

// UpdateBatchReservedQuantity adjusts the reserved_quantity on an inventory batch (delta-based)
func (r *inventoryRepository) UpdateBatchReservedQuantity(ctx context.Context, batchID string, quantity float64) error {
	return r.DB(ctx).Table("inventory_batches").Where("id = ?", batchID).
		Update("reserved_quantity", gorm.Expr("COALESCE(reserved_quantity, 0) + ?", quantity)).Error
}

// GetBatchesByProductAndWarehouse returns all active batches for a product in a specific warehouse (FIFO order)
func (r *inventoryRepository) GetBatchesByProductAndWarehouse(ctx context.Context, productID, warehouseID string) ([]dto.InventoryBatchItem, error) {
	var items []dto.InventoryBatchItem

	query := r.DB(ctx).Table("inventory_batches ib").
		Select(`
			ib.id,
			ib.batch_number,
			ib.expiry_date,
			ib.created_at as received_at,
			ib.current_quantity,
			ib.reserved_quantity,
			(ib.current_quantity - ib.reserved_quantity) as available
		`).
		Where("ib.deleted_at IS NULL").
		Where("ib.product_id = ?", productID).
		Where("ib.warehouse_id = ?", warehouseID).
		Where("ib.is_active = true").
		Order("ib.created_at ASC")

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
