package models

import (
	"time"

	"gorm.io/gorm"
)

// EvaluationCriteria represents a single evaluation criterion within a group
type EvaluationCriteria struct {
	ID                string         `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EvaluationGroupID string         `gorm:"type:char(36);not null;index" json:"evaluation_group_id"`
	Name              string         `gorm:"type:varchar(200);not null" json:"name"`
	Description       *string        `gorm:"type:text" json:"description"`
	Weight            float64        `gorm:"type:decimal(5,2);not null" json:"weight"` // percentage, sum per group must be 100
	MaxScore          float64        `gorm:"type:decimal(5,2);not null;default:100" json:"max_score"`
	SortOrder         int            `gorm:"type:int;not null;default:0" json:"sort_order"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (EvaluationCriteria) TableName() string {
	return "evaluation_criteria"
}
