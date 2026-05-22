package dto

import "time"

// ─── Config JSON schema structures ────────────────────────────────────────────

// LoyaltyConfigJSON is the parsed form of the JSONB config stored in loyalty_programs.
type LoyaltyConfigJSON struct {
	PointRules      PointRules     `json:"point_rules"`
	Tiers           []TierConfig   `json:"tiers"`
	Rewards         []RewardConfig `json:"rewards"`
	PointExpiryDays int            `json:"point_expiry_days"`
}

type PointRules struct {
	// PointsPerAmount is how many points are awarded per AmountPerPoint IDR spent.
	PointsPerAmount float64 `json:"points_per_amount"`
	// AmountPerPoint is the transaction amount (in local currency) required to earn PointsPerAmount points.
	AmountPerPoint       float64 `json:"amount_per_point"`
	MinTransactionAmount float64 `json:"min_transaction_amount"`
}

type TierConfig struct {
	Name              string  `json:"name"`
	MinPoints         int64   `json:"min_lifetime_points"` // aligned with frontend field name
	Multiplier        float64 `json:"multiplier"`
	BadgeColor        string  `json:"badge_color"`
}

type RewardConfig struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"` // discount_percent | discount_fixed | merch
	Value           float64 `json:"value"`
	PointsRequired  int64   `json:"points_required"`
	IsActive        bool    `json:"is_active"`
	MerchDescription *string `json:"merch_description,omitempty"` // used when Type = "merch"
}

// ─── Program DTOs ─────────────────────────────────────────────────────────────

type CreateLoyaltyProgramRequest struct {
	OutletID    *string           `json:"outlet_id"`
	Name        string            `json:"name" binding:"required,max=255"`
	Description *string           `json:"description"`
	Config      LoyaltyConfigJSON `json:"config" binding:"required"`
	IsActive    bool              `json:"is_active"`
}

type UpdateLoyaltyProgramRequest struct {
	Name        *string            `json:"name"`
	Description *string            `json:"description"`
	Config      *LoyaltyConfigJSON `json:"config"`
	IsActive    *bool              `json:"is_active"`
}

