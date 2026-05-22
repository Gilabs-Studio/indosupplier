package dto

import "time"

type NotificationResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	EntityLink string     `json:"entity_link,omitempty"`
	IsRead     bool       `json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

type MarkAllAsReadResponse struct {
	Marked int64 `json:"marked"`
}
