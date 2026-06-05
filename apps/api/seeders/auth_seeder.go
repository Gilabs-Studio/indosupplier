package seeders

import (
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

type marketplaceUserSeed struct {
	Email       string
	Name        string
	CompanyName string
	Industry    string
	IsSupplier  bool
}

func SeedUsers() error {
	const seedPassword = "password123"

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
		{Email: "admin@example.com", Name: "Raka Wijaya", CompanyName: "PT Baja Sentosa", Industry: "Steel Manufacturing", IsSupplier: true},
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
	var count int64
	if err := database.DB.Model(&buyerModels.BuyerProfile{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return database.DB.Create(&buyerModels.BuyerProfile{
		UserID:              userID,
		FullName:            seed.Name,
		CompanyName:         seed.CompanyName,
		CountryCode:         "ID",
		Industry:            seed.Industry,
		PurchaseFrequency:   "monthly",
		ProfileCompleteness: 70,
	}).Error
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

func seedSupplierProducts(db *gorm.DB, supplierProfileID string, industry string) error {
	categoryName := "Industrial Minerals"
	categorySlug := "industrial-minerals"
	if industry == "Textile" {
		categoryName = "Textiles & Fabrics"
		categorySlug = "textiles-fabrics"
	} else if industry == "Agriculture" {
		categoryName = "Agricultural Products"
		categorySlug = "agricultural-products"
	} else if industry == "Steel Manufacturing" {
		categoryName = "Steel & Metal"
		categorySlug = "steel-metal"
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

	type seedProd struct {
		Name          string
		Description   string
		MOQ           string
		StartingPrice float64
		Capacity      string
		Photos        []string
	}

	var prods []seedProd
	if categorySlug == "steel-metal" {
		prods = []seedProd{
			{
				Name:          "Reinforced Steel Bar (Rebar) D10",
				Description:   "High quality deformed steel rebar D10 for heavy construction, building frameworks, and civil engineering projects. SNI standard certified.",
				MOQ:           "10 Ton",
				StartingPrice: 12500000,
				Capacity:      "500 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1516216628859-9bccecad13de?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1581094288338-2314dddb7ecc?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Hot Rolled Carbon Steel Plate 6mm",
				Description:   "ASTM A36 certified hot rolled carbon steel plate, 6mm thickness, ideal for structural fabrications, ship building, and general manufacturing.",
				MOQ:           "5 Ton",
				StartingPrice: 14200000,
				Capacity:      "300 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Galvanized Steel Sheet Coil",
				Description:   "Anti-corrosive hot-dipped galvanized steel coil, zinc coating thickness 120g/m2, suitable for roofing sheets and outdoor structural applications.",
				MOQ:           "8 Ton",
				StartingPrice: 16800000,
				Capacity:      "200 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1581092160607-ee22621dd758?auto=format&fit=crop&w=800&q=80",
				},
			},
		}
	} else if categorySlug == "textiles-fabrics" {
		prods = []seedProd{
			{
				Name:          "100% Organic Ring-Spun Cotton Yarn",
				Description:   "Premium combed ring-spun organic cotton yarn, count Ne 30/1, suitable for weaving high-quality soft fabrics, t-shirts, and baby garments.",
				MOQ:           "500 Kg",
				StartingPrice: 48000,
				Capacity:      "10000 Kg / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1528459801416-a9e53bbf4e17?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1606760227091-3dd870d97f1d?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Raw Indigo Denim Fabric 12oz",
				Description:   "Heavy-duty raw indigo selvedge denim fabric 12oz, 100% cotton, standard width 150cm, perfect for authentic jeans and heavy outerwear creation.",
				MOQ:           "1000 Meters",
				StartingPrice: 35000,
				Capacity:      "15000 Meters / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1541099649105-f69ad21f3246?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Polyester Staple Fiber Grade A",
				Description:   "Recycled semi-dull polyester staple fiber, 1.4D x 38mm, ideal for spinning high strength threads and non-woven fabric linings.",
				MOQ:           "2 Ton",
				StartingPrice: 26000,
				Capacity:      "50 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1558085324-2f298b28c714?auto=format&fit=crop&w=800&q=80",
				},
			},
		}
	} else if categorySlug == "agricultural-products" {
		prods = []seedProd{
			{
				Name:          "Fresh Indonesian Organic Ginger",
				Description:   "Export quality fresh big ginger (Gajah) and red ginger, organically grown in Central Java. Hand-washed, sorted, and packed in mesh bags.",
				MOQ:           "1 Ton",
				StartingPrice: 22000,
				Capacity:      "30 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1615485290382-441e4d049cb5?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1598170845058-32b9d6a5da37?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Premium Sumatra Gayo Arabica Coffee Beans",
				Description:   "Single-origin Sumatra Gayo Arabica green coffee beans, semi-washed process, Grade 1 double picked. Moisture content 12-13%. Deep herbal notes.",
				MOQ:           "250 Kg",
				StartingPrice: 92000,
				Capacity:      "10 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1447933601403-0c6688de566e?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Natural Coconut Sugar Blocks",
				Description:   "100% pure organic coconut sap sugar blocks, traditionally processed with no chemicals or additives. Perfect healthy sweetener alternative.",
				MOQ:           "1 Ton",
				StartingPrice: 19500,
				Capacity:      "15 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1596450514735-2d88a7051187?auto=format&fit=crop&w=800&q=80",
				},
			},
		}
	} else {
		prods = []seedProd{
			{
				Name:          "Garnet Sand Mesh 80 Almandine",
				Description:   "High grade almandine garnet sand mesh 80, highly abrasive and clean, optimized for waterjet cutting machines and steel surface sandblasting.",
				MOQ:           "20 Ton",
				StartingPrice: 3800000,
				Capacity:      "500 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1605281317010-fe5fed93a4c2?auto=format&fit=crop&w=800&q=80",
					"https://images.unsplash.com/photo-1581092160607-ee22621dd758?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Sodium Bentonite Clay Powder",
				Description:   "Premium expandable sodium bentonite powder for civil engineering, drilling mud stabilizer, and bonding agent in foundry sands.",
				MOQ:           "10 Ton",
				StartingPrice: 4500000,
				Capacity:      "300 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1576086213369-97a306d36557?auto=format&fit=crop&w=800&q=80",
				},
			},
			{
				Name:          "Activated Carbon Powder Mesh 325",
				Description:   "Coal-based activated carbon powder 325 mesh, high iodine value (900 mg/g), suitable for municipal water purification and gas adsorption.",
				MOQ:           "2 Ton",
				StartingPrice: 16500,
				Capacity:      "50 Ton / Month",
				Photos: []string{
					"https://images.unsplash.com/photo-1607619056574-7b8f30413b46?auto=format&fit=crop&w=800&q=80",
				},
			},
		}
	}

	for i, p := range prods {
		prod := supplierModels.SupplierProduct{
			SupplierProfileID: supplierProfileID,
			CategoryID:        cat.ID,
			Name:              p.Name,
			Description:       p.Description,
			MOQ:               p.MOQ,
			StartingPrice:     p.StartingPrice,
			Currency:          "IDR",
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
