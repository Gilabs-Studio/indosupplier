package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	buyer "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	supplier "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

type ChatRoom struct {
	ID                string                    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BuyerProfileID    string                    `gorm:"type:uuid;not null;index:idx_chat_room_participants,unique" json:"buyer_profile_id"`
	SupplierProfileID string                    `gorm:"type:uuid;not null;index:idx_chat_room_participants,unique" json:"supplier_profile_id"`
	LastMessageID     *string                   `gorm:"type:uuid;index" json:"last_message_id"`
	CreatedAt         time.Time                 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time                 `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt         gorm.DeletedAt            `gorm:"index" json:"-"`

	// Relations
	BuyerProfile    *buyer.BuyerProfile       `gorm:"foreignKey:BuyerProfileID" json:"buyer_profile,omitempty"`
	SupplierProfile *supplier.SupplierProfile `gorm:"foreignKey:SupplierProfileID" json:"supplier_profile,omitempty"`
	LastMessage     *ChatMessage              `gorm:"foreignKey:LastMessageID" json:"last_message,omitempty"`
}

func (ChatRoom) TableName() string {
	return "chat_rooms"
}

func (c *ChatRoom) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type ChatMessage struct {
	ID         string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChatRoomID string         `gorm:"type:uuid;not null;index" json:"chat_room_id"`
	SenderID   string         `gorm:"type:uuid;not null;index" json:"sender_id"`
	SenderType string         `gorm:"type:varchar(40);not null" json:"sender_type"` // "buyer" or "supplier"
	Body       string         `gorm:"type:text;not null" json:"body"`
	IsRead     bool           `gorm:"not null;default:false;index" json:"is_read"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (c *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
