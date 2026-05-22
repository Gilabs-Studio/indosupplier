package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateEmployeeEducationHistoryRequest struct {
	Institution  string   `json:"institution" binding:"required,max=200"`
	Degree       string   `json:"degree" binding:"required,oneof=ELEMENTARY JUNIOR_HIGH SENIOR_HIGH DIPLOMA BACHELOR MASTER DOCTORATE"`
	FieldOfStudy string   `json:"field_of_study" binding:"max=200"`
	StartDate    string   `json:"start_date" binding:"required"`
	EndDate      string   `json:"end_date,omitempty"`
	GPA          *float32 `json:"gpa" binding:"omitempty,min=1,max=4"`
	Description  string   `json:"description"`
	DocumentPath string   `json:"document_path" binding:"max=255"`
}

type UpdateEmployeeEducationHistoryRequest struct {
	Institution  string   `json:"institution" binding:"omitempty,max=200"`
	Degree       string   `json:"degree" binding:"omitempty,oneof=ELEMENTARY JUNIOR_HIGH SENIOR_HIGH DIPLOMA BACHELOR MASTER DOCTORATE"`
	FieldOfStudy string   `json:"field_of_study,omitempty"`
	StartDate    string   `json:"start_date,omitempty"`
	EndDate      string   `json:"end_date,omitempty"`
	GPA          *float32 `json:"gpa" binding:"omitempty,min=1,max=4"`
	Description  string   `json:"description,omitempty"`
	DocumentPath string   `json:"document_path,omitempty" binding:"omitempty,max=255"`
}

type EmployeeEducationHistoryResponse struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	Institution   string    `json:"institution"`
	Degree        string    `json:"degree"`
	FieldOfStudy  string    `json:"field_of_study"`
	StartDate     string    `json:"start_date"`
	EndDate       *string   `json:"end_date,omitempty"`
	GPA           *float32  `json:"gpa,omitempty"`
	Description   string    `json:"description"`
	DocumentPath  string    `json:"document_path"`
	IsCompleted   bool      `json:"is_completed"`
	DurationYears float64   `json:"duration_years"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EmployeeEducationBriefResponse struct {
	ID           string   `json:"id"`
	Institution  string   `json:"institution"`
	Degree       string   `json:"degree"`
	FieldOfStudy string   `json:"field_of_study"`
	StartDate    string   `json:"start_date"`
	EndDate      *string  `json:"end_date,omitempty"`
	GPA          *float32 `json:"gpa,omitempty"`
	DocumentPath string   `json:"document_path"`
	IsCompleted  bool     `json:"is_completed"`
}
