package dto

import "time"

type BuyerProfileResponse struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	Email               string     `json:"email"`
	FullName            string     `json:"full_name"`
	CompanyName         string     `json:"company_name"`
	CountryCode         string     `json:"country_code"`
	Industry            string     `json:"industry"`
	PurchaseFrequency   string     `json:"purchase_frequency,omitempty"`
	Phone               string     `json:"phone,omitempty"`
	Website             string     `json:"website,omitempty"`
	Address             string     `json:"address,omitempty"`
	ProfileCompleteness int        `json:"profile_completeness"`
	CompanyVerifiedAt   *time.Time `json:"company_verified_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type UpdateBuyerPersonalRequest struct {
	FullName string `json:"full_name" binding:"required,min=3,max=100"`
	Phone    string `json:"phone" binding:"required,min=8,max=20"`
}

type UpdateBuyerCompanyRequest struct {
	CompanyName string `json:"company_name" binding:"required,min=3,max=100"`
	Industry    string `json:"industry" binding:"required"`
	Website     string `json:"website" binding:"omitempty,url"`
	Address     string `json:"address" binding:"required,min=5,max=500"`
}

type BuyerDocumentResponse struct {
	ID             string     `json:"id"`
	BuyerProfileID string     `json:"buyer_profile_id"`
	DocumentType   string     `json:"document_type"`
	DocumentNumber string     `json:"document_number"`
	FileURL        string     `json:"file_url"`
	Status         string     `json:"status"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	ReviewReason   string     `json:"review_reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type UploadBuyerDocumentRequest struct {
	DocumentType   string `json:"document_type" binding:"required,min=2,max=80"`
	DocumentNumber string `json:"document_number" binding:"required,max=120"`
	FileURL        string `json:"file_url" binding:"required,url"`
}
