package seeders

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedUsers() error {
	const (
		seedPassword = "admin123"
		defaultEmail = "admin@example.com"
	)

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
