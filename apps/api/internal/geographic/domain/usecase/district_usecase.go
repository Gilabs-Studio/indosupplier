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
	ErrDistrictNotFound      = errors.New("district not found")
	ErrDistrictAlreadyExists = errors.New("district with this code already exists")
	ErrDistrictHasVillages   = errors.New("cannot delete district with existing villages")
)

// DistrictUsecase defines the interface for district business logic
type DistrictUsecase interface {
	List(ctx context.Context, req *dto.ListDistrictsRequest) ([]dto.DistrictResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.DistrictResponse, error)
	Create(ctx context.Context, req *dto.CreateDistrictRequest) (*dto.DistrictResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateDistrictRequest) (*dto.DistrictResponse, error)
	Delete(ctx context.Context, id string) error
}

type districtUsecase struct {
	districtRepo repositories.DistrictRepository
	cityRepo     repositories.CityRepository
}

// NewDistrictUsecase creates a new DistrictUsecase
func NewDistrictUsecase(
	districtRepo repositories.DistrictRepository,
	cityRepo repositories.CityRepository,
) DistrictUsecase {
	return &districtUsecase{
		districtRepo: districtRepo,
		cityRepo:     cityRepo,
	}
}

func (u *districtUsecase) List(ctx context.Context, req *dto.ListDistrictsRequest) ([]dto.DistrictResponse, *utils.PaginationResult, error) {
	districts, total, err := u.districtRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToDistrictResponses(districts)

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

func (u *districtUsecase) GetByID(ctx context.Context, id string) (*dto.DistrictResponse, error) {
	district, err := u.districtRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDistrictNotFound
		}
		return nil, err
	}

	return mapper.ToDistrictResponse(district), nil
}

func (u *districtUsecase) Create(ctx context.Context, req *dto.CreateDistrictRequest) (*dto.DistrictResponse, error) {
	// Validate city exists
	_, err := u.cityRepo.FindByID(ctx, req.CityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCityNotFound
		}
		return nil, err
	}

	// Check if code already exists
	existing, err := u.districtRepo.FindByCode(ctx, req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDistrictAlreadyExists
	}

	district := mapper.DistrictFromCreateRequest(req)
	if err := u.districtRepo.Create(ctx, district); err != nil {
		return nil, err
	}

	district, err = u.districtRepo.FindByID(ctx, district.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToDistrictResponse(district), nil
}

func (u *districtUsecase) Update(ctx context.Context, id string, req *dto.UpdateDistrictRequest) (*dto.DistrictResponse, error) {
	district, err := u.districtRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDistrictNotFound
		}
		return nil, err
	}

	if req.CityID != "" && req.CityID != district.CityID {
		_, err := u.cityRepo.FindByID(ctx, req.CityID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCityNotFound
			}
			return nil, err
		}
		district.CityID = req.CityID
	}

	if req.Code != "" && req.Code != district.Code {
		existing, err := u.districtRepo.FindByCode(ctx, req.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrDistrictAlreadyExists
		}
		district.Code = req.Code
	}

	if req.Name != "" {
		district.Name = req.Name
	}
	if req.IsActive != nil {
		district.IsActive = *req.IsActive
	}

	if err := u.districtRepo.Update(ctx, district); err != nil {
		return nil, err
	}

	district, err = u.districtRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToDistrictResponse(district), nil
}

func (u *districtUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.districtRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDistrictNotFound
		}
		return err
	}

	hasVillages, err := u.districtRepo.HasVillages(ctx, id)
	if err != nil {
		return err
	}
	if hasVillages {
		return ErrDistrictHasVillages
	}

	return u.districtRepo.Delete(ctx, id)
}
