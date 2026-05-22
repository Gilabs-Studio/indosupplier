# Web Application API

Backend API untuk Web Application Platform menggunakan Go dan Gin framework.

## Tech Stack API 

- **Go**: 1.25+
- **Gin**: Web framework
- **PostgreSQL**: Database
- **GORM**: ORM
- **JWT**: Authentication
- **Docker**: Containerization

## Arsitektur

### Layered Architecture

```
Handler Layer (internal/api/handlers/)
    ↓
Service Layer (internal/service/)
    ↓
Repository Layer (internal/repository/)
    ↓
Database (PostgreSQL)
```

### Dependency Injection

- Repository di-inject ke Service
- Service di-inject ke Handler
- JWT Manager di-inject ke Service dan Middleware

### Interface-Based Design

- Repository menggunakan interface (`internal/repository/interfaces/`)
- Implementasi di `internal/repository/postgres/`
- Memudahkan testing dan perubahan database

## Setup Development

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 16+
- pnpm (untuk monorepo)
- Docker & Docker Compose (optional)

### Local Development dengan pnpm

1. Dari root monorepo:

```bash
pnpm dev --filter=api
```

2. Atau dari folder api:

```bash
cd apps/api
pnpm dev
```

3. Atau langsung dengan Go:

```bash
cd apps/api
go run cmd/server/main.go
```

### Maintenance Tools

**Backfill snapshot (immutability transaksi Purchase + Finance)**

Jika kolom snapshot baru sudah ada di DB (AutoMigrate), jalankan tool ini sekali untuk mengisi snapshot pada data transaksi yang sudah ada agar tampilan historis tidak berubah saat master data berubah:

```bash
cd apps/api
go run ./cmd/tools/backfill-snapshots
```

### Environment Variables

Copy `.env.example` ke `.env` dan sesuaikan:

```bash
cp .env.example .env
```

**Required Variables:**

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL user (default: postgres)
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: Database name (default: gims_erp)
- `JWT_SECRET`: JWT secret key (min 32 characters)

**Storage Configuration:**

- `STORAGE_TYPE`: Storage type - `local` (default) or `r2` for Cloudflare R2
- `STORAGE_UPLOAD_DIR`: Directory for local storage (default: `./uploads`)
- `STORAGE_BASE_URL`: Base URL for serving files (default: `/uploads`)

**Cloudflare R2 Configuration (if using R2):**

- `R2_ENDPOINT`: R2 endpoint URL (e.g., `https://<account-id>.r2.cloudflarestorage.com`)
- `R2_ACCESS_KEY_ID`: R2 Access Key ID
- `R2_SECRET_ACCESS_KEY`: R2 Secret Access Key
- `R2_BUCKET`: R2 Bucket name
- `R2_PUBLIC_URL`: Public URL for R2 bucket (custom domain or r2.dev subdomain)

Untuk detail setup storage, lihat [STORAGE_SETUP.md](./STORAGE_SETUP.md)

### Database Setup

#### Option 1: Docker Compose (Recommended)

```bash
cd apps/api
docker-compose up -d postgres
```

**Note**: Docker Compose menggunakan port **5434** (bukan 5432) untuk menghindari konflik dengan PostgreSQL lain. Pastikan `.env` file menggunakan `DB_PORT=5434`.

#### Option 2: Local PostgreSQL

1. Install PostgreSQL
2. Create database:

```sql
CREATE DATABASE gims_erp;
```

**Note**: Jika menggunakan local PostgreSQL, gunakan `DB_PORT=5432` di `.env` file.

### Run Migrations & Seeders

Migrations dan seeders akan otomatis dijalankan saat server start.

**Manual migration:**

```bash
cd apps/api
go run cmd/server/main.go
```

**Seed data:**

- Admin: `admin@example.com` / `admin123`
- Doctor: `doctor@example.com` / `admin123`
- Pharmacist: `pharmacist@example.com` / `admin123`

### Build

Dari root monorepo:

```bash
pnpm build --filter=api
```

Atau dari folder api:

```bash
cd apps/api
pnpm build
# atau
go build -o bin/server ./cmd/server/main.go
```

### Docker Development

1. Build and run with Docker Compose:

```bash
cd apps/api
docker-compose up --build
```

2. Atau build image manually:

```bash
cd apps/api
docker build -t gims-api .
docker run -p 8087:8087 gims-api
```

## API Endpoints

### Health Check

- `GET /health` - Health check endpoint
- `GET /ping` - Ping endpoint

### Authentication

- `POST /api/v1/auth/login` - Login dengan email/password
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout (client-side)

## Project Structure

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                      # HTTP layer
│   │   ├── handlers/             # Request handlers
│   │   ├── middleware/           # HTTP middleware
│   │   └── routes/               # Route definitions
│   ├── domain/                   # Business logic & entities
│   │   └── auth/                 # Auth domain
│   ├── repository/               # Data access layer
│   │   ├── interfaces/           # Repository interfaces
│   │   └── postgres/             # PostgreSQL implementation
│   ├── service/                  # Application services
│   │   └── auth/                 # Auth service
│   ├── config/                   # Configuration
│   └── database/                 # Database connection
├── pkg/                          # Public packages
│   ├── response/                 # API response helpers
│   ├── errors/                   # Error handling
│   ├── jwt/                      # JWT utilities
│   └── logger/                   # Logging
├── migrations/                   # Database migrations
├── seeders/                      # Database seeders
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Development Guidelines

- Follow Go best practices and conventions
- Use the API response standards defined in `/docs/api-standart/`
- Implement error codes as defined in `/docs/api-standart/api-error-codes.md`
- Use layered architecture (Handler → Service → Repository)
- Use interface-based design for repositories
- Implement dependency injection pattern

## Error Handling

Semua error mengikuti format standar:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": {},
    "field_errors": []
  },
  "timestamp": "2024-01-15T10:30:45+07:00",
  "request_id": "req_abc123xyz"
}
```

## Logging

- Structured logging menggunakan `pkg/logger`
- Request logging via middleware
- Error logging dengan context

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/service/auth/...
```

## Next Steps

- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Setup CI/CD
- [ ] Add API documentation (Swagger)
- [ ] Performance optimization
- [ ] Add rate limiting
- [ ] Add caching layer (Redis)