type LoyaltyProgramResponse struct {
	ID          string            `json:"id"`
	OutletID    *string           `json:"outlet_id,omitempty"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Config      LoyaltyConfigJSON `json:"config"`
	MemberCount int64             `json:"member_count"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ─── Member DTOs ──────────────────────────────────────────────────────────────

// EnrollMemberRequest registers a customer into a loyalty program from POS.
// If CustomerID is provided, the existing customer record is used and Phone/Email are optional.
// If CustomerID is empty, a new customer is created — Name is still required.
type EnrollMemberRequest struct {
	CustomerID *string `json:"customer_id"`
	ProgramID  *string `json:"program_id"` // resolved from outlet if omitted
	OutletID   *string `json:"outlet_id"`
	Name       string  `json:"name" binding:"required,max=255"`
	// Phone is optional when CustomerID is provided; the backend looks up the customer record.
	Phone *string `json:"phone" binding:"omitempty,max=30"`
	Email  *string `json:"email"`
	// PosOrderID links the enrollment to the current transaction so points can be credited immediately.
	PosOrderID *string `json:"pos_order_id"`
}

type LoyaltyMemberResponse struct {
	ID               string     `json:"id"`
	CustomerID       string     `json:"customer_id"`
	ProgramID        string     `json:"program_id"`
	MemberCode       string     `json:"member_code"`
	EnrolledOutletID *string    `json:"enrolled_outlet_id,omitempty"`
	EnrolledOutletName *string  `json:"enrolled_outlet_name,omitempty"`
	CurrentTier      string     `json:"current_tier"`
	TierBadgeColor   string     `json:"tier_badge_color"`
	LifetimePoints   int64      `json:"lifetime_points"`
	PointBalance     int64      `json:"point_balance"`
	JoinedAt         time.Time  `json:"joined_at"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty"`
	TotalTransactions int        `json:"total_transactions"`
	// CustomerName is joined from the customers table for display purposes.
	CustomerName  string  `json:"customer_name"`
	CustomerPhone *string `json:"customer_phone,omitempty"`
}

type ChangeProgramRequest struct {
	ProgramID string `json:"program_id" binding:"required"`
}

// LookupMemberResponse is the lightweight response used by the POS terminal
// when detecting whether a customer is a loyalty member.
type LookupMemberResponse struct {
	Found          bool    `json:"found"`
	MemberID       *string `json:"member_id,omitempty"`
	MemberCode     *string `json:"member_code,omitempty"`
	CustomerID     *string `json:"customer_id,omitempty"`
	CustomerName   *string `json:"customer_name,omitempty"`
	CurrentTier    *string `json:"current_tier,omitempty"`
	TierBadgeColor *string `json:"tier_badge_color,omitempty"`
	PointBalance   *int64  `json:"point_balance,omitempty"`
	// AvailableRewards is the list of rewards the customer can currently redeem.
	AvailableRewards []RewardConfig `json:"available_rewards,omitempty"`
}

// ─── Points + Ledger DTOs ─────────────────────────────────────────────────────

// EarnPointsRequest is called internally after a POS/Sales payment is confirmed.
type EarnPointsRequest struct {
	MemberID        string  `json:"member_id"`
	TransactionID   string  `json:"transaction_id"`
	TransactionType string  `json:"transaction_type"` // pos_order | sales_order
	TotalAmount     float64 `json:"total_amount"`
	ProcessedBy     *string `json:"processed_by,omitempty"`
}

// RedeemPointsRequest applies a reward to a POS order.
type RedeemPointsRequest struct {
	MemberID        string  `json:"member_id" binding:"required"`
	RewardID        string  `json:"reward_id" binding:"required"`
	TransactionID   string  `json:"transaction_id" binding:"required"`
	TransactionType string  `json:"transaction_type" binding:"required"`
	ProcessedBy     *string `json:"processed_by,omitempty"`
}

type RedeemPointsResponse struct {
	PointsDeducted   int64   `json:"points_deducted"`
	NewBalance       int64   `json:"new_balance"`
	DiscountAmount   float64 `json:"discount_amount"`
	DiscountType     string  `json:"discount_type"` // discount_percent | discount_fixed | merch
	RewardName       string  `json:"reward_name"`
	// MerchDescription is populated when the reward type is "merch".
	MerchDescription *string `json:"merch_description,omitempty"`
}

// AdjustPointsRequest allows admin to manually add or subtract points.
type AdjustPointsRequest struct {
	MemberID    string  `json:"member_id" binding:"required"`
	Points      int64   `json:"points" binding:"required"`
	Notes       *string `json:"notes"`
	ProcessedBy *string `json:"processed_by,omitempty"`
}

type PointLedgerResponse struct {
	ID              string     `json:"id"`
	MemberID        string     `json:"member_id"`
	EntryType       string     `json:"entry_type"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	TransactionID   *string    `json:"transaction_id,omitempty"`
	Points          int64      `json:"points"`
	BalanceAfter    int64      `json:"balance_after"`
	LifetimeAfter   int64      `json:"lifetime_after"`
	Multiplier      *float64   `json:"multiplier,omitempty"`
	RewardID        *string    `json:"reward_id,omitempty"`
	RewardName      *string    `json:"reward_name,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ─── Public self-register ─────────────────────────────────────────────────────

type PublicSelfRegisterRequest struct {
	Name     string  `json:"name" binding:"required,max=255"`
	Phone    string  `json:"phone" binding:"required,max=30"`
	Email    *string `json:"email"`
	OrderID  string  `json:"order_id" binding:"required"`
	OutletID string  `json:"outlet_id" binding:"required"`
}

type PublicSelfRegisterResponse struct {
	MemberCode     string `json:"member_code"`
	CurrentTier    string `json:"current_tier"`
	TierBadgeColor string `json:"tier_badge_color"`
	PointsEarned   int64  `json:"points_earned"`
	PointBalance   int64  `json:"point_balance"`
	AlreadyMember  bool   `json:"already_member"`
}
