package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

var (
	ErrMaxComparisonReached        = errors.New("maximum comparison limit of 5 suppliers reached")
	ErrMaxProductComparisonReached = errors.New("maximum comparison limit of 5 products reached")
)

type CompareUsecase interface {
	List(ctx context.Context, userID string) ([]dto.ComparedSupplierResponse, error)
	Add(ctx context.Context, userID string, supplierProfileID string) ([]dto.ComparedSupplierResponse, error)
	Delete(ctx context.Context, userID string, supplierProfileID string) ([]dto.ComparedSupplierResponse, error)

	ListProducts(ctx context.Context, userID string) ([]dto.ComparedProductResponse, error)
	AddProduct(ctx context.Context, userID string, supplierProductID string) ([]dto.ComparedProductResponse, error)
	DeleteProduct(ctx context.Context, userID string, supplierProductID string) ([]dto.ComparedProductResponse, error)
}

type compareUsecase struct {
	db *gorm.DB
}

func NewCompareUsecase(db *gorm.DB) CompareUsecase {
	return &compareUsecase{db: db}
}

func (u *compareUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *compareUsecase) getOrCreateSession(ctx context.Context, buyerProfileID string) (*buyerModels.ComparisonSession, error) {
	var session buyerModels.ComparisonSession
	err := u.db.WithContext(ctx).Where("buyer_profile_id = ?", buyerProfileID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			session = buyerModels.ComparisonSession{
				ID:             uuid.New().String(),
				BuyerProfileID: buyerProfileID,
				ShareToken:     uuid.New().String(),
			}
			if err := u.db.WithContext(ctx).Create(&session).Error; err != nil {
				return nil, err
			}
			return &session, nil
		}
		return nil, err
	}
	return &session, nil
}

