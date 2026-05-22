package service

import (
	"context"
	"fmt"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
)

// TimezoneService provides timezone-related functionality
type TimezoneService interface {
	// DetectTimezoneFromCoordinates detects timezone based on lat/long
	DetectTimezoneFromCoordinates(ctx context.Context, latitude, longitude float64) (*coreModels.TimezoneInfo, error)

	// ValidateTimezone checks if a timezone name is valid
	ValidateTimezone(ctx context.Context, timezone string) error

	// GetTimezoneForCompany returns timezone for a company location
	GetTimezoneForCompany(ctx context.Context, latitude, longitude *float64, currentTimezone string) (string, error)
}

type timezoneService struct {
	timezoneRepo coreRepos.TimeZoneRepository
}

// NewTimezoneService creates a new TimezoneService
func NewTimezoneService(timezoneRepo coreRepos.TimeZoneRepository) TimezoneService {
	return &timezoneService{timezoneRepo: timezoneRepo}
}

func (s *timezoneService) DetectTimezoneFromCoordinates(ctx context.Context, latitude, longitude float64) (*coreModels.TimezoneInfo, error) {
	return s.timezoneRepo.DetectTimezoneFromCoordinates(ctx, latitude, longitude)
}

func (s *timezoneService) ValidateTimezone(ctx context.Context, timezone string) error {
	_, err := s.timezoneRepo.GetCurrentTimezone(ctx, timezone)
	return err
}

func (s *timezoneService) GetTimezoneForCompany(ctx context.Context, latitude, longitude *float64, currentTimezone string) (string, error) {
	// If timezone is already set, validate it
	if currentTimezone != "" {
		err := s.ValidateTimezone(ctx, currentTimezone)
		if err == nil {
			return currentTimezone, nil
		}
	}

	// Auto-detect from coordinates if available
	if latitude != nil && longitude != nil {
		info, err := s.DetectTimezoneFromCoordinates(ctx, *latitude, *longitude)
		if err != nil {
			return "", fmt.Errorf("failed to detect timezone: %w", err)
		}
		return info.ZoneName, nil
	}

	// Default to Asia/Jakarta
	return "Asia/Jakarta", nil
}

// Indonesia timezone info for reference
var IndonesiaTimezones = []struct {
	Code        string
	Name        string
	ZoneName    string
	UTCOffset   int
	Description string
}{
	{
		Code:        "WIB",
		Name:        "Western Indonesia Time",
		ZoneName:    "Asia/Jakarta",
		UTCOffset:   7,
		Description: "Sumatra, Java, Kalimantan Barat, Kalimantan Tengah",
	},
	{
		Code:        "WITA",
		Name:        "Central Indonesia Time",
		ZoneName:    "Asia/Makassar",
		UTCOffset:   8,
		Description: "Sulawesi, Bali, Nusa Tenggara, Kalimantan Timur/Utara/Selatan",
	},
	{
		Code:        "WIT",
		Name:        "Eastern Indonesia Time",
		ZoneName:    "Asia/Jayapura",
		UTCOffset:   9,
		Description: "Papua, Maluku",
	},
}
