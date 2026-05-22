# API Performance Standards

## Enterprise-Level Performance Guidelines

**Version**: 1.0  
**Status**: Active

---

## Overview

This document defines performance standards and requirements for enterprise-level API development. All endpoints must meet these standards to ensure scalability, reliability, and optimal user experience.

**Target Metrics**:
- Response time: < 200ms for 95th percentile
- Throughput: Handle 1000+ requests per second per instance
- Database query time: < 100ms for 95th percentile
- Memory usage: Efficient, no memory leaks
- Connection pool: Properly configured and monitored

---

## Database Performance

### Query Optimization

**❌ DO NOT:**
- Fetch all records then filter/process in application layer
- Use `SELECT *` without specific columns
- Query without indexes on WHERE/JOIN columns
- Use N+1 queries for relationships
- Process large datasets in memory

**✅ DO:**
- Use database aggregation (COUNT, SUM, GROUP BY, AVG)
- Add indexes on frequently queried columns
- Use `SELECT` with specific columns only
- Use batch loading with IN queries for relationships
- Use `Preload()` for eager loading
- Use database-level filtering and sorting

### Index Strategy

**Search Optimization (Prefix Search + GIN)**:
For text search columns (e.g., name, email), use GIN indexes with `pg_trgm` for high performance.

**✅ REQUIRED IMPLEMENTATION**:
1. Migration:
```go
DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
DB.Exec("CREATE INDEX idx_users_name_gin ON users USING gin (name gin_trgm_ops)")
```
2. Repository (Prefix Search):
```go
// Use prefix search (text%) instead of wildcard (%text%) for index usage
search := req.Search + "%"
db.Where("name ILIKE ?", search)
```

### Index Strategy

**Required Indexes**:
- Primary keys (automatic)
- Foreign keys
- Frequently queried columns in WHERE clauses
- Columns used in JOIN operations
- Columns used in ORDER BY
- Composite indexes for multi-column queries

**Index Best Practices**:
- Monitor index usage and remove unused indexes
- Avoid over-indexing (slows down writes)
- Use partial indexes for filtered queries
- Consider covering indexes for read-heavy queries

### Connection Pooling

**✅ REQUIRED CONFIGURATION:**

```go
sqlDB.SetMaxOpenConns(100)        // Max open connections
sqlDB.SetMaxIdleConns(25)         // Max idle connections
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Idle timeout
```

**Connection Pool Sizing**:
- Max open connections: Based on server capacity (typically 50-200)
- Max idle connections: 25-50% of max open connections
- Monitor connection pool usage and adjust as needed

### Query Timeout

**✅ REQUIRED IMPLEMENTATION:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := db.WithContext(ctx).Where(...).Find(&results).Error
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return errors.New("QUERY_TIMEOUT")
    }
    return err
}
```

**Timeout Guidelines**:
- Simple queries: 5-10 seconds
- Complex queries: 15-30 seconds
- Report/aggregation queries: 30-60 seconds
- Always handle timeout errors gracefully

### Pagination Limits

**✅ ENFORCE STRICT LIMITS:**

```go
perPage := req.PerPage
if perPage < 1 {
    perPage = 20
}
if perPage > 100 {
    perPage = 100  // CRITICAL: Maximum limit
}
// NEVER allow per_page > 100
```

**Pagination Best Practices**:
- Default: 20 items per page
- Maximum: 100 items per page
- Use cursor-based pagination for large datasets
- Always include pagination metadata in response

---

## Caching Strategy

### Cache Levels

**1. Application-Level Cache (In-Memory)**
- Use for frequently accessed, rarely changed data
- TTL: 1-5 minutes
- Examples: User roles, permissions, reference data

**2. Distributed Cache (Redis) - ✅ PREFERRED**
- Standard for shared cache across instances
- **Serialization**: JSON (standard) or MsgPack (compact)
- **Key Naming**: `resource:action:params` (e.g., `users:list:page:1`)
- TTL: 5-60 minutes
- Examples: Session data, API responses, computed results

**3. Database Query Cache**
- Use for expensive queries
- TTL: 1-15 minutes
- Examples: Dashboard summaries, report data

### Cache Patterns

**Cache-Aside Pattern**:
```go
// Check cache first
if cached, found := cache.Get(key); found {
    return cached
}

