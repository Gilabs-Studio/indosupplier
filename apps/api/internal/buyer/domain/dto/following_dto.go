package dto

import "time"

type CreateFollowingRequest struct {
	SupplierProfileID string `json:"supplierProfileId" binding:"required,uuid"`
}

type FollowingProductItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type FollowingResponse struct {
	ID                string                 `json:"id"`
	SupplierProfileID string                 `json:"supplierProfileId"`
	SupplierSlug      string                 `json:"supplierSlug"`
	CompanyName       string                 `json:"companyName"`
	Category          string                 `json:"category"`
	Location          string                 `json:"location"`
	BusinessType      string                 `json:"businessType"`
	EstablishedYear   int                    `json:"establishedYear"`
	Rating            float64                `json:"rating"`
	ReviewCount       int                    `json:"reviewCount"`
	IsVerified        bool                   `json:"isVerified"`
	KeyProducts       []FollowingProductItem `json:"keyProducts"`
	Logo              string                 `json:"logo"`
	CreatedAt         time.Time              `json:"createdAt"`
}
