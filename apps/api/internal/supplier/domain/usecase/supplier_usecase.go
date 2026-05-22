package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
	"github.com/gilabs/gims/api/internal/supplier/domain/mapper"
	"github.com/google/uuid"
)

// SupplierUsecase defines the interface for supplier business logic
type SupplierUsecase interface {
	Create(ctx context.Context, userID string, req dto.CreateSupplierRequest) (dto.SupplierResponse, error)
	GetByID(ctx context.Context, id string) (dto.SupplierResponse, error)
	List(ctx context.Context, params repositories.SupplierListParams) ([]dto.SupplierResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateSupplierRequest) (dto.SupplierResponse, error)
	Delete(ctx context.Context, id string) error
	// Approval workflow
	Submit(ctx context.Context, id string) (dto.SupplierResponse, error)
	Approve(ctx context.Context, id string, userID string, req dto.ApproveSupplierRequest) (dto.SupplierResponse, error)
	// Nested operations
	AddContact(ctx context.Context, supplierID string, req dto.CreateContactRequest) (dto.ContactResponse, error)
	UpdateContact(ctx context.Context, id string, req dto.UpdateContactRequest) (dto.ContactResponse, error)
	DeleteContact(ctx context.Context, id string) error
	AddBankAccount(ctx context.Context, supplierID string, req dto.CreateSupplierBankRequest) (dto.SupplierBankResponse, error)
	UpdateBankAccount(ctx context.Context, id string, req dto.UpdateSupplierBankRequest) (dto.SupplierBankResponse, error)
	DeleteBankAccount(ctx context.Context, id string) error
}

type supplierUsecase struct {
	repo repositories.SupplierRepository
}

// NewSupplierUsecase creates a new SupplierUsecase
func NewSupplierUsecase(repo repositories.SupplierRepository) SupplierUsecase {
	return &supplierUsecase{repo: repo}
}

func (u *supplierUsecase) Create(ctx context.Context, userID string, req dto.CreateSupplierRequest) (dto.SupplierResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var supplier *models.Supplier
	for attempt := 0; attempt < 5; attempt++ {
		code, err := u.repo.GetNextCode(ctx)
		if err != nil {
			return dto.SupplierResponse{}, errors.New("failed to generate supplier code")
		}

		supplier = &models.Supplier{
			ID:            uuid.New().String(),
			Code:          code,
			Name:          req.Name,
			Address:       req.Address,
			Email:         req.Email,
			Website:       req.Website,
			NPWP:          req.NPWP,
			ContactPerson: req.ContactPerson,
			Notes:         req.Notes,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Status:        models.SupplierStatusApproved,
			IsApproved:    true,
			CreatedBy:     &userID,
			IsActive:      isActive,
		}

		if req.SupplierTypeID != "" {
			supplier.SupplierTypeID = &req.SupplierTypeID
		}
		if req.PaymentTermsID != "" {
			supplier.PaymentTermsID = &req.PaymentTermsID
		}
		if req.BusinessUnitID != "" {
			supplier.BusinessUnitID = &req.BusinessUnitID
		}
		if req.ProvinceID != "" {
			supplier.ProvinceID = &req.ProvinceID
		}
		if req.CityID != "" {
			supplier.CityID = &req.CityID
		}
		if req.DistrictID != "" {
			supplier.DistrictID = &req.DistrictID
		}
		if req.VillageID != "" {
			supplier.VillageID = &req.VillageID
		}
		if req.VillageName != "" {
			supplier.VillageName = &req.VillageName
		}

		if err := u.repo.Create(ctx, supplier); err != nil {
			if isUniqueSupplierCodeError(err) {
				supplier = nil
				continue
			}
			return dto.SupplierResponse{}, err
		}

		break
	}

	if supplier == nil {
		return dto.SupplierResponse{}, fmt.Errorf("failed to generate unique supplier code after retries")
	}

	// Create phone numbers
	for _, phone := range req.Contacts {
		contactRoleID := normalizeOptionalUUIDPointer(phone.ContactRoleID)
		phoneModel := &models.SupplierContact{
			ID:            uuid.New().String(),
			SupplierID:    supplier.ID,
			ContactRoleID: contactRoleID,
			Name:          phone.Name,
			Email:         phone.Email,
			Phone:         phone.Phone,
			Notes:         phone.Notes,
			IsPrimary:     phone.IsPrimary,
		}
		if phone.IsActive != nil {
			phoneModel.IsActive = *phone.IsActive
		} else {
			phoneModel.IsActive = true
		}
		if err := u.repo.CreateContact(ctx, phoneModel); err != nil {
			return dto.SupplierResponse{}, err
		}
	}

	// Create bank accounts
	for _, bank := range req.BankAccounts {
		bankModel := &models.SupplierBank{
			ID:            uuid.New().String(),
			SupplierID:    supplier.ID,
			BankID:        bank.BankID,
			CurrencyID:    &bank.CurrencyID,
			AccountNumber: bank.AccountNumber,
			AccountName:   bank.AccountName,
			Branch:        bank.Branch,
			IsPrimary:     bank.IsPrimary,
		}
		if err := u.repo.CreateBankAccount(ctx, bankModel); err != nil {
			return dto.SupplierResponse{}, err
		}
	}

	// Reload to get all relations
	reloaded, err := u.repo.FindByID(ctx, supplier.ID)
	if err != nil {
		return dto.SupplierResponse{}, err
	}

	return mapper.ToSupplierResponse(reloaded), nil
}

