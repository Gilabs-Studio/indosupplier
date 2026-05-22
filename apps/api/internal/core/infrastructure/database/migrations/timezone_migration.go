package migrations

import (
	"log"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"gorm.io/gorm"
)

// MigrateTimezoneTables creates timezone tables and inserts Indonesia data
func MigrateTimezoneTables(db *gorm.DB) error {
	log.Println("Migrating timezone tables...")

	// Create tables
	if err := db.AutoMigrate(
		&models.Country{},
		&models.TimeZone{},
	); err != nil {
		return err
	}

	// Insert Indonesia country
	indonesia := models.Country{
		CountryCode: "ID",
		CountryName: "Indonesia",
	}

	err := db.Where("country_code = ?", indonesia.CountryCode).
		FirstOrCreate(&indonesia).Error
	if err != nil {
		return err
	}

	// Insert Indonesia timezones
	indonesiaTimezones := []models.TimeZone{
		{
			ZoneName:     "Asia/Jakarta",
			CountryCode:  "ID",
			Abbreviation: "WIB",
			TimeStart:    0,
			GMTOffset:    25200, // UTC+7 in seconds
			DST:          "0",
		},
		{
			ZoneName:     "Asia/Makassar",
			CountryCode:  "ID",
			Abbreviation: "WITA",
			TimeStart:    0,
			GMTOffset:    28800, // UTC+8 in seconds
			DST:          "0",
		},
		{
			ZoneName:     "Asia/Jayapura",
			CountryCode:  "ID",
			Abbreviation: "WIT",
			TimeStart:    0,
			GMTOffset:    32400, // UTC+9 in seconds
			DST:          "0",
		},
	}

	for _, tz := range indonesiaTimezones {
		err := db.Where("zone_name = ? AND time_start = ?", tz.ZoneName, tz.TimeStart).
			FirstOrCreate(&tz).Error
		if err != nil {
			return err
		}
	}

	log.Println("Timezone migration completed")
	return nil
}
