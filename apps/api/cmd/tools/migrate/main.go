package main

import (
	"log"
	"os"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/logger"
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

	if len(os.Args) > 1 && os.Args[1] == "reset" {
		log.Println("Reset flag detected. Dropping all tables...")
		if err := database.DropAllTables(); err != nil {
			log.Fatal("Failed to drop tables:", err)
		}
		// In this tool, we just drop. It will be followed by a normal migrate run
		return
	}

	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Migrations completed")
}
