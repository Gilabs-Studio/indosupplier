package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/gilabs/gims/api/internal/core/apptime"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrInventorySettingsNotFound = errors.New("inventory settings not found")
	ErrInventorySettingsLocked   = errors.New("inventory settings already locked and cannot be changed")
)

type InventorySettingsUsecase interface {
	GetByCompanyID(ctx context.Context, companyID string) (*dto.InventorySettingsResponse, error)
	Upsert(ctx context.Context, req *dto.UpdateInventorySettingsRequest) (*dto.InventorySettingsResponse, error)
	GetAverageCost(ctx context.Context, companyID, productID string) (*dto.InventoryAverageCostResponse, error)
	RecalculateOnReceive(ctx context.Context, companyID, productID string, receivedQty, unitPrice float64) (float64, error)
	CalculateCOGS(ctx context.Context, companyID, productID string, qty float64) (float64, error)
	LockSettings(ctx context.Context, companyID string) error
}

type inventorySettingsUsecase struct {
	repo repositories.InventorySettingsRepository
}

func NewInventorySettingsUsecase(repo repositories.InventorySettingsRepository) InventorySettingsUsecase {
	return &inventorySettingsUsecase{repo: repo}
}

func mapInventorySettings(item *financeModels.InventorySettings) *dto.InventorySettingsResponse {
	return &dto.InventorySettingsResponse{
		ID:              item.ID,
		CompanyID:       item.CompanyID,
		ValuationMethod: item.ValuationMethod,
		IsLocked:        item.IsLocked,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func (uc *inventorySettingsUsecase) GetByCompanyID(ctx context.Context, companyID string) (*dto.InventorySettingsResponse, error) {
	item, err := uc.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultItem := &financeModels.InventorySettings{
				CompanyID:       companyID,
				ValuationMethod: financeModels.InventoryValuationMethodAverageCost,
				IsLocked:        false,
			}
			if createErr := uc.repo.Upsert(ctx, defaultItem); createErr != nil {
				return nil, createErr
			}
			created, refetchErr := uc.repo.FindByCompanyID(ctx, companyID)
			if refetchErr != nil {
				return nil, refetchErr
			}
			return mapInventorySettings(created), nil
		}
		return nil, err
	}
	return mapInventorySettings(item), nil
}

func (uc *inventorySettingsUsecase) Upsert(ctx context.Context, req *dto.UpdateInventorySettingsRequest) (*dto.InventorySettingsResponse, error) {
	current, err := uc.repo.FindByCompanyID(ctx, req.CompanyID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if current != nil && current.IsLocked {
		if current.ValuationMethod != req.ValuationMethod {
			return nil, ErrInventorySettingsLocked
		}
		return mapInventorySettings(current), nil
	}

	item := &financeModels.InventorySettings{
		CompanyID:       req.CompanyID,
		ValuationMethod: req.ValuationMethod,
		IsLocked:        false,
	}
	if current != nil {
		item.ID = current.ID
		item.IsLocked = current.IsLocked
	}

	if err := uc.repo.Upsert(ctx, item); err != nil {
		return nil, err
	}
	updated, err := uc.repo.FindByCompanyID(ctx, req.CompanyID)
	if err != nil {
		return nil, err
	}
	return mapInventorySettings(updated), nil
}

func (uc *inventorySettingsUsecase) GetAverageCost(ctx context.Context, companyID, productID string) (*dto.InventoryAverageCostResponse, error) {
	item, err := uc.repo.GetAverageCostByProduct(ctx, companyID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.InventoryAverageCostResponse{
				CompanyID:     companyID,
				ProductID:     productID,
				AverageCost:   0,
				TotalQuantity: 0,
				TotalValue:    0,
				LastUpdated:   apptime.Now(),
			}, nil
		}
		return nil, err
	}
	return &dto.InventoryAverageCostResponse{
		CompanyID:     item.CompanyID,
		ProductID:     item.ProductID,
		AverageCost:   item.AverageCost,
		TotalQuantity: item.TotalQuantity,
		TotalValue:    item.TotalValue,
		LastUpdated:   item.LastUpdated,
	}, nil
}

func (uc *inventorySettingsUsecase) RecalculateOnReceive(ctx context.Context, companyID, productID string, receivedQty, unitPrice float64) (float64, error) {
	if receivedQty <= 0 {
		return 0, errors.New("received_qty must be greater than zero")
	}
	if unitPrice < 0 {
		return 0, errors.New("unit_price cannot be negative")
	}

	current, err := uc.repo.GetAverageCostByProduct(ctx, companyID, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if current == nil {
		current = &financeModels.InventoryAverageCost{
			CompanyID:     companyID,
			ProductID:     productID,
			AverageCost:   0,
			TotalQuantity: 0,
			TotalValue:    0,
		}
	}

	newTotalQty := current.TotalQuantity + receivedQty
	if newTotalQty <= 0 {
		return 0, errors.New("invalid quantity after recalculation")
	}
	newTotalValue := current.TotalValue + (receivedQty * unitPrice)
	newAvgCost := roundTo(newTotalValue/newTotalQty, 6)

	current.TotalQuantity = roundTo(newTotalQty, 4)
	current.TotalValue = roundTo(newTotalValue, 2)
	current.AverageCost = newAvgCost
	current.LastUpdated = apptime.Now()

	if err := uc.repo.UpsertAverageCost(ctx, current); err != nil {
		return 0, err
	}
	return newAvgCost, uc.LockSettings(ctx, companyID)
}

func (uc *inventorySettingsUsecase) CalculateCOGS(ctx context.Context, companyID, productID string, qty float64) (float64, error) {
	if qty <= 0 {
		return 0, errors.New("qty must be greater than zero")
	}
	current, err := uc.repo.GetAverageCostByProduct(ctx, companyID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return roundTo(qty*current.AverageCost, 2), nil
}

func (uc *inventorySettingsUsecase) LockSettings(ctx context.Context, companyID string) error {
	item, err := uc.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultItem := &financeModels.InventorySettings{
				CompanyID:       companyID,
				ValuationMethod: financeModels.InventoryValuationMethodAverageCost,
				IsLocked:        true,
			}
			return uc.repo.Upsert(ctx, defaultItem)
		}
		return err
	}
	if item.IsLocked {
		return nil
	}
	return uc.repo.SetLocked(ctx, companyID)
}

func roundTo(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}
