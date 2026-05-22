package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CustomerInvoiceRepository defines the interface for customer invoice data access
type CustomerInvoiceRepository interface {
	FindByID(ctx context.Context, id string) (*models.CustomerInvoice, error)
	FindByCode(ctx context.Context, code string) (*models.CustomerInvoice, error)
	List(ctx context.Context, req *dto.ListCustomerInvoicesRequest) ([]models.CustomerInvoice, int64, error)
	ListItems(ctx context.Context, invoiceID string, req *dto.ListCustomerInvoiceItemsRequest) ([]models.CustomerInvoiceItem, int64, error)
	Create(ctx context.Context, invoice *models.CustomerInvoice) error
	Update(ctx context.Context, invoice *models.CustomerInvoice) error
	Delete(ctx context.Context, id string) error
	GetNextInvoiceNumber(ctx context.Context, prefix string) (string, error)
	UpdateStatus(ctx context.Context, id string, status models.CustomerInvoiceStatus, paidAmount *float64, paymentAt *time.Time, userID *string) error
}

type customerInvoiceRepository struct {
	db *gorm.DB
}

// NewCustomerInvoiceRepository creates a new CustomerInvoiceRepository
func NewCustomerInvoiceRepository(db *gorm.DB) CustomerInvoiceRepository {
	return &customerInvoiceRepository{db: db}
}

func (r *customerInvoiceRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *customerInvoiceRepository) FindByID(ctx context.Context, id string) (*models.CustomerInvoice, error) {
	var invoice models.CustomerInvoice
	err := r.getDB(ctx).
		Preload("PaymentTerms").
		Preload("SalesOrder").
		Preload("DeliveryOrder").
		Preload("DownPaymentInvoice").
		Preload("Items.Product").
		Where("id = ?", id).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *customerInvoiceRepository) FindByCode(ctx context.Context, code string) (*models.CustomerInvoice, error) {
	var invoice models.CustomerInvoice
	err := r.getDB(ctx).
		Preload("PaymentTerms").
		Preload("SalesOrder").
		Preload("DeliveryOrder").
		Preload("Items.Product").
		Where("code = ?", code).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *customerInvoiceRepository) List(ctx context.Context, req *dto.ListCustomerInvoicesRequest) ([]models.CustomerInvoice, int64, error) {
	var invoices []models.CustomerInvoice
	var total int64
	var query *gorm.DB

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = r.db.WithContext(ctx).Model(&models.CustomerInvoice{}).
			Joins("LEFT JOIN sales_orders ON sales_orders.id = customer_invoices.sales_order_id").
			Joins("LEFT JOIN employees ON employees.id = sales_orders.sales_rep_id")

		// Apply tenant filter manually since we are using joins
		var err error
		query, err = applyTenantFilter(ctx, query, "customer_invoices.tenant_id", "sales_orders.tenant_id")
		if err != nil {
			return nil, 0, err
		}

		query = query.Where("sales_orders.customer_name ILIKE ? OR employees.name ILIKE ? OR customer_invoices.invoice_number ILIKE ? OR customer_invoices.code ILIKE ? OR customer_invoices.notes ILIKE ?", search, search, search, search, search)
	} else {
		query = r.db.WithContext(ctx).Model(&models.CustomerInvoice{})
		var err error
		query, err = applyTenantFilter(ctx, query, "customer_invoices.tenant_id")
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	// Use fully-qualified column name to prevent ambiguity when JOINs are active (search path)
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OwnerUserIDColumn: "customer_invoices.created_by",
	})

	// Apply status filter
	if req.Status != "" {
		query = query.Where("customer_invoices.status = ?", req.Status)
	}

	// Apply type filter
	if req.Type != "" {
		query = query.Where("customer_invoices.type = ?", req.Type)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("customer_invoices.invoice_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("customer_invoices.invoice_date <= ?", req.DateTo)
	}

	// Apply due date range filter
	if req.DueDateFrom != "" {
		query = query.Where("customer_invoices.due_date >= ?", req.DueDateFrom)
	}
	if req.DueDateTo != "" {
		query = query.Where("customer_invoices.due_date <= ?", req.DueDateTo)
	}

	// Apply sales order filter
	if req.SalesOrderID != "" {
		query = query.Where("sales_order_id = ?", req.SalesOrderID)
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
		"invoice_number": "customer_invoices.invoice_number",
		"invoice_date":   "customer_invoices.invoice_date",
		"due_date":       "customer_invoices.due_date",
		"amount":         "customer_invoices.amount",
		"status":         "customer_invoices.status",
		"created_at":     "customer_invoices.created_at",
		"updated_at":     "customer_invoices.updated_at",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(req.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "customer_invoices.invoice_date"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err := query.
		Preload("PaymentTerms").
		Preload("SalesOrder").
		Preload("DownPaymentInvoice").
		Preload("Items.Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "current_hpp", "image_url")
		}).
		Limit(perPage).
		Offset(offset).
		Find(&invoices).Error
	if err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *customerInvoiceRepository) Create(ctx context.Context, invoice *models.CustomerInvoice) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Generate code inside the transaction with advisory lock to prevent duplicates under concurrency.
		if invoice.Code == "" {
			code, err := r.getNextCodeLocked(ctx, tx, "INV")
			if err != nil {
				return err
			}
			invoice.Code = code
		}

		// Retry loop with savepoint to handle rare race conditions on unique constraint.
		items := invoice.Items
		invoice.Items = nil

		for attempt := 0; attempt < 3; attempt++ {
			sp := fmt.Sprintf("sp_inv_create_%d", attempt)
			if err := tx.SavePoint(sp).Error; err != nil {
				return err
			}

			if err := tx.Create(invoice).Error; err != nil {
				if rbErr := tx.RollbackTo(sp).Error; rbErr != nil {
					return rbErr
				}
				if isInvoiceUniqueCodeConflict(err) {
					code, genErr := r.getNextCodeLocked(ctx, tx, "INV")
					if genErr != nil {
						return genErr
					}
					invoice.Code = code
					continue
				}
				return err
			}
			break
		}

		// Create items with the invoice ID
		if len(items) > 0 {
			for i := range items {
				items[i].CustomerInvoiceID = invoice.ID
				if err := tx.Create(&items[i]).Error; err != nil {
					return err
				}
			}
			invoice.Items = items
		}

		return nil
	})
}

