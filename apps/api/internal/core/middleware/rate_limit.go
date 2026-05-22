package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	infraRedis "github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const redisFixedWindowLua = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
	redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1]))
end
local ttl = redis.call('TTL', KEYS[1])
return {current, ttl}
`

// rateLimiter stores rate limiters per IP address for in-memory fallback
type rateLimiter struct {
	limiter        *rate.Limiter
	lastSeen       time.Time
	firstLimitTime *time.Time 
	window         int        
}

// rateLimiters is a map of IP addresses to their rate limiters
type rateLimiters struct {
	mu       sync.RWMutex
	limiters map[string]*rateLimiter
	cleanup  *time.Ticker
}

var (
	// In-memory fallbacks
	memoryLimiters = make(map[string]*rateLimiters)
	memoryMu       sync.Mutex
)

func getMemoryLimiter(typeKey string) *rateLimiters {
	memoryMu.Lock()
	defer memoryMu.Unlock()
	
	if memoryLimiters[typeKey] == nil {
		rl := &rateLimiters{
			limiters: make(map[string]*rateLimiter),
			cleanup:  time.NewTicker(5 * time.Minute),
		}
		go func() {
			for range rl.cleanup.C {
				rl.cleanupLimiters()
			}
		}()
		memoryLimiters[typeKey] = rl
	}
	return memoryLimiters[typeKey]
}

// cleanupLimiters removes limiters that haven't been used in the last hour
func (rl *rateLimiters) cleanupLimiters() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := apptime.Now()
	for ip, limiter := range rl.limiters {
		if now.Sub(limiter.lastSeen) > 1*time.Hour {
			delete(rl.limiters, ip)
		}
	}
}

// getLimiter returns a rate limiter for the given key (IP, email, etc.)
func (rl *rateLimiters) getLimiter(key string, requests int, window int) (*rate.Limiter, *rateLimiter) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		interval := time.Duration(window) * time.Second / time.Duration(requests)
		limiter = &rateLimiter{
			limiter:        rate.NewLimiter(rate.Every(interval), requests),
			lastSeen:       apptime.Now(),
			firstLimitTime: nil,
			window:         window,
		}
		rl.limiters[key] = limiter
	} else {
		limiter.lastSeen = apptime.Now()
		limiter.window = window
	}

	return limiter.limiter, limiter
}

// Redis logic using Fixed Window
func checkRedisRateLimit(ctx context.Context, client *redis.Client, key string, limit int, window int) (allowed bool, remaining int, reset int64, err error) {
	// Key: rate_limit:{key}
	// Value: count
	// TTL: window

	res, err := client.Eval(ctx, redisFixedWindowLua, []string{key}, window).Result()
	if err != nil {
		return false, 0, 0, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, 0, 0, fmt.Errorf("unexpected redis eval result")
	}

	count, _ := arr[0].(int64)
	ttlSec, _ := arr[1].(int64)
	if ttlSec <= 0 {
		// If TTL is missing (e.g., -1) or key vanished (-2), treat as full window.
		ttlSec = int64(window)
	}
	reset = apptime.Now().Add(time.Duration(ttlSec) * time.Second).Unix()

	remaining = limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return int(count) <= limit, remaining, reset, nil
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limitType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass rate limiting in development and when explicitly disabled via env var.
		// This allows local development to avoid hitting 429s while keeping production protected.
		if config.AppConfig != nil {
			env := config.AppConfig.Server.Env
			if env != "production" {
				if env == "development" || os.Getenv("DISABLE_RATE_LIMIT") == "true" {
					c.Next()
					return
				}
			}
		}
		ip := c.ClientIP()
		var rule config.RateLimitRule
		var key string

		// Determine rule and key
		switch limitType {
		case "login":
			// Level 1: IP-based
			rule = config.AppConfig.RateLimit.Login
			key = fmt.Sprintf("rate_limit:login:ip:%s", ip)
			
			// Check Level 3: Global Login
			globalRule := config.AppConfig.RateLimit.LoginGlobal
			globalKey := "rate_limit:login:global"
			if !checkLimit(c, globalKey, globalRule, "login_global") {
				return
			}
			
			// Check Level 2: Email (if present)
			var loginReq struct {
				Email string `json:"email"`
			}
			// Limit how much we read to avoid memory DoS.
			// BodySizeLimitMiddleware should already enforce a global cap, but we harden here too.
			const maxLoginInspectBytes = 64 << 10 // 64KB
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLoginInspectBytes+1))
			if int64(len(bodyBytes)) > maxLoginInspectBytes {
				errors.ErrorResponse(c, "REQUEST_BODY_TOO_LARGE", nil, nil)
				c.Abort()
				return
			}
			// Always restore body for downstream handlers.
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if err != nil {
				// If we couldn't read/parse properly, continue without email-level rate limit.
				break
			}
			if json.Unmarshal(bodyBytes, &loginReq) == nil && loginReq.Email != "" {
				emailRule := config.AppConfig.RateLimit.LoginByEmail
				emailKey := fmt.Sprintf("rate_limit:login:email:%s", loginReq.Email)
				if !checkLimit(c, emailKey, emailRule, "login_email") {
					return
				}
			}
			
		case "refresh":
			rule = config.AppConfig.RateLimit.Refresh
			key = fmt.Sprintf("rate_limit:refresh:%s", ip)
		case "upload":
			rule = config.AppConfig.RateLimit.Upload
			key = fmt.Sprintf("rate_limit:upload:%s", ip)
		case "public":
			rule = config.AppConfig.RateLimit.Public
			key = fmt.Sprintf("rate_limit:public:%s", ip)
		default: // general
			rule = config.AppConfig.RateLimit.General
			key = fmt.Sprintf("rate_limit:general:%s", ip)
		}

		// Perform the check for the primary rule
		if !checkLimit(c, key, rule, limitType) {
			return
		}

		c.Next()
	}
}

// checkLimit handles the actual check using Redis or Memory
func checkLimit(c *gin.Context, key string, rule config.RateLimitRule, typeKey string) bool {
	redisClient := infraRedis.GetClient()
	
	if redisClient != nil {
		allowed, remaining, reset, err := checkRedisRateLimit(c.Request.Context(), redisClient, key, rule.Requests, rule.Window)
		if err == nil {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Requests))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
			
			if !allowed {
				errors.ErrorResponse(c, "RATE_LIMIT_EXCEEDED", nil, nil)
				c.Abort()
				return false
			}
			return true
		}
		// Redis failed: for production hardening, fail closed to avoid multi-instance bypass.
		fmt.Printf("Redis rate limit error: %v\n", err)
		if config.AppConfig != nil && config.AppConfig.RateLimit.FailClosedOnRedisError {
			c.Header("Retry-After", "1")
			errors.ErrorResponse(c, "SERVICE_UNAVAILABLE", map[string]interface{}{
				"reason": "rate_limiter_backend_unavailable",
			}, nil)
			c.Abort()
			return false
		}

		// Non-production or explicit fail-open mode: fallback to in-memory limiter.
		fmt.Printf("Falling back to in-memory rate limiter for key=%s\n", key)
	}

	// In-Memory Fallback
	memConfig := getMemoryLimiter(typeKey)
	// We use the "key" from Redis logic as the identifier for memory map too (safe enough)
	limiter, limiterStruct := memConfig.getLimiter(key, rule.Requests, rule.Window)
	
	if !limiter.Allow() {
		// Calculate reset time
		var resetTime int64
		memConfig.mu.Lock()
		if limiterStruct.firstLimitTime == nil {
			now := apptime.Now()
			limiterStruct.firstLimitTime = &now
		}
		resetTime = limiterStruct.firstLimitTime.Add(time.Duration(rule.Window) * time.Second).Unix()
		memConfig.mu.Unlock()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Requests))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
		errors.ErrorResponse(c, "RATE_LIMIT_EXCEEDED", nil, nil)
		c.Abort()
		return false
	}
	
	// Reset FirstLimitTime if allowed
	memConfig.mu.Lock()
	limiterStruct.firstLimitTime = nil
	memConfig.mu.Unlock()

	// Remaining estimation for Token Bucket
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Requests))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rule.Requests-1)) // rough estimate
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", apptime.Now().Add(time.Duration(rule.Window)*time.Second).Unix()))
	
	return true
}
