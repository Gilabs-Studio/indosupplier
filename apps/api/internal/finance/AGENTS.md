# Finance Domain Knowledge Base

**Location:** `apps/api/internal/finance/`
**Files:** ~171 Go files
**Sub-packages:** `accounting`, `financesettings`, `reference`, `service`

## Structure

```
finance/
├── data/
│   ├── models/         # Asset, Account, JournalEntry, etc.
│   └── repositories/   # Finance-specific repositories
├── domain/
│   ├── accounting/     # Sub-ledger, GL logic
│   ├── dto/            # Request/Response DTOs
│   ├── financesettings/# Company finance settings
│   ├── mapper/         # Model ↔ DTO
│   ├── reference/      # Reference data helpers
│   ├── service/        # Domain services
│   └── usecase/        # Business logic
└── presentation/
    ├── handler/
    └── router/
```

## Key Conventions

### Accounting Sub-package

`domain/accounting/` contains core accounting logic:
- General ledger entries
- Sub-ledger reconciliation
- Balance calculations

Use this for cross-cutting accounting operations, not usecase-specific logic.

### Reference Data

`domain/reference/` holds reference tables and lookups:
- Account categories
- Transaction types
- Fiscal periods

### Complex Usecases

Finance has some of the largest usecases in the project (e.g., `asset_usecase.go` ~1373 lines). Keep usecases focused:
- One usecase per entity/aggregate
- Extract shared calculation logic into `domain/service/` or `domain/accounting/`

## Where to Look

| Task | Location |
|------|----------|
| Add journal entry | `domain/usecase/journal_*.go` |
| Asset management | `domain/usecase/asset_usecase.go` |
| Account/COA | `data/models/account.go`, `domain/accounting/` |
| Company settings | `domain/financesettings/` |

## Anti-Patterns

- Duplicating accounting calculations across usecases
- Putting GL posting logic directly in entity usecases
- Forgetting to use transactions for multi-leg journal entries