func isInvoiceUniqueCodeConflict(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate key value") && strings.Contains(s, "idx_customer_invoices_code")
}

func (r *customerInvoiceRepository) getNextCodeLocked(ctx context.Context, tx *gorm.DB, prefix string) (string, error) {
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")
	codePrefix := prefix + "-" + dateStr

	// Use advisory lock to serialize code generation per day
	lockKey := "customer_invoice_code:" + codePrefix
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
		return "", err
	}

	var count int64
	if err := tx.WithContext(ctx).Model(&models.CustomerInvoice{}).
		Unscoped().
		Where("code LIKE ?", codePrefix+"-%").
		Count(&count).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%04d", codePrefix, count+1), nil
}

func (r *customerInvoiceRepository) Update(ctx context.Context, invoice *models.CustomerInvoice) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update invoice WITHOUT associations
		if err := tx.Omit("Items").Save(invoice).Error; err != nil {
			return err
		}

		// Delete existing items
		if err := tx.Where("customer_invoice_id = ?", invoice.ID).Delete(&models.CustomerInvoiceItem{}).Error; err != nil {
			return err
		}

		// Create new items
		if len(invoice.Items) > 0 {
			for i := range invoice.Items {
				invoice.Items[i].CustomerInvoiceID = invoice.ID
				invoice.Items[i].CreatedAt = apptime.Now()
				invoice.Items[i].UpdatedAt = apptime.Now()
				if err := tx.Create(&invoice.Items[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *customerInvoiceRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete items first
		if err := tx.Where("customer_invoice_id = ?", id).Delete(&models.CustomerInvoiceItem{}).Error; err != nil {
			return err
		}

		// Delete invoice
		return tx.Delete(&models.CustomerInvoice{}, "id = ?", id).Error
	})
}

func (r *customerInvoiceRepository) GetNextInvoiceNumber(ctx context.Context, prefix string) (string, error) {
	var lastInvoice models.CustomerInvoice
	var sequence int

	// Find the last invoice with the same prefix
	err := r.getDB(ctx).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		First(&lastInvoice).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			sequence = 1
		} else {
			return "", err
		}
	} else {
		var count int64
		r.getDB(ctx).Model(&models.CustomerInvoice{}).
			Where("code LIKE ?", prefix+"%").
			Count(&count)
		sequence = int(count) + 1
	}

	// Generate new code: PREFIX-YYYYMMDD-XXXX (e.g., INV-20240115-0001)
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")
	code := prefix + "-" + dateStr + "-" + fmt.Sprintf("%04d", sequence)

	return code, nil
}

func (r *customerInvoiceRepository) UpdateStatus(ctx context.Context, id string, status models.CustomerInvoiceStatus, paidAmount *float64, paymentAt *time.Time, userID *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case models.CustomerInvoiceStatusPaid:
		if paidAmount != nil {
			updates["paid_amount"] = *paidAmount
			updates["remaining_amount"] = 0
		}
		if paymentAt != nil {
			updates["payment_at"] = *paymentAt
		} else {
			updates["payment_at"] = database.GetDB(ctx, r.db).NowFunc()
		}
	case models.CustomerInvoiceStatusPartial:
		if paidAmount != nil {
			updates["paid_amount"] = *paidAmount
		}
	case models.CustomerInvoiceStatusCancelled:
		updates["cancelled_by"] = userID
		updates["cancelled_at"] = database.GetDB(ctx, r.db).NowFunc()
	}

	return r.getDB(ctx).Model(&models.CustomerInvoice{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListItems retrieves invoice items with pagination
func (r *customerInvoiceRepository) ListItems(ctx context.Context, invoiceID string, req *dto.ListCustomerInvoiceItemsRequest) ([]models.CustomerInvoiceItem, int64, error) {
	var items []models.CustomerInvoiceItem
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
	if err := r.getDB(ctx).Model(&models.CustomerInvoiceItem{}).
		Where("customer_invoice_id = ?", invoiceID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated items
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "current_hpp", "image_url")
		}).
		Where("customer_invoice_id = ?", invoiceID).
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
