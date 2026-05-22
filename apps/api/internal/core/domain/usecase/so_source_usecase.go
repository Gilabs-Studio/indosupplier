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

// SOSourceUsecase defines the interface for SO source business logic
type SOSourceUsecase interface {
	Create(ctx context.Context, req dto.CreateSOSourceRequest) (dto.SOSourceResponse, error)
	GetByID(ctx context.Context, id string) (dto.SOSourceResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.SOSourceResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateSOSourceRequest) (dto.SOSourceResponse, error)
	Delete(ctx context.Context, id string) error
}

type soSourceUsecase struct {
	repo repositories.SOSourceRepository
}

// NewSOSourceUsecase creates a new SOSourceUsecase
func NewSOSourceUsecase(repo repositories.SOSourceRepository) SOSourceUsecase {
	return &soSourceUsecase{repo: repo}
}

func (u *soSourceUsecase) Create(ctx context.Context, req dto.CreateSOSourceRequest) (dto.SOSourceResponse, error) {
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
	code := fmt.Sprintf("SOS-%s-%s", slug, id[:4])

	soSource := &models.SOSource{
		ID:          id,
		Code:        code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, soSource); err != nil {
		return dto.SOSourceResponse{}, err
	}

	return mapper.ToSOSourceResponse(soSource), nil
}

func (u *soSourceUsecase) GetByID(ctx context.Context, id string) (dto.SOSourceResponse, error) {
	soSource, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SOSourceResponse{}, err
	}
	return mapper.ToSOSourceResponse(soSource), nil
}

func (u *soSourceUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.SOSourceResponse, int64, error) {
	soSources, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToSOSourceResponseList(soSources), total, nil
}

func (u *soSourceUsecase) Update(ctx context.Context, id string, req dto.UpdateSOSourceRequest) (dto.SOSourceResponse, error) {
	soSource, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.SOSourceResponse{}, err
	}

	if req.Code != "" {
		soSource.Code = req.Code
	}
	if req.Name != "" {
		soSource.Name = req.Name
	}
	if req.Description != "" {
		soSource.Description = req.Description
	}
	if req.IsActive != nil {
		soSource.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, soSource); err != nil {
		return dto.SOSourceResponse{}, err
	}

	return mapper.ToSOSourceResponse(soSource), nil
}

func (u *soSourceUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("SO source not found")
	}
	return u.repo.Delete(ctx, id)
}
