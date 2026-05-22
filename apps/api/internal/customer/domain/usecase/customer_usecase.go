package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/core/utils"
	crmRepos "github.com/gilabs/gims/api/internal/crm/data/repositories"
	coreDB "github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/customer/data/models"
	"github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/customer/domain/dto"
	"github.com/gilabs/gims/api/internal/customer/domain/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerUsecase defines the interface for customer business logic
type CustomerUsecase interface {
	Create(ctx context.Context, userID string, req dto.CreateCustomerRequest) (dto.CustomerResponse, error)
	GetByID(ctx context.Context, id string) (dto.CustomerResponse, error)
	List(ctx context.Context, params repositories.CustomerListParams) ([]dto.CustomerResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateCustomerRequest) (dto.CustomerResponse, error)
	Delete(ctx context.Context, id string) error
	// Nested operations
	AddBankAccount(ctx context.Context, customerID string, req dto.CreateCustomerBankRequest) (dto.CustomerBankResponse, error)
	UpdateBankAccount(ctx context.Context, id string, req dto.UpdateCustomerBankRequest) (dto.CustomerBankResponse, error)
	DeleteBankAccount(ctx context.Context, id string) error
	// Form data for dropdowns
	GetFormData(ctx context.Context) (*dto.CustomerFormDataResponse, error)
}

type customerUsecase struct {
	repo        repositories.CustomerRepository
	typeRepo    repositories.CustomerTypeRepository
	contactRepo crmRepos.ContactRepository
	db          *gorm.DB
}

// NewCustomerUsecase creates a new CustomerUsecase
func NewCustomerUsecase(repo repositories.CustomerRepository, typeRepo repositories.CustomerTypeRepository, contactRepo crmRepos.ContactRepository, db *gorm.DB) CustomerUsecase {
	return &customerUsecase{repo: repo, typeRepo: typeRepo, contactRepo: contactRepo, db: db}
}

func (u *customerUsecase) Create(ctx context.Context, userID string, req dto.CreateCustomerRequest) (dto.CustomerResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	providedCode := strings.TrimSpace(req.Code)
	var customer *models.Customer

	for attempt := 0; attempt < 5; attempt++ {
		nextCode := providedCode
		if nextCode == "" {
			genCode, err := u.repo.GetNextCode(ctx)
			if err != nil {
				return dto.CustomerResponse{}, err
			}
			nextCode = genCode
		}

		existing, _ := u.repo.FindByCode(ctx, nextCode)
		if existing != nil {
			if providedCode != "" {
				return dto.CustomerResponse{}, errors.New("customer code already exists")
			}
			continue
		}

		customer = &models.Customer{
			ID:            uuid.New().String(),
			Code:          nextCode,
			Name:          req.Name,
			Address:       req.Address,
			Email:         req.Email,
			Website:       req.Website,
			NPWP:          req.NPWP,
			ContactPerson: req.ContactPerson,
			Notes:         req.Notes,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			CreatedBy:     &userID,
			IsActive:      isActive,
			CustomerTypeID: req.CustomerTypeID,
			ProvinceID:     req.ProvinceID,
			CityID:         req.CityID,
			DistrictID:     req.DistrictID,
			VillageID:      req.VillageID,
			VillageName:    req.VillageName,
			DefaultBusinessTypeID: req.DefaultBusinessTypeID,
			DefaultAreaID:         req.DefaultAreaID,
			DefaultSalesRepID:     req.DefaultSalesRepID,
			DefaultPaymentTermsID: req.DefaultPaymentTermsID,
			DefaultTaxRate:        req.DefaultTaxRate,
			CreditLimit:           utils.Float64Value(req.CreditLimit),
			CreditIsActive:        utils.BoolValue(req.CreditIsActive),
		}

		if err := u.repo.Create(ctx, customer); err != nil {
			if isUniqueCustomerCodeError(err) && providedCode == "" {
				customer = nil
				continue
			}
			if isUniqueCustomerCodeError(err) {
				return dto.CustomerResponse{}, errors.New("customer code already exists")
			}
			return dto.CustomerResponse{}, err
		}

		break
	}

	if customer == nil {
		return dto.CustomerResponse{}, fmt.Errorf("failed to generate unique customer code after retries")
	}

	// Create nested bank accounts
	for _, bank := range req.BankAccounts {
		bankModel := &models.CustomerBank{
			ID:            uuid.New().String(),
			CustomerID:    customer.ID,
			BankID:        bank.BankID,
			CurrencyID:    &bank.CurrencyID,
			AccountNumber: bank.AccountNumber,
			AccountName:   bank.AccountName,
			Branch:        bank.Branch,
			IsPrimary:     bank.IsPrimary,
		}
		if err := u.repo.CreateBankAccount(ctx, bankModel); err != nil {
			return dto.CustomerResponse{}, err
		}
	}

	// Reload to get all populated relations
	customer, err := u.repo.FindByID(ctx, customer.ID)
	if err != nil {
		return dto.CustomerResponse{}, err
	}

	return mapper.ToCustomerResponse(customer), nil
}

