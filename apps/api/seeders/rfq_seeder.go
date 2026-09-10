package seeders

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedRFQs() error {
	var buyer buyerModels.BuyerProfile
	var user userModels.User
	if err := database.DB.Where("email = ?", "admin@example.com").First(&user).Error; err != nil {
		return nil // skip if the dedicated dev user has not been seeded yet
	}
	if err := database.DB.Where("user_id = ?", user.ID).First(&buyer).Error; err != nil {
		return nil // skip if the dedicated dev buyer profile has not been seeded yet
	}

	var suppliers []supplierModels.SupplierProfile
	if err := database.DB.
		Where("user_id <> ?", user.ID).
		Order("verification_level DESC, updated_at DESC").
		Limit(5).
		Find(&suppliers).Error; err != nil || len(suppliers) == 0 {
		return nil // skip if no suppliers
	}

	var category supplierModels.Category
	if err := database.DB.First(&category).Error; err != nil {
		return nil // skip if no categories
	}

	rfqData := []struct {
		SeedCode     string
		Product      string
		CategorySlug string
		Quantity     float64
		Unit         string
		TargetPort   string
		Description  string
		Status       string
	}{
		{
			SeedCode:     "RFQ-2026-004",
			Product:      "Garnet Sand Mesh 80",
			CategorySlug: "industrial-minerals",
			Quantity:     50,
			Unit:         "Ton",
			TargetPort:   "Tanjung Priok, Jakarta",
			Description:  "Membutuhkan Garnet Sand Mesh 80 kualitas industri untuk keperluan sandblasting lambung kapal. Pengiriman ke Tanjung Priok Jakarta.",
			Status:       "waiting",
		},
		{
			SeedCode:     "RFQ-2026-003",
			Product:      "Bentonite Clay Powder",
			CategorySlug: "chemicals",
			Quantity:     20,
			Unit:         "Ton",
			TargetPort:   "Tanjung Perak, Surabaya",
			Description:  "Membutuhkan Bentonite Clay Powder kualitas industri untuk konstruksi sipil. Kemasan bag 25kg, total kebutuhan 20 ton dikirim ke pelabuhan Surabaya. Sertakan Certificate of Analysis (COA) terbaru.",
			Status:       "received",
		},
		{
			SeedCode:     "RFQ-2026-002",
			Product:      "Quartz Powder 325 Mesh",
			CategorySlug: "industrial-minerals",
			Quantity:     100,
			Unit:         "Ton",
			TargetPort:   "Tanjung Priok, Jakarta",
			Description:  "Quartz powder 325 mesh untuk bahan baku industri keramik.",
			Status:       "completed",
		},
		{
			SeedCode:     "RFQ-2026-001",
			Product:      "Organic Coconut Sugar Organic Grade",
			CategorySlug: "agricultural-products",
			Quantity:     5,
			Unit:         "Ton",
			TargetPort:   "Port of Rotterdam (CIF)",
			Description:  "Organic coconut sugar.",
			Status:       "completed",
		},
	}

	for _, d := range rfqData {
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

		rfq := models.RFQ{}
		err := database.DB.
			Where("buyer_profile_id = ? AND title = ? AND destination_location = ?", buyer.ID, d.Product, d.TargetPort).
			First(&rfq).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rfq = models.RFQ{
				BuyerProfileID:      buyer.ID,
				Title:               d.Product,
				ProductDescription:  d.Description,
				QuantityValue:       d.Quantity,
				QuantityUnit:        d.Unit,
				DestinationLocation: d.TargetPort,
				CategoryID:          &cat.ID,
				VisibilityStatus:    "open",
				Mode:                "broadcast",
				CreatedAt:           now.AddDate(0, 0, -10),
			}
			if d.Status == "completed" {
				t := now.AddDate(0, 0, -5)
				rfq.ClosedAt = &t
			}
			if err := database.DB.Create(&rfq).Error; err != nil {
				return err
			}
		} else {
			updates := map[string]interface{}{
				"product_description":  d.Description,
				"quantity_value":       d.Quantity,
				"quantity_unit":        d.Unit,
				"destination_location": d.TargetPort,
				"category_id":          cat.ID,
				"visibility_status":    "open",
				"mode":                 "broadcast",
				"updated_at":           now,
			}
			if d.Status == "completed" && rfq.ClosedAt == nil {
				updates["closed_at"] = closedAt
			}
			if err := database.DB.Model(&rfq).Updates(updates).Error; err != nil {
				return err
			}
		}

		if d.SeedCode == "RFQ-2026-004" || d.SeedCode == "RFQ-2026-003" {
			if err := ensureRFQAttachment(rfq.ID, d.SeedCode, now); err != nil {
				return err
			}
		}

		if err := ensureRFQRecipients(rfq.ID, d.Status, suppliers, now); err != nil {
			return err
		}
	}

	fmt.Println("seeded mock RFQs, attachments, recipients, and bids for admin@example.com")
	return nil
}

