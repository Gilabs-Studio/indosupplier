package seeders

import (
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

type marketplaceUserSeed struct {
	Email       string
	Name        string
	CompanyName string
	Industry    string
	Phone       string
	Website     string
	Address     string
	IsSupplier  bool
}

func SeedUsers() error {
	const seedPassword = "admin123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	seeds := []marketplaceUserSeed{
		{Email: "admin2@example.com", Name: "Yohanes Pratama", CompanyName: "PT Sumber Procurement", Industry: "Manufacturing"},
		{Email: "buyer2@indosupplier.local", Name: "Nadia Kirana", CompanyName: "CV Nusantara Retail", Industry: "Retail"},
		{Email: "buyer3@indosupplier.local", Name: "Arif Santoso", CompanyName: "PT Global Konstruksi", Industry: "Construction"},
		{Email: "buyer4@indosupplier.local", Name: "Maya Lestari", CompanyName: "PT Agro Makmur", Industry: "Agriculture"},
		{Email: "buyer5@indosupplier.local", Name: "Bima Hartono", CompanyName: "CV Logistik Prima", Industry: "Logistics"},
		{
			Email:       "admin@example.com",
			Name:        "Raka Wijaya",
			CompanyName: "PT Baja Sentosa",
			Industry:    "Steel Manufacturing",
			Phone:       "+6281234567890",
			Website:     "https://bajasentosa.indosupplier.local",
			Address:     "Jl. Industri Baja No. 88, Cakung, Jakarta Timur",
			IsSupplier:  true,
		},
		{Email: "supplier2@indosupplier.local", Name: "Sinta Maharani", CompanyName: "CV Tekstil Nusantara", Industry: "Textile", IsSupplier: true},
		{Email: "supplier3@indosupplier.local", Name: "Dimas Saputra", CompanyName: "PT Agro Indo Sejahtera", Industry: "Agriculture", IsSupplier: true},
		{Email: "supplier4@indosupplier.local", Name: "Laras Permata", CompanyName: "PT Kimia Cemerlang", Industry: "Chemical", IsSupplier: true},
		{Email: "supplier5@indosupplier.local", Name: "Fajar Mahendra", CompanyName: "CV Mesin Karya", Industry: "Machinery", IsSupplier: true},
	}

	for index, seed := range seeds {
		if err := seedMarketplaceUser(seed, string(hashedPassword), index+1); err != nil {
			return err
		}
	}

	fmt.Printf("seeded %d marketplace users; password: %s\n", len(seeds), seedPassword)
	return nil
}

func seedMarketplaceUser(seed marketplaceUserSeed, hashedPassword string, sequence int) error {
	var user userModels.User
	if err := database.DB.Where("email = ?", seed.Email).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		user = userModels.User{
			Email:     seed.Email,
			Password:  hashedPassword,
			Name:      seed.Name,
			AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=" + url.QueryEscape(seed.Email),
			Status:    "active",
		}
		if err := database.DB.Create(&user).Error; err != nil {
			return err
		}
	}

	if err := ensureBuyerProfile(user.ID, seed); err != nil {
		return err
	}

	if seed.IsSupplier {
		return ensureSupplierProfile(user.ID, seed, sequence)
	}

	return nil
}

func ensureBuyerProfile(userID string, seed marketplaceUserSeed) error {
	var profile buyerModels.BuyerProfile
	err := database.DB.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		profile = buyerProfileFromSeed(userID, seed)
		if err := database.DB.Create(&profile).Error; err != nil {
			return err
		}
	}

	if seed.Email == "admin@example.com" {
		now := apptime.Now()
		updates := map[string]interface{}{
			"full_name":            seed.Name,
			"company_name":         seed.CompanyName,
			"country_code":         "ID",
			"industry":             seed.Industry,
			"purchase_frequency":   "monthly",
			"phone":                seed.Phone,
			"website":              seed.Website,
			"address":              seed.Address,
			"profile_completeness": 100,
			"company_verified_at":  &now,
			"updated_at":           now,
		}
		if err := database.DB.Model(&profile).Updates(updates).Error; err != nil {
			return err
		}
		return ensureAdminBuyerDocuments(profile.ID)
	}

	return nil
}

