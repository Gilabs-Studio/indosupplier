package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PosTableStatus represents the real-time occupancy status of a floor-plan table object
type PosTableStatus string

const (
	PosTableStatusAvailable PosTableStatus = "AVAILABLE"
	PosTableStatusOccupied  PosTableStatus = "OCCUPIED"
	PosTableStatusReserved  PosTableStatus = "RESERVED"
	PosTableStatusCleaning  PosTableStatus = "CLEANING"
)

// PosTableStatusRecord tracks real-time table occupancy within an active POS session
type PosTableStatusRecord struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SessionID      string         `gorm:"type:uuid;not null;index" json:"session_id"`
	FloorPlanID    string         `gorm:"type:uuid;not null;index" json:"floor_plan_id"`
	TableObjectID  string         `gorm:"type:varchar(100);not null;index" json:"table_object_id"`
	TableLabel     string         `gorm:"type:varchar(50)" json:"table_label"`
	Status         PosTableStatus `gorm:"type:varchar(20);not null;default:'AVAILABLE';index" json:"status"`
	OccupiedSince  *time.Time     `json:"occupied_since"`
	CurrentOrderID *string        `gorm:"type:uuid;index" json:"current_order_id"`
	GuestCount     int            `gorm:"default:0" json:"guest_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PosTableStatusRecord) TableName() string {
	return "pos_table_statuses"
}

func (p *PosTableStatusRecord) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
