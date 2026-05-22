# GIMS - GILABS Integrated Management System

Monorepo untuk GIMS (GILABS Integrated Management System) menggunakan Turborepo.

## Tech Stack

### Backend

- **Go 1.25+** dengan **Gin Framework**
- **PostgreSQL** (akan ditambahkan)
- **Docker** untuk containerization

### Frontend

- **Next.js 16** (App Router, Server Components)
- **Tailwind CSS v4**
- **TypeScript**
- **Zustand** untuk state management

### Monorepo

- **Turborepo** untuk build system
- **pnpm** untuk package management

## Struktur Project

```
gims/
├── apps/
│   ├── api/          # Go API backend (Gin)
│   └── web/          # Next.js frontend
├── packages/
│   ├── eslint-config/      # Shared ESLint config
│   └── typescript-config/   # Shared TypeScript config
└── docs/             # Dokumentasi
```

## Getting Started

### Prerequisites

- **Node.js** >= 18
- **pnpm** >= 9.0.0
- **Go** >= 1.25
- **Docker** (optional, untuk development)

### Install Dependencies

```bash
pnpm install
```

### Development

Jalankan semua apps dalam mode development:

```bash
pnpm dev
```

Jalankan app tertentu:

```bash
# Frontend only
pnpm dev --filter=web

# API only
pnpm dev --filter=api

# API + Frontend
pnpm dev --filter=web --filter=api
```

### Build

Build semua apps:

```bash
pnpm build
```

Build app tertentu:

```bash
pnpm build --filter=web
pnpm build --filter=api
```

### Lint & Type Check

```bash
# Lint semua apps
pnpm lint

# Type check semua apps
pnpm check-types
```

## Apps

### API (`apps/api`)

Backend API menggunakan Go dan Gin framework.

**Development:**

```bash
cd apps/api
pnpm dev
# atau
go run cmd/server/main.go
```

**Build:**

```bash
cd apps/api
pnpm build
# atau
go build -o bin/server ./cmd/server/main.go
```

**Docker:**

```bash
cd apps/api
docker-compose up --build
```

Server akan berjalan di `http://localhost:8080`

**Endpoints:**

- `GET /health` - Health check
- `GET /ping` - Ping endpoint
- `GET /api/v1/` - API version info

**Menjalankan API via Docker + Web dari root:**

Di root project (bukan di folder lain), tersedia script helper:

```bash
cd D:\Files\Documents\Pekerjaan\Gilabs\gims
pnpm run dev:web-api-docker
```

Script ini akan:

- `cd apps/api && docker compose down` (reset container)
- `docker compose up -d` (menjalankan API + database di background)
- `cd ../web && pnpm dev` (menjalankan web Next.js dalam mode development)

### Web (`apps/web`)

Frontend menggunakan Next.js 16.

**Development:**

```bash
cd apps/web
pnpm dev
```

Server akan berjalan di `http://localhost:3000`

## Packages

### `@repo/eslint-config`

Shared ESLint configuration untuk semua apps.

### `@repo/typescript-config`

Shared TypeScript configuration untuk semua apps.

## Documentation


## Agent Skills

- **k6 Skill**: User-level skill for implementing, documenting, and maintaining k6 transaction workflow tests, with Gill Me inquiry and clean architecture documentation. See `.agents/skills/k6/` and `k6/README_SKILL.md` for details.
## Development Guidelines

### Backend (Go)

- Follow Go best practices dan conventions
- Gunakan API response standards yang sudah didefinisikan
- Implement error codes sesuai dokumentasi
- Support bilingual error messages (ID & EN)

### Frontend (Next.js)

- Follow Next.js 16 App Router best practices
- Gunakan Server Components by default
- State management dengan Zustand
- Styling dengan Tailwind CSS v4
- Components menggunakan shadcn/ui v4

### Monorepo

- Gunakan Turborepo untuk build orchestration
- Shared configs di `packages/`
- Apps di `apps/`
- Dokumentasi di `docs/`

## Scripts

- `pnpm dev` - Run all apps in development mode
- `pnpm build` - Build all apps
- `pnpm lint` - Lint all apps
- `pnpm check-types` - Type check all apps
- `pnpm test` - Run tests
- `pnpm clean` - Clean build artifacts

## Remote Caching

Turborepo dapat menggunakan Remote Caching untuk share build cache dengan team dan CI/CD.

Setup:

```bash
turbo login
turbo link
```

## Useful Links

- [Turborepo Docs](https://turborepo.com/docs)
- [Next.js 16 Docs](https://nextjs.org/docs)
- [Gin Framework Docs](https://gin-gonic.com/docs/)
- [Tailwind CSS v4 Docs](https://tailwindcss.com/docs)
