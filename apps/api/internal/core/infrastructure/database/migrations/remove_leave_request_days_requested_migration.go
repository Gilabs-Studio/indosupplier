package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// LeaveRequestMigration is a temporary struct for migration purposes only
type LeaveRequestMigration struct {
	ID            string `gorm:"type:char(36);primaryKey"`
	DaysRequested int    `gorm:"not null"`
}

func (LeaveRequestMigration) TableName() string {
	return "leave_requests"
}

// RemoveLeaveRequestDaysRequestedMigration removes the days_requested column from leave_requests table
// WHY: Consolidate to using TotalDays only with inclusive calendar days calculation
func RemoveLeaveRequestDaysRequestedMigration(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Check if column exists
	if db.Migrator().HasColumn(&LeaveRequestMigration{}, "days_requested") {
		if err := db.Migrator().DropColumn(&LeaveRequestMigration{}, "days_requested"); err != nil {
			return fmt.Errorf("failed to drop days_requested column: %w", err)
		}
		fmt.Println("Successfully removed days_requested column from leave_requests table")
	} else {
		fmt.Println("Leave requests days_requested column does not exist, skipping migration")
	}

	return nil
}
