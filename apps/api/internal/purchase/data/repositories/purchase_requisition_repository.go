package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseRequisitionRepository interface {
	List(ctx context.Context, params PurchaseRequisitionListParams) ([]*models.PurchaseRequisition, int64, error)
	GetByID(ctx context.Context, id string) (*models.PurchaseRequisition, error)
	Create(ctx context.Context, pr *models.PurchaseRequisition) (*models.PurchaseRequisition, error)
	Update(ctx context.Context, pr *models.PurchaseRequisition) (*models.PurchaseRequisition, error)
	// UpdateStatus transitions the PR to the given status and optionally applies extra column updates (e.g. converted_to_purchase_order_id).
	UpdateStatus(ctx context.Context, id string, status models.PurchaseRequisitionStatus, extra ...map[string]interface{}) (*models.PurchaseRequisition, error)
	Delete(ctx context.Context, id string) error
	GetNextCode(ctx context.Context, prefix string) (string, error)
}

type PurchaseRequisitionListParams struct {
	Search  string
	Status  string
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}

type purchaseRequisitionRepository struct {
	db *gorm.DB
}

func NewPurchaseRequisitionRepository(db *gorm.DB) PurchaseRequisitionRepository {
	return &purchaseRequisitionRepository{db: db}
}

func (r *purchaseRequisitionRepository) List(ctx context.Context, params PurchaseRequisitionListParams) ([]*models.PurchaseRequisition, int64, error) {
	var results []*models.PurchaseRequisition
	var total int64

	query := r.db.WithContext(ctx).Model(&models.PurchaseRequisition{})
	var err error
	query, err = applyTenantFilter(ctx, query, "purchase_requisitions.tenant_id")
	if err != nil {
		return nil, 0, err
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.PurchaseScopeQueryOptions())

	if params.Status != "" {
		query = query.Where("purchase_requisitions.status = ?", params.Status)
	}

	if s := strings.TrimSpace(params.Search); s != "" {
		pattern := "%" + s + "%"
		query = query.Joins("LEFT JOIN employees ON employees.id = purchase_requisitions.employee_id")
		query = query.Joins("LEFT JOIN suppliers ON suppliers.id = purchase_requisitions.supplier_id")
		query = query.Where(
			"purchase_requisitions.supplier_name_snapshot ILIKE ? OR suppliers.name ILIKE ? OR suppliers.code ILIKE ? OR employees.name ILIKE ? OR purchase_requisitions.code ILIKE ? OR purchase_requisitions.notes ILIKE ? OR purchase_requisitions.request_date::text ILIKE ?",
			pattern,
			pattern,
			pattern,
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"code":         "purchase_requisitions.code",
		"request_date": "purchase_requisitions.request_date",
		"status":       "purchase_requisitions.status",
		"total_amount": "purchase_requisitions.total_amount",
		"created_at":   "purchase_requisitions.created_at",
		"updated_at":   "purchase_requisitions.updated_at",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "purchase_requisitions.created_at"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) != "asc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	if params.Limit > 0 {
		query = query.Limit(params.Limit).Offset(params.Offset)
	}

	// N+1 safe: preload relations in bulk
	query = query.
		Preload("Supplier").
		Preload("PaymentTerms").
		Preload("BusinessUnit").
		Preload("Employee.User")

	if err := query.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *purchaseRequisitionRepository) GetByID(ctx context.Context, id string) (*models.PurchaseRequisition, error) {
	var pr models.PurchaseRequisition
	err := database.GetDB(ctx, r.db).
		Model(&models.PurchaseRequisition{}).
		Preload("Supplier").
		Preload("PaymentTerms").
		Preload("BusinessUnit").
		Preload("Employee.User").
		Preload("Items.Product").
		First(&pr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *purchaseRequisitionRepository) Create(ctx context.Context, pr *models.PurchaseRequisition) (*models.PurchaseRequisition, error) {
	if pr == nil {
		return nil, fmt.Errorf("purchase requisition is nil")
	}

	return pr, database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Generate Code inside the same transaction to prevent duplicates under concurrency.
		if pr.Code == "" {
			code, err := r.getNextCodeLocked(ctx, tx, "PR")
			if err != nil {
				return err
			}
			pr.Code = code
		}

		// In case of an unlikely race/manual insert, retry a few times.
		for attempt := 0; attempt < 3; attempt++ {
			sp := fmt.Sprintf("sp_pr_create_%d", attempt)
			if err := tx.SavePoint(sp).Error; err != nil {
				return err
			}

			if err := tx.Create(pr).Error; err != nil {
				if rbErr := tx.RollbackTo(sp).Error; rbErr != nil {
					return rbErr
				}

				// Unique constraint violations abort the transaction in Postgres. Savepoints let us recover and retry.
				if isUniqueConstraintViolation(err, "idx_purchase_requisitions_code") {
					code, genErr := r.getNextCodeLocked(ctx, tx, "PR")
					if genErr != nil {
						return genErr
					}
					pr.Code = code
					continue
				}
				return err
			}
			return nil
		}

		return fmt.Errorf("failed to create purchase requisition: code conflict")
	})
}

