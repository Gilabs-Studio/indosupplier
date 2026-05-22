package repositories

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SupplierRepository defines the interface for supplier data access
type SupplierRepository interface {
	Create(ctx context.Context, supplier *models.Supplier) error
	FindByID(ctx context.Context, id string) (*models.Supplier, error)
	FindByCode(ctx context.Context, code string) (*models.Supplier, error)
	List(ctx context.Context, params SupplierListParams) ([]models.Supplier, int64, error)
	Update(ctx context.Context, supplier *models.Supplier) error
	Delete(ctx context.Context, id string) error
	// Code generation
	GetNextCode(ctx context.Context) (string, error)
	// Nested operations
	CreateContact(ctx context.Context, phone *models.SupplierContact) error
	UpdateContact(ctx context.Context, phone *models.SupplierContact) error
	DeleteContact(ctx context.Context, id string) error
	ClearPrimaryContacts(ctx context.Context, supplierID string) error
	ClearPrimaryContactsExcept(ctx context.Context, contactID string) error
	CreateBankAccount(ctx context.Context, bank *models.SupplierBank) error
	UpdateBankAccount(ctx context.Context, bank *models.SupplierBank) error
	DeleteBankAccount(ctx context.Context, id string) error
	ClearPrimaryBankAccounts(ctx context.Context, supplierID string) error
	ClearPrimaryBankAccountsExcept(ctx context.Context, bankID string) error
}

type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new instance of SupplierRepository
func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *models.Supplier) error {
	return database.GetDB(ctx, r.db).Create(supplier).Error
}

func (r *supplierRepository) FindByID(ctx context.Context, id string) (*models.Supplier, error) {
	var supplier models.Supplier
	err := database.GetDB(ctx, r.db).
		Preload("SupplierType").
		Preload("PaymentTerms").
		Preload("BusinessUnit").
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province").
		Preload("Contacts.ContactRole").
		Preload("BankAccounts.Bank").
		Preload("BankAccounts.Currency").
		First(&supplier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) FindByCode(ctx context.Context, code string) (*models.Supplier, error) {
	var supplier models.Supplier
	err := database.GetDB(ctx, r.db).
		Preload("SupplierType").
		Preload("PaymentTerms").
		Preload("BusinessUnit").
		First(&supplier, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) List(ctx context.Context, params SupplierListParams) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Supplier{})

	// Apply search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR email ILIKE ? OR contact_person ILIKE ?", search, search, search, search)
	}

	// Apply supplier type filter
	if params.SupplierTypeID != "" {
		query = query.Where("supplier_type_id = ?", params.SupplierTypeID)
	}

	// Apply status filter
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Apply approval filter
	if params.IsApproved != nil {
		query = query.Where("is_approved = ?", *params.IsApproved)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"name":       "name",
		"code":       "code",
		"email":      "email",
		"status":     "status",
		"is_active":  "is_active",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(params.SortBy))]
	if sortBy == "" {
		sortBy = "name"
	}

	isDesc := strings.ToLower(strings.TrimSpace(params.SortDir)) == "desc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Preload relations
	query = query.Preload("Province").
		Preload("City").
		Preload("District").
		Preload("SupplierType").
		Preload("PaymentTerms").
		Preload("BusinessUnit").
		Preload("Contacts.ContactRole").
		Preload("BankAccounts.Bank").
		Preload("BankAccounts.Currency")

	if err := query.Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *models.Supplier) error {
	return database.GetDB(ctx, r.db).Save(supplier).Error
}

func (r *supplierRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Supplier{}, "id = ?", id).Error
}

// GetNextCode generates the next supplier code in the format SUP-XXXXX
func (r *supplierRepository) GetNextCode(ctx context.Context) (string, error) {
	var latestCode string
	if err := database.GetDB(ctx, r.db).
		Model(&models.Supplier{}).
		Where("code ~ ?", `^SUP-[0-9]+$`).
		Select("code").
		Order("LENGTH(code) DESC, code DESC").
		Limit(1).
		Scan(&latestCode).Error; err != nil {
		return "", err
	}

	return generateSupplierCode(extractSupplierCodeSequence(latestCode) + 1), nil
}

func generateSupplierCode(seq int) string {
	return "SUP-" + padSupplierNumber(seq, 5)
}

var supplierCodePattern = regexp.MustCompile(`^SUP-(\d+)$`)

func extractSupplierCodeSequence(code string) int {
	matches := supplierCodePattern.FindStringSubmatch(strings.TrimSpace(strings.ToUpper(code)))
	if len(matches) != 2 {
		return 0
	}

	seq, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	return seq
}

func padSupplierNumber(n, width int) string {
	if width <= 0 {
		width = 1
	}
	return fmt.Sprintf("%0*d", width, n)
}

// Nested Phone Number operations
func (r *supplierRepository) CreateContact(ctx context.Context, phone *models.SupplierContact) error {
	return database.GetDB(ctx, r.db).Create(phone).Error
}

func (r *supplierRepository) UpdateContact(ctx context.Context, phone *models.SupplierContact) error {
	return database.GetDB(ctx, r.db).Save(phone).Error
}

func (r *supplierRepository) DeleteContact(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.SupplierContact{}, "id = ?", id).Error
}

func (r *supplierRepository) ClearPrimaryContacts(ctx context.Context, supplierID string) error {
	return database.GetDB(ctx, r.db).Model(&models.SupplierContact{}).Where("supplier_id = ? AND is_primary = true", supplierID).Update("is_primary", false).Error
}

func (r *supplierRepository) ClearPrimaryContactsExcept(ctx context.Context, contactID string) error {
	return database.GetDB(ctx, r.db).Model(&models.SupplierContact{}).Where(
		"supplier_id = (SELECT supplier_id FROM supplier_contacts WHERE id = ?) AND id != ? AND is_primary = true",
		contactID, contactID,
	).Update("is_primary", false).Error
}

// Nested Bank Account operations
func (r *supplierRepository) CreateBankAccount(ctx context.Context, bank *models.SupplierBank) error {
	return database.GetDB(ctx, r.db).Create(bank).Error
}

func (r *supplierRepository) UpdateBankAccount(ctx context.Context, bank *models.SupplierBank) error {
	return database.GetDB(ctx, r.db).Save(bank).Error
}

func (r *supplierRepository) DeleteBankAccount(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.SupplierBank{}, "id = ?", id).Error
}

func (r *supplierRepository) ClearPrimaryBankAccounts(ctx context.Context, supplierID string) error {
	return database.GetDB(ctx, r.db).Model(&models.SupplierBank{}).Where("supplier_id = ? AND is_primary = true", supplierID).Update("is_primary", false).Error
}

func (r *supplierRepository) ClearPrimaryBankAccountsExcept(ctx context.Context, bankID string) error {
	return database.GetDB(ctx, r.db).Model(&models.SupplierBank{}).Where(
		"supplier_id = (SELECT supplier_id FROM supplier_banks WHERE id = ?) AND id != ? AND is_primary = true",
		bankID, bankID,
	).Update("is_primary", false).Error
}
