# API Response Standards

## Backend API Response Guidelines

**Version**: 1.0  
**Status**: Active

---

## Overview

All API responses follow a consistent standard format to ensure:
- **Consistency**: Same format across all endpoints
- **Predictability**: Developers know what to expect
- **Error Handling**: Clear and actionable error handling
- **Type Safety**: Clear structure for frontend typing

---

## Base Response Structure

### Standard Success Response

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "timestamp": "2024-01-15T10:30:45+07:00",
  "request_id": "req_abc123xyz"
}
```

### Standard Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {},
    "field_errors": []
  },
  "meta": {},
  "timestamp": "2024-01-15T10:30:45+07:00",
  "request_id": "req_abc123xyz"
}
```

---

## Success Response Patterns

### Single Resource Response

**GET /api/v1/{resource}/{id}**

- Return single object in `data`
- Include full resource details
- Include related resources if needed

### Collection Response (with Pagination)

**GET /api/v1/{resource}**

- Return array in `data`
- **CRITICAL**: Always include pagination meta
- Default `per_page`: 20, maximum: 100
- Include filters and sort info in meta

### Create Resource Response

**POST /api/v1/{resource}**

- Return created resource in `data`
- HTTP status: 201 Created
- Include `created_by` in meta if applicable

### Update Resource Response

**PUT/PATCH /api/v1/{resource}/{id}**

- Return updated resource in `data`
- Include `updated_by` in meta
- Optional: include `changes` in meta for audit

### Delete Resource Response

**DELETE /api/v1/{resource}/{id}**

- Return minimal info (id, deleted_at)
- HTTP status: 204 No Content or 200 OK
- Include `deleted_by` in meta

### Action Response (Non-CRUD)

**POST /api/v1/{resource}/{id}/{action}**

- Return action result in `data`
- Include relevant metadata in meta
- Include status and action timestamp

---

## Error Response

### Standard Error Structure

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      "additional_info": "value"
    },
    "field_errors": [
      {
        "field": "field_name",
        "code": "ERROR_CODE",
        "message": "Field-specific error message"
      }
    ]
  }
}
```

### Error Categories

- **Validation Errors (400)**: Invalid request data, missing fields, wrong types
- **Authentication Errors (401)**: Invalid/expired token, missing credentials
- **Authorization Errors (403)**: Insufficient permissions
- **Not Found Errors (404)**: Resource not found
- **Conflict Errors (409)**: Duplicate values, resource in use
- **Business Logic Errors (422)**: Business rule violations
- **Rate Limit Errors (429)**: Too many requests
- **Server Errors (500)**: Internal server errors
- **Service Unavailable (503)**: Maintenance mode, service down

**CRITICAL**: All error codes must follow standards in `api-error-codes.md`

---

## Pagination

### Pagination Meta Structure

```json
{
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false,
    "next_page": 2,
    "prev_page": null
  }
}
```

### Query Parameters

- `page`: Requested page (default: 1)
- `per_page`: Items per page (default: 20, **max: 100**)
- **CRITICAL**: Enforce maximum `per_page` = 100 to prevent memory issues

### Cursor-based Pagination (Alternative)

For large datasets, use cursor-based pagination with `cursor` and `limit` parameters.

---

## Date & Time Format

### Standard Format

- **Format**: ISO 8601 with timezone
- **Timezone**: WIB (UTC+7) for Indonesia
- **Format String**: `2006-01-02T15:04:05+07:00` (Go time format)
- **Example**: `2024-01-15T10:30:45+07:00`

### Timestamp Fields

All resources have at minimum:
- `created_at`: Creation time (required)
- `updated_at`: Update time (required)
- `deleted_at`: Deletion time (optional, for soft delete)

### Timezone Handling

- All datetime stored in database as UTC
- Response always in WIB format (UTC+7)
- Frontend responsible for converting to user's local timezone if needed

---

## Data Types & Validation

### Currency (Money)

- **Type**: Integer (in smallest unit, e.g., cents for currency)
- **Display**: Format with thousand separator
- **Max Value**: According to business requirements

### Percentage

- **Type**: Float (0-100)
- **Precision**: 2 decimal places

### Boolean

- **Type**: Boolean
- **Values**: `true` or `false`

### Enum/Status

- **Type**: String
- **Values**: Predefined constants
- Document all possible values

### ID Format

- **Format**: `{resource_type}_{random_string}`
- Use consistent prefix for each resource type

### Phone Number

- **Format**: String, without spaces and dashes
- **Validation**: According to country format

### Email

- **Format**: Standard email format
- **Validation**: RFC 5322 compliant

### URL

- **Format**: Full URL with protocol
- **Example**: `https://cdn.example.com/path/to/file.jpg`

### Nested Resource

- **Reference**: Minimal fields (id, name) for list responses
- **Full**: Complete object for detail responses
- Use query parameter `?include=relation1,relation2` for control

---

## HTTP Status Codes

### Success Codes

