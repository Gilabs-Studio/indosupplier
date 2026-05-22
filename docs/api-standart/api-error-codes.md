# API Error Codes Reference

## Backend API Error Code Guidelines

**Version**: 1.0  
**Status**: Active

---

## Overview

All error responses use consistent error codes to ensure:
- **Consistency**: Predictable error handling
- **Type Safety**: Frontend can handle errors in a type-safe manner
- **Debugging**: Error codes help with tracking and debugging
- **User Experience**: Clear and actionable error messages

---

## Error Response Format

### Single Error

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      "additional_info": "value"
    }
  }
}
```

### Multiple Field Errors

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "field_errors": [
      {
        "field": "field_name",
        "code": "ERROR_CODE",
        "message": "Field-specific error message",
        "constraint": {
          "min": 0
        }
      }
    ]
  }
}
```

---

## Error Code Categories

### Validation Errors (400 Bad Request)

**Pattern**: Field validation and request validation

**Common Codes**:
- `VALIDATION_ERROR`: General validation error
- `REQUIRED`: Required field missing
- `INVALID_TYPE`: Invalid data type
- `INVALID_FORMAT`: Invalid format (email, phone, URL, etc.)
- `INVALID_LENGTH`: Invalid length
- `MIN_VALUE`: Value below minimum
- `MAX_VALUE`: Value above maximum
- `INVALID_ENUM`: Value not in enum
- `INVALID_DATE`: Invalid date format
- `INVALID_TIME`: Invalid time format
- `DUPLICATE_VALUE`: Duplicate value

**Request Validation Codes**:
- `MISSING_REQUIRED_FIELD`: Required field missing
- `INVALID_REQUEST_BODY`: Invalid request body
- `INVALID_QUERY_PARAM`: Invalid query parameter
- `INVALID_PATH_PARAM`: Invalid path parameter
- `UNSUPPORTED_MEDIA_TYPE`: Unsupported Content-Type

---

### Authentication & Authorization (401/403)

**Authentication Codes (401)**:
- `UNAUTHORIZED`: Authentication token invalid or expired
- `TOKEN_EXPIRED`: Token expired
- `TOKEN_INVALID`: Token invalid
- `TOKEN_MISSING`: Token not found in header
- `INVALID_CREDENTIALS`: Email or password incorrect
- `ACCOUNT_DISABLED`: Account disabled
- `ACCOUNT_LOCKED`: Account locked (too many login attempts)
- `SESSION_EXPIRED`: Session expired
- `REFRESH_TOKEN_INVALID`: Refresh token invalid
- `REFRESH_TOKEN_EXPIRED`: Refresh token expired

**Authorization Codes (403)**:
- `FORBIDDEN`: Insufficient permissions to access this resource
- `PERMISSION_DENIED`: Insufficient permissions
- `ROLE_INSUFFICIENT`: Role does not have access
- `RESOURCE_OWNERSHIP_REQUIRED`: Only resource owner can access

---

### Resource Errors (404/409)

**Not Found Codes (404)**:
- `NOT_FOUND`: Resource not found
- `{RESOURCE}_NOT_FOUND`: Specific resource not found (e.g., `PRODUCT_NOT_FOUND`, `USER_NOT_FOUND`)

**Conflict Codes (409)**:
- `CONFLICT`: Conflict with current state
- `RESOURCE_ALREADY_EXISTS`: Resource already exists
- `RESOURCE_IN_USE`: Resource currently in use
- `CANNOT_DELETE`: Resource cannot be deleted (has dependencies)

---

### Business Logic Errors (422 Unprocessable Entity)

**Pattern**: Business rule violations that cannot be processed

**Common Categories**:
- Stock & Inventory: `INSUFFICIENT_STOCK`, `STOCK_NEGATIVE`, `STOCK_LOCKED`
- Transactions: `CART_EMPTY`, `SALE_ALREADY_COMPLETED`, `REFUND_AMOUNT_EXCEEDED`
- State Management: `SHIFT_NOT_OPEN`, `SHIFT_ALREADY_CLOSED`
- Business Rules: Custom error codes according to business logic

**CRITICAL**: Business logic errors must be clear and actionable for users

---

### Payment Errors (422/408)

