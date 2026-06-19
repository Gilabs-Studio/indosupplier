package utils

import "testing"

func TestNormalizePaginationUsesDefaultAndMax(t *testing.T) {
	page, perPage := NormalizePagination(0, 100, 12)

	if page != 1 {
		t.Fatalf("expected page 1, got %d", page)
	}
	if perPage != 20 {
		t.Fatalf("expected perPage capped to 20, got %d", perPage)
	}
}

func TestTotalPagesNeverBelowOne(t *testing.T) {
	if got := TotalPages(0, 20); got != 1 {
		t.Fatalf("expected total pages 1 for empty result, got %d", got)
	}
}
