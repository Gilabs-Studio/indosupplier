# Backfill Snapshots Tool

This tool backfills **snapshot/denormalized fields** for existing transactional data (Purchase + Finance) so historical documents remain **immutable** when related master data changes (supplier/product/payment terms/COA/bank account/business unit).

> Runtime immutability is handled by snapshot hydration/preservation in the domain usecases.
> This tool is only to backfill **already-existing** rows.

## When to run

Run this **after** deploying code that adds snapshot columns (GORM AutoMigrate / migrations), especially when:

- You added new `*_snapshot` columns on transactional tables.
- You changed snapshot strategy (e.g., adding `payment_terms_days_snapshot`).
- You need to make old transactions display stable values in the UI.

## Safety / behavior

- The SQL steps are written to be **idempotent**: they update rows only when snapshot columns are empty/NULL.
- It is safe to run multiple times.
- The tool performs bulk `UPDATE ... FROM ...` statements; run during low-traffic maintenance windows for large databases.

## Where the SQL lives

Each step is stored as an ordered SQL file in:

- `cmd/tools/backfill-snapshots/sql/*.sql`

The Go program embeds these files (via `go:embed`) so execution order is explicit and the tool remains self-contained.

## Prerequisites

- Database is reachable and credentials are configured.
- Environment variables are loaded via `config.Load()`:
  - In non-production, `.env` is loaded automatically if present.
  - In production, use injected environment variables.

Relevant DB env vars (see `apps/api/SETUP.md`):

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

## Run

From `apps/api/`:

```bash
go run ./cmd/tools/backfill-snapshots
```

Expected output:

- `Backfill OK (<step>): rows_affected=<n>` for each step
- `Backfill snapshots completed`

If a step fails, the process exits immediately with:

- `Backfill failed (<step>): <error>`

## Notes

- This tool does not create schema/columns; it only fills data.
- If you are adding a **new** snapshot column, always ensure:
  1) schema column exists
  2) create/update usecases hydrate/preserve it
  3) mappers prefer snapshot values in API responses
  4) backfill fills existing rows (this tool)
