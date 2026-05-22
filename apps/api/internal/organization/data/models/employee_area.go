package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmployeeArea represents the M:N relationship between Employee and Area.
// IsSupervisor=true means the employee is a supervisor of the area,
// IsSupervisor=false means the employee is a regular member.
type EmployeeArea struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID   string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_area" json:"employee_id"`
	Employee     *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	AreaID       string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_area" json:"area_id"`
	Area         *Area     `gorm:"foreignKey:AreaID" json:"area,omitempty"`
	IsSupervisor bool      `gorm:"default:false;index;not null" json:"is_supervisor"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for EmployeeArea
func (EmployeeArea) TableName() string {
	return "employee_areas"
}

// BeforeCreate hook to generate UUID
func (ea *EmployeeArea) BeforeCreate(tx *gorm.DB) error {
	if ea.ID == "" {
		ea.ID = uuid.New().String()
	}
	return nil
}
