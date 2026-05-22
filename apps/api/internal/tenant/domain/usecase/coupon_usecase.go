package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrCouponNotFound    = errors.New("coupon not found")
	ErrCouponInactive    = errors.New("coupon is inactive")
	ErrCouponExpired     = errors.New("coupon has expired")
	ErrCouponExhausted   = errors.New("coupon usage limit reached")
	ErrCouponDuplicate   = errors.New("coupon code already exists")
	ErrCouponAlreadyUsed = errors.New("coupon already used by this email")
	ErrCouponUserLimit   = errors.New("coupon user limit exceeded")
)

// CouponUsecase handles coupon management for system admins and validation during registration.
type CouponUsecase interface {
	Create(ctx context.Context, req *dto.CreateCouponRequest, adminEmail string) (*dto.CouponResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCouponRequest) (*dto.CouponResponse, error)
	List(ctx context.Context, params dto.CouponListParams) ([]*dto.CouponResponse, *response.PaginationMeta, error)
	SetActive(ctx context.Context, id string, active bool) error
	// Validate checks whether a coupon code is usable (public, no email context).
	Validate(ctx context.Context, code string) (*dto.ValidateCouponResponse, error)
	// ValidateForEmail additionally checks if the given email has already used this coupon.
	ValidateForEmail(ctx context.Context, code, email string) (*dto.ValidateCouponResponse, error)
	// ValidateForPlan checks coupon validity for a specific plan slug (enforces tier_specific scope).
	ValidateForPlan(ctx context.Context, code, email, planSlug string) (*dto.ValidateCouponResponse, error)
	// RedeemForTenant validates the coupon, records the email usage, and creates a TenantSubscription.
	RedeemForTenant(ctx context.Context, code, email, tenantID string, userCount int) (*models.TenantSubscription, error)
	// ApplyDiscount computes the discounted invoice amount for a paid-plan coupon.
	// Returns original amount when the coupon is a trial type or not applicable.
	ApplyDiscount(ctx context.Context, code, planSlug string, baseAmount int64, userCount int, billingPeriod string) (int64, error)
}

type couponUsecase struct {
	couponRepo repositories.CouponRepository
	subRepo    repositories.SubscriptionRepository
	db         *gorm.DB
}

// NewCouponUsecase creates a new CouponUsecase.
func NewCouponUsecase(
	couponRepo repositories.CouponRepository,
	subRepo repositories.SubscriptionRepository,
	db *gorm.DB,
) CouponUsecase {
	return &couponUsecase{
		couponRepo: couponRepo,
		subRepo:    subRepo,
		db:         db,
	}
}

func (u *couponUsecase) Create(ctx context.Context, req *dto.CreateCouponRequest, adminEmail string) (*dto.CouponResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		code = generateCouponCode()
	}

	coupon := &models.Coupon{
		Code:                   code,
		Description:            req.Description,
		CouponType:             models.CouponType(req.CouponType),
		DiscountType:           models.CouponDiscountType(req.DiscountType),
		DiscountValue:          req.DiscountValue,
		MaxUserCount:           req.MaxUserCount,
		LockUserCount:          req.LockUserCount,
		PackagePriceMonthlyIDR: req.PackagePriceMonthlyIDR,
		PackagePriceYearlyIDR:  req.PackagePriceYearlyIDR,
		Scope:                  models.CouponScope(req.Scope),
		DurationDays:           req.DurationDays,
		MaxUses:                req.MaxUses,
		MaxUsesPerEmail:        req.MaxUsesPerEmail,
		IsActive:               true,
		ExpiresAt:              req.ExpiresAt,
		CreatedBy:              adminEmail,
	}
	if req.TargetPlanSlug != "" {
		slug := req.TargetPlanSlug
		coupon.TargetPlanSlug = &slug
	}
	if coupon.MaxUsesPerEmail < 1 {
		coupon.MaxUsesPerEmail = 1
	}
	// Apply defaults for optional fields not provided by the caller.
	if coupon.DiscountType == "" {
		coupon.DiscountType = models.CouponDiscountTrial
	}
	if coupon.Scope == "" {
		coupon.Scope = models.CouponScopeGeneral
	}
	if coupon.PackagePriceYearlyIDR <= 0 && coupon.PackagePriceMonthlyIDR > 0 {
		coupon.PackagePriceYearlyIDR = coupon.PackagePriceMonthlyIDR * 12 * 0.9
	}

	if err := u.couponRepo.Create(ctx, coupon); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, ErrCouponDuplicate
		}
		return nil, err
	}

	return toCouponDTO(coupon), nil
}

