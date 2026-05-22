package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PipelineStage represents a configurable pipeline stage for deal management
type PipelineStage struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_pipeline_stages_tenant_code" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_pipeline_stages_tenant_code" json:"code"`
	Order       int            `gorm:"not null;index" json:"order"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	Probability int            `gorm:"type:int;default:0" json:"probability"`
	IsWon       bool           `gorm:"default:false" json:"is_won"`
	IsLost      bool           `gorm:"default:false" json:"is_lost"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for PipelineStage
func (PipelineStage) TableName() string {
	return "crm_pipeline_stages"
}

// BeforeCreate hook to generate UUID
func (p *PipelineStage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
