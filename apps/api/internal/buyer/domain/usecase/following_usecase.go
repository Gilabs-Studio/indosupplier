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
	ErrFollowingNotFound      = errors.New("following not found")
	ErrFollowingAlreadyExists = errors.New("supplier already followed")
)

func slugifySupplierName(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type FollowingUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateFollowingRequest) (*dto.FollowingResponse, error)
	Delete(ctx context.Context, userID string, supplierProfileID string) error
	List(ctx context.Context, userID string) ([]dto.FollowingResponse, error)
}

type followingUsecase struct {
	db            *gorm.DB
	followingRepo repositories.FollowingRepository
}

func NewFollowingUsecase(db *gorm.DB, followingRepo repositories.FollowingRepository) FollowingUsecase {
	return &followingUsecase{
		db:            db,
		followingRepo: followingRepo,
	}
}

func (u *followingUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *followingUsecase) Create(ctx context.Context, userID string, req *dto.CreateFollowingRequest) (*dto.FollowingResponse, error) {
	buyerProfileID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ?", req.SupplierProfileID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierProfileNotFound
		}
		return nil, err
	}

	if _, err := u.followingRepo.FindByBuyerAndSupplier(ctx, buyerProfileID, req.SupplierProfileID); err == nil {
		return nil, ErrFollowingAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	following := &buyerModels.SupplierFollowing{
		BuyerProfileID:    buyerProfileID,
		SupplierProfileID: req.SupplierProfileID,
	}
	if err := u.followingRepo.Create(ctx, following); err != nil {
		return nil, err
	}

	response, err := u.buildFollowingResponse(ctx, *following, supplier)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (u *followingUsecase) Delete(ctx context.Context, userID string, supplierProfileID string) error {
	buyerProfileID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return err
	}

	if _, err := u.followingRepo.FindByBuyerAndSupplier(ctx, buyerProfileID, supplierProfileID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFollowingNotFound
		}
		return err
	}

	return u.followingRepo.DeleteByBuyerAndSupplier(ctx, buyerProfileID, supplierProfileID)
}

func (u *followingUsecase) List(ctx context.Context, userID string) ([]dto.FollowingResponse, error) {
	buyerProfileID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	followings, err := u.followingRepo.ListByBuyer(ctx, buyerProfileID)
	if err != nil {
		return nil, err
	}
	if len(followings) == 0 {
		return []dto.FollowingResponse{}, nil
	}

	supplierIDs := make([]string, 0, len(followings))
	for _, item := range followings {
		supplierIDs = append(supplierIDs, item.SupplierProfileID)
	}

	var suppliers []supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id IN ?", supplierIDs).Find(&suppliers).Error; err != nil {
		return nil, err
	}

	supplierMap := make(map[string]supplierModels.SupplierProfile, len(suppliers))
	for _, supplier := range suppliers {
		supplierMap[supplier.ID] = supplier
	}

	responses := make([]dto.FollowingResponse, 0, len(followings))
	for _, item := range followings {
		supplier, exists := supplierMap[item.SupplierProfileID]
		if !exists {
			continue
		}

		response, err := u.buildFollowingResponse(ctx, item, supplier)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (u *followingUsecase) buildFollowingResponse(
	ctx context.Context,
	following buyerModels.SupplierFollowing,
	supplier supplierModels.SupplierProfile,
) (dto.FollowingResponse, error) {
	categoryName := ""
	var categoryInfo struct {
		Name string
	}

	err := u.db.WithContext(ctx).
		Table("supplier_categories").
		Select("categories.name").
		Joins("join categories on categories.id = supplier_categories.category_id").
		Where("supplier_categories.supplier_profile_id = ?", supplier.ID).
		Order("supplier_categories.is_primary DESC, supplier_categories.created_at ASC").
		Limit(1).
		Scan(&categoryInfo).Error
	if err == nil {
		categoryName = categoryInfo.Name
	}

	var avatarURL string
	_ = u.db.WithContext(ctx).
		Table("users").
		Select("avatar_url").
		Where("id = ?", supplier.UserID).
		Limit(1).
		Row().
		Scan(&avatarURL)

	var keyProducts []dto.FollowingProductItem
	var dbProducts []supplierModels.SupplierProduct
	err = u.db.WithContext(ctx).
		Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("supplier_profile_id = ?", supplier.ID).
		Order("sort_order ASC").
		Limit(3).
		Find(&dbProducts).Error
	if err == nil {
		for _, p := range dbProducts {
			img := ""
			if len(p.Photos) > 0 {
				img = p.Photos[0].FileURL
			}
			keyProducts = append(keyProducts, dto.FollowingProductItem{
				ID:    p.ID,
				Name:  p.Name,
				Image: img,
			})
		}
	}
	if keyProducts == nil {
		keyProducts = []dto.FollowingProductItem{}
	}

	establishedYear, _ := strconv.Atoi(supplier.EstablishedYear)

	return dto.FollowingResponse{
		ID:                following.ID,
		SupplierProfileID: supplier.ID,
		SupplierSlug:      slugifySupplierName(supplier.CompanyName),
		CompanyName:       supplier.CompanyName,
		Category:          categoryName,
		Location:          supplier.CityID,
		BusinessType:      supplier.CompanyType,
		EstablishedYear:   establishedYear,
		Rating:            supplier.StarRating,
		ReviewCount:       supplier.ReviewCount,
		IsVerified:        supplier.VerificationLevel >= 2,
		KeyProducts:       keyProducts,
		Logo:              avatarURL,
		CreatedAt:         following.CreatedAt,
	}, nil
}