func (u *couponUsecase) Update(ctx context.Context, id string, req *dto.UpdateCouponRequest) (*dto.CouponResponse, error) {
	coupon, err := u.couponRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}

	coupon.Description = req.Description
	coupon.CouponType = models.CouponType(req.CouponType)
	coupon.DiscountType = models.CouponDiscountType(req.DiscountType)
	coupon.DiscountValue = req.DiscountValue
	coupon.MaxUserCount = req.MaxUserCount
	coupon.LockUserCount = req.LockUserCount
	coupon.PackagePriceMonthlyIDR = req.PackagePriceMonthlyIDR
	coupon.PackagePriceYearlyIDR = req.PackagePriceYearlyIDR
	coupon.Scope = models.CouponScope(req.Scope)
	coupon.DurationDays = req.DurationDays
	coupon.MaxUses = req.MaxUses
	coupon.MaxUsesPerEmail = req.MaxUsesPerEmail
	coupon.ExpiresAt = req.ExpiresAt

	if req.TargetPlanSlug != "" {
		slug := req.TargetPlanSlug
		coupon.TargetPlanSlug = &slug
	} else {
		coupon.TargetPlanSlug = nil
	}

	if coupon.DiscountType == "" {
		coupon.DiscountType = models.CouponDiscountTrial
	}
	if coupon.Scope == "" {
		coupon.Scope = models.CouponScopeGeneral
	}
	if coupon.MaxUsesPerEmail < 1 {
		coupon.MaxUsesPerEmail = 1
	}
	if coupon.PackagePriceYearlyIDR <= 0 && coupon.PackagePriceMonthlyIDR > 0 {
		coupon.PackagePriceYearlyIDR = coupon.PackagePriceMonthlyIDR * 12 * 0.9
	}

	if err := u.db.WithContext(ctx).Save(coupon).Error; err != nil {
		return nil, err
	}

	return toCouponDTO(coupon), nil
}

func (u *couponUsecase) List(ctx context.Context, params dto.CouponListParams) ([]*dto.CouponResponse, *response.PaginationMeta, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}

	coupons, total, err := u.couponRepo.List(ctx, params.ActiveOnly, params.Page, params.PerPage)
	if err != nil {
		return nil, nil, err
	}

	result := make([]*dto.CouponResponse, 0, len(coupons))
	for _, c := range coupons {
		result = append(result, toCouponDTO(c))
	}

	pagination := response.NewPaginationMeta(params.Page, params.PerPage, int(total))
	return result, pagination, nil
}

func (u *couponUsecase) SetActive(ctx context.Context, id string, active bool) error {
	_, err := u.couponRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	return u.couponRepo.SetActive(ctx, id, active)
}

func (u *couponUsecase) Validate(ctx context.Context, code string) (*dto.ValidateCouponResponse, error) {
	coupon, err := u.couponRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.ValidateCouponResponse{Valid: false}, nil
		}
		return nil, err
	}

	if !coupon.IsUsable() {
		return &dto.ValidateCouponResponse{Valid: false}, nil
	}

	return &dto.ValidateCouponResponse{
		Valid:                  true,
		Description:            coupon.Description,
		CouponType:             string(coupon.CouponType),
		DiscountType:           string(coupon.DiscountType),
		DiscountValue:          coupon.DiscountValue,
		MaxUserCount:           coupon.MaxUserCount,
		LockUserCount:          coupon.LockUserCount,
		PackagePriceMonthlyIDR: coupon.PackagePriceMonthlyIDR,
		PackagePriceYearlyIDR:  coupon.PackagePriceYearlyIDR,
		Scope:                  string(coupon.Scope),
		DurationDays:           coupon.DurationDays,
	}, nil
}

