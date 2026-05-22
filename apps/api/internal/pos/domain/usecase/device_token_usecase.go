package usecase

import (
	"context"
	"errors"

	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

var ErrPOSDeviceTokenForbidden = errors.New("device token outlet is outside user scope")

type POSDeviceTokenUsecase interface {
	Register(ctx context.Context, userID string, req *dto.RegisterPOSDeviceTokenRequest) (*dto.POSDeviceTokenResponse, error)
}

type posDeviceTokenUsecase struct {
	repo       repositories.POSDeviceTokenRepository
	outletRepo orgRepo.OutletRepository
}

func NewPOSDeviceTokenUsecase(repo repositories.POSDeviceTokenRepository, outletRepo orgRepo.OutletRepository) POSDeviceTokenUsecase {
	return &posDeviceTokenUsecase{repo: repo, outletRepo: outletRepo}
}

func (u *posDeviceTokenUsecase) Register(ctx context.Context, userID string, req *dto.RegisterPOSDeviceTokenRequest) (*dto.POSDeviceTokenResponse, error) {
	if req.TenantID != scopeString(ctx, "tenant_id") {
		return nil, ErrPOSDeviceTokenForbidden
	}
	outlet, err := u.outletRepo.GetByID(ctx, req.OutletID)
	if err != nil {
		return nil, err
	}
	if outlet == nil || outlet.TenantID != req.TenantID {
		return nil, ErrPOSDeviceTokenForbidden
	}
	allowedOutletIDs, err := resolveScopedPOSOutletIDs(ctx, u.outletRepo)
	if err != nil {
		return nil, err
	}
	if allowedOutletIDs != nil && !isOutletAllowed(allowedOutletIDs, req.OutletID) {
		return nil, ErrPOSDeviceTokenForbidden
	}

	token := &models.POSDeviceToken{
		UserID:   userID,
		TenantID: req.TenantID,
		OutletID: req.OutletID,
		Platform: req.Platform,
		Token:    req.Token,
	}
	if err := u.repo.Upsert(ctx, token); err != nil {
		return nil, err
	}
	return &dto.POSDeviceTokenResponse{
		ID:       token.ID,
		OutletID: token.OutletID,
		TenantID: token.TenantID,
		Platform: token.Platform,
	}, nil
}
