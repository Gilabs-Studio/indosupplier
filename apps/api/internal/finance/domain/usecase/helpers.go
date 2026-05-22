package usecase

import (
	"errors"
	"strings"
	"time"
)

// parseDate parses a required date string in "2006-01-02" format.
// Returns an error if the date is empty or invalid.
func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("date is required")
	}
	return time.Parse("2006-01-02", value)
}

// parseAssetDateStrict parses a required asset date in "2006-01-02" format.
// Returns a user-friendly validation error when format is invalid.
func parseAssetDateStrict(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("date is required")
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return parsed, nil
}

// parseOptionalDate parses an optional date string in "2006-01-02" format.
// Returns nil if the date is empty; returns error if invalid format.
func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return &parsed, nil
}

// parseOptionalDateFilter parses an optional date filter string (used in List methods).
// Returns nil if input is nil or empty string. Used for startDate/endDate filtering.
func parseOptionalDateFilter(input *string, label string) (*time.Time, error) {
	if input == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, errors.New("invalid " + label + " format, use YYYY-MM-DD")
	}

	return &parsed, nil
}

// normalizePagination is already defined in common_helpers.go - use that instead
