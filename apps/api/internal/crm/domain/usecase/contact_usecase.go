package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/google/uuid"
)

// ContactUsecase defines the interface for contact business logic
type ContactUsecase interface {
	Create(ctx context.Context, req dto.CreateContactRequest) (dto.ContactResponse, error)
	GetByID(ctx context.Context, id string) (dto.ContactResponse, error)
	List(ctx context.Context, params repositories.ContactListParams) ([]dto.ContactResponse, int64, error)
	ListByCustomerID(ctx context.Context, customerID string, params repositories.ListParams) ([]dto.ContactResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateContactRequest) (dto.ContactResponse, error)
	Delete(ctx context.Context, id string) error
	GetFormData(ctx context.Context) (*dto.ContactFormDataResponse, error)
}

type contactUsecase struct {
	contactRepo     repositories.ContactRepository
	contactRoleRepo repositories.ContactRoleRepository
	customerRepo    customerRepos.CustomerRepository
}

// NewContactUsecase creates a new contact usecase
func NewContactUsecase(
	contactRepo repositories.ContactRepository,
	contactRoleRepo repositories.ContactRoleRepository,
	customerRepo customerRepos.CustomerRepository,
) ContactUsecase {
	return &contactUsecase{
		contactRepo:     contactRepo,
		contactRoleRepo: contactRoleRepo,
		customerRepo:    customerRepo,
	}
}

func (u *contactUsecase) Create(ctx context.Context, req dto.CreateContactRequest) (dto.ContactResponse, error) {
	// Validate customer exists
	_, err := u.customerRepo.FindByID(ctx, req.CustomerID)
	if err != nil {
		return dto.ContactResponse{}, errors.New("customer not found")
	}

	// Validate contact role exists if provided
	if req.ContactRoleID != nil && *req.ContactRoleID != "" {
		_, err := u.contactRoleRepo.FindByID(ctx, *req.ContactRoleID)
		if err != nil {
			return dto.ContactResponse{}, errors.New("contact role not found")
		}
	}

	// Check unique name per customer
	exists, err := u.contactRepo.ExistsByNameAndCustomer(ctx, req.Name, req.CustomerID, "")
	if err != nil {
		return dto.ContactResponse{}, err
	}
	if exists {
		return dto.ContactResponse{}, errors.New("contact with this name already exists for this customer")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	contact := &models.Contact{
		ID:            uuid.New().String(),
		CustomerID:    req.CustomerID,
		ContactRoleID: req.ContactRoleID,
		Name:          req.Name,
		Phone:         req.Phone,
		Email:         req.Email,
		Notes:         req.Notes,
		IsActive:      isActive,
	}

	if err := u.contactRepo.Create(ctx, contact); err != nil {
		return dto.ContactResponse{}, err
	}

	// Reload with preloaded relations
	created, err := u.contactRepo.FindByID(ctx, contact.ID)
	if err != nil {
		return dto.ContactResponse{}, err
	}

	return mapper.ToContactResponse(created), nil
}

func (u *contactUsecase) GetByID(ctx context.Context, id string) (dto.ContactResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.Contact{}, id, security.DefaultScopeQueryOptions()) {
		return dto.ContactResponse{}, errors.New("contact not found")
	}
	contact, err := u.contactRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ContactResponse{}, err
	}
	return mapper.ToContactResponse(contact), nil
}

func (u *contactUsecase) List(ctx context.Context, params repositories.ContactListParams) ([]dto.ContactResponse, int64, error) {
	contacts, total, err := u.contactRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToContactResponseList(contacts), total, nil
}

func (u *contactUsecase) ListByCustomerID(ctx context.Context, customerID string, params repositories.ListParams) ([]dto.ContactResponse, int64, error) {
	contacts, total, err := u.contactRepo.ListByCustomerID(ctx, customerID, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToContactResponseList(contacts), total, nil
}

func (u *contactUsecase) Update(ctx context.Context, id string, req dto.UpdateContactRequest) (dto.ContactResponse, error) {
	contact, err := u.contactRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ContactResponse{}, errors.New("contact not found")
	}

	// Validate customer if changing
	if req.CustomerID != "" && req.CustomerID != contact.CustomerID {
		_, err := u.customerRepo.FindByID(ctx, req.CustomerID)
		if err != nil {
			return dto.ContactResponse{}, errors.New("customer not found")
		}
		contact.CustomerID = req.CustomerID
	}

	// Validate contact role if changing
	if req.ContactRoleID != nil {
		if *req.ContactRoleID != "" {
			_, err := u.contactRoleRepo.FindByID(ctx, *req.ContactRoleID)
			if err != nil {
				return dto.ContactResponse{}, errors.New("contact role not found")
			}
		}
		contact.ContactRoleID = req.ContactRoleID
	}

	// Check unique name per customer if name is changing
	if req.Name != "" && req.Name != contact.Name {
		customerID := contact.CustomerID
		if req.CustomerID != "" {
			customerID = req.CustomerID
		}
		exists, err := u.contactRepo.ExistsByNameAndCustomer(ctx, req.Name, customerID, id)
		if err != nil {
			return dto.ContactResponse{}, err
		}
		if exists {
			return dto.ContactResponse{}, errors.New("contact with this name already exists for this customer")
		}
		contact.Name = req.Name
	}

	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.Email != "" {
		contact.Email = req.Email
	}
	if req.Notes != "" {
		contact.Notes = req.Notes
	}
	if req.IsActive != nil {
		contact.IsActive = *req.IsActive
	}

	if err := u.contactRepo.Update(ctx, contact); err != nil {
		return dto.ContactResponse{}, err
	}

	// Reload with preloaded relations
	updated, err := u.contactRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ContactResponse{}, err
	}

	return mapper.ToContactResponse(updated), nil
}

func (u *contactUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.contactRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("contact not found")
	}
	return u.contactRepo.Delete(ctx, id)
}

func (u *contactUsecase) GetFormData(ctx context.Context) (*dto.ContactFormDataResponse, error) {
	// Fetch customers
	customers, _, err := u.customerRepo.List(ctx, customerRepos.CustomerListParams{
		ListParams: customerRepos.ListParams{
			Limit: 500,
		},
	})
	if err != nil {
		return nil, err
	}

	customerOptions := make([]dto.ContactCustomerOption, 0, len(customers))
	for _, c := range customers {
		customerOptions = append(customerOptions, dto.ContactCustomerOption{
			ID:   c.ID,
			Code: c.Code,
			Name: c.Name,
		})
	}

	// Fetch contact roles
	roles, _, err := u.contactRoleRepo.List(ctx, repositories.ListParams{
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}

	roleOptions := make([]dto.ContactRoleOptionForForm, 0, len(roles))
	for _, r := range roles {
		roleOptions = append(roleOptions, dto.ContactRoleOptionForForm{
			ID:         r.ID,
			Name:       r.Name,
			Code:       r.Code,
			BadgeColor: r.BadgeColor,
		})
	}

	return &dto.ContactFormDataResponse{
		Customers:    customerOptions,
		ContactRoles: roleOptions,
	}, nil
}
