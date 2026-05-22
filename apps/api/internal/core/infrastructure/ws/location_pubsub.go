package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	// redisPubSubChannel is the global Redis Pub/Sub channel for real-time location events.
	// All hub instances across server pods subscribe and publish here, enabling
	// horizontal scaling without sticky sessions.
	redisPubSubChannel = "travel:locations"
)

// pubSubEnvelope wraps a broadcast payload with the originating instance ID so
// each hub can discard events it published itself (avoiding echo-loops).
type pubSubEnvelope struct {
	Origin    string                 `json:"origin"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}

// LocationPubSub bridges the local LocationHub with Redis Pub/Sub to enable
// multi-instance deployments. It holds one subscription goroutine per hub.
type LocationPubSub struct {
	instanceID string
	hub        *LocationHub
	cancel     context.CancelFunc
	once       sync.Once
}

var defaultPubSub *LocationPubSub

// InitLocationPubSub starts the Redis Pub/Sub bridge for the default hub.
// Call this once during application startup AFTER Redis is initialised.
// If Redis is not configured, this is a no-op — the in-process hub continues to work.
func InitLocationPubSub() {
	client := redis.GetClient()
	if client == nil {
		log.Println("[location_pubsub] Redis client unavailable – running in single-instance mode")
		return
	}

	ps := &LocationPubSub{
		instanceID: uuid.New().String(),
		hub:        defaultLocationHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	defaultPubSub = ps

	go ps.subscribe(ctx, client)
	log.Printf("[location_pubsub] started (instance=%s channel=%s)", ps.instanceID, redisPubSubChannel)
}

// StopLocationPubSub gracefully shuts down the subscription goroutine.
func StopLocationPubSub() {
	if defaultPubSub != nil {
		defaultPubSub.once.Do(func() {
			defaultPubSub.cancel()
		})
	}
}

// publishToRedis sends a location or navigation event to the Redis channel so
// other hub instances can fan-out the event to their local WebSocket clients.
// Errors are logged but never returned — local delivery always continues.
func publishToRedis(eventType string, data map[string]interface{}) {
	if defaultPubSub == nil {
		return // single-instance mode, nothing to do
	}

	client := redis.GetClient()
	if client == nil {
		return
	}

	env := pubSubEnvelope{
		Origin:    defaultPubSub.instanceID,
		EventType: eventType,
		Data:      data,
	}

	payload, err := json.Marshal(env)
	if err != nil {
		log.Printf("[location_pubsub] marshal error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Publish(ctx, redisPubSubChannel, payload).Err(); err != nil {
		log.Printf("[location_pubsub] publish error: %v", err)
	}
}

// subscribe blocks and processes Redis messages until ctx is cancelled.
func (ps *LocationPubSub) subscribe(ctx context.Context, client *goredis.Client) {
	sub := client.Subscribe(ctx, redisPubSubChannel)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			ps.handleMessage(msg.Payload)
		}
	}
}

// handleMessage deserialises a Redis envelope and fans out to local WS clients,
// skipping events that originated from this instance (echo prevention).
func (ps *LocationPubSub) handleMessage(payload string) {
	var env pubSubEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		return
	}

	// Discard our own events — we already delivered them locally.
	if env.Origin == ps.instanceID {
		return
	}

	rawPayload, err := json.Marshal(map[string]interface{}{
		"type": env.EventType,
		"data": env.Data,
	})
	if err != nil {
		return
	}

	// Fan-out to local WS clients whose subscription filters match the employee.
	employeeID, _ := env.Data["employee_id"].(string)
	lat, _ := env.Data["lat"].(float64)
	lng, _ := env.Data["lng"].(float64)

	ps.hub.mu.RLock()
	subs := make([]*subscriber, 0, len(ps.hub.subscribers))
	for _, sub := range ps.hub.subscribers {
		subs = append(subs, sub)
	}
	ps.hub.mu.RUnlock()

	for _, sub := range subs {
		if !matchesFilter(sub.filter, employeeID, lat, lng) {
			continue
		}
		if err := sub.write(rawPayload); err != nil {
			ps.hub.Unregister(sub.id)
		}
	}
}