func (u *compareUsecase) List(ctx context.Context, userID string) ([]dto.ComparedSupplierResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	var items []buyerModels.ComparisonSessionItem
	if err := u.db.WithContext(ctx).Where("comparison_session_id = ?", session.ID).Order("sort_order ASC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []dto.ComparedSupplierResponse{}, nil
	}

	var supplierIDs []string
	for _, item := range items {
		supplierIDs = append(supplierIDs, item.SupplierProfileID)
	}

	var suppliers []supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id IN ?", supplierIDs).Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[string]supplierModels.SupplierProfile)
	for _, s := range suppliers {
		supplierMap[s.ID] = s
	}

	// Fetch primary categories
	type supplierCatInfo struct {
		SupplierProfileID string
		CategoryName      string
	}
	var catInfos []supplierCatInfo
	u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("supplier_categories.supplier_profile_id, categories.name as category_name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id IN ?", supplierIDs).
		Order("supplier_categories.is_primary DESC").
		Scan(&catInfos)

	categoryMap := make(map[string]string)
	for _, info := range catInfos {
		if _, exists := categoryMap[info.SupplierProfileID]; !exists {
			categoryMap[info.SupplierProfileID] = info.CategoryName
		}
	}

	// Fetch first product of each supplier for MOQ & Capacity fallback
	var products []supplierModels.SupplierProduct
	u.db.WithContext(ctx).Where("supplier_profile_id IN ?", supplierIDs).Order("sort_order ASC").Find(&products)

	moqMap := make(map[string]string)
	capacityMap := make(map[string]string)
	for _, p := range products {
		if _, exists := moqMap[p.SupplierProfileID]; !exists {
			moqMap[p.SupplierProfileID] = p.MOQ
			capacityMap[p.SupplierProfileID] = p.CapacityText
		}
	}

	// Fetch approved certifications
	type certInfo struct {
		SupplierProfileID string
		Name              string
	}
	var certs []certInfo
	u.db.WithContext(ctx).
		Table("supplier_certifications").
		Select("supplier_certifications.supplier_profile_id, certifications.name").
		Joins("join certifications on certifications.id = supplier_certifications.certification_id").
		Where("supplier_certifications.supplier_profile_id IN ? AND supplier_certifications.status = ?", supplierIDs, "approved").
		Scan(&certs)

	certMap := make(map[string][]string)
	for _, c := range certs {
		certMap[c.SupplierProfileID] = append(certMap[c.SupplierProfileID], c.Name)
	}

	var responses []dto.ComparedSupplierResponse
	// Return in sorted session items order
	for _, item := range items {
		s, exists := supplierMap[item.SupplierProfileID]
		if !exists {
			continue
		}

		establishedYear, _ := strconv.Atoi(s.EstablishedYear)
		if establishedYear == 0 {
			establishedYear = 2015
		}

		moq := moqMap[s.ID]
		if moq == "" {
			moq = "100 Pcs"
		}

		capText := capacityMap[s.ID]
		if capText == "" {
			capText = "5 Ton / Bulan"
		}

		respHrs := s.AvgResponseTimeMinutes / 60
		if respHrs == 0 {
			respHrs = 1
		}
		responseTime := fmt.Sprintf("%d Jam", respHrs)

		certifications := certMap[s.ID]
		if certifications == nil {
			certifications = []string{}
		}

		type reviewWithBuyer struct {
			ID         string    `json:"id"`
			Rating     int       `json:"rating"`
			ReviewText string    `json:"review_text"`
			BuyerName  string    `json:"buyer_name"`
			CreatedAt  time.Time `json:"created_at"`
		}
		var dbReviews []reviewWithBuyer
		u.db.WithContext(ctx).
			Table("supplier_reviews").
			Select("supplier_reviews.id, supplier_reviews.rating, supplier_reviews.review_text, buyer_profiles.company_name as buyer_name, supplier_reviews.created_at").
			Joins("join buyer_profiles on buyer_profiles.id = supplier_reviews.buyer_profile_id").
			Where("supplier_reviews.supplier_profile_id = ? AND supplier_reviews.status = ?", s.ID, "approved").
			Order("supplier_reviews.created_at DESC").
			Scan(&dbReviews)

		var comparedReviews []dto.ComparedReviewResponse
		for _, r := range dbReviews {
			comparedReviews = append(comparedReviews, dto.ComparedReviewResponse{
				ID:         r.ID,
				BuyerName:  r.BuyerName,
				Rating:     r.Rating,
				ReviewText: r.ReviewText,
				CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		if comparedReviews == nil {
			comparedReviews = []dto.ComparedReviewResponse{}
		}

		responses = append(responses, dto.ComparedSupplierResponse{
			ID:              s.ID,
			CompanyName:     s.CompanyName,
			Slug:            slugify(s.CompanyName),
			Location:        s.CityID,
			BusinessType:    s.CompanyType,
			EstablishedYear: establishedYear,
			Rating:          s.StarRating,
			ReviewCount:     s.ReviewCount,
			Verified:        s.VerificationLevel >= 2,
			MOQ:             moq,
			ResponseTime:    responseTime,
			Capacity:        capText,
			Certifications:  certifications,
			Reviews:         comparedReviews,
		})
	}

	return responses, nil
}

func (u *compareUsecase) Add(ctx context.Context, userID string, supplierProfileID string) ([]dto.ComparedSupplierResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	// Check current count
	var count int64
	if err := u.db.WithContext(ctx).Model(&buyerModels.ComparisonSessionItem{}).Where("comparison_session_id = ?", session.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, ErrMaxComparisonReached
	}

	// Check duplicate
	var existing buyerModels.ComparisonSessionItem
	err = u.db.WithContext(ctx).Where("comparison_session_id = ? AND supplier_profile_id = ?", session.ID, supplierProfileID).First(&existing).Error
	if err == nil {
		return u.List(ctx, userID) // already exists
	}

	item := buyerModels.ComparisonSessionItem{
		ID:                  uuid.New().String(),
		ComparisonSessionID: session.ID,
		SupplierProfileID:   supplierProfileID,
		SortOrder:           int(count),
	}

	if err := u.db.WithContext(ctx).Create(&item).Error; err != nil {
		return nil, err
	}

	return u.List(ctx, userID)
}

func (u *compareUsecase) Delete(ctx context.Context, userID string, supplierProfileID string) ([]dto.ComparedSupplierResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	if err := u.db.WithContext(ctx).
		Where("comparison_session_id = ? AND supplier_profile_id = ?", session.ID, supplierProfileID).
		Delete(&buyerModels.ComparisonSessionItem{}).Error; err != nil {
		return nil, err
	}

	return u.List(ctx, userID)
}

func (u *compareUsecase) ListProducts(ctx context.Context, userID string) ([]dto.ComparedProductResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	var items []buyerModels.ComparisonProductSessionItem
	if err := u.db.WithContext(ctx).Where("comparison_session_id = ?", session.ID).Order("sort_order ASC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []dto.ComparedProductResponse{}, nil
	}

	var productIDs []string
	for _, item := range items {
		productIDs = append(productIDs, item.SupplierProductID)
	}

	var products []supplierModels.SupplierProduct
	if err := u.db.WithContext(ctx).Preload("Photos").Preload("Category").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}

	productMap := make(map[string]supplierModels.SupplierProduct)
	var supplierIDs []string
	for _, p := range products {
		productMap[p.ID] = p
		supplierIDs = append(supplierIDs, p.SupplierProfileID)
	}

	var suppliers []supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id IN ?", supplierIDs).Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[string]supplierModels.SupplierProfile)
	for _, s := range suppliers {
		supplierMap[s.ID] = s
	}

	var responses []dto.ComparedProductResponse
	for _, item := range items {
		p, exists := productMap[item.SupplierProductID]
		if !exists {
			continue
		}

		s, sExists := supplierMap[p.SupplierProfileID]
		if !sExists {
			continue
		}

		imageUrl := ""
		if len(p.Photos) > 0 {
			imageUrl = p.Photos[0].FileURL
		}

		catName := ""
		if p.Category != nil {
			catName = p.Category.Name
		}

		respHrs := s.AvgResponseTimeMinutes / 60
		if respHrs == 0 {
			respHrs = 1
		}
		responseTime := fmt.Sprintf("%d Jam", respHrs)

		type reviewWithBuyer struct {
			ID         string    `json:"id"`
			Rating     int       `json:"rating"`
			ReviewText string    `json:"review_text"`
			BuyerName  string    `json:"buyer_name"`
			CreatedAt  time.Time `json:"created_at"`
		}
		var dbReviews []reviewWithBuyer
		u.db.WithContext(ctx).
			Table("supplier_reviews").
			Select("supplier_reviews.id, supplier_reviews.rating, supplier_reviews.review_text, buyer_profiles.company_name as buyer_name, supplier_reviews.created_at").
			Joins("join buyer_profiles on buyer_profiles.id = supplier_reviews.buyer_profile_id").
			Where("supplier_reviews.supplier_profile_id = ? AND supplier_reviews.status = ?", s.ID, "approved").
			Order("supplier_reviews.created_at DESC").
			Scan(&dbReviews)

		var comparedReviews []dto.ComparedReviewResponse
		for _, r := range dbReviews {
			comparedReviews = append(comparedReviews, dto.ComparedReviewResponse{
				ID:         r.ID,
				BuyerName:  r.BuyerName,
				Rating:     r.Rating,
				ReviewText: r.ReviewText,
				CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		if comparedReviews == nil {
			comparedReviews = []dto.ComparedReviewResponse{}
		}

		responses = append(responses, dto.ComparedProductResponse{
			ID:                   p.ID,
			Name:                 p.Name,
			ImageUrl:             imageUrl,
			Price:                p.StartingPrice,
			MOQ:                  p.MOQ,
			Capacity:             p.CapacityText,
			CategoryName:         catName,
			Description:          p.Description,
			SupplierID:           s.ID,
			SupplierCompanyName:  s.CompanyName,
			SupplierSlug:         slugify(s.CompanyName),
			SupplierRating:       s.StarRating,
			SupplierReviewCount:  s.ReviewCount,
			SupplierVerified:     s.VerificationLevel >= 2,
			SupplierLocation:     s.CityID,
			SupplierResponseTime: responseTime,
			Reviews:              comparedReviews,
		})
	}

	return responses, nil
}

func (u *compareUsecase) AddProduct(ctx context.Context, userID string, supplierProductID string) ([]dto.ComparedProductResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	var count int64
	if err := u.db.WithContext(ctx).Model(&buyerModels.ComparisonProductSessionItem{}).Where("comparison_session_id = ?", session.ID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, ErrMaxProductComparisonReached
	}

	var existing buyerModels.ComparisonProductSessionItem
	err = u.db.WithContext(ctx).Where("comparison_session_id = ? AND supplier_product_id = ?", session.ID, supplierProductID).First(&existing).Error
	if err == nil {
		return u.ListProducts(ctx, userID)
	}

	item := buyerModels.ComparisonProductSessionItem{
		ID:                  uuid.New().String(),
		ComparisonSessionID: session.ID,
		SupplierProductID:   supplierProductID,
		SortOrder:           int(count),
	}

	if err := u.db.WithContext(ctx).Create(&item).Error; err != nil {
		return nil, err
	}

	return u.ListProducts(ctx, userID)
}

func (u *compareUsecase) DeleteProduct(ctx context.Context, userID string, supplierProductID string) ([]dto.ComparedProductResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := u.getOrCreateSession(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	if err := u.db.WithContext(ctx).
		Where("comparison_session_id = ? AND supplier_product_id = ?", session.ID, supplierProductID).
		Delete(&buyerModels.ComparisonProductSessionItem{}).Error; err != nil {
		return nil, err
	}

	return u.ListProducts(ctx, userID)
}
