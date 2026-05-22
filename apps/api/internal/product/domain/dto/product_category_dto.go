package dto

import "time"

// === ProductCategory DTOs ===

type CreateProductCategoryRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=100"`
	Description  string  `json:"description" binding:"max=500"`
	CategoryType string  `json:"category_type" binding:"omitempty,oneof=GOODS FNB"`
	ParentID     *string `json:"parent_id"`
	IsActive     *bool   `json:"is_active"`
}

type UpdateProductCategoryRequest struct {
	Name         string  `json:"name" binding:"omitempty,min=2,max=100"`
	Description  string  `json:"description" binding:"max=500"`
	CategoryType string  `json:"category_type" binding:"omitempty,oneof=GOODS FNB"`
	ParentID     *string `json:"parent_id"`
	IsActive     *bool   `json:"is_active"`
}

type ProductCategoryResponse struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	CategoryType string                   `json:"category_type"`
	ParentID     *string                  `json:"parent_id"`
	Parent       *ProductCategoryResponse `json:"parent,omitempty"`
	IsActive     bool                     `json:"is_active"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

// CategoryTreeResponse represents a category in tree structure with product count
type CategoryTreeResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CategoryType string                 `json:"category_type"`
	ParentID     *string                `json:"parent_id"`
	ProductCount int64                  `json:"product_count"`
	Children     []CategoryTreeResponse `json:"children"`
	HasChildren  bool                   `json:"has_children"`
	IsActive     bool                   `json:"is_active"`
	Level        int                    `json:"level"`
}

// CategoryTreeParams parameters for fetching category tree
type CategoryTreeParams struct {
	ParentID      *string // nil = root categories
	MaxDepth      int     // 0 = unlimited, 1 = only direct children
	IncludeCount  bool    // include product count per category
	OnlyActive    bool    // filter only active categories
}
