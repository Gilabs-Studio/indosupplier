package utils

import "github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"

const (
	fallbackDefaultPerPage = 20
	fallbackMaxPerPage     = 20
)

type PaginationResult struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func DefaultPerPage() int {
	if config.AppConfig != nil && config.AppConfig.Pagination.DefaultPerPage > 0 {
		return config.AppConfig.Pagination.DefaultPerPage
	}
	return fallbackDefaultPerPage
}

func MaxPerPage() int {
	if config.AppConfig != nil && config.AppConfig.Pagination.MaxPerPage > 0 {
		return config.AppConfig.Pagination.MaxPerPage
	}
	return fallbackMaxPerPage
}

func NormalizePagination(page int, perPage int, defaultPerPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if defaultPerPage < 1 {
		defaultPerPage = DefaultPerPage()
	}
	maxPerPage := MaxPerPage()
	if defaultPerPage > maxPerPage {
		defaultPerPage = maxPerPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}

func PaginationOffset(page int, perPage int) int {
	return (page - 1) * perPage
}

func TotalPages(total int64, perPage int) int {
	if perPage < 1 {
		perPage = DefaultPerPage()
	}
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	if totalPages < 1 {
		return 1
	}
	return totalPages
}

func NewPaginationResult(page int, perPage int, total int64) *PaginationResult {
	return &PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: TotalPages(total, perPage),
	}
}
