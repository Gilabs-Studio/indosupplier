package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant represents a SaaS tenant (organization) in the platform
type Tenant struct {
	ID                     string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name                   string         `gorm:"type:varchar(255);not null;index" json:"name"`
	Slug                   string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	OwnerUserID            *string        `gorm:"type:uuid" json:"owner_user_id"`
	Status                 string         `gorm:"type:varchar(20);not null;default:'active';index" json:"status"` // active, suspended, trial, pending_deletion
	Plan                   string         `gorm:"type:varchar(50);not null;default:'free'" json:"plan"`           // free, starter, professional, enterprise
	MaxUsers               int            `gorm:"type:int;not null;default:5" json:"max_users"`
	DeletionRequestedAt    *time.Time     `gorm:"type:timestamptz;index" json:"deletion_requested_at,omitempty"`
	DeletionScheduledAt    *time.Time     `gorm:"type:timestamptz;index" json:"deletion_scheduled_at,omitempty"`
	DeletionRequestedBy    *string        `gorm:"type:uuid" json:"deletion_requested_by,omitempty"`
	DeletionRecoveredAt    *time.Time     `gorm:"type:timestamptz" json:"deletion_recovered_at,omitempty"`
	DeletionPreviousStatus *string        `gorm:"type:varchar(20)" json:"deletion_previous_status,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Tenant
func (Tenant) TableName() string {
	return "tenants"
}

// BeforeCreate hook to generate UUID
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
