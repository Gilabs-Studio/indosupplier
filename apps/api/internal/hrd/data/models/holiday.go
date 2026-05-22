package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HolidayType represents the type of holiday
type HolidayType string

const (
	HolidayTypeNational   HolidayType = "NATIONAL"
	HolidayTypeCollective HolidayType = "COLLECTIVE" // Cuti Bersama
	HolidayTypeCompany    HolidayType = "COMPANY"
)

// Holiday represents a holiday entry
type Holiday struct {
	ID          string      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Date        time.Time   `gorm:"type:date;not null;index" json:"date"`
	Name        string      `gorm:"size:200;not null" json:"name"`
	Description string      `gorm:"size:500" json:"description"`
	Type        HolidayType `gorm:"size:20;not null;default:'NATIONAL'" json:"type"`
	Year        int         `gorm:"not null;index" json:"year"`

	// For collective leave, this determines if it cuts annual leave
	IsCollectiveLeave  bool `gorm:"default:false" json:"is_collective_leave"`
	CutsAnnualLeave    bool `gorm:"default:false" json:"cuts_annual_leave"` // If true, deducts from employee's annual leave quota

	// Company scoping — NULL means global (NATIONAL/COLLECTIVE), non-NULL means company-specific
	// WHY: Different companies may have different company holidays while sharing national ones
	CompanyID *string `gorm:"type:uuid;index" json:"company_id"`

	// Recurring flag for annual holidays
	IsRecurring bool `gorm:"default:false" json:"is_recurring"`

	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (h *Holiday) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	if h.Year == 0 {
		h.Year = h.Date.Year()
	}
	return nil
}
