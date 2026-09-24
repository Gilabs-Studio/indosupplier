package seeders

import (
	"fmt"
	"math"
	"time"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	rfqModels "github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

func SeedSupplierDashboard() error {
	var user userModels.User
	if err := database.DB.Where("email = ?", "admin@example.com").First(&user).Error; err != nil {
		return nil // skip if admin user not found yet
	}

	var profile supplierModels.SupplierProfile
	if err := database.DB.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		return nil // skip if profile not found
	}

	// 1. Update PT Baja Sentosa profile stats & level
	updates := map[string]interface{}{
		"verification_level":        2, // Gold Level Verified
		"is_premium_verified":       true,
		"star_rating":               4.8,
		"review_count":              24,
		"response_rate":             98.5,
		"avg_response_time_minutes": 120, // 2 jam
		"status":                    "active",
		"established_year":          "2015",
	}
	if err := database.DB.Model(&profile).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update supplier profile for dashboard: %w", err)
	}

	// 2. Ensure Steel & Metal category is linked to PT Baja Sentosa
	var cat supplierModels.Category
	if err := database.DB.Where("slug = ?", "steel-metal").First(&cat).Error; err == nil {
		var supCat supplierModels.SupplierCategory
		if err := database.DB.Where("supplier_profile_id = ? AND category_id = ?", profile.ID, cat.ID).First(&supCat).Error; err != nil {
			supCat = supplierModels.SupplierCategory{
				SupplierProfileID: profile.ID,
				CategoryID:        cat.ID,
			}
			_ = database.DB.Create(&supCat).Error
		}
	}

	// 3. Find a buyer profile for PO buyer
	var buyer buyerModels.BuyerProfile
	if err := database.DB.First(&buyer).Error; err != nil {
		return nil
	}

	// 4. Seed 12 months of monthly Purchase Orders for 2026 to render Sales Performance Graph
	// Check if already seeded for 2026
	var count2026 int64
	database.DB.Model(&buyerModels.PurchaseOrder{}).
		Where("supplier_profile_id = ? AND EXTRACT(YEAR FROM created_at) = 2026", profile.ID).
		Count(&count2026)

	if count2026 < 12 {
		monthlyTargets := []struct {
			month   int
			day     int
			product string
			qty     float64
			unit    string
			price   float64
			total   float64
		}{
			{1, 15, "Reinforced Steel Bar (Rebar) D10", 3, "Ton", 15000000, 45000000},
			{2, 14, "Hot Rolled Carbon Steel Plate 6mm", 4, "Ton", 15000000, 60000000},
			{3, 20, "Galvanized Steel Sheet Coil", 2, "Ton", 25000000, 50000000},
			{4, 18, "Reinforced Steel Bar (Rebar) D12", 5, "Ton", 15000000, 75000000},
			{5, 22, "Hot Rolled Carbon Steel Plate 10mm", 6, "Ton", 15000000, 90000000},
			{6, 17, "Baja Profil H-Beam 150 Standard SNI", 5, "Ton", 17000000, 85000000},
			{7, 12, "Reinforced Steel Bar (Rebar) D16", 7, "Ton", 15714285, 110000000},
			{8, 19, "Galvanized Steel Sheet Coil 1.2mm", 5, "Ton", 19000000, 95000000},
			{9, 10, "Hot Rolled Carbon Steel Plate 12mm", 8, "Ton", 15000000, 120000000},
			{10, 25, "Baja Profil WF 200 Heavy Duty", 8, "Ton", 16250000, 130000000},
			{11, 20, "Reinforced Steel Bar (Rebar) D19", 7, "Ton", 16428571, 115000000},
			{12, 15, "Baja Profil I-Beam & H-Beam Structural", 9, "Ton", 15555555, 140000000},
		}

		for _, mt := range monthlyTargets {
			poDate := time.Date(2026, time.Month(mt.month), mt.day, 10, 30, 0, 0, time.UTC)
			po := buyerModels.PurchaseOrder{
				PONumber:          fmt.Sprintf("PO-2026%02d-%s", mt.month, profile.ID[:6]),
				BuyerProfileID:    buyer.ID,
				SupplierProfileID: profile.ID,
				ProductName:       mt.product,
				QuantityValue:     mt.qty,
				QuantityUnit:      mt.unit,
				PricePerUnit:      mt.price,
				TotalAmount:       mt.total,
				Status:            "completed",
				PaymentStatus:     "paid",
				DeliveryAddress:   "Kawasan Industri Jababeka, Cikarang, Jawa Barat",
				Notes:             "Pesanan material konstruksi SNI bersertifikat.",
				CreatedAt:         poDate,
				UpdatedAt:         poDate,
			}
			_ = database.DB.Create(&po).Error
		}
	}

	// 5. Seed open RFQs matching Steel & Metal category
	if cat.ID != "" {
		var openRFQCount int64
		database.DB.Model(&rfqModels.RFQ{}).
			Where("category_id = ? AND visibility_status = 'open' AND closed_at IS NULL", cat.ID).
			Count(&openRFQCount)

		if openRFQCount < 4 {
			mockRFQs := []struct {
				title       string
				desc        string
				qty         float64
				unit        string
				destination string
			}{
				{
					title:       "Pengadaan Besi Beton Ulir D16 & D19 Standard SNI Proyek Flyover",
					desc:        "Dibutuhkan besi beton ulir mutu TS420B sertifikat SNI untuk konstruksi jalan layang.",
					qty:         50,
					unit:        "Ton",
					destination: "Pelabuhan Tanjung Priok, Jakarta",
				},
				{
					title:       "Kebutuhan Hot Rolled Steel Coil SS400 Ketebalan 4mm",
					desc:        "Mencari supplier pabrik baja coil hot rolled sertifikasi mill test report terjamin.",
					qty:         30,
					unit:        "Ton",
					destination: "Kawasan Industri MM2100, Cikarang",
				},
				{
					title:       "Baja Profil WF 200 & H-Beam Konstruksi Pabrik Baru",
					desc:        "Tender terbuka pengadaan profil baja struktural WF 200 dan H-Beam 150.",
					qty:         25,
					unit:        "Ton",
					destination: "Kawasan Industri Cilegon, Banten",
				},
				{
					title:       "Pipa Seamless Baja Karbon ASTM A106 Grade B Sch 40",
					desc:        "Pengadaan pipa seamless schedule 40 diameter 4 inch untuk instalasi piping industri.",
					qty:         15,
					unit:        "Ton",
					destination: "Pelabuhan Tanjung Perak, Surabaya",
				},
			}

			catIDPtr := cat.ID
			now := apptime.Now()
			for i, mr := range mockRFQs {
				rfq := rfqModels.RFQ{
					BuyerProfileID:      buyer.ID,
					Title:               mr.title,
					ProductDescription:  mr.desc,
					QuantityValue:       mr.qty,
					QuantityUnit:        mr.unit,
					DeliveryTimeline:    "45 Hari Kerja",
					DestinationLocation: mr.destination,
					BudgetMin:           mr.qty * 14000000,
					BudgetMax:           mr.qty * 17000000,
					Mode:                "public",
					CategoryID:          &catIDPtr,
					VisibilityStatus:    "open",
					CreatedAt:           now.AddDate(0, 0, -((i + 1) * 3)),
					UpdatedAt:           now.AddDate(0, 0, -((i + 1) * 3)),
				}
				if err := database.DB.Create(&rfq).Error; err == nil {
					// Invite PT Baja Sentosa as recipient
					rec := rfqModels.RFQRecipient{
						RFQID:             rfq.ID,
						SupplierProfileID: profile.ID,
						Status:            "new",
					}
					_ = database.DB.Create(&rec).Error

					// Create a sample quote proposal message for realism
					msg := rfqModels.RFQMessage{
						RFQID:       rfq.ID,
						SenderType:  "supplier",
						SenderID:    user.ID,
						MessageType: "proposal",
						Body:        "Kami dari PT Baja Sentosa siap memenuhi spesifikasi tersebut dengan standar SNI dan mill test certificate.",
					}
					_ = database.DB.Create(&msg).Error
				}
			}
		}
	}

	if err := seedSupplierReviewsForDashboard(profile.ID); err != nil {
		fmt.Printf("warning: failed to seed supplier reviews: %v\n", err)
	}

	return nil
}

