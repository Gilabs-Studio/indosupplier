package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PasswordResetStatusPending = "pending"
	PasswordResetStatusUsed    = "used"
	PasswordResetStatusExpired = "expired"
)

// PasswordResetRequest represents a password reset request.
type PasswordResetRequest struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"token" binding:"-"`
	Status    string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	ExpiresAt time.Time      `gorm:"type:timestamptz;not null;index" json:"expires_at"`
	UsedAt    *time.Time     `gorm:"type:timestamptz" json:"used_at"`
	IPAddress string         `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PasswordResetRequest) TableName() string {
	return "password_reset_requests"
}

func (p *PasswordResetRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// IsExpired checks if the reset token has expired.
func (p *PasswordResetRequest) IsExpired() bool {
	return apptime.Now().After(p.ExpiresAt)
}

// IsValid checks if the reset token is still pending and not expired.
func (p *PasswordResetRequest) IsValid() bool {
	return p.Status == PasswordResetStatusPending && !p.IsExpired()
}