**Common Codes**:
- `PAYMENT_FAILED`: Payment failed
- `PAYMENT_TIMEOUT`: Payment timeout
- `PAYMENT_CANCELLED`: Payment cancelled
- `PAYMENT_ALREADY_PROCESSED`: Payment already processed
- `PAYMENT_METHOD_NOT_AVAILABLE`: Payment method not available
- `INSUFFICIENT_BALANCE`: Insufficient balance
- `INVALID_PAYMENT_AMOUNT`: Invalid payment amount

---

### System Errors (500/503)

**Common Codes**:
- `INTERNAL_SERVER_ERROR`: Internal server error occurred
- `SERVICE_UNAVAILABLE`: Service under maintenance
- `DATABASE_ERROR`: Database error
- `CACHE_ERROR`: General cache error
- `REDIS_CONNECTION_FAILED`: Unable to connect to Redis
- `REDIS_COMMAND_FAILED`: Redis command execution failed
- `LOCK_ACQUISITION_FAILED`: Failed to acquire distributed lock
- `STORAGE_ERROR`: Storage error
- `QUEUE_ERROR`: Message queue error
- `TIMEOUT`: Request timeout
- `QUERY_TIMEOUT`: Database query timeout
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `MAINTENANCE_MODE`: System under maintenance
- `CONNECTION_POOL_EXHAUSTED`: Database connection pool exhausted
- `TOO_MANY_RESULTS`: Requested too many results (max per_page exceeded)

---

### Integration Errors (502/503)

**Pattern**: Errors from external services

**Common Categories**:
- Payment Gateway: `PAYMENT_GATEWAY_ERROR`, `PAYMENT_GATEWAY_TIMEOUT`
- Third-party APIs: `{SERVICE}_ERROR`, `{SERVICE}_UNAVAILABLE`
- Hardware: `PRINTER_ERROR`, `SCANNER_ERROR`

---

## Error Code Naming Conventions

### Format

- **UPPERCASE**: All error codes in UPPERCASE
- **SNAKE_CASE**: Use underscore as separator
- **Descriptive**: Name must be clear and descriptive

### Patterns

- **Resource Not Found**: `{RESOURCE}_NOT_FOUND` (e.g., `PRODUCT_NOT_FOUND`)
- **Resource Already Exists**: `{RESOURCE}_ALREADY_EXISTS`
- **Invalid {Thing}**: `INVALID_{THING}` (e.g., `INVALID_FORMAT`)
- **{Action} Failed**: `{ACTION}_FAILED` (e.g., `PAYMENT_FAILED`)
- **{Action} Already {State}**: `{ACTION}_ALREADY_{STATE}` (e.g., `SHIFT_ALREADY_OPEN`)

---

## Error Code Mapping (Go Implementation)

### Pattern

```go
var ErrorCodeMap = map[string]ErrorInfo{
    "ERROR_CODE": {
        HTTPStatus: http.StatusXXX,
        Message:    "Default error message",
    },
}
```

### ErrorInfo Struct

```go
type ErrorInfo struct {
    HTTPStatus int
    Message    string
}
```

---

## Best Practices

### 1. Error Code Consistency

- Use predefined error codes
- Do not create new error codes without documentation
- Update documentation when adding new error codes

### 2. Error Messages

- Use user-friendly language for user-facing messages
- Include error code for programmatic handling
- Provide sufficient context in `details`
- Do not expose sensitive information

### 3. Field Errors

- Use `field_errors` for validation errors
- Each field error must have `field`, `code`, and `message`
- Optional: include `constraint` for validation rules

### 4. Error Details

- Include information helpful for debugging
- Do not expose sensitive data (passwords, tokens, etc.)
- Include resource IDs and relevant context

### 5. HTTP Status Codes

- Use HTTP status code appropriate for error category
- 400: Client errors (validation, bad request)
- 401: Authentication errors
- 403: Authorization errors
- 404: Not found errors
- 409: Conflict errors
- 422: Business logic errors
- 429: Rate limit errors
- 500: Server errors
- 503: Service unavailable

---

## Adding New Error Codes

### Checklist

- [ ] Error code follows naming convention
- [ ] HTTP status code appropriate for error category
- [ ] Error message clear and actionable
- [ ] Documentation updated
- [ ] Error code added to ErrorCodeMap
- [ ] Test cases for error scenario

---

**This document will be updated according to API development and addition of new error codes.**