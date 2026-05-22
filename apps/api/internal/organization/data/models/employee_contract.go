package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContractType represents the type of employment contract
type ContractType string

const (
	ContractTypePKWTT  ContractType = "PKWTT"  // Perjanjian Kerja Waktu Tidak Tertentu (Permanent)
	ContractTypePKWT   ContractType = "PKWT"   // Perjanjian Kerja Waktu Tertentu (Contract)
	ContractTypeIntern ContractType = "Intern" // Magang/Internship
)

// ContractStatus represents the status of an employment contract
type ContractStatus string

const (
	ContractStatusActive     ContractStatus = "ACTIVE"
	ContractStatusExpired    ContractStatus = "EXPIRED"
	ContractStatusTerminated ContractStatus = "TERMINATED"
)

// EmployeeContract represents an employee's employment contract
type EmployeeContract struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_employee_contracts_employee" json:"employee_id"`
	ContractNumber string         `gorm:"type:varchar(50);uniqueIndex:idx_employee_contracts_number;not null" json:"contract_number"`
	ContractType   ContractType   `gorm:"type:varchar(20);not null;index:idx_employee_contracts_type" json:"contract_type"`
	StartDate      time.Time      `gorm:"type:date;not null;index:idx_employee_contracts_dates" json:"start_date"`
	EndDate        *time.Time     `gorm:"type:date;index:idx_employee_contracts_dates" json:"end_date"`
	DocumentPath   string         `gorm:"type:varchar(255)" json:"document_path"`
	Status         ContractStatus `gorm:"type:varchar(20);not null;default:'ACTIVE';index:idx_employee_contracts_status" json:"status"`

	// Lifecycle audit fields
	TerminatedAt            *time.Time `gorm:"type:timestamp" json:"terminated_at"`
	TerminationReason       string     `gorm:"type:varchar(100)" json:"termination_reason"`
	TerminationNotes        string     `gorm:"type:text" json:"termination_notes"`
	ExpiredAt               *time.Time `gorm:"type:timestamp" json:"expired_at"`
	CorrectedFromContractID *uuid.UUID `gorm:"type:uuid" json:"corrected_from_contract_id"`

	CreatedBy uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for EmployeeContract
func (EmployeeContract) TableName() string {
	return "employee_contracts"
}

// IsActive checks if the contract is currently active
func (ec *EmployeeContract) IsActive() bool {
	now := apptime.Now()
	if ec.Status != ContractStatusActive {
		return false
	}
	if now.Before(ec.StartDate) {
		return false
	}
	if ec.EndDate != nil && now.After(*ec.EndDate) {
		return false
	}
	return true
}

// IsExpiringSoon checks if contract is expiring within given days
func (ec *EmployeeContract) IsExpiringSoon(days int) bool {
	if ec.EndDate == nil {
		return false
	}
	threshold := apptime.Now().AddDate(0, 0, days)
	return ec.EndDate.Before(threshold) && ec.EndDate.After(apptime.Now())
}

// DaysUntilExpiry returns days until contract expires
func (ec *EmployeeContract) DaysUntilExpiry() int {
	if ec.EndDate == nil {
		return -1
	}
	return int(ec.EndDate.Sub(apptime.Now()).Hours() / 24)
}
