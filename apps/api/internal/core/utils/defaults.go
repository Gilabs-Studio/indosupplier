package utils

import (
	"regexp"
	"strings"
)

const (
	DefaultCategoryName    = "Industrial Minerals"
	DefaultSupplierName    = "Unknown Supplier"
	DefaultMOQ             = "5 Ton"
	DefaultCapacityText    = "5 Ton / Bulan"
	DefaultQuotaLimitLabel = "Unlimited"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

