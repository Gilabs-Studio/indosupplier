package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

type ReviewRepository interface {
	GetEligibleTransactions(ctx context.Context, buyerID string) ([]dto.EligibleTransactionResponse, error)
	GetReviewHistory(ctx context.Context, buyerID string) ([]dto.ReviewHistoryResponse, error)
	Create(ctx context.Context, review *trustModels.SupplierReview) error
	HasReviewed(ctx context.Context, buyerID string, purchaseOrderID string) (bool, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *reviewRepository) GetEligibleTransactions(ctx context.Context, buyerID string) ([]dto.EligibleTransactionResponse, error) {
	var results []dto.EligibleTransactionResponse
	err := r.getDB(ctx).
		Table("purchase_orders").
		Select("purchase_orders.id, purchase_orders.po_number, purchase_orders.supplier_profile_id, supplier_profiles.company_name as supplier_name, purchase_orders.product_name, purchase_orders.quantity_value, purchase_orders.quantity_unit, purchase_orders.total_amount, purchase_orders.updated_at as date").
		Joins("JOIN supplier_profiles ON supplier_profiles.id = purchase_orders.supplier_profile_id").
		Joins("LEFT JOIN supplier_reviews ON supplier_reviews.purchase_order_id = purchase_orders.id").
		Where("purchase_orders.buyer_profile_id = ? AND purchase_orders.status = ? AND supplier_reviews.id IS NULL", buyerID, "completed").
		Order("purchase_orders.updated_at DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	
	// Format dates to YYYY-MM-DD
	for i := range results {
		if len(results[i].Date) > 10 {
			results[i].Date = results[i].Date[:10]
		}
	}
	
	if results == nil {
		results = []dto.EligibleTransactionResponse{}
	}
	return results, nil
}

func (r *reviewRepository) GetReviewHistory(ctx context.Context, buyerID string) ([]dto.ReviewHistoryResponse, error) {
	type rawHistory struct {
		ID                string
		PONumber          string
		SupplierProfileID string
		SupplierName      string
		ProductName       string
		QuantityValue     float64
		QuantityUnit      string
		TotalAmount       float64
		Rating            int
		ReviewText        string
		Status            string
		CreatedAt         time.Time
	}
	
	var raws []rawHistory
	err := r.getDB(ctx).
		Table("supplier_reviews").
		Select("supplier_reviews.id, COALESCE(purchase_orders.po_number, 'N/A') as po_number, supplier_reviews.supplier_profile_id, supplier_profiles.company_name as supplier_name, COALESCE(purchase_orders.product_name, 'Sourcing Review') as product_name, COALESCE(purchase_orders.quantity_value, 0) as quantity_value, COALESCE(purchase_orders.quantity_unit, '') as quantity_unit, COALESCE(purchase_orders.total_amount, 0) as total_amount, supplier_reviews.rating, supplier_reviews.review_text, supplier_reviews.status, supplier_reviews.created_at").
		Joins("LEFT JOIN purchase_orders ON purchase_orders.id = supplier_reviews.purchase_order_id").
		Joins("JOIN supplier_profiles ON supplier_profiles.id = supplier_reviews.supplier_profile_id").
		Where("supplier_reviews.buyer_profile_id = ?", buyerID).
		Order("supplier_reviews.created_at DESC").
		Scan(&raws).Error

	if err != nil {
		return nil, err
	}

	results := make([]dto.ReviewHistoryResponse, len(raws))
	for i, raw := range raws {
		results[i] = dto.ReviewHistoryResponse{
			ID:                raw.ID,
			PONumber:          raw.PONumber,
			SupplierProfileID: raw.SupplierProfileID,
			SupplierName:      raw.SupplierName,
			ProductName:       raw.ProductName,
			QuantityValue:     raw.QuantityValue,
			QuantityUnit:      raw.QuantityUnit,
			TotalAmount:       raw.TotalAmount,
			Rating:            raw.Rating,
			ReviewText:        raw.ReviewText,
			Status:            raw.Status,
			CreatedAt:         raw.CreatedAt.Format("2006-01-02"),
		}
	}
	
	if results == nil {
		results = []dto.ReviewHistoryResponse{}
	}
	return results, nil
}

func (r *reviewRepository) Create(ctx context.Context, review *trustModels.SupplierReview) error {
	return r.getDB(ctx).Create(review).Error
}

func (r *reviewRepository) HasReviewed(ctx context.Context, buyerID string, purchaseOrderID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).
		Model(&trustModels.SupplierReview{}).
		Where("buyer_profile_id = ? AND purchase_order_id = ?", buyerID, purchaseOrderID).
		Count(&count).Error
	return count > 0, err
}
