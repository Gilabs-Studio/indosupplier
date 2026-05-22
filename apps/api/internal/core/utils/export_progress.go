package utils

// NoopProgress is used in sync export mode where progress reporting is not needed.
func NoopProgress(_ int) {
	// Intentionally empty: sync exports do not publish progress updates.
}

// LinearProgress maps current page progress into a bounded range.
func LinearProgress(current, total, minValue, maxValue int) int {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}
	if total <= 0 {
		return minValue
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}

	span := maxValue - minValue
	if span <= 0 {
		return minValue
	}

	return minValue + (current*span)/total
}
