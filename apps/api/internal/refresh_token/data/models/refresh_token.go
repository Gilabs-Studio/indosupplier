package models

import (
	"time"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a refresh token entity
type RefreshToken struct {
	ID        string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string     `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	UserID    string     `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenID   string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"token_id"` // JWT ID (jti) claim
	ExpiresAt time.Time  `gorm:"type:timestamp;not null;index" json:"expires_at"`
	Revoked   bool       `gorm:"type:boolean;default:false;not null" json:"revoked"`
	RevokedAt *time.Time `gorm:"type:timestamp;null" json:"revoked_at,omitempty"`
	CreatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not null" json:"updated_at"`
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate hook to generate UUID
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == "" {
		rt.ID = uuid.New().String()
	}
	return nil
}

// IsExpired checks if the token is expired
func (rt *RefreshToken) IsExpired() bool {
	return apptime.Now().After(rt.ExpiresAt)
}

// IsValid checks if the token is valid (not revoked and not expired)
func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}

// Revoke marks the token as revoked
func (rt *RefreshToken) Revoke() {
	now := apptime.Now()
	rt.Revoked = true
	rt.RevokedAt = &now
}
