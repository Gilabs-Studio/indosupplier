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
	ErrBusinessTypeNotFound      = errors.New("business type not found")
	ErrBusinessTypeAlreadyExists = errors.New("business type with this name already exists")
)

// BusinessTypeUsecase defines the interface for business type business logic
type BusinessTypeUsecase interface {
	List(ctx context.Context, req *dto.ListBusinessTypesRequest) ([]dto.BusinessTypeResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.BusinessTypeResponse, error)
	Create(ctx context.Context, req *dto.CreateBusinessTypeRequest) (*dto.BusinessTypeResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateBusinessTypeRequest) (*dto.BusinessTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type businessTypeUsecase struct {
	businessTypeRepo repositories.BusinessTypeRepository
}

// NewBusinessTypeUsecase creates a new BusinessTypeUsecase
func NewBusinessTypeUsecase(businessTypeRepo repositories.BusinessTypeRepository) BusinessTypeUsecase {
	return &businessTypeUsecase{businessTypeRepo: businessTypeRepo}
}

func (u *businessTypeUsecase) List(ctx context.Context, req *dto.ListBusinessTypesRequest) ([]dto.BusinessTypeResponse, *utils.PaginationResult, error) {
	businessTypes, total, err := u.businessTypeRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToBusinessTypeResponses(businessTypes)

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

func (u *businessTypeUsecase) GetByID(ctx context.Context, id string) (*dto.BusinessTypeResponse, error) {
	businessType, err := u.businessTypeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessTypeNotFound
		}
		return nil, err
	}

	return mapper.ToBusinessTypeResponse(businessType), nil
}

func (u *businessTypeUsecase) Create(ctx context.Context, req *dto.CreateBusinessTypeRequest) (*dto.BusinessTypeResponse, error) {
	existing, err := u.businessTypeRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrBusinessTypeAlreadyExists
	}

	businessType := mapper.BusinessTypeFromCreateRequest(req)
	if err := u.businessTypeRepo.Create(ctx, businessType); err != nil {
		return nil, err
	}

	return mapper.ToBusinessTypeResponse(businessType), nil
}

func (u *businessTypeUsecase) Update(ctx context.Context, id string, req *dto.UpdateBusinessTypeRequest) (*dto.BusinessTypeResponse, error) {
	businessType, err := u.businessTypeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessTypeNotFound
		}
		return nil, err
	}

	if req.Name != "" && req.Name != businessType.Name {
		existing, err := u.businessTypeRepo.FindByName(ctx, req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrBusinessTypeAlreadyExists
		}
		businessType.Name = req.Name
	}

	if req.Description != "" {
		businessType.Description = req.Description
	}
	if req.IsActive != nil {
		businessType.IsActive = *req.IsActive
	}

	if err := u.businessTypeRepo.Update(ctx, businessType); err != nil {
		return nil, err
	}

	return mapper.ToBusinessTypeResponse(businessType), nil
}

func (u *businessTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.businessTypeRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBusinessTypeNotFound
		}
		return err
	}

	return u.businessTypeRepo.Delete(ctx, id)
}