func (u *customerUsecase) GetByID(ctx context.Context, id string) (dto.CustomerResponse, error) {
	customer, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CustomerResponse{}, err
	}
	resp := mapper.ToCustomerResponse(customer)

	// Enrich with contacts count
	if u.contactRepo != nil {
		count, err := u.contactRepo.CountByCustomerID(ctx, id)
		if err != nil {
			log.Printf("Warning: failed to get contacts count for customer %s: %v", id, err)
		} else {
			resp.ContactsCount = count
		}
	}

	return resp, nil
}

func (u *customerUsecase) List(ctx context.Context, params repositories.CustomerListParams) ([]dto.CustomerResponse, int64, error) {
	customers, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	responses := mapper.ToCustomerResponseList(customers)

	// Enrich with contacts counts via batch query
	if u.contactRepo != nil && len(customers) > 0 {
		ids := make([]string, len(customers))
		for i, c := range customers {
			ids[i] = c.ID
		}
		counts, err := u.contactRepo.CountByCustomerIDs(ctx, ids)
		if err != nil {
			log.Printf("Warning: failed to batch count contacts: %v", err)
		} else {
			for i := range responses {
				responses[i].ContactsCount = counts[responses[i].ID]
			}
		}
	}

	return responses, total, nil
}

func (u *customerUsecase) Update(ctx context.Context, id string, req dto.UpdateCustomerRequest) (dto.CustomerResponse, error) {
	customer, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CustomerResponse{}, err
	}

	// Apply partial updates — only overwrite fields that are explicitly provided
	if req.Code != nil {
		existing, _ := u.repo.FindByCode(ctx, *req.Code)
		if existing != nil && existing.ID != id {
			return dto.CustomerResponse{}, errors.New("customer code already exists")
		}
		customer.Code = *req.Code
	}
	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Address != nil {
		customer.Address = *req.Address
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Website != nil {
		customer.Website = *req.Website
	}
	if req.NPWP != nil {
		customer.NPWP = *req.NPWP
	}
	if req.ContactPerson != nil {
		customer.ContactPerson = *req.ContactPerson
	}
	if req.Notes != nil {
		customer.Notes = *req.Notes
	}
	if req.CustomerTypeID != nil {
		customer.CustomerTypeID = req.CustomerTypeID
		// Clear preloaded association so GORM does not override the FK with the old relation ID.
		customer.CustomerType = nil
	}
	// Geographic fields: nil keeps existing, pointer with empty string clears it
	if req.ProvinceID != nil {
		customer.ProvinceID = req.ProvinceID
		customer.Province = nil
	}
	if req.CityID != nil {
		customer.CityID = req.CityID
		customer.City = nil
	}
	if req.DistrictID != nil {
		customer.DistrictID = req.DistrictID
		customer.District = nil
	}
	if req.VillageID != nil {
		customer.VillageID = req.VillageID
		customer.Village = nil
	}
	if req.VillageName != nil {
		customer.VillageName = req.VillageName
	}
	if req.Latitude != nil {
		customer.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		customer.Longitude = req.Longitude
	}
	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}
	// Sales defaults: nil pointer means "keep existing", non-nil means "set or clear"
	if req.DefaultBusinessTypeID != nil {
		customer.DefaultBusinessTypeID = req.DefaultBusinessTypeID
		customer.DefaultBusinessType = nil
	}
	if req.DefaultAreaID != nil {
		customer.DefaultAreaID = req.DefaultAreaID
		customer.DefaultArea = nil
	}
	if req.DefaultSalesRepID != nil {
		customer.DefaultSalesRepID = req.DefaultSalesRepID
		customer.DefaultSalesRep = nil
	}
	if req.DefaultPaymentTermsID != nil {
		customer.DefaultPaymentTermsID = req.DefaultPaymentTermsID
		customer.DefaultPaymentTerms = nil
	}
	if req.DefaultTaxRate != nil {
		customer.DefaultTaxRate = req.DefaultTaxRate
	}
	if req.CreditLimit != nil {
		customer.CreditLimit = *req.CreditLimit
	}
	if req.CreditIsActive != nil {
		customer.CreditIsActive = *req.CreditIsActive
	}

	if err := u.repo.Update(ctx, customer); err != nil {
		return dto.CustomerResponse{}, err
	}

	// Reload for fresh relations
	customer, _ = u.repo.FindByID(ctx, id)
	return mapper.ToCustomerResponse(customer), nil
}

