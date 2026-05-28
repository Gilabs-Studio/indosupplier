package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	infraRedis "github.com/gilabs/indosupplier/api/internal/core/infrastructure/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const idempotencyTTL = 24 * time.Hour

// idempotencyCacheEntry is the payload stored in Redis for a completed request.
type idempotencyCacheEntry struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
}

// GetOrSetIdempotency checks Redis for an existing response keyed by userID+key.
// If found it writes the cached response and returns true (caller should abort).
// If not found it returns false; the caller processes the request normally and
// must call SaveIdempotencyResult to persist the response.
func GetOrSetIdempotency(
	ctx context.Context,
	userID, key string,
) (*idempotencyCacheEntry, bool) {
	client := infraRedis.GetClient()
	if client == nil || key == "" {
		return nil, false
	}

	redisKey := fmt.Sprintf("idempotency:%s:%s", userID, key)
	raw, err := client.Get(ctx, redisKey).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		// Redis unavailable — fail open (process the request normally).
		return nil, false
	}

	var entry idempotencyCacheEntry
	if unmarshalErr := json.Unmarshal(raw, &entry); unmarshalErr != nil {
		return nil, false
	}
	return &entry, true
}

// SaveIdempotencyResult persists a completed response in Redis.
func SaveIdempotencyResult(
	ctx context.Context,
	userID, key string,
	statusCode int,
	body json.RawMessage,
) {
	client := infraRedis.GetClient()
	if client == nil || key == "" {
		return
	}

	entry := idempotencyCacheEntry{
		StatusCode: statusCode,
		Body:       body,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	redisKey := fmt.Sprintf("idempotency:%s:%s", userID, key)
	// Ignore Redis errors here — failing to cache is not fatal.
	_ = client.Set(ctx, redisKey, data, idempotencyTTL).Err()
}

// idempotentResponseWriter captures the response body so it can be cached.
type idempotentResponseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *idempotentResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *idempotentResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// IdempotentRequest is a Gin middleware that enforces idempotency for mutating
// requests. It reads the X-Idempotency-Key header and, if present:
//   - Returns the cached response immediately for duplicate requests.
//   - Stores the successful response in Redis after the first execution.
//
// Only 2xx responses are cached — errors are not idempotent.
func IdempotentRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		userID, _ := c.Get("user_id")
		uid, _ := userID.(string)
		if uid == "" {
			c.Next()
			return
		}

		// Return cached response if this key was already processed.
		if entry, found := GetOrSetIdempotency(c.Request.Context(), uid, key); found {
			c.Data(entry.StatusCode, "application/json; charset=utf-8", entry.Body)
			c.Abort()
			return
		}

		// Wrap the ResponseWriter to capture the response body.
		wrapper := &idempotentResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}
		c.Writer = wrapper

		c.Next()

		// Persist successful responses only.
		if wrapper.statusCode >= 200 && wrapper.statusCode < 300 && len(wrapper.body) > 0 {
			SaveIdempotencyResult(
				c.Request.Context(),
				uid,
				key,
				wrapper.statusCode,
				json.RawMessage(wrapper.body),
			)
		}
	}
}
