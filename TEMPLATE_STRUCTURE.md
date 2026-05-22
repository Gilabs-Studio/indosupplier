# Template Structure Documentation

Dokumentasi ini menjelaskan struktur project template yang hanya berisi **Auth Feature** dan **Reusable Components**. Project ini dapat digunakan sebagai basis untuk membangun aplikasi baru.

## 📁 Struktur Project

```
ticketing-konser/
├── apps/
│   ├── api/              # Backend API (Go + Gin)
│   └── web/              # Frontend (Next.js 16 + React)
├── packages/             # Shared packages
└── docs/                 # Dokumentasi
```

---

## 🔧 Backend API (`apps/api/`)

### Struktur Folder

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point aplikasi
├── internal/
│   ├── api/
│   │   ├── handlers/                 # HTTP handlers
│   │   │   └── auth_handler.go       # Handler untuk authentication
│   │   ├── middleware/               # HTTP middleware
│   │   │   ├── auth.go               # JWT authentication middleware
│   │   │   ├── cors.go               # CORS middleware
│   │   │   ├── logger.go              # Request logging middleware
│   │   │   └── request_id.go         # Request ID middleware
│   │   └── routes/                    # Route definitions
│   │       └── auth_routes.go        # Auth routes setup
│   ├── config/
│   │   └── config.go                 # Configuration management
│   ├── database/
│   │   └── database.go                # Database connection & migration
│   ├── domain/                        # Domain entities (business models)
│   │   ├── auth/
│   │   │   └── entity.go             # Auth domain entity
│   │   ├── role/
│   │   │   └── entity.go             # Role domain entity
│   │   └── user/
│   │       └── entity.go             # User domain entity
│   ├── repository/
│   │   ├── interfaces/                # Repository interfaces
│   │   │   ├── auth_repository.go    # Auth repository interface
│   │   │   ├── role_repository.go    # Role repository interface
│   │   │   └── user_repository.go    # User repository interface
│   │   └── postgres/                  # PostgreSQL implementations
│   │       ├── auth/
│   │       │   └── repository.go     # Auth repository implementation
│   │       ├── role/
│   │       │   └── repository.go     # Role repository implementation
│   │       └── user/
│   │           └── repository.go     # User repository implementation
│   └── service/                       # Business logic layer
│       ├── auth/
│       │   └── service.go             # Auth service (login, logout, refresh)
│       └── user/
│           └── service.go             # User service (CRUD operations)
├── pkg/                               # Shared packages
│   ├── errors/
│   │   └── errors.go                 # Error handling utilities
│   ├── jwt/
│   │   └── jwt.go                    # JWT token management
│   ├── logger/
│   │   └── logger.go                 # Logger configuration
│   └── response/
│       └── response.go                # Standard API response format
└── seeders/                           # Database seeders
    ├── auth_seeder.go                 # Seed roles and users
    ├── helpers.go                     # Seeder helper functions
    └── seed_all.go                    # Main seeder function
