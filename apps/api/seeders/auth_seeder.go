package seeders

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedUsers() error {
	seedPassword := os.Getenv("SEED_DEFAULT_PASSWORD")
	if seedPassword == "" {
		seedPassword = "password123"
	}

	defaultEmail := os.Getenv("SEED_DEFAULT_EMAIL")
	if defaultEmail == "" {
		defaultEmail = "admin@indosupplier.local"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUser := &models.User{
		Email:     defaultEmail,
		Password:  string(hashedPassword),
		Name:      "Admin",
		AvatarURL: fmt.Sprintf("https://api.dicebear.com/7.x/lorelei/svg?seed=%s", defaultEmail),
		Status:    "active",
	}

	var existing models.User
	if err := database.DB.Where("email = ?", defaultEmail).First(&existing).Error; err == nil {
		return nil
	}

	return database.DB.Create(adminUser).Error
}
