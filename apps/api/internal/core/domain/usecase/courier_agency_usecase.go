package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/mapper"
	"github.com/google/uuid"
)

// CourierAgencyUsecase defines the interface for courier agency business logic
type CourierAgencyUsecase interface {
	Create(ctx context.Context, req dto.CreateCourierAgencyRequest) (dto.CourierAgencyResponse, error)
	GetByID(ctx context.Context, id string) (dto.CourierAgencyResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.CourierAgencyResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateCourierAgencyRequest) (dto.CourierAgencyResponse, error)
	Delete(ctx context.Context, id string) error
}

type courierAgencyUsecase struct {
	repo repositories.CourierAgencyRepository
}

// NewCourierAgencyUsecase creates a new CourierAgencyUsecase
func NewCourierAgencyUsecase(repo repositories.CourierAgencyRepository) CourierAgencyUsecase {
	return &courierAgencyUsecase{repo: repo}
}

func (u *courierAgencyUsecase) Create(ctx context.Context, req dto.CreateCourierAgencyRequest) (dto.CourierAgencyResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	id := uuid.New().String()
	slug := "GEN"
	if len(req.Name) >= 3 {
		slug = req.Name[:3]
	} else {
		slug = req.Name
	}
	code := fmt.Sprintf("CA-%s-%s", slug, id[:4])

	courierAgency := &models.CourierAgency{
		ID:          id,
		Code:        code,
		Name:        req.Name,
		Description: req.Description,
		Phone:       req.Phone,
		Address:     req.Address,
		TrackingURL: req.TrackingURL,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, courierAgency); err != nil {
		return dto.CourierAgencyResponse{}, err
	}

	return mapper.ToCourierAgencyResponse(courierAgency), nil
}

func (u *courierAgencyUsecase) GetByID(ctx context.Context, id string) (dto.CourierAgencyResponse, error) {
	courierAgency, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CourierAgencyResponse{}, err
	}
	return mapper.ToCourierAgencyResponse(courierAgency), nil
}

func (u *courierAgencyUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.CourierAgencyResponse, int64, error) {
	courierAgencies, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToCourierAgencyResponseList(courierAgencies), total, nil
}

func (u *courierAgencyUsecase) Update(ctx context.Context, id string, req dto.UpdateCourierAgencyRequest) (dto.CourierAgencyResponse, error) {
	courierAgency, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CourierAgencyResponse{}, err
	}

	if req.Code != "" {
		courierAgency.Code = req.Code
	}
	if req.Name != "" {
		courierAgency.Name = req.Name
	}
	if req.Description != "" {
		courierAgency.Description = req.Description
	}
	if req.Phone != "" {
		courierAgency.Phone = req.Phone
	}
	if req.Address != "" {
		courierAgency.Address = req.Address
	}
	if req.TrackingURL != "" {
		courierAgency.TrackingURL = req.TrackingURL
	}
	if req.IsActive != nil {
		courierAgency.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, courierAgency); err != nil {
		return dto.CourierAgencyResponse{}, err
	}

	return mapper.ToCourierAgencyResponse(courierAgency), nil
}

func (u *courierAgencyUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("courier agency not found")
	}
	return u.repo.Delete(ctx, id)
}
