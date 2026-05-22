package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	"github.com/google/uuid"
)

// ContactRoleUsecase defines the interface for contact role business logic
type ContactRoleUsecase interface {
	Create(ctx context.Context, req dto.CreateContactRoleRequest) (dto.ContactRoleResponse, error)
	GetByID(ctx context.Context, id string) (dto.ContactRoleResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ContactRoleResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateContactRoleRequest) (dto.ContactRoleResponse, error)
	Delete(ctx context.Context, id string) error
}

type contactRoleUsecase struct {
	repo repositories.ContactRoleRepository
}

// NewContactRoleUsecase creates a new contact role usecase
func NewContactRoleUsecase(repo repositories.ContactRoleRepository) ContactRoleUsecase {
	return &contactRoleUsecase{repo: repo}
}

func (u *contactRoleUsecase) Create(ctx context.Context, req dto.CreateContactRoleRequest) (dto.ContactRoleResponse, error) {
	exists, err := u.repo.ExistsByName(ctx, req.Name, "")
	if err != nil {
		return dto.ContactRoleResponse{}, err
	}
	if exists {
		return dto.ContactRoleResponse{}, errors.New("contact role with this name already exists")
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	roleID := uuid.New().String()

	role := &models.ContactRole{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        generateContactRoleCode(req.Name, roleID),
		Description: req.Description,
		BadgeColor:  req.BadgeColor,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, role); err != nil {
		return dto.ContactRoleResponse{}, err
	}

	return mapper.ToContactRoleResponse(role), nil
}

func (u *contactRoleUsecase) GetByID(ctx context.Context, id string) (dto.ContactRoleResponse, error) {
	role, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ContactRoleResponse{}, err
	}
	return mapper.ToContactRoleResponse(role), nil
}

func (u *contactRoleUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ContactRoleResponse, int64, error) {
	roles, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToContactRoleResponseList(roles), total, nil
}

func (u *contactRoleUsecase) Update(ctx context.Context, id string, req dto.UpdateContactRoleRequest) (dto.ContactRoleResponse, error) {
	role, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ContactRoleResponse{}, errors.New("contact role not found")
	}

	if req.Name != "" {
		exists, existsErr := u.repo.ExistsByName(ctx, req.Name, id)
		if existsErr != nil {
			return dto.ContactRoleResponse{}, existsErr
		}
		if exists {
			return dto.ContactRoleResponse{}, errors.New("contact role with this name already exists")
		}
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.BadgeColor != "" {
		role.BadgeColor = req.BadgeColor
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, role); err != nil {
		return dto.ContactRoleResponse{}, err
	}

	return mapper.ToContactRoleResponse(role), nil
}

func (u *contactRoleUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("contact role not found")
	}
	return u.repo.Delete(ctx, id)
}

func generateContactRoleCode(name, roleID string) string {
	base := normalizeCodeBase(name, "ROLE")
	return fmt.Sprintf("%s-%s", base, strings.Split(roleID, "-")[0])
}
