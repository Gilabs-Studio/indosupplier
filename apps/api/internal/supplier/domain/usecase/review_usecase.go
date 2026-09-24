package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

var (
	ErrSupplierReviewProfileNotFound = errors.New("supplier profile not found")
	ErrSupplierReviewNotFound        = errors.New("review not found")
	ErrSupplierReviewNotApproved     = errors.New("review is not approved")
	ErrSupplierReviewEmptyReply      = errors.New("reply text cannot be empty")
)

type SupplierReviewUsecase interface {
	GetReviews(ctx context.Context, userID string, page, limit int, rating *int, status string, search string) (*dto.SupplierReviewsListResponseDto, error)
	ReplyReview(ctx context.Context, userID string, reviewID string, req *dto.SupplierReplyReviewRequestDto) (*dto.SupplierReviewItemDto, error)
}

type supplierReviewUsecase struct {
	portalRepo repositories.PortalRepository
	reviewRepo repositories.SupplierReviewRepository
}

func NewSupplierReviewUsecase(portalRepo repositories.PortalRepository, reviewRepo repositories.SupplierReviewRepository) SupplierReviewUsecase {
	return &supplierReviewUsecase{
		portalRepo: portalRepo,
		reviewRepo: reviewRepo,
	}
}

func (u *supplierReviewUsecase) GetReviews(ctx context.Context, userID string, page, limit int, rating *int, status string, search string) (*dto.SupplierReviewsListResponseDto, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierReviewProfileNotFound
		}
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	reviews, totalItems, err := u.reviewRepo.GetSupplierReviews(ctx, profile.ID, page, limit, rating, status, search)
	if err != nil {
		return nil, err
	}

	stats, err := u.reviewRepo.GetSupplierReviewStats(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	if totalPages < 1 && totalItems == 0 {
		totalPages = 1
	}

	return &dto.SupplierReviewsListResponseDto{
		Reviews: reviews,
		Stats:   *stats,
		Pagination: dto.SupplierReviewsPaginationDto{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			Limit:       limit,
		},
	}, nil
}

func (u *supplierReviewUsecase) ReplyReview(ctx context.Context, userID string, reviewID string, req *dto.SupplierReplyReviewRequestDto) (*dto.SupplierReviewItemDto, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierReviewProfileNotFound
		}
		return nil, err
	}

	trimmedReply := strings.TrimSpace(req.Reply)
	if len(trimmedReply) < 3 {
		return nil, ErrSupplierReviewEmptyReply
	}

	review, err := u.reviewRepo.GetReviewByID(ctx, reviewID, profile.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierReviewNotFound
		}
		return nil, err
	}

	if review.Status != "approved" {
		return nil, ErrSupplierReviewNotApproved
	}

	now := apptime.Now()
	if err := u.reviewRepo.UpdateReply(ctx, review.ID, trimmedReply, now); err != nil {
		return nil, fmt.Errorf("failed to update reply: %w", err)
	}

	// Create in-app notification for buyer
	notif := trustModels.Notification{
		RecipientType: "buyer",
		RecipientID:   review.BuyerProfileID,
		Type:          "review_reply",
		Title:         fmt.Sprintf("Balasan Ulasan dari %s", profile.CompanyName),
		Body:          fmt.Sprintf("%s telah membalas ulasan Anda: \"%s\"", profile.CompanyName, trimmedReply),
		Channel:       "in_app",
		IsRead:        false,
		RelatedType:   "review",
		RelatedID:     &review.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_ = u.reviewRepo.CreateNotification(ctx, &notif)

	// Fetch updated review item to return
	updatedItem, err := u.reviewRepo.GetSupplierReviewItemByID(ctx, review.ID, profile.ID)
	if err == nil && updatedItem != nil {
		return updatedItem, nil
	}

	return &dto.SupplierReviewItemDto{
		ID:                review.ID,
		BuyerProfileID:    review.BuyerProfileID,
		Rating:            review.Rating,
		ReviewText:        review.ReviewText,
		SupplierReply:     trimmedReply,
		SupplierRepliedAt: &now,
		Status:            review.Status,
		CreatedAt:         review.CreatedAt,
	}, nil
}
