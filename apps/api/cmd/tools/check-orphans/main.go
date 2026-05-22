package main

import (
	"log"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
    "github.com/gilabs/gims/api/internal/role/data/models"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	var count int64
	database.DB.Model(&models.RolePermission{}).
		Where("NOT EXISTS (SELECT 1 FROM permissions WHERE permissions.id = role_permissions.permission_id)").
		Count(&count)

	log.Printf("Found %d orphan role_permissions", count)
    
    if count > 0 {
        log.Println("Deleting orphans...")
        if err := database.DB.Exec("DELETE FROM role_permissions WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permissions.id = role_permissions.permission_id)").Error; err != nil {
            log.Fatal(err)
        }
        log.Println("Done")
    }
}