func (u *supplierUsecase) GetByID(ctx context.Context, id string) (dto.SupplierResponse, error) {
	supplier, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierResponse{}, err
	}
	return mapper.ToSupplierResponse(supplier), nil
}

func (u *supplierUsecase) List(ctx context.Context, params repositories.SupplierListParams) ([]dto.SupplierResponse, int64, error) {
	suppliers, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToSupplierResponseList(suppliers), total, nil
}

func (u *supplierUsecase) Update(ctx context.Context, id string, req dto.UpdateSupplierRequest) (dto.SupplierResponse, error) {
	supplier, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierResponse{}, err
	}

	// Validate email format when a non-empty email is provided.
	// (Binding tag removed from DTO because omitempty on *string doesn't skip
	// for non-nil pointer to "", which blocks the entire request.)
	if req.Email != nil && *req.Email != "" {
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return dto.SupplierResponse{}, errors.New("invalid email format")
		}
	}

	// Update fields when explicitly sent (pointer non-nil)
	if req.Name != nil {
		supplier.Name = *req.Name
	}
	if req.Address != nil {
		supplier.Address = *req.Address
	}
	if req.Email != nil {
		supplier.Email = *req.Email
	}
	if req.Website != nil {
		supplier.Website = *req.Website
	}
	if req.NPWP != nil {
		supplier.NPWP = *req.NPWP
	}
	if req.ContactPerson != nil {
		supplier.ContactPerson = *req.ContactPerson
	}
	if req.Notes != nil {
		supplier.Notes = *req.Notes
	}

	// Nullable FK fields: empty string means clear, non-empty means set
	if req.SupplierTypeID != nil {
		if *req.SupplierTypeID == "" {
			supplier.SupplierTypeID = nil
		} else {
			supplier.SupplierTypeID = req.SupplierTypeID
		}
	}
	if req.PaymentTermsID != nil {
		if *req.PaymentTermsID == "" {
			supplier.PaymentTermsID = nil
		} else {
			supplier.PaymentTermsID = req.PaymentTermsID
		}
	}
	if req.BusinessUnitID != nil {
		if *req.BusinessUnitID == "" {
			supplier.BusinessUnitID = nil
		} else {
			supplier.BusinessUnitID = req.BusinessUnitID
		}
	}
	if req.ProvinceID != nil {
		if *req.ProvinceID == "" {
			supplier.ProvinceID = nil
		} else {
			supplier.ProvinceID = req.ProvinceID
		}
	}
	if req.CityID != nil {
		if *req.CityID == "" {
			supplier.CityID = nil
		} else {
			supplier.CityID = req.CityID
		}
	}
	if req.DistrictID != nil {
		if *req.DistrictID == "" {
			supplier.DistrictID = nil
		} else {
			supplier.DistrictID = req.DistrictID
		}
	}
	if req.VillageID != nil {
		if *req.VillageID == "" {
			supplier.VillageID = nil
		} else {
			supplier.VillageID = req.VillageID
		}
	}
	if req.VillageName != nil {
		supplier.VillageName = req.VillageName
	}

	if req.Latitude != nil {
		supplier.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		supplier.Longitude = req.Longitude
	}
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}

	// Detach preloaded belongs_to associations so GORM uses FK field values directly.
	// Without this, GORM syncs FK columns back to the loaded association's PK,
	// reverting any FK changes made above.
	supplier.SupplierType = nil
	supplier.PaymentTerms = nil
	supplier.BusinessUnit = nil
	supplier.Province = nil
	supplier.City = nil
	supplier.District = nil
	supplier.Village = nil
	supplier.Contacts = nil
	supplier.BankAccounts = nil

	if err := u.repo.Update(ctx, supplier); err != nil {
		return dto.SupplierResponse{}, err
	}

	// Reload
	supplier, _ = u.repo.FindByID(ctx, id)
	return mapper.ToSupplierResponse(supplier), nil
}

