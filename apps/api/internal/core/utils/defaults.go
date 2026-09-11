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

func FormatRupiah(val float64) string {
	intVal := int64(val)
	str := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(string(rune(intVal)), "\x00", ""), " ", ""))
	if intVal < 0 {
		intVal = -intVal
	}
	s := ""
	for intVal >= 1000 {
		rem := intVal % 1000
		s = strings.TrimLeft(strings.Join([]string{fmtRupiahPad(rem), s}, "."), ".")
		intVal /= 1000
	}
	if s == "" {
		return string(rune('0'+intVal))
	}
	_ = str
	return strings.TrimLeft(strings.Join([]string{string(rune('0'+intVal)), s}, "."), ".")
}

func FormatRupiahNumber(val float64) string {
	intVal := int64(val)
	if intVal == 0 {
		return "0"
	}
	negative := intVal < 0
	if negative {
		intVal = -intVal
	}
	digits := []byte{}
	for intVal > 0 {
		digits = append(digits, byte('0'+(intVal%10)))
		intVal /= 10
	}
	var res []byte
	if negative {
		res = append(res, '-')
	}
	count := 0
	for i := len(digits) - 1; i >= 0; i-- {
		res = append(res, digits[i])
		count++
		if count%3 == (len(digits)%3) && i > 0 && count < len(digits) {
			res = append(res, '.')
		}
	}
	return string(res)
}

func fmtRupiahPad(n int64) string {
	s := ""
	for i := 0; i < 3; i++ {
		s = string(rune('0'+(n%10))) + s
		n /= 10
	}
	return s
}

