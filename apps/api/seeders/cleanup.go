package seeders

import (
	"fmt"
	"log"
	"os"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// CleanupDatabase truncates all tables except system tables
func CleanupDatabase() error {
	// Safety check: Never allow in production
	env := ""
	if config.AppConfig != nil {
		env = config.AppConfig.Server.Env
	}
	if env == "" {
		env = os.Getenv("APP_ENV")
	}

	if env == "production" || env == "prod" {
		log.Println("🔒 Production mode detected: Cleanup is disabled (safety protection)")
		return nil
	}

	log.Println("🧹 Cleaning up database tables...")

	// Get all table names in the current schema
	var tables []string
	err := database.DB.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = CURRENT_SCHEMA()
		AND tablename NOT LIKE 'pg_%'
		AND tablename NOT LIKE '_prisma_%'
		AND tablename != 'migrations'
	`).Scan(&tables).Error

	if err != nil {
		return fmt.Errorf("failed to get table list: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	// Truncate all tables
	for _, table := range tables {
		truncateSQL := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if err := database.DB.Exec(truncateSQL).Error; err != nil {
			log.Printf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	log.Println("✅ Database tables cleaned up successfully")
	return nil
}
