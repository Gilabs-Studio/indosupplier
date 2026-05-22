package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/mapper"
	"gorm.io/gorm"
)

var ErrCurrencyNotFound = errors.New("currency not found")

type CurrencyUsecase interface {
	Create(ctx context.Context, req dto.CreateCurrencyRequest) (dto.CurrencyResponse, error)
	GetByID(ctx context.Context, id string) (dto.CurrencyResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.CurrencyResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateCurrencyRequest) (dto.CurrencyResponse, error)
	Delete(ctx context.Context, id string) error
}

type currencyUsecase struct {
	repo repositories.CurrencyRepository
}

func NewCurrencyUsecase(repo repositories.CurrencyRepository) CurrencyUsecase {
	return &currencyUsecase{repo: repo}
}

func (u *currencyUsecase) Create(ctx context.Context, req dto.CreateCurrencyRequest) (dto.CurrencyResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	decimalPlaces := 2
	if req.DecimalPlaces != nil {
		decimalPlaces = *req.DecimalPlaces
	}

	currency := &models.Currency{
		Code:          strings.ToUpper(strings.TrimSpace(req.Code)),
		Name:          strings.TrimSpace(req.Name),
		Symbol:        strings.TrimSpace(req.Symbol),
		DecimalPlaces: decimalPlaces,
		IsActive:      isActive,
	}
	if err := u.repo.Create(ctx, currency); err != nil {
		return dto.CurrencyResponse{}, err
	}
	return mapper.ToCurrencyResponse(currency), nil
}

func (u *currencyUsecase) GetByID(ctx context.Context, id string) (dto.CurrencyResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CurrencyResponse{}, ErrCurrencyNotFound
		}
		return dto.CurrencyResponse{}, err
	}
	return mapper.ToCurrencyResponse(item), nil
}

func (u *currencyUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.CurrencyResponse, int64, error) {
	items, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToCurrencyResponseList(items), total, nil
}

func (u *currencyUsecase) Update(ctx context.Context, id string, req dto.UpdateCurrencyRequest) (dto.CurrencyResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CurrencyResponse{}, ErrCurrencyNotFound
		}
		return dto.CurrencyResponse{}, err
	}

	if req.Code != "" {
		item.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	}
	if req.Name != "" {
		item.Name = strings.TrimSpace(req.Name)
	}
	if req.Symbol != "" {
		item.Symbol = strings.TrimSpace(req.Symbol)
	}
	if req.DecimalPlaces != nil {
		item.DecimalPlaces = *req.DecimalPlaces
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, item); err != nil {
		return dto.CurrencyResponse{}, err
	}
	return mapper.ToCurrencyResponse(item), nil
}

func (u *currencyUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCurrencyNotFound
		}
		return err
	}
	return u.repo.Delete(ctx, id)
}
