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

// ProcurementTypeUsecase defines the interface for procurement type business logic
type ProcurementTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateProcurementTypeRequest) (dto.ProcurementTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProcurementTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ProcurementTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProcurementTypeRequest) (dto.ProcurementTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type procurementTypeUsecase struct {
	repo repositories.ProcurementTypeRepository
}

// NewProcurementTypeUsecase creates a new ProcurementTypeUsecase
func NewProcurementTypeUsecase(repo repositories.ProcurementTypeRepository) ProcurementTypeUsecase {
	return &procurementTypeUsecase{repo: repo}
}

func (u *procurementTypeUsecase) Create(ctx context.Context, req dto.CreateProcurementTypeRequest) (dto.ProcurementTypeResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	procurementType := &models.ProcurementType{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, procurementType); err != nil {
		return dto.ProcurementTypeResponse{}, err
	}

	return mapper.ToProcurementTypeResponse(procurementType), nil
}

func (u *procurementTypeUsecase) GetByID(ctx context.Context, id string) (dto.ProcurementTypeResponse, error) {
	procurementType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProcurementTypeResponse{}, err
	}
	return mapper.ToProcurementTypeResponse(procurementType), nil
}

func (u *procurementTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ProcurementTypeResponse, int64, error) {
	procurementTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProcurementTypeResponseList(procurementTypes), total, nil
}

func (u *procurementTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateProcurementTypeRequest) (dto.ProcurementTypeResponse, error) {
	procurementType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProcurementTypeResponse{}, err
	}

	if req.Name != "" {
		procurementType.Name = req.Name
	}
	if req.Description != "" {
		procurementType.Description = req.Description
	}
	if req.IsActive != nil {
		procurementType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, procurementType); err != nil {
		return dto.ProcurementTypeResponse{}, err
	}

	return mapper.ToProcurementTypeResponse(procurementType), nil
}

func (u *procurementTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("procurement type not found")
	}
	return u.repo.Delete(ctx, id)
}
