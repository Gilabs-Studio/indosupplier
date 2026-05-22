package main

import (
	"log"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/logger"
	"github.com/gilabs/gims/api/seeders"
)

func main() {
	logger.Init()

	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	if err := seeders.SeedRoles(); err != nil {
		log.Fatal("Failed to seed roles:", err)
	}
	if err := seeders.SeedMenus(); err != nil {
		log.Fatal("Failed to seed menus:", err)
	}
	if err := seeders.UpdateMenuStructure(); err != nil {
		log.Fatal("Failed to update menu structure:", err)
	}
	if err := seeders.SeedPermissions(); err != nil {
		log.Fatal("Failed to seed permissions:", err)
	}

	log.Println("RBAC sync completed")
}
