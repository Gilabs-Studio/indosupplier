package migrations

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// NormalizeLifetimeSubscriptions sets tenant_subscriptions.expires_at = NULL
// for subscriptions that were granted by coupons with DurationDays == 0
// (treated as lifetime). This fixes cases where a lifetime coupon previously
// wrote a near-term expires_at and caused the tenant to appear overdue.
func NormalizeLifetimeSubscriptions(db *gorm.DB) error {
    log.Println("Running migration: NormalizeLifetimeSubscriptions")

    // Only operate when coupon exists with duration_days = 0
    // Update subscriptions that reference such coupons and currently have expires_at set.
    // Also clear next_billing_at since lifetime shouldn't have recurring billing.
    res := db.Exec(`
        UPDATE tenant_subscriptions ts
        SET expires_at = NULL,
            next_billing_at = NULL,
            status = 'active',
            updated_at = ?
        FROM coupons c
        WHERE ts.coupon_id = c.id
          AND c.duration_days = 0
          AND ts.expires_at IS NOT NULL
    `, time.Now())

    if res.Error != nil {
        log.Printf("Warning: NormalizeLifetimeSubscriptions failed: %v", res.Error)
        return res.Error
    }

    log.Printf("NormalizeLifetimeSubscriptions: rows affected=%d", res.RowsAffected)
    return nil
}
