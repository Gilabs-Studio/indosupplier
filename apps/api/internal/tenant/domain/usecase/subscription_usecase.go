package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"gorm.io/gorm"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

// SubscriptionUsecase handles tenant subscription queries for system admins.
type SubscriptionUsecase interface {
	ListAll(ctx context.Context, params dto.SubscriptionListParams) ([]*dto.SubscriptionResponse, *response.PaginationMeta, error)
	GetActiveByTenant(ctx context.Context, tenantID string) (*dto.SubscriptionResponse, error)
}

type subscriptionUsecase struct {
	subRepo    repositories.SubscriptionRepository
	tenantRepo repositories.TenantRepository
}

// NewSubscriptionUsecase creates a new SubscriptionUsecase.
func NewSubscriptionUsecase(
	subRepo repositories.SubscriptionRepository,
	tenantRepo repositories.TenantRepository,
) SubscriptionUsecase {
	return &subscriptionUsecase{
		subRepo:    subRepo,
		tenantRepo: tenantRepo,
	}
}

func (u *subscriptionUsecase) ListAll(ctx context.Context, params dto.SubscriptionListParams) ([]*dto.SubscriptionResponse, *response.PaginationMeta, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}

	subs, total, err := u.subRepo.ListAll(ctx, params.Page, params.PerPage)
	if err != nil {
		return nil, nil, err
	}

	result := make([]*dto.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		r := subModelToDTO(s)
		r.TenantName = u.resolveTenantName(ctx, s.TenantID)
		result = append(result, r)
	}

	pagination := response.NewPaginationMeta(params.Page, params.PerPage, int(total))
	return result, pagination, nil
}

func (u *subscriptionUsecase) GetActiveByTenant(ctx context.Context, tenantID string) (*dto.SubscriptionResponse, error) {
	sub, err := u.subRepo.FindActiveByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}
	r := subModelToDTO(sub)
	if tenant, tenantErr := u.tenantRepo.FindByIDWithCounts(ctx, tenantID); tenantErr == nil && tenant != nil {
		// Legacy tenants can still have seat_limit/user_count defaulted to 1 in tenant_subscriptions,
		// so the repository normalizes MaxUsers from the active subscription seat_limit first and
		// only falls back to tenants.max_users when subscription data is unavailable.
		if tenant.MaxUsers > r.SeatLimit {
			r.SeatLimit = tenant.MaxUsers
		}
		if tenant.MaxUsers > r.UserCount {
			r.UserCount = tenant.MaxUsers
		}
		if tenant.OwnerEmail != "" {
			r.TenantName = tenant.OwnerEmail
		} else if tenant.Name != "" {
			r.TenantName = tenant.Name
		}
	}
	if activeUsers, countErr := u.tenantRepo.CountActiveUsers(ctx, tenantID); countErr == nil {
		r.ActiveUsers = int(activeUsers)
	}
	if r.TenantName == "" {
		r.TenantName = u.resolveTenantName(ctx, tenantID)
	}
	return r, nil
}

// resolveTenantName fetches a tenant name by ID (best-effort, returns empty on failure).
func (u *subscriptionUsecase) resolveTenantName(ctx context.Context, tenantID string) string {
	if tenantID == "" {
		return ""
	}
	tenant, err := u.tenantRepo.FindByIDWithCounts(ctx, tenantID)
	if err == nil && tenant != nil {
		if tenant.OwnerEmail != "" {
			return tenant.OwnerEmail
		}
		if tenant.Name != "" {
			return tenant.Name
		}
	}
	return ""
}

// subModelToDTO converts a TenantSubscription model to its DTO representation.
func subModelToDTO(s *models.TenantSubscription) *dto.SubscriptionResponse {
	r := &dto.SubscriptionResponse{
		ID:            s.ID,
		TenantID:      s.TenantID,
		Plan:          string(s.Plan),
		BillingPeriod: string(s.BillingPeriod),
		Status:        string(s.Status),
		StartsAt:      s.StartsAt,
		ExpiresAt:     s.ExpiresAt,
		NextBillingAt: s.NextBillingAt,
		UserCount:     s.UserCount,
		SeatLimit:     s.SeatLimit,
		OutletLimit:   s.OutletLimit,
		AmountPaidIDR: s.AmountPaidIDR,
		Notes:         s.Notes,
		CreatedAt:     s.CreatedAt,
	}
	if r.SeatLimit < s.UserCount {
		r.SeatLimit = s.UserCount
	}
	if r.SeatLimit == 0 {
		r.SeatLimit = s.UserCount
	}
	if r.UserCount == 0 {
		r.UserCount = r.SeatLimit
	}
	if r.OutletLimit <= 0 {
		r.OutletLimit = 1
	}
	if s.XenditPaymentID != nil {
		r.XenditPaymentID = *s.XenditPaymentID
	}
	if s.Coupon != nil {
		r.CouponCode = s.Coupon.Code
	}
	return r
}