func (u *couponUsecase) ValidateForEmail(ctx context.Context, code, email string) (*dto.ValidateCouponResponse, error) {
	coupon, err := u.couponRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.ValidateCouponResponse{Valid: false, Reason: "not_found"}, nil
		}
		return nil, err
	}

	if !coupon.IsUsable() {
		reason := "inactive"
		if coupon.ExpiresAt != nil && apptime.Now().After(*coupon.ExpiresAt) {
			reason = "expired"
		} else if coupon.UsedCount >= coupon.MaxUses {
			reason = "exhausted"
		}
		return &dto.ValidateCouponResponse{Valid: false, Reason: reason}, nil
	}

	if email != "" {
		used, err := u.couponRepo.IsEmailUsedCoupon(ctx, coupon.ID, email)
		if err != nil {
			return nil, err
		}
		if used {
			return &dto.ValidateCouponResponse{Valid: false, Reason: "already_used_by_email"}, nil
		}
	}

	return &dto.ValidateCouponResponse{
		Valid:                  true,
		Description:            coupon.Description,
		CouponType:             string(coupon.CouponType),
		DiscountType:           string(coupon.DiscountType),
		DiscountValue:          coupon.DiscountValue,
		MaxUserCount:           coupon.MaxUserCount,
		LockUserCount:          coupon.LockUserCount,
		PackagePriceMonthlyIDR: coupon.PackagePriceMonthlyIDR,
		PackagePriceYearlyIDR:  coupon.PackagePriceYearlyIDR,
		Scope:                  string(coupon.Scope),
		TargetPlanSlug:         targetPlanSlug(coupon),
		DurationDays:           coupon.DurationDays,
	}, nil
}

// RedeemForTenant validates the coupon under a DB lock and creates a TenantSubscription.

func (u *couponUsecase) ValidateForPlan(ctx context.Context, code, email, planSlug string) (*dto.ValidateCouponResponse, error) {
	resp, err := u.ValidateForEmail(ctx, code, email)
	if err != nil {
		return nil, err
	}
	if !resp.Valid {
		return resp, nil
	}
	// Tier-specific coupons must match the requested plan.
	if resp.Scope == string(models.CouponScopeTierSpecific) && resp.TargetPlanSlug != "" {
		if planSlug != resp.TargetPlanSlug {
			return &dto.ValidateCouponResponse{Valid: false, Reason: "plan_mismatch"}, nil
		}
	}
	return resp, nil
}

// ApplyDiscount computes the discounted invoice amount for a percent/amount coupon.
// Trial-type coupons set the amount to zero (they bypass the payment flow entirely).
// For amount coupons, DiscountValue is interpreted as a monthly amount in IDR.
func (u *couponUsecase) ApplyDiscount(ctx context.Context, code, planSlug string, baseAmount int64, userCount int, billingPeriod string) (int64, error) {
	if code == "" {
		return baseAmount, nil
	}
	if userCount < 1 {
		userCount = 1
	}
	coupon, err := u.couponRepo.FindByCode(ctx, code)
	if err != nil {
		return baseAmount, nil // Best-effort: return original if coupon not found
	}
	if !coupon.IsUsable() {
		return baseAmount, nil
	}
	if coupon.Scope == models.CouponScopeTierSpecific && coupon.TargetPlanSlug != nil {
		if planSlug != *coupon.TargetPlanSlug {
			return baseAmount, nil
		}
	}

	if coupon.MaxUserCount > 0 && userCount > coupon.MaxUserCount {
		return baseAmount, ErrCouponUserLimit
	}

	if packageAmount, ok := packageCouponAmount(coupon, userCount, billingPeriod); ok {
		return packageAmount, nil
	}

	return applyStandardCouponDiscount(coupon, baseAmount, billingPeriod), nil
}

