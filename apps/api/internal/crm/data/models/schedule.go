package models

import (
	"time"

	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Schedule represents a calendar entry for planned activities
type Schedule struct {
	ID                    string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TaskID                *string    `gorm:"type:uuid;index" json:"task_id"`
	Task                  *Task      `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	EmployeeID            string     `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee              *orgModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Title                 string     `gorm:"type:varchar(255);not null" json:"title"`
	Description           string     `gorm:"type:text" json:"description"`
	ScheduledAt           time.Time  `gorm:"not null;index" json:"scheduled_at"`
	EndAt                 *time.Time `json:"end_at"`
	Status                string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, confirmed, completed, cancelled
	ReminderMinutesBefore int        `gorm:"default:30" json:"reminder_minutes_before"`
	CreatedBy             *string    `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the database table name
func (Schedule) TableName() string {
	return "crm_schedules"
}

// BeforeCreate generates a UUID if not set
func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
