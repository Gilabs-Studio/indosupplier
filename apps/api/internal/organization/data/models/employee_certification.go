package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeCertification struct {
	ID                string         `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID        string         `gorm:"type:uuid;not null;index:idx_employee_certification_employee" json:"employee_id"`
	CertificateName   string         `gorm:"type:varchar(200);not null" json:"certificate_name"`
	IssuedBy          string         `gorm:"type:varchar(200);not null" json:"issued_by"`
	IssueDate         time.Time      `gorm:"type:date;not null" json:"issue_date"`
	ExpiryDate        *time.Time     `gorm:"type:date" json:"expiry_date"`
	CertificateFile   string         `gorm:"type:varchar(255)" json:"certificate_file"`
	CertificateNumber string         `gorm:"type:varchar(100)" json:"certificate_number"`
	Description       string         `gorm:"type:text" json:"description"`
	CreatedBy         string         `gorm:"type:varchar(255)" json:"created_by"`
	UpdatedBy         string         `gorm:"type:varchar(255)" json:"updated_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (e *EmployeeCertification) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

func (EmployeeCertification) TableName() string {
	return "employee_certifications"
}

func (e *EmployeeCertification) IsExpired() bool {
	if e.ExpiryDate == nil {
		return false
	}
	return e.ExpiryDate.Before(apptime.Now())
}

func (e *EmployeeCertification) DaysUntilExpiry() int {
	if e.ExpiryDate == nil {
		return 999999
	}
	duration := time.Until(*e.ExpiryDate)
	return int(duration.Hours() / 24)
}
