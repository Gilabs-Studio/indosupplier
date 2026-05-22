package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAssignmentHistory tracks the assignment history of assets to employees
type AssetAssignmentHistory struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     *uuid.UUID `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	AssetID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_assignment_asset" json:"asset_id"`
	EmployeeID   *uuid.UUID `gorm:"type:uuid;index" json:"employee_id,omitempty"`
	DepartmentID *uuid.UUID `gorm:"type:uuid;index" json:"department_id,omitempty"`
	LocationID   *uuid.UUID `gorm:"type:uuid;index" json:"location_id,omitempty"`
	AssignedAt   time.Time  `gorm:"type:timestamptz;default:now();index" json:"assigned_at"`
	AssignedBy   *uuid.UUID `gorm:"type:uuid" json:"assigned_by,omitempty"`
	ReturnedAt   *time.Time `gorm:"type:timestamptz;index" json:"returned_at,omitempty"`
	ReturnReason *string    `gorm:"type:text" json:"return_reason,omitempty"`
	Notes        *string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`

	// Relations
	Asset      Asset          `gorm:"foreignKey:AssetID" json:"-"`
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Department *Department    `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Location   *AssetLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
	Assigner   *User          `gorm:"foreignKey:AssignedBy" json:"assigned_by_user,omitempty"`
}

// TableName specifies the table name for AssetAssignmentHistory
func (AssetAssignmentHistory) TableName() string {
	return "asset_assignment_histories"
}

// BeforeCreate hook untuk generate UUID dan set timestamp
func (a *AssetAssignmentHistory) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.AssignedAt.IsZero() {
		a.AssignedAt = time.Now()
	}
	return nil
}

// IsActive returns true if the assignment is currently active (not returned)
func (a *AssetAssignmentHistory) IsActive() bool {
	return a.ReturnedAt == nil || a.ReturnedAt.IsZero()
}

// Duration returns the duration of the assignment
func (a *AssetAssignmentHistory) Duration() time.Duration {
	endTime := time.Now()
	if a.ReturnedAt != nil && !a.ReturnedAt.IsZero() {
		endTime = *a.ReturnedAt
	}
	return endTime.Sub(a.AssignedAt)
}
