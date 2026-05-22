# Architecture Proposal (Go Monolith With Vertical Slices)

Dokumen ini adalah standarisasi arsitektur layanan backend berbasis Go. Desain mengikuti pola vertical slice per domain, dengan layer yang jelas (data, domain, presentation) dan core cross-cutting (config, database, router, middleware).

Tujuan:
- Konsistensi struktur proyek dan penamaan file.
- Skalabel, mudah di-maintain, dan mudah menambah fitur/domain baru.
- Siap untuk pengembangan lokal, pengujian, dan deployment multi-lingkungan (dev/staging/prod).

Ringkasan:
- Bahasa: Go (100%)
- Eksekutables: API server (HTTP), worker background, seeder, generator.
- Infrastruktur inti: konfigurasi, koneksi DB, migrasi & seeding, router, middleware.
- Domain vertical slice: data (models, repositories, seeders), domain (dto, mapper, usecase), presentation (handler, router).
- Orkestrasi: Dockerfile, docker-compose per lingkungan.
- Migrasi: Atlas (berbasis SQL dengan timestamp).
- Testing: unit dan e2e.

---

## 1. High-Level Architecture

Komponen:
- cmd/api: Entry point HTTP API.
- cmd/<worker>: Worker untuk job background (opsional; contoh: attendance-worker).
- cmd/seed: Runner seeding data.
- cmd/gen: Generator scaffolding (opsional).
- internal/core: Concern lintas domain (config, database, router, middleware, utils, constants).
- internal/<domain>: Vertical slice per domain (data → domain → presentation).

Alur request (sederhana):
```
Client → Router → Middleware → Handler → Usecase → Repository → Database
                       │
                     (DTO/Validation, Mapper)
```

Diagram data flow (ASCII):
```
+--------+     +---------+     +----------+     +----------+     +---------+
| Client | --> |  Router | --> | Middleware| --> | Handler  | --> | Usecase |
+--------+     +---------+     +----------+     +----------+     +----+----+
                                                                     |
                                                                     v
                                                               +-----+------+
                                                               | Repository |
                                                               +-----+------+
                                                                     |
                                                                     v
                                                                +----+-----+
                                                                | Database |
                                                                +----------+
```

Prinsip:
- Separation of concerns per layer.
- Domain-first (entity dan usecase mandiri dari framework).
- Dependency mengalir dari luar ke dalam (presentation → domain → data).
- Cross-cutting (auth, config, logging) tersentral di core.

---

## 2. Project Structure

Struktur direktori utama:
```
.
├─ cmd/
│  ├─ api/                  # HTTP server entry point
│  ├─ <worker-name>/        # Worker background (opsional)
│  ├─ gen/                  # Generator scaffolding (opsional)
│  └─ seed/                 # Seeder runner
├─ internal/
│  ├─ core/
│  │  ├─ constants/
│  │  ├─ events/                # Domain events (user_events.go, role_events.go, etc.)
│  │  ├─ infrastructure/
│  │  │  ├─ config/
│  │  │  ├─ database/       # connection.go, migrate.go, seed.go
│  │  │  ├─ events/         # Event infrastructure (types.go, publisher.go, noop_publisher.go)
│  │  │  ├─ redis/          # redis.go (client init)
│  │  │  └─ router/
│  │  ├─ middleware/
│  │  └─ utils/
│  └─ <domain>/
│     ├─ data/
│     │  ├─ models/
│     │  ├─ repositories/
│     │  └─ seeders/
│     ├─ domain/
│     │  ├─ dto/
│     │  ├─ mapper/
│     │  └─ usecase/
│     └─ presentation/
│        ├─ handler/
│        ├─ router/         # <entity>_routers.go
│        └─ routers.go      # aggregator router domain
├─ migrations/              # Atlas SQL migrations (timestamped)
├─ templates/               # scaffolding templates (dto/handler/repository/routers/usecase)
├─ test/
│  └─ e2e/
├─ .env.example
├─ Dockerfile(.dev/.test/.worker)
├─ docker-compose(.dev/.staging/.prod/.test).yml
├─ go.mod / go.sum
└─ README.md
```

---

## 3. Domain Module Layout (Vertical Slice)

Contoh domain: `master-data`

```
internal/master-data/
├─ data/
│  ├─ models/
│  │  └─ product.go
│  ├─ repositories/
│  │  └─ product_repository.go
│  └─ seeders/
│     └─ product_seeder.go
├─ domain/
│  ├─ dto/
│  │  └─ product_dto.go
│  ├─ mapper/
│  │  └─ product_mapper.go
│  └─ usecase/
│     └─ product_usecase.go
└─ presentation/
│  ├─ handler/
│  │  └─ product_handler.go
│  ├─ router/
│  │  └─ product_routers.go
│  └─ routers.go
```