func seedSupplierReviewsForDashboard(supplierProfileID string) error {
	if !database.DB.Migrator().HasTable(&trustModels.SupplierReview{}) {
		return nil
	}

	// Fetch buyers
	var buyers []buyerModels.BuyerProfile
	database.DB.Limit(5).Find(&buyers)
	if len(buyers) == 0 {
		return nil
	}

	// Fetch supplier product
	var prod supplierModels.SupplierProduct
	var prodID *string
	if err := database.DB.Where("supplier_profile_id = ?", supplierProfileID).First(&prod).Error; err == nil {
		prodID = &prod.ID
	}

	now := apptime.Now()
	repliedAt1 := now.AddDate(0, 0, -10)
	repliedAt2 := now.AddDate(0, 0, -16)

	seeds := []struct {
		buyerIndex int
		rating     int
		reviewText string
		reply      string
		repliedAt  *time.Time
		daysAgo    int
	}{
		{
			buyerIndex: 0,
			rating:     5,
			reviewText: "Besi beton ulir D10 presisi, sertifikat SNI valid dan delivery on time untuk proyek jembatan kami. Sangat profesional!",
			reply:      "Terima kasih atas kepercayaannya. Kami selalu menjaga standar mutu SNI dan kepatuhan jadwal pengiriman untuk mendukung kesuksesan proyek infrastruktur Anda.",
			repliedAt:  &repliedAt1,
			daysAgo:    12,
		},
		{
			buyerIndex: 1,
			rating:     5,
			reviewText: "Kualitas plat baja sangat memuaskan, toleransi ketebalan sesuai standar ASTM dan potongan rapi tanpa cacat pinggiran.",
			reply:      "",
			repliedAt:  nil,
			daysAgo:    7,
		},
		{
			buyerIndex: 2,
			rating:     4,
			reviewText: "Barang sesuai spesifikasi teknis dan packing bundle rapi. Sedikit catatan di proses konfirmasi surat jalan yang butuh 1 hari kerja, selebihnya memuaskan.",
			reply:      "",
			repliedAt:  nil,
			daysAgo:    5,
		},
		{
			buyerIndex: 3,
			rating:     5,
			reviewText: "Repeat order ketiga kali, respon marketing cepat tanggap dan penyesuaian termin pembayaran B2B sangat membantu cashflow proyek kami.",
			reply:      "Terima kasih atas kerja samanya. Kami berkomitmen memberikan fleksibilitas dan solusi rantai pasok terbaik untuk mitra jangka panjang kami.",
			repliedAt:  &repliedAt2,
			daysAgo:    18,
		},
		{
			buyerIndex: 4,
			rating:     3,
			reviewText: "Kualitas material bagus, namun mohon untuk ketersediaan armada truk trailer diatur lebih awal saat high-demand musim konstruksi agar tidak ada antrean bongkar muat.",
			reply:      "",
			repliedAt:  nil,
			daysAgo:    2,
		},
	}

	for _, s := range seeds {
		bIdx := s.buyerIndex
		if bIdx >= len(buyers) {
			bIdx = 0
		}
		buyer := buyers[bIdx]

		var count int64
		database.DB.Model(&trustModels.SupplierReview{}).
			Where("buyer_profile_id = ? AND supplier_profile_id = ? AND review_text = ?", buyer.ID, supplierProfileID, s.reviewText).
			Count(&count)
		if count > 0 {
			continue
		}

		rev := trustModels.SupplierReview{
			BuyerProfileID:    buyer.ID,
			SupplierProfileID: supplierProfileID,
			ProductID:         prodID,
			Rating:            s.rating,
			ReviewText:        s.reviewText,
			SupplierReply:     s.reply,
			SupplierRepliedAt: s.repliedAt,
			Status:            "approved",
			CreatedAt:         now.AddDate(0, 0, -s.daysAgo),
			UpdatedAt:         now.AddDate(0, 0, -s.daysAgo),
		}
		_ = database.DB.Create(&rev).Error
	}

	// Update supplier profile aggregate rating and review count
	type stat struct {
		Avg   float64 `gorm:"column:avg"`
		Count int64   `gorm:"column:count"`
	}
	var st stat
	database.DB.Table("supplier_reviews").
		Select("COALESCE(AVG(rating), 0) AS avg, COUNT(id) AS count").
		Where("supplier_profile_id = ? AND status = 'approved' AND deleted_at IS NULL", supplierProfileID).
		Scan(&st)

	if st.Count > 0 {
		database.DB.Model(&supplierModels.SupplierProfile{}).
			Where("id = ?", supplierProfileID).
			Updates(map[string]interface{}{
				"star_rating":  math.Round(st.Avg*10) / 10,
				"review_count": int(st.Count),
			})
	}

	return nil
}
