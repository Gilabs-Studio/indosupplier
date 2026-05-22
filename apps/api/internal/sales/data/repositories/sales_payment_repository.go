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

type SalesPaymentListParams struct {
	Search    string
	Status    string
	Method    string
	InvoiceID string
	StartDate string
	EndDate   string
	SortBy    string
	SortDir   string
	Limit     int
	Offset    int
}

type SalesPaymentRepository interface {
	List(ctx context.Context, params SalesPaymentListParams) ([]*models.SalesPayment, int64, error)
	GetByID(ctx context.Context, id string) (*models.SalesPayment, error)
	Create(ctx context.Context, p *models.SalesPayment) error
	Delete(ctx context.Context, id string) error
}

type salesPaymentRepository struct {
	db *gorm.DB
}

func NewSalesPaymentRepository(db *gorm.DB) SalesPaymentRepository {
	return &salesPaymentRepository{db: db}
}

func (r *salesPaymentRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

var salesPaymentAllowedSort = map[string]string{
	"created_at":       "sales_payments.created_at",
	"updated_at":       "sales_payments.updated_at",
	"payment_date":     "sales_payments.payment_date",
	"amount":           "sales_payments.amount",
	"method":           "sales_payments.method",
	"status":           "sales_payments.status",
	"reference_number": "sales_payments.reference_number",
}

func (r *salesPaymentRepository) List(ctx context.Context, params SalesPaymentListParams) ([]*models.SalesPayment, int64, error) {
	var items []*models.SalesPayment
	var total int64

	q := r.getDB(ctx).Model(&models.SalesPayment{}).
		Preload("CustomerInvoice").
		Preload("BankAccount")

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.SalesPaymentScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Joins("LEFT JOIN customer_invoices ON customer_invoices.id = sales_payments.customer_invoice_id").
			Joins("LEFT JOIN customers ON customers.id = customer_invoices.customer_id").
			Where(
				"sales_payments.reference_number ILIKE ? OR sales_payments.notes ILIKE ? OR sales_payments.method ILIKE ? OR customer_invoices.code ILIKE ? OR customer_invoices.invoice_number ILIKE ? OR customers.name ILIKE ?",
				like, like, like, like, like, like,
			)
	}
	if st := strings.TrimSpace(params.Status); st != "" {
		q = q.Where("sales_payments.status = ?", st)
	}
	if m := strings.TrimSpace(params.Method); m != "" {
		q = q.Where("sales_payments.method = ?", m)
	}
	if inv := strings.TrimSpace(params.InvoiceID); inv != "" {
		q = q.Where("sales_payments.customer_invoice_id = ?", inv)
	}
	if sd := strings.TrimSpace(params.StartDate); sd != "" {
		q = q.Where("sales_payments.payment_date >= ?", sd)
	}
	if ed := strings.TrimSpace(params.EndDate); ed != "" {
		q = q.Where("sales_payments.payment_date <= ?", ed)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := salesPaymentAllowedSort[params.SortBy]
	if sortCol == "" {
		sortCol = salesPaymentAllowedSort["created_at"]
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

func (r *salesPaymentRepository) GetByID(ctx context.Context, id string) (*models.SalesPayment, error) {
	var p models.SalesPayment
	if err := r.getDB(ctx).
		Preload("CustomerInvoice").
		Preload("BankAccount").
		First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *salesPaymentRepository) Create(ctx context.Context, p *models.SalesPayment) error {
	db := r.getDB(ctx)
	if err := db.Create(p).Error; err != nil {
		if isMissingSalesPaymentCOAColumnErr(err) {
			return db.Omit("SnapshotCOAID", "ResolvedCOAID").Create(p).Error
		}
		return err
	}
	return nil
}

func (r *salesPaymentRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&models.SalesPayment{}, "id = ?", id).Error
}

func isMissingSalesPaymentCOAColumnErr(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return (strings.Contains(msg, "snapshot_coa_id") || strings.Contains(msg, "resolved_coa_id")) &&
		(strings.Contains(msg, "does not exist") || strings.Contains(msg, "undefined column") || strings.Contains(msg, "sqlstate 42703"))
}
