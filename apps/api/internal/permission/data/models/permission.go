package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents a permission entity
type Permission struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Code        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"` // Format: resource.action
	MenuID      *string        `gorm:"type:uuid;index" json:"menu_id,omitempty"`
	Menu        *Menu          `gorm:"foreignKey:MenuID" json:"menu,omitempty"`
	Resource    string         `gorm:"type:varchar(100)" json:"resource"`       // Added to match DB schema
	Action      string         `gorm:"type:varchar(50);not null" json:"action"` // VIEW, CREATE, EDIT, DELETE, etc.
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate hook to generate UUID and validate code format
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if !isValidPermissionCode(p.Code) {
		return errors.New("invalid permission code format")
	}
	return nil
}

// BeforeUpdate hook to validate code format
func (p *Permission) BeforeUpdate(tx *gorm.DB) error {
	if p.Code != "" && !isValidPermissionCode(p.Code) {
		return errors.New("invalid permission code format")
	}
	return nil
}

func isValidPermissionCode(code string) bool {
	// Simple check for dot separator
	for _, c := range code {
		if c == '.' {
			return true
		}
	}
	return false
}
