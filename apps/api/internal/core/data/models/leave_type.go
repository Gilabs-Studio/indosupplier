package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeaveType represents a type of employee leave
type LeaveType struct {
	ID               string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code             string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Name             string         `gorm:"type:varchar(100);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	MaxDays          int            `gorm:"not null;default:0" json:"max_days"`
	IsPaid           bool           `gorm:"default:true" json:"is_paid"`
	IsCutAnnualLeave bool           `gorm:"default:true" json:"is_cut_annual_leave"` // Whether this leave type deducts from annual leave quota
	IsActive         bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for LeaveType
func (LeaveType) TableName() string {
	return "leave_types"
}

// BeforeCreate hook to generate UUID
func (l *LeaveType) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
