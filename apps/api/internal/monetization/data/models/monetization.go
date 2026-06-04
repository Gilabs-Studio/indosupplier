package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdProduct struct {
	ID            string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code          string    `gorm:"type:varchar(80);not null;uniqueIndex" json:"code"`
	Name          string    `gorm:"type:varchar(180);not null;index" json:"name"`
	AdType        string    `gorm:"type:varchar(60);not null;index" json:"ad_type"`
	PlacementType string    `gorm:"type:varchar(80);not null;index" json:"placement_type"`
	PricingModel  string    `gorm:"type:varchar(60);not null;index" json:"pricing_model"`
	Description   string    `gorm:"type:text" json:"description"`
	IsActive      bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (AdProduct) TableName() string {
	return "ad_products"
}

func (a *AdProduct) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

type AdCampaign struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	AdProductID       string         `gorm:"type:uuid;not null;index" json:"ad_product_id"`
	CategoryID        string         `gorm:"type:uuid;index" json:"category_id"`
	Title             string         `gorm:"type:varchar(255);not null;index" json:"title"`
	Description       string         `gorm:"type:text" json:"description"`
	ImageURL          string         `gorm:"type:text" json:"image_url"`
	StartDate         *time.Time     `json:"start_date"`
	EndDate           *time.Time     `json:"end_date"`
	Status            string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	ApprovalStatus    string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"approval_status"`
	SearchBoostWeight float64        `gorm:"not null;default:0" json:"search_boost_weight"`
	TargetKeywords    string         `gorm:"type:text" json:"target_keywords"`
	ReviewedBy        string         `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt        *time.Time     `json:"reviewed_at"`
	ReviewReason      string         `gorm:"type:text" json:"review_reason"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AdCampaign) TableName() string {
	return "ad_campaigns"
}

func (a *AdCampaign) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	if a.ApprovalStatus == "" {
		a.ApprovalStatus = "pending"
	}
	return nil
}

type AuctionSession struct {
	ID               string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CategoryID       string         `gorm:"type:uuid;not null;index" json:"category_id"`
	SlotCount        int            `gorm:"not null;default:1" json:"slot_count"`
	SlotDurationDays int            `gorm:"not null;default:7" json:"slot_duration_days"`
	MinBidAmount     float64        `gorm:"not null;default:0" json:"min_bid_amount"`
	BiddingStartAt   *time.Time     `json:"bidding_start_at"`
	BiddingEndAt     *time.Time     `json:"bidding_end_at"`
	Status           string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	CreatedBy        string         `gorm:"type:uuid;index" json:"created_by"`
	ClosedBy         string         `gorm:"type:uuid;index" json:"closed_by"`
	ClosedAt         *time.Time     `json:"closed_at"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuctionSession) TableName() string {
	return "auction_sessions"
}

func (a *AuctionSession) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	return nil
}

type AuctionBid struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AuctionSessionID  string         `gorm:"type:uuid;not null;index:idx_auction_supplier" json:"auction_session_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index:idx_auction_supplier" json:"supplier_profile_id"`
	BidAmount         float64        `gorm:"not null;default:0" json:"bid_amount"`
	DepositAmount     float64        `gorm:"not null;default:0" json:"deposit_amount"`
	RankPosition      int            `gorm:"not null;default:0;index" json:"rank_position"`
	Status            string         `gorm:"type:varchar(40);not null;default:'active';index" json:"status"`
	WithdrawnAt       *time.Time     `json:"withdrawn_at"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuctionBid) TableName() string {
	return "auction_bids"
}

func (a *AuctionBid) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Status == "" {
		a.Status = "active"
	}
	return nil
}

type SubscriptionPlan struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code         string    `gorm:"type:varchar(80);not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"type:varchar(180);not null;index" json:"name"`
	BillingCycle string    `gorm:"type:varchar(40);not null;index" json:"billing_cycle"`
	Price        float64   `gorm:"not null;default:0" json:"price"`
	Description  string    `gorm:"type:text" json:"description"`
	BenefitsJSON string    `gorm:"type:jsonb;not null;default:'[]'" json:"benefits_json"`
	IsActive     bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;index" json:"updated_at"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

func (s *SubscriptionPlan) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.BenefitsJSON == "" {
		s.BenefitsJSON = "[]"
	}
	return nil
}

type SupplierSubscription struct {
	ID                 string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID  string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	SubscriptionPlanID string         `gorm:"type:uuid;not null;index" json:"subscription_plan_id"`
	StartAt            *time.Time     `json:"start_at"`
	EndAt              *time.Time     `json:"end_at"`
	Status             string         `gorm:"type:varchar(40);not null;default:'active';index" json:"status"`
	AutoRenew          bool           `gorm:"not null;default:false" json:"auto_renew"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierSubscription) TableName() string {
	return "supplier_subscriptions"
}

func (s *SupplierSubscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	return nil
}

type Payment struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	RelatedType       string         `gorm:"type:varchar(80);not null;index:idx_payment_related" json:"related_type"`
	RelatedID         string         `gorm:"type:uuid;not null;index:idx_payment_related" json:"related_id"`
	Amount            float64        `gorm:"not null;default:0" json:"amount"`
	Currency          string         `gorm:"type:varchar(10);not null;default:'IDR'" json:"currency"`
	Method            string         `gorm:"type:varchar(80);index" json:"method"`
	Status            string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	PaidAt            *time.Time     `json:"paid_at"`
	FailedAt          *time.Time     `json:"failed_at"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Currency == "" {
		p.Currency = "IDR"
	}
	if p.Status == "" {
		p.Status = "pending"
	}
	return nil
}

type Invoice struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID     string         `gorm:"type:uuid;not null;index" json:"payment_id"`
	InvoiceNumber string         `gorm:"type:varchar(120);not null;uniqueIndex" json:"invoice_number"`
	FileURL       string         `gorm:"type:text" json:"file_url"`
	IssuedAt      *time.Time     `json:"issued_at"`
	DueAt         *time.Time     `json:"due_at"`
	PaidAt        *time.Time     `json:"paid_at"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Invoice) TableName() string {
	return "invoices"
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

type Refund struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID   string         `gorm:"type:uuid;not null;index" json:"payment_id"`
	Amount      float64        `gorm:"not null;default:0" json:"amount"`
	Reason      string         `gorm:"type:text" json:"reason"`
	Status      string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	ProcessedAt *time.Time     `json:"processed_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Refund) TableName() string {
	return "refunds"
}

func (r *Refund) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "pending"
	}
	return nil
}