// If not found, query database
data, err := db.Query(...)
if err != nil {
    return err
}

// Store in cache
cache.Set(key, data, ttl)
return data
```

**Cache Invalidation**:
- Invalidate on write operations
- Use cache tags for related data invalidation
- Implement cache warming for critical data

### List Caching Pattern

**Problem**: List endpoints with strict filters (search, status, role) are expensive to compute repeatedly.
**Solution**: Cache the entire result set with a short TTL (5 mins) and invalidate on mutations.

**Key Generation**:
```go
key := fmt.Sprintf("resource:list:page:%d:search:%s:status:%s", page, search, status)
```

**Invalidation Strategy**:
On Create/Update/Delete, ensure **all** list variations are invalidated.
```go
// Invalidate all list caches for this resource
pattern := "resource:list:*"
iter := redis.Scan(ctx, 0, pattern, 0).Iterator()
for iter.Next(ctx) {
    redis.Del(ctx, iter.Val())
}
```

### Cache TTL Guidelines

- **Dashboard data**: 1-5 minutes
- **User lists**: 5-15 minutes
- **Reference data**: 1 hour
- **Report data**: 5-15 minutes
- **Session data**: 30 minutes - 24 hours
- **API responses**: 1-5 minutes (if applicable)

---

## Response Time Optimization

### Database Query Optimization

**Use Database Aggregation**:
```go
// ❌ BAD: Fetch all then count
var reports []Report
db.Find(&reports)
count := len(reports)

// ✅ GOOD: Database aggregation
var count int64
db.Model(&Report{}).Count(&count)
```

**Use Batch Loading**:
```go
// ❌ BAD: N+1 queries
for _, user := range users {
    target, _ := repo.GetTarget(user.ID)  // Query per user
}

