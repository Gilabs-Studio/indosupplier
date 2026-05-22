package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	"github.com/gin-gonic/gin"
)

type metricsStore struct {
	totalRequests   uint64
	totalErrors     uint64
	totalDurationMs uint64
	byStatus        sync.Map // int -> *uint64
}

var metrics = &metricsStore{}
var metricsStartedAt = time.Now()

type MetricsSnapshotData struct {
	TotalRequests    uint64         `json:"total_requests"`
	TotalErrors5xx   uint64         `json:"total_errors_5xx"`
	TotalDurationMs  uint64         `json:"total_duration_ms"`
	AvgDurationMs    float64        `json:"avg_duration_ms"`
	RequestsByStatus map[int]uint64 `json:"requests_by_statuscode"`
	UptimeSeconds    int64          `json:"uptime_seconds"`
}

func MetricsSnapshot() MetricsSnapshotData {
	byStatus := make(map[int]uint64)
	metrics.byStatus.Range(func(k, v any) bool {
		status, ok := k.(int)
		if !ok {
			return true
		}
		byStatus[status] = atomic.LoadUint64(v.(*uint64))
		return true
	})

	total := atomic.LoadUint64(&metrics.totalRequests)
	durationMs := atomic.LoadUint64(&metrics.totalDurationMs)
	avg := float64(0)
	if total > 0 {
		avg = float64(durationMs) / float64(total)
	}

	return MetricsSnapshotData{
		TotalRequests:    total,
		TotalErrors5xx:   atomic.LoadUint64(&metrics.totalErrors),
		TotalDurationMs:  durationMs,
		AvgDurationMs:    avg,
		RequestsByStatus: byStatus,
		UptimeSeconds:    int64(time.Since(metricsStartedAt).Seconds()),
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := apptime.Now()
		c.Next()

		durMs := uint64(time.Since(start).Milliseconds())
		atomic.AddUint64(&metrics.totalRequests, 1)
		atomic.AddUint64(&metrics.totalDurationMs, durMs)
		status := c.Writer.Status()
		if status >= 500 {
			atomic.AddUint64(&metrics.totalErrors, 1)
		}

		ptr, _ := metrics.byStatus.LoadOrStore(status, new(uint64))
		atomic.AddUint64(ptr.(*uint64), 1)
	}
}

func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Token gate
		token := config.AppConfig.Observability.MetricsToken
		if token == "" || c.GetHeader("X-Internal-Token") != token {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, MetricsSnapshot())
	}
}
