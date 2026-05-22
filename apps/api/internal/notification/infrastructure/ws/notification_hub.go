package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	notificationModels "github.com/gilabs/gims/api/internal/notification/data/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	NotificationEventCreated = "notification.created"
	NotificationEventUpdated = "notification.updated"
	NotificationEventDeleted = "notification.deleted"
)

// NotificationEvent is the envelope sent to notification WebSocket clients.
type NotificationEvent struct {
	Event     string                       `json:"event"`
	Type      string                       `json:"type"`
	Data      notificationModels.Notification `json:"data"`
	TenantID  string                       `json:"tenant_id,omitempty"`
	UserID    string                       `json:"user_id"`
	Timestamp string                       `json:"timestamp"`
}

type notificationSubscriber struct {
	id       string
	conn     *websocket.Conn
	tenantID string
	userID   string
	mu       sync.Mutex
}

// NotificationHub manages user-scoped notification WebSocket subscribers.
type NotificationHub struct {
	mu          sync.RWMutex
	subscribers map[string]*notificationSubscriber
}

var defaultNotificationHub = NewNotificationHub()

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		subscribers: make(map[string]*notificationSubscriber),
	}
}

func DefaultNotificationHub() *NotificationHub {
	return defaultNotificationHub
}

func (h *NotificationHub) Register(conn *websocket.Conn, tenantID, userID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.subscribers[id] = &notificationSubscriber{
		id:       id,
		conn:     conn,
		tenantID: tenantID,
		userID:   userID,
	}
	return id
}

func (h *NotificationHub) Unregister(id string) {
	h.mu.Lock()
	sub, exists := h.subscribers[id]
	if exists {
		delete(h.subscribers, id)
	}
	h.mu.Unlock()

	if exists {
		sub.mu.Lock()
		_ = sub.conn.Close()
		sub.mu.Unlock()
	}
}

func (h *NotificationHub) PublishCreated(notification notificationModels.Notification) {
	h.Publish(NotificationEventCreated, notification)
}

func (h *NotificationHub) Publish(eventType string, notification notificationModels.Notification) {
	if notification.UserID == "" || eventType == "" {
		return
	}

	event := NotificationEvent{
		Event:     eventType,
		Type:      eventType,
		Data:      notification,
		TenantID:  notification.TenantID,
		UserID:    notification.UserID,
		Timestamp: apptime.Now().UTC().Format(time.RFC3339),
	}

	h.deliver(event)
	publishNotificationToRedis(event)
}

func (h *NotificationHub) deliver(event NotificationEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	targets := make([]*notificationSubscriber, 0)
	for _, sub := range h.subscribers {
		if sub.userID != event.UserID {
			continue
		}
		if event.TenantID != "" && sub.tenantID != "" && sub.tenantID != event.TenantID {
			continue
		}
		targets = append(targets, sub)
	}
	h.mu.RUnlock()

	for _, sub := range targets {
		sub.mu.Lock()
		if err := sub.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			sub.mu.Unlock()
			h.Unregister(sub.id)
			continue
		}
		sub.mu.Unlock()
	}
}
