package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierReview struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID    string         `gorm:"type:uuid;not null;index" json:"buyer_profile_id"`
	SupplierProfileID string         `gorm:"type:uuid;not null;index" json:"supplier_profile_id"`
	RFQID             string         `gorm:"type:uuid;index" json:"rfq_id"`
	Rating            int            `gorm:"not null;default:0;index" json:"rating"`
	ReviewText        string         `gorm:"type:text" json:"review_text"`
	SupplierReply     string         `gorm:"type:text" json:"supplier_reply"`
	SupplierRepliedAt *time.Time     `json:"supplier_replied_at"`
	Status            string         `gorm:"type:varchar(40);not null;default:'pending';index" json:"status"`
	ModeratedBy       string         `gorm:"type:uuid;index" json:"moderated_by"`
	ModeratedAt       *time.Time     `json:"moderated_at"`
	ModerationReason  string         `gorm:"type:text" json:"moderation_reason"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SupplierReview) TableName() string {
	return "supplier_reviews"
}

func (s *SupplierReview) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "pending"
	}
	return nil
}

type Notification struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RecipientType string         `gorm:"type:varchar(40);not null;index:idx_notification_recipient" json:"recipient_type"`
	RecipientID   string         `gorm:"type:uuid;not null;index:idx_notification_recipient" json:"recipient_id"`
	Type          string         `gorm:"type:varchar(80);not null;index" json:"type"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Body          string         `gorm:"type:text" json:"body"`
	Channel       string         `gorm:"type:varchar(40);not null;default:'in_app';index" json:"channel"`
	IsRead        bool           `gorm:"not null;default:false;index" json:"is_read"`
	ReadAt        *time.Time     `json:"read_at"`
	RelatedType   string         `gorm:"type:varchar(80);index" json:"related_type"`
	RelatedID     string         `gorm:"type:uuid;index" json:"related_id"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if n.Channel == "" {
		n.Channel = "in_app"
	}
	return nil
}
