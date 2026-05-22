package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ContactRepository defines the interface for contact data access
type ContactRepository interface {
	Create(ctx context.Context, contact *models.Contact) error
	FindByID(ctx context.Context, id string) (*models.Contact, error)
	List(ctx context.Context, params ContactListParams) ([]models.Contact, int64, error)
	ListByCustomerID(ctx context.Context, customerID string, params ListParams) ([]models.Contact, int64, error)
	Update(ctx context.Context, contact *models.Contact) error
	Delete(ctx context.Context, id string) error
	CountByCustomerID(ctx context.Context, customerID string) (int64, error)
	CountByCustomerIDs(ctx context.Context, customerIDs []string) (map[string]int64, error)
	ExistsByNameAndCustomer(ctx context.Context, name string, customerID string, excludeID string) (bool, error)
}

// ContactListParams extends ListParams with contact-specific filters
type ContactListParams struct {
	ListParams
	CustomerID    string
	ContactRoleID string
}

type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) Create(ctx context.Context, contact *models.Contact) error {
	return database.GetDB(ctx, r.db).Create(contact).Error
}

func (r *contactRepository) FindByID(ctx context.Context, id string) (*models.Contact, error) {
	var contact models.Contact
	err := database.GetDB(ctx, r.db).
		Preload("Customer").
		Preload("ContactRole").
		First(&contact, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *contactRepository) List(ctx context.Context, params ContactListParams) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Contact{})
	query = security.ApplyScopeFilter(query, ctx, security.DefaultScopeQueryOptions())

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", search, search, search)
	}

	if params.CustomerID != "" {
		query = query.Where("customer_id = ?", params.CustomerID)
	}

	if params.ContactRoleID != "" {
		query = query.Where("contact_role_id = ?", params.ContactRoleID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: params.SortBy}, Desc: params.SortDir == "desc"})
	} else {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: false})
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Preload("Customer").Preload("ContactRole").Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *contactRepository) ListByCustomerID(ctx context.Context, customerID string, params ListParams) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Contact{}).Where("customer_id = ?", customerID)
	query = security.ApplyScopeFilter(query, ctx, security.DefaultScopeQueryOptions())

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelist to prevent SQL injection
	allowedSortColumns := map[string]string{
		"id":              "id",
		"name":            "name",
		"email":           "email",
		"phone":           "phone",
		"position":        "position",
		"is_active":       "is_active",
		"created_at":      "created_at",
		"updated_at":      "updated_at",
		"contact_role_id": "contact_role_id",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "name"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) == "desc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Preload("ContactRole").Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (r *contactRepository) Update(ctx context.Context, contact *models.Contact) error {
	return database.GetDB(ctx, r.db).Save(contact).Error
}

func (r *contactRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Contact{}, "id = ?", id).Error
}

func (r *contactRepository) CountByCustomerID(ctx context.Context, customerID string) (int64, error) {
	var count int64
	err := database.GetDB(ctx, r.db).Model(&models.Contact{}).Where("customer_id = ?", customerID).Count(&count).Error
	return count, err
}

// CountByCustomerIDs returns contact counts for multiple customers in a single GROUP BY query
func (r *contactRepository) CountByCustomerIDs(ctx context.Context, customerIDs []string) (map[string]int64, error) {
	if len(customerIDs) == 0 {
		return make(map[string]int64), nil
	}

	type countResult struct {
		CustomerID string
		Count      int64
	}
	var results []countResult

	err := database.GetDB(ctx, r.db).
		Model(&models.Contact{}).
		Select("customer_id, count(*) as count").
		Where("customer_id IN ?", customerIDs).
		Group("customer_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(results))
	for _, r := range results {
		counts[r.CustomerID] = r.Count
	}
	return counts, nil
}

func (r *contactRepository) ExistsByNameAndCustomer(ctx context.Context, name string, customerID string, excludeID string) (bool, error) {
	var count int64
	query := database.GetDB(ctx, r.db).Model(&models.Contact{}).
		Where("name = ? AND customer_id = ?", name, customerID)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
