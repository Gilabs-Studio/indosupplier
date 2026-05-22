# HRD Domain Knowledge Base

**Location:** `apps/api/internal/hrd/`
**Files:** ~80 Go files
**Special:** Per-employee timezone, attendance rules, background workers

## Structure

```
hrd/
├── data/
│   ├── models/         # Employee, Attendance, LeaveRequest, etc.
│   └── repositories/   # HRD-specific repositories
├── domain/
│   ├── dto/            # Request/Response DTOs
│   ├── mapper/         # Model ↔ DTO
│   └── usecase/        # Business logic
├── presentation/
│   ├── handler/
│   └── router/
└── worker/             # Background workers (attendance processing, etc.)
```

## Critical Rules

### Timezone (HRD-specific)

HRD **must** use per-employee/company timezone functions:

```go
now := apptime.NowForEmployee(employeeID)      // Attendance, leave, overtime
loc := apptime.LocationForEmployee(employeeID) // Date math
```

- DB columns: `timestamptz` (not `timestamp`)
- Holiday queries: use `IsHolidayForCompany()`, `FindByDateRangeForCompany()`
- Holiday `CompanyID`: NULL = global, non-NULL = company-specific

### Attendance Check-in Validation

Employee can only check in **at or after** scheduled start time:

```go
var earliestCheckInTime string
if ws.IsFlexible && ws.FlexibleStartTime != "" {
    earliestCheckInTime = ws.FlexibleStartTime
} else {
    earliestCheckInTime = ws.StartTime
}

if now.Before(earliestCheckInToday) {
    return fmt.Errorf("TOO_EARLY_TO_CHECK_IN: Cannot check in before %s", earliestCheckInTime)
}
```

Frontend must also disable the button (dual validation).

### Workers

`worker/` contains background processors:
- Attendance snapshot/backfill
- Leave balance recalculation
- Scheduled report generation

Workers run independently of HTTP handlers. Use proper graceful shutdown.

## Where to Look

| Task | Location |
|------|----------|
| Attendance | `domain/usecase/attendance_*.go`, `worker/` |
| Leave requests | `domain/usecase/leave_*.go` |
| Employee data | `data/models/employee.go` |
| Work schedule | `data/models/work_schedule.go` |

## Anti-Patterns

- Using `time.Now()` instead of `apptime.NowForEmployee()`
- Hardcoding `Asia/Jakarta` timezone
- Skipping company filter on holiday queries
- Missing backend validation for check-in time
