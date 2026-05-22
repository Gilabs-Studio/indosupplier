package usecase

import (
	"errors"
	"math"
	"strings"
	"time"
)

// clamp returns val clamped between minVal and maxVal.
func clamp(val, minVal, maxVal float64) float64 {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// parseOptionalDateFilter parses an optional date filter string (used in List methods).
// Returns nil if input is nil or empty string.
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

// normalizePagination enforces pagination constraints:
// page < 1 becomes 1, perPage < 1 becomes 10, perPage > 100 becomes 100.
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

// roundTo2Decimals rounds a float64 to exactly 2 decimal places.
func roundTo2Decimals(value float64) float64 {
	return math.Round(value*100) / 100
}
