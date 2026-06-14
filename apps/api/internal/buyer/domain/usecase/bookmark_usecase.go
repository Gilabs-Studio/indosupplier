package usecase

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

var (
	ErrBookmarkNotFound      = errors.New("bookmark not found")
	ErrBookmarkAlreadyExists = errors.New("supplier already bookmarked")
)

func slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

type BookmarkUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateBookmarkRequest) (*dto.BookmarkResponse, error)
	Delete(ctx context.Context, userID string, id string) error
	List(ctx context.Context, userID string) ([]dto.BookmarkResponse, error)
}

type bookmarkUsecase struct {
	db           *gorm.DB
	bookmarkRepo repositories.BookmarkRepository
}

func NewBookmarkUsecase(db *gorm.DB, bookmarkRepo repositories.BookmarkRepository) BookmarkUsecase {
	return &bookmarkUsecase{
		db:           db,
		bookmarkRepo: bookmarkRepo,
	}
}

func (u *bookmarkUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *bookmarkUsecase) Create(ctx context.Context, userID string, req *dto.CreateBookmarkRequest) (*dto.BookmarkResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify supplier exists
	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ?", req.SupplierProfileID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierProfileNotFound
		}
		return nil, err
	}

	var isProductBookmark bool
	var product supplierModels.SupplierProduct
	if req.SupplierProductID != nil && *req.SupplierProductID != "" {
		isProductBookmark = true
		if err := u.db.WithContext(ctx).Where("id = ? AND supplier_profile_id = ?", *req.SupplierProductID, req.SupplierProfileID).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("supplier product not found")
			}
			return nil, err
		}
	}

	// Check duplicate
	var existing buyerModels.Bookmark
	var queryErr error
	if isProductBookmark {
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND supplier_product_id = ?", buyerID, req.SupplierProfileID, *req.SupplierProductID).First(&existing).Error
	} else {
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND (supplier_product_id IS NULL OR supplier_product_id = '')", buyerID, req.SupplierProfileID).First(&existing).Error
	}

	if queryErr == nil {
		return nil, ErrBookmarkAlreadyExists
	}

	bookmark := &buyerModels.Bookmark{
		BuyerProfileID:    buyerID,
		SupplierProfileID: req.SupplierProfileID,
		Notes:             req.Notes,
	}
	if isProductBookmark {
		bookmark.SupplierProductID = req.SupplierProductID
	}

	if err := u.bookmarkRepo.Create(ctx, bookmark); err != nil {
		return nil, err
	}

	// Fetch category
	var categoryName string
	var supplierCat struct {
		Name string
	}
	err = u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("categories.name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id = ?", supplier.ID).
		Order("supplier_categories.is_primary DESC").
		Limit(1).
		Scan(&supplierCat).Error
	if err == nil {
		categoryName = supplierCat.Name
	}

	establishedYear, _ := strconv.Atoi(supplier.EstablishedYear)
	if establishedYear == 0 {
		establishedYear = 2015 // Default fallback
	}

	var bookmarkType string
	var prodName, prodMinOrder, prodImage string
	var prodPrice float64

	if isProductBookmark {
		bookmarkType = "product"
		prodName = product.Name
		prodPrice = product.StartingPrice
		prodMinOrder = product.MOQ
		
		var firstPhoto supplierModels.SupplierProductPhoto
		if err := u.db.WithContext(ctx).Where("supplier_product_id = ?", product.ID).Order("sort_order ASC").First(&firstPhoto).Error; err == nil {
			prodImage = firstPhoto.FileURL
		}
	} else {
		bookmarkType = "supplier"
	}

	// Fetch key products (only relevant for supplier list/cards)
	var productNames []string
	if !isProductBookmark {
		u.db.WithContext(ctx).
			Model(&supplierModels.SupplierProduct{}).
			Where("supplier_profile_id = ?", supplier.ID).
			Order("sort_order ASC").
			Limit(3).
			Pluck("name", &productNames)
	}

	return &dto.BookmarkResponse{
		ID:                bookmark.ID,
		SupplierProfileID: supplier.ID,
		SupplierProductID: bookmark.SupplierProductID,
		Type:              bookmarkType,
		SupplierSlug:      slugify(supplier.CompanyName),
		CompanyName:       supplier.CompanyName,
		Category:          categoryName,
		Location:          supplier.CityID,
		BusinessType:      supplier.CompanyType,
		EstablishedYear:   establishedYear,
		Rating:            supplier.StarRating,
		ReviewCount:       supplier.ReviewCount,
		IsVerified:        supplier.VerificationLevel >= 2,
		KeyProducts:       productNames,
		ProductName:       prodName,
		ProductPrice:      prodPrice,
		ProductMinOrder:   prodMinOrder,
		ProductImage:      prodImage,
		CreatedAt:         bookmark.CreatedAt,
	}, nil
}