func packageCouponAmount(coupon *models.Coupon, userCount int, billingPeriod string) (int64, bool) {
	if coupon == nil || coupon.PackagePriceMonthlyIDR <= 0 || coupon.MaxUserCount <= 0 {
		return 0, false
	}

	packageAmount := int64(coupon.PackagePriceMonthlyIDR)
	if billingPeriod == "yearly" {
		if coupon.PackagePriceYearlyIDR > 0 {
			packageAmount = int64(coupon.PackagePriceYearlyIDR)
		} else {
			packageAmount = int64(coupon.PackagePriceMonthlyIDR * 12 * 0.9)
		}
	}

	if userCount <= coupon.MaxUserCount {
		return packageAmount, true
	}
	return 0, false
}

func applyStandardCouponDiscount(coupon *models.Coupon, baseAmount int64, billingPeriod string) int64 {
	if coupon == nil {
		return baseAmount
	}

	switch coupon.DiscountType {
	case models.CouponDiscountTrial:
		return 0
	case models.CouponDiscountPercent:
		if coupon.DiscountValue <= 0 {
			return baseAmount
		}
		discount := int64(float64(baseAmount) * coupon.DiscountValue / 100)
		if final := baseAmount - discount; final >= 0 {
			return final
		}
		return 0
	case models.CouponDiscountAmount:
		discount := int64(coupon.DiscountValue)
		if billingPeriod == "yearly" {
			discount *= 12
		}
		if final := baseAmount - discount; final >= 0 {
			return final
		}
		return 0
	default:
		return baseAmount
	}
}

func (u *couponUsecase) RedeemForTenant(ctx context.Context, code, email, tenantID string, userCount int) (*models.TenantSubscription, error) {
	var sub *models.TenantSubscription

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coupon, err := loadCouponForRedemption(tx, code)
		if err != nil {
			return err
		}
		if err := validateCouponForRedemption(tx, coupon, email); err != nil {
			return err
		}
		if err := recordCouponUsage(tx, coupon.ID, email); err != nil {
			return err
		}
		if err := incrementCouponUsedCount(tx, coupon.ID); err != nil {
			return err
		}

		if userCount < 1 {
			userCount = 1
		}
		if coupon.MaxUserCount > 0 {
			if coupon.LockUserCount {
				userCount = coupon.MaxUserCount
			} else if userCount > coupon.MaxUserCount {
				return ErrCouponUserLimit
			}
		}

		sub = buildTrialSubscriptionFromCoupon(coupon, tenantID, userCount)
		return tx.Create(sub).Error
	})

	return sub, err
}

func loadCouponForRedemption(tx *gorm.DB, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("code = ? AND deleted_at IS NULL", strings.ToUpper(strings.TrimSpace(code))).
		First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return &coupon, nil
}

func validateCouponForRedemption(tx *gorm.DB, coupon *models.Coupon, email string) error {
	if coupon == nil {
		return ErrCouponNotFound
	}
	if !coupon.IsActive {
		return ErrCouponInactive
	}
	if coupon.UsedCount >= coupon.MaxUses {
		return ErrCouponExhausted
	}
	if coupon.ExpiresAt != nil && apptime.Now().After(*coupon.ExpiresAt) {
		return ErrCouponExpired
	}

	var usageCount int64
	if err := tx.Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND email = ?", coupon.ID, strings.ToLower(strings.TrimSpace(email))).
		Count(&usageCount).Error; err != nil {
		return err
	}
	if usageCount > 0 {
		return ErrCouponAlreadyUsed
	}
	return nil
}

func recordCouponUsage(tx *gorm.DB, couponID, email string) error {
	usage := &models.CouponUsage{
		CouponID: couponID,
		Email:    strings.ToLower(strings.TrimSpace(email)),
		UsedAt:   apptime.Now(),
	}
	return tx.Create(usage).Error
}

