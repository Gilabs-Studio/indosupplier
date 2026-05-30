package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WaitingList struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	CompanyName string         `gorm:"type:varchar(255);not null" json:"company_name"`
	CompanyType string         `gorm:"type:varchar(50);not null" json:"company_type"` // e.g. supplier, buyer, other
	Phone       string         `gorm:"type:varchar(50)" json:"phone"`
	Notes       string         `gorm:"type:text" json:"notes"`
	Status      string         `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"` // pending, approved, contacted, rejected
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WaitingList) TableName() string {
	return "waiting_list"
}

func (w *WaitingList) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	if w.Status == "" {
		w.Status = "pending"
	}
	return nil
}
