package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PosTableQRToken stores a stable QR token for a specific table object.
// The token UUID is embedded in the customer-facing URL and can be
// independently revoked without affecting the floor plan layout.
type PosTableQRToken struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string         `gorm:"column:tenant_id;type:uuid;not null;index" json:"tenant_id"`
	OutletID      string         `gorm:"type:uuid;not null;index" json:"outlet_id"`
	FloorPlanID   string         `gorm:"type:uuid;not null;index" json:"floor_plan_id"`
	TableObjectID string         `gorm:"type:varchar(100);not null;index" json:"table_object_id"`
	TableLabel    string         `gorm:"type:varchar(50);not null" json:"table_label"`
	Token         string         `gorm:"type:uuid;uniqueIndex;not null" json:"token"`
	IsActive      bool           `gorm:"type:boolean;not null;default:true;index" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PosTableQRToken) TableName() string {
	return "pos_table_qr_tokens"
}

func (t *PosTableQRToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.Token == "" {
		t.Token = uuid.New().String()
	}
	return nil
}
