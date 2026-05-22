# Setup Guide - API

## Quick Start

### Option 1: Using Docker Compose (Recommended)

1. **Start PostgreSQL with Docker Compose:**

```bash
cd apps/api
docker-compose up -d postgres
```

2. **Copy environment file:**

```bash
cp .env.example .env
```

**Note**: Jika port 5432 sudah digunakan, docker-compose akan menggunakan port 5434. Pastikan `.env` file menggunakan port yang sesuai (5434 untuk Docker, 5432 untuk local PostgreSQL).

3. **Run the server:**

```bash
# From root
pnpm dev --filter=@repo/api

# Or from apps/api
go run cmd/api/main.go
```

### Option 2: Using Local PostgreSQL

1. **Install PostgreSQL** (if not installed):

```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# macOS
brew install postgresql
```

2. **Create database:**

```bash
sudo -u postgres psql
```

Then in PostgreSQL shell:

```sql
CREATE DATABASE gims_erp;
CREATE USER postgres WITH PASSWORD 'postgres';
GRANT ALL PRIVILEGES ON DATABASE gims_erp TO postgres;
\q
```

3. **Update .env file** with your PostgreSQL credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=gims_erp
DB_SSLMODE=disable
```

4. **Run the server:**

```bash
go run cmd/api/main.go
```

## Environment Variables

Copy `.env.example` to `.env` and update the values:

```env
# Server Configuration
PORT=8080
ENV=development

# Reverse Proxy / Load Balancer
# Enable ONLY when you are behind a trusted proxy that sets X-Forwarded-Proto.
PROXY_HEADERS_ENABLED=false
# Comma-separated list of trusted proxy IPs/CIDRs (required when PROXY_HEADERS_ENABLED=true)
TRUSTED_PROXIES=127.0.0.1

# Startup Safety (recommended defaults)
# Production default is false; enable explicitly when needed.
RUN_MIGRATIONS=true
RUN_SEEDERS=true

# Request Limits
SERVER_MAX_BODY_BYTES=1048576
SERVER_MAX_MULTIPART_BODY_BYTES=52428800
SERVER_MAX_MULTIPART_MEMORY_BYTES=8388608

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gims_erp
DB_SSLMODE=disable

# DB Pool / GORM (optional)
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME_MINUTES=30
DB_CONN_MAX_IDLE_TIME_MINUTES=10
DB_PREPARE_STMT=false
DB_SKIP_DEFAULT_TRANSACTION=false

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production-min-32-chars
JWT_ISSUER=template-1n

# Optional (recommended): split secrets
# If these are set, they override JWT_SECRET.
JWT_ACCESS_SECRET=
JWT_REFRESH_SECRET=

# Optional (recommended for key rotation): key rings + kid
# Format: comma-separated "kid:secret" pairs.
# If JWT_ACCESS_KID is set and matches a key in JWT_ACCESS_KEYS, that key is used for signing.
# Verification will accept any key in the ring (useful for rotation).
JWT_ACCESS_KEYS=
JWT_ACCESS_KID=
JWT_REFRESH_KEYS=
JWT_REFRESH_KID=
JWT_ACCESS_TTL=24
JWT_REFRESH_TTL=7

# Seeder (only used when RUN_SEEDERS=true)
# In production you MUST set this.
SEED_DEFAULT_PASSWORD=change-me

# Redis tuning (optional)
REDIS_DIAL_TIMEOUT_SEC=10
REDIS_READ_TIMEOUT_SEC=30
REDIS_WRITE_TIMEOUT_SEC=30
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5

# Observability (non-production only; token required)
PPROF_ENABLED=false
PPROF_TOKEN=
METRICS_ENABLED=false
METRICS_TOKEN=
```

## Ops Commands (recommended)

From `apps/api`:

- Run migrations: `make migrate`
- Run seeders: `make seed`
- Run API: `make run`

## Troubleshooting

### Database Connection Error

**Error**: `password authentication failed for user "postgres"`

**Solutions**:

1. Check if PostgreSQL is running:

```bash
# Docker
docker ps | grep postgres

# System service
sudo systemctl status postgresql
```

2. Verify database credentials in `.env` file

3. If using Docker Compose, make sure the service is up:

```bash
docker-compose up -d postgres
docker-compose ps
```

4. Test connection manually:

```bash
psql -h localhost -p 5432 -U postgres -d gims_erp
```

### Database Doesn't Exist

Create the database:

```bash
# Using Docker
docker exec -it gims-db psql -U postgres
CREATE DATABASE gims_erp;
\q

# Using local PostgreSQL
createdb -U postgres gims_erp
```

## Default Seeded Users

If `RUN_SEEDERS=true` and the database is empty, these users will be created:

- **Admin**: `admin@example.com`
- **Doctor**: `doctor@example.com`
- **Pharmacist**: `pharmacist@example.com`

Password:
- If `SEED_DEFAULT_PASSWORD` is set, all seeded users use that password.
- If not set (development only), a strong random password is generated and printed in the API logs.
