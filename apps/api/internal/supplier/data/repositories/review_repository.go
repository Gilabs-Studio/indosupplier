package repositories

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

type SupplierReviewRepository interface {
	GetSupplierReviews(ctx context.Context, supplierProfileID string, page, limit int, rating *int, status string, search string) ([]dto.SupplierReviewItemDto, int64, error)
	GetSupplierReviewItemByID(ctx context.Context, reviewID string, supplierProfileID string) (*dto.SupplierReviewItemDto, error)
	GetSupplierReviewStats(ctx context.Context, supplierProfileID string) (*dto.SupplierReviewStatsDto, error)
	GetReviewByID(ctx context.Context, reviewID string, supplierProfileID string) (*trustModels.SupplierReview, error)
	UpdateReply(ctx context.Context, reviewID string, replyText string, repliedAt time.Time) error
	CreateNotification(ctx context.Context, notification *trustModels.Notification) error
}

type supplierReviewRepository struct {
	db *gorm.DB
}

func NewSupplierReviewRepository(db *gorm.DB) SupplierReviewRepository {
	return &supplierReviewRepository{db: db}
}

func (r *supplierReviewRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *supplierReviewRepository) GetSupplierReviews(ctx context.Context, supplierProfileID string, page, limit int, rating *int, status string, search string) ([]dto.SupplierReviewItemDto, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	baseQuery := r.getDB(ctx).Table("supplier_reviews sr").
		Joins("LEFT JOIN buyer_profiles bp ON bp.id = sr.buyer_profile_id").
		Joins("LEFT JOIN users u ON u.id = bp.user_id").
		Joins("LEFT JOIN purchase_orders po ON po.id = sr.purchase_order_id").
		Joins("LEFT JOIN supplier_products sp ON sp.id = sr.product_id").
		Where("sr.supplier_profile_id = ? AND sr.status = ? AND sr.deleted_at IS NULL", supplierProfileID, "approved")

	if rating != nil && *rating >= 1 && *rating <= 5 {
		baseQuery = baseQuery.Where("sr.rating = ?", *rating)
	}

	if status == "replied" {
		baseQuery = baseQuery.Where("sr.supplier_reply IS NOT NULL AND sr.supplier_reply != ''")
	} else if status == "unreplied" {
		baseQuery = baseQuery.Where("sr.supplier_reply IS NULL OR sr.supplier_reply = ''")
	}

	if search = strings.TrimSpace(search); search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		baseQuery = baseQuery.Where("(LOWER(sr.review_text) LIKE ? OR LOWER(bp.company_name) LIKE ? OR LOWER(u.name) LIKE ? OR LOWER(po.product_name) LIKE ? OR LOWER(sp.name) LIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	var totalItems int64
	if err := baseQuery.Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	type rawReview struct {
		ID                string
		BuyerProfileID    string
		BuyerCompanyName  *string
		BuyerUserName     *string
		BuyerAvatarURL    *string
		ProductID         *string
		ProductName       *string
		PurchaseOrderID   *string
		PONumber          *string
		Rating            int
		ReviewText        string
		SupplierReply     string
		SupplierRepliedAt *time.Time
		Status            string
		CreatedAt         time.Time
	}

	var raws []rawReview
	err := baseQuery.
		Select(`
			sr.id,
			sr.buyer_profile_id,
			bp.company_name AS buyer_company_name,
			u.name AS buyer_user_name,
			u.avatar_url AS buyer_avatar_url,
			sr.product_id,
			COALESCE(sp.name, po.product_name, 'Pesanan Produk') AS product_name,
			sr.purchase_order_id,
			po.po_number,
			sr.rating,
			sr.review_text,
			sr.supplier_reply,
			sr.supplier_replied_at,
			sr.status,
			sr.created_at
		`).
		Order("sr.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&raws).Error

	if err != nil {
		return nil, 0, err
	}

	items := make([]dto.SupplierReviewItemDto, len(raws))
	for i, raw := range raws {
		buyerComp := "Pembeli Terverifikasi"
		if raw.BuyerCompanyName != nil && *raw.BuyerCompanyName != "" {
			buyerComp = *raw.BuyerCompanyName
		}
		buyerUser := "Buyer"
		if raw.BuyerUserName != nil && *raw.BuyerUserName != "" {
			buyerUser = *raw.BuyerUserName
		}
		buyerAvatar := ""
		if raw.BuyerAvatarURL != nil {
			buyerAvatar = *raw.BuyerAvatarURL
		}
		prodName := "Pesanan Produk"
		if raw.ProductName != nil && *raw.ProductName != "" {
			prodName = *raw.ProductName
		}
		poNum := "-"
		if raw.PONumber != nil && *raw.PONumber != "" {
			poNum = *raw.PONumber
		}

		items[i] = dto.SupplierReviewItemDto{
			ID:                raw.ID,
			BuyerProfileID:    raw.BuyerProfileID,
			BuyerCompanyName:  buyerComp,
			BuyerUserName:     buyerUser,
			BuyerAvatarURL:    buyerAvatar,
			ProductID:         raw.ProductID,
			ProductName:       prodName,
			PurchaseOrderID:   raw.PurchaseOrderID,
			PONumber:          poNum,
			Rating:            raw.Rating,
			ReviewText:        raw.ReviewText,
			SupplierReply:     raw.SupplierReply,
			SupplierRepliedAt: raw.SupplierRepliedAt,
			Status:            raw.Status,
			CreatedAt:         raw.CreatedAt,
		}
	}

	return items, totalItems, nil
}

func (r *supplierReviewRepository) GetSupplierReviewStats(ctx context.Context, supplierProfileID string) (*dto.SupplierReviewStatsDto, error) {
	type statsResult struct {
		TotalReviews   int64   `gorm:"column:total_reviews"`
		AverageRating  float64 `gorm:"column:average_rating"`
		RepliedCount   int64   `gorm:"column:replied_count"`
		UnrepliedCount int64   `gorm:"column:unreplied_count"`
		Count5         int64   `gorm:"column:count_5"`
		Count4         int64   `gorm:"column:count_4"`
		Count3         int64   `gorm:"column:count_3"`
		Count2         int64   `gorm:"column:count_2"`
		Count1         int64   `gorm:"column:count_1"`
	}

	var res statsResult
	err := r.getDB(ctx).Table("supplier_reviews").
		Select(`
			COUNT(id) AS total_reviews,
			COALESCE(AVG(rating), 0) AS average_rating,
			COUNT(CASE WHEN supplier_reply IS NOT NULL AND supplier_reply != '' THEN 1 END) AS replied_count,
			COUNT(CASE WHEN supplier_reply IS NULL OR supplier_reply = '' THEN 1 END) AS unreplied_count,
			COUNT(CASE WHEN rating = 5 THEN 1 END) AS count_5,
			COUNT(CASE WHEN rating = 4 THEN 1 END) AS count_4,
			COUNT(CASE WHEN rating = 3 THEN 1 END) AS count_3,
			COUNT(CASE WHEN rating = 2 THEN 1 END) AS count_2,
			COUNT(CASE WHEN rating = 1 THEN 1 END) AS count_1
		`).
		Where("supplier_profile_id = ? AND status = ? AND deleted_at IS NULL", supplierProfileID, "approved").
		Scan(&res).Error

	if err != nil {
		return nil, err
	}

	breakdown := map[int]int64{
		5: res.Count5,
		4: res.Count4,
		3: res.Count3,
		2: res.Count2,
		1: res.Count1,
	}

	return &dto.SupplierReviewStatsDto{
		TotalReviews:    res.TotalReviews,
		AverageRating:   res.AverageRating,
		RepliedCount:    res.RepliedCount,
		UnrepliedCount:  res.UnrepliedCount,
		RatingBreakdown: breakdown,
	}, nil
}

func (r *supplierReviewRepository) GetSupplierReviewItemByID(ctx context.Context, reviewID string, supplierProfileID string) (*dto.SupplierReviewItemDto, error) {
	type rawReview struct {
		ID                string
		BuyerProfileID    string
		BuyerCompanyName  *string
		BuyerUserName     *string
		BuyerAvatarURL    *string
		ProductID         *string
		ProductName       *string
		PurchaseOrderID   *string
		PONumber          *string
		Rating            int
		ReviewText        string
		SupplierReply     string
		SupplierRepliedAt *time.Time
		Status            string
		CreatedAt         time.Time
	}

	var raw rawReview
	err := r.getDB(ctx).Table("supplier_reviews sr").
		Joins("LEFT JOIN buyer_profiles bp ON bp.id = sr.buyer_profile_id").
		Joins("LEFT JOIN users u ON u.id = bp.user_id").
		Joins("LEFT JOIN purchase_orders po ON po.id = sr.purchase_order_id").
		Joins("LEFT JOIN supplier_products sp ON sp.id = sr.product_id").
		Where("sr.id = ? AND sr.supplier_profile_id = ? AND sr.deleted_at IS NULL", reviewID, supplierProfileID).
		Select(`
			sr.id,
			sr.buyer_profile_id,
			bp.company_name AS buyer_company_name,
			u.name AS buyer_user_name,
			u.avatar_url AS buyer_avatar_url,
			sr.product_id,
			COALESCE(sp.name, po.product_name, 'Pesanan Produk') AS product_name,
			sr.purchase_order_id,
			po.po_number,
			sr.rating,
			sr.review_text,
			sr.supplier_reply,
			sr.supplier_replied_at,
			sr.status,
			sr.created_at
		`).
		First(&raw).Error

	if err != nil {
		return nil, err
	}

	buyerComp := "Pembeli Terverifikasi"
	if raw.BuyerCompanyName != nil && *raw.BuyerCompanyName != "" {
		buyerComp = *raw.BuyerCompanyName
	}
	buyerUser := "Buyer"
	if raw.BuyerUserName != nil && *raw.BuyerUserName != "" {
		buyerUser = *raw.BuyerUserName
	}
	buyerAvatar := ""
	if raw.BuyerAvatarURL != nil {
		buyerAvatar = *raw.BuyerAvatarURL
	}
	prodName := "Pesanan Produk"
	if raw.ProductName != nil && *raw.ProductName != "" {
		prodName = *raw.ProductName
	}
	poNum := "-"
	if raw.PONumber != nil && *raw.PONumber != "" {
		poNum = *raw.PONumber
	}

	return &dto.SupplierReviewItemDto{
		ID:                raw.ID,
		BuyerProfileID:    raw.BuyerProfileID,
		BuyerCompanyName:  buyerComp,
		BuyerUserName:     buyerUser,
		BuyerAvatarURL:    buyerAvatar,
		ProductID:         raw.ProductID,
		ProductName:       prodName,
		PurchaseOrderID:   raw.PurchaseOrderID,
		PONumber:          poNum,
		Rating:            raw.Rating,
		ReviewText:        raw.ReviewText,
		SupplierReply:     raw.SupplierReply,
		SupplierRepliedAt: raw.SupplierRepliedAt,
		Status:            raw.Status,
		CreatedAt:         raw.CreatedAt,
	}, nil
}

func (r *supplierReviewRepository) GetReviewByID(ctx context.Context, reviewID string, supplierProfileID string) (*trustModels.SupplierReview, error) {
	var review trustModels.SupplierReview
	err := r.getDB(ctx).
		Where("id = ? AND supplier_profile_id = ? AND deleted_at IS NULL", reviewID, supplierProfileID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *supplierReviewRepository) UpdateReply(ctx context.Context, reviewID string, replyText string, repliedAt time.Time) error {
	return r.getDB(ctx).
		Model(&trustModels.SupplierReview{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"supplier_reply":      replyText,
			"supplier_replied_at": repliedAt,
		}).Error
}

func (r *supplierReviewRepository) CreateNotification(ctx context.Context, notification *trustModels.Notification) error {
	return r.getDB(ctx).Create(notification).Error
}
