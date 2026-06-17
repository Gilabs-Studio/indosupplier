package dto

type ContentArticleResponse struct {
	ID                string  `json:"id"`
	Type              string  `json:"type"`
	Locale            string  `json:"locale"`
	Title             string  `json:"title"`
	Slug              string  `json:"slug"`
	Excerpt           string  `json:"excerpt"`
	Body              string  `json:"body"`
	AuthorName        string  `json:"authorName"`
	ImageURL          string  `json:"imageUrl"`
	VideoURL          string  `json:"videoUrl"`
	Duration          string  `json:"duration"`
	ViewCount         int     `json:"viewCount"`
	SupplierProfileID *string `json:"supplierProfileId,omitempty"`
	SupplierProductID *string `json:"supplierProductId,omitempty"`
	Status            string  `json:"status"`
	IsFeatured        bool    `json:"isFeatured"`
	SortOrder         int     `json:"sortOrder"`
	PublishedAt       string  `json:"publishedAt,omitempty"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type ListContentArticlesRequest struct {
	Type    string `form:"type"`
	Locale  string `form:"locale"`
	Status  string `form:"status"`
	Search  string `form:"search"`
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
}

type CreateContentArticleRequest struct {
	Type              string  `json:"type" binding:"required,oneof=news feature tips editorial_review video"`
	Locale            string  `json:"locale" binding:"omitempty,oneof=id en both"`
	Title             string  `json:"title" binding:"required"`
	Slug              string  `json:"slug" binding:"omitempty"`
	Excerpt           string  `json:"excerpt" binding:"omitempty"`
	Body              string  `json:"body" binding:"omitempty"`
	AuthorName        string  `json:"authorName" binding:"omitempty"`
	ImageURL          string  `json:"imageUrl" binding:"omitempty,url"`
	VideoURL          string  `json:"videoUrl" binding:"omitempty,url"`
	Duration          string  `json:"duration" binding:"omitempty"`
	ViewCount         int     `json:"viewCount" binding:"omitempty,min=0"`
	SupplierProfileID *string `json:"supplierProfileId" binding:"omitempty,uuid"`
	SupplierProductID *string `json:"supplierProductId" binding:"omitempty,uuid"`
	Status            string  `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsFeatured        bool    `json:"isFeatured"`
	SortOrder         int     `json:"sortOrder"`
}

type UpdateContentArticleRequest struct {
	Type              *string `json:"type" binding:"omitempty,oneof=news feature tips editorial_review video"`
	Locale            *string `json:"locale" binding:"omitempty,oneof=id en both"`
	Title             *string `json:"title" binding:"omitempty"`
	Slug              *string `json:"slug" binding:"omitempty"`
	Excerpt           *string `json:"excerpt" binding:"omitempty"`
	Body              *string `json:"body" binding:"omitempty"`
	AuthorName        *string `json:"authorName" binding:"omitempty"`
	ImageURL          *string `json:"imageUrl" binding:"omitempty,url"`
	VideoURL          *string `json:"videoUrl" binding:"omitempty,url"`
	Duration          *string `json:"duration" binding:"omitempty"`
	ViewCount         *int    `json:"viewCount" binding:"omitempty,min=0"`
	SupplierProfileID *string `json:"supplierProfileId" binding:"omitempty,uuid"`
	SupplierProductID *string `json:"supplierProductId" binding:"omitempty,uuid"`
	Status            *string `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsFeatured        *bool   `json:"isFeatured"`
	SortOrder         *int    `json:"sortOrder"`
}
