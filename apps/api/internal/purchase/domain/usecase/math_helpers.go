package usecase

import "math"

// clampRange constrains val to [minVal, maxVal].
func clampRange(val, minVal, maxVal float64) float64 {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// round2dp rounds to 2 decimal places (for currency).
func round2dp(v float64) float64 {
	return math.Round(v*100) / 100
}