func (u *supplierUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("supplier not found")
	}

	return u.repo.Delete(ctx, id)
}

// Approval workflow
func (u *supplierUsecase) Submit(ctx context.Context, id string) (dto.SupplierResponse, error) {
	supplier, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierResponse{}, err
	}

	if supplier.Status != models.SupplierStatusDraft && supplier.Status != models.SupplierStatusRejected {
		return dto.SupplierResponse{}, errors.New("supplier can only be submitted from draft or rejected status")
	}

	supplier.Status = models.SupplierStatusPending
	if err := u.repo.Update(ctx, supplier); err != nil {
		return dto.SupplierResponse{}, err
	}

	return mapper.ToSupplierResponse(supplier), nil
}

func (u *supplierUsecase) Approve(ctx context.Context, id string, userID string, req dto.ApproveSupplierRequest) (dto.SupplierResponse, error) {
	supplier, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierResponse{}, err
	}

	if supplier.Status != models.SupplierStatusPending {
		return dto.SupplierResponse{}, errors.New("supplier must be in pending status to approve/reject")
	}

	now := apptime.Now()
	supplier.ApprovedBy = &userID
	supplier.ApprovedAt = &now

	if req.Action == "approve" {
		supplier.Status = models.SupplierStatusApproved
		supplier.IsApproved = true
	} else {
		supplier.Status = models.SupplierStatusRejected
		supplier.IsApproved = false
	}

	if err := u.repo.Update(ctx, supplier); err != nil {
		return dto.SupplierResponse{}, err
	}

	return mapper.ToSupplierResponse(supplier), nil
}

// Nested contact operations
func (u *supplierUsecase) AddContact(ctx context.Context, supplierID string, req dto.CreateContactRequest) (dto.ContactResponse, error) {
	contactRoleID := normalizeOptionalUUIDPointer(req.ContactRoleID)
	phone := &models.SupplierContact{
		ID:            uuid.New().String(),
		SupplierID:    supplierID,
		ContactRoleID: contactRoleID,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Notes:         req.Notes,
		IsPrimary:     req.IsPrimary,
	}
	
	if req.IsActive != nil {
		phone.IsActive = *req.IsActive
	} else {
		phone.IsActive = true
	}

	// Enforce single primary contact per supplier (#343)
	if req.IsPrimary {
		if err := u.repo.ClearPrimaryContacts(ctx, supplierID); err != nil {
			return dto.ContactResponse{}, err
		}
	}

	if err := u.repo.CreateContact(ctx, phone); err != nil {
		return dto.ContactResponse{}, err
	}

	return dto.ContactResponse{
		ID:            phone.ID,
		SupplierID:    phone.SupplierID,
		ContactRoleID: phone.ContactRoleID,
		Name:          phone.Name,
		Email:         phone.Email,
		Phone:         phone.Phone,
		Notes:         phone.Notes,
		IsPrimary:     phone.IsPrimary,
		IsActive:      phone.IsActive,
		CreatedAt:     phone.CreatedAt,
		UpdatedAt:     phone.UpdatedAt,
	}, nil
}

