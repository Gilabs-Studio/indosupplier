package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type AttendanceEventType string

const (
	AttendanceEventTodayUpdated AttendanceEventType = "hrd.attendance.today_updated"
)

type AttendanceEvent struct {
	Event     AttendanceEventType    `json:"event"`
	Type      AttendanceEventType    `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	TenantID  string                 `json:"tenant_id"`
	Timestamp string                 `json:"timestamp"`
}

type attendanceSubscriber struct {
	id       string
	conn     *websocket.Conn
	tenantID string
	mu       sync.Mutex
}

type AttendanceHub struct {
	mu          sync.RWMutex
	subscribers map[string]*attendanceSubscriber
}

var defaultAttendanceHub = NewAttendanceHub()

func NewAttendanceHub() *AttendanceHub {
	return &AttendanceHub{
		subscribers: make(map[string]*attendanceSubscriber),
	}
}

func DefaultAttendanceHub() *AttendanceHub {
	return defaultAttendanceHub
}

func (h *AttendanceHub) Register(conn *websocket.Conn, tenantID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.subscribers[id] = &attendanceSubscriber{
		id:       id,
		conn:     conn,
		tenantID: tenantID,
	}

	return id
}

func (h *AttendanceHub) Unregister(id string) {
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

func (h *AttendanceHub) Publish(tenantID string, eventType string, payload map[string]interface{}) {
	if tenantID == "" || eventType == "" {
		return
	}

	if payload == nil {
		payload = map[string]interface{}{}
	}

	event := AttendanceEvent{
		Event:     AttendanceEventType(eventType),
		Type:      AttendanceEventType(eventType),
		Data:      payload,
		Payload:   payload,
		TenantID:  tenantID,
		Timestamp: apptime.Now().UTC().Format(time.RFC3339),
	}

	h.deliver(event)
}

func (h *AttendanceHub) deliver(event AttendanceEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	targets := make([]*attendanceSubscriber, 0)
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
