package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoyaltyMember links a customer to a loyalty program and tracks their point balance.
// LifetimePoints is used for tier calculation; PointBalance is the redeemable balance.
type LoyaltyMember struct {
	ID               string     `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CustomerID       string     `gorm:"type:uuid;not null;index" json:"customer_id"`
	ProgramID        string     `gorm:"type:uuid;not null;index" json:"program_id"`
	MemberCode       string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"member_code"`
	EnrolledOutletID *string    `gorm:"type:uuid;index" json:"enrolled_outlet_id,omitempty"`
	CurrentTier      string     `gorm:"type:varchar(50);not null;default:'Bronze'" json:"current_tier"`
	TierBadgeColor   string     `gorm:"type:varchar(20);not null;default:'#CD7F32'" json:"tier_badge_color"`
	// LifetimePoints tracks total ever-earned points used for tier upgrades.
	LifetimePoints int64 `gorm:"default:0;not null" json:"lifetime_points"`
	// PointBalance is the current redeemable balance (earned minus redeemed/expired).
	PointBalance      int64      `gorm:"default:0;not null" json:"point_balance"`
	JoinedAt          time.Time  `json:"joined_at"`
	LastTransactionAt *time.Time `gorm:"type:timestamptz" json:"last_transaction_at,omitempty"`
	TotalTransactions int        `gorm:"default:0;not null" json:"total_transactions"`
	CreatedBy         *string    `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy         *string    `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LoyaltyMember) TableName() string { return "loyalty_members" }

func (lm *LoyaltyMember) BeforeCreate(tx *gorm.DB) error {
	if lm.ID == "" {
		lm.ID = uuid.New().String()
	}
	return nil
}