### 3.1 Contoh DTO
```go
// internal/master-data/domain/dto/product_dto.go
package dto

type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required,min=3"`
    SKU         string  `json:"sku" binding:"required,alphanum"`
    CategoryID  string  `json:"category_id" binding:"required,uuid"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Description string  `json:"description"`
}

type ProductResponse struct {
    ID         string  `json:"id"`
    Name       string  `json:"name"`
    SKU        string  `json:"sku"`
    CategoryID string  `json:"category_id"`
    Price      float64 `json:"price"`
    CreatedAt  string  `json:"created_at"`
    UpdatedAt  string  `json:"updated_at"`
}
```

### 3.2 Contoh Model
```go
// internal/master-data/data/models/product.go
package models

import "time"

type Product struct {
    ID         string    `gorm:"type:uuid;primaryKey"`
    Name       string    `gorm:"size:255;not null"`
    SKU        string    `gorm:"size:64;uniqueIndex;not null"`
    CategoryID string    `gorm:"type:uuid;index"`
    Price      float64   `gorm:"not null"`
    Description string   `gorm:"type:text"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### 3.3 Mapper
```go
// internal/master-data/domain/mapper/product_mapper.go
package mapper

import (
    "internal/master-data/data/models"
    "internal/master-data/domain/dto"
)

func ToProductResponse(m *models.Product) dto.ProductResponse {
    return dto.ProductResponse{
        ID:         m.ID,
        Name:       m.Name,
        SKU:        m.SKU,
        CategoryID: m.CategoryID,
        Price:      m.Price,
        CreatedAt:  m.CreatedAt.Format(time.RFC3339),
        UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
    }
}
```

### 3.4 Repository
```go
// internal/master-data/data/repositories/product_repository.go
package repositories

import (
    "context"

    "gorm.io/gorm"
    "internal/master-data/data/models"
)

type ProductRepository interface {
    Create(ctx context.Context, p *models.Product) error
    FindByID(ctx context.Context, id string) (*models.Product, error)
    List(ctx context.Context, limit, offset int, q string) ([]models.Product, int64, error)
    Update(ctx context.Context, p *models.Product) error
    Delete(ctx context.Context, id string) error
}

type productRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *models.Product) error {
    return r.db.WithContext(ctx).Create(p).Error
}
```

### 3.5 Usecase
```go
// internal/master-data/domain/usecase/product_usecase.go
package usecase

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "internal/master-data/data/models"
    "internal/master-data/data/repositories"
    "internal/master-data/domain/dto"
    "internal/master-data/domain/mapper"
)

type ProductUsecase interface {
    Create(ctx context.Context, req dto.CreateProductRequest) (dto.ProductResponse, error)
    // ... List, Get, Update, Delete
}

type productUsecase struct {
    repo repositories.ProductRepository
}

func NewProductUsecase(repo repositories.ProductRepository) ProductUsecase {
    return &productUsecase{repo: repo}
}

func (u *productUsecase) Create(ctx context.Context, req dto.CreateProductRequest) (dto.ProductResponse, error) {
    if req.Price <= 0 {
        return dto.ProductResponse{}, errors.New("price must be > 0")
    }
    m := &models.Product{
        ID:          uuid.NewString(),
        Name:        req.Name,
        SKU:         req.SKU,
        CategoryID:  req.CategoryID,
        Price:       req.Price,
        Description: req.Description,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    if err := u.repo.Create(ctx, m); err != nil {
        return dto.ProductResponse{}, err
    }
    return mapper.ToProductResponse(m), nil
}
```

### 3.6 Handler
```go
// internal/master-data/presentation/handler/product_handler.go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "internal/master-data/domain/dto"
    "internal/master-data/domain/usecase"
)

type ProductHandler struct {
    uc usecase.ProductUsecase
}

func NewProductHandler(uc usecase.ProductUsecase) *ProductHandler {
    return &ProductHandler{uc: uc}
}

func (h *ProductHandler) Create(c *gin.Context) {
    var req dto.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "detail": err.Error()})
        return
    }
    res, err := h.uc.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, res)
}
```

### 3.7 Router per Entity
```go
// internal/master-data/presentation/router/product_routers.go
package router

import (
    "github.com/gin-gonic/gin"
    "internal/master-data/presentation/handler"
)

func RegisterProductRoutes(rg *gin.RouterGroup, h *handler.ProductHandler) {
    g := rg.Group("/products")
    g.POST("", h.Create)
    // g.GET("", h.List)
    // g.GET("/:id", h.GetByID)
    // g.PUT("/:id", h.Update)
    // g.DELETE("/:id", h.Delete)
}
```

