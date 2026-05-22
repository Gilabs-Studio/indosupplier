package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIIntentRegistry defines available intents the AI assistant can execute
type AIIntentRegistry struct {
	ID                   string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	IntentCode           string    `gorm:"type:varchar(100);uniqueIndex:idx_ai_intent_code;not null" json:"intent_code"`
	DisplayName          string    `gorm:"type:varchar(255);not null" json:"display_name"`
	Description          string    `gorm:"type:text" json:"description,omitempty"`
	Module               string    `gorm:"type:varchar(50);not null" json:"module"`
	ActionType           string    `gorm:"type:varchar(20);not null" json:"action_type"`
	RequiredPermission   string    `gorm:"type:varchar(100);not null" json:"required_permission"`
	RequiresConfirmation bool      `gorm:"default:true" json:"requires_confirmation"`
	EndpointPath         string    `gorm:"type:varchar(255)" json:"endpoint_path,omitempty"`
	ParameterSchema      *string   `gorm:"type:jsonb" json:"parameter_schema,omitempty"`
	IsActive             bool      `gorm:"default:true" json:"is_active"`
	CreatedAt            time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt            time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`
}

// TableName returns the table name for AIIntentRegistry
func (AIIntentRegistry) TableName() string {
	return "ai_intent_registry"
}

// BeforeCreate generates UUID before inserting
func (r *AIIntentRegistry) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
