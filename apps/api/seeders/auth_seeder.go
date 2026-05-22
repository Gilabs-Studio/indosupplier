package seeders

import (
	"fmt"
	"log"
	"os"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	user "github.com/gilabs/gims/api/internal/user/data/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

func getSeedPassword() (string, error) {
	if pw := os.Getenv("SEED_DEFAULT_PASSWORD"); pw != "" {
		return pw, nil
	}

	// Production must never silently create accounts with unknown/weak password.
	if config.AppConfig != nil && config.AppConfig.Server.Env == "production" {
		return "", fmt.Errorf("SEED_DEFAULT_PASSWORD must be set when RUN_SEEDERS=true in production")
	}

	// For development, we use a predictable default if none is provided.
	// This avoids the need to log a "generated" password, which triggers security alerts
	// regarding clear-text logging of sensitive information.
	return "password123", nil
}

// SeedUsers seeds initial users
func SeedUsers() error {
	// Refactored to Upsert/Ensure existence of Fixed Users
	// We want to ensure Admin/Manager/Staff exist with these IDs.

	seedPassword, err := getSeedPassword()
	if err != nil {
		return err
	}
	// Only print password if we are actually going to try creating users (simplified check)
	// Actually we print it if env is empty, regardless of creation success/failure (maybe not ideal but okay).
	if os.Getenv("SEED_DEFAULT_PASSWORD") == "" {
		log.Printf("Using fixed default password for dev seeding.")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var defaultEmail = os.Getenv("SEED_DEFAULT_EMAIL")
	if defaultEmail == "" {
		defaultEmail = "admin@example.com"
	}

	users := []user.User{
		{
			ID:        AdminUserID,
			TenantID:  DefaultTenantID,
			Email:     defaultEmail,
			Password:  string(hashedPassword),
			Name:      "Admin User",
			AvatarURL: fmt.Sprintf("https://api.dicebear.com/7.x/lorelei/svg?seed=%s", defaultEmail),
			RoleID:    AdminRoleID,
			Status:    "active",
		},
		{
			ID:        ManagerUserID,
			TenantID:  DefaultTenantID,
			Email:     "manager@example.com",
			Password:  string(hashedPassword),
			Name:      "Manager User",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=manager@example.com",
			RoleID:    ManagerRoleID,
			Status:    "active",
		},
		{
			ID:        StaffUserID,
			TenantID:  DefaultTenantID,
			Email:     "staff@example.com",
			Password:  string(hashedPassword),
			Name:      "Staff User",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=staff@example.com",
			RoleID:    StaffRoleID,
			Status:    "active",
		},
		{
			ID:        ViewerUserID,
			TenantID:  DefaultTenantID,
			Email:     "viewer@example.com",
			Password:  string(hashedPassword),
			Name:      "Viewer User",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=viewer@example.com",
			RoleID:    ViewerRoleID,
			Status:    "active",
		},
		{
			ID:        OutletManager1UserID,
			TenantID:  DefaultTenantID,
			Email:     "outlet-manager1@example.com",
			Password:  string(hashedPassword),
			Name:      "Outlet Manager Sudirman",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=outlet-manager1@example.com",
			RoleID:    OutletManagerRoleID,
			Status:    "active",
		},
		{
			ID:        OutletManager2UserID,
			TenantID:  DefaultTenantID,
			Email:     "outlet-manager2@example.com",
			Password:  string(hashedPassword),
			Name:      "Outlet Manager Kelapa Gading",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=outlet-manager2@example.com",
			RoleID:    OutletManagerRoleID,
			Status:    "active",
		},
		{
			ID:        OutletManager3UserID,
			TenantID:  DefaultTenantID,
			Email:     "outlet-manager3@example.com",
			Password:  string(hashedPassword),
			Name:      "Outlet Manager Kimia Farma Sudirman",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=outlet-manager3@example.com",
			RoleID:    OutletManagerRoleID,
			Status:    "active",
		},
		{
			ID:        OutletManager4UserID,
			TenantID:  DefaultTenantID,
			Email:     "outlet-manager4@example.com",
			Password:  string(hashedPassword),
			Name:      "Outlet Manager Kimia Farma Pondok Indah",
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=outlet-manager4@example.com",
			RoleID:    OutletManagerRoleID,
			Status:    "active",
		},
	}

	for _, u := range users {
		// 1. Check if user with same Email exists but different ID
		var existing user.User
		err := database.DB.Unscoped().Where("email = ?", u.Email).First(&existing).Error
		if err == nil && existing.ID != u.ID {
			log.Printf("User email conflict for %s: Existing ID %s, Expected %s. Moving existing...", u.Email, existing.ID, u.ID)
			// Instead of deleting (which fails due to FKs), rename the email to clear the unique constraint
			database.DB.Unscoped().Model(&existing).Update("email", u.Email+"_old_"+existing.ID[:8])
		}

		// 2. Use OnConflict for ID safety
		if err := database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&u).Error; err != nil {
			log.Printf("Warning: Failed to ensure user %s: %v", u.Email, err)
			return err
		}
		log.Printf("Ensured user: %s (role_id: %s)", u.Email, u.RoleID)
	}

	log.Println("Users seeded successfully")
	return nil
}
