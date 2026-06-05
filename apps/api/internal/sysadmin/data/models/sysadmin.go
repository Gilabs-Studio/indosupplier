package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemAdmin struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"type:varchar(255);not null" json:"-"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	PermissionSet string         `gorm:"column:role;type:varchar(50);not null;default:'super_admin';index" json:"permission_set"` // super_admin, content_admin, ads_admin, cs_admin, finance_admin, moderator
	Status        string         `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`                          // active, inactive
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SystemAdmin) TableName() string {
	return "system_admins"
}

func (sa *SystemAdmin) BeforeCreate(tx *gorm.DB) error {
	if sa.ID == "" {
		sa.ID = uuid.New().String()
	}
	if sa.Status == "" {
		sa.Status = "active"
	}
	if sa.PermissionSet == "" {
		sa.PermissionSet = "super_admin"
	}
	return nil
}
