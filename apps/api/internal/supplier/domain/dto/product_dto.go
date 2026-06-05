package dto

import "time"

type PhotoDTO struct {
	FileURL   string `json:"file_url" binding:"required"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order"`
}

type CreateProductRequest struct {
	CategoryID    string     `json:"category_id" binding:"required"`
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description" binding:"required"`
	MOQ           string     `json:"moq"`
	StartingPrice float64    `json:"starting_price"`
	Currency      string     `json:"currency"`
	CapacityText  string     `json:"capacity_text"`
	IsFeatured    bool       `json:"is_featured"`
	SortOrder     int        `json:"sort_order"`
	Photos        []PhotoDTO `json:"photos"`
}

type UpdateProductRequest struct {
	CategoryID    string     `json:"category_id" binding:"required"`
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description" binding:"required"`
	MOQ           string     `json:"moq"`
	StartingPrice float64    `json:"starting_price"`
	Currency      string     `json:"currency"`
	CapacityText  string     `json:"capacity_text"`
	IsFeatured    bool       `json:"is_featured"`
	SortOrder     int        `json:"sort_order"`
	Photos        []PhotoDTO `json:"photos"`
}

type CategoryResponse struct {
	ID          string  `json:"id"`
	ParentID    *string `json:"parent_id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
}

type PhotoResponse struct {
	ID        string    `json:"id"`
	FileURL   string    `json:"file_url"`
	Caption   string    `json:"caption"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductResponse struct {
	ID                string            `json:"id"`
	SupplierProfileID string            `json:"supplier_profile_id"`
	CategoryID        string            `json:"category_id"`
	Category          *CategoryResponse `json:"category,omitempty"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	MOQ               string            `json:"moq"`
	StartingPrice     float64           `json:"starting_price"`
	Currency          string            `json:"currency"`
	CapacityText      string            `json:"capacity_text"`
	IsFeatured        bool              `json:"is_featured"`
	SortOrder         int               `json:"sort_order"`
	Photos            []PhotoResponse   `json:"photos"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type ListProductsRequest struct {
	Search     string `form:"search"`
	CategoryID string `form:"category_id"`
	Page       int    `form:"page,default=1"`
	PerPage    int    `form:"per_page,default=20"`
}
