package usecase

import (
	"context"
	"errors"
	"fmt"

	xenditClient "github.com/gilabs/gims/api/internal/core/infrastructure/xendit"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"gorm.io/gorm"
)

var (
	ErrPaymentMethodNotFound   = errors.New("payment method not found")
	ErrXenditTokenInvalid      = errors.New("xendit card token is invalid or expired")
	ErrXenditTokenAlreadySaved = errors.New("this card token is already saved")
)

// PaymentMethodUsecase manages saved card tokens for a tenant.
type PaymentMethodUsecase interface {
	List(ctx context.Context, tenantID string) ([]*models.TenantPaymentMethod, error)
	Add(ctx context.Context, tenantID, xenditTokenID string) (*models.TenantPaymentMethod, error)
	SetDefault(ctx context.Context, id, tenantID string) error
	Remove(ctx context.Context, id, tenantID string) error
}

type paymentMethodUsecase struct {
	repo   repositories.PaymentMethodRepository
	xendit *xenditClient.Client
}

// NewPaymentMethodUsecase creates a PaymentMethodUsecase.
func NewPaymentMethodUsecase(repo repositories.PaymentMethodRepository, xendit *xenditClient.Client) PaymentMethodUsecase {
	return &paymentMethodUsecase{repo: repo, xendit: xendit}
}

func (u *paymentMethodUsecase) List(ctx context.Context, tenantID string) ([]*models.TenantPaymentMethod, error) {
	return u.repo.ListByTenant(ctx, tenantID)
}

// Add verifies the Xendit token, then persists the masked card reference.
func (u *paymentMethodUsecase) Add(ctx context.Context, tenantID, xenditTokenID string) (*models.TenantPaymentMethod, error) {
	if u.xendit == nil || !u.xendit.IsConfigured() {
		return nil, fmt.Errorf("payment gateway not configured")
	}

	tokenData, err := u.xendit.GetCardToken(ctx, xenditTokenID)
	if err != nil || tokenData.Status != "VALID" {
		return nil, ErrXenditTokenInvalid
	}

	existing, err := u.repo.FindByID(ctx, xenditTokenID, tenantID)
	if err == nil && existing != nil {
		return nil, ErrXenditTokenAlreadySaved
	}

	m := &models.TenantPaymentMethod{
		TenantID:         tenantID,
		XenditTokenID:    tokenData.ID,
		MaskedCardNumber: tokenData.MaskedCardNumber,
		CardBrand:        tokenData.CardBrand,
		CardHolderName:   tokenData.CardHolderName,
		ExpiryMonth:      tokenData.ExpiryMonth,
		ExpiryYear:       tokenData.ExpiryYear,
	}

	existing2, listErr := u.repo.ListByTenant(ctx, tenantID)
	if listErr == nil && len(existing2) == 0 {
		m.IsDefault = true
	}

	if createErr := u.repo.Create(ctx, m); createErr != nil {
		return nil, fmt.Errorf("save payment method: %w", createErr)
	}
	return m, nil
}

func (u *paymentMethodUsecase) SetDefault(ctx context.Context, id, tenantID string) error {
	if err := u.repo.SetDefault(ctx, id, tenantID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentMethodNotFound
		}
		return err
	}
	return nil
}

func (u *paymentMethodUsecase) Remove(ctx context.Context, id, tenantID string) error {
	if err := u.repo.Delete(ctx, id, tenantID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentMethodNotFound
		}
		return err
	}
	return nil
}
