package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BuyerProfile struct {
	ID                  string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID              string         `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	FullName            string         `gorm:"type:varchar(255);not null;index" json:"full_name"`
	CompanyName         string         `gorm:"type:varchar(255);not null;index" json:"company_name"`
	CountryCode         string         `gorm:"type:varchar(10);index" json:"country_code"`
	Industry            string         `gorm:"type:varchar(150);index" json:"industry"`
	PurchaseFrequency   string         `gorm:"type:varchar(80)" json:"purchase_frequency"`
	CompanyVerifiedAt   *time.Time     `json:"company_verified_at"`
	ProfileCompleteness int            `gorm:"not null;default:0" json:"profile_completeness"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BuyerProfile) TableName() string {
	return "buyer_profiles"
}

func (b *BuyerProfile) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

type BuyerDocument struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID string         `gorm:"type:uuid;not null;index" json:"buyer_profile_id"`
	DocumentType   string         `gorm:"type:varchar(80);not null;index" json:"document_type"`
	DocumentNumber string         `gorm:"type:varchar(120)" json:"document_number"`
	FileURL        string         `gorm:"type:text;not null" json:"file_url"`
	Status         string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	ReviewedBy     string         `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt     *time.Time     `json:"reviewed_at"`
	ReviewReason   string         `gorm:"type:text" json:"review_reason"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BuyerDocument) TableName() string {
	return "buyer_documents"
}

func (b *BuyerDocument) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.Status == "" {
		b.Status = "pending"
	}
	return nil
}

type Bookmark struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID    string         `gorm:"type:uuid;not null;index:idx_bookmark_buyer_supplier,unique" json:"buyer_profile_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index:idx_bookmark_buyer_supplier,unique" json:"supplier_profile_id"`
	Notes             string         `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Bookmark) TableName() string {
	return "bookmarks"
}

func (b *Bookmark) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

type ComparisonSession struct {
	ID             string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID string     `gorm:"type:uuid;not null;index" json:"buyer_profile_id"`
	ShareToken     string     `gorm:"type:varchar(120);uniqueIndex" json:"share_token"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (ComparisonSession) TableName() string {
	return "comparison_sessions"
}

func (c *ComparisonSession) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type ComparisonSessionItem struct {
	ID                  string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ComparisonSessionID string    `gorm:"type:uuid;not null;index:idx_comparison_session_supplier,unique" json:"comparison_session_id"`
	SupplierProfileID   string    `gorm:"type:uuid;not null;index:idx_comparison_session_supplier,unique" json:"supplier_profile_id"`
	SortOrder           int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (ComparisonSessionItem) TableName() string {
	return "comparison_session_items"
}

func (c *ComparisonSessionItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
