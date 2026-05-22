package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesQuotationRepository defines the interface for sales quotation data access
type SalesQuotationRepository interface {
	FindByID(ctx context.Context, id string) (*models.SalesQuotation, error)
	FindByCode(ctx context.Context, code string) (*models.SalesQuotation, error)
	List(ctx context.Context, req *dto.ListSalesQuotationsRequest) ([]models.SalesQuotation, int64, error)
	ListItems(ctx context.Context, quotationID string, req *dto.ListSalesQuotationItemsRequest) ([]models.SalesQuotationItem, int64, error)
	Create(ctx context.Context, sq *models.SalesQuotation) error
	Update(ctx context.Context, sq *models.SalesQuotation) error
	Delete(ctx context.Context, id string) error
	GetNextQuotationNumber(ctx context.Context, prefix string) (string, error)
	UpdateStatus(ctx context.Context, id string, status models.SalesQuotationStatus, userID *string, reason *string) error
}

type salesQuotationRepository struct {
	db *gorm.DB
}

// NewSalesQuotationRepository creates a new SalesQuotationRepository
func NewSalesQuotationRepository(db *gorm.DB) SalesQuotationRepository {
	return &salesQuotationRepository{db: db}
}

func (r *salesQuotationRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *salesQuotationRepository) FindByID(ctx context.Context, id string) (*models.SalesQuotation, error) {
	var quotation models.SalesQuotation
	err := r.getDB(ctx).
		Preload("Customer").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Preload("Items.Product").
		Where("id = ?", id).
		First(&quotation).Error
	if err != nil {
		return nil, err
	}
	return &quotation, nil
}

func (r *salesQuotationRepository) FindByCode(ctx context.Context, code string) (*models.SalesQuotation, error) {
	var quotation models.SalesQuotation
	err := r.getDB(ctx).
		Preload("Customer").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Preload("Items.Product").
		Where("code = ?", code).
		First(&quotation).Error
	if err != nil {
		return nil, err
	}
	return &quotation, nil
}