### 3.8 Aggregator Router Domain
```go
// internal/master-data/presentation/routers.go
package presentation

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "internal/master-data/data/repositories"
    "internal/master-data/domain/usecase"
    "internal/master-data/presentation/handler"
    "internal/master-data/presentation/router"
)

func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup, db *gorm.DB) {
    // Wiring dependency (di domain ini)
    productRepo := repositories.NewProductRepository(db)
    productUC := usecase.NewProductUsecase(productRepo)
    productH := handler.NewProductHandler(productUC)

    // Group path untuk domain master-data
    group := api.Group("/master-data")
    router.RegisterProductRoutes(group, productH)
    // Tambahkan entity lain di sini...
}
```

---

## 4. Core Layer

### 4.1 Config
- Single source of truth untuk konfigurasi aplikasi (ENV-first).
- Bentuk:
  - App: PORT, ENV, LOG_LEVEL
  - Database: DB_DSN/DB_HOST/DB_NAME/DB_USER/DB_PASS
  - Security: JWT_SECRET, ALLOW_ORIGINS
- Rekomendasi: var binding + fallback default + validasi awal.
```go
// internal/core/infrastructure/config/config.go
type Config struct {
    AppEnv   string
    HTTPPort string
    DBDSN    string
    JWTSecret string
    // ...
}
```

### 4.2 Database
- Satu paket koneksi (`connection.go`) yang mengembalikan *gorm.DB.
- Migrations (Atlas) dipanggil saat startup (opsional) berdasarkan ENV.
- Seeding per environment (seed.go, production_seed.go).
- Transaksi ditangani di usecase bila perlu (pass *gorm.DB tx).

### 4.3 Router
- Inisialisasi Gin/Fiber (contoh: Gin).
- Versioning: prefix `/api/v1`.
- Registrasi per domain: panggil `<domain>/presentation.RegisterRoutes`.
- Middlewares global: logging, recovery, CORS, auth (opsional).

### 4.4 Middleware
- Auth: verifikasi token, inject user context.
- Authorize: cek permission berdasarkan role/action.
- RequestID, Rate limiting (opsional), CORS.

---

## 5. Konvensi Penamaan

- Go files: lower_snake_case untuk multi-kata.
  - Contoh: `auth_middleware.go`, `production_seed.go`, `connection.go`, `router.go`.
- Domain directories: singular, deskriptif; jika multi-kata bisa memakai tanda hubung (mis. `master-data`).
- Router per entity: `<entity>_routers.go`
- Handler per entity: `<entity>_handler.go`
- Usecase per entity: `<entity>_usecase.go`
- Repository per entity: `<entity>_repository.go`
- DTO per entity: `<entity>_dto.go`
- Mapper per entity: `<entity>_mapper.go`
- Migrations (Atlas): `YYYYMMDDhhmmss_<optional-description>.sql`
  - Contoh: `20260101120000_add_products.sql`
- Docker Compose variants: `docker-compose.<env>.yml`
  - Contoh: `docker-compose.prod.yml`, `docker-compose.staging.yml`, `docker-compose.test.yml`
- Aset: deskriptif sesuai konteks (mis. `mika-logo.jpeg`, `header_company.jpeg`)

---

## 6. Error Handling & Response

Standar respons JSON mengikuti [API Response Standards](./api-response-standards.md).

Contoh ringkas:
```json
// Sukses (collection)
{
  "data": [{ "id": "..." }],
  "meta": { "page": 1, "size": 10, "total": 100 }
}

// Sukses (single)
{ "data": { "id": "...", "name": "..." } }

// Error
{ "error": "validation_error", "detail": "field 'name' is required" }
```

Pedoman:
- Validasi input di handler (binding tags) → 400.
- Bisnis rule violation → 422.
- Not found → 404.
- Internal error → 500.
- Gunakan kode error yang konsisten untuk diobservasi/logging.

---

## 7. Testing Strategy

- Unit test:
  - Usecase: mock repository.
  - Repository: gunakan test DB (in-memory/isolated schema).
- E2E test (test/e2e):
  - Jalankan service dengan docker-compose.test.yml.
  - Seed data minimal.
  - Uji alur utama (create → get → update → delete).
- Contoh file:
  - `internal/<domain>/domain/usecase/product_usecase_test.go`
  - `test/e2e/product_flow_test.go`

---

## 8. Migrations (Atlas)

- Penamaan waktu: `YYYYMMDDhhmmss` agar kronologis.
- Satu perubahan skema per file (kecil, mudah direview/rollback).
- Simpan checksum di `migrations/atlas.sum`.
- Lifecycle:
  - Dev: jalankan otomatis saat startup (opsional).
  - Staging/Prod: jalankan langkah terkontrol di pipeline.

Contoh:
```
migrations/
├─ 20260101120000_initial_schema.sql
├─ 20260105103000_add_products.sql
└─ atlas.sum
```

