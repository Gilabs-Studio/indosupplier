package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateEmployeeContractRequest represents the request to create an employee contract
type CreateEmployeeContractRequest struct {
	ContractNumber string `json:"contract_number" binding:"required,max=50"`
	ContractType   string `json:"contract_type" binding:"required,oneof=PKWTT PKWT Intern"`
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date,omitempty"`
	DocumentPath   string `json:"document_path" binding:"max=255"`
}

// UpdateEmployeeContractRequest represents the request to update an employee contract
type UpdateEmployeeContractRequest struct {
	ContractNumber string `json:"contract_number" binding:"omitempty,max=50"`
	ContractType   string `json:"contract_type" binding:"omitempty,oneof=PKWTT PKWT Intern"`
	StartDate      string `json:"start_date,omitempty"`
	EndDate        string `json:"end_date,omitempty"`
	DocumentPath   string `json:"document_path,omitempty" binding:"omitempty,max=255"`
}

// TerminateEmployeeContractRequest represents the request to terminate a contract
type TerminateEmployeeContractRequest struct {
	Reason string `json:"reason" binding:"required,max=100"`
	Notes  string `json:"notes" binding:"max=1000"`
}

// RenewEmployeeContractRequest represents the request to renew a contract
type RenewEmployeeContractRequest struct {
	ContractNumber string `json:"contract_number" binding:"required,max=50"`
	ContractType   string `json:"contract_type" binding:"required,oneof=PKWTT PKWT Intern"`
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date,omitempty"`
	DocumentPath   string `json:"document_path,omitempty" binding:"omitempty,max=255"`
}

// CorrectEmployeeContractRequest represents the request to correct a contract
type CorrectEmployeeContractRequest struct {
	EndDate      string `json:"end_date,omitempty"`
	DocumentPath string `json:"document_path,omitempty" binding:"omitempty,max=255"`
}

// EmployeeContractResponse represents the employee contract response
type EmployeeContractResponse struct {
	ID                      uuid.UUID  `json:"id"`
	EmployeeID              uuid.UUID  `json:"employee_id"`
	ContractNumber          string     `json:"contract_number"`
	ContractType            string     `json:"contract_type"`
	StartDate               string     `json:"start_date"`
	EndDate                 *string    `json:"end_date,omitempty"`
	DocumentPath            string     `json:"document_path"`
	Status                  string     `json:"status"`
	IsExpiringSoon          bool       `json:"is_expiring_soon"`
	DaysUntilExpiry         *int       `json:"days_until_expiry,omitempty"`
	TerminatedAt            *time.Time `json:"terminated_at,omitempty"`
	TerminationReason       string     `json:"termination_reason,omitempty"`
	TerminationNotes        string     `json:"termination_notes,omitempty"`
	ExpiredAt               *time.Time `json:"expired_at,omitempty"`
	CorrectedFromContractID *uuid.UUID `json:"corrected_from_contract_id,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// ListEmployeeContractsRequest represents the request to list employee contracts
type ListEmployeeContractsRequest struct {
	Page         int     `form:"page" binding:"omitempty,min=1"`
	PerPage      int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	EmployeeID   *string `form:"employee_id" binding:"omitempty"`
	Status       *string `form:"status" binding:"omitempty,oneof=ACTIVE EXPIRED TERMINATED"`
	ContractType *string `form:"contract_type" binding:"omitempty,oneof=PKWTT PKWT Intern"`
}
