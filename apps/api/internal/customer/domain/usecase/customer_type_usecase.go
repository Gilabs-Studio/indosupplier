package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/customer/data/models"
	"github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/customer/domain/dto"
	"github.com/gilabs/gims/api/internal/customer/domain/mapper"
	"github.com/google/uuid"
)

// CustomerTypeUsecase defines the interface for customer type business logic
type CustomerTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateCustomerTypeRequest) (dto.CustomerTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.CustomerTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.CustomerTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateCustomerTypeRequest) (dto.CustomerTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type customerTypeUsecase struct {
	repo repositories.CustomerTypeRepository
}

// NewCustomerTypeUsecase creates a new CustomerTypeUsecase
func NewCustomerTypeUsecase(repo repositories.CustomerTypeRepository) CustomerTypeUsecase {
	return &customerTypeUsecase{repo: repo}
}

func (u *customerTypeUsecase) Create(ctx context.Context, req dto.CreateCustomerTypeRequest) (dto.CustomerTypeResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	customerType := &models.CustomerType{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, customerType); err != nil {
		return dto.CustomerTypeResponse{}, err
	}

	return mapper.ToCustomerTypeResponse(customerType), nil
}

func (u *customerTypeUsecase) GetByID(ctx context.Context, id string) (dto.CustomerTypeResponse, error) {
	customerType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CustomerTypeResponse{}, err
	}
	return mapper.ToCustomerTypeResponse(customerType), nil
}

func (u *customerTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.CustomerTypeResponse, int64, error) {
	customerTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToCustomerTypeResponseList(customerTypes), total, nil
}

func (u *customerTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateCustomerTypeRequest) (dto.CustomerTypeResponse, error) {
	customerType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CustomerTypeResponse{}, err
	}

	if req.Name != "" {
		customerType.Name = req.Name
	}
	if req.Description != "" {
		customerType.Description = req.Description
	}
	if req.IsActive != nil {
		customerType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, customerType); err != nil {
		return dto.CustomerTypeResponse{}, err
	}

	return mapper.ToCustomerTypeResponse(customerType), nil
}

func (u *customerTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("customer type not found")
	}
	return u.repo.Delete(ctx, id)
}
