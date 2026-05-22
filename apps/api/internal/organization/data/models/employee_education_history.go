package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DegreeLevel string

const (
	DegreeLevelElementary DegreeLevel = "ELEMENTARY"
	DegreeLevelJuniorHigh DegreeLevel = "JUNIOR_HIGH"
	DegreeLevelSeniorHigh DegreeLevel = "SENIOR_HIGH"
	DegreeLevelDiploma    DegreeLevel = "DIPLOMA"
	DegreeLevelBachelor   DegreeLevel = "BACHELOR"
	DegreeLevelMaster     DegreeLevel = "MASTER"
	DegreeLevelDoctorate  DegreeLevel = "DOCTORATE"
)

type EmployeeEducationHistory struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_employee_education_employee" json:"employee_id"`
	Institution  string         `gorm:"type:varchar(200);not null" json:"institution"`
	Degree       DegreeLevel    `gorm:"type:varchar(20);not null" json:"degree"`
	FieldOfStudy string         `gorm:"type:varchar(200)" json:"field_of_study"`
	StartDate    time.Time      `gorm:"type:date;not null" json:"start_date"`
	EndDate      *time.Time     `gorm:"type:date" json:"end_date"`
	GPA          *float32       `gorm:"type:decimal(3,2)" json:"gpa"`
	Description  string         `gorm:"type:text" json:"description"`
	DocumentPath string         `gorm:"type:varchar(255)" json:"document_path"`
	CreatedBy    uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy    *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (EmployeeEducationHistory) TableName() string {
	return "employee_education_histories"
}

func (eeh *EmployeeEducationHistory) BeforeCreate(tx *gorm.DB) error {
	if eeh.ID == uuid.Nil {
		eeh.ID = uuid.New()
	}
	return nil
}

func (eeh *EmployeeEducationHistory) IsCompleted() bool {
	return eeh.EndDate != nil
}

func (eeh *EmployeeEducationHistory) GetDurationYears() float64 {
	endDate := apptime.Now()
	if eeh.EndDate != nil {
		endDate = *eeh.EndDate
	}
	duration := endDate.Sub(eeh.StartDate)
	return duration.Hours() / (24 * 365)
}
