package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VerificationRequest struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	RequestType       string         `gorm:"type:varchar(80);not null;index" json:"request_type"`
	Status            string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	LegalName         string         `gorm:"type:varchar(255)" json:"legal_name"`
	NIBNumber         string         `gorm:"type:varchar(80)" json:"nib_number"`
	EstablishedDate   string         `gorm:"type:varchar(40)" json:"established_date"`
	Industry          string         `gorm:"type:varchar(120)" json:"industry"`
	BusinessPhone     string         `gorm:"type:varchar(50)" json:"business_phone"`
	BusinessAddress   string         `gorm:"type:text" json:"business_address"`
	BusinessDesc      string         `gorm:"type:text" json:"business_desc"`
	NPWPNumber        string         `gorm:"type:varchar(80)" json:"npwp_number"`
	NIBFileURL        string         `gorm:"type:text" json:"nib_file_url"`
	NPWPFileURL       string         `gorm:"type:text" json:"npwp_file_url"`
	AktaFileURL       string         `gorm:"type:text" json:"akta_file_url"`
	DirectorName      string         `gorm:"type:varchar(255)" json:"director_name"`
	DirectorNIK       string         `gorm:"type:varchar(80)" json:"director_nik"`
	BankName          string         `gorm:"type:varchar(120)" json:"bank_name"`
	AccountNumber     string         `gorm:"type:varchar(80)" json:"account_number"`
	AccountName       string         `gorm:"type:varchar(255)" json:"account_name"`
	SubmittedAt       *time.Time     `json:"submitted_at"`
	ReviewedBy        *string        `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt        *time.Time     `json:"reviewed_at"`
	ReviewReason      string         `gorm:"type:text" json:"review_reason"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VerificationRequest) TableName() string {
	return "verification_requests"
}

func (v *VerificationRequest) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	if v.Status == "" {
		v.Status = "draft"
	}
	return nil
}

type SiteVisit struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	ScheduledAt       *time.Time     `json:"scheduled_at"`
	CompletedAt       *time.Time     `json:"completed_at"`
	Result            string         `gorm:"type:varchar(80);index" json:"result"`
	Notes             string         `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SiteVisit) TableName() string {
	return "site_visits"
}

func (s *SiteVisit) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
