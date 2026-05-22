package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrBusinessUnitNotFound      = errors.New("business unit not found")
	ErrBusinessUnitAlreadyExists = errors.New("business unit with this name already exists")
)

// BusinessUnitUsecase defines the interface for business unit business logic
type BusinessUnitUsecase interface {
	List(ctx context.Context, req *dto.ListBusinessUnitsRequest) ([]dto.BusinessUnitResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.BusinessUnitResponse, error)
	Create(ctx context.Context, req *dto.CreateBusinessUnitRequest) (*dto.BusinessUnitResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateBusinessUnitRequest) (*dto.BusinessUnitResponse, error)
	Delete(ctx context.Context, id string) error
}

type businessUnitUsecase struct {
	businessUnitRepo repositories.BusinessUnitRepository
}

// NewBusinessUnitUsecase creates a new BusinessUnitUsecase
func NewBusinessUnitUsecase(businessUnitRepo repositories.BusinessUnitRepository) BusinessUnitUsecase {
	return &businessUnitUsecase{businessUnitRepo: businessUnitRepo}
}

func (u *businessUnitUsecase) List(ctx context.Context, req *dto.ListBusinessUnitsRequest) ([]dto.BusinessUnitResponse, *utils.PaginationResult, error) {
	businessUnits, total, err := u.businessUnitRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToBusinessUnitResponses(businessUnits)

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

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *businessUnitUsecase) GetByID(ctx context.Context, id string) (*dto.BusinessUnitResponse, error) {
	businessUnit, err := u.businessUnitRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessUnitNotFound
		}
		return nil, err
	}

	return mapper.ToBusinessUnitResponse(businessUnit), nil
}

func (u *businessUnitUsecase) Create(ctx context.Context, req *dto.CreateBusinessUnitRequest) (*dto.BusinessUnitResponse, error) {
	existing, err := u.businessUnitRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrBusinessUnitAlreadyExists
	}

	businessUnit := mapper.BusinessUnitFromCreateRequest(req)
	if err := u.businessUnitRepo.Create(ctx, businessUnit); err != nil {
		return nil, err
	}

	return mapper.ToBusinessUnitResponse(businessUnit), nil
}

func (u *businessUnitUsecase) Update(ctx context.Context, id string, req *dto.UpdateBusinessUnitRequest) (*dto.BusinessUnitResponse, error) {
	businessUnit, err := u.businessUnitRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessUnitNotFound
		}
		return nil, err
	}

	if req.Name != "" && req.Name != businessUnit.Name {
		existing, err := u.businessUnitRepo.FindByName(ctx, req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrBusinessUnitAlreadyExists
		}
		businessUnit.Name = req.Name
	}

	if req.Description != "" {
		businessUnit.Description = req.Description
	}
	if req.IsActive != nil {
		businessUnit.IsActive = *req.IsActive
	}

	if err := u.businessUnitRepo.Update(ctx, businessUnit); err != nil {
		return nil, err
	}

	return mapper.ToBusinessUnitResponse(businessUnit), nil
}

func (u *businessUnitUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.businessUnitRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBusinessUnitNotFound
		}
		return err
	}

	return u.businessUnitRepo.Delete(ctx, id)
}
