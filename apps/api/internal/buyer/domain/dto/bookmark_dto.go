package dto

import "time"

type CreateBookmarkRequest struct {
	SupplierProfileID string  `json:"supplierProfileId" binding:"required,uuid"`
	SupplierProductID *string `json:"supplierProductId" binding:"omitempty,uuid"`
	Notes             string  `json:"notes"`
}

type BookmarkResponse struct {
	ID                string    `json:"id"`
	SupplierProfileID string    `json:"supplierProfileId"`
	SupplierProductID *string   `json:"supplierProductId,omitempty"`
	Type              string    `json:"type"` // "supplier" or "product"
	SupplierSlug      string    `json:"supplierSlug"`
	CompanyName       string    `json:"companyName"`
	Category          string    `json:"category"`
	Location          string    `json:"location"`
	BusinessType      string    `json:"businessType"`
	EstablishedYear   int       `json:"establishedYear"`
	Rating            float64   `json:"rating"`
	ReviewCount       int       `json:"reviewCount"`
	IsVerified        bool      `json:"isVerified"`
	KeyProducts       []string  `json:"keyProducts,omitempty"`
	
	// Product specific details
	ProductName       string    `json:"productName,omitempty"`
	ProductPrice      float64   `json:"productPrice,omitempty"`
	ProductMinOrder   string    `json:"productMinOrder,omitempty"`
	ProductImage      string    `json:"productImage,omitempty"`
	
	CreatedAt         time.Time `json:"created_at"`
}
