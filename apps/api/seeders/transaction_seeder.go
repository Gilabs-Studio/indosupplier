package seeders

import (
	"fmt"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

func SeedTransactions() error {
	var buyer buyerModels.BuyerProfile
	if err := database.DB.First(&buyer).Error; err != nil {
		return nil // skip if no buyers
	}

	var suppliers []supplierModels.SupplierProfile
	if err := database.DB.Limit(3).Find(&suppliers).Error; err != nil || len(suppliers) == 0 {
		return nil // skip if no suppliers
	}

	if !database.DB.Migrator().HasTable(&buyerModels.PurchaseOrder{}) {
		return nil // skip if table does not exist yet
	}

	var count int64
	if err := database.DB.Model(&buyerModels.PurchaseOrder{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	txs := []buyerModels.PurchaseOrder{
		{
			PONumber:          "PO-20260613-A01B2C",
			BuyerProfileID:    buyer.ID,
			SupplierProfileID: suppliers[0].ID,
			ProductName:       "Garnet Sand Mesh 80 Almandine",
			QuantityValue:     20,
			QuantityUnit:      "Ton",
			PricePerUnit:      3800000,
			TotalAmount:       76000000,
			Status:            "processing",
			PaymentStatus:     "paid",
			DeliveryAddress:   "Pelabuhan Tanjung Priok, CIF Jakarta",
			Notes:             "Mohon pastikan kemasan karung goni ganda agar tidak bocor.",
		},
		{
			PONumber:          "PO-20260612-C02D3E",
			BuyerProfileID:    buyer.ID,
			SupplierProfileID: suppliers[0].ID,
			ProductName:       "Sodium Bentonite Clay Powder",
			QuantityValue:     10,
			QuantityUnit:      "Ton",
			PricePerUnit:      4500000,
			TotalAmount:       45000000,
			Status:            "completed",
			PaymentStatus:     "paid",
			DeliveryAddress:   "Pelabuhan Tanjung Perak, Surabaya",
			Notes:             "Kirimkan COA (Certificate of Analysis) bersama dengan dokumen pengapalan.",
		},
		{
			PONumber:          "PO-20260611-F03G4H",
			BuyerProfileID:    buyer.ID,
			SupplierProfileID: suppliers[0].ID,
			ProductName:       "Activated Carbon Powder Mesh 325",
			QuantityValue:     2,
			QuantityUnit:      "Ton",
			PricePerUnit:      16500000,
			TotalAmount:       33000000,
			Status:            "pending",
			PaymentStatus:     "unpaid",
			DeliveryAddress:   "Pelabuhan Tanjung Priok, CIF Jakarta",
			Notes:             "Kirim secepatnya.",
		},
	}

	for _, tx := range txs {
		if err := database.DB.Create(&tx).Error; err != nil {
			return err
		}
	}

	fmt.Println("seeded mock purchase orders")
	return nil
}
