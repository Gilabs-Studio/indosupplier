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

// ProductSegmentUsecase defines the interface for product segment business logic
type ProductSegmentUsecase interface {
	Create(ctx context.Context, req dto.CreateProductSegmentRequest) (dto.ProductSegmentResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductSegmentResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ProductSegmentResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProductSegmentRequest) (dto.ProductSegmentResponse, error)
	Delete(ctx context.Context, id string) error
}

type productSegmentUsecase struct {
	repo repositories.ProductSegmentRepository
}

// NewProductSegmentUsecase creates a new ProductSegmentUsecase
func NewProductSegmentUsecase(repo repositories.ProductSegmentRepository) ProductSegmentUsecase {
	return &productSegmentUsecase{repo: repo}
}

func (u *productSegmentUsecase) Create(ctx context.Context, req dto.CreateProductSegmentRequest) (dto.ProductSegmentResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	segment := &models.ProductSegment{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, segment); err != nil {
		return dto.ProductSegmentResponse{}, err
	}

	return mapper.ToProductSegmentResponse(segment), nil
}

func (u *productSegmentUsecase) GetByID(ctx context.Context, id string) (dto.ProductSegmentResponse, error) {
	segment, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductSegmentResponse{}, err
	}
	return mapper.ToProductSegmentResponse(segment), nil
}

func (u *productSegmentUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ProductSegmentResponse, int64, error) {
	segments, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProductSegmentResponseList(segments), total, nil
}

func (u *productSegmentUsecase) Update(ctx context.Context, id string, req dto.UpdateProductSegmentRequest) (dto.ProductSegmentResponse, error) {
	segment, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductSegmentResponse{}, err
	}

	if req.Name != "" {
		segment.Name = req.Name
	}
	if req.Description != "" {
		segment.Description = req.Description
	}
	if req.IsActive != nil {
		segment.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, segment); err != nil {
		return dto.ProductSegmentResponse{}, err
	}

	return mapper.ToProductSegmentResponse(segment), nil
}

func (u *productSegmentUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("product segment not found")
	}
	return u.repo.Delete(ctx, id)
}
