# CRM Domain Knowledge Base

**Location:** `apps/api/internal/crm/`
**Module:** `github.com/gilabs/gims/api/internal/crm`

## Architecture

Vertical slice following the standard GIMS pattern. All business logic must reside in `domain/usecase/`, with handlers remaining thin.

### Directory Structure
- `data/models/`: GORM entities and domain-specific logic (e.g., code generation).
- `data/repositories/`: Database abstraction layer.
- `domain/dto/`: Request and Response objects with Gin binding tags.
- `domain/mapper/`: Conversion logic between Models and DTOs.
- `domain/usecase/`: Core business logic and cross-domain orchestrations.
- `presentation/handler/`: HTTP request/response handling.
- `presentation/router/`: Route definitions.

### Key Models
- **Lead**: `lead.go` (Lead + LeadProductItem)
- **Deal**: `deal.go` (Deal + DealProductItem + DealHistory)
- **Activities**:
  - `task.go`: General tasks linked to Leads or Deals.
  - `schedule.go`: Calendar events.
  - `reminder.go`: Notifications for follow-ups.
  - `visit_report.go`: Field visit documentation.
  - `activity.go`: Unified activity log.
- **Master Data**:
  - `pipeline_stage.go`, `lead_source.go`, `lead_status.go`, `contact_role.go`, `activity_type.go`.
- **Others**: `contact.go`, `area_capture.go`.

## Domain Patterns

### Auto-generated Codes
All codes are generated using `apptime.Now()` to ensure timezone consistency. Never use `time.Now()`.
- **Deal**: `DEAL-YYYYMM-XXXXX` (via `generateDealCode()` in `deal.go`)
- **Lead**: `LEAD-YYYYMM-XXXXX` (via `generateLeadCode()` in `lead.go`)
- **Visit**: `VISIT-YYYYMM-XXXXX`

### Lead → Deal Conversion
- **Pattern**: Copy-then-link. Data is duplicated from Lead to Deal to preserve the historical state of the Lead.
- **Process**:
  1. Create new `Deal` record.
  2. Copy `LeadProductItems` to new `DealProductItems`.
  3. Set Lead status to `converted`.
  4. Link Lead to Deal via `deal_id` foreign key.
- **Note**: Ensure `TenantID` is propagated to all child items during conversion.

### Deal → Quotation Conversion
- **Cross-domain**: Invokes the Sales domain usecase.
- **Idempotency**: Tracked via `source_deal_id` on the `SalesQuotation` model (1:1 relationship).
- **Validation**:
  - Deal status must be `won`.
  - Must have at least one `DealProductItem`.
  - Must have an associated `Customer`.

### Archived Products Handling
- **Immutability**: Deals and Leads must remain readable even if the original Product is archived or deleted.
- **Snapshot Pattern**: `DealProductItem` and `LeadProductItem` store `ProductName` and `ProductSKU` at the time of addition.
- **Usecase Logic**: `deal_usecase.go` must not reject updates if a previously added product is now archived; it should rely on the snapshot data for display.

## Critical Rules & Gotchas

### TenantID on Child Models (CRITICAL)
The `database.GetDB(ctx, r.db)` helper automatically adds `WHERE tenant_id = ?` to all queries. GORM propagates these fields to INSERT columns. If a child model lacks the `TenantID` field, GORM throws `gorm.ErrInvalidField`, resulting in an HTTP 500.

**Required Field Definition:**
```go
TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
```

**Audit Status:**
- ✅ `DealProductItem` (deal.go:118) - Fixed.
- ✅ `DealHistory` (deal.go:159) - Fixed.
- ⚠️ `LeadProductItem` (lead.go:171) - **MISSING**. Will fail on create in tenant context.

### Timezone Support
- **Rule**: Never use bare `time.Now()`.
- **Correct Usage**:
  ```go
  import "github.com/gilabs/gims/api/internal/core/apptime"
  now := apptime.Now()
  ```

### Cross-Domain Dependencies
- **Sales**: For `SalesQuotation` generation.
- **Product**: `productModels.Product` for item selection.
- **Customer**: `customerModels.Customer` for Lead/Deal association.
- **Organization**: `orgModels.Employee` for `AssignedTo` and `ChangedBy` fields.
- **Geographic**: Province, City, and District lookups.
- **Inventory**: Stock availability checks.

## Known Bugs & Fixes

### TenantID Propagation Bug
- **Problem**: Child models without `TenantID` fail during creation because GORM tries to insert the `tenant_id` from the scoped DB into a non-existent field.
- **Fix**: Add the `TenantID` field to the model struct with proper GORM tags.
- **Deep Dive**: See `docs/features/core/gorm-tenant-scoping.md`.

### LeadProductItem Audit Warning
- **Warning**: `LeadProductItem` in `lead.go` is currently missing the `TenantID` field.
- **Impact**: Any attempt to create a Lead with products while logged into a tenant will result in a 500 Internal Server Error.
- **Action**: Add `TenantID` to `LeadProductItem` struct immediately when modifying CRM models.

## Key Gotchas
1. **Task Reference**: Both `Deal.Tasks` and `Lead.Tasks` reference the same `Task` model via `DealID` and `LeadID` foreign keys respectively.
2. **Geographic Data**: When querying geographic or platform-wide tables, avoid `GetDB()` if they are not tenant-scoped. Use `r.db.WithContext(ctx)` directly.
3. **Conversion Logic**: When converting Lead to Deal, ensure all metadata (including custom fields) is correctly mapped to the new Deal record.
