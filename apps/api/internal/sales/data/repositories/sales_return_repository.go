package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesReturnListParams struct {
	Search     string
	Status     string
	Action     string
	InvoiceID  string
	DeliveryID string
	SortBy     string
	SortDir    string
	Limit      int
	Offset     int
}

type SalesReturnRepository interface {
	List(ctx context.Context, params SalesReturnListParams) ([]*models.SalesReturn, int64, error)
	GetByID(ctx context.Context, id string) (*models.SalesReturn, error)
	Create(ctx context.Context, row *models.SalesReturn) error
	Update(ctx context.Context, id string, row *models.SalesReturn) error
	UpdateStatus(ctx context.Context, id string, status models.SalesReturnStatus) error
	Delete(ctx context.Context, id string) error
}

type salesReturnRepository struct {
	db *gorm.DB
}

func NewSalesReturnRepository(db *gorm.DB) SalesReturnRepository {
	return &salesReturnRepository{db: db}
}

var salesReturnAllowedSort = map[string]string{
	"created_at": "sales_returns.created_at",
	"updated_at": "sales_returns.updated_at",
	"code":       "sales_returns.code",
}

const salesReturnIDFilter = "id = ?"

func (r *salesReturnRepository) List(ctx context.Context, params SalesReturnListParams) ([]*models.SalesReturn, int64, error) {
	rows := make([]*models.SalesReturn, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&models.SalesReturn{}).Preload("Items")
	var err error
	q, err = applyTenantFilter(ctx, q, "sales_returns.tenant_id")
	if err != nil {
		return nil, 0, err
	}
	q = security.ApplyScopeFilter(q, ctx, security.SalesScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Joins("LEFT JOIN customers ON customers.id = sales_returns.customer_id")
		q, err = applyTenantFilter(ctx, q, "customers.tenant_id")
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("customers.name ILIKE ? OR sales_returns.code ILIKE ? OR sales_returns.reason ILIKE ?", like, like, like)
	}
	if s := strings.TrimSpace(params.Status); s != "" {
		q = q.Where("sales_returns.status = ?", strings.ToUpper(s))
	}
	if a := strings.TrimSpace(params.Action); a != "" {
		q = q.Where("sales_returns.action = ?", strings.ToUpper(a))
	}
	if inv := strings.TrimSpace(params.InvoiceID); inv != "" {
		q = q.Where("sales_returns.invoice_id = ?", inv)
	}
	if deliveryID := strings.TrimSpace(params.DeliveryID); deliveryID != "" {
		q = q.Where("sales_returns.delivery_id = ?", deliveryID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := salesReturnAllowedSort[params.SortBy]
	if sortCol == "" {
		sortCol = salesReturnAllowedSort["created_at"]
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

func (r *salesReturnRepository) GetByID(ctx context.Context, id string) (*models.SalesReturn, error) {
	var row models.SalesReturn
	if err := database.GetDB(ctx, r.db).Preload("Items").First(&row, salesReturnIDFilter, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *salesReturnRepository) Create(ctx context.Context, row *models.SalesReturn) error {
	return database.GetDB(ctx, r.db).Create(row).Error
}

func (r *salesReturnRepository) Update(ctx context.Context, id string, row *models.SalesReturn) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Update header fields
		if err := tx.Model(&models.SalesReturn{}).Where(salesReturnIDFilter, id).Updates(map[string]interface{}{
			"invoice_id":   row.InvoiceID,
			"warehouse_id": row.WarehouseID,
			"customer_id":  row.CustomerID,
			"reason":       row.Reason,
			"action":       row.Action,
			"notes":        row.Notes,
			"total_amount": row.TotalAmount,
		}).Error; err != nil {
			return err
		}

		// Replace items: delete existing, insert new
		if err := tx.Where("sales_return_id = ?", id).Delete(&models.SalesReturnItem{}).Error; err != nil {
			return err
		}

		for i := range row.Items {
			row.Items[i].SalesReturnID = id
			if err := tx.Create(&row.Items[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *salesReturnRepository) UpdateStatus(ctx context.Context, id string, status models.SalesReturnStatus) error {
	return database.GetDB(ctx, r.db).
		Model(&models.SalesReturn{}).
		Where(salesReturnIDFilter, id).
		Update("status", status).Error
}

func (r *salesReturnRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.SalesReturn{}, salesReturnIDFilter, id).Error
}
