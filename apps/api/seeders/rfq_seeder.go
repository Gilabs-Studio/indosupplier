package seeders

import (
	"encoding/json"
	"fmt"
	"time"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/mapper"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedRFQs() error {
	var buyer buyerModels.BuyerProfile
	var user userModels.User
	if err := database.DB.Where("email = ?", "admin@example.com").First(&user).Error; err == nil {
		if errProfile := database.DB.Where("user_id = ?", user.ID).First(&buyer).Error; errProfile != nil {
			if errFirst := database.DB.First(&buyer).Error; errFirst != nil {
				return nil // skip if no buyers
			}
		}
	} else {
		if errFirst := database.DB.First(&buyer).Error; errFirst != nil {
			return nil // skip if no buyers
		}
	}

	var suppliers []supplierModels.SupplierProfile
	if err := database.DB.Limit(5).Find(&suppliers).Error; err != nil || len(suppliers) == 0 {
		return nil // skip if no suppliers
	}

	var category supplierModels.Category
	if err := database.DB.First(&category).Error; err != nil {
		return nil // skip if no categories
	}

	rfqData := []struct {
		AlphanumericID string
		Product        string
		CategorySlug   string
		Quantity       float64
		Unit           string
		TargetPort     string
		Description    string
		Status         string
	}{
		{
			AlphanumericID: "RFQ-2026-004",
			Product:        "Garnet Sand Mesh 80",
			CategorySlug:   "industrial-minerals",
			Quantity:       50,
			Unit:           "Ton",
			TargetPort:     "Tanjung Priok, Jakarta",
			Description:    "Membutuhkan Garnet Sand Mesh 80 kualitas industri untuk keperluan sandblasting lambung kapal. Pengiriman ke Tanjung Priok Jakarta.",
			Status:         "waiting",
		},
		{
			AlphanumericID: "RFQ-2026-003",
			Product:        "Bentonite Clay Powder",
			CategorySlug:   "chemicals",
			Quantity:       20,
			Unit:           "Ton",
			TargetPort:     "Tanjung Perak, Surabaya",
			Description:    "Membutuhkan Bentonite Clay Powder kualitas industri untuk konstruksi sipil. Kemasan bag 25kg, total kebutuhan 20 ton dikirim ke pelabuhan Surabaya. Sertakan Certificate of Analysis (COA) terbaru.",
			Status:         "received",
		},
		{
			AlphanumericID: "RFQ-2026-002",
			Product:        "Quartz Powder 325 Mesh",
			CategorySlug:   "industrial-minerals",
			Quantity:       100,
			Unit:           "Ton",
			TargetPort:     "Tanjung Priok, Jakarta",
			Description:    "Quartz powder 325 mesh untuk bahan baku industri keramik.",
			Status:         "completed",
		},
		{
			AlphanumericID: "RFQ-2026-001",
			Product:        "Organic Coconut Sugar Organic Grade",
			CategorySlug:   "agricultural-products",
			Quantity:       5,
			Unit:           "Ton",
			TargetPort:     "Port of Rotterdam (CIF)",
			Description:    "Organic coconut sugar.",
			Status:         "completed",
		},
	}

	for _, d := range rfqData {
		uuidID := mapper.ResolveRFQID(d.AlphanumericID)

		var existing models.RFQ
		if err := database.DB.Where("id = ?", uuidID).First(&existing).Error; err == nil {
			if existing.BuyerProfileID != buyer.ID {
				database.DB.Model(&existing).Update("buyer_profile_id", buyer.ID)
			}
			continue
		}

		// Find category matching category slug, or fallback
		var cat supplierModels.Category
		if err := database.DB.Where("slug = ?", d.CategorySlug).First(&cat).Error; err != nil {
			cat = category
		}

		now := apptime.Now()
		var closedAt *time.Time
		if d.Status == "completed" {
			t := now.AddDate(0, 0, -5)
			closedAt = &t
		}

		rfq := models.RFQ{
			ID:                  uuidID,
			BuyerProfileID:      buyer.ID,
			Title:               d.Product,
			ProductDescription:  d.Description,
			QuantityValue:       d.Quantity,
			QuantityUnit:        d.Unit,
			DestinationLocation: d.TargetPort,
			CategoryID:          cat.ID,
			VisibilityStatus:    "open",
			Mode:                "broadcast",
			CreatedAt:           now.AddDate(0, 0, -10),
			ClosedAt:            closedAt,
		}

		if err := database.DB.Create(&rfq).Error; err != nil {
			return err
		}

		// Seed specifications attachment for RFQ-2026-004 and RFQ-2026-003
		if d.AlphanumericID == "RFQ-2026-004" || d.AlphanumericID == "RFQ-2026-003" {
			attachment := models.RFQAttachment{
				RFQID:     rfq.ID,
				FileURL:   "https://indosupplier.local/uploads/coa.pdf",
				FileName:  "COA_Bentonite_Clay_2026.pdf",
				MimeType:  "application/pdf",
				FileSize:  1400000,
				CreatedAt: now.AddDate(0, 0, -10),
			}
			database.DB.Create(&attachment)
		}

		// Seed bids/recipients for RFQs that received quotes or completed
		if d.Status == "received" || d.Status == "completed" {
			bidsData := []struct {
				SupplierIndex int
				Price         string
				MOQ           string
			}{
				{SupplierIndex: 0, Price: "Rp 12.500 / Kg", MOQ: "5 Ton"},
				{SupplierIndex: 1, Price: "Rp 11.800 / Kg", MOQ: "10 Ton"},
				{SupplierIndex: 2, Price: "Rp 13.000 / Kg", MOQ: "1 Ton"},
			}

			for idx, bd := range bidsData {
				if bd.SupplierIndex >= len(suppliers) {
					continue
				}
				supp := suppliers[bd.SupplierIndex]

				recStatus := "responded"
				if d.Status == "completed" && idx == 0 {
					recStatus = "accepted"
				}

				respTime := now.AddDate(0, 0, -9)
				recipient := models.RFQRecipient{
					RFQID:             rfq.ID,
					SupplierProfileID: supp.ID,
					Status:            recStatus,
					RespondedAt:       &respTime,
					CreatedAt:         now.AddDate(0, 0, -10),
				}

				if err := database.DB.Create(&recipient).Error; err != nil {
					return err
				}

				metadataMap := map[string]string{
					"price": bd.Price,
					"moq":   bd.MOQ,
				}
				metadataBytes, _ := json.Marshal(metadataMap)

				msg := models.RFQMessage{
					RFQID:       rfq.ID,
					SenderType:  "supplier",
					SenderID:    supp.ID,
					MessageType: "offer",
					Body:        string(metadataBytes),
					Metadata:    string(metadataBytes),
					CreatedAt:   now.AddDate(0, 0, -9),
				}
				database.DB.Create(&msg)
			}
		}
	}

	fmt.Println("seeded mock RFQs, attachments, and bids")
	return nil
}
