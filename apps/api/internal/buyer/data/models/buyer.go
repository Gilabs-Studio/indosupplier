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
	BuyerProfileID    string         `gorm:"type:uuid;not null;index:idx_bookmark_buyer_item" json:"buyer_profile_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index:idx_bookmark_buyer_item" json:"supplier_profile_id"`
	SupplierProductID *string        `gorm:"type:uuid;index:idx_bookmark_buyer_item;nullable" json:"supplier_product_id"`
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

type SupplierFollowing struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID    string         `gorm:"type:uuid;not null;index:idx_buyer_supplier_following,unique" json:"buyer_profile_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index:idx_buyer_supplier_following,unique" json:"supplier_profile_id"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierFollowing) TableName() string {
	return "buyer_supplier_followings"
}

func (f *SupplierFollowing) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
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

type ComparisonProductSessionItem struct {
	ID                  string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ComparisonSessionID string    `gorm:"type:uuid;not null;index:idx_comparison_session_product,unique" json:"comparison_session_id"`
	SupplierProductID   string    `gorm:"type:uuid;not null;index:idx_comparison_session_product,unique" json:"supplier_product_id"`
	SortOrder           int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (ComparisonProductSessionItem) TableName() string {
	return "comparison_product_session_items"
}

func (c *ComparisonProductSessionItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type PurchaseOrder struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PONumber          string         `gorm:"type:varchar(120);not null;uniqueIndex" json:"po_number"`
	BuyerProfileID    string         `gorm:"type:uuid;not null;index" json:"buyer_profile_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	RFQID             *string        `gorm:"type:uuid;index" json:"rfq_id"`
	ProductName       string         `gorm:"type:varchar(255);not null" json:"product_name"`
	QuantityValue     float64        `gorm:"not null;default:0" json:"quantity_value"`
	QuantityUnit      string         `gorm:"type:varchar(60)" json:"quantity_unit"`
	PricePerUnit      float64        `gorm:"not null;default:0" json:"price_per_unit"`
	TotalAmount       float64        `gorm:"not null;default:0" json:"total_amount"`
	Status            string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	PaymentStatus     string         `gorm:"type:varchar(40);not null;default:'unpaid';index" json:"payment_status"`
	DeliveryAddress   string         `gorm:"type:text" json:"delivery_address"`
	Notes             string         `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

func (p *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Status == "" {
		p.Status = "pending"
	}
	if p.PaymentStatus == "" {
		p.PaymentStatus = "unpaid"
	}
	return nil
}
