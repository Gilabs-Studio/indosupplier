package dto

import "time"

// ─── Coupon DTOs ─────────────────────────────────────────────────────────────

// CreateCouponRequest is the system-admin payload for generating a new coupon.
type CreateCouponRequest struct {
	// Code must be uppercase alphanumeric, 6-32 chars. Leave empty to auto-generate.
	Code                   string  `json:"code"              binding:"omitempty,min=4,max=32,alphanum"`
	Description            string  `json:"description"       binding:"required,min=3,max=255"`
	CouponType             string  `json:"coupon_type"       binding:"required,oneof=full_access pos_only erp_only crm_only hr_only"`
	DiscountType           string  `json:"discount_type"     binding:"omitempty,oneof=trial percent amount"`
	DiscountValue          float64 `json:"discount_value"    binding:"omitempty,min=0"`
	MaxUserCount           int     `json:"max_user_count"    binding:"omitempty,min=0,max=999"`
	LockUserCount          bool    `json:"lock_user_count"`
	PackagePriceMonthlyIDR float64 `json:"package_price_monthly_idr" binding:"omitempty,min=0"`
	PackagePriceYearlyIDR  float64 `json:"package_price_yearly_idr"  binding:"omitempty,min=0"`
	Scope                  string  `json:"scope"             binding:"omitempty,oneof=general tier_specific"`
	TargetPlanSlug         string  `json:"target_plan_slug"  binding:"omitempty,max=64"`
	// DurationDays is how many days access is granted (trial type only). Minimum 1.
	// DurationDays is how many days access is granted (trial type only).
	// Use 0 to indicate lifetime (no duration limit).
	DurationDays    int        `json:"duration_days"     binding:"omitempty,min=0,max=36500"`
	MaxUses         int        `json:"max_uses"          binding:"required,min=1,max=100000"`
	MaxUsesPerEmail int        `json:"max_uses_per_email" binding:"omitempty,min=1,max=100"`
	ExpiresAt       *time.Time `json:"expires_at"         binding:"omitempty"`
}

// UpdateCouponRequest updates an existing coupon while preserving immutable fields (code, created_by).
type UpdateCouponRequest struct {
	Description            string     `json:"description"       binding:"required,min=3,max=255"`
	CouponType             string     `json:"coupon_type"       binding:"required,oneof=full_access pos_only erp_only crm_only hr_only"`
	DiscountType           string     `json:"discount_type"     binding:"omitempty,oneof=trial percent amount"`
	DiscountValue          float64    `json:"discount_value"    binding:"omitempty,min=0"`
	MaxUserCount           int        `json:"max_user_count"    binding:"omitempty,min=0,max=999"`
	LockUserCount          bool       `json:"lock_user_count"`
	PackagePriceMonthlyIDR float64    `json:"package_price_monthly_idr" binding:"omitempty,min=0"`
	PackagePriceYearlyIDR  float64    `json:"package_price_yearly_idr"  binding:"omitempty,min=0"`
	Scope                  string     `json:"scope"             binding:"omitempty,oneof=general tier_specific"`
	TargetPlanSlug         string     `json:"target_plan_slug"  binding:"omitempty,max=64"`
	// DurationDays: allow 0 for lifetime. Make optional so frontend can omit when not applicable.
	DurationDays           int        `json:"duration_days"     binding:"omitempty,min=0,max=36500"`
	MaxUses                int        `json:"max_uses"          binding:"required,min=1,max=100000"`
	MaxUsesPerEmail        int        `json:"max_uses_per_email" binding:"omitempty,min=1,max=100"`
	ExpiresAt              *time.Time `json:"expires_at"         binding:"omitempty"`
}

