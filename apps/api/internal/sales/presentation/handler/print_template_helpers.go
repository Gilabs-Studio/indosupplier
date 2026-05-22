package handler

import (
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// moneyFmt is a CLDR-compliant number printer for English locale (1,234.56)
var moneyFmt = message.NewPrinter(language.English)

// printFuncMap exposes formatting helpers to all HTML print templates in this package.
// Handlers use this shared FuncMap when parsing their respective template files.
var printFuncMap = template.FuncMap{
	"formatMoney": fmtMoney,
	"formatQty":   fmtQty,
	"formatTax":   fmtTax,
	// inc increments an integer index by 1 (for 1-based numbering in templates)
	"inc": func(i int) int { return i + 1 },
	// gt0 reports whether a float64 is strictly greater than zero
	"gt0": func(v float64) bool { return v > 0 },
}

// fmtMoney formats a float64 as a 2-decimal number with thousand separators.
func fmtMoney(v float64) string {
	return moneyFmt.Sprintf("%.2f", v)
}

// fmtQty formats a quantity, trimming trailing decimal zeros up to 3 places.
func fmtQty(v float64) string {
	s := fmt.Sprintf("%.3f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// fmtTax formats a tax rate, trimming trailing decimal zeros up to 2 places.
func fmtTax(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
