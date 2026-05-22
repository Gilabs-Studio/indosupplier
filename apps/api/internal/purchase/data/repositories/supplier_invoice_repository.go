package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SupplierInvoiceListParams struct {
	Search  string
	Status  string
	Type    string
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}

type SupplierInvoiceRepository interface {
	List(ctx context.Context, params SupplierInvoiceListParams) ([]*models.SupplierInvoice, int64, error)
	GetByID(ctx context.Context, id string) (*models.SupplierInvoice, error)
	Create(ctx context.Context, si *models.SupplierInvoice) error
	Update(ctx context.Context, si *models.SupplierInvoice) error
	Delete(ctx context.Context, id string) error
	GetLatestDownPaymentByPO(ctx context.Context, purchaseOrderID string) (*models.SupplierInvoice, error)
}

type supplierInvoiceRepository struct {
	db *gorm.DB
}

func NewSupplierInvoiceRepository(db *gorm.DB) SupplierInvoiceRepository {
	return &supplierInvoiceRepository{db: db}
}

func (r *supplierInvoiceRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

var supplierInvoiceAllowedSort = map[string]string{
	"created_at":     "supplier_invoices.created_at",
	"updated_at":     "supplier_invoices.updated_at",
	"invoice_date":   "supplier_invoices.invoice_date",
	"due_date":       "supplier_invoices.due_date",
	"code":           "supplier_invoices.code",
	"invoice_number": "supplier_invoices.invoice_number",
	"amount":         "supplier_invoices.amount",
	"status":         "supplier_invoices.status",
}

func (r *supplierInvoiceRepository) List(ctx context.Context, params SupplierInvoiceListParams) ([]*models.SupplierInvoice, int64, error) {
	var items []*models.SupplierInvoice
	var total int64

	q := r.getDB(ctx).Model(&models.SupplierInvoice{}).
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("GoodsReceipt").
		Preload("PaymentTerms").
		Preload("DownPaymentInvoice").
		Preload("RegularInvoices")

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.PurchaseScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("supplier_invoices.supplier_name_snapshot ILIKE ? OR supplier_invoices.invoice_number ILIKE ? OR supplier_invoices.code ILIKE ?", like, like, like)
	}
	if st := strings.TrimSpace(params.Status); st != "" {
		q = q.Where("supplier_invoices.status = ?", st)
	}
	if tp := strings.TrimSpace(params.Type); tp != "" {
		q = q.Where("supplier_invoices.type = ?", tp)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := supplierInvoiceAllowedSort[params.SortBy]
	if sortCol == "" {
		sortCol = supplierInvoiceAllowedSort["created_at"]
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

	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *supplierInvoiceRepository) GetByID(ctx context.Context, id string) (*models.SupplierInvoice, error) {
	var si models.SupplierInvoice
	err := r.getDB(ctx).
		Preload("PurchaseOrder").
		Preload("GoodsReceipt").
		Preload("PaymentTerms").
		Preload("Items").
		Preload("Items.Product").
		Preload("DownPaymentInvoice").
		Preload("RegularInvoices").
		First(&si, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &si, nil
}

func (r *supplierInvoiceRepository) Create(ctx context.Context, si *models.SupplierInvoice) error {
	return r.getDB(ctx).Create(si).Error
}

func (r *supplierInvoiceRepository) Update(ctx context.Context, si *models.SupplierInvoice) error {
	return r.getDB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(si).Error
}

func (r *supplierInvoiceRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.SupplierInvoice{}, "id = ?", id).Error
}

func (r *supplierInvoiceRepository) GetLatestDownPaymentByPO(ctx context.Context, purchaseOrderID string) (*models.SupplierInvoice, error) {
	var si models.SupplierInvoice
	err := r.getDB(ctx).
		Model(&models.SupplierInvoice{}).
		Where("purchase_order_id = ?", purchaseOrderID).
		Where("type = ?", models.SupplierInvoiceTypeDownPayment).
		Order("created_at DESC").
		First(&si).Error
	if err != nil {
		return nil, err
	}
	return &si, nil
}
