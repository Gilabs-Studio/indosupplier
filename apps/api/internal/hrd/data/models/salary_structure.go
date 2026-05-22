package models

import (
	"time"

	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalaryStructureStatus string

const (
	SalaryStructureStatusDraft    SalaryStructureStatus = "draft"
	SalaryStructureStatusActive   SalaryStructureStatus = "active"
	SalaryStructureStatusInactive SalaryStructureStatus = "inactive"
)

type SalaryStructure struct {
	ID            string                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string                `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID    string                `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee      *orgModels.Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	EffectiveDate time.Time             `gorm:"type:date;not null;index" json:"effective_date"`
	BasicSalary   float64               `gorm:"type:numeric(15,2);not null" json:"basic_salary"`
	Notes         string                `gorm:"type:text" json:"notes"`
	Status        SalaryStructureStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`

	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalaryStructure) TableName() string {
	return "salary_structures"
}

func (s *SalaryStructure) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