func incrementCouponUsedCount(tx *gorm.DB, couponID string) error {
	return tx.Model(&models.Coupon{}).Where("id = ?", couponID).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func buildTrialSubscriptionFromCoupon(coupon *models.Coupon, tenantID string, userCount int) *models.TenantSubscription {
	if coupon == nil {
		return nil
	}

	now := apptime.Now()
	var expiresAt time.Time
	var expiresAtPtr *time.Time
	if coupon.DurationDays > 0 {
		expiresAt = now.Add(time.Duration(coupon.DurationDays) * 24 * time.Hour)
		expiresAtPtr = &expiresAt
	} else {
		// DurationDays == 0 indicates lifetime — leave ExpiresAt nil
		expiresAtPtr = nil
	}
	couponID := coupon.ID
	plan := couponTypeToPlan(coupon.CouponType)
	if coupon.TargetPlanSlug != nil && *coupon.TargetPlanSlug != "" {
		plan = models.SubscriptionPlan(*coupon.TargetPlanSlug)
	}
	if userCount < 1 {
		userCount = 1
	}

	return &models.TenantSubscription{
		TenantID:      tenantID,
		Plan:          plan,
		BillingPeriod: models.BillingMonthly,
		Status:        models.SubscriptionTrial,
		UserCount:     userCount,
		SeatLimit:     userCount,
		StartsAt:      now,
		ExpiresAt:     expiresAtPtr,
		CouponID:      &couponID,
		Notes:         fmt.Sprintf("Granted via coupon %s (%s)", coupon.Code, coupon.Description),
	}
}

// ─── helpers ────────────────────────────────────────────────────────────────

// toCouponDTO converts a model to the safe response DTO.
func toCouponDTO(c *models.Coupon) *dto.CouponResponse {
	r := &dto.CouponResponse{
		ID:                     c.ID,
		Code:                   c.Code,
		Description:            c.Description,
		CouponType:             string(c.CouponType),
		DiscountType:           string(c.DiscountType),
		DiscountValue:          c.DiscountValue,
		MaxUserCount:           c.MaxUserCount,
		LockUserCount:          c.LockUserCount,
		PackagePriceMonthlyIDR: c.PackagePriceMonthlyIDR,
		PackagePriceYearlyIDR:  c.PackagePriceYearlyIDR,
		Scope:                  string(c.Scope),
		DurationDays:           c.DurationDays,
		MaxUses:                c.MaxUses,
		UsedCount:              c.UsedCount,
		MaxUsesPerEmail:        c.MaxUsesPerEmail,
		IsActive:               c.IsActive,
		ExpiresAt:              c.ExpiresAt,
		CreatedBy:              c.CreatedBy,
		CreatedAt:              c.CreatedAt,
		UpdatedAt:              c.UpdatedAt,
	}
	if c.TargetPlanSlug != nil {
		r.TargetPlanSlug = *c.TargetPlanSlug
	}
	return r
}

// targetPlanSlug returns the target plan slug string from a coupon (empty when nil).
func targetPlanSlug(c *models.Coupon) string {
	if c.TargetPlanSlug != nil {
		return *c.TargetPlanSlug
	}
	return ""
}

// couponTypeToPlan maps a CouponType to the closest SubscriptionPlan.
func couponTypeToPlan(ct models.CouponType) models.SubscriptionPlan {
	switch ct {
	case models.CouponTypePOSOnly:
		return models.PlanPOSGrowth
	case models.CouponTypeERPOnly:
		return models.PlanERPCore
	case models.CouponTypeCRMOnly:
		return models.PlanCRMGrowth
	case models.CouponTypeHROnly:
		return models.PlanHRGrowth
	default:
		return models.PlanUltimateSuite
	}
}

// generateCouponCode produces a random 10-character uppercase alphanumeric code.
func generateCouponCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Omit I, O, 0, 1 for readability
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use time-based suffix — should not happen in practice.
		return fmt.Sprintf("GIMS%d", time.Now().UnixNano()%100000)
	}
	result := make([]byte, 10)
	for i := range result {
		result[i] = charset[int(b[i])%len(charset)]
	}
	return string(result)
}
