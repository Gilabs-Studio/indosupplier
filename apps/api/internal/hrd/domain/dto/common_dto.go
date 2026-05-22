package dto

import "github.com/google/uuid"

// EmployeeSimpleResponse represents a minimal employee response
// This is used across HRD modules for employee references
type EmployeeSimpleResponse struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Position     string    `json:"position"`
	Department   string    `json:"department"`
}

// EmployeeFormOption represents an employee option for dropdowns
// This is used across HRD modules for form selections
type EmployeeFormOption struct {
	ID           uuid.UUID `json:"id"`
	EmployeeCode string    `json:"employee_code"`
	Name         string    `json:"name"`
}
