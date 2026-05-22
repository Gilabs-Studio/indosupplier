package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type OnboardingEventType string

const (
	OnboardingEventUpdated OnboardingEventType = "general.onboarding.updated"
)

type OnboardingEvent struct {
	Event     OnboardingEventType    `json:"event"`
	Type      OnboardingEventType    `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	TenantID  string                 `json:"tenant_id"`
	Timestamp string                 `json:"timestamp"`
}

type onboardingSubscriber struct {
	id       string
	conn     *websocket.Conn
	tenantID string
	mu       sync.Mutex
}

type OnboardingHub struct {
	mu          sync.RWMutex
	subscribers map[string]*onboardingSubscriber
}

var defaultOnboardingHub = NewOnboardingHub()

func NewOnboardingHub() *OnboardingHub {
	return &OnboardingHub{
		subscribers: make(map[string]*onboardingSubscriber),
	}
}

func DefaultOnboardingHub() *OnboardingHub {
	return defaultOnboardingHub
}

func (h *OnboardingHub) Register(conn *websocket.Conn, tenantID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.subscribers[id] = &onboardingSubscriber{
		id:       id,
		conn:     conn,
		tenantID: tenantID,
	}

	return id
}

func (h *OnboardingHub) Unregister(id string) {
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

func (h *OnboardingHub) Publish(tenantID string, eventType string, payload map[string]interface{}) {
	if tenantID == "" || eventType == "" {
		return
	}

	if payload == nil {
		payload = map[string]interface{}{}
	}

	event := OnboardingEvent{
		Event:     OnboardingEventType(eventType),
		Type:      OnboardingEventType(eventType),
		Data:      payload,
		Payload:   payload,
		TenantID:  tenantID,
		Timestamp: apptime.Now().UTC().Format(time.RFC3339),
	}

	h.deliver(event)
}

func (h *OnboardingHub) deliver(event OnboardingEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	targets := make([]*onboardingSubscriber, 0)
	for _, sub := range h.subscribers {
		if sub.tenantID == event.TenantID {
			targets = append(targets, sub)
		}
	}
	h.mu.RUnlock()

	for _, sub := range targets {
		sub.mu.Lock()
		_ = sub.conn.WriteMessage(websocket.TextMessage, data)
		sub.mu.Unlock()
	}
}