// ✅ GOOD: Batch load
userIDs := extractIDs(users)
targets := repo.GetBatchTargets(userIDs)  // Single query
```

### Response Size Optimization

**Minimize Response Payload**:
- Return only required fields
- Use sparse fieldsets: `?fields=id,name,email`
- Compress large responses (gzip)
- Use pagination for collections
- Avoid deep nesting (max 3-4 levels)

### Lazy Loading

**Use Query Parameters for Related Data**:
- `?include=relation1,relation2` for optional relations
- Default: return minimal data
- Full data: only when explicitly requested

---

## Memory Management

### Memory Leak Prevention

**✅ DO:**
- Close database connections properly
- Release resources in defer statements
- Use connection pooling
- Limit result set sizes
- Clear large data structures after use

**❌ DO NOT:**
- Keep large datasets in memory
- Create unbounded slices/maps
- Store references to large objects
- Forget to close file handles/connections

### Memory Profiling

**Regular Monitoring**:
- Monitor memory usage per endpoint
- Profile memory leaks in development
- Set memory limits for containers
- Alert on high memory usage

---

## Rate Limiting

### Rate Limit Strategy

**Per-User Rate Limits**:
- Authentication endpoints: 5 requests/minute
- General API: 100 requests/minute
- File upload: 10 requests/minute
- Report generation: 5 requests/minute

**Per-IP Rate Limits**:
- Public endpoints: 200 requests/minute
- Protect against DDoS attacks

### Implementation

```go
// Use middleware for rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 100)
    return func(c *gin.Context) {
        if !limiter.Allow() {
            errors.RateLimitResponse(c, ...)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## Panic Attack Prevention & Runtime Safety

### Mandatory Safeguards

**✅ REQUIRED:**
- Global panic recovery middleware must be enabled for all routes.
- Request parameter bounds must be validated before repository/usecase execution.
- All handlers must handle malformed payloads without panic.
- Body and multipart size limits must be enforced globally.
- Background workers must use timeout-bound contexts.

**Example (Query Bounds):**
```go
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

if page < 1 {
    page = 1
}
if perPage < 1 {
    perPage = 20
}
if perPage > 100 {
    perPage = 100
}
```

---

## Retry Storm & Burst Mitigation

### Server-Side Rules

**✅ REQUIRED:**
- Use distributed rate limiting (Redis-backed) in multi-instance deployments.
- Define limiter backend failure policy explicitly:
  - Public/auth critical surfaces: prefer fail-closed.
  - Internal low-risk surfaces: fail-open only with strict observability.
- Return `Retry-After` for `429` and temporary overload scenarios.
- Protect expensive endpoints with tighter per-endpoint limits than general endpoints.

### Retry Policy Rules

**✅ REQUIRED:**
- Retries must use exponential backoff with jitter.
- Retries must have max attempts and retry budget.
- Do not retry non-idempotent writes without idempotency key.
- Idempotent mutation endpoints should support idempotency key to avoid duplicate effects.

### Client Coordination Rules

**✅ REQUIRED:**
- Client must not aggressively auto-retry on `429`/`503`.
- Respect `Retry-After` header if present.
- Use cache/stale data and controlled refetch to avoid thundering herd after reconnect/focus.

---

## Monitoring & Observability

### Metrics to Monitor

**Performance Metrics**:
- Response time (p50, p95, p99)
- Request rate (requests per second)
- Error rate (errors per second)
- Database query time
- Connection pool usage
- Cache hit rate

**Resource Metrics**:
- CPU usage
- Memory usage
- Database connections
- Network I/O
- Disk I/O

### Logging

**Structured Logging**:
```go
logger.Info("Request processed",
    zap.String("request_id", requestID),
    zap.String("endpoint", endpoint),
    zap.Duration("duration", duration),
    zap.Int("status", statusCode),
)
```

**Log Levels**:
- ERROR: Errors that need attention
- WARN: Warnings that might need attention
- INFO: Important business events
- DEBUG: Detailed debugging information (development only)

### Alerting

**Critical Alerts**:
- Response time > 1 second (p95)
- Error rate > 5%
- Database connection pool exhausted
- Memory usage > 80%
- CPU usage > 90%

---

## Load Testing

### Performance Testing Requirements

**Before Production Deployment**:
- Load test with expected traffic (2x peak load)
- Test database under load
- Test cache effectiveness
- Test rate limiting
- Test error handling under load

### Load Testing Tools

- **Recommended**: k6, Apache JMeter, Gatling
- Test scenarios: Normal load, peak load, stress test
- Monitor: Response time, error rate, resource usage

---

## Best Practices Checklist

### Database
- [ ] All queries use indexes
- [ ] No N+1 queries
- [ ] Pagination limits enforced (max 100)
- [ ] Query timeouts implemented (30s max)
- [ ] Connection pool configured
- [ ] Database aggregation used for summaries

### Caching
- [ ] Cache strategy defined for each endpoint
- [ ] Cache TTL appropriate for data type
- [ ] Cache invalidation implemented
- [ ] Cache hit rate monitored

### Response Time
- [ ] Response time < 200ms (p95)
- [ ] Database queries < 100ms (p95)
- [ ] Large responses paginated
- [ ] Response compression enabled

### Memory
- [ ] No memory leaks
- [ ] Large datasets not loaded to memory
- [ ] Resources properly released
- [ ] Memory usage monitored

### Rate Limiting
- [ ] Rate limits configured
- [ ] Rate limit errors handled gracefully
- [ ] Rate limit headers included in response

### Monitoring
- [ ] Performance metrics collected
- [ ] Error rates monitored
- [ ] Alerts configured
- [ ] Logging structured

---

**This document will be updated according to performance requirements and feedback from the development team.**
