package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// PosEventType identifies the kind of real-time POS event.
type PosEventType string

const (
	// PosEventNewCustomerOrder fires when a customer places an order via the self-order QR page.
	PosEventNewCustomerOrder PosEventType = "pos.new_customer_order"
	// PosEventOrderStatusChanged fires when order status transitions (e.g. IN_PROGRESS → READY).
	PosEventOrderStatusChanged PosEventType = "pos.order_status_changed"
	// PosEventPaymentConfirmed fires when order payment is confirmed (cash or webhook-settled).
	PosEventPaymentConfirmed PosEventType = "pos.payment_confirmed"
	// PosEventTableUpdated fires when a table state can be updated from the payload.
	PosEventTableUpdated PosEventType = "pos.table_updated"
)

// PosEvent is the envelope sent over every WebSocket connection.
type PosEvent struct {
	Event     PosEventType           `json:"event"`
	Type      PosEventType           `json:"type"`
	Data      map[string]interface{} `json:"data"`
	// Payload is kept as a backward-compatible alias for older clients.
	Payload   map[string]interface{} `json:"payload,omitempty"`
	OutletID  string                 `json:"outlet_id"`
	TenantID  string                 `json:"tenant_id"`
	Timestamp string                 `json:"timestamp"`
}

type posSubscriber struct {
	id       string
	conn     *websocket.Conn
	tenantID string
	outletID string
	mu       sync.Mutex
}

// PosHub manages WebSocket connections for the POS real-time event system.
// Subscriptions are keyed by tenant_id + outlet_id so events fan-out only to relevant staff.
type PosHub struct {
	mu          sync.RWMutex
	subscribers map[string]*posSubscriber
}

var defaultPosHub = NewPosHub()

// NewPosHub creates an isolated hub instance.
func NewPosHub() *PosHub {
	return &PosHub{
		subscribers: make(map[string]*posSubscriber),
	}
}

// DefaultPosHub returns the process-wide singleton hub.
func DefaultPosHub() *PosHub {
	return defaultPosHub
}

// Register adds a WebSocket connection subscribed to events for one tenant/outlet channel.
// Returns the subscriber ID needed to unregister later.
func (h *PosHub) Register(conn *websocket.Conn, tenantID, outletID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.subscribers[id] = &posSubscriber{
		id:       id,
		conn:     conn,
		tenantID: tenantID,
		outletID: outletID,
	}
	return id
}

// Unregister removes a subscriber and closes its connection.
func (h *PosHub) Unregister(id string) {
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

// PublishEvent broadcasts an event using a typed PosEventType.
func (h *PosHub) PublishEvent(tenantID, outletID string, eventType PosEventType, payload map[string]interface{}) {
	h.Publish(tenantID, outletID, string(eventType), payload)
}

// Publish satisfies the usecase.POSHubPublisher interface with a plain string event type.
func (h *PosHub) Publish(tenantID, outletID string, eventType string, payload map[string]interface{}) {
	if tenantID == "" || outletID == "" || eventType == "" {
		log.Printf("[pos_ws] refused broadcast missing tenant/outlet/event tenant_id=%q outlet_id=%q event=%q", tenantID, outletID, eventType)
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	event := PosEvent{
		Event:     PosEventType(eventType),
		Type:      PosEventType(eventType),
		Data:      payload,
		Payload:   payload,
		OutletID:  outletID,
		TenantID:  tenantID,
		Timestamp: apptime.Now().UTC().Format(time.RFC3339),
	}

	h.deliver(event)
	publishPosToRedis(event)
}

func (h *PosHub) deliver(event PosEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	targets := make([]*posSubscriber, 0)
	for _, sub := range h.subscribers {
		if sub.tenantID == event.TenantID && sub.outletID == event.OutletID {
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
