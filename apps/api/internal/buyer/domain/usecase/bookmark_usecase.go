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
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/mapper"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
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
	Toggle(ctx context.Context, userID string, req *dto.CreateBookmarkRequest) (*dto.ToggleBookmarkResponse, error)
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
			// Auto-provision default BuyerProfile so user is never blocked
			var user userModels.User
			name := "Buyer User"
			if errUser := u.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; errUser == nil && user.Name != "" {
				name = user.Name
			}
			newBuyer := buyerModels.BuyerProfile{
				UserID:              userID,
				FullName:            name,
				CompanyName:         name + " Company",
				ProfileCompleteness: 50,
			}
			if errCreate := u.db.WithContext(ctx).Create(&newBuyer).Error; errCreate == nil {
				return newBuyer.ID, nil
			}
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

	supplierProfileID := req.GetSupplierProfileID()
	supplierProductID := req.GetSupplierProductID()

	if supplierProfileID == "" {
		return nil, errors.New("supplierProfileId is required")
	}

	// Verify supplier exists
	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ?", supplierProfileID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierProfileNotFound
		}
		return nil, err
	}

	var isProductBookmark bool
	var product supplierModels.SupplierProduct
	if supplierProductID != nil && *supplierProductID != "" {
		isProductBookmark = true
		if err := u.db.WithContext(ctx).Where("id = ? AND supplier_profile_id = ?", *supplierProductID, supplierProfileID).First(&product).Error; err != nil {
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
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND supplier_product_id = ?", buyerID, supplierProfileID, *supplierProductID).First(&existing).Error
	} else {
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND (supplier_product_id IS NULL OR supplier_product_id = '')", buyerID, supplierProfileID).First(&existing).Error
	}

	if queryErr == nil {
		return nil, ErrBookmarkAlreadyExists
	}

	bookmark := &buyerModels.Bookmark{
		BuyerProfileID:    buyerID,
		SupplierProfileID: supplierProfileID,
		Notes:             req.Notes,
	}
	if isProductBookmark {
		bookmark.SupplierProductID = supplierProductID
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

	resp := mapper.ToBookmarkResponse(
		bookmark,
		&supplier,
		categoryName,
		productNames,
		&product,
		establishedYear,
		slugify(supplier.CompanyName),
	)
	return &resp, nil
}

func (u *bookmarkUsecase) Delete(ctx context.Context, userID string, id string) error {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return err
	}

	// Try finding by primary key ID first
	var bookmark buyerModels.Bookmark
	err = u.db.WithContext(ctx).Where("id = ? AND buyer_profile_id = ?", id, buyerID).First(&bookmark).Error
	if err != nil {
		// Also try finding by supplier_product_id or supplier_profile_id
		err = u.db.WithContext(ctx).Where("(supplier_product_id = ? OR supplier_profile_id = ?) AND buyer_profile_id = ?", id, id, buyerID).First(&bookmark).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBookmarkNotFound
			}
			return err
		}
	}

	return u.bookmarkRepo.Delete(ctx, bookmark.ID, buyerID)
}

func (u *bookmarkUsecase) Toggle(ctx context.Context, userID string, req *dto.CreateBookmarkRequest) (*dto.ToggleBookmarkResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	supplierProfileID := req.GetSupplierProfileID()
	supplierProductID := req.GetSupplierProductID()

	if supplierProfileID == "" {
		return nil, errors.New("supplierProfileId is required")
	}

	// Check if already exists
	var existing buyerModels.Bookmark
	var queryErr error
	if supplierProductID != nil && *supplierProductID != "" {
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND supplier_product_id = ?", buyerID, supplierProfileID, *supplierProductID).First(&existing).Error
	} else {
		queryErr = u.db.WithContext(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ? AND (supplier_product_id IS NULL OR supplier_product_id = '')", buyerID, supplierProfileID).First(&existing).Error
	}

	if queryErr == nil {
		// Already exists -> Remove it
		if err := u.bookmarkRepo.Delete(ctx, existing.ID, buyerID); err != nil {
			return nil, err
		}
		return mapper.ToToggleBookmarkResponse(false, "removed", nil), nil
	}

	// Does not exist -> Create it
	created, err := u.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return mapper.ToToggleBookmarkResponse(true, "added", created), nil
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

		var productPtr *supplierModels.SupplierProduct
		if b.SupplierProductID != nil && *b.SupplierProductID != "" {
			if p, ok := productMap[*b.SupplierProductID]; ok {
				productPtr = &p
			}
		}

		resp := mapper.ToBookmarkResponse(
			&b,
			&s,
			catName,
			prods,
			productPtr,
			establishedYear,
			slugify(s.CompanyName),
		)
		responses = append(responses, resp)
	}

	return responses, nil
}
