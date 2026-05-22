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

// PackagingUsecase defines the interface for packaging business logic
type PackagingUsecase interface {
	Create(ctx context.Context, req dto.CreatePackagingRequest) (dto.PackagingResponse, error)
	GetByID(ctx context.Context, id string) (dto.PackagingResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.PackagingResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdatePackagingRequest) (dto.PackagingResponse, error)
	Delete(ctx context.Context, id string) error
}

type packagingUsecase struct {
	repo repositories.PackagingRepository
}

// NewPackagingUsecase creates a new PackagingUsecase
func NewPackagingUsecase(repo repositories.PackagingRepository) PackagingUsecase {
	return &packagingUsecase{repo: repo}
}

func (u *packagingUsecase) Create(ctx context.Context, req dto.CreatePackagingRequest) (dto.PackagingResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	packaging := &models.Packaging{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, packaging); err != nil {
		return dto.PackagingResponse{}, err
	}

	return mapper.ToPackagingResponse(packaging), nil
}

func (u *packagingUsecase) GetByID(ctx context.Context, id string) (dto.PackagingResponse, error) {
	packaging, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	return mapper.ToPackagingResponse(packaging), nil
}

func (u *packagingUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.PackagingResponse, int64, error) {
	packagings, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToPackagingResponseList(packagings), total, nil
}

func (u *packagingUsecase) Update(ctx context.Context, id string, req dto.UpdatePackagingRequest) (dto.PackagingResponse, error) {
	packaging, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PackagingResponse{}, err
	}

	if req.Name != "" {
		packaging.Name = req.Name
	}
	if req.Description != "" {
		packaging.Description = req.Description
	}
	if req.IsActive != nil {
		packaging.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, packaging); err != nil {
		return dto.PackagingResponse{}, err
	}

	return mapper.ToPackagingResponse(packaging), nil
}

func (u *packagingUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("packaging not found")
	}
	return u.repo.Delete(ctx, id)
}
