package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AISearchLog struct {
	ID               string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID   string         `gorm:"type:uuid;index" json:"buyer_profile_id"`
	QueryText        string         `gorm:"type:text;not null" json:"query_text"`
	ParsedIntentJSON string         `gorm:"type:jsonb;not null;default:'{}'" json:"parsed_intent_json"`
	FilterJSON       string         `gorm:"type:jsonb;not null;default:'{}'" json:"filter_json"`
	ResultCount      int            `gorm:"not null;default:0" json:"result_count"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AISearchLog) TableName() string {
	return "ai_search_logs"
}

func (a *AISearchLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.ParsedIntentJSON == "" {
		a.ParsedIntentJSON = "{}"
	}
	if a.FilterJSON == "" {
		a.FilterJSON = "{}"
	}
	return nil
}

type SearchBoostCampaign struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	AdProductID       string         `gorm:"type:uuid;index" json:"ad_product_id"`
	CategoryID        string         `gorm:"type:uuid;index" json:"category_id"`
	SearchBoostWeight float64        `gorm:"not null;default:0" json:"search_boost_weight"`
	TargetKeywords    string         `gorm:"type:text" json:"target_keywords"`
	StartAt           *time.Time     `json:"start_at"`
	EndAt             *time.Time     `json:"end_at"`
	Status            string         `gorm:"type:varchar(40);not null;default:'draft';index" json:"status"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SearchBoostCampaign) TableName() string {
	return "search_boost_campaigns"
}

func (s *SearchBoostCampaign) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "draft"
	}
	return nil
}
