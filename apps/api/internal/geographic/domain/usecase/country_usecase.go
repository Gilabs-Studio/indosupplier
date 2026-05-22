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
	ErrCountryNotFound      = errors.New("country not found")
	ErrCountryAlreadyExists = errors.New("country with this code already exists")
	ErrCountryHasProvinces  = errors.New("cannot delete country with existing provinces")
)

// CountryUsecase defines the interface for country business logic
type CountryUsecase interface {
	List(ctx context.Context, req *dto.ListCountriesRequest) ([]dto.CountryResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.CountryResponse, error)
	Create(ctx context.Context, req *dto.CreateCountryRequest) (*dto.CountryResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCountryRequest) (*dto.CountryResponse, error)
	Delete(ctx context.Context, id string) error
}

type countryUsecase struct {
	countryRepo repositories.CountryRepository
}

// NewCountryUsecase creates a new CountryUsecase
func NewCountryUsecase(countryRepo repositories.CountryRepository) CountryUsecase {
	return &countryUsecase{countryRepo: countryRepo}
}

func (u *countryUsecase) List(ctx context.Context, req *dto.ListCountriesRequest) ([]dto.CountryResponse, *utils.PaginationResult, error) {
	countries, total, err := u.countryRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToCountryResponses(countries)

	// Calculate pagination
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

func (u *countryUsecase) GetByID(ctx context.Context, id string) (*dto.CountryResponse, error) {
	country, err := u.countryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}

	return mapper.ToCountryResponse(country), nil
}

func (u *countryUsecase) Create(ctx context.Context, req *dto.CreateCountryRequest) (*dto.CountryResponse, error) {
	// Check if code already exists
	existing, err := u.countryRepo.FindByCode(ctx, req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCountryAlreadyExists
	}

	country := mapper.CountryFromCreateRequest(req)
	if err := u.countryRepo.Create(ctx, country); err != nil {
		return nil, err
	}

	return mapper.ToCountryResponse(country), nil
}

func (u *countryUsecase) Update(ctx context.Context, id string, req *dto.UpdateCountryRequest) (*dto.CountryResponse, error) {
	country, err := u.countryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}

	// Check if code already exists (for different country)
	if req.Code != "" && req.Code != country.Code {
		existing, err := u.countryRepo.FindByCode(ctx, req.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrCountryAlreadyExists
		}
		country.Code = req.Code
	}

	if req.Name != "" {
		country.Name = req.Name
	}
	if req.PhoneCode != "" {
		country.PhoneCode = req.PhoneCode
	}
	if req.IsActive != nil {
		country.IsActive = *req.IsActive
	}

	if err := u.countryRepo.Update(ctx, country); err != nil {
		return nil, err
	}

	return mapper.ToCountryResponse(country), nil
}

func (u *countryUsecase) Delete(ctx context.Context, id string) error {
	// Check if country exists
	_, err := u.countryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCountryNotFound
		}
		return err
	}

	// Check if country has provinces
	hasProvinces, err := u.countryRepo.HasProvinces(ctx, id)
	if err != nil {
		return err
	}
	if hasProvinces {
		return ErrCountryHasProvinces
	}

	return u.countryRepo.Delete(ctx, id)
}
