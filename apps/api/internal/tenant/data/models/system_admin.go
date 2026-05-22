package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemAdmin represents a platform-level administrator (completely separate from tenant users)
type SystemAdmin struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Status    string         `gorm:"type:varchar(20);not null;default:'active';index" json:"status"` // active, disabled
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SystemAdmin
func (SystemAdmin) TableName() string {
	return "system_admins"
}

// BeforeCreate hook to generate UUID
func (s *SystemAdmin) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
