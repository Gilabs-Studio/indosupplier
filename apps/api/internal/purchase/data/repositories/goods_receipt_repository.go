package repositories

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GoodsReceiptRepository interface {
	List(ctx context.Context, params GoodsReceiptListParams) ([]*models.GoodsReceipt, int64, error)
	GetByID(ctx context.Context, id string) (*models.GoodsReceipt, error)
	Create(ctx context.Context, gr *models.GoodsReceipt) (*models.GoodsReceipt, error)
	Update(ctx context.Context, gr *models.GoodsReceipt) (*models.GoodsReceipt, error)
	Delete(ctx context.Context, id string) error
}

type GoodsReceiptListParams struct {
	Search      string
	Status      string
	WarehouseID string
	SortBy      string
	SortDir     string
	Limit       int
	Offset      int
}

type goodsReceiptRepository struct {
	db *gorm.DB
}

func NewGoodsReceiptRepository(db *gorm.DB) GoodsReceiptRepository {
	return &goodsReceiptRepository{db: db}
}

func (r *goodsReceiptRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *goodsReceiptRepository) List(ctx context.Context, params GoodsReceiptListParams) ([]*models.GoodsReceipt, int64, error) {
	var results []*models.GoodsReceipt
	var total int64

	base := r.getDB(ctx).Model(&models.GoodsReceipt{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	base = security.ApplyScopeFilter(base, ctx, security.PurchaseScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		pattern := "%" + s + "%"
		base = base.Where("goods_receipts.supplier_name_snapshot ILIKE ? OR goods_receipts.code ILIKE ? OR goods_receipts.notes ILIKE ?", pattern, pattern, pattern)
	}
	if strings.TrimSpace(params.Status) != "" {
		base = base.Where("status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}
	if strings.TrimSpace(params.WarehouseID) != "" {
		base = base.Where("warehouse_id = ?", strings.TrimSpace(params.WarehouseID))
	}
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.getDB(ctx).Model(&models.GoodsReceipt{})

	// Apply scope-based data filtering (must match count query scope)
	query = security.ApplyScopeFilter(query, ctx, security.PurchaseScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		pattern := "%" + s + "%"
		query = query.Where("goods_receipts.supplier_name_snapshot ILIKE ? OR goods_receipts.code ILIKE ? OR goods_receipts.notes ILIKE ?", pattern, pattern, pattern)
	}
	if strings.TrimSpace(params.WarehouseID) != "" {
		query = query.Where("warehouse_id = ?", strings.TrimSpace(params.WarehouseID))
	}
	if strings.TrimSpace(params.Status) != "" {
		query = query.Where("status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"code":         "code",
		"receipt_date": "receipt_date",
		"status":       "status",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "created_at"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	if params.Limit > 0 {
		query = query.Limit(params.Limit).Offset(params.Offset)
	}

	query = query.
		Preload("Supplier").
		Preload("Warehouse").
		Preload("Creator").
		Preload("PurchaseOrder").
		Preload("Items")

	if err := query.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *goodsReceiptRepository) GetByID(ctx context.Context, id string) (*models.GoodsReceipt, error) {
	var gr models.GoodsReceipt
	q := r.getDB(ctx).
		Model(&models.GoodsReceipt{}).
		Preload("Supplier").
		Preload("Warehouse").
		Preload("Creator").
		Preload("PurchaseOrder.Supplier").
		Preload("PurchaseOrder.Items.Product").
		Preload("Items.Product").
		Preload("Items.PurchaseOrderItem.Product")

	// Accept either a UUID (primary key) or a code/reference number.
	if _, err := uuid.Parse(id); err == nil {
		q = q.Where("goods_receipts.id = ?", id)
	} else {
		q = q.Where("goods_receipts.code = ?", id)
	}

	if err := q.First(&gr).Error; err != nil {
		return nil, err
	}
	return &gr, nil
}

func (r *goodsReceiptRepository) Create(ctx context.Context, gr *models.GoodsReceipt) (*models.GoodsReceipt, error) {
	if gr == nil {
		return nil, fmt.Errorf("goods receipt is nil")
	}

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		code, err := r.getNextCodeLocked(ctx, tx, "GR")
		if err != nil {
			return err
		}
		gr.Code = code

		if err := tx.Create(gr).Error; err != nil {
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
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, gr.ID)
}

func (r *goodsReceiptRepository) Update(ctx context.Context, gr *models.GoodsReceipt) (*models.GoodsReceipt, error) {
	if gr == nil {
		return nil, fmt.Errorf("goods receipt is nil")
	}

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.GoodsReceipt
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&existing, "id = ?", gr.ID).Error; err != nil {
			return err
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"warehouse_id":    gr.WarehouseID,
			"notes":           gr.Notes,
			"proof_image_url": gr.ProofImageURL,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("goods_receipt_id = ?", gr.ID).Delete(&models.GoodsReceiptItem{}).Error; err != nil {
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, gr.ID)
}

func (r *goodsReceiptRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("goods_receipt_id = ?", id).Delete(&models.GoodsReceiptItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.GoodsReceipt{}, "id = ?", id).Error
	})
}

func (r *goodsReceiptRepository) getNextCodeLocked(ctx context.Context, tx *gorm.DB, prefix string) (string, error) {
	now := database.GetDB(ctx, r.db).NowFunc()
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
		if n, convErr := parseIntSafe(lastSeqStr); convErr == nil {
			seq = n + 1
		}
	}

	return fmt.Sprintf("%s%04d", codePrefix, seq), nil
}

func parseIntSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func normalizeGRSortField(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "code":
		return "code"
	case "receipt_date":
		return "receipt_date"
	case "status":
		return "status"
	case "created_at":
		return "created_at"
	default:
		return "created_at"
	}
}
