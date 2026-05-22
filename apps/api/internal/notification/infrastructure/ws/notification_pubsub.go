package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type notificationPubSubEnvelope struct {
	Origin string            `json:"origin"`
	Event  NotificationEvent `json:"event"`
}

type NotificationPubSub struct {
	instanceID string
	hub        *NotificationHub
	cancel     context.CancelFunc
	once       sync.Once
}

var defaultNotificationPubSub *NotificationPubSub

func notificationChannelName(tenantID, userID string) string {
	tenant := tenantID
	if tenant == "" {
		tenant = "global"
	}
	return fmt.Sprintf("notifications:%s:%s", tenant, userID)
}

// InitNotificationPubSub starts the Redis Pub/Sub bridge for notification realtime events.
func InitNotificationPubSub() {
	client := redis.GetClient()
	if client == nil {
		return
	}

	ps := &NotificationPubSub{
		instanceID: uuid.New().String(),
		hub:        defaultNotificationHub,
	}
	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	defaultNotificationPubSub = ps

	go ps.subscribe(ctx, client)
}

// StopNotificationPubSub gracefully stops Redis subscription goroutine.
func StopNotificationPubSub() {
	if defaultNotificationPubSub != nil {
		defaultNotificationPubSub.once.Do(func() {
			defaultNotificationPubSub.cancel()
		})
	}
}

func publishNotificationToRedis(event NotificationEvent) {
	if defaultNotificationPubSub == nil {
		return
	}
	client := redis.GetClient()
	if client == nil {
		return
	}

	env := notificationPubSubEnvelope{
		Origin: defaultNotificationPubSub.instanceID,
		Event:  event,
	}
	payload, err := json.Marshal(env)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = client.Publish(ctx, notificationChannelName(event.TenantID, event.UserID), payload).Err()
}

func (ps *NotificationPubSub) subscribe(ctx context.Context, client *goredis.Client) {
	sub := client.PSubscribe(ctx, "notifications:*")
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

func (ps *NotificationPubSub) handleMessage(payload string) {
	var env notificationPubSubEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		return
	}
	if env.Origin == ps.instanceID {
		return
	}
	if env.Event.UserID == "" || env.Event.Type == "" {
		return
	}
	ps.hub.deliver(env.Event)
}
