package ws

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/redis"
)

const (
	// locationThrottleTTL is the minimum interval between location broadcasts per employee.
	// Setting this to 5 seconds keeps WebSocket traffic and Redis write volume low while
	// still providing near-real-time tracking — a good balance for SaaS cost efficiency.
	locationThrottleTTL = 5 * time.Second
)

// locationThrottleKey returns the Redis key used to gate broadcasts for a specific employee.
func locationThrottleKey(employeeID string) string {
	return fmt.Sprintf("travel:loc:throttle:%s", employeeID)
}

// ShouldThrottleLocation returns true when a location update for the given employee
// should be suppressed because one was already broadcast within the throttle window.
//
// Implementation uses Redis SET NX EX (atomic set-if-not-exists with TTL):
//   - Returns false (allow) on first call within the window → sets the key
//   - Returns true  (suppress) on subsequent calls until the key expires
//   - Falls back to false (allow all) when Redis is unavailable so location
//     tracking always works in degraded environments.
func ShouldThrottleLocation(employeeID string) bool {
	client := redis.GetClient()
	if client == nil {
		return false // no Redis → allow all updates
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	key := locationThrottleKey(employeeID)
	// SET NX returns true (key was set) when the slot was free.
	set, err := client.SetNX(ctx, key, 1, locationThrottleTTL).Result()
	if err != nil {
		return false // Redis error → allow the update rather than silently dropping
	}

	return !set // key already existed → throttle
}
