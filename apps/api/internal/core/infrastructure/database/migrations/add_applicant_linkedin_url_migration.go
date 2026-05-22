package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// RecruitmentApplicantMigration is a temporary struct for migration purposes only
type RecruitmentApplicantMigration struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	LinkedinURL string `gorm:"type:varchar(500)"`
}

func (RecruitmentApplicantMigration) TableName() string {
	return "recruitment_applicants"
}

// AddApplicantLinkedInURLMigration adds the linkedin_url column to recruitment_applicants table
func AddApplicantLinkedInURLMigration(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Check if column exists
	if !db.Migrator().HasColumn(&RecruitmentApplicantMigration{}, "linkedin_url") {
		if err := db.Migrator().AddColumn(&RecruitmentApplicantMigration{}, "linkedin_url"); err != nil {
			return fmt.Errorf("failed to add linkedin_url column: %w", err)
		}
		fmt.Println("Successfully added linkedin_url column to recruitment_applicants table")
	} else {
		fmt.Println("linkedin_url column already exists, skipping migration")
	}

	return nil
}
