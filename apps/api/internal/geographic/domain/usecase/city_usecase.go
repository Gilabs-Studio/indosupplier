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
	ErrCityNotFound      = errors.New("city not found")
	ErrCityAlreadyExists = errors.New("city with this code already exists")
	ErrCityHasDistricts  = errors.New("cannot delete city with existing districts")
)

// CityUsecase defines the interface for city business logic
type CityUsecase interface {
	List(ctx context.Context, req *dto.ListCitiesRequest) ([]dto.CityResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.CityResponse, error)
	Create(ctx context.Context, req *dto.CreateCityRequest) (*dto.CityResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCityRequest) (*dto.CityResponse, error)
	Delete(ctx context.Context, id string) error
}

type cityUsecase struct {
	cityRepo     repositories.CityRepository
	provinceRepo repositories.ProvinceRepository
}

// NewCityUsecase creates a new CityUsecase
func NewCityUsecase(
	cityRepo repositories.CityRepository,
	provinceRepo repositories.ProvinceRepository,
) CityUsecase {
	return &cityUsecase{
		cityRepo:     cityRepo,
		provinceRepo: provinceRepo,
	}
}

func (u *cityUsecase) List(ctx context.Context, req *dto.ListCitiesRequest) ([]dto.CityResponse, *utils.PaginationResult, error) {
	cities, total, err := u.cityRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToCityResponses(cities)

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

func (u *cityUsecase) GetByID(ctx context.Context, id string) (*dto.CityResponse, error) {
	city, err := u.cityRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCityNotFound
		}
		return nil, err
	}

	return mapper.ToCityResponse(city), nil
}

func (u *cityUsecase) Create(ctx context.Context, req *dto.CreateCityRequest) (*dto.CityResponse, error) {
	// Validate province exists
	_, err := u.provinceRepo.FindByID(ctx, req.ProvinceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProvinceNotFound
		}
		return nil, err
	}

	// Check if code already exists
	existing, err := u.cityRepo.FindByCode(ctx, req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCityAlreadyExists
	}

	city := mapper.CityFromCreateRequest(req)
	if err := u.cityRepo.Create(ctx, city); err != nil {
		return nil, err
	}

	city, err = u.cityRepo.FindByID(ctx, city.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCityResponse(city), nil
}

func (u *cityUsecase) Update(ctx context.Context, id string, req *dto.UpdateCityRequest) (*dto.CityResponse, error) {
	city, err := u.cityRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCityNotFound
		}
		return nil, err
	}

	if req.ProvinceID != "" && req.ProvinceID != city.ProvinceID {
		_, err := u.provinceRepo.FindByID(ctx, req.ProvinceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProvinceNotFound
			}
			return nil, err
		}
		city.ProvinceID = req.ProvinceID
	}

	if req.Code != "" && req.Code != city.Code {
		existing, err := u.cityRepo.FindByCode(ctx, req.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrCityAlreadyExists
		}
		city.Code = req.Code
	}

	if req.Name != "" {
		city.Name = req.Name
	}
	if req.Type != "" {
		city.Type = req.Type
	}
	if req.IsActive != nil {
		city.IsActive = *req.IsActive
	}

	if err := u.cityRepo.Update(ctx, city); err != nil {
		return nil, err
	}

	city, err = u.cityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToCityResponse(city), nil
}

func (u *cityUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.cityRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCityNotFound
		}
		return err
	}

	hasDistricts, err := u.cityRepo.HasDistricts(ctx, id)
	if err != nil {
		return err
	}
	if hasDistricts {
		return ErrCityHasDistricts
	}

	return u.cityRepo.Delete(ctx, id)
}
