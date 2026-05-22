package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlan represents the tier identifier, aligned with the pricing document.
type SubscriptionPlan string

const (
	// POS tiers
	PlanPOSEssential  SubscriptionPlan = "pos_essential"
	PlanPOSGrowth     SubscriptionPlan = "pos_growth"
	PlanPOSEnterprise SubscriptionPlan = "pos_enterprise"

	// ERP tiers
	PlanERPCore       SubscriptionPlan = "erp_core"
	PlanERPPro        SubscriptionPlan = "erp_pro"
	PlanERPEnterprise SubscriptionPlan = "erp_enterprise"

	// CRM tiers
	PlanCRMBasic      SubscriptionPlan = "crm_basic"
	PlanCRMGrowth     SubscriptionPlan = "crm_growth"
	PlanCRMEnterprise SubscriptionPlan = "crm_enterprise"

	// HR tiers
	PlanHRBasic      SubscriptionPlan = "hr_basic"
	PlanHRGrowth     SubscriptionPlan = "hr_growth"
	PlanHREnterprise SubscriptionPlan = "hr_enterprise"

	// Bundle plans (per-user pricing, Decoy Effect strategy)
	PlanBundleStarter  SubscriptionPlan = "bundle_starter"  // POS + CRM
	PlanBundleUltimate SubscriptionPlan = "bundle_ultimate" // All modules at a discount

	// Bundle plans (per-user pricing, sold via Xendit)
	PlanGrowthSuite   SubscriptionPlan = "growth_suite"   // POS + ERP + CRM
	PlanUltimateSuite SubscriptionPlan = "ultimate_suite" // All modules

	// Meta plans
	PlanFullAccess SubscriptionPlan = "full_access" // All modules, used for coupon-granted trial
)

// SubscriptionBillingPeriod represents the recurrence of billing.
type SubscriptionBillingPeriod string

const (
	BillingMonthly  SubscriptionBillingPeriod = "monthly"
	BillingYearly   SubscriptionBillingPeriod = "yearly"
	BillingLifetime SubscriptionBillingPeriod = "lifetime"
)

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionPastDue   SubscriptionStatus = "past_due"
	SubscriptionSuspended SubscriptionStatus = "suspended"
	SubscriptionExpired   SubscriptionStatus = "expired"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionTrial     SubscriptionStatus = "trial"
)

// TenantSubscription tracks the active subscription for a tenant.
// A tenant can have a single active subscription record; historical records are kept
// for audit purposes (status = expired / cancelled).
type TenantSubscription struct {
	ID            string                    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string                    `gorm:"type:uuid;not null;index"                        json:"tenant_id"`
	Plan          SubscriptionPlan          `gorm:"type:varchar(50);not null"                       json:"plan"`
	BillingPeriod SubscriptionBillingPeriod `gorm:"type:varchar(20);not null;default:'monthly'"     json:"billing_period"`
	Status        SubscriptionStatus        `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	UserCount     int                       `gorm:"type:int;not null;default:1"                     json:"user_count"`
	SeatLimit     int                       `gorm:"type:int;not null;default:1"                     json:"seat_limit"`
	OutletLimit   int                       `gorm:"type:int;not null;default:1"                     json:"outlet_limit"`
	AmountPaidIDR int64                     `gorm:"type:bigint;not null;default:0"                  json:"amount_paid_idr"`
	StartsAt      time.Time                 `gorm:"type:timestamptz;not null"                       json:"starts_at"`
	ExpiresAt     *time.Time                `gorm:"type:timestamptz"                                json:"expires_at"` // nil = lifetime
	NextBillingAt *time.Time                `gorm:"type:timestamptz"                                json:"next_billing_at,omitempty"`
	CouponID      *string                   `gorm:"type:uuid"                                       json:"coupon_id,omitempty"`
	// Xendit invoice fields (one-time payment)
	XenditPaymentID *string `gorm:"type:varchar(255)"                               json:"xendit_payment_id,omitempty"`
	XenditInvoiceID *string `gorm:"type:varchar(255)"                               json:"xendit_invoice_id,omitempty"`
	// Xendit subscription fields (recurring billing)
	XenditSubscriptionID *string        `gorm:"type:varchar(255)"                               json:"xendit_subscription_id,omitempty"`
	XenditCustomerID     *string        `gorm:"type:varchar(255)"                               json:"xendit_customer_id,omitempty"`
	Notes                string         `gorm:"type:text"                                       json:"notes,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index"                                           json:"-"`

	// Associations (read-only)
	Coupon *Coupon `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
}

// TableName specifies the table name for TenantSubscription.
func (TenantSubscription) TableName() string {
	return "tenant_subscriptions"
}

// BeforeCreate hook to generate UUID.
func (s *TenantSubscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// IsActive returns true when the subscription is currently valid.
func (s *TenantSubscription) IsActive() bool {
	if s.Status != SubscriptionActive && s.Status != SubscriptionTrial {
		return false
	}
	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return false
	}
	return true
}
