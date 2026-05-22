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

// UnitOfMeasureUsecase defines the interface for unit of measure business logic
type UnitOfMeasureUsecase interface {
	Create(ctx context.Context, req dto.CreateUnitOfMeasureRequest) (dto.UnitOfMeasureResponse, error)
	GetByID(ctx context.Context, id string) (dto.UnitOfMeasureResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.UnitOfMeasureResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateUnitOfMeasureRequest) (dto.UnitOfMeasureResponse, error)
	Delete(ctx context.Context, id string) error
}

type unitOfMeasureUsecase struct {
	repo repositories.UnitOfMeasureRepository
}

// NewUnitOfMeasureUsecase creates a new UnitOfMeasureUsecase
func NewUnitOfMeasureUsecase(repo repositories.UnitOfMeasureRepository) UnitOfMeasureUsecase {
	return &unitOfMeasureUsecase{repo: repo}
}

func (u *unitOfMeasureUsecase) Create(ctx context.Context, req dto.CreateUnitOfMeasureRequest) (dto.UnitOfMeasureResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	uom := &models.UnitOfMeasure{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Symbol:      req.Symbol,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, uom); err != nil {
		return dto.UnitOfMeasureResponse{}, err
	}

	return mapper.ToUnitOfMeasureResponse(uom), nil
}

func (u *unitOfMeasureUsecase) GetByID(ctx context.Context, id string) (dto.UnitOfMeasureResponse, error) {
	uom, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.UnitOfMeasureResponse{}, err
	}
	return mapper.ToUnitOfMeasureResponse(uom), nil
}

func (u *unitOfMeasureUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.UnitOfMeasureResponse, int64, error) {
	uoms, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToUnitOfMeasureResponseList(uoms), total, nil
}

func (u *unitOfMeasureUsecase) Update(ctx context.Context, id string, req dto.UpdateUnitOfMeasureRequest) (dto.UnitOfMeasureResponse, error) {
	uom, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.UnitOfMeasureResponse{}, err
	}

	if req.Name != "" {
		uom.Name = req.Name
	}
	if req.Symbol != "" {
		uom.Symbol = req.Symbol
	}
	if req.Description != "" {
		uom.Description = req.Description
	}
	if req.IsActive != nil {
		uom.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, uom); err != nil {
		return dto.UnitOfMeasureResponse{}, err
	}

	return mapper.ToUnitOfMeasureResponse(uom), nil
}

func (u *unitOfMeasureUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("unit of measure not found")
	}
	return u.repo.Delete(ctx, id)
}
