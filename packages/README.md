# Packages

Folder `packages/` berisi shared configurations dan utilities yang digunakan oleh semua apps dalam monorepo.

## Struktur

### `@repo/eslint-config`

Shared ESLint configuration untuk semua apps.

**Usage:**

```json
{
  "devDependencies": {
    "@repo/eslint-config": "workspace:*"
  }
}
```

**Configs:**

- `base.js` - Base ESLint config
- `next.js` - Next.js specific config
- `react-internal.js` - React internal config

### `@repo/typescript-config`

Shared TypeScript configuration untuk semua apps.

**Usage:**

```json
{
  "devDependencies": {
    "@repo/typescript-config": "workspace:*"
  }
}
```

**Configs:**

- `base.json` - Base TypeScript config
- `nextjs.json` - Next.js specific config
- `react-library.json` - React library config

## Mengapa Packages?

Packages digunakan untuk:

1. **Code Sharing**: Share configurations, utilities, dan types antar apps
2. **Consistency**: Memastikan semua apps menggunakan config yang sama
3. **Maintainability**: Update config di satu tempat, semua apps terupdate
4. **Type Safety**: Shared types untuk API contracts

## Catatan

- **UI Components**: UI components sekarang berada langsung di `apps/web/src/components/` karena hanya digunakan oleh web app
- Jika di masa depan ada kebutuhan untuk share UI components antar apps, bisa dibuat `packages/ui` kembali
