package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// AddPerformanceIndexesMigration adds composite indexes for slow list endpoints
// identified by k6 endpoint profiler at 100 VUs.
func AddPerformanceIndexesMigration(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	indexes := []struct {
		name  string
		query string
	}{
		{
			name:  "idx_sales_orders_tenant_date_status",
			query: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sales_orders_tenant_date_status ON sales_orders(tenant_id, order_date DESC, status) WHERE deleted_at IS NULL`,
		},
		{
			name:  "idx_sales_orders_tenant_created",
			query: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sales_orders_tenant_created ON sales_orders(tenant_id, created_at DESC) WHERE deleted_at IS NULL`,
		},
		{
			name:  "idx_loyalty_members_tenant_status_created",
			query: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_loyalty_members_tenant_status_created ON loyalty_members(tenant_id, status, created_at DESC)`,
		},
	}

	for _, idx := range indexes {
		// Check if index already exists
		var count int64
		db.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", idx.name).Scan(&count)
		if count > 0 {
			fmt.Printf("Index %s already exists, skipping\n", idx.name)
			continue
		}

		// CONCURRENTLY cannot run inside a transaction, use Exec directly
		if err := db.Exec(idx.query).Error; err != nil {
			fmt.Printf("Warning: Failed to create index %s: %v\n", idx.name, err)
			// Don't fail migration — index creation is non-critical
			continue
		}
		fmt.Printf("Created index: %s\n", idx.name)
	}

	return nil
}
