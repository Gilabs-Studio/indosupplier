package seeders

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/models"
)

func SeedSystemAdmins() error {
	seedPassword := os.Getenv("SEED_DEFAULT_PASSWORD")
	if seedPassword == "" {
		seedPassword = "password123"
	}

	defaultEmail := "sysadmin@indosupplier.local"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.SystemAdmin{
		Email:     defaultEmail,
		Password:  string(hashedPassword),
		Name:      "System Admin",
		Role:      "super_admin",
		Status:    "active",
	}

	var existing models.SystemAdmin
	if err := database.DB.Where("email = ?", defaultEmail).First(&existing).Error; err == nil {
		return nil
	}

	fmt.Printf("seeding default system admin: %s\n", defaultEmail)
	return database.DB.Create(admin).Error
}
