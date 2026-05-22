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
	ErrProvinceNotFound      = errors.New("province not found")
	ErrProvinceAlreadyExists = errors.New("province with this code already exists")
	ErrProvinceHasCities     = errors.New("cannot delete province with existing cities")
)

// ProvinceUsecase defines the interface for province business logic
type ProvinceUsecase interface {
	List(ctx context.Context, req *dto.ListProvincesRequest) ([]dto.ProvinceResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.ProvinceResponse, error)
	Create(ctx context.Context, req *dto.CreateProvinceRequest) (*dto.ProvinceResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateProvinceRequest) (*dto.ProvinceResponse, error)
	Delete(ctx context.Context, id string) error
}

type provinceUsecase struct {
	provinceRepo repositories.ProvinceRepository
	countryRepo  repositories.CountryRepository
}

// NewProvinceUsecase creates a new ProvinceUsecase
func NewProvinceUsecase(
	provinceRepo repositories.ProvinceRepository,
	countryRepo repositories.CountryRepository,
) ProvinceUsecase {
	return &provinceUsecase{
		provinceRepo: provinceRepo,
		countryRepo:  countryRepo,
	}
}

func (u *provinceUsecase) List(ctx context.Context, req *dto.ListProvincesRequest) ([]dto.ProvinceResponse, *utils.PaginationResult, error) {
	provinces, total, err := u.provinceRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToProvinceResponses(provinces)

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

func (u *provinceUsecase) GetByID(ctx context.Context, id string) (*dto.ProvinceResponse, error) {
	province, err := u.provinceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProvinceNotFound
		}
		return nil, err
	}

	return mapper.ToProvinceResponse(province), nil
}

func (u *provinceUsecase) Create(ctx context.Context, req *dto.CreateProvinceRequest) (*dto.ProvinceResponse, error) {
	// Validate country exists
	_, err := u.countryRepo.FindByID(ctx, req.CountryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}

	// Check if code already exists
	existing, err := u.provinceRepo.FindByCode(ctx, req.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrProvinceAlreadyExists
	}

	province := mapper.ProvinceFromCreateRequest(req)
	if err := u.provinceRepo.Create(ctx, province); err != nil {
		return nil, err
	}

	// Reload with relations
	province, err = u.provinceRepo.FindByID(ctx, province.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToProvinceResponse(province), nil
}

func (u *provinceUsecase) Update(ctx context.Context, id string, req *dto.UpdateProvinceRequest) (*dto.ProvinceResponse, error) {
	province, err := u.provinceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProvinceNotFound
		}
		return nil, err
	}

	// Validate country if provided
	if req.CountryID != "" && req.CountryID != province.CountryID {
		_, err := u.countryRepo.FindByID(ctx, req.CountryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCountryNotFound
			}
			return nil, err
		}
		province.CountryID = req.CountryID
	}

	// Check if code already exists
	if req.Code != "" && req.Code != province.Code {
		existing, err := u.provinceRepo.FindByCode(ctx, req.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrProvinceAlreadyExists
		}
		province.Code = req.Code
	}

	if req.Name != "" {
		province.Name = req.Name
	}
	if req.IsActive != nil {
		province.IsActive = *req.IsActive
	}

	if err := u.provinceRepo.Update(ctx, province); err != nil {
		return nil, err
	}

	// Reload with relations
	province, err = u.provinceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapper.ToProvinceResponse(province), nil
}

func (u *provinceUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.provinceRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProvinceNotFound
		}
		return err
	}

	hasCities, err := u.provinceRepo.HasCities(ctx, id)
	if err != nil {
		return err
	}
	if hasCities {
		return ErrProvinceHasCities
	}

	return u.provinceRepo.Delete(ctx, id)
}
