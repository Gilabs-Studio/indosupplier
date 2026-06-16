package dto

import "time"

type CreateRoomRequest struct {
	SupplierProfileID string `json:"supplier_profile_id" binding:"required,uuid"`
}

type SendMessageRequest struct {
	Body string `json:"body" binding:"required"`
}

type MessageResponse struct {
	ID         string    `json:"id"`
	ChatRoomID string    `json:"chat_room_id"`
	SenderID   string    `json:"sender_id"`
	SenderType string    `json:"sender_type"`
	Body       string    `json:"body"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type RoomResponse struct {
	ID                string           `json:"id"`
	BuyerProfileID    string           `json:"buyer_profile_id"`
	SupplierProfileID string           `json:"supplier_profile_id"`
	CompanyName       string           `json:"company_name"`
	IsVerified        bool             `json:"is_verified"`
	LastMessage       *MessageResponse `json:"last_message,omitempty"`
	UnreadCount       int              `json:"unread_count"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}
