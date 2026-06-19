package utils

import "testing"

func TestFormatMoneyUsesDefaultIDRFormat(t *testing.T) {
	got := FormatMoney(12500000, "")
	want := "Rp 12.500.000"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
