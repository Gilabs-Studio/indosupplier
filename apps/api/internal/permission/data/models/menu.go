package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Menu represents a menu entity (hierarchical structure)
type Menu struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Icon        string         `gorm:"type:varchar(100)" json:"icon"`
	URL         string         `gorm:"type:varchar(255);not null" json:"url"`
	ParentID    *string        `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent      *Menu          `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Menu         `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Order       int            `gorm:"type:integer;default:0" json:"order"`
	Status      string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Module      string         `gorm:"type:varchar(100);index" json:"module"`
	Slug        string         `gorm:"type:varchar(255);index" json:"slug"`
	Access      bool           `gorm:"type:boolean;default:true" json:"access"`
	IsActive    bool           `gorm:"type:boolean;default:true" json:"is_active"`
	IsClickable bool           `gorm:"type:boolean;default:true" json:"is_clickable"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Menu
func (Menu) TableName() string {
	return "menus"
}

// BeforeCreate hook to generate UUID
func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
