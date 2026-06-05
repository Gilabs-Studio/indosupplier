package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentID    *string        `gorm:"type:uuid;index;uniqueIndex:idx_categories_slug_parent" json:"parent_id"`
	Slug        string         `gorm:"type:varchar(160);not null;index;uniqueIndex:idx_categories_slug_parent" json:"slug"`
	Name        string         `gorm:"type:varchar(180);not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IconURL     string         `gorm:"type:text" json:"icon_url"`
	SortOrder   int            `gorm:"not null;default:0;index" json:"sort_order"`
	IsActive    bool           `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Category) TableName() string {
	return "categories"
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type SupplierProfile struct {
	ID                     string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID                 string         `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	CompanyName            string         `gorm:"type:varchar(255);not null;index" json:"company_name"`
	CompanyType            string         `gorm:"type:varchar(80);index" json:"company_type"`
	TaxStatus              string         `gorm:"type:varchar(80);index" json:"tax_status"`
	NPWP                   string         `gorm:"type:varchar(80);index" json:"npwp"`
	CountryCode            string         `gorm:"type:varchar(10);index" json:"country_code"`
	ProvinceID             string         `gorm:"type:varchar(80);index" json:"province_id"`
	CityID                 string         `gorm:"type:varchar(80);index" json:"city_id"`
	Address                string         `gorm:"type:text" json:"address"`
	Latitude               *float64       `json:"latitude"`
	Longitude              *float64       `json:"longitude"`
	BusinessHours          string         `gorm:"type:text" json:"business_hours"`
	Timezone               string         `gorm:"type:varchar(80);not null;default:'Asia/Jakarta'" json:"timezone"`
	Description            string         `gorm:"type:text" json:"description"`
	Phone                  string         `gorm:"type:varchar(50)" json:"phone"`
	WhatsApp               string         `gorm:"type:varchar(50)" json:"whatsapp"`
	Email                  string         `gorm:"type:varchar(255)" json:"email"`
	Website                string         `gorm:"type:text" json:"website"`
	VerificationLevel      int            `gorm:"not null;default:1;index" json:"verification_level"`
	IsPremiumVerified      bool           `gorm:"not null;default:false;index" json:"is_premium_verified"`
	ResponseRate           float64        `gorm:"not null;default:0" json:"response_rate"`
	AvgResponseTimeMinutes int            `gorm:"not null;default:0" json:"avg_response_time_minutes"`
	StarRating             float64        `gorm:"not null;default:0" json:"star_rating"`
	ReviewCount            int            `gorm:"not null;default:0" json:"review_count"`
	ProfileCompleteness    int            `gorm:"not null;default:0" json:"profile_completeness"`
	Status                 string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	CreatedAt              time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierProfile) TableName() string {
	return "supplier_profiles"
}

func (s *SupplierProfile) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Timezone == "" {
		s.Timezone = "Asia/Jakarta"
	}
	if s.Status == "" {
		s.Status = "draft"
	}
	return nil
}

type SupplierCategory struct {
	ID                string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string    `gorm:"type:uuid;not null;index:idx_supplier_category,unique" json:"supplier_profile_id"`
	CategoryID        string    `gorm:"type:uuid;not null;index:idx_supplier_category,unique" json:"category_id"`
	IsPrimary         bool      `gorm:"not null;default:false;index" json:"is_primary"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (SupplierCategory) TableName() string {
	return "supplier_categories"
}

func (s *SupplierCategory) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type SupplierProduct struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	CategoryID        string         `gorm:"type:uuid;index" json:"category_id"`
	Name              string         `gorm:"type:varchar(255);not null;index" json:"name"`
	Description       string         `gorm:"type:text" json:"description"`
	MOQ               string         `gorm:"type:varchar(120)" json:"moq"`
	StartingPrice     float64        `gorm:"not null;default:0" json:"starting_price"`
	Currency          string         `gorm:"type:varchar(10);not null;default:'IDR'" json:"currency"`
	CapacityText      string         `gorm:"type:varchar(255)" json:"capacity_text"`
	IsFeatured        bool           `gorm:"not null;default:false;index" json:"is_featured"`
	SortOrder         int            `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierProduct) TableName() string {
	return "supplier_products"
}

func (s *SupplierProduct) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Currency == "" {
		s.Currency = "IDR"
	}
	return nil
}

type SupplierProductPhoto struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProductID string         `gorm:"type:uuid;not null;index" json:"supplier_product_id"`
	FileURL           string         `gorm:"type:text;not null" json:"file_url"`
	Caption           string         `gorm:"type:varchar(255)" json:"caption"`
	SortOrder         int            `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierProductPhoto) TableName() string {
	return "supplier_product_photos"
}

func (s *SupplierProductPhoto) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type SupplierProductTag struct {
	ID                string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProductID string    `gorm:"type:uuid;not null;index" json:"supplier_product_id"`
	Tag               string    `gorm:"type:varchar(120);not null;index" json:"tag"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SupplierProductTag) TableName() string {
	return "supplier_product_tags"
}

func (s *SupplierProductTag) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type SupplierPhoto struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	Type              string         `gorm:"type:varchar(60);not null;index" json:"type"`
	FileURL           string         `gorm:"type:text;not null" json:"file_url"`
	Caption           string         `gorm:"type:varchar(255)" json:"caption"`
	SortOrder         int            `gorm:"not null;default:0;index" json:"sort_order"`
	IsApproved        bool           `gorm:"not null;default:false;index" json:"is_approved"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierPhoto) TableName() string {
	return "supplier_photos"
}

func (s *SupplierPhoto) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type Certification struct {
	ID          string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string    `gorm:"type:varchar(80);not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"type:varchar(180);not null;index" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (Certification) TableName() string {
	return "certifications"
}

func (c *Certification) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type SupplierCertification struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	CertificationID   string         `gorm:"type:uuid;not null;index" json:"certification_id"`
	CertificateNumber string         `gorm:"type:varchar(160);index" json:"certificate_number"`
	IssuedBy          string         `gorm:"type:varchar(255)" json:"issued_by"`
	IssuedAt          *time.Time     `json:"issued_at"`
	ExpiredAt         *time.Time     `json:"expired_at"`
	FileURL           string         `gorm:"type:text;not null" json:"file_url"`
	Status            string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	ReviewedBy        string         `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt        *time.Time     `json:"reviewed_at"`
	ReviewReason      string         `gorm:"type:text" json:"review_reason"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierCertification) TableName() string {
	return "supplier_certifications"
}

func (s *SupplierCertification) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "pending"
	}
	return nil
}

type SupplierDocument struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	DocumentType      string         `gorm:"type:varchar(80);not null;index" json:"document_type"`
	DocumentNumber    string         `gorm:"type:varchar(120)" json:"document_number"`
	FileURL           string         `gorm:"type:text;not null" json:"file_url"`
	Status            string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	ReviewedBy        string         `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt        *time.Time     `json:"reviewed_at"`
	ReviewReason      string         `gorm:"type:text" json:"review_reason"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierDocument) TableName() string {
	return "supplier_documents"
}

func (s *SupplierDocument) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "pending"
	}
	return nil
}
