package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LoyaltyProgram stores the configuration for a loyalty program.
// OutletID = NULL means the program is global (applies to all outlets of a company).
type LoyaltyProgram struct {
	ID          string         `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OutletID    *string        `gorm:"type:uuid;index" json:"outlet_id,omitempty"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	// ConfigJSON holds tiers, point_rules, and rewards as JSONB.
	// Schema: { point_rules, tiers[], rewards[], point_expiry_days }
	ConfigJSON datatypes.JSON `gorm:"type:jsonb;not null" json:"config_json"`
	IsActive    bool          `gorm:"default:true;index" json:"is_active"`
	MemberCount int64         `gorm:"column:member_count;->" json:"member_count"`
	CreatedBy  *string        `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy  *string        `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LoyaltyProgram) TableName() string { return "loyalty_programs" }

func (lp *LoyaltyProgram) BeforeCreate(tx *gorm.DB) error {
	if lp.ID == "" {
		lp.ID = uuid.New().String()
	}
	return nil
}
