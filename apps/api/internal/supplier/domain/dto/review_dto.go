package dto

import "time"

type SupplierReviewItemDto struct {
	ID                string     `json:"id"`
	BuyerProfileID    string     `json:"buyer_profile_id"`
	BuyerCompanyName  string     `json:"buyer_company_name"`
	BuyerUserName     string     `json:"buyer_user_name"`
	BuyerAvatarURL    string     `json:"buyer_avatar_url"`
	ProductID         *string    `json:"product_id,omitempty"`
	ProductName       string     `json:"product_name"`
	PurchaseOrderID   *string    `json:"purchase_order_id,omitempty"`
	PONumber          string     `json:"po_number"`
	Rating            int        `json:"rating"`
	ReviewText        string     `json:"review_text"`
	SupplierReply     string     `json:"supplier_reply"`
	SupplierRepliedAt *time.Time `json:"supplier_replied_at"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
}

type SupplierReviewStatsDto struct {
	TotalReviews    int64         `json:"total_reviews"`
	AverageRating   float64       `json:"average_rating"`
	RepliedCount    int64         `json:"replied_count"`
	UnrepliedCount  int64         `json:"unreplied_count"`
	RatingBreakdown map[int]int64 `json:"rating_breakdown"`
}

type SupplierReviewsPaginationDto struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	Limit       int   `json:"limit"`
}

type SupplierReviewsListResponseDto struct {
	Reviews    []SupplierReviewItemDto      `json:"reviews"`
	Stats      SupplierReviewStatsDto       `json:"stats"`
	Pagination SupplierReviewsPaginationDto `json:"pagination"`
}

type SupplierReplyReviewRequestDto struct {
	Reply string `json:"reply" binding:"required,min=3,max=1000"`
}