func (r *salesQuotationRepository) List(ctx context.Context, req *dto.ListSalesQuotationsRequest) ([]models.SalesQuotation, int64, error) {
	var quotations []models.SalesQuotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SalesQuotation{})
	var err error
	query, err = applyTenantFilter(ctx, query, "sales_quotations.tenant_id")
	if err != nil {
		return nil, 0, err
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.SalesScopeQueryOptions())

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = query.Joins("LEFT JOIN employees ON employees.id = sales_quotations.sales_rep_id")
		query, err = applyTenantFilter(ctx, query, "employees.tenant_id")
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("sales_quotations.customer_name ILIKE ? OR employees.name ILIKE ? OR sales_quotations.code ILIKE ? OR sales_quotations.notes ILIKE ?", search, search, search, search)
	}

	// Apply status filter
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("quotation_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("quotation_date <= ?", req.DateTo)
	}

	// Apply sales rep filter
	if req.SalesRepID != "" {
		query = query.Where("sales_rep_id = ?", req.SalesRepID)
	}

	// Apply business unit filter
	if req.BusinessUnitID != "" {
		query = query.Where("business_unit_id = ?", req.BusinessUnitID)
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
		"code":           "code",
		"quotation_date": "quotation_date",
		"valid_until":    "valid_until",
		"customer_name":  "customer_name",
		"total_amount":   "total_amount",
		"status":         "status",
		"created_at":     "created_at",
		"updated_at":     "updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(req.SortBy))]
	if sortBy == "" {
		sortBy = "quotation_date"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err = query.
		Preload("Customer").
		Preload("PaymentTerms").
		Preload("SalesRep").
		Preload("BusinessUnit").
		Preload("BusinessType").
		Limit(perPage).
		Offset(offset).
		Find(&quotations).Error
	if err != nil {
		return nil, 0, err
	}

	return quotations, total, nil
}

func (r *salesQuotationRepository) Create(ctx context.Context, sq *models.SalesQuotation) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Create quotation with items (GORM handles cascade)
		// Store items temporarily
		items := sq.Items
		sq.Items = nil

		// Create quotation without items
		if err := tx.Create(sq).Error; err != nil {
			return err
		}

		// Create items with the quotation ID
		if len(items) > 0 {
			for i := range items {
				items[i].SalesQuotationID = sq.ID
				if err := tx.Create(&items[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesQuotationRepository) Update(ctx context.Context, sq *models.SalesQuotation) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update scalar columns explicitly to avoid stale preloaded associations
		// overriding foreign keys (e.g. customer_id, sales_rep_id, business_type_id).
		quotationUpdates := map[string]interface{}{
			"quotation_date":      sq.QuotationDate,
			"valid_until":         sq.ValidUntil,
			"payment_terms_id":    sq.PaymentTermsID,
			"sales_rep_id":        sq.SalesRepID,
			"business_unit_id":    sq.BusinessUnitID,
			"business_type_id":    sq.BusinessTypeID,
			"customer_id":         sq.CustomerID,
			"customer_contact_id": sq.CustomerContactID,
			"customer_name":       sq.CustomerName,
			"customer_contact":    sq.CustomerContact,
			"customer_phone":      sq.CustomerPhone,
			"customer_email":      sq.CustomerEmail,
			"subtotal":            sq.Subtotal,
			"discount_amount":     sq.DiscountAmount,
			"tax_rate":            sq.TaxRate,
			"tax_amount":          sq.TaxAmount,
			"delivery_cost":       sq.DeliveryCost,
			"other_cost":          sq.OtherCost,
			"total_amount":        sq.TotalAmount,
			"notes":               sq.Notes,
			"updated_at":          sq.UpdatedAt,
		}

		if err := tx.Model(&models.SalesQuotation{}).
			Where("id = ?", sq.ID).
			Updates(quotationUpdates).Error; err != nil {
			return err
		}

		// Delete existing items (soft delete)
		if err := tx.Where("sales_quotation_id = ?", sq.ID).Delete(&models.SalesQuotationItem{}).Error; err != nil {
			return err
		}

		// Create new items
		if len(sq.Items) > 0 {
			for i := range sq.Items {
				sq.Items[i].SalesQuotationID = sq.ID
				sq.Items[i].CreatedAt = apptime.Now()
				sq.Items[i].UpdatedAt = apptime.Now()
				if err := tx.Create(&sq.Items[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesQuotationRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete items first (CASCADE should handle this, but explicit for safety)
		if err := tx.Where("sales_quotation_id = ?", id).Delete(&models.SalesQuotationItem{}).Error; err != nil {
			return err
		}

		// Delete quotation
		return tx.Delete(&models.SalesQuotation{}, "id = ?", id).Error
	})
}

func (r *salesQuotationRepository) GetNextQuotationNumber(ctx context.Context, prefix string) (string, error) {
	var lastQuotation models.SalesQuotation
	var sequence int

	// Find the last quotation with the same prefix
	err := r.getDB(ctx).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		First(&lastQuotation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No previous quotation, start from 1
			sequence = 1
		} else {
			return "", err
		}
	} else {
		// Extract sequence number from last code
		// Format: PREFIX-YYYYMMDD-0001
		// We'll use a simpler format: PREFIX-YYYYMMDD-XXXX
		// For now, just increment based on count
		var count int64
		r.getDB(ctx).Model(&models.SalesQuotation{}).
			Where("code LIKE ?", prefix+"%").
			Count(&count)
		sequence = int(count) + 1
	}

	// Generate new code: PREFIX-YYYYMMDD-XXXX
	// Format: SQ-20240115-0001
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")

	// Format sequence with 4 digits
	code := prefix + "-" + dateStr + "-" + formatSequence(sequence)

	return code, nil
}

// formatSequence formats sequence number with 4 digits
// formatSequence formats sequence number with 4 digits (0001, 0002, etc.)
func formatSequence(seq int) string {
	return fmt.Sprintf("%04d", seq)
}

func (r *salesQuotationRepository) UpdateStatus(ctx context.Context, id string, status models.SalesQuotationStatus, userID *string, reason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case models.SalesQuotationStatusApproved:
		updates["approved_by"] = userID
		updates["approved_at"] = database.GetDB(ctx, r.db).NowFunc()
	case models.SalesQuotationStatusRejected:
		updates["rejected_by"] = userID
		updates["rejected_at"] = database.GetDB(ctx, r.db).NowFunc()
		if reason != nil {
			updates["rejection_reason"] = *reason
		}
	case models.SalesQuotationStatusConverted:
		updates["converted_at"] = database.GetDB(ctx, r.db).NowFunc()
	}

	return r.getDB(ctx).Model(&models.SalesQuotation{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListItems retrieves quotation items with pagination
func (r *salesQuotationRepository) ListItems(ctx context.Context, quotationID string, req *dto.ListSalesQuotationItemsRequest) ([]models.SalesQuotationItem, int64, error) {
	var items []models.SalesQuotationItem
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
	if err := r.getDB(ctx).Model(&models.SalesQuotationItem{}).
		Where("sales_quotation_id = ?", quotationID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated items with minimal preload (only product info)
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "image_url")
		}).
		Where("sales_quotation_id = ?", quotationID).
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&items).Error

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
