package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditChange represents a single field change in an audit log
type AuditChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// AuditChanges is a slice of AuditChange that implements SQL driver interfaces
type AuditChanges []AuditChange

// Value implements the driver.Valuer interface
func (ac AuditChanges) Value() (driver.Value, error) {
	if ac == nil {
		return nil, nil
	}
	return json.Marshal(ac)
}

// Scan implements the sql.Scanner interface
func (ac *AuditChanges) Scan(value interface{}) error {
	if value == nil {
		*ac = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ac)
}

// AssetAuditLog represents an audit trail entry for asset changes
type AssetAuditLog struct {
	ID          uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssetID     uuid.UUID          `gorm:"type:uuid;not null;index:idx_audit_asset" json:"asset_id"`
	Action      string             `gorm:"type:varchar(50);not null;index:idx_audit_action" json:"action"` // created, updated, deleted, depreciated, transferred, disposed, sold, revalued, assigned, approved
	Changes     AuditChanges       `gorm:"type:jsonb" json:"changes,omitempty"`
	PerformedBy *uuid.UUID         `gorm:"type:uuid;index" json:"performed_by,omitempty"`
	PerformedAt time.Time          `gorm:"type:timestamptz;default:now();index:idx_audit_date" json:"performed_at"`
	IPAddress   *string            `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent   *string            `gorm:"type:text" json:"user_agent,omitempty"`
	Metadata    MapStringInterface `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional context data
	CreatedAt   time.Time          `gorm:"type:timestamptz;default:now()" json:"created_at"`

	// Relations
	Asset Asset `gorm:"foreignKey:AssetID" json:"-"`
	User  *User `gorm:"foreignKey:PerformedBy" json:"performed_by_user,omitempty"`
}

// TableName specifies the table name for AssetAuditLog
func (AssetAuditLog) TableName() string {
	return "asset_audit_logs"
}

// BeforeCreate hook untuk generate UUID dan set timestamp
func (a *AssetAuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.PerformedAt.IsZero() {
		a.PerformedAt = time.Now()
	}
	return nil
}

// MapStringInterface untuk metadata field yang flexible
type MapStringInterface map[string]interface{}

// Value implements the driver.Valuer interface
func (m MapStringInterface) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *MapStringInterface) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, m)
}
