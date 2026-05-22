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

	if err := seeders.SeedAll(); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	log.Println("Seeding completed")
}
