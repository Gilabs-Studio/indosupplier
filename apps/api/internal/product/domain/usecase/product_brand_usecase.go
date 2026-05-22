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

// ProductBrandUsecase defines the interface for product brand business logic
type ProductBrandUsecase interface {
	Create(ctx context.Context, req dto.CreateProductBrandRequest) (dto.ProductBrandResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductBrandResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ProductBrandResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProductBrandRequest) (dto.ProductBrandResponse, error)
	Delete(ctx context.Context, id string) error
}

type productBrandUsecase struct {
	repo repositories.ProductBrandRepository
}

// NewProductBrandUsecase creates a new ProductBrandUsecase
func NewProductBrandUsecase(repo repositories.ProductBrandRepository) ProductBrandUsecase {
	return &productBrandUsecase{repo: repo}
}

func (u *productBrandUsecase) Create(ctx context.Context, req dto.CreateProductBrandRequest) (dto.ProductBrandResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	brand := &models.ProductBrand{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, brand); err != nil {
		return dto.ProductBrandResponse{}, err
	}

	return mapper.ToProductBrandResponse(brand), nil
}

func (u *productBrandUsecase) GetByID(ctx context.Context, id string) (dto.ProductBrandResponse, error) {
	brand, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductBrandResponse{}, err
	}
	return mapper.ToProductBrandResponse(brand), nil
}

func (u *productBrandUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ProductBrandResponse, int64, error) {
	brands, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProductBrandResponseList(brands), total, nil
}

func (u *productBrandUsecase) Update(ctx context.Context, id string, req dto.UpdateProductBrandRequest) (dto.ProductBrandResponse, error) {
	brand, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductBrandResponse{}, err
	}

	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Description != "" {
		brand.Description = req.Description
	}
	if req.IsActive != nil {
		brand.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, brand); err != nil {
		return dto.ProductBrandResponse{}, err
	}

	return mapper.ToProductBrandResponse(brand), nil
}

func (u *productBrandUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("product brand not found")
	}
	return u.repo.Delete(ctx, id)
}
