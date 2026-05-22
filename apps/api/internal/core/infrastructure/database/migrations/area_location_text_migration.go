package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// EnsureAreaLocationTextMigration relaxes the area location columns so they can
// store free-form values without arbitrary length constraints.
func EnsureAreaLocationTextMigration(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	statements := []string{
		`ALTER TABLE areas ALTER COLUMN province TYPE TEXT USING province::text`,
		`ALTER TABLE areas ALTER COLUMN regency TYPE TEXT USING regency::text`,
		`ALTER TABLE areas ALTER COLUMN district TYPE TEXT USING district::text`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("failed to relax area location column: %w", err)
		}
	}

	return nil
}
