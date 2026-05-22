package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	once   sync.Once
)

// InitRedis initializes the Redis client
func InitRedis(cfg *config.Config) error {
	var err error
	once.Do(func() {
		addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
		Client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
				DialTimeout:  time.Duration(cfg.Redis.DialTimeoutSec) * time.Second,
				ReadTimeout:  time.Duration(cfg.Redis.ReadTimeoutSec) * time.Second,
				WriteTimeout: time.Duration(cfg.Redis.WriteTimeoutSec) * time.Second,
				PoolSize:     cfg.Redis.PoolSize,
				MinIdleConns: cfg.Redis.MinIdleConns,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = Client.Ping(ctx).Err(); err != nil {
			log.Printf("Failed to connect to Redis at %s: %v", addr, err)
			return
		}
		log.Printf("Connected to Redis at %s", addr)
	})
	return err
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	return Client
}

// Close closes the Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
