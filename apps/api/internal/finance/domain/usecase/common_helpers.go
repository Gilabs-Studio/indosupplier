package usecase

import (
	"errors"
	"strings"
	"time"
)

// parseDateRequired parses a date in "2006-01-02" format, returning an error if empty.
func parseDateRequired(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("date is required")
	}
	return time.Parse("2006-01-02", value)
}

// parseDateOptional parses an optional date from a *string pointer.
// Returns nil if the pointer is nil or the value is empty.
func parseDateOptional(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// parseEndDateOptional parses an optional date and sets it to the end of the day (23:59:59).
func parseEndDateOptional(value *string) (*time.Time, error) {
	parsed, err := parseDateOptional(value)
	if err != nil || parsed == nil {
		return parsed, err
	}
	
	// Set to end of day
	eod := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
	return &eod, nil
}

// normalizePagination ensures page and perPage are within valid bounds.
func normalizePagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}
