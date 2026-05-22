package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/google/uuid"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseReturnListParams struct {
	Search         string
	Status         string
	Action         string
	GoodsReceiptID string
	SortBy         string
	SortDir        string
	Limit          int
	Offset         int
}

type PurchaseReturnRepository interface {
	List(ctx context.Context, params PurchaseReturnListParams) ([]*models.PurchaseReturn, int64, error)
	GetByID(ctx context.Context, id string) (*models.PurchaseReturn, error)
	Create(ctx context.Context, row *models.PurchaseReturn) error
	Update(ctx context.Context, row *models.PurchaseReturn) error
	UpdateStatus(ctx context.Context, id string, status models.PurchaseReturnStatus) error
	Delete(ctx context.Context, id string) error
}

type purchaseReturnRepository struct {
	db *gorm.DB
}

func NewPurchaseReturnRepository(db *gorm.DB) PurchaseReturnRepository {
	return &purchaseReturnRepository{db: db}
}

var purchaseReturnAllowedSort = map[string]string{
	"created_at": "purchase_returns.created_at",
	"updated_at": "purchase_returns.updated_at",
	"code":       "purchase_returns.code",
}

const purchaseReturnIDFilter = "id = ?"

func (r *purchaseReturnRepository) List(ctx context.Context, params PurchaseReturnListParams) ([]*models.PurchaseReturn, int64, error) {
	rows := make([]*models.PurchaseReturn, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&models.PurchaseReturn{}).Preload("Items")
	var err error
	q, err = applyTenantFilter(ctx, q, "purchase_returns.tenant_id")
	if err != nil {
		return nil, 0, err
	}
	q = security.ApplyScopeFilter(q, ctx, security.PurchaseScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Joins("LEFT JOIN suppliers ON suppliers.id = purchase_returns.supplier_id")
		q, err = applyTenantFilter(ctx, q, "suppliers.tenant_id")
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("suppliers.name ILIKE ? OR purchase_returns.code ILIKE ? OR purchase_returns.reason ILIKE ?", like, like, like)
	}
	if s := strings.TrimSpace(params.Status); s != "" {
		q = q.Where("purchase_returns.status = ?", strings.ToUpper(s))
	}
	if a := strings.TrimSpace(params.Action); a != "" {
		q = q.Where("purchase_returns.action = ?", strings.ToUpper(a))
	}
	if gr := strings.TrimSpace(params.GoodsReceiptID); gr != "" {
		// Accept UUID or code/reference number for goods_receipt_id filter.
		if _, err := uuid.Parse(gr); err == nil {
			q = q.Where("purchase_returns.goods_receipt_id = ?", gr)
		} else {
			q = q.Where("purchase_returns.goods_receipt_id = (SELECT id FROM goods_receipts WHERE code = ? LIMIT 1)", gr)
		}
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := purchaseReturnAllowedSort[params.SortBy]
	if sortCol == "" {
		sortCol = purchaseReturnAllowedSort["created_at"]
	}
	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) != "asc"

	// Split table and column for clause.Column
	parts := strings.Split(sortCol, ".")
	column := clause.Column{Name: parts[len(parts)-1]}
	if len(parts) > 1 {
		column.Table = parts[0]
	}

	q = q.Order(clause.OrderByColumn{Column: column, Desc: isDesc})

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *purchaseReturnRepository) GetByID(ctx context.Context, id string) (*models.PurchaseReturn, error) {
	var row models.PurchaseReturn
	if err := database.GetDB(ctx, r.db).Preload("Items").First(&row, purchaseReturnIDFilter, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *purchaseReturnRepository) Create(ctx context.Context, row *models.PurchaseReturn) error {
	return database.GetDB(ctx, r.db).Session(&gorm.Session{NewDB: true}).Create(row).Error
}

func (r *purchaseReturnRepository) Update(ctx context.Context, row *models.PurchaseReturn) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PurchaseReturn{}).Where(purchaseReturnIDFilter, row.ID).Updates(map[string]interface{}{
			"warehouse_id": row.WarehouseID,
			"reason":       row.Reason,
			"action":       row.Action,
			"notes":        row.Notes,
			"total_amount": row.TotalAmount,
			"updated_at":   row.UpdatedAt,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("purchase_return_id = ?", row.ID).Delete(&models.PurchaseReturnItem{}).Error; err != nil {
			return err
		}
		if len(row.Items) > 0 {
			if err := tx.Create(row.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *purchaseReturnRepository) UpdateStatus(ctx context.Context, id string, status models.PurchaseReturnStatus) error {
	return database.GetDB(ctx, r.db).
		Model(&models.PurchaseReturn{}).
		Where(purchaseReturnIDFilter, id).
		Update("status", status).Error
}

func (r *purchaseReturnRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.PurchaseReturn{}, purchaseReturnIDFilter, id).Error
}