func (r *purchaseRequisitionRepository) getNextCodeLocked(ctx context.Context, tx *gorm.DB, prefix string) (string, error) {
	now := database.GetDB(ctx, r.db).NowFunc()
	yearStr := now.Format("2006")
	codePrefix := prefix + yearStr

	// Serialize code generation per-year to avoid duplicates.
	lockKey := "purchase_requisition_code:" + codePrefix
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
		return "", err
	}

	var last models.PurchaseRequisition
	err := tx.WithContext(ctx).
		Unscoped().
		Model(&models.PurchaseRequisition{}).
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
		if n, convErr := strconv.Atoi(lastSeqStr); convErr == nil {
			seq = n + 1
		}
	}

	return codePrefix + formatSequence(seq), nil
}

func isUniqueConstraintViolation(err error, constraintName string) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, "SQLSTATE 23505") && strings.Contains(errStr, constraintName) {
		return true
	}
	// fallback heuristic
	return strings.Contains(strings.ToLower(errStr), "duplicate key value") && strings.Contains(errStr, constraintName)
}

func (r *purchaseRequisitionRepository) Update(ctx context.Context, pr *models.PurchaseRequisition) (*models.PurchaseRequisition, error) {
	if pr == nil {
		return nil, fmt.Errorf("purchase requisition is nil")
	}

	err := database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var existing models.PurchaseRequisition
		if err := database.GetDB(ctx, tx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&existing, "id = ?", pr.ID).Error; err != nil {
			return err
		}

		// Replace scalar fields (keep Code/Status as provided by caller)
		if err := database.GetDB(ctx, tx).Model(&existing).Updates(map[string]interface{}{
			"supplier_id":                 pr.SupplierID,
			"supplier_code_snapshot":      pr.SupplierCodeSnapshot,
			"supplier_name_snapshot":      pr.SupplierNameSnapshot,
			"payment_terms_id":            pr.PaymentTermsID,
			"payment_terms_name_snapshot": pr.PaymentTermsNameSnapshot,
			"payment_terms_days_snapshot": pr.PaymentTermsDaysSnapshot,
			"business_unit_id":            pr.BusinessUnitID,
			"business_unit_name_snapshot": pr.BusinessUnitNameSnapshot,
			"employee_id":                 pr.EmployeeID,
			"request_date":                pr.RequestDate,
			"address":                     pr.Address,
			"notes":                       pr.Notes,
			"tax_rate":                    pr.TaxRate,
			"tax_amount":                  pr.TaxAmount,
			"delivery_cost":               pr.DeliveryCost,
			"other_cost":                  pr.OtherCost,
			"subtotal":                    pr.Subtotal,
			"total_amount":                pr.TotalAmount,
		}).Error; err != nil {
			return err
		}

		// Replace items
		if err := database.GetDB(ctx, tx).Where("purchase_requisition_id = ?", pr.ID).Delete(&models.PurchaseRequisitionItem{}).Error; err != nil {
			return err
		}
		if len(pr.Items) > 0 {
			for i := range pr.Items {
				pr.Items[i].PurchaseRequisitionID = pr.ID
			}
			if err := database.GetDB(ctx, tx).Create(&pr.Items).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, pr.ID)
}

func (r *purchaseRequisitionRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := database.GetDB(ctx, tx).Where("purchase_requisition_id = ?", id).Delete(&models.PurchaseRequisitionItem{}).Error; err != nil {
			return err
		}
		return database.GetDB(ctx, tx).Delete(&models.PurchaseRequisition{}, "id = ?", id).Error
	})
}

func (r *purchaseRequisitionRepository) UpdateStatus(ctx context.Context, id string, status models.PurchaseRequisitionStatus, extra ...map[string]interface{}) (*models.PurchaseRequisition, error) {
	err := database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var existing models.PurchaseRequisition
		if err := database.GetDB(ctx, tx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&existing, "id = ?", id).Error; err != nil {
			return err
		}

		now := apptime.Now()
		updates := map[string]interface{}{"status": status}

		// Set workflow timestamp for the new status.
		switch status {
		case models.PurchaseRequisitionStatusSubmitted:
			updates["submitted_at"] = now
		case models.PurchaseRequisitionStatusApproved:
			updates["approved_at"] = now
		case models.PurchaseRequisitionStatusRejected:
			updates["rejected_at"] = now
		case models.PurchaseRequisitionStatusConverted:
			updates["converted_at"] = now
		}

		// Merge any caller-provided extra updates (e.g. converted_to_purchase_order_id).
		for _, m := range extra {
			for k, v := range m {
				updates[k] = v
			}
		}

		return database.GetDB(ctx, tx).Model(&existing).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *purchaseRequisitionRepository) GetNextCode(ctx context.Context, prefix string) (string, error) {
	var out string
	err := database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		code, err := r.getNextCodeLocked(ctx, tx, prefix)
		if err != nil {
			return err
		}
		out = code
		return nil
	})
	return out, err
}

func formatSequence(seq int) string {
	return fmt.Sprintf("%04d", seq)
}

func normalizePRSortField(raw string) string {
	switch raw {
	case "code":
		return "code"
	case "request_date":
		return "request_date"
	case "status":
		return "status"
	case "total_amount":
		return "total_amount"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}