func (u *bookmarkUsecase) Delete(ctx context.Context, userID string, id string) error {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if bookmark exists and belongs to this buyer
	_, err = u.bookmarkRepo.FindByIDAndBuyer(ctx, id, buyerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookmarkNotFound
		}
		return err
	}

	return u.bookmarkRepo.Delete(ctx, id, buyerID)
}

func (u *bookmarkUsecase) List(ctx context.Context, userID string) ([]dto.BookmarkResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	bookmarks, err := u.bookmarkRepo.List(ctx, buyerID)
	if err != nil {
		return nil, err
	}

	if len(bookmarks) == 0 {
		return []dto.BookmarkResponse{}, nil
	}

	var supplierIDs []string
	var productIDs []string
	for _, b := range bookmarks {
		supplierIDs = append(supplierIDs, b.SupplierProfileID)
		if b.SupplierProductID != nil && *b.SupplierProductID != "" {
			productIDs = append(productIDs, *b.SupplierProductID)
		}
	}

	// Batch fetch suppliers
	var suppliers []supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id IN ?", supplierIDs).Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[string]supplierModels.SupplierProfile)
	for _, s := range suppliers {
		supplierMap[s.ID] = s
	}

	// Batch fetch categories
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
		Order("supplier_categories.is_primary DESC, supplier_categories.created_at ASC").
		Scan(&catInfos)

	categoryMap := make(map[string]string)
	for _, info := range catInfos {
		if _, exists := categoryMap[info.SupplierProfileID]; !exists {
			categoryMap[info.SupplierProfileID] = info.CategoryName
		}
	}

	// Batch fetch key products for supplier cards
	type productInfo struct {
		SupplierProfileID string
		Name              string
	}
	var prodInfos []productInfo
	u.db.WithContext(ctx).
		Table("supplier_products").
		Select("supplier_profile_id, name").
		Where("supplier_profile_id IN ?", supplierIDs).
		Order("sort_order ASC, created_at DESC").
		Scan(&prodInfos)

	productsMap := make(map[string][]string)
	for _, info := range prodInfos {
		if len(productsMap[info.SupplierProfileID]) < 3 {
			productsMap[info.SupplierProfileID] = append(productsMap[info.SupplierProfileID], info.Name)
		}
	}

	// Batch fetch bookmarked products if any
	var products []supplierModels.SupplierProduct
	if len(productIDs) > 0 {
		if err := u.db.WithContext(ctx).Preload("Photos").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			return nil, err
		}
	}
	productMap := make(map[string]supplierModels.SupplierProduct)
	for _, p := range products {
		productMap[p.ID] = p
	}

	var responses []dto.BookmarkResponse
	for _, b := range bookmarks {
		s, exists := supplierMap[b.SupplierProfileID]
		if !exists {
			continue
		}

		catName := categoryMap[s.ID]
		if catName == "" {
			catName = "General"
		}

		prods := productsMap[s.ID]
		if prods == nil {
			prods = []string{}
		}

		establishedYear, _ := strconv.Atoi(s.EstablishedYear)
		if establishedYear == 0 {
			establishedYear = 2015
		}

		var bookmarkType string
		var prodName, prodMinOrder, prodImage string
		var prodPrice float64

		if b.SupplierProductID != nil && *b.SupplierProductID != "" {
			bookmarkType = "product"
			if p, ok := productMap[*b.SupplierProductID]; ok {
				prodName = p.Name
				prodPrice = p.StartingPrice
				prodMinOrder = p.MOQ
				if len(p.Photos) > 0 {
					prodImage = p.Photos[0].FileURL
				}
			}
		} else {
			bookmarkType = "supplier"
		}

		responses = append(responses, dto.BookmarkResponse{
			ID:                b.ID,
			SupplierProfileID: s.ID,
			SupplierProductID: b.SupplierProductID,
			Type:              bookmarkType,
			SupplierSlug:      slugify(s.CompanyName),
			CompanyName:       s.CompanyName,
			Category:          catName,
			Location:          s.CityID,
			BusinessType:      s.CompanyType,
			EstablishedYear:   establishedYear,
			Rating:            s.StarRating,
			ReviewCount:       s.ReviewCount,
			IsVerified:        s.VerificationLevel >= 2,
			KeyProducts:       prods,
			ProductName:       prodName,
			ProductPrice:      prodPrice,
			ProductMinOrder:   prodMinOrder,
			ProductImage:      prodImage,
			CreatedAt:         b.CreatedAt,
		})
	}

	return responses, nil
}
