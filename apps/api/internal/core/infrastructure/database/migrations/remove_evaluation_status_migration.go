package migrations

import (
	"gorm.io/gorm"
)

// RemoveEvaluationStatusColumn removes the status column from employee_evaluations table
func RemoveEvaluationStatusColumn(db *gorm.DB) error {
	// Check if column exists first
	if db.Migrator().HasColumn(&EmployeeEvaluationMigration{}, "status") {
		if err := db.Migrator().DropColumn(&EmployeeEvaluationMigration{}, "status"); err != nil {
			return err
		}
	}
	return nil
}

// EmployeeEvaluationMigration is a temporary struct for migration purposes only
type EmployeeEvaluationMigration struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	Status string `gorm:"type:varchar(20)"`
}

func (EmployeeEvaluationMigration) TableName() string {
	return "employee_evaluations"
}
