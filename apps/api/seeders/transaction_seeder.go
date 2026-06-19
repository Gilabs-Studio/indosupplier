package seeders

import (
	"fmt"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
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
	hasFollowingTable := database.DB.Migrator().HasTable(&buyerModels.SupplierFollowing{})

	// Seed product bookmarks and followed suppliers for both buyer2 and admin
	emailsToSeed := []string{"buyer2@indosupplier.local", "admin@example.com"}
	for _, email := range emailsToSeed {
		var u userModels.User
		if err := database.DB.Where("email = ?", email).First(&u).Error; err == nil {
			var bp buyerModels.BuyerProfile
			if err := database.DB.Where("user_id = ?", u.ID).First(&bp).Error; err == nil {
				if hasFollowingTable {
					var followingCount int64
					if err := database.DB.Model(&buyerModels.SupplierFollowing{}).Where("buyer_profile_id = ?", bp.ID).Count(&followingCount).Error; err == nil && followingCount == 0 {
						followings := []buyerModels.SupplierFollowing{
							{
								BuyerProfileID:    bp.ID,
								SupplierProfileID: suppliers[0].ID,
							},
						}
						if len(suppliers) > 1 {
							followings = append(followings, buyerModels.SupplierFollowing{
								BuyerProfileID:    bp.ID,
								SupplierProfileID: suppliers[1].ID,
							})
						}
						for _, following := range followings {
							if err := database.DB.Create(&following).Error; err != nil {
								fmt.Printf("warning: failed to seed following supplier for %s: %v\n", email, err)
							}
						}
						fmt.Printf("seeded mock following suppliers for %s\n", email)
					}
				}

				var bCount int64
				if err := database.DB.Model(&buyerModels.Bookmark{}).Where("buyer_profile_id = ?", bp.ID).Count(&bCount).Error; err == nil && bCount == 0 {
					var prod supplierModels.SupplierProduct
					hasProduct := database.DB.Where("supplier_profile_id = ?", suppliers[0].ID).First(&prod).Error == nil

					bookmarks := []buyerModels.Bookmark{}
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

	// Seed purchase orders and reviews for both buyer2 and admin
	for _, email := range emailsToSeed {
		var u userModels.User
		if err := database.DB.Where("email = ?", email).First(&u).Error; err == nil {
			var bp buyerModels.BuyerProfile
			if err := database.DB.Where("user_id = ?", u.ID).First(&bp).Error; err == nil {
				var poCount int64
				if err := database.DB.Model(&buyerModels.PurchaseOrder{}).Where("buyer_profile_id = ?", bp.ID).Count(&poCount).Error; err == nil && poCount == 0 {
					// Seed supplier profiles count check
					var sCount int64
					database.DB.Model(&supplierModels.SupplierProfile{}).Count(&sCount)
					if sCount == 0 {
						continue
					}

					txs := []buyerModels.PurchaseOrder{
						{
							PONumber:          fmt.Sprintf("PO-20260613-%s", bp.ID[:6]),
							BuyerProfileID:    bp.ID,
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
							PONumber:          fmt.Sprintf("PO-20260612-%s", bp.ID[len(bp.ID)-6:]),
							BuyerProfileID:    bp.ID,
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
							PONumber:          fmt.Sprintf("PO-20260611-%s", bp.ID[9:15]),
							BuyerProfileID:    bp.ID,
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

					// Completed and reviewed transaction
					po4 := buyerModels.PurchaseOrder{
						PONumber:          fmt.Sprintf("PO-20260610-%s", bp.ID[2:8]),
						BuyerProfileID:    bp.ID,
						SupplierProfileID: suppliers[0].ID,
						ProductName:       "Raw Indigo Denim Fabric 12oz",
						QuantityValue:     15,
						QuantityUnit:      "Roll",
						PricePerUnit:      1200000,
						TotalAmount:       18000000,
						Status:            "completed",
						PaymentStatus:     "paid",
						DeliveryAddress:   "Gudang Utama GIMS, Jakarta",
						Notes:             "Denim kualitas ekspor.",
					}
					txs = append(txs, po4)

					for i := range txs {
						if err := database.DB.Create(&txs[i]).Error; err != nil {
							fmt.Printf("warning: failed to seed PO for %s: %v\n", email, err)
						}
					}

					// Seed review for po4
					if database.DB.Migrator().HasTable(&trustModels.SupplierReview{}) {
						r1 := trustModels.SupplierReview{
							BuyerProfileID:    bp.ID,
							SupplierProfileID: suppliers[0].ID,
							PurchaseOrderID:   &txs[3].ID,
							Rating:            5,
							ReviewText:        fmt.Sprintf("Sangat puas dengan denim dari %s. Bahan tebal, warna solid, pengiriman cepat.", suppliers[0].CompanyName),
							Status:            "approved",
						}
						if err := database.DB.Create(&r1).Error; err != nil {
							fmt.Printf("warning: failed to seed review for %s: %v\n", email, err)
						}
					}

					fmt.Printf("seeded mock purchase orders and reviews for %s\n", email)
				}
			}
		}
	}

	if err := seedDiscoveryDemoReviews(); err != nil {
		fmt.Printf("warning: failed to seed discovery demo reviews: %v\n", err)
	}

	return nil
}

func seedDiscoveryDemoReviews() error {
	if !database.DB.Migrator().HasTable(&trustModels.SupplierReview{}) {
		return nil
	}

	var supplier supplierModels.SupplierProfile
	if err := database.DB.Where("company_name = ?", "PT Agro Indo Sejahtera").First(&supplier).Error; err != nil {
		return nil
	}

	reviews := []struct {
		Email string
		Rate  int
		Text  string
	}{
		{
			Email: "buyer2@indosupplier.local",
			Rate:  5,
			Text:  "Biji kopi Gayo konsisten, kadar air sesuai spesifikasi, dan dokumen traceability lengkap untuk audit internal kami.",
		},
		{
			Email: "buyer4@indosupplier.local",
			Rate:  5,
			Text:  "Respons cepat untuk negosiasi MOQ. Sampel green bean datang rapi dan profil roasting sesuai catatan cupping.",
		},
	}

	for _, seed := range reviews {
		var user userModels.User
		if err := database.DB.Where("email = ?", seed.Email).First(&user).Error; err != nil {
			continue
		}

		var buyer buyerModels.BuyerProfile
		if err := database.DB.Where("user_id = ?", user.ID).First(&buyer).Error; err != nil {
			continue
		}

		var count int64
		if err := database.DB.Model(&trustModels.SupplierReview{}).
			Where("buyer_profile_id = ? AND supplier_profile_id = ? AND review_text = ?", buyer.ID, supplier.ID, seed.Text).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		review := trustModels.SupplierReview{
			BuyerProfileID:    buyer.ID,
			SupplierProfileID: supplier.ID,
			Rating:            seed.Rate,
			ReviewText:        seed.Text,
			Status:            "approved",
		}
		if err := database.DB.Create(&review).Error; err != nil {
			return err
		}
	}

	type reviewStats struct {
		AverageRating float64 `gorm:"column:average_rating"`
		TotalCount    int64   `gorm:"column:total_count"`
	}

	var stats reviewStats
	if err := database.DB.Table("supplier_reviews").
		Select("COALESCE(AVG(rating), 0) AS average_rating, COUNT(*) AS total_count").
		Where("supplier_profile_id = ? AND status = ?", supplier.ID, "approved").
		Scan(&stats).Error; err != nil {
		return err
	}

	if stats.TotalCount > 0 {
		return database.DB.Model(&supplierModels.SupplierProfile{}).
			Where("id = ?", supplier.ID).
			Updates(map[string]interface{}{
				"star_rating":  stats.AverageRating,
				"review_count": int(stats.TotalCount),
			}).Error
	}

	return nil
}
