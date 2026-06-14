package seeders

import (
	"fmt"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedTransactions() error {
	var buyer buyerModels.BuyerProfile
	var user userModels.User
	if err := database.DB.Where("email = ?", "buyer2@indosupplier.local").First(&user).Error; err == nil {
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
	if err := database.DB.Limit(3).Find(&suppliers).Error; err != nil || len(suppliers) == 0 {
		return nil // skip if no suppliers
	}

	if !database.DB.Migrator().HasTable(&buyerModels.PurchaseOrder{}) {
		return nil // skip if table does not exist yet
	}

	if !database.DB.Migrator().HasTable(&buyerModels.Bookmark{}) {
		return nil
	}

	// Seed bookmarks for both buyer2 and admin
	emailsToSeed := []string{"buyer2@indosupplier.local", "admin@example.com"}
	for _, email := range emailsToSeed {
		var u userModels.User
		if err := database.DB.Where("email = ?", email).First(&u).Error; err == nil {
			var bp buyerModels.BuyerProfile
			if err := database.DB.Where("user_id = ?", u.ID).First(&bp).Error; err == nil {
				var bCount int64
				if err := database.DB.Model(&buyerModels.Bookmark{}).Where("buyer_profile_id = ?", bp.ID).Count(&bCount).Error; err == nil && bCount == 0 {
					var prod supplierModels.SupplierProduct
					hasProduct := database.DB.Where("supplier_profile_id = ?", suppliers[0].ID).First(&prod).Error == nil

					bookmarks := []buyerModels.Bookmark{
						{
							BuyerProfileID:    bp.ID,
							SupplierProfileID: suppliers[0].ID,
							Notes:             "Supplier utama untuk bahan baku mineral.",
						},
					}
					if hasProduct {
						bookmarks = append(bookmarks, buyerModels.Bookmark{
							BuyerProfileID:    bp.ID,
							SupplierProfileID: suppliers[0].ID,
							SupplierProductID: &prod.ID,
							Notes:             "Bahan baku pasir berkualitas.",
						})
					}
					if len(suppliers) > 1 {
						var prod2 supplierModels.SupplierProduct
						hasProduct2 := database.DB.Where("supplier_profile_id = ?", suppliers[1].ID).First(&prod2).Error == nil

						bookmarks = append(bookmarks, buyerModels.Bookmark{
							BuyerProfileID:    bp.ID,
							SupplierProfileID: suppliers[1].ID,
							Notes:             "Supplier alternatif untuk tekstil.",
						})

						if hasProduct2 {
							bookmarks = append(bookmarks, buyerModels.Bookmark{
								BuyerProfileID:    bp.ID,
								SupplierProfileID: suppliers[1].ID,
								SupplierProductID: &prod2.ID,
								Notes:             "Produk tekstil unggulan.",
							})
						}
					}
					for _, b := range bookmarks {
						if err := database.DB.Create(&b).Error; err != nil {
							fmt.Printf("warning: failed to seed bookmark for %s: %v\n", email, err)
						}
					}
					fmt.Printf("seeded mock bookmarks for %s\n", email)
				}
			}
		}
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
