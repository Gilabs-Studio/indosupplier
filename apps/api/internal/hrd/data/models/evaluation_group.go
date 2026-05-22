package models

import (
	"time"

	"gorm.io/gorm"
)

// EvaluationGroup represents a template/group for evaluation criteria
type EvaluationGroup struct {
	ID          string         `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"type:boolean;not null" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Criteria []EvaluationCriteria `gorm:"foreignKey:EvaluationGroupID;references:ID" json:"criteria,omitempty"`
}

func (EvaluationGroup) TableName() string {
	return "evaluation_groups"
}

// TotalWeight returns the sum of all criteria weights in this group
func (eg *EvaluationGroup) TotalWeight() float64 {
	var total float64
	for _, c := range eg.Criteria {
		total += c.Weight
	}
	return total
}
