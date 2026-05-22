package database

import (
	"fmt"

	core "github.com/gilabs/indosupplier/api/internal/core/data/models"
	refreshToken "github.com/gilabs/indosupplier/api/internal/refresh_token/data/models"
	user "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

// AutoMigrate runs minimal migrations for the cleaned baseline project.
func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&user.User{},
		&refreshToken.RefreshToken{},
		&core.AuditLog{},
		&core.TimeZone{},
		&core.Country{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
