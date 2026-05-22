package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/stock_opname/data/models"
	"github.com/gilabs/gims/api/internal/stock_opname/domain/dto"
	"github.com/gilabs/gims/api/internal/stock_opname/domain/repository"
	"gorm.io/gorm"
)

type stockOpnameRepository struct {
	db *gorm.DB
}

const whereStockOpnameID = "stock_opname_id = ?"

func NewStockOpnameRepository(db *gorm.DB) repository.StockOpnameRepository {
	return &stockOpnameRepository{db: db}
}

func (r *stockOpnameRepository) Create(ctx context.Context, opname *models.StockOpname) error {
	return database.GetDB(ctx, r.db).Session(&gorm.Session{NewDB: true}).Create(opname).Error
}

func (r *stockOpnameRepository) Update(ctx context.Context, opname *models.StockOpname) error {
	return database.GetDB(ctx, r.db).Save(opname).Error
}

func (r *stockOpnameRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.StockOpname{}, "id = ?", id).Error
}

func (r *stockOpnameRepository) FindByID(ctx context.Context, id string) (*models.StockOpname, error) {
	var opname models.StockOpname
	err := database.GetDB(ctx, r.db).
		Preload("Warehouse").
		Preload("Items").
		Preload("Warehouse").
		Preload("OrderedBy").
		Preload("AssignedTo").
		First(&opname, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &opname, nil
}

func (r *stockOpnameRepository) List(ctx context.Context, req *dto.ListStockOpnamesRequest) ([]models.StockOpname, int64, error) {
	var opnames []models.StockOpname
	var total int64

	query := database.GetDB(ctx, r.db).
		Model(&models.StockOpname{}).
		Preload("Warehouse").
		Preload("Items").
		Preload("OrderedBy").
		Preload("AssignedTo")

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.DefaultScopeQueryOptions())

	// For OUTLET-scoped users (user_warehouses assignment), restrict to their warehouses.
	// Only applies when no explicit warehouse_id filter is provided by the caller.
	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	if len(warehouseIDs) > 0 && req.WarehouseID == "" {
		query = query.Where("warehouse_id IN ?", warehouseIDs)
	}

	if req.Search != "" {
		search := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(opname_number) LIKE ? OR LOWER(description) LIKE ?", search, search)
	}

	if req.WarehouseID != "" {
		query = query.Where("warehouse_id = ?", req.WarehouseID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", normalizeStatusFilter(req.Status))
	}

	if req.StartDate != "" {
		query = query.Where("date >= ?", req.StartDate)
	}

	if req.EndDate != "" {
		query = query.Where("date <= ?", req.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PerPage
	err := query.Order("created_at DESC").
		Limit(req.PerPage).
		Offset(offset).
		Find(&opnames).Error

	return opnames, total, err
}

func (r *stockOpnameRepository) ReplaceItems(ctx context.Context, opnameID string, items []models.StockOpnameItem) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Delete existing items
		if err := tx.Where(whereStockOpnameID, opnameID).Delete(&models.StockOpnameItem{}).Error; err != nil {
			return err
		}

		// Create new items
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		// Update stats on Opname
		var totalVariance float64
		for _, item := range items {
			totalVariance += item.VarianceQty
		}

		if err := tx.Model(&models.StockOpname{}).Where("id = ?", opnameID).Updates(map[string]interface{}{
			"total_items":        len(items),
			"total_variance_qty": totalVariance,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *stockOpnameRepository) ListItems(ctx context.Context, opnameID string) ([]models.StockOpnameItem, error) {
	var items []models.StockOpnameItem
	err := database.GetDB(ctx, r.db).
		Where(whereStockOpnameID, opnameID).
		Preload("Product").
		Find(&items).Error
	return items, err
}

func (r *stockOpnameRepository) ListItemsPaginated(ctx context.Context, opnameID string, page, perPage int) ([]models.StockOpnameItem, int64, error) {
	var items []models.StockOpnameItem
	var total int64

	query := database.GetDB(ctx, r.db).
		Model(&models.StockOpnameItem{}).
		Where(whereStockOpnameID, opnameID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := query.
		Preload("Product").
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&items).Error

	return items, total, err
}

func (r *stockOpnameRepository) ListWarehouseStockSnapshot(ctx context.Context, warehouseID, scopeType string, categoryIDs, brandIDs []string) ([]dto.WarehouseStockSnapshot, error) {
	var rows []dto.WarehouseStockSnapshot

	query := r.db.WithContext(ctx).
		Table("inventory_batches ib").
		Select("ib.product_id AS product_id, COALESCE(ib.current_quantity - ib.reserved_quantity, 0) AS system_qty, ib.id AS batch_id, ib.batch_number AS batch_number, ib.current_quantity AS batch_qty").
		Joins("JOIN products p ON p.id = ib.product_id AND p.deleted_at IS NULL").
		Where("ib.warehouse_id = ? AND ib.deleted_at IS NULL", warehouseID)

	if tid, ok := ctx.Value("tenant_id").(string); ok && tid != "" {
		query = query.Where("ib.tenant_id = ? AND p.tenant_id = ?", tid, tid)
	}

	switch strings.ToLower(strings.TrimSpace(scopeType)) {
	case "category":
		if len(categoryIDs) > 0 {
			query = query.Where("p.category_id IN ?", categoryIDs)
		}
	case "brand":
		if len(brandIDs) > 0 {
			query = query.Where("p.brand_id IN ?", brandIDs)
		}
	}

	err := query.
		Where("COALESCE(ib.current_quantity - ib.reserved_quantity, 0) <> 0").
		Order("ib.product_id ASC, ib.batch_number ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *stockOpnameRepository) UpdateStatus(ctx context.Context, id string, status models.StockOpnameStatus, userID *string) error {
	updates := map[string]interface{}{"status": status, "updated_at": apptime.Now()}
	if userID != nil {
		updates["updated_by"] = *userID
	}
	return database.GetDB(ctx, r.db).Model(&models.StockOpname{}).Where("id = ?", id).Updates(updates).Error
}

func (r *stockOpnameRepository) GetNextOpnameNumber(ctx context.Context) (string, error) {
	prefix := fmt.Sprintf("OP-%s-", apptime.Now().Format("200601"))
	var lastOpname models.StockOpname
	err := database.GetDB(ctx, r.db).
		Where("opname_number LIKE ?", prefix+"%").
		Order("opname_number DESC").
		First(&lastOpname).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	seq := 1
	if err == nil {
		// Parse last number
		parts := strings.Split(lastOpname.OpnameNumber, "-")
		if len(parts) == 3 {
			var lastSeq int
			fmt.Sscanf(parts[2], "%d", &lastSeq)
			seq = lastSeq + 1
		}
	}

	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

func normalizeStatusFilter(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending_approval":
		return string(models.StockOpnameStatusPending)
	case "completed":
		return string(models.StockOpnameStatusPosted)
	default:
		return status
	}
}

// GetMyWarehouses returns the warehouses assigned to a specific user via user_warehouses.
func (r *stockOpnameRepository) GetMyWarehouses(ctx context.Context, userID string) ([]dto.UserWarehouseInfo, error) {
	type row struct {
		ID   string
		Name string
		Code string
	}
	var rows []row
	query := r.db.WithContext(ctx).
		Table("user_warehouses uw").
		Joins("JOIN warehouses w ON w.id = uw.warehouse_id AND w.deleted_at IS NULL").
		Where("uw.user_id = ? AND uw.deleted_at IS NULL", userID)

	if tid, ok := ctx.Value("tenant_id").(string); ok && tid != "" {
		query = query.Where("w.tenant_id = ?", tid)
	}

	err := query.
		Select("w.id, w.name, w.code").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserWarehouseInfo, len(rows))
	for i, rw := range rows {
		result[i] = dto.UserWarehouseInfo{ID: rw.ID, Name: rw.Name, Code: rw.Code}
	}
	return result, nil
}
