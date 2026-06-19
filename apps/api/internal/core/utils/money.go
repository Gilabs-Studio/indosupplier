package utils

import (
	"math"
	"strconv"
	"strings"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
)

const (
	fallbackCurrency          = "IDR"
	fallbackCurrencySymbol    = "Rp"
	fallbackThousandSeparator = "."
	fallbackDecimalSeparator  = ","
)

func DefaultCurrency() string {
	if config.AppConfig != nil && strings.TrimSpace(config.AppConfig.Localization.DefaultCurrency) != "" {
		return strings.TrimSpace(config.AppConfig.Localization.DefaultCurrency)
	}
	return fallbackCurrency
}

func FormatMoney(amount float64, currency string) string {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		currency = DefaultCurrency()
	}

	symbol := currency
	thousandSeparator := fallbackThousandSeparator
	decimalSeparator := fallbackDecimalSeparator
	if config.AppConfig != nil {
		if strings.EqualFold(currency, config.AppConfig.Localization.DefaultCurrency) {
			symbol = strings.TrimSpace(config.AppConfig.Localization.CurrencySymbol)
		}
		if config.AppConfig.Localization.ThousandSeparator != "" {
			thousandSeparator = config.AppConfig.Localization.ThousandSeparator
		}
		if config.AppConfig.Localization.DecimalSeparator != "" {
			decimalSeparator = config.AppConfig.Localization.DecimalSeparator
		}
	} else if strings.EqualFold(currency, fallbackCurrency) {
		symbol = fallbackCurrencySymbol
	}
	if symbol == "" {
		symbol = fallbackCurrencySymbol
	}

	return strings.TrimSpace(symbol + " " + FormatAmount(amount, 0, thousandSeparator, decimalSeparator))
}

func FormatAmount(amount float64, decimals int, thousandSeparator string, decimalSeparator string) string {
	if decimals < 0 {
		decimals = 0
	}
	if thousandSeparator == "" {
		thousandSeparator = fallbackThousandSeparator
	}
	if decimalSeparator == "" {
		decimalSeparator = fallbackDecimalSeparator
	}

	multiplier := math.Pow10(decimals)
	rounded := math.Round(amount*multiplier) / multiplier
	sign := ""
	if rounded < 0 {
		sign = "-"
		rounded = math.Abs(rounded)
	}

	raw := strconv.FormatFloat(rounded, 'f', decimals, 64)
	parts := strings.SplitN(raw, ".", 2)
	whole := groupThousands(parts[0], thousandSeparator)
	if decimals == 0 {
		return sign + whole
	}
	return sign + whole + decimalSeparator + parts[1]
}

func groupThousands(value string, separator string) string {
	if len(value) <= 3 {
		return value
	}
	var builder strings.Builder
	prefix := len(value) % 3
	if prefix == 0 {
		prefix = 3
	}
	builder.WriteString(value[:prefix])
	for i := prefix; i < len(value); i += 3 {
		builder.WriteString(separator)
		builder.WriteString(value[i : i+3])
	}
	return builder.String()
}
