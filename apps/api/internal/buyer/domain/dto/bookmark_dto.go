package dto

import "time"

type CreateBookmarkRequest struct {
	SupplierProfileID      string  `json:"supplierProfileId"`
	SupplierProductID      *string `json:"supplierProductId"`
	SupplierProfileIDSnake string  `json:"supplier_profile_id"`
	SupplierProductIDSnake *string `json:"supplier_product_id"`
	Notes                  string  `json:"notes"`
}

func (r *CreateBookmarkRequest) GetSupplierProfileID() string {
	if r.SupplierProfileID != "" {
		return r.SupplierProfileID
	}
	return r.SupplierProfileIDSnake
}

func (r *CreateBookmarkRequest) GetSupplierProductID() *string {
	if r.SupplierProductID != nil && *r.SupplierProductID != "" {
		return r.SupplierProductID
	}
	if r.SupplierProductIDSnake != nil && *r.SupplierProductIDSnake != "" {
		return r.SupplierProductIDSnake
	}
	return nil
}

type ToggleBookmarkResponse struct {
	Bookmarked bool              `json:"bookmarked"`
	Action     string            `json:"action"` // "added" or "removed"
	Bookmark   *BookmarkResponse `json:"bookmark,omitempty"`
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
