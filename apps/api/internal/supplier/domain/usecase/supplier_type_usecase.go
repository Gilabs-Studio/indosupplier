package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
	"github.com/gilabs/gims/api/internal/supplier/domain/mapper"
	"github.com/google/uuid"
)

// SupplierTypeUsecase defines the interface for supplier type business logic
type SupplierTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateSupplierTypeRequest) (dto.SupplierTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.SupplierTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.SupplierTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateSupplierTypeRequest) (dto.SupplierTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type supplierTypeUsecase struct {
	repo repositories.SupplierTypeRepository
}

// NewSupplierTypeUsecase creates a new SupplierTypeUsecase
func NewSupplierTypeUsecase(repo repositories.SupplierTypeRepository) SupplierTypeUsecase {
	return &supplierTypeUsecase{repo: repo}
}

func (u *supplierTypeUsecase) Create(ctx context.Context, req dto.CreateSupplierTypeRequest) (dto.SupplierTypeResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	supplierType := &models.SupplierType{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, supplierType); err != nil {
		return dto.SupplierTypeResponse{}, err
	}

	return mapper.ToSupplierTypeResponse(supplierType), nil
}

func (u *supplierTypeUsecase) GetByID(ctx context.Context, id string) (dto.SupplierTypeResponse, error) {
	supplierType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierTypeResponse{}, err
	}
	return mapper.ToSupplierTypeResponse(supplierType), nil
}

func (u *supplierTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.SupplierTypeResponse, int64, error) {
	supplierTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToSupplierTypeResponseList(supplierTypes), total, nil
}

func (u *supplierTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateSupplierTypeRequest) (dto.SupplierTypeResponse, error) {
	supplierType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SupplierTypeResponse{}, err
	}

	if req.Name != "" {
		supplierType.Name = req.Name
	}
	if req.Description != "" {
		supplierType.Description = req.Description
	}
	if req.IsActive != nil {
		supplierType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, supplierType); err != nil {
		return dto.SupplierTypeResponse{}, err
	}

	return mapper.ToSupplierTypeResponse(supplierType), nil
}

func (u *supplierTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("supplier type not found")
	}
	return u.repo.Delete(ctx, id)
}
