package seeders

import (
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
	var count int64
	if err := database.DB.Model(&supplierModels.SupplierProfile{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return database.DB.Create(&supplierModels.SupplierProfile{
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
	}).Error
}