- **200 OK**: Request successful (GET, PUT, PATCH)
- **201 Created**: Resource created successfully (POST)
- **204 No Content**: Request successful, no content (DELETE)

### Client Error Codes

- **400 Bad Request**: Invalid request (validation error)
- **401 Unauthorized**: Not authenticated
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Conflict with current state
- **422 Unprocessable Entity**: Valid request but cannot be processed
- **429 Too Many Requests**: Rate limit exceeded

### Server Error Codes

- **500 Internal Server Error**: Server error
- **502 Bad Gateway**: Error from upstream service
- **503 Service Unavailable**: Service under maintenance
- **504 Gateway Timeout**: Timeout from upstream service

---

## Best Practices

### 1. Consistent Field Naming

- Use **snake_case** for all fields
- Use descriptive and clear names
- Avoid unclear abbreviations

### 2. Always Include Metadata

- `timestamp`: Response creation time
- `request_id`: Unique ID for request tracking
- `tenant_id`: Tenant ID (for multi-tenant)
- `outlet_id`: Outlet ID (if applicable)

### 3. Error Messages

- Use user-friendly language for user-facing messages
- Include error code for programmatic handling
- Provide sufficient context for debugging
- Do not expose sensitive information in production

### 4. Pagination

- Default `per_page`: 20
- Maximum `per_page`: 100
- **CRITICAL**: DO NOT use `per_page` > 100 in production
- Always include pagination meta for collection responses
- For large datasets, use cursor-based pagination
- Use database aggregation (COUNT, SUM, GROUP BY) instead of fetching all data

### 5. Date/Time

- Always use ISO 8601 format
- Always include timezone (WIB/UTC+7)
- Store in database as UTC, convert to WIB in response

### 6. Currency

- Store as integer (smallest unit)
- Include formatted version if needed
- Use format according to locale

### 7. Null vs Empty

- Use `null` for fields that don't exist
- Use empty array `[]` for empty collections
- Use empty string `""` only if field is actually an empty string

### 8. Nested Resources

- Use reference (id + minimal fields) for list responses
- Use full object for detail responses
- Include query parameter `?include=relation1,relation2` for control
- **CRITICAL**: Use `Preload()` for eager loading relationships
- **CRITICAL**: Avoid N+1 queries with batch loading

### 9. Versioning

- Use URL versioning: `/api/v1/`
- Maintain backward compatibility within major version
- Document breaking changes

### 10. Request ID

- Generate unique request ID for each request
- Include in response for correlation
- Log request ID in all logs for tracing

---

## Enterprise-Scale Performance (CRITICAL)

### Database Query Optimization

**❌ DO NOT:**
- Fetch all data to memory then process in application layer
- Query without pagination limit
- N+1 queries for relationships

**✅ DO:**
- Use database aggregation (COUNT, SUM, GROUP BY)
- Batch load relationships with IN queries
- Use Preload() for eager loading
- Enforce pagination limits (max per_page = 100)

### Connection Pooling

**✅ REQUIRED CONFIGURATION:**
- Set max open connections according to server capacity
- Set max idle connections
- Configure connection lifetime and idle timeout

### Query Timeout

**✅ REQUIRED IMPLEMENTATION:**
- All queries use context with timeout (30 seconds)
- Handle timeout errors gracefully
- Return appropriate error code (QUERY_TIMEOUT)

### Caching Strategy

**✅ IMPLEMENT CACHING:**
- Dashboard data: TTL 1-5 minutes
- User lists: TTL 5-15 minutes
- Reference data: TTL 1 hour
- Report data: TTL 5-15 minutes
- Use cache-aside pattern
- Implement cache invalidation

### Pagination Limits

**✅ ENFORCE LIMITS:**
- Default: 20
- Maximum: 100
- **CRITICAL**: NEVER allow per_page > 100

### Large Dataset Handling

**✅ FOR LARGE DATASETS:**
- Use database aggregation (SUM, COUNT, GROUP BY)
- Use materialized views for dashboards
- Use cursor-based pagination
- Stream processing for very large datasets
- DO NOT fetch all data to memory

---

## Implementation Notes (Go Backend)

### Response Struct Pattern

```go
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Meta      Meta        `json:"meta,omitempty"`
    Timestamp string      `json:"timestamp"`
    RequestID string      `json:"request_id"`
}
```

### Helper Functions Pattern

- `SuccessResponse()`: For success responses
- `SuccessResponseCreated()`: For created resources (201)
- `ErrorResponse()`: For error responses
- `ValidationErrorResponse()`: For validation errors
- `NotFoundResponse()`: For not found errors
- `NewPaginationMeta()`: For pagination metadata

### Performance Best Practices

- Use context with timeout for all queries
- Enforce pagination limits in ParsePaginationParams()
- Use batch loading to avoid N+1 queries
- Use database aggregation for summaries
- Configure connection pool correctly

---

**This document will be updated according to API development and feedback from the development team.**