```

### File Penting

#### 1. `cmd/server/main.go`

**Fungsi**: Entry point aplikasi
**Isi**:

- Initialize logger, config, database
- Setup JWT manager
- Setup repositories, services, handlers
- Setup router dengan middleware
- Start HTTP server

#### 2. `internal/api/routes/auth_routes.go`

**Fungsi**: Define auth routes
**Routes**:

- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user (requires auth)

#### 3. `internal/domain/auth/entity.go`

**Fungsi**: Auth domain model
**Struktur**:

- `User` - User entity dengan email, password, name, role, status
- `LoginRequest` - DTO untuk login request
- `LoginResponse` - DTO untuk login response (user, token, refresh_token)

#### 4. `internal/service/auth/service.go`

**Fungsi**: Business logic untuk authentication
**Methods**:

- `Login(req *LoginRequest)` - Authenticate user, return JWT tokens
- `RefreshToken(refreshToken string)` - Generate new access token
- `Logout(userID string)` - Invalidate tokens

#### 5. `internal/database/database.go`

**Fungsi**: Database connection & migration
**Isi**:

- `Connect()` - Connect ke PostgreSQL
- `AutoMigrate()` - Auto migrate tables (users, roles)
- `DropAllTables()` - Drop all tables (development only, dengan safety checks)

#### 6. `seeders/auth_seeder.go`

**Fungsi**: Seed initial data
**Isi**:

- `SeedRoles()` - Seed roles (admin, doctor, pharmacist, viewer)
- `SeedUsers()` - Seed users dengan default password "admin123"

---

## 🎨 Frontend Web (`apps/web/`)

### Struktur Folder

```
apps/web/
├── app/                               # Next.js App Router
│   ├── [locale]/                      # Internationalized routes
│   │   ├── layout.tsx                 # Locale layout wrapper
│   │   ├── page.tsx                   # Home page (redirects to login)
│   │   └── login/
│   │       └── page.tsx                # Login page
│   ├── layout.tsx                     # Root layout
│   ├── page.tsx                       # Root page
│   └── not-found.tsx                  # 404 page
├── src/
│   ├── components/                     # Reusable UI components
│   │   ├── ui/                        # shadcn/ui components
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   └── ... (36 UI components)
│   │   ├── layouts/
│   │   │   ├── app-layout.tsx         # Main app layout wrapper
│   │   │   └── sidebar-wrapper.tsx    # Sidebar component
│   │   ├── navigation/
│   │   │   ├── breadcrumb.tsx
│   │   │   ├── main-nav.tsx
│   │   │   └── sidebar.tsx
│   │   ├── providers/
│   │   │   └── theme-provider.tsx     # Theme provider (dark/light mode)
│   │   ├── error-boundary.tsx         # Error boundary component
│   │   └── loading.tsx                # Loading component
│   ├── features/                      # Feature modules
│   │   └── auth/                      # Authentication feature
│   │       ├── components/
│   │       │   ├── auth-guard.tsx     # Route protection component
│   │       │   ├── auth-layout.tsx    # Auth page layout
│   │       │   ├── login-form.tsx     # Login form component
│   │       │   └── permission-guard.tsx # Permission guard (simplified)
│   │       ├── hooks/
│   │       │   ├── useAuthGuard.ts    # Hook untuk route protection
│   │       │   ├── useLogin.ts        # Hook untuk login logic
│   │       │   ├── useLogout.ts       # Hook untuk logout logic
│   │       │   └── useRefreshSession.ts # Hook untuk refresh token
│   │       ├── schemas/
│   │       │   └── login.schema.ts    # Zod schema untuk login form
│   │       ├── services/
│   │       │   └── authService.ts     # API client untuk auth endpoints
│   │       ├── stores/
│   │       │   └── useAuthStore.ts    # Zustand store untuk auth state
│   │       └── types/
│   │           ├── errors.d.ts       # Auth error types
│   │           └── index.d.ts         # Auth types
│   ├── hooks/                         # Global hooks
│   │   ├── use-debounce.ts
│   │   └── use-mobile.ts
│   ├── i18n/                          # Internationalization
│   │   ├── request.ts                 # i18n request config
│   │   └── routing.ts                 # i18n routing config
│   ├── lib/                           # Utility libraries
│   │   ├── api-client.ts              # Axios client dengan interceptors
│   │   ├── react-query.tsx           # TanStack Query provider
│   │   ├── utils.ts                  # Utility functions
│   │   ├── badge-variant.ts
│   │   ├── icon-utils.tsx
│   │   └── menu-icons.tsx
│   └── types/                         # Global types
│       └── locale.d.ts                # Locale type definition
├── public/                            # Static assets
│   ├── avatar-placeholder.svg
│   └── login.webp
├── components.json                    # shadcn/ui config
├── next.config.ts                     # Next.js configuration
├── tailwind.config.ts                 # Tailwind CSS configuration
└── tsconfig.json                      # TypeScript configuration
```

### File Penting

#### 1. `app/[locale]/login/page.tsx`

**Fungsi**: Login page
**Isi**: Render `LoginForm` component dengan `AuthLayout`

#### 2. `src/features/auth/components/login-form.tsx`

**Fungsi**: Login form component
**Isi**:

- Form dengan email & password fields
- Validation menggunakan Zod schema
- Handle submit dengan `useLogin` hook
- Error handling & loading states

#### 3. `src/features/auth/stores/useAuthStore.ts`

**Fungsi**: Auth state management (Zustand)
**State**:

- `user` - Current user data
- `token` - Access token
- `refreshToken` - Refresh token
- `isAuthenticated` - Auth status
- `isLoading` - Loading state
- `error` - Error message

#### 4. `src/features/auth/services/authService.ts`

**Fungsi**: API client untuk auth endpoints
**Methods**:

- `login(credentials)` - POST /api/v1/auth/login
- `refreshToken(token)` - POST /api/v1/auth/refresh
- `logout()` - POST /api/v1/auth/logout

#### 5. `src/features/auth/hooks/useLogin.ts`

**Fungsi**: Login business logic hook
**Isi**:

- Handle form submission
- Call authService.login()
- Update auth store
- Set tokens di localStorage & cookies
- Redirect setelah login berhasil

#### 6. `src/components/ui/`

**Fungsi**: Reusable UI components (shadcn/ui)
**Components**: 36+ components termasuk Button, Input, Card, Dialog, Table, dll.

---

## 🔐 Authentication Flow

### Login Flow

1. User mengisi form di `/login`
2. `LoginForm` component memanggil `useLogin` hook
3. `useLogin` memanggil `authService.login()`
4. API mengembalikan user data + JWT tokens
5. Tokens disimpan di:
   - Zustand store (in-memory)
   - localStorage (persist)
   - Cookies (untuk middleware)
6. User di-redirect ke home page

### Protected Routes

- `AuthGuard` component mengecek `isAuthenticated` dari store
- Jika tidak authenticated, redirect ke `/login`
- Jika authenticated, render children

### Token Refresh

- `useRefreshSession` hook otomatis refresh token sebelum expired
- Menggunakan `refreshToken` dari localStorage
- Update store dengan token baru

---

## 📦 Dependencies

### Backend (Go)

- `github.com/gin-gonic/gin` - HTTP framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/google/uuid` - UUID generation

