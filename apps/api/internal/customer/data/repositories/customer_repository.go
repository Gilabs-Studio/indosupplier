package repositories

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/customer/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	Create(ctx context.Context, customer *models.Customer) error
	FindByID(ctx context.Context, id string) (*models.Customer, error)
	FindByCode(ctx context.Context, code string) (*models.Customer, error)
	List(ctx context.Context, params CustomerListParams) ([]models.Customer, int64, error)
	Update(ctx context.Context, customer *models.Customer) error
	Delete(ctx context.Context, id string) error
	// Nested bank account operations
	CreateBankAccount(ctx context.Context, bank *models.CustomerBank) error
	UpdateBankAccount(ctx context.Context, bank *models.CustomerBank) error
	DeleteBankAccount(ctx context.Context, id string) error
	// Code generation
	GetNextCode(ctx context.Context) (string, error)
}

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *models.Customer) error {
	return database.GetDB(ctx, r.db).Create(customer).Error
}

func (r *customerRepository) FindByID(ctx context.Context, id string) (*models.Customer, error) {
	var customer models.Customer
	err := database.GetDB(ctx, r.db).
		Preload("CustomerType").
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Village.District.City.Province").
		Preload("BankAccounts.Bank").
		Preload("BankAccounts.Currency").
		Preload("DefaultBusinessType").
		Preload("DefaultArea").
		Preload("DefaultSalesRep").
		Preload("DefaultPaymentTerms").
		First(&customer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) FindByCode(ctx context.Context, code string) (*models.Customer, error) {
	var customer models.Customer
	err := database.GetDB(ctx, r.db).
		Preload("CustomerType").
		First(&customer, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) List(ctx context.Context, params CustomerListParams) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.Customer{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where(
			"name ILIKE ? OR code ILIKE ? OR email ILIKE ? OR contact_person ILIKE ?",
			search, search, search, search,
		)
	}
	if params.CustomerTypeID != "" {
		query = query.Where("customer_type_id = ?", params.CustomerTypeID)
	}

	if params.IsLoyaltyMember != nil {
		if *params.IsLoyaltyMember {
			query = query.Where("EXISTS (SELECT 1 FROM loyalty_members WHERE loyalty_members.customer_id = customers.id)")
		} else {
			query = query.Where("NOT EXISTS (SELECT 1 FROM loyalty_members WHERE loyalty_members.customer_id = customers.id)")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: params.SortBy},
			Desc:   params.SortDir == "desc",
		})
	} else {
		query = query.Order("name ASC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Preload("Province").
		Preload("City").
		Preload("District").
		Preload("CustomerType").
		Preload("BankAccounts.Bank").
		Preload("BankAccounts.Currency")

	if err := query.Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *models.Customer) error {
	return database.GetDB(ctx, r.db).Save(customer).Error
}

func (r *customerRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Customer{}, "id = ?", id).Error
}

func (r *customerRepository) CreateBankAccount(ctx context.Context, bank *models.CustomerBank) error {
	return database.GetDB(ctx, r.db).Create(bank).Error
}

func (r *customerRepository) UpdateBankAccount(ctx context.Context, bank *models.CustomerBank) error {
	return database.GetDB(ctx, r.db).Save(bank).Error
}

func (r *customerRepository) DeleteBankAccount(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.CustomerBank{}, "id = ?", id).Error
}

// GetNextCode generates the next customer code in the format CUST-XXXXX
func (r *customerRepository) GetNextCode(ctx context.Context) (string, error) {
	var latestCode string
	if err := database.GetDB(ctx, r.db).
		Model(&models.Customer{}).
		Where("code ~ ?", `^CUST-[0-9]+$`).
		Select("code").
		Order("LENGTH(code) DESC, code DESC").
		Limit(1).
		Scan(&latestCode).Error; err != nil {
		return "", err
	}

	return generateCustomerCode(extractCustomerCodeSequence(latestCode) + 1), nil
}

func generateCustomerCode(seq int) string {
	return "CUST-" + padNumber(seq, 5)
}

var customerCodePattern = regexp.MustCompile(`^CUST-(\d+)$`)

func extractCustomerCodeSequence(code string) int {
	matches := customerCodePattern.FindStringSubmatch(strings.TrimSpace(strings.ToUpper(code)))
	if len(matches) != 2 {
		return 0
	}

	seq, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	return seq
}

func padNumber(n, width int) string {
	if width <= 0 {
		width = 1
	}
	return fmt.Sprintf("%0*d", width, n)
}
