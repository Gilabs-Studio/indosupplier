# Frontend Knowledge Base

**Location:** `apps/web/`
**Stack:** Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4, shadcn/ui

## Architecture

Feature-based structure:

```
src/features/<feature>/
├── types/           # `.d.ts` declarations only
├── schemas/         # Zod schemas (`<feature>.schema.ts`)
├── stores/          # Zustand (`use<Domain>Store.ts`) — STATE ONLY
├── hooks/           # Business logic (`use<Action>.ts`) — ALL LOGIC HERE
├── services/        # API calls (`<feature>-service.ts`)
├── components/      # UI only — NO business logic
└── i18n/            # `en.ts`, `id.ts`
```

## Critical Rules

### Component Logic Separation

**NEVER** put business logic in components. Extract to hooks.

```tsx
// ✅ GOOD — hook handles logic
const { mutate: login, isPending } = useLogin();
return <form onSubmit={handleSubmit((d) => login(d))}>...</form>;
```

### Type Safety

- **NEVER** use `any`. Use `unknown` with type guards.
- Always define return types.
- Use optional chaining (`?.`) and nullish coalescing (`??`).

### i18n

- Global: `src/i18n/messages/{en,id}.json`
- Feature: `src/features/<feature>/i18n/{en,id}.ts`
- Register feature translations in `src/i18n/request.ts`
- Navigation: `import { Link } from "@/i18n/routing"` (NOT `next/link`)

### UI/UX Safety

- Handle `isLoading`, `isError`, empty states from TanStack Query
- Add `cursor-pointer` to ALL clickable elements
- Use `PageMotion` for page transitions
- Never use arbitrary Tailwind values (`bg-[#e53e3e]`)

## Performance (Critical for Dashboard Apps)

1. **Default to Server Components** — Client Components only when interactivity needed
2. **Use `next/link`** for all navigation (enables prefetch)
3. **Route-level `loading.tsx`** for every page — use `PageMotion` + Skeleton
4. **Dynamic import** for heavy components (charts, maps, editors)
5. **Lazy load tabs** — only render active `TabsContent`
6. **Cache server data** — `fetch` with `revalidate` or `unstable_cache`
7. **Avoid global state for server data** — use Server Components + cache

## Adding a New Feature

1. Create `src/features/<feature>/` with full structure
2. Add route: `app/[locale]/(dashboard)/<feature>/page.tsx`
3. Add `loading.tsx` for route-level loading
4. Register route in `src/lib/route-validator.ts`
5. Register i18n in `src/i18n/request.ts`

## Commands

```bash
cd apps/web

# Dev
pnpm dev                          # localhost:3000

# Build
pnpm build
pnpm analyze                      # Bundle analysis

# Quality
pnpm lint
pnpm check-types
```

## Anti-Patterns

- Business logic in components (use hooks)
- `any` type anywhere
- Arbitrary Tailwind values
- Missing `cursor-pointer` on interactive elements
- `useEffect + fetch` for initial data (use Server Components)
- Full-page Client Components (split into Server + small Client wrappers)
