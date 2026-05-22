package models

import (
	"time"
)

// FeedbackTokenStatus represents the lifecycle state of a one-time feedback token.
type FeedbackTokenStatus string

const (
	FeedbackTokenStatusPending FeedbackTokenStatus = "pending"
	FeedbackTokenStatusUsed    FeedbackTokenStatus = "used"
	FeedbackTokenStatusExpired FeedbackTokenStatus = "expired"
)

// FeedbackToken is a one-time, short-lived token emitted per POS transaction.
// It is embedded as a QR code on the receipt so customers can open the public
// feedback form without authentication.
type FeedbackToken struct {
	ID           string              `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Token        string              `gorm:"type:varchar(128);uniqueIndex;not null" json:"token"`
	FormID       string              `gorm:"type:uuid;not null;index" json:"form_id"`
	OutletID     string              `gorm:"type:uuid;not null;index" json:"outlet_id"`
	// PosOrderID ties this token back to the originating transaction for traceability.
	PosOrderID   *string             `gorm:"type:uuid;index" json:"pos_order_id,omitempty"`
	CustomerName *string             `gorm:"type:varchar(255)" json:"customer_name,omitempty"`
	Status       FeedbackTokenStatus `gorm:"type:varchar(20);default:'pending';not null;index" json:"status"`
	// ExpiresAt defaults to 30 days from creation so tokens don't linger indefinitely.
	ExpiresAt    time.Time           `gorm:"type:timestamptz;not null;index" json:"expires_at"`
	UsedAt       *time.Time          `gorm:"type:timestamptz" json:"used_at,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

func (FeedbackToken) TableName() string { return "feedback_tokens" }

// IsValid returns true when the token can still accept a response.
func (t *FeedbackToken) IsValid() bool {
	return t.Status == FeedbackTokenStatusPending && time.Now().Before(t.ExpiresAt)
}
