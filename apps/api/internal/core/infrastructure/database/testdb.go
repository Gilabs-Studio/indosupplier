package database

import (
    "fmt"
    "hash/fnv"
    "os"
    "strings"
    "testing"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    roleModels "github.com/gilabs/gims/api/internal/role/data/models"
    userModels "github.com/gilabs/gims/api/internal/user/data/models"
    salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
    financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

// OpenTestDB returns a *gorm.DB suitable for integration tests. It respects
// the `TEST_DB` environment variable:
// - if TEST_DB=postgres and TEST_DATABASE_DSN is set, it opens Postgres
// - otherwise it falls back to an in-memory SQLite DB
// It also returns a cleanup function to call with defer.
func OpenTestDB(t *testing.T) (*gorm.DB, func()) {
    t.Helper()

    if os.Getenv("TEST_DB") == "postgres" {
        dsn := os.Getenv("TEST_DATABASE_DSN")
        if dsn == "" {
            t.Fatalf("TEST_DB=postgres but TEST_DATABASE_DSN is not set")
        }

        adminDSN := replaceDBName(dsn, "postgres")
        adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
        if err != nil {
            t.Fatalf("failed opening postgres admin db: %v", err)
        }

        testDBName := shortTestDBName(t.Name())
        if err := adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName)).Error; err != nil {
            t.Fatalf("failed creating postgres test db %s: %v", testDBName, err)
        }

        testDSN := replaceDBName(dsn, testDBName)
        db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
        if err != nil {
            _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName)).Error
            t.Fatalf("failed opening postgres test db: %v", err)
        }

        // Auto-migrate real Role, User, and key finance/sales models into the
        // test DB so GORM migrations and foreign keys match production schema.
        if err := db.AutoMigrate(
            &roleModels.Role{}, &userModels.User{},
            &financeModels.ChartOfAccount{}, &financeModels.JournalEntry{}, &financeModels.JournalLine{}, &financeModels.JournalReversal{}, &financeModels.FiscalYear{}, &financeModels.AccountingPeriod{}, &financeModels.ValuationRun{}, &financeModels.ValuationRunDetail{}, &financeModels.FinancialClosing{}, &financeModels.Budget{}, &financeModels.BudgetItem{}, &financeModels.AdjustmentJournalApproval{},
            &financeModels.Asset{}, &financeModels.AssetCategory{}, &financeModels.AssetLocation{}, &financeModels.AssetDepreciation{}, &financeModels.AssetDepreciationSchedule{}, &financeModels.AssetDisposal{}, &financeModels.AssetTransfer{}, &financeModels.AssetMaintenanceLog{}, &financeModels.AssetTransaction{}, &financeModels.AssetAttachment{}, &financeModels.AssetAuditLog{}, &financeModels.AssetAssignmentHistory{},
            &salesModels.CustomerInvoice{},
        ); err != nil {
            _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName)).Error
            t.Fatalf("failed auto-migrating role/user models: %v", err)
        }

        const testUUID = "11111111-1111-1111-1111-111111111111"
        const roleUUID = "22222222-2222-2222-2222-222222222222"

        // Seed a minimal protected role required by User.RoleID (do nothing if exists)
        if err := db.Exec(`INSERT INTO roles (id, name, code, status, is_protected, data_scope, created_at, updated_at) VALUES (?, ?, ?, 'active', false, 'ALL', now(), now()) ON CONFLICT (id) DO NOTHING`, roleUUID, "Test Role", "test_role").Error; err != nil {
            t.Fatalf("failed seeding test role: %v", err)
        }

        // Seed a minimal user with the required non-null fields
        if err := db.Exec(`INSERT INTO users (id, email, password, name, role_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'active', now(), now()) ON CONFLICT (id) DO NOTHING`, testUUID, "test@example.com", "test-password", "Test User", roleUUID).Error; err != nil {
            t.Fatalf("failed seeding test user: %v", err)
        }

        // Also seed the canonical system user ID used by many finance tests
        const systemUserID = "00000000-0000-0000-0000-000000000001"
        if err := db.Exec(`INSERT INTO users (id, email, password, name, role_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'active', now(), now()) ON CONFLICT (id) DO NOTHING`, systemUserID, "system@example.com", "test-password", "System User", roleUUID).Error; err != nil {
            t.Fatalf("failed seeding system test user: %v", err)
        }
        if err := db.Exec(`CREATE TABLE IF NOT EXISTS fiscal_years (
            id uuid PRIMARY KEY,
            tenant_id uuid,
            company_id uuid NOT NULL,
            name varchar(100) NOT NULL,
            start_date date NOT NULL,
            end_date date NOT NULL,
            status varchar(20) NOT NULL DEFAULT 'draft',
            created_by uuid,
            created_at timestamp,
            updated_at timestamp,
            deleted_at timestamp
        )`).Error; err != nil {
            t.Fatalf("failed creating test fiscal_years table: %v", err)
        }
        if err := db.Exec(`CREATE TABLE IF NOT EXISTS accounting_periods (
            id uuid PRIMARY KEY,
            tenant_id uuid,
            period_name varchar(50) NOT NULL,
            start_date date NOT NULL,
            end_date date NOT NULL,
            status varchar(20) NOT NULL DEFAULT 'open',
            locked_at timestamp,
            locked_by uuid,
            created_at timestamp,
            updated_at timestamp,
            deleted_at timestamp
        )`).Error; err != nil {
            t.Fatalf("failed creating test accounting_periods table: %v", err)
        }
        db.Callback().Create().Before("gorm:before_create").Register("testdb:fill_uuid_fields", func(tx *gorm.DB) {
            if tx.Statement == nil || tx.Statement.Schema == nil {
                return
            }

            for _, field := range tx.Statement.Schema.Fields {
                if field.Name == "ID" {
                    continue
                }
                if field.TagSettings["TYPE"] != "uuid" {
                    continue
                }

                if value, zero := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue); zero {
                    if stringValue, ok := value.(string); ok && stringValue == "" {
                        tx.Statement.SetColumn(field.Name, testUUID)
                    }
                }
            }
        })

        return db, func() {
            if sqlDB, err := db.DB(); err == nil {
                _ = sqlDB.Close()
            }
            _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName)).Error
            if sqlDB, err := adminDB.DB(); err == nil {
                _ = sqlDB.Close()
            }
        }
    }

    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed opening sqlite test db: %v", err)
    }
    return db, func() {}
}

func parseKeyValueDSN(dsn string) map[string]string {
    parts := strings.Fields(dsn)
    values := make(map[string]string, len(parts))
    for _, part := range parts {
        key, value, found := strings.Cut(part, "=")
        if found {
            values[key] = value
        }
    }
    return values
}

func replaceDBName(dsn, dbName string) string {
    parts := strings.Fields(dsn)
    for i, part := range parts {
        if strings.HasPrefix(part, "dbname=") {
            parts[i] = "dbname=" + dbName
        }
    }
    return strings.Join(parts, " ")
}

func shortTestDBName(value string) string {
    hash := fnv.New64a()
    _, _ = hash.Write([]byte(value + time.Now().Format(time.RFC3339Nano)))
    return fmt.Sprintf("t_%x", hash.Sum64())
}