func ensureRFQAttachment(rfqID string, seedCode string, now time.Time) error {
	var count int64
	if err := database.DB.Model(&models.RFQAttachment{}).
		Where("rfq_id = ?", rfqID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	fileName := "RFQ_Specification.pdf"
	if seedCode == "RFQ-2026-004" {
		fileName = "Spec_Garnet_Sand_Mesh_80.pdf"
	}
	if seedCode == "RFQ-2026-003" {
		fileName = "COA_Bentonite_Clay_2026.pdf"
	}

	attachment := models.RFQAttachment{
		RFQID:     rfqID,
		FileURL:   "https://indosupplier.local/uploads/rfqs/" + fileName,
		FileName:  fileName,
		MimeType:  "application/pdf",
		FileSize:  1400000,
		CreatedAt: now.AddDate(0, 0, -10),
		UpdatedAt: now,
	}
	return database.DB.Create(&attachment).Error
}

func ensureRFQRecipients(rfqID string, seedStatus string, suppliers []supplierModels.SupplierProfile, now time.Time) error {
	bidsData := []struct {
		SupplierIndex int
		Price         float64
		MOQ           string
		DeliveryTime  string
	}{
		{SupplierIndex: 0, Price: 12500, MOQ: "5 Ton", DeliveryTime: "2 Jam"},
		{SupplierIndex: 1, Price: 11800, MOQ: "10 Ton", DeliveryTime: "3 Jam"},
		{SupplierIndex: 2, Price: 13000, MOQ: "1 Ton", DeliveryTime: "1 Jam"},
	}
	bidsBySupplierIndex := make(map[int]struct {
		Price        float64
		MOQ          string
		DeliveryTime string
	}, len(bidsData))
	for _, bid := range bidsData {
		bidsBySupplierIndex[bid.SupplierIndex] = struct {
			Price        float64
			MOQ          string
			DeliveryTime string
		}{
			Price:        bid.Price,
			MOQ:          bid.MOQ,
			DeliveryTime: bid.DeliveryTime,
		}
	}

	for idx, supplier := range suppliers {
		recipient := models.RFQRecipient{}
		err := database.DB.
			Where("rfq_id = ? AND supplier_profile_id = ?", rfqID, supplier.ID).
			First(&recipient).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			recipient = models.RFQRecipient{
				RFQID:             rfqID,
				SupplierProfileID: supplier.ID,
				Status:            "new",
				RankPosition:      idx + 1,
				CreatedAt:         now.AddDate(0, 0, -10),
				UpdatedAt:         now,
			}
			if err := database.DB.Create(&recipient).Error; err != nil {
				return err
			}
		} else if recipient.RankPosition == 0 {
			if err := database.DB.Model(&recipient).Updates(map[string]interface{}{
				"rank_position": idx + 1,
				"updated_at":    now,
			}).Error; err != nil {
				return err
			}
		}

		if seedStatus != "received" && seedStatus != "completed" {
			continue
		}

		bid, ok := bidsBySupplierIndex[idx]
		if !ok {
			continue
		}

		respTime := now.AddDate(0, 0, -9)
		recStatus := "responded"
		if seedStatus == "completed" && idx == 0 {
			recStatus = "accepted"
		}
		if err := database.DB.Model(&models.RFQRecipient{}).
			Where("id = ?", recipient.ID).
			Updates(map[string]interface{}{
				"status":       recStatus,
				"responded_at": &respTime,
				"updated_at":   now,
			}).Error; err != nil {
			return err
		}

		if err := ensureRFQOffer(rfqID, supplier.ID, bid.Price, bid.MOQ, bid.DeliveryTime, now); err != nil {
			return err
		}
	}

	return nil
}

func ensureRFQOffer(rfqID string, supplierID string, price float64, moq string, deliveryTime string, now time.Time) error {
	metadataMap := map[string]string{
		"price":        fmt.Sprintf("%s / Kg", utils.FormatMoney(price, utils.DefaultCurrency())),
		"moq":          moq,
		"deliveryTime": deliveryTime,
	}
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return err
	}

	body := "Penawaran awal dari supplier untuk kebutuhan RFQ ini."
	message := models.RFQMessage{}
	err = database.DB.
		Where("rfq_id = ? AND sender_id = ? AND message_type = ?", rfqID, supplierID, "offer").
		First(&message).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		message = models.RFQMessage{
			RFQID:       rfqID,
			SenderType:  "supplier",
			SenderID:    supplierID,
			MessageType: "offer",
			Body:        body,
			Metadata:    string(metadataBytes),
			CreatedAt:   now.AddDate(0, 0, -9),
			UpdatedAt:   now,
		}
		return database.DB.Create(&message).Error
	}

	return database.DB.Model(&message).Updates(map[string]interface{}{
		"body":       body,
		"metadata":   string(metadataBytes),
		"updated_at": now,
	}).Error
}