---

## 9. Seeding

- `seed.go`: data dasar (mis. role, permission).
- `production_seed.go`: seeding aman untuk production (idempotent).
- Domain-specific seeders di `internal/<domain>/data/seeders/`.

---

## 10. Observability

- Logging terstruktur (level, trace/request_id, user_id).
- Metrics (Prometheus): HTTP latency, DB latency, error rate (opsional).
- Tracing (OpenTelemetry) untuk jalur kritikal (opsional).

---

## 11. Security

- Secrets via ENV (jangan commit .env real).
- JWT untuk auth, role/permission untuk authorize.
- Validasi input ketat (binding & custom validator).
- CORS ketat sesuai origin.
- Sanitasi output; hindari leak detail internal pada error.

---

## 12. Performance & Scalability

- Pagination default di listing endpoints.
- Indeks DB via migrasi untuk kolom pencarian/relasi.
- Hindari N+1 (preload yang diperlukan).
- Cache layer (opsional) untuk lookup statis (mis: Redis).

---

## 13. CI/CD (Garis Besar)

- Pipeline:
  - Lint + Format
  - Unit test
  - Build image (API/worker)
  - Migrate DB (terkontrol)
  - Deploy ke staging
  - E2E smoke test
  - Manual approval → Prod
- GitHub Actions contoh job:
  - go build/test
  - docker buildx
  - atlas migrate apply (dengan guard)

---

## 14. Local Development

- Live reload: air (`.air.toml`).
- Jalankan:
```
go run ./cmd/api
# atau
docker compose up api db
```
- ENV: salin `.env.example` → `.env` dan sesuaikan.
- Makefile (opsional) untuk perintah umum:
```
make run
make test
make migrate
make seed
```

---

## 15. Background Jobs

- Struktur `cmd/<worker>` + paket usecase spesifik job.
- Queue: Redis/DB (opsional).
- Scheduling: cron (opsional) atau external scheduler.

---

## 16. API Versioning & Deprecation

- Prefix route: `/api/v1`.
- Perubahan breaking → `/api/v2`.
- Deprecation policy: tandai endpoint deprecated + grace period.

---

## 17. Checklist Tambah Entity Baru (Template)

1) Model
- [ ] Buat `internal/<domain>/data/models/<entity>.go`
- [ ] Tambah migrasi: `migrations/YYYYMMDDhhmmss_add_<entity>.sql`

2) Repository
- [ ] Buat `internal/<domain>/data/repositories/<entity>_repository.go`
- [ ] Tambahkan method: Create, FindByID, List(pagination), Update, Delete

3) DTO & Mapper
- [ ] `internal/<domain>/domain/dto/<entity>_dto.go`
- [ ] `internal/<domain>/domain/mapper/<entity>_mapper.go`

4) Usecase
- [ ] `internal/<domain>/domain/usecase/<entity>_usecase.go`
- [ ] Unit test

5) Handler & Router
- [ ] `internal/<domain>/presentation/handler/<entity>_handler.go`
- [ ] `internal/<domain>/presentation/router/<entity>_routers.go`
- [ ] Daftarkan di `internal/<domain>/presentation/routers.go`

6) E2E Test
- [ ] `test/e2e/<entity>_flow_test.go`

7) Docs
- [ ] Update README / docs domain

---

## 18. Contoh Registrasi Global (Router Core)

```go
// internal/core/infrastructure/router/router.go
func Register(r *gin.Engine, db *gorm.DB) {
    api := r.Group("/api/v1")

    // domain master-data
    masterdata.RegisterRoutes(r, api, db)

    // domain lain ...
}
```

---

## 19. Template Scaffolding (Opsional)

Direktori `templates/`:
- dto.tmpl
- handler.tmpl
- repository.tmpl
- routers.tmpl
- usecase.tmpl

Contoh variabel template:
- {{.Entity}}: Product
- {{.EntityVar}}: product
- {{.Package}}: master-data

Generator `cmd/gen` dapat menggantikan placeholder dan menulis file ke path yang tepat.

---

## 20. Contoh Konvensi Routing (REST)

- Collection:
  - GET `/api/v1/<domain>/<entities>?page=1&size=10&q=foo`
  - POST `/api/v1/<domain>/<entities>`
- Item:
  - GET `/api/v1/<domain>/<entities>/:id`
  - PUT `/api/v1/<domain>/<entities>/:id`
  - DELETE `/api/v1/<domain>/<entities>/:id`

---

Dengan arsitektur ini, setiap domain berdiri mandiri (vertical slice), dependency tertata rapi, penamaan konsisten, dan siap untuk skala pengembangan tim. Dokumen ini dapat digunakan sebagai blueprint untuk memulai proyek Go baru atau menstandarkan proyek yang sudah berjalan.