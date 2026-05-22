package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrDivisionNotFound      = errors.New("division not found")
	ErrDivisionAlreadyExists = errors.New("division with this name already exists")
)

// DivisionUsecase defines the interface for division business logic
type DivisionUsecase interface {
	List(ctx context.Context, req *dto.ListDivisionsRequest) ([]dto.DivisionResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.DivisionResponse, error)
	Create(ctx context.Context, req *dto.CreateDivisionRequest) (*dto.DivisionResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateDivisionRequest) (*dto.DivisionResponse, error)
	Delete(ctx context.Context, id string) error
}

type divisionUsecase struct {
	divisionRepo repositories.DivisionRepository
}

// NewDivisionUsecase creates a new DivisionUsecase
func NewDivisionUsecase(divisionRepo repositories.DivisionRepository) DivisionUsecase {
	return &divisionUsecase{divisionRepo: divisionRepo}
}

func (u *divisionUsecase) List(ctx context.Context, req *dto.ListDivisionsRequest) ([]dto.DivisionResponse, *utils.PaginationResult, error) {
	divisions, total, err := u.divisionRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToDivisionResponses(divisions)

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *divisionUsecase) GetByID(ctx context.Context, id string) (*dto.DivisionResponse, error) {
	division, err := u.divisionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDivisionNotFound
		}
		return nil, err
	}

	return mapper.ToDivisionResponse(division), nil
}

func (u *divisionUsecase) Create(ctx context.Context, req *dto.CreateDivisionRequest) (*dto.DivisionResponse, error) {
	// Check if name already exists
	existing, err := u.divisionRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDivisionAlreadyExists
	}

	division := mapper.DivisionFromCreateRequest(req)
	if err := u.divisionRepo.Create(ctx, division); err != nil {
		return nil, err
	}

	return mapper.ToDivisionResponse(division), nil
}

func (u *divisionUsecase) Update(ctx context.Context, id string, req *dto.UpdateDivisionRequest) (*dto.DivisionResponse, error) {
	division, err := u.divisionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDivisionNotFound
		}
		return nil, err
	}

	// Check if name already exists (for different division)
	if req.Name != "" && req.Name != division.Name {
		existing, err := u.divisionRepo.FindByName(ctx, req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrDivisionAlreadyExists
		}
		division.Name = req.Name
	}

	if req.Description != "" {
		division.Description = req.Description
	}
	if req.IsActive != nil {
		division.IsActive = *req.IsActive
	}

	if err := u.divisionRepo.Update(ctx, division); err != nil {
		return nil, err
	}

	return mapper.ToDivisionResponse(division), nil
}

func (u *divisionUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.divisionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDivisionNotFound
		}
		return err
	}

	return u.divisionRepo.Delete(ctx, id)
}