// CouponResponse is the safe representation of a coupon returned to system admins.
type CouponResponse struct {
	ID                     string     `json:"id"`
	Code                   string     `json:"code"`
	Description            string     `json:"description"`
	CouponType             string     `json:"coupon_type"`
	DiscountType           string     `json:"discount_type"`
	DiscountValue          float64    `json:"discount_value"`
	MaxUserCount           int        `json:"max_user_count"`
	LockUserCount          bool       `json:"lock_user_count"`
	PackagePriceMonthlyIDR float64    `json:"package_price_monthly_idr"`
	PackagePriceYearlyIDR  float64    `json:"package_price_yearly_idr"`
	Scope                  string     `json:"scope"`
	TargetPlanSlug         string     `json:"target_plan_slug,omitempty"`
	DurationDays           int        `json:"duration_days"`
	MaxUses                int        `json:"max_uses"`
	UsedCount              int        `json:"used_count"`
	MaxUsesPerEmail        int        `json:"max_uses_per_email"`
	IsActive               bool       `json:"is_active"`
	ExpiresAt              *time.Time `json:"expires_at"`
	CreatedBy              string     `json:"created_by"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// SetCouponStatusRequest toggles a coupon active/inactive.
type SetCouponStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// CouponListParams query parameters for listing coupons.
type CouponListParams struct {
	Page       int  `form:"page"`
	PerPage    int  `form:"per_page"`
	ActiveOnly bool `form:"active_only"`
}

// ValidateCouponRequest is used during tenant registration to check a coupon code.
type ValidateCouponRequest struct {
	Code string `json:"code" binding:"required,min=1,max=64"`
}

// ValidateCouponResponse is returned to the registration client before they commit.
type ValidateCouponResponse struct {
	Valid bool `json:"valid"`
	// Reason is a machine-readable code for why the coupon is invalid (empty when valid).
	// Values: "not_found", "inactive", "expired", "exhausted", "already_used_by_email", "plan_mismatch"
	Reason                 string  `json:"reason,omitempty"`
	Description            string  `json:"description,omitempty"`
	CouponType             string  `json:"coupon_type,omitempty"`
	DiscountType           string  `json:"discount_type,omitempty"`
	DiscountValue          float64 `json:"discount_value,omitempty"`
	MaxUserCount           int     `json:"max_user_count,omitempty"`
	LockUserCount          bool    `json:"lock_user_count,omitempty"`
	PackagePriceMonthlyIDR float64 `json:"package_price_monthly_idr,omitempty"`
	PackagePriceYearlyIDR  float64 `json:"package_price_yearly_idr,omitempty"`
	Scope                  string  `json:"scope,omitempty"`
	TargetPlanSlug         string  `json:"target_plan_slug,omitempty"`
	DurationDays           int     `json:"duration_days,omitempty"`
}

// ─── Subscription DTOs ───────────────────────────────────────────────────────

// SubscriptionResponse is the representation of a tenant subscription.
type SubscriptionResponse struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	TenantName      string     `json:"tenant_name,omitempty"`
	Plan            string     `json:"plan"`
	BillingPeriod   string     `json:"billing_period"`
	Status          string     `json:"status"`
	StartsAt        time.Time  `json:"starts_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	NextBillingAt   *time.Time `json:"next_billing_at,omitempty"`
	UserCount       int        `json:"user_count"`
	SeatLimit       int        `json:"seat_limit,omitempty"`
	OutletLimit     int        `json:"outlet_limit,omitempty"`
	ActiveUsers     int        `json:"active_users,omitempty"`
	AmountPaidIDR   int64      `json:"amount_paid_idr"`
	CouponCode      string     `json:"coupon_code,omitempty"`
	XenditPaymentID string     `json:"xendit_payment_id,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// SubscriptionListParams query parameters for listing subscriptions.
type SubscriptionListParams struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

// ─── Subscription Plan Config DTOs ───────────────────────────────────────────

// SubscriptionPlanResponse is the public-facing representation of a plan.
type SubscriptionPlanResponse struct {
	Slug                  string         `json:"slug"`
	Name                  string         `json:"name"`
	Category              string         `json:"category"`
	Description           string         `json:"description,omitempty"`
	BillingType           string         `json:"billing_type"`
	PriceMonthlyIDR       int64          `json:"price_monthly_idr"`
	PriceYearlyIDR        int64          `json:"price_yearly_idr"`
	OutletAddonMonthlyIDR int64          `json:"outlet_addon_monthly_idr"`
	OutletAddonYearlyIDR  int64          `json:"outlet_addon_yearly_idr"`
	MinUsers              int            `json:"min_users"`
	MaxUsers              int            `json:"max_users"`
	IsActive              bool           `json:"is_active"`
	IsHighlighted         bool           `json:"is_highlighted"`
	SortOrder             int            `json:"sort_order"`
	Features              []string       `json:"features"`
	RoleTemplates         []RoleTemplate `json:"role_templates"`
	Modules               []string       `json:"modules"` // enabled module slugs
	MenuURLs              []string       `json:"menu_urls"`
}

// RoleTemplate is a plan-provisioned role template.
type RoleTemplate struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpsertPlanRequest is the system-admin payload for creating or updating a plan.
type UpsertPlanRequest struct {
	Slug                  string         `json:"slug"              binding:"required,min=2,max=64"`
	Name                  string         `json:"name"              binding:"required,min=2,max=128"`
	Category              string         `json:"category"          binding:"required,oneof=pos erp crm hr bundle"`
	Description           string         `json:"description"       binding:"omitempty,max=1000"`
	BillingType           string         `json:"billing_type"      binding:"required,oneof=per_user flat"`
	PriceMonthlyIDR       int64          `json:"price_monthly_idr" binding:"required,min=0"`
	PriceYearlyIDR        int64          `json:"price_yearly_idr"  binding:"required,min=0"`
	OutletAddonMonthlyIDR int64          `json:"outlet_addon_monthly_idr" binding:"omitempty,min=0"`
	OutletAddonYearlyIDR  int64          `json:"outlet_addon_yearly_idr"  binding:"omitempty,min=0"`
	MinUsers              int            `json:"min_users"         binding:"omitempty,min=1"`
	MaxUsers              int            `json:"max_users"         binding:"omitempty,min=1"`
	IsHighlighted         bool           `json:"is_highlighted"`
	SortOrder             int            `json:"sort_order"`
	Features              []string       `json:"features"          binding:"omitempty"`
	RoleTemplates         []RoleTemplate `json:"role_templates" binding:"required,min=1"`
	Modules               []string       `json:"modules"           binding:"omitempty"` // module slugs to entitle
	MenuURLs              []string       `json:"menu_urls"         binding:"omitempty"`
}

// ComputePriceRequest is the payload for the public price-computation endpoint.
type ComputePriceRequest struct {
	Slug          string `json:"slug"           binding:"required"`
	BillingPeriod string `json:"billing_period" binding:"required,oneof=monthly yearly"`
	UserCount     int    `json:"user_count"     binding:"omitempty,min=1,max=500"`
	CouponCode    string `json:"coupon_code"    binding:"omitempty,max=64"`
}

// ComputePriceResponse is the server-computed price breakdown.
type ComputePriceResponse struct {
	Slug           string  `json:"slug"`
	BillingPeriod  string  `json:"billing_period"`
	UserCount      int     `json:"user_count"`
	BaseAmountIDR  int64   `json:"base_amount_idr"`
	DiscountIDR    int64   `json:"discount_idr"`
	FinalAmountIDR int64   `json:"final_amount_idr"`
	CouponApplied  bool    `json:"coupon_applied"`
	CouponCode     string  `json:"coupon_code,omitempty"`
	DiscountType   string  `json:"discount_type,omitempty"`
	DiscountValue  float64 `json:"discount_value,omitempty"`
}

// BillingChangeAction describes the self-service billing lifecycle action.
type BillingChangeAction string

const (
	BillingChangeActionAddSeat     BillingChangeAction = "add_seat"
	BillingChangeActionAddOutlet   BillingChangeAction = "add_outlet"
	BillingChangeActionUpgradePlan BillingChangeAction = "upgrade_plan"
	BillingChangeActionAddModule   BillingChangeAction = "add_module"
	BillingChangeActionDowngrade   BillingChangeAction = "downgrade"
	BillingChangeActionDowngradeOutlet BillingChangeAction = "downgrade_outlet"
)

// BillingChangeRequest is the tenant-facing payload for requesting a billing change.
type BillingChangeRequest struct {
	Action           BillingChangeAction `json:"action" binding:"required,oneof=add_seat add_outlet upgrade_plan add_module downgrade downgrade_outlet"`
	Target           string              `json:"target" binding:"required,min=1,max=128"`
	ActionDate       string              `json:"action_date" binding:"required,datetime=2006-01-02"`
	CouponCode       string              `json:"coupon_code,omitempty" binding:"omitempty,max=64"`
	IdempotencyKey   string              `json:"idempotency_key,omitempty" binding:"omitempty,max=255"`
	XenditCustomerID string              `json:"xendit_customer_id,omitempty" binding:"omitempty,max=255"`
	XenditPlanID     string              `json:"xendit_plan_id,omitempty" binding:"omitempty,max=255"`
}

// BillingChangeResponse represents the computed billing delta and checkout payload.
type BillingChangeResponse struct {
	Status        string `json:"status"`
	ErrorCode     string `json:"error_code,omitempty"`
	SyncRequired  bool   `json:"sync_required,omitempty"`
	BillingChange struct {
		OldSeatLimit          int    `json:"old_seat_limit"`
		NewSeatLimit          int    `json:"new_seat_limit"`
		OldOutletLimit        int    `json:"old_outlet_limit"`
		NewOutletLimit        int    `json:"new_outlet_limit"`
		OldAmountPerCycle     int64  `json:"old_amount_per_cycle"`
		NewAmountPerCycle     int64  `json:"new_amount_per_cycle"`
		ProrationAmount       int64  `json:"proration_amount"`
		ProrationWaivedReason string `json:"proration_waived_reason"`
	} `json:"billing_change"`
	CouponApplied struct {
		Code           string `json:"code"`
		DiscountAmount int64  `json:"discount_amount"`
		MarkAsUsed     bool   `json:"mark_as_used"`
	} `json:"coupon_applied"`
	XenditAction     string         `json:"xendit_action"`
	XenditPayload    map[string]any `json:"xendit_payload"`
	UserNotification struct {
		Title        string `json:"title"`
		Message      string `json:"message"`
		AmountDueNow int64  `json:"amount_due_now"`
	} `json:"user_notification"`
}