func (u *customerUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("customer not found")
	}
	return u.repo.Delete(ctx, id)
}

// Nested bank account operations

func (u *customerUsecase) AddBankAccount(ctx context.Context, customerID string, req dto.CreateCustomerBankRequest) (dto.CustomerBankResponse, error) {
	bank := &models.CustomerBank{
		ID:            uuid.New().String(),
		CustomerID:    customerID,
		BankID:        req.BankID,
		CurrencyID:    &req.CurrencyID,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Branch:        req.Branch,
		IsPrimary:     req.IsPrimary,
	}

	if err := u.repo.CreateBankAccount(ctx, bank); err != nil {
		return dto.CustomerBankResponse{}, err
	}

	return dto.CustomerBankResponse{
		ID:            bank.ID,
		CustomerID:    bank.CustomerID,
		BankID:        bank.BankID,
		CurrencyID:    bank.CurrencyID,
		AccountNumber: bank.AccountNumber,
		AccountName:   bank.AccountName,
		Branch:        bank.Branch,
		IsPrimary:     bank.IsPrimary,
		CreatedAt:     bank.CreatedAt,
		UpdatedAt:     bank.UpdatedAt,
	}, nil
}

func (u *customerUsecase) UpdateBankAccount(ctx context.Context, id string, req dto.UpdateCustomerBankRequest) (dto.CustomerBankResponse, error) {
	bank := &models.CustomerBank{ID: id}
	if req.BankID != "" {
		bank.BankID = req.BankID
	}
	if req.CurrencyID != "" {
		bank.CurrencyID = &req.CurrencyID
	}
	if req.AccountNumber != "" {
		bank.AccountNumber = req.AccountNumber
	}
	if req.AccountName != "" {
		bank.AccountName = req.AccountName
	}
	if req.Branch != "" {
		bank.Branch = req.Branch
	}
	if req.IsPrimary != nil {
		bank.IsPrimary = *req.IsPrimary
	}

	if err := u.repo.UpdateBankAccount(ctx, bank); err != nil {
		return dto.CustomerBankResponse{}, err
	}

	return dto.CustomerBankResponse{
		ID:            bank.ID,
		BankID:        bank.BankID,
		CurrencyID:    bank.CurrencyID,
		AccountNumber: bank.AccountNumber,
		AccountName:   bank.AccountName,
		Branch:        bank.Branch,
		IsPrimary:     bank.IsPrimary,
	}, nil
}

