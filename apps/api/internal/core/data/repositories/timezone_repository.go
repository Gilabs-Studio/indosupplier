package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"gorm.io/gorm"
)

// TimeZoneRepository defines the interface for timezone data operations
type TimeZoneRepository interface {
	// GetCurrentTimezone returns the current timezone info for a given zone name
	GetCurrentTimezone(ctx context.Context, zoneName string) (*models.TimezoneInfo, error)

	// GetTimezoneByCountry returns timezone info for a country
	GetTimezoneByCountry(ctx context.Context, countryCode string) ([]*models.TimezoneInfo, error)

	// DetectTimezoneFromCoordinates detects timezone based on lat/long coordinates
	// For Indonesia, uses longitude-based detection
	DetectTimezoneFromCoordinates(ctx context.Context, latitude, longitude float64) (*models.TimezoneInfo, error)

	// GetSupportedTimezones returns list of all supported timezone names
	GetSupportedTimezones(ctx context.Context) ([]string, error)
}

type timeZoneRepository struct {
	db *gorm.DB
}

// NewTimeZoneRepository creates a new TimeZoneRepository
func NewTimeZoneRepository(db *gorm.DB) TimeZoneRepository {
	return &timeZoneRepository{db: db}
}

func (r *timeZoneRepository) GetCurrentTimezone(ctx context.Context, zoneName string) (*models.TimezoneInfo, error) {
	now := time.Now().Unix()

	var tz models.TimeZone
	err := r.db.WithContext(ctx).
		Where("zone_name = ? AND time_start <= ?", zoneName, now).
		Order("time_start DESC").
		First(&tz).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("timezone not found: %s", zoneName)
		}
		return nil, err
	}

	// Get country name if available
	var country models.Country
	r.db.WithContext(ctx).
		Where("country_code = ?", tz.CountryCode).
		First(&country)

	return &models.TimezoneInfo{
		ZoneName:     tz.ZoneName,
		CountryCode:  tz.CountryCode,
		CountryName:  country.CountryName,
		Abbreviation: tz.Abbreviation,
		GMTOffset:    tz.GMTOffset,
		IsDST:        tz.DST == "1",
	}, nil
}

func (r *timeZoneRepository) GetTimezoneByCountry(ctx context.Context, countryCode string) ([]*models.TimezoneInfo, error) {
	now := time.Now().Unix()

	var timezones []models.TimeZone
	err := r.db.WithContext(ctx).
		Where("country_code = ? AND time_start <= ?", countryCode, now).
		Order("time_start DESC").
		Find(&timezones).Error

	if err != nil {
		return nil, err
	}

	var result []*models.TimezoneInfo
	seen := make(map[string]bool)

	for _, tz := range timezones {
		if !seen[tz.ZoneName] {
			seen[tz.ZoneName] = true
			result = append(result, &models.TimezoneInfo{
				ZoneName:     tz.ZoneName,
				CountryCode:  tz.CountryCode,
				Abbreviation: tz.Abbreviation,
				GMTOffset:    tz.GMTOffset,
				IsDST:        tz.DST == "1",
			})
		}
	}

	return result, nil
}

// Indonesia timezone boundaries based on longitude
var indonesiaTimezones = []struct {
	Name      string
	ZoneName  string
	MinLong   float64
	MaxLong   float64
	UTCOffset int
	Cities    []string
}{
	{
		Name:      "WIB",
		ZoneName:  "Asia/Jakarta",
		MinLong:   95.0,
		MaxLong:   120.0,
		UTCOffset: 7,
		Cities:    []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Palembang"},
	},
	{
		Name:      "WITA",
		ZoneName:  "Asia/Makassar",
		MinLong:   120.0,
		MaxLong:   135.0,
		UTCOffset: 8,
		Cities:    []string{"Makassar", "Denpasar", "Manado", "Palu"},
	},
	{
		Name:      "WIT",
		ZoneName:  "Asia/Jayapura",
		MinLong:   135.0,
		MaxLong:   141.0,
		UTCOffset: 9,
		Cities:    []string{"Jayapura", "Ambon", "Sorong"},
	},
}

func (r *timeZoneRepository) DetectTimezoneFromCoordinates(ctx context.Context, latitude, longitude float64) (*models.TimezoneInfo, error) {
	// For Indonesia, use longitude-based detection
	// Indonesia spans approximately 95°E to 141°E

	for _, tz := range indonesiaTimezones {
		if longitude >= tz.MinLong && longitude < tz.MaxLong {
			// Get full timezone info from database
			info, err := r.GetCurrentTimezone(ctx, tz.ZoneName)
			if err != nil {
				// Fallback to basic info if database lookup fails
				return &models.TimezoneInfo{
					ZoneName:     tz.ZoneName,
					CountryCode:  "ID",
					CountryName:  "Indonesia",
					Abbreviation: tz.Name,
					GMTOffset:    tz.UTCOffset * 3600,
					IsDST:        false,
				}, nil
			}
			return info, nil
		}
	}

	// Fallback to Asia/Jakarta for coordinates outside Indonesia
	return r.GetCurrentTimezone(ctx, "Asia/Jakarta")
}

func (r *timeZoneRepository) GetSupportedTimezones(ctx context.Context) ([]string, error) {
	var zoneNames []string

	err := r.db.WithContext(ctx).
		Model(&models.TimeZone{}).
		Select("DISTINCT zone_name").
		Pluck("zone_name", &zoneNames).Error

	return zoneNames, err
}

// RegisterTimezoneModels registers timezone models in the migration
func RegisterTimezoneModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.TimeZone{},
		&models.Country{},
	)
}
