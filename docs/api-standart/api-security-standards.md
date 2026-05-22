# API Security Standards

This document outlines the security standards and best practices implemented in the API.

## 1. Authentication

### 1.1 Token Storage
- **Access Tokens** and **Refresh Tokens** MUST be stored in **HttpOnly, Secure, SameSite=Strict** cookies.
- Tokens MUST NOT be exposed in JSON response bodies to prevent token theft via XSS.
- The `Authorization` header is NOT used for client-server communication in browser environments; the server reads directly from cookies.

### 1.2 Token Rotation
- Refresh tokens are rotated on every use.
- The old refresh token is revoked immediately.
- **Row-Level Locking (`FOR UPDATE`)** is used during rotation to prevent race conditions where a reused token could spawn multiple valid chains.

## 2. CSRF Protection

### 2.1 Double-Submit Cookie Pattern
- The API implements the Double-Submit Cookie pattern.
- A non-HttpOnly cookie `csrf_token` is set by the server.
- All mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) MUST include an `X-CSRF-Token` header matching the cookie value.
- The server validates this header in the `CSRF` middleware.

## 3. SQL Injection Prevention

### 3.1 Parameterized Queries
- **ALWAYS** use parameterized queries for all database operations.
- Using GORM's built-in methods (e.g., `Where("id = ?", id)`) satisfies this requirement.
- **NEVER** use `fmt.Sprintf` or string concatenation to build SQL queries.

### 3.2 Secure Sorting (ORDER BY)
- Direct string concatenation in `ORDER BY` clauses is a common vector for SQL injection.
- **NEVER** pass raw strings from user input (e.g., `req.SortBy`) directly into `.Order()`.
- Use **Whitelisting** and **GORM Clauses**:
```go
// ✅ CORRECT: Using OrderByColumn clause
query = query.Order(clause.OrderByColumn{
    Column: clause.Column{Name: params.SortBy},
    Desc:   params.SortDir == "desc",
})
```
- For operations involving joins, always qualify the column name (e.g., `tableName.columnName`) to prevent ambiguity.

## 4. XSS Prevention

### 4.1 URL Sanitization
- When handling user-provided URLs (e.g., LinkedIn profiles, external sites), perform robust validation.
- **NEVER** use simple substring checks like `.includes("linkedin.com")` as they can be bypassed (e.g., `evil.com/linkedin.com`).
- Use the **URL Constructor** to validate hosts:
```typescript
// ✅ CORRECT (Frontend)
try {
  const url = new URL(inputUrl);
  const isValidHost = ["linkedin.com", "www.linkedin.com"].includes(url.hostname);
} catch (e) {
  // Invalid URL format
}
```
- Reject pseudo-protocols like `javascript:` entirely.

### 4.2 Handling User-Provided HTML
- The API follows a **JSON-Only** approach for most data exchange.
- If data must be rendered in HTML context, it MUST be properly escaped.
- For the frontend (Next.js/React), standard JSX automatically escapes content. Avoid using `dangerouslySetInnerHTML`.

## 5. Rate Limiting

### 5.1 Redis-Backed Limiter
- Rate limiting is implemented using **Redis** (Fixed Window algorithm) for distributed tracking.
- For production/multi-instance environments, limiter failure policy MUST be explicit:
	- Critical public/auth endpoints: prefer fail-closed to avoid distributed bypass during incidents.
	- Non-critical internal endpoints: controlled fail-open allowed only with alerting.

### 5.2 Limit Levels
- **Global Login Limit**: Per-IP limit on login attempts to prevent DOS.
- **Email-Based Limit**: Limits login attempts per email address to prevent brute-force attacks on specific accounts.
- **General Limit**: Default limit for all authenticated endpoints.

### 5.3 Retry Storm Prevention
- Return `429 Too Many Requests` with `Retry-After` when limits are exceeded.
- Avoid expensive work before limiter checks (fail fast).
- Enforce tighter limits on authentication, upload, and expensive report endpoints.

## 6. Input Validation & Data Integrity

### 6.1 Strict Validation (Backend)
- All request bodies MUST be bound using `ShouldBindJSON` with strict structure definitions.
- `binding` tags (e.g., `required`, `email`, `min=8`) MUST be present on all DTO fields.

### 6.2 Schema Validation (Frontend)
- Use **Zod** for complex form validation and type safety.
- Validate data at the edge of the application (API Client / Forms).

### 6.3 Transactional Integrity
- Critical flows (Login, Token Refresh, Logout) MUST be wrapped in database transactions (`db.Transaction`) to ensure atomicity.

## 7. Permissions-Policy (Feature Policy)

The `Permissions-Policy` HTTP header controls which browser features the application is allowed to use.

### 7.1 Current Policy

| Feature        | Policy   | Reason                                                   |
| -------------- | -------- | -------------------------------------------------------- |
| `geolocation`  | `(self)` | Required for GPS-based attendance clock-in/out           |
| `camera`       | `(self)` | Required for WFH photo proof during clock-in             |
| `microphone`   | `()`     | Not used — blocked entirely                              |
| `fullscreen`   | `(self)` | Allowed for the app's own origin                         |
| `payment`      | `()`     | Not used — blocked entirely                              |
| `usb`          | `()`     | Not used — blocked entirely                              |

> **⚠️ CRITICAL:** `geolocation` and `camera` MUST be set to `(self)`. Setting `()` completely blocks the API at the HTTP level.

## 8. Security Headers

The following headers MUST be present on all responses:
- `Strict-Transport-Security` (HSTS): Enforces HTTPS.
- `Content-Security-Policy` (CSP): Restricts origins for scripts, styles, and images.
- `X-Content-Type-Options: nosniff`: Prevents MIME type sniffing.
- `X-Frame-Options: DENY`: Prevents clickjacking.
- `Referrer-Policy: strict-origin-when-cross-origin`: Controls referral data.

## 9. Dependency Degradation Safety
- All dependency calls (DB/Redis/external API) MUST use timeout-bound contexts.
- On dependency timeout/failure, API MUST return controlled error responses (no panic, no process crash).
- Background workers MUST stop gracefully during shutdown before resource teardown.
