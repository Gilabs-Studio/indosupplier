# Backend Knowledge Base

**Location:** `apps/api/`
**Stack:** Go 1.25+, Gin, GORM, PostgreSQL, Redis
**Module:** `github.com/gilabs/indosupplier/api`

## Architecture

Vertical slice per domain:

```
internal/<domain>/
├── data/
│   ├── models/         # GORM entities
│   ├── repositories/   # Interface + implementation
│   └── seeders/        # Domain seeders
├── domain/
│   ├── dto/            # Request/Response with `binding` tags
│   ├── mapper/         # Model ↔ DTO
│   ├── usecase/        # Business logic (NEVER in handlers)
│   └── accounting/     # Finance-specific sub-package
└── presentation/
    ├── handler/         # Thin HTTP handlers
    ├── router/          # `<entity>_routers.go`
    └── routers.go       # Domain aggregator
```

## Critical Rules

### Model Registration (Mandatory)

After creating a new model, register in:

```go
// apps/api/internal/core/infrastructure/database/migrate.go
import finance "github.com/gilabs/indosupplier/api/internal/finance/data/models"

// In migrateWithErrorHandling():
&finance.Asset{},
```

### Import Order

```go
import (
    "context"    // stdlib
    "github.com/gin-gonic/gin"  // external
    "github.com/gilabs/indosupplier/api/internal/finance/data/models"  // internal
)
```

### API Response Format

```json
{
  "success": true,
  "data": {},
  "meta": { "pagination": {...} },
  "timestamp": "2024-01-15T10:30:45+07:00",
  "request_id": "req_abc123"
}
```

- Pagination: max `per_page` = 100, default 20
- Error codes: `VALIDATION_ERROR`, `UNAUTHORIZED`, `{RESOURCE}_NOT_FOUND`

### Seeder UUIDs (CRITICAL)

PostgreSQL enforces hex-only UUIDs (`0-9`, `a-f`). Never use non-hex letters.

```go
// ❌ WRONG — "rr" not valid hex
RecruitmentID = "rr000001-0000-0000-0000-000000000001"

// ✅ CORRECT
RecruitmentID = "ae000001-0000-0000-0000-000000000001"
```

Go cannot take address of constants — use local variables for pointer fields.

### GetFormData Pattern

For features with foreign key dropdowns, create `GET /<entity>/form-data` returning all options in one call. Place route **before** parameterized routes.

### TenantID pada Child Models (CRITICAL)

Setiap GORM model yang di-INSERT via repository yang menggunakan `database.GetDB()` **HARUS** memiliki field:

```go
TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
```

**Mengapa:** `GetDB()` menambahkan `WHERE tenant_id = ?` ke semua query. Saat GORM melakukan `CREATE`, ia mencoba meng-assign `tenant_id` ke model. Jika field tidak ada → `gorm.ErrInvalidField` → HTTP 500 "invalid field".

**Ini hanya perlu** untuk model yang di-create di tenant context (authenticated request). Model platform-wide (user, role, global config) yang pakai `r.db.WithContext(ctx)` langsung tidak perlu field ini.

**Detail:** `docs/features/core/gorm-tenant-scoping.md`

## Adding a New Domain

1. Create `internal/<domain>/` with full vertical slice
2. Register models in `migrate.go`
3. Add domain router in `cmd/api/main.go` or root router
4. Add seeders in `seeders/` and register in `seed_all.go`
5. Update Postman: `docs/postman/postman.json`

## Commands

```bash
cd apps/api

# Dev (Full Docker — rebuilds container)
pnpm dev                          # Docker + auto migrate/seed

# Dev (Local Go + Air — hot reload, recommended for daily dev)
docker compose up -d postgres redis   # Start DB + Redis only
air                                   # Hot reload (or: make dev)
# Or from root: pnpm dev:local

# Run without hot reload
go run ./cmd/api/main.go

# DB
make migrate                      # Migrations only
make seed                         # Seeders only
docker-compose up -d postgres     # PostgreSQL on port 5434

# Quality
go test ./...
go test ./internal/finance/...
go test -run TestName ./pkg/...
go vet ./...
go fmt ./...
go mod tidy

# ERP Analyzer
make analyze                      # Full analysis
make analyze MODULES=finance,sales
```

### CGO / WebP Note

`kolesa-team/go-webp` requires CGO + libwebp. Air builds with `CGO_ENABLED=0` on Windows.
Image upload works in both modes — WebP conversion is skipped when CGO is unavailable.
See: `docs/features/core/webp-build-tags.md`

## Anti-Patterns

- Relative imports (`internal/...` without module path)
- Business logic in handlers
- Bare `time.Now()` in business logic (use `apptime`)
- Missing `go mod tidy` after new files
- Forgetting model registration in `migrate.go`
- Missing `TenantID` field on child models created via `database.GetDB()` (see: [gorm-tenant-scoping.md](../../docs/features/core/gorm-tenant-scoping.md))
