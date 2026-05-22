package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/geographic/data/repositories"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrVillageNotFound      = errors.New("village not found")
	ErrVillageAlreadyExists = errors.New("village with this code already exists")
)

// VillageUsecase defines the interface for village business logic
type VillageUsecase interface {
	List(ctx context.Context, req *dto.ListVillagesRequest) ([]dto.VillageResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.VillageResponse, error)
	Create(ctx context.Context, req *dto.CreateVillageRequest) (*dto.VillageResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateVillageRequest) (*dto.VillageResponse, error)
	Delete(ctx context.Context, id string) error
}

type villageUsecase struct {
	villageRepo  repositories.VillageRepository
	districtRepo repositories.DistrictRepository
}

// NewVillageUsecase creates a new VillageUsecase
func NewVillageUsecase(
	villageRepo repositories.VillageRepository,
	districtRepo repositories.DistrictRepository,
) VillageUsecase {
	return &villageUsecase{
		villageRepo:  villageRepo,
		districtRepo: districtRepo,
	}
}

func (u *villageUsecase) List(ctx context.Context, req *dto.ListVillagesRequest) ([]dto.VillageResponse, *utils.PaginationResult, error) {
	villages, total, err := u.villageRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToVillageResponses(villages)

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

func (u *villageUsecase) GetByID(ctx context.Context, id string) (*dto.VillageResponse, error) {
	village, err := u.villageRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVillageNotFound
		}
		return nil, err
	}

	return mapper.ToVillageResponse(village), nil
}

func (u *villageUsecase) Create(ctx context.Context, req *dto.CreateVillageRequest) (*dto.VillageResponse, error) {
	// Validate district exists
	_, err := u.districtRepo.FindByID(ctx, req.DistrictID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDistrictNotFound
		}
		return nil, err
	}

	// Check if code already exists
	existing, err := u.villageRepo.FindByCode(ctx, req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrVillageAlreadyExists
	}

	village := mapper.VillageFromCreateRequest(req)
	if err := u.villageRepo.Create(ctx, village); err != nil {
		return nil, err
	}

	village, err = u.villageRepo.FindByID(ctx, village.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToVillageResponse(village), nil
}

func (u *villageUsecase) Update(ctx context.Context, id string, req *dto.UpdateVillageRequest) (*dto.VillageResponse, error) {
	village, err := u.villageRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVillageNotFound
		}
		return nil, err
	}

	if req.DistrictID != "" && req.DistrictID != village.DistrictID {
		_, err := u.districtRepo.FindByID(ctx, req.DistrictID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrDistrictNotFound
			}
			return nil, err
		}
		village.DistrictID = req.DistrictID
	}

	if req.Code != "" && req.Code != village.Code {
		existing, err := u.villageRepo.FindByCode(ctx, req.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrVillageAlreadyExists
		}
		village.Code = req.Code
	}

	if req.Name != "" {
		village.Name = req.Name
	}
	if req.PostalCode != "" {
		village.PostalCode = req.PostalCode
	}
	if req.Type != "" {
		village.Type = req.Type
	}
	if req.IsActive != nil {
		village.IsActive = *req.IsActive
	}

	if err := u.villageRepo.Update(ctx, village); err != nil {
		return nil, err
	}

	village, err = u.villageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToVillageResponse(village), nil
}

func (u *villageUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.villageRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVillageNotFound
		}
		return err
	}

	return u.villageRepo.Delete(ctx, id)
}
