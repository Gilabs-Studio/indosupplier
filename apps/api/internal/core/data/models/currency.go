package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Currency represents a master-data currency used across banking and finance forms.
type Currency struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code          string         `gorm:"type:varchar(10);not null;uniqueIndex" json:"code"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Symbol        string         `gorm:"type:varchar(10)" json:"symbol"`
	DecimalPlaces int            `gorm:"not null;default:2" json:"decimal_places"`
	IsActive      bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Currency) TableName() string {
	return "currencies"
}

func (c *Currency) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
