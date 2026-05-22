package main

import (
	"log"
	"encoding/json"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
    "github.com/gilabs/gims/api/internal/sales/data/models"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	id := "6d673932-d881-460e-83b7-e08879f7ab7b"
	var invoice models.CustomerInvoice
	if err := database.DB.Preload("Items").Where("id = ?", id).First(&invoice).Error; err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(invoice, "", "  ")
	log.Printf("Invoice Details:\n%s", string(out))
}
