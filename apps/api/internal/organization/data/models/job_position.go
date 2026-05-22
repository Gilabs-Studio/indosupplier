package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobPosition represents a job position/role within the organization
type JobPosition struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for JobPosition
func (JobPosition) TableName() string {
	return "job_positions"
}

// BeforeCreate hook to generate UUID
func (j *JobPosition) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	return nil
}
