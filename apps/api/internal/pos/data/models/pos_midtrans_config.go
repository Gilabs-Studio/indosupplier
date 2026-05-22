package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// XenditEnvironment defines the Xendit API environment
type XenditEnvironment string

const (
	XenditEnvironmentSandbox    XenditEnvironment = "sandbox"
	XenditEnvironmentProduction XenditEnvironment = "production"
)

// XenditConnectionStatus represents the merchant's Xendit account connection state
type XenditConnectionStatus string

const (
	XenditStatusNotConnected XenditConnectionStatus = "not_connected"
	XenditStatusConnected    XenditConnectionStatus = "connected"
	XenditStatusSuspended    XenditConnectionStatus = "suspended"
)

// XenditConfig holds per-company Xendit payment gateway credentials.
// Each company connects their own Xendit sub-account (XenPlatform model).
// All digital payments route through this sub-account via the for-user-id header.
type XenditConfig struct {
	ID               string                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CompanyID        string                 `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	SecretKey        string                 `gorm:"type:varchar(255);not null" json:"-"` // never expose in response
	XenditAccountID  string                 `gorm:"type:varchar(255)" json:"xendit_account_id"` // sub-account ID for XenPlatform routing
	BusinessName     string                 `gorm:"type:varchar(255)" json:"business_name"`
	Environment      XenditEnvironment      `gorm:"type:varchar(20);default:'sandbox'" json:"environment"`
	ConnectionStatus XenditConnectionStatus `gorm:"type:varchar(20);default:'not_connected'" json:"connection_status"`
	IsActive         bool                   `gorm:"default:true" json:"is_active"`
	WebhookToken     string                 `gorm:"type:varchar(255)" json:"-"` // Xendit webhook verification token
	UpdatedBy        *string                `gorm:"type:uuid" json:"updated_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (XenditConfig) TableName() string {
	return "pos_xendit_configs"
}

func (x *XenditConfig) BeforeCreate(tx *gorm.DB) error {
	if x.ID == "" {
		x.ID = uuid.New().String()
	}
	return nil
}

// IsConnected returns true when the account is fully connected and active
func (x *XenditConfig) IsConnected() bool {
	return x.ConnectionStatus == XenditStatusConnected && x.IsActive && x.SecretKey != ""
}