func (u *supplierUsecase) UpdateContact(ctx context.Context, id string, req dto.UpdateContactRequest) (dto.ContactResponse, error) {
	// For simplicity, we'll update directly. In production, you'd want a FindPhoneByID method
	phone := &models.SupplierContact{ID: id}
	if req.ContactRoleID != nil {
		phone.ContactRoleID = normalizeOptionalUUIDPointer(req.ContactRoleID)
	}
	if req.Name != "" {
		phone.Name = req.Name
	}
	if req.Email != "" {
		phone.Email = req.Email
	}
	if req.Phone != "" {
		phone.Phone = req.Phone
	}
	if req.Notes != "" {
		phone.Notes = req.Notes
	}
	if req.IsPrimary != nil {
		phone.IsPrimary = *req.IsPrimary
	}
	if req.IsActive != nil {
		phone.IsActive = *req.IsActive
	}

	// Enforce single primary contact per supplier (#343)
	if req.IsPrimary != nil && *req.IsPrimary {
		if err := u.repo.ClearPrimaryContactsExcept(ctx, id); err != nil {
			return dto.ContactResponse{}, err
		}
	}

	if err := u.repo.UpdateContact(ctx, phone); err != nil {
		return dto.ContactResponse{}, err
	}

	return dto.ContactResponse{
		ID:            phone.ID,
		ContactRoleID: phone.ContactRoleID,
		Name:          phone.Name,
		Email:         phone.Email,
		Phone:         phone.Phone,
		Notes:         phone.Notes,
		IsPrimary:     phone.IsPrimary,
		IsActive:      phone.IsActive,
	}, nil
}

func (u *supplierUsecase) DeleteContact(ctx context.Context, id string) error {
	return u.repo.DeleteContact(ctx, id)
}

// Nested bank account operations
func (u *supplierUsecase) AddBankAccount(ctx context.Context, supplierID string, req dto.CreateSupplierBankRequest) (dto.SupplierBankResponse, error) {
	bank := &models.SupplierBank{
		ID:            uuid.New().String(),
		SupplierID:    supplierID,
		BankID:        req.BankID,
		CurrencyID:    &req.CurrencyID,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Branch:        req.Branch,
		IsPrimary:     req.IsPrimary,
	}

	// Enforce single primary bank account per supplier (#343)
	if req.IsPrimary {
		if err := u.repo.ClearPrimaryBankAccounts(ctx, supplierID); err != nil {
			return dto.SupplierBankResponse{}, err
		}
	}

	if err := u.repo.CreateBankAccount(ctx, bank); err != nil {
		return dto.SupplierBankResponse{}, err
	}

	return dto.SupplierBankResponse{
		ID:            bank.ID,
		SupplierID:    bank.SupplierID,
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

func (u *supplierUsecase) UpdateBankAccount(ctx context.Context, id string, req dto.UpdateSupplierBankRequest) (dto.SupplierBankResponse, error) {
	bank := &models.SupplierBank{ID: id}
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

	// Enforce single primary bank account per supplier (#343)
	if req.IsPrimary != nil && *req.IsPrimary {
		if err := u.repo.ClearPrimaryBankAccountsExcept(ctx, id); err != nil {
			return dto.SupplierBankResponse{}, err
		}
	}

	if err := u.repo.UpdateBankAccount(ctx, bank); err != nil {
		return dto.SupplierBankResponse{}, err
	}

	return dto.SupplierBankResponse{
		ID:            bank.ID,
		BankID:        bank.BankID,
		CurrencyID:    bank.CurrencyID,
		AccountNumber: bank.AccountNumber,
		AccountName:   bank.AccountName,
		Branch:        bank.Branch,
		IsPrimary:     bank.IsPrimary,
	}, nil
}

func (u *supplierUsecase) DeleteBankAccount(ctx context.Context, id string) error {
	return u.repo.DeleteBankAccount(ctx, id)
}

func isUniqueSupplierCodeError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "23505") && !strings.Contains(msg, "duplicate key") && !strings.Contains(msg, "unique") {
		return false
	}

	return strings.Contains(msg, "suppliers_code") ||
		strings.Contains(msg, "uq_suppliers_tenant_code_active") ||
		strings.Contains(msg, "supplier code")
}

func normalizeOptionalUUIDPointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