### Frontend (TypeScript/React)

- `next` - Next.js framework
- `react` - React library
- `zustand` - State management
- `@tanstack/react-query` - Server state management
- `zod` - Schema validation
- `react-hook-form` - Form handling
- `axios` - HTTP client
- `next-intl` - Internationalization
- `tailwindcss` - CSS framework
- `shadcn/ui` - UI component library

---

## 🚀 Getting Started

### 1. Setup Environment Variables

**Backend** (`apps/api/.env`):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ticketing_konser
JWT_SECRET=your-secret-key
JWT_ACCESS_TOKEN_TTL=24
JWT_REFRESH_TOKEN_TTL=168
SERVER_PORT=8080
ENV=development
```

**Frontend** (`apps/web/.env.local`):

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 2. Install Dependencies

```bash
pnpm install
```

### 3. Run Database Migrations

```bash
cd apps/api
go run cmd/server/main.go
# Migrations akan otomatis berjalan
```

### 4. Start Development Servers

```bash
# Root directory
pnpm dev

# Atau secara terpisah:
# Terminal 1 - Backend
cd apps/api
go run cmd/server/main.go

# Terminal 2 - Frontend
cd apps/web
pnpm dev
```

### 5. Access Application

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Login dengan:
  - Email: `admin@example.com`
  - Password: `admin123`

---

## 📝 Notes

1. **Hanya Auth Feature**: Project ini hanya berisi authentication, tidak ada features lain (dashboard, CRUD, dll)
2. **Reusable Components**: Semua UI components di `src/components/ui/` dapat digunakan untuk features baru
3. **Clean Architecture**: Backend menggunakan layered architecture (Handler → Service → Repository)
4. **Type Safety**: Frontend menggunakan TypeScript dengan strict mode
5. **Internationalization**: Support untuk multiple locales (id, en)

---

## 🔄 Menambahkan Feature Baru

### Backend

1. Buat domain entity di `internal/domain/{feature}/entity.go`
2. Buat repository interface di `internal/repository/interfaces/{feature}_repository.go`
3. Implement repository di `internal/repository/postgres/{feature}/repository.go`
4. Buat service di `internal/service/{feature}/service.go`
5. Buat handler di `internal/api/handlers/{feature}_handler.go`
6. Setup routes di `internal/api/routes/{feature}_routes.go`
7. Register routes di `cmd/server/main.go`

### Frontend

1. Buat feature folder di `src/features/{feature}/`
2. Struktur:
   - `components/` - UI components
   - `hooks/` - Business logic hooks
   - `services/` - API clients
   - `stores/` - State management (jika perlu)
   - `types/` - Type definitions
   - `schemas/` - Zod schemas (jika ada form)
3. Buat route di `app/[locale]/{feature}/page.tsx`

---

## 📚 Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Zustand](https://zustand-demo.pmnd.rs/)
- [TanStack Query](https://tanstack.com/query/latest)
- [shadcn/ui](https://ui.shadcn.com/)
