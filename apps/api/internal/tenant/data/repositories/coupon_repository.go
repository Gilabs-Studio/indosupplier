package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// CouponRepository defines data-access operations for Coupon.
type CouponRepository interface {
	Create(ctx context.Context, coupon *models.Coupon) error
	FindByCode(ctx context.Context, code string) (*models.Coupon, error)
	FindByID(ctx context.Context, id string) (*models.Coupon, error)
	List(ctx context.Context, activeOnly bool, page, perPage int) ([]*models.Coupon, int64, error)
	SetActive(ctx context.Context, id string, active bool) error
	IncrementUsed(ctx context.Context, id string) error
	// IsEmailUsedCoupon returns true when the given email has already redeemed this coupon.
	IsEmailUsedCoupon(ctx context.Context, couponID, email string) (bool, error)
	// RecordCouponUsage creates a CouponUsage record inside an existing tx.
	RecordCouponUsage(tx interface{ Create(value interface{}) *gorm.DB }, couponID, email string) error
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new CouponRepository.
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *couponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	// Normalise code to uppercase to avoid case-sensitive collisions.
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	return r.getDB(ctx).Create(coupon).Error
}

func (r *couponRepository) FindByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var c models.Coupon
	err := r.getDB(ctx).
		Where("code = ? AND deleted_at IS NULL", strings.ToUpper(strings.TrimSpace(code))).
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *couponRepository) FindByID(ctx context.Context, id string) (*models.Coupon, error) {
	var c models.Coupon
	err := r.getDB(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *couponRepository) List(ctx context.Context, activeOnly bool, page, perPage int) ([]*models.Coupon, int64, error) {
	var coupons []*models.Coupon
	var total int64

	q := r.getDB(ctx).Model(&models.Coupon{}).Where("deleted_at IS NULL")
	if activeOnly {
		q = q.Where("is_active = true")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := q.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

func (r *couponRepository) SetActive(ctx context.Context, id string, active bool) error {
	return r.getDB(ctx).
		Model(&models.Coupon{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("is_active", active).Error
}

func (r *couponRepository) IncrementUsed(ctx context.Context, id string) error {
	return r.getDB(ctx).
		Model(&models.Coupon{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func (r *couponRepository) IsEmailUsedCoupon(ctx context.Context, couponID, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND email = ?", couponID, strings.ToLower(strings.TrimSpace(email))).
		Count(&count).Error
	return count > 0, err
}

func (r *couponRepository) RecordCouponUsage(tx interface{ Create(value interface{}) *gorm.DB }, couponID, email string) error {
	usage := &models.CouponUsage{
		CouponID: couponID,
		Email:    strings.ToLower(strings.TrimSpace(email)),
		UsedAt:   time.Now().UTC(),
	}
	return tx.Create(usage).Error
}
