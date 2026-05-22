package usecase

import (
	"context"
	"errors"
	"strings"

	posModels "github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrPOSConfigForbidden = errors.New("pos config forbidden")
)

// POSConfigUsecase manages POS configuration per outlet.
type POSConfigUsecase interface {
	// GetOrCreate returns the existing config or creates a default one if absent.
	GetOrCreate(ctx context.Context, outletID string) (*dto.POSConfigResponse, error)
	// Upsert creates or updates the POS config for the given outlet.
	Upsert(ctx context.Context, outletID string, req *dto.UpsertPOSConfigRequest, userID string) (*dto.POSConfigResponse, error)
	// UpdateReceiptWhatsAppTemplate updates only the receipt WhatsApp template for the given outlet.
	UpdateReceiptWhatsAppTemplate(ctx context.Context, outletID string, req *dto.UpdateReceiptWhatsAppTemplateRequest, userID string, isOwner bool) (*dto.POSConfigResponse, error)
}

type posConfigUsecase struct {
	repo repositories.POSConfigRepository
}

// NewPOSConfigUsecase constructs a POSConfigUsecase.
func NewPOSConfigUsecase(repo repositories.POSConfigRepository) POSConfigUsecase {
	return &posConfigUsecase{repo: repo}
}

func (u *posConfigUsecase) GetOrCreate(ctx context.Context, outletID string) (*dto.POSConfigResponse, error) {
	cfg, err := u.repo.FindByOutletID(ctx, outletID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if cfg == nil {
		cfg = &posModels.POSConfig{
			OutletID:           outletID,
			TaxRate:            0.11,
			ServiceChargeRate:  0,
			AllowDiscount:      true,
			MaxDiscountPercent: 100,
			PrintReceiptAuto:   false,
			Currency:           "IDR",
		}
		if upsertErr := u.repo.Upsert(ctx, cfg); upsertErr != nil {
			return nil, upsertErr
		}

		cfg, err = u.repo.FindByOutletID(ctx, outletID)
		if err != nil {
			return nil, err
		}
	}

	return mapper.ToPOSConfigResponse(cfg), nil
}

func (u *posConfigUsecase) Upsert(ctx context.Context, outletID string, req *dto.UpsertPOSConfigRequest, userID string) (*dto.POSConfigResponse, error) {
	existing, _ := u.repo.FindByOutletID(ctx, outletID)

	cfg := &posModels.POSConfig{OutletID: outletID}
	if existing != nil {
		cfg.TaxRate = existing.TaxRate
		cfg.ServiceChargeRate = existing.ServiceChargeRate
		cfg.AllowDiscount = existing.AllowDiscount
		cfg.MaxDiscountPercent = existing.MaxDiscountPercent
		cfg.PrintReceiptAuto = existing.PrintReceiptAuto
		cfg.ReceiptFooter = existing.ReceiptFooter
		cfg.ReceiptWhatsAppTemplate = existing.ReceiptWhatsAppTemplate
		cfg.Currency = existing.Currency
	} else {
		cfg.Currency = "IDR"
		cfg.AllowDiscount = true
	}

	if req.TaxRate != nil {
		cfg.TaxRate = *req.TaxRate
	}
	if req.ServiceChargeRate != nil {
		cfg.ServiceChargeRate = *req.ServiceChargeRate
	}
	if req.AllowDiscount != nil {
		cfg.AllowDiscount = *req.AllowDiscount
	}
	if req.MaxDiscountPercent != nil {
		cfg.MaxDiscountPercent = *req.MaxDiscountPercent
	}
	if req.PrintReceiptAuto != nil {
		cfg.PrintReceiptAuto = *req.PrintReceiptAuto
	}
	if req.ReceiptFooter != nil {
		cfg.ReceiptFooter = req.ReceiptFooter
	}
	if req.ReceiptWhatsAppTemplate != nil {
		template := strings.TrimSpace(*req.ReceiptWhatsAppTemplate)
		if template == "" {
			cfg.ReceiptWhatsAppTemplate = nil
		} else {
			cfg.ReceiptWhatsAppTemplate = &template
		}
	}
	if req.Currency != nil {
		cfg.Currency = *req.Currency
	}

	cfg.UpdatedBy = &userID
	if err := u.repo.Upsert(ctx, cfg); err != nil {
		return nil, err
	}

	saved, err := u.repo.FindByOutletID(ctx, outletID)
	if err != nil {
		return nil, err
	}

	return mapper.ToPOSConfigResponse(saved), nil
}

func (u *posConfigUsecase) UpdateReceiptWhatsAppTemplate(ctx context.Context, outletID string, req *dto.UpdateReceiptWhatsAppTemplateRequest, userID string, isOwner bool) (*dto.POSConfigResponse, error) {
	if !isOwner {
		return nil, ErrPOSConfigForbidden
	}

	cfg, err := u.repo.FindByOutletID(ctx, outletID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		cfg = &posModels.POSConfig{
			OutletID:           outletID,
			TaxRate:            0.11,
			ServiceChargeRate:  0,
			AllowDiscount:      true,
			MaxDiscountPercent: 100,
			PrintReceiptAuto:   false,
			Currency:           "IDR",
		}
	}

	if req != nil && req.ReceiptWhatsAppTemplate != nil {
		template := strings.TrimSpace(*req.ReceiptWhatsAppTemplate)
		if template == "" {
			cfg.ReceiptWhatsAppTemplate = nil
		} else {
			cfg.ReceiptWhatsAppTemplate = &template
		}
	} else {
		cfg.ReceiptWhatsAppTemplate = nil
	}

	cfg.UpdatedBy = &userID
	if err := u.repo.Upsert(ctx, cfg); err != nil {
		return nil, err
	}

	saved, err := u.repo.FindByOutletID(ctx, outletID)
	if err != nil {
		return nil, err
	}

	return mapper.ToPOSConfigResponse(saved), nil
}
