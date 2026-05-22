package models

import (
	"time"

	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus represents the lifecycle state of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the urgency level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// Task represents an actionable item with assignment and priority
type Task struct {
	ID          string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Title       string `gorm:"type:varchar(255);not null;index" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"type:varchar(30);default:'general'" json:"type"` // general, call, email, meeting, follow_up
	Status      string `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	Priority    string `gorm:"type:varchar(10);default:'medium';index" json:"priority"`
	DueDate     *time.Time `gorm:"type:date;index" json:"due_date"`
	CompletedAt *time.Time `json:"completed_at"`

	// Assignment
	AssignedTo   *string             `gorm:"type:uuid;index" json:"assigned_to"`
	AssignedFrom *string             `gorm:"type:uuid" json:"assigned_from"`
	AssignedEmployee *orgModels.Employee `gorm:"foreignKey:AssignedTo" json:"assigned_employee,omitempty"`
	AssignerEmployee *orgModels.Employee `gorm:"foreignKey:AssignedFrom" json:"assigner_employee,omitempty"`

	// Relations (all optional)
	CustomerID *string                  `gorm:"type:uuid;index" json:"customer_id"`
	Customer   *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContactID  *string                  `gorm:"type:uuid;index" json:"contact_id"`
	Contact    *Contact                 `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	DealID     *string                  `gorm:"type:uuid;index" json:"deal_id"`
	Deal       *Deal                    `gorm:"foreignKey:DealID" json:"deal,omitempty"`
	LeadID     *string                  `gorm:"type:uuid;index" json:"lead_id"`
	Lead       *Lead                    `gorm:"foreignKey:LeadID" json:"lead,omitempty"`

	// Metadata
	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Reminders
	Reminders []Reminder `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"reminders,omitempty"`
}

// TableName returns the database table name
func (Task) TableName() string {
	return "crm_tasks"
}

// BeforeCreate generates a UUID if not set
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