func buyerProfileFromSeed(userID string, seed marketplaceUserSeed) buyerModels.BuyerProfile {
	profile := buyerModels.BuyerProfile{
		UserID:              userID,
		FullName:            seed.Name,
		CompanyName:         seed.CompanyName,
		CountryCode:         "ID",
		Industry:            seed.Industry,
		PurchaseFrequency:   "monthly",
		ProfileCompleteness: 70,
	}
	if seed.Phone != "" {
		profile.Phone = seed.Phone
	}
	if seed.Website != "" {
		profile.Website = seed.Website
	}
	if seed.Address != "" {
		profile.Address = seed.Address
	}
	if seed.Email == "admin@example.com" {
		now := apptime.Now()
		profile.CompanyVerifiedAt = &now
		profile.ProfileCompleteness = 100
	}
	return profile
}

func ensureAdminBuyerDocuments(buyerProfileID string) error {
	var count int64
	if err := database.DB.Model(&buyerModels.BuyerDocument{}).
		Where("buyer_profile_id = ?", buyerProfileID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	reviewedAt := apptime.Now().AddDate(0, 0, -2)
	documents := []buyerModels.BuyerDocument{
		{
			BuyerProfileID: buyerProfileID,
			DocumentType:   "npwp",
			DocumentNumber: "01.234.567.8-999.000",
			FileURL:        "https://indosupplier.local/uploads/documents/admin-buyer-npwp.pdf",
			Status:         "verified",
			ReviewedAt:     &reviewedAt,
		},
		{
			BuyerProfileID: buyerProfileID,
			DocumentType:   "nib",
			DocumentNumber: "9120400000001",
			FileURL:        "https://indosupplier.local/uploads/documents/admin-buyer-nib.pdf",
			Status:         "verified",
			ReviewedAt:     &reviewedAt,
		},
	}

	for i := range documents {
		if err := database.DB.Create(&documents[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureSupplierProfile(userID string, seed marketplaceUserSeed, sequence int) error {
	var profile supplierModels.SupplierProfile
	err := database.DB.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profile = supplierModels.SupplierProfile{
				UserID:                 userID,
				CompanyName:            seed.CompanyName,
				CompanyType:            "manufacturer",
				TaxStatus:              "pkp",
				NPWP:                   fmt.Sprintf("01.234.%03d.8-999.000", sequence),
				CountryCode:            "ID",
				ProvinceID:             "DKI Jakarta",
				CityID:                 "Jakarta",
				Address:                fmt.Sprintf("Jl. Industri Raya No. %d, Jakarta", sequence),
				BusinessHours:          "Monday-Friday 08:00-17:00",
				Timezone:               "Asia/Jakarta",
				Description:            fmt.Sprintf("%s adalah supplier terverifikasi untuk industri %s.", seed.CompanyName, seed.Industry),
				Phone:                  fmt.Sprintf("+62812%08d", 34000000+sequence),
				WhatsApp:               fmt.Sprintf("+62812%08d", 34000000+sequence),
				Email:                  seed.Email,
				Website:                fmt.Sprintf("https://supplier%d.indosupplier.local", sequence),
				VerificationLevel:      2,
				IsPremiumVerified:      sequence%2 == 0,
				ResponseRate:           92,
				AvgResponseTimeMinutes: 120,
				StarRating:             4.5,
				ReviewCount:            12 + sequence,
				ProfileCompleteness:    85,
				Status:                 "active",
			}
			if err := database.DB.Create(&profile).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var productCount int64
	if err := database.DB.Model(&supplierModels.SupplierProduct{}).
		Where("supplier_profile_id = ?", profile.ID).
		Count(&productCount).Error; err != nil {
		return err
	}
	if productCount == 0 {
		return seedSupplierProducts(database.DB, profile.ID, seed.Industry)
	}

	return nil
}

type seedProd struct {
	Name          string
	Description   string
	MOQ           string
	StartingPrice float64
	Capacity      string
	Photos        []string
}

func getSeedProductsForCategory(categorySlug string) []seedProd {
	if categorySlug == "office-stationery" || categorySlug == "kantor-atk" {
		return []seedProd{
			{
				Name:          "Kertas HVS A4 80gsm (1 Karton / 5 Rim)",
				Description:   "Kertas HVS A4 80gsm kualitas premium, ultra white 98%, cocok untuk fotokopi berkecepatan tinggi, printer laser & inkjet kantor.",
				MOQ:           "10 Karton",
				StartingPrice: 215000,
				Capacity:      "2000 Karton / Bulan",
				Photos: []string{
					"/images/products/prod-hvs.webp",
				},
			},
			{
				Name:          "Ballpoint Pen Gel 0.5mm Hitam (1 Box / 12 pcs)",
				Description:   "Pena gel tinta hitam pekat cepat kering 0.5mm, pegangan karet ergonomis untuk kenyamanan menulis dokumen bisnis kantor.",
				MOQ:           "20 Box",
				StartingPrice: 45000,
				Capacity:      "5000 Box / Bulan",
				Photos: []string{
					"/images/categories/cat-kantor-atk.webp",
				},
			},
		}
	} else if categorySlug == "safety-k3" || categorySlug == "kebersihan-k3" {
		return []seedProd{
			{
				Name:          "Masker Medis 3 Ply (1 Box / 50 pcs)",
				Description:   "Masker bedah medis 3 ply dengan meltblown filter BFE 99%, izin edar Kemenkes RI, tali elastis nyaman dipakai sepanjang hari.",
				MOQ:           "20 Box",
				StartingPrice: 28500,
				Capacity:      "10000 Box / Bulan",
				Photos: []string{
					"/images/products/prod-masker.webp",
				},
			},
			{
				Name:          "Safety Helmet SNI Pro Guard",
				Description:   "Helm keselamatan proyek standar SNI & CE EN397, material ABS high impact dengan inner suspension tali putar 6 titik dan chin strap.",
				MOQ:           "50 Unit",
				StartingPrice: 42000,
				Capacity:      "3000 Unit / Bulan",
				Photos: []string{
					"/images/products/prod-helmet.webp",
				},
			},
		}
	} else if categorySlug == "electronics-it" || categorySlug == "elektronik" {
		return []seedProd{
			{
				Name:          "Laptop Business i5 8GB / 512GB SSD",
				Description:   "Laptop bisnis handal prosesor Intel Core i5 Gen-12, RAM 8GB DDR4 upgradeable, SSD NVMe 512GB, layar 14 inci FHD anti-glare, garansi resmi 2 tahun.",
				MOQ:           "5 Unit",
				StartingPrice: 8950000,
				Capacity:      "200 Unit / Bulan",
				Photos: []string{
					"/images/products/prod-laptop.webp",
				},
			},
		}
	} else if categorySlug == "machinery-industrial" || categorySlug == "mesin" {
		return []seedProd{
			{
				Name:          "Pompa Air Sentrifugal Industri 3HP 3-Phase",
				Description:   "Pompa sentrifugal heavy-duty motor tembaga 3HP 2.2kW 380V, debit air 600 L/menit, head max 32m untuk sirkulasi air pabrik dan pendingin.",
				MOQ:           "2 Unit",
				StartingPrice: 4850000,
				Capacity:      "100 Unit / Bulan",
				Photos: []string{
					"/images/products/prod-water-pump.webp",
				},
			},
		}
	} else if categorySlug == "furniture" {
		return []seedProd{
			{
				Name:          "Kursi Kantor Ergonomis Mesh Headrest",
				Description:   "Kursi kerja manajer ergonomis jaring breathable, lumbar support adjustable, mekanisme recline hidrolik kelas 4 garansi 2 tahun.",
				MOQ:           "5 Unit",
				StartingPrice: 1250000,
				Capacity:      "500 Unit / Bulan",
				Photos: []string{
					"/images/products/prod-office-chair.webp",
				},
			},
		}
	} else if categorySlug == "packaging" {
		return []seedProd{
			{
				Name:          "Karton Box Double Wall 40x30x30cm (Pack 25 pcs)",
				Description:   "Kardus packing corrugated double wall tebal K200/M150/K200 CB Flute, kuat menahan beban hingga 30kg pengiriman ekspedisi dan logistik.",
				MOQ:           "10 Bundle",
				StartingPrice: 350000,
				Capacity:      "5000 Bundle / Bulan",
				Photos: []string{
					"/images/products/prod-carton-boxes.webp",
				},
			},
		}
	} else if categorySlug == "steel-metal" {
		return []seedProd{
			{
				Name:          "Reinforced Steel Bar (Rebar) D10 SNI",
				Description:   "High quality deformed steel rebar D10 for heavy construction, building frameworks, and civil engineering projects. SNI standard certified.",
				MOQ:           "10 Ton",
				StartingPrice: 12500000,
				Capacity:      "500 Ton / Month",
				Photos: []string{
					"/images/categories/cat-bahan-baku.webp",
				},
			},
		}
	} else if categorySlug == "agricultural-products" {
		return []seedProd{
			{
				Name:          "Premium Sumatra Gayo Arabica Green Coffee Beans",
				Description:   "Single-origin Sumatra Gayo Arabica green coffee beans, semi-washed process, Grade 1 double picked. Moisture content 12-13%. Deep herbal notes.",
				MOQ:           "250 Kg",
				StartingPrice: 92000,
				Capacity:      "10 Ton / Month",
				Photos: []string{
					"/images/categories/cat-makanan-minuman.webp",
				},
			},
		}
	}

	return []seedProd{
		{
			Name:          "Sodium Bentonite Clay Powder 25kg",
			Description:   "Premium expandable sodium bentonite powder for civil engineering, drilling mud stabilizer, and bonding agent in foundry sands.",
			MOQ:           "10 Ton",
			StartingPrice: 4500000,
			Capacity:      "300 Ton / Month",
			Photos: []string{
				"/images/products/prod-mineral-powder.webp",
			},
		},
	}
}

func seedSupplierProducts(db *gorm.DB, supplierProfileID string, industry string) error {
	categories := []struct {
		Slug string
		Name string
	}{
		{Slug: "kantor-atk", Name: "Kantor & ATK"},
		{Slug: "safety-k3", Name: "Kebersihan & K3"},
		{Slug: "electronics-it", Name: "Elektronik & IT"},
		{Slug: "machinery-industrial", Name: "Mesin & Industrial"},
		{Slug: "furniture", Name: "Furniture"},
		{Slug: "packaging", Name: "Packaging"},
		{Slug: "steel-metal", Name: "Bahan Baku"},
		{Slug: "agricultural-products", Name: "Komoditas Tani"},
	}

	categorySlug := "kantor-atk"
	categoryName := "Kantor & ATK"

	if industry == "Chemical" || industry == "Manufacturing" {
		categorySlug = "safety-k3"
		categoryName = "Kebersihan & K3"
	} else if industry == "Machinery" {
		categorySlug = "machinery-industrial"
		categoryName = "Mesin & Industrial"
	} else if industry == "Steel Manufacturing" {
		categorySlug = "steel-metal"
		categoryName = "Bahan Baku"
	} else if industry == "Agriculture" {
		categorySlug = "agricultural-products"
		categoryName = "Komoditas Tani"
	} else if industry == "Textile" {
		categorySlug = "packaging"
		categoryName = "Packaging"
	}

	// Ensure all standard categories exist
	for _, c := range categories {
		var existingCat supplierModels.Category
		if err := db.Where("slug = ?", c.Slug).First(&existingCat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newCat := supplierModels.Category{
					Slug:        c.Slug,
					Name:        c.Name,
					Description: "Kategori produk " + c.Name,
					IsActive:    true,
				}
				_ = db.Create(&newCat)
			}
		}
	}

	var cat supplierModels.Category
	if err := db.Where("slug = ?", categorySlug).First(&cat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cat = supplierModels.Category{
				Slug:        categorySlug,
				Name:        categoryName,
				Description: "Products related to " + categoryName,
				IsActive:    true,
			}
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	prods := getSeedProductsForCategory(categorySlug)

	for i, p := range prods {
		prod := supplierModels.SupplierProduct{
			SupplierProfileID: supplierProfileID,
			CategoryID:        cat.ID,
			Name:              p.Name,
			Description:       p.Description,
			MOQ:               p.MOQ,
			StartingPrice:     p.StartingPrice,
			Currency:          utils.DefaultCurrency(),
			CapacityText:      p.Capacity,
			IsFeatured:        i == 0,
			SortOrder:         i + 1,
		}

		if err := db.Create(&prod).Error; err != nil {
			return err
		}

		for j, fileURL := range p.Photos {
			photo := supplierModels.SupplierProductPhoto{
				SupplierProductID: prod.ID,
				FileURL:           fileURL,
				Caption:           fmt.Sprintf("Product Photo of %s %d", p.Name, j+1),
				SortOrder:         j,
			}
			if err := db.Create(&photo).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
