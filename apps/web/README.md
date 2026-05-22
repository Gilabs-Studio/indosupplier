# SalesView Web

Frontend application untuk SalesView menggunakan Next.js 16.

## Tech Stack

- **Next.js**: 16.0.3 (App Router, Server Components)
- **React**: 19.2.0
- **TypeScript**: 5.x
- **Tailwind CSS**: v4
- **shadcn/ui**: v4 components
- **Zustand**: State management
- **TanStack Query**: Server state management
- **React Hook Form**: Form handling
- **Zod**: Schema validation
- **Axios**: HTTP client

## Getting Started

### Prerequisites

- Node.js 18+
- pnpm 9.0.0

### Installation

1. Install dependencies:

```bash
pnpm install
```

2. Copy environment variables:

```bash
cp .env.example .env
```

3. Update `.env` dengan konfigurasi yang sesuai:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Development

Run the development server:

```bash
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser.

### Build

```bash
pnpm build
```

### Start Production Server

```bash
pnpm start
```

## Environment Variables

### Required Variables

- `NEXT_PUBLIC_API_URL`: Base URL untuk API backend (default: `http://localhost:8080`)

### Optional Variables

- `NODE_ENV`: Environment mode (`development`, `production`, `test`)

**Note**: Di Next.js, hanya variable yang dimulai dengan `NEXT_PUBLIC_` yang bisa diakses di client-side code.

## Project Structure

```
apps/web/
├── app/                          # Next.js App Router
│   ├── dashboard/               # Protected dashboard page
│   ├── layout.tsx               # Root layout
│   ├── page.tsx                 # Login page (root route)
│   └── globals.css              # Global styles
├── src/
│   ├── components/
│   │   ├── ui/                  # shadcn/ui components
│   │   ├── error-boundary.tsx   # Error boundary
│   │   └── loading.tsx          # Loading skeletons
│   ├── features/
│   │   └── auth/
│   │       ├── components/      # Auth components
│   │       ├── hooks/           # Auth hooks
│   │       ├── schemas/         # Zod schemas
│   │       ├── services/        # API services
│   │       ├── stores/          # Zustand stores
│   │       └── types/           # TypeScript types
│   └── lib/
│       ├── api-client.ts        # Axios client
│       ├── react-query.tsx      # TanStack Query provider
│       └── utils.ts             # Utilities
└── public/                      # Static assets
```

## Features

### Authentication

- Login dengan email/password
- JWT token management
- Protected routes dengan AuthGuard
- Auto token refresh
- Auto redirect on 401

### Form Handling

- React Hook Form untuk form state
- Zod untuk validation
- Error display per field
- Loading states

### State Management

- **Zustand**: Client state (auth, UI state)
- **TanStack Query**: Server state (API data, caching)

### Error Handling

- Error Boundary untuk React errors
- API error handling dengan interceptors
- User-friendly error messages

## Development Guidelines

- Follow Next.js 16 App Router conventions
- Use Server Components by default
- Use Client Components only when needed (`"use client"`)
- Feature-based folder structure
- Type-safe dengan TypeScript
- Use Zod schemas for validation

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS v4](https://tailwindcss.com/blog/tailwindcss-v4)
- [shadcn/ui](https://ui.shadcn.com)
- [Zustand](https://zustand-demo.pmnd.rs)
- [TanStack Query](https://tanstack.com/query)
