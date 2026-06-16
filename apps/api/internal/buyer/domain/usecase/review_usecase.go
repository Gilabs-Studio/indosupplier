package usecase

import (
	"context"
	"errors"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

var (
	ErrTransactionNotCompleted = errors.New("transaction is not completed yet")
	ErrReviewAlreadyExists     = errors.New("review already exists for this transaction")
)

type ReviewUsecase interface {
	ListEligible(ctx context.Context, userID string) ([]dto.EligibleTransactionResponse, error)
	ListHistory(ctx context.Context, userID string) ([]dto.ReviewHistoryResponse, error)
	Create(ctx context.Context, userID string, req *dto.CreateReviewRequest) (*trustModels.SupplierReview, error)
}

type reviewUsecase struct {
	db         *gorm.DB
	reviewRepo repositories.ReviewRepository
}

func NewReviewUsecase(db *gorm.DB, reviewRepo repositories.ReviewRepository) ReviewUsecase {
	return &reviewUsecase{
		db:         db,
		reviewRepo: reviewRepo,
	}
}

func (u *reviewUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *reviewUsecase) ListEligible(ctx context.Context, userID string) ([]dto.EligibleTransactionResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.reviewRepo.GetEligibleTransactions(ctx, buyerID)
}

func (u *reviewUsecase) ListHistory(ctx context.Context, userID string) ([]dto.ReviewHistoryResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.reviewRepo.GetReviewHistory(ctx, buyerID)
}

func (u *reviewUsecase) Create(ctx context.Context, userID string, req *dto.CreateReviewRequest) (*trustModels.SupplierReview, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Check if the purchase order exists and belongs to the buyer
	var po buyerModels.PurchaseOrder
	if err := u.db.WithContext(ctx).Where("id = ? AND buyer_profile_id = ?", req.PurchaseOrderID, buyerID).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	// 2. Validate PO status is completed
	if po.Status != "completed" {
		return nil, ErrTransactionNotCompleted
	}

	// 3. Verify they haven't already reviewed this PO
	hasReviewed, err := u.reviewRepo.HasReviewed(ctx, buyerID, req.PurchaseOrderID)
	if err != nil {
		return nil, err
	}
	if hasReviewed {
		return nil, ErrReviewAlreadyExists
	}

	// 4. Create the review
	review := &trustModels.SupplierReview{
		BuyerProfileID:    buyerID,
		SupplierProfileID: po.SupplierProfileID,
		PurchaseOrderID:   &po.ID,
		Rating:            req.Rating,
		ReviewText:        req.ReviewText,
		Status:            "approved", // Immediately approve for real-time visibility in GIMS platform
	}
	if po.RFQID != nil {
		review.RFQID = *po.RFQID
	}

	if err := u.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// 5. Recalculate supplier average rating & review count
	var stats struct {
		AverageRating float64
		TotalCount    int64
	}
	
	// Query average rating and count of approved reviews for this supplier
	err = u.db.WithContext(ctx).
		Table("supplier_reviews").
		Select("COALESCE(AVG(rating), 0) as average_rating, COUNT(*) as total_count").
		Where("supplier_profile_id = ? AND status = ?", po.SupplierProfileID, "approved").
		Row().Scan(&stats.AverageRating, &stats.TotalCount)
		
	if err == nil {
		u.db.WithContext(ctx).
			Table("supplier_profiles").
			Where("id = ?", po.SupplierProfileID).
			Updates(map[string]interface{}{
				"star_rating":  stats.AverageRating,
				"review_count": int(stats.TotalCount),
			})
	}

	return review, nil
}
