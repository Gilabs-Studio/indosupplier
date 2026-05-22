package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
	"github.com/gilabs/gims/api/internal/product/domain/mapper"
	"github.com/google/uuid"
)

// ProductTypeUsecase defines the interface for product type business logic
type ProductTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateProductTypeRequest) (dto.ProductTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ProductTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProductTypeRequest) (dto.ProductTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type productTypeUsecase struct {
	repo repositories.ProductTypeRepository
}

// NewProductTypeUsecase creates a new ProductTypeUsecase
func NewProductTypeUsecase(repo repositories.ProductTypeRepository) ProductTypeUsecase {
	return &productTypeUsecase{repo: repo}
}

func (u *productTypeUsecase) Create(ctx context.Context, req dto.CreateProductTypeRequest) (dto.ProductTypeResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	productType := &models.ProductType{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, productType); err != nil {
		return dto.ProductTypeResponse{}, err
	}

	return mapper.ToProductTypeResponse(productType), nil
}

func (u *productTypeUsecase) GetByID(ctx context.Context, id string) (dto.ProductTypeResponse, error) {
	productType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductTypeResponse{}, err
	}
	return mapper.ToProductTypeResponse(productType), nil
}

func (u *productTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ProductTypeResponse, int64, error) {
	productTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProductTypeResponseList(productTypes), total, nil
}

func (u *productTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateProductTypeRequest) (dto.ProductTypeResponse, error) {
	productType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductTypeResponse{}, err
	}

	if req.Name != "" {
		productType.Name = req.Name
	}
	if req.Description != "" {
		productType.Description = req.Description
	}
	if req.IsActive != nil {
		productType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, productType); err != nil {
		return dto.ProductTypeResponse{}, err
	}

	return mapper.ToProductTypeResponse(productType), nil
}

func (u *productTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("product type not found")
	}
	return u.repo.Delete(ctx, id)
}
