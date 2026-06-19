# API Configuration & Seed Data Standards

## Purpose

Runtime API code must not own business seed data, display defaults, or environment-specific policy. API code should read configuration and persisted reference data; seeders should create demo/default data.

---

## Configuration Rules

### Do

- Put runtime policy in `internal/core/infrastructure/config`.
- Put reusable formatting helpers in `internal/core/utils`.
- Read values through config/helper functions instead of literal strings in handlers/usecases/repositories.
- Use `apptime.Now()` for application timestamps.
- Use `utils.NormalizePagination()` for all paginated queries.
- Use `utils.FormatMoney()` and `utils.DefaultCurrency()` for currency formatting/defaulting.

### Do Not

- Do not hardcode business values in usecases, handlers, or repositories.
- Do not seed or auto-create business data inside request paths.
- Do not use bare `time.Now()` outside `internal/core/apptime`.
- Do not introduce endpoint-specific pagination caps.
- Do not use `limit` as a public query parameter for page size; use `per_page`.

---

## Seed Data Ownership

Seed/demo/reference data belongs in `apps/api/seeders`.

Examples:

- Subscription plans and prices.
- Demo suppliers/products.
- Demo RFQ bids.
- Demo content/articles.
- Initial admin accounts.

Usecases may read this data from repositories, but must not create it automatically.

---

## Pagination Standard

Public list endpoints use:

```text
page
per_page
```

Rules:

- Default `per_page`: `PAGINATION_DEFAULT_PER_PAGE`, fallback `20`.
- Maximum `per_page`: `PAGINATION_MAX_PER_PAGE`, fallback `20`.
- All handlers/usecases/repositories must normalize through `utils.NormalizePagination()`.
- Deprecated aliases such as `limit` must not be introduced.

---

## Date & Time Standard

- Application timestamps: `apptime.Now()`.
- Company-specific timestamps: `apptime.NowForCompany(companyID)`.
- Employee-specific timestamps: `apptime.NowForEmployee(employeeID)`.
- Date display formatting should use named helpers/constants in `core/utils`, not inline layouts in business usecases.

---

## Currency Standard

- Currency codes must come from request data, persisted data, or `utils.DefaultCurrency()`.
- Money display strings must use `utils.FormatMoney(amount, currency)`.
- Database model hooks can default empty currency values through `utils.DefaultCurrency()`.
- Do not add static DB defaults like `default:'IDR'` when the value is configurable.
