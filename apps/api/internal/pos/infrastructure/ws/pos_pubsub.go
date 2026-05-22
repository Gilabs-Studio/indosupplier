package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type posPubSubEnvelope struct {
	Origin string   `json:"origin"`
	Event  PosEvent `json:"event"`
}

type PosPubSub struct {
	instanceID string
	hub        *PosHub
	cancel     context.CancelFunc
	once       sync.Once
}

var defaultPosPubSub *PosPubSub

func posChannelName(tenantID, outletID string) string {
	return fmt.Sprintf("pos:%s:%s", tenantID, outletID)
}

// InitPosPubSub starts the Redis Pub/Sub bridge for POS realtime events.
func InitPosPubSub() {
	client := redis.GetClient()
	if client == nil {
		log.Println("[pos_pubsub] Redis client unavailable - running in single-instance mode")
		return
	}

	ps := &PosPubSub{
		instanceID: uuid.New().String(),
		hub:        defaultPosHub,
	}
	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	defaultPosPubSub = ps

	go ps.subscribe(ctx, client)
	log.Printf("[pos_pubsub] started instance=%s pattern=pos:*", ps.instanceID)
}

// StopPosPubSub gracefully shuts down the POS Redis subscription goroutine.
func StopPosPubSub() {
	if defaultPosPubSub != nil {
		defaultPosPubSub.once.Do(func() {
			defaultPosPubSub.cancel()
		})
	}
}

func publishPosToRedis(event PosEvent) {
	if defaultPosPubSub == nil {
		return
	}
	client := redis.GetClient()
	if client == nil {
		return
	}

	env := posPubSubEnvelope{
		Origin: defaultPosPubSub.instanceID,
		Event:  event,
	}
	payload, err := json.Marshal(env)
	if err != nil {
		log.Printf("[pos_pubsub] marshal error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Publish(ctx, posChannelName(event.TenantID, event.OutletID), payload).Err(); err != nil {
		log.Printf("[pos_pubsub] publish error: %v", err)
	}
}

func (ps *PosPubSub) subscribe(ctx context.Context, client *goredis.Client) {
	sub := client.PSubscribe(ctx, "pos:*")
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

func (ps *PosPubSub) handleMessage(payload string) {
	var env posPubSubEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		return
	}
	if env.Origin == ps.instanceID {
		return
	}
	if env.Event.TenantID == "" || env.Event.OutletID == "" || env.Event.Event == "" {
		return
	}
	ps.hub.deliver(env.Event)
}