func (u *customerUsecase) DeleteBankAccount(ctx context.Context, id string) error {
	return u.repo.DeleteBankAccount(ctx, id)
}

// GetFormData returns dropdown options for customer forms
func (u *customerUsecase) GetFormData(ctx context.Context) (*dto.CustomerFormDataResponse, error) {
	customerTypes, _, err := u.typeRepo.List(ctx, repositories.ListParams{
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}

	// Filter to only active customer types for selection
	var activeCustomerTypes []models.CustomerType
	for _, ct := range customerTypes {
		if ct.IsActive {
			activeCustomerTypes = append(activeCustomerTypes, ct)
		}
	}

	// Load business types
	type btRow struct {
		ID   string
		Name string
	}
	var btRows []btRow
	coreDB.GetDB(ctx, u.db).Table("business_types").
		Select("id, name").Where("deleted_at IS NULL AND is_active = true").
		Order("name").Scan(&btRows)
	btOptions := make([]dto.SalesDefaultOptionBrief, 0, len(btRows))
	for _, r := range btRows {
		btOptions = append(btOptions, dto.SalesDefaultOptionBrief{ID: r.ID, Name: r.Name})
	}

	// Load areas
	type areaRow struct {
		ID       string
		Name     string
		Province string
	}
	var areaRows []areaRow
	coreDB.GetDB(ctx, u.db).Table("areas").
		Select("id, name, province").Where("deleted_at IS NULL AND is_active = true").
		Order("name").Scan(&areaRows)
	areaOptions := make([]dto.CustomerAreaFormOption, 0, len(areaRows))
	for _, r := range areaRows {
		areaOptions = append(areaOptions, dto.CustomerAreaFormOption{ID: r.ID, Name: r.Name, Province: r.Province})
	}

	// Load employees (sales reps)
	type empRow struct {
		ID           string
		EmployeeCode string
		Name         string
	}
	var empRows []empRow
	coreDB.GetDB(ctx, u.db).Table("employees").
		Select("id, employee_code, name").Where("deleted_at IS NULL").
		Order("name").Scan(&empRows)
	salesRepOptions := make([]dto.SalesRepBrief, 0, len(empRows))
	for _, r := range empRows {
		salesRepOptions = append(salesRepOptions, dto.SalesRepBrief{
			ID:           r.ID,
			EmployeeCode: r.EmployeeCode,
			Name:         r.Name,
		})
	}

	// Load payment terms
	type ptRow struct {
		ID   string
		Code string
		Name string
		Days int
	}
	var ptRows []ptRow
	coreDB.GetDB(ctx, u.db).Table("payment_terms").
		Select("id, code, name, days").Where("deleted_at IS NULL AND is_active = true").
		Order("days").Scan(&ptRows)
	ptOptions := make([]dto.PaymentTermsFormOption, 0, len(ptRows))
	for _, r := range ptRows {
		ptOptions = append(ptOptions, dto.PaymentTermsFormOption{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
			Days: r.Days,
		})
	}

	return &dto.CustomerFormDataResponse{
		CustomerTypes: mapper.ToCustomerTypeResponseList(activeCustomerTypes),
		BusinessTypes: btOptions,
		Areas:         areaOptions,
		SalesReps:     salesRepOptions,
		PaymentTerms:  ptOptions,
	}, nil
}

func isUniqueCustomerCodeError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "23505") && !strings.Contains(msg, "duplicate key") && !strings.Contains(msg, "unique") {
		return false
	}

	return strings.Contains(msg, "customers_code") ||
		strings.Contains(msg, "uq_customers_tenant_code_active") ||
		strings.Contains(msg, "customer code")
}
