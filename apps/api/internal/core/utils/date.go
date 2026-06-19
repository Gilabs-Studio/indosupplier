package utils

import "time"

const DisplayDateLayout = "January 2, 2006"

func FormatDisplayDate(t time.Time) string {
	return t.Format(DisplayDateLayout)
}
