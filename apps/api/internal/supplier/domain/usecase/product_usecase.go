package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/indosupplier/api/internal/core/utils"
	userRepo "github.com/gilabs/indosupplier/api/internal/user/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/mapper"
)

var (
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrProductNotFound         = errors.New("product not found")
	ErrProductUnauthorized     = errors.New("you do not own this product")
)

type ProductUsecase interface {
	List(ctx context.Context, userID string, req *dto.ListProductsRequest) ([]dto.ProductResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, userID string, id string) (*dto.ProductResponse, error)
	Create(ctx context.Context, userID string, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	Update(ctx context.Context, userID string, id string, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(ctx context.Context, userID string, id string) error
	ListCategories(ctx context.Context) ([]dto.CategoryResponse, error)
}

type productUsecase struct {
	userRepo    userRepo.UserRepository
	productRepo repositories.ProductRepository
}

func NewProductUsecase(userRepo userRepo.UserRepository, productRepo repositories.ProductRepository) ProductUsecase {
	return &productUsecase{
		userRepo:    userRepo,
		productRepo: productRepo,
	}
}

func (u *productUsecase) getSupplierProfileID(ctx context.Context, userID string) (string, error) {
	accountCtx, err := u.userRepo.FindAccountContext(ctx, userID)
	if err != nil {
		return "", err
	}
	if accountCtx.SupplierProfileID == "" {
		return "", ErrSupplierProfileNotFound
	}
	return accountCtx.SupplierProfileID, nil
}

func (u *productUsecase) List(ctx context.Context, userID string, req *dto.ListProductsRequest) ([]dto.ProductResponse, *utils.PaginationResult, error) {
	supplierProfileID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	products, total, err := u.productRepo.List(ctx, supplierProfileID, req.Search, req.CategoryID, req.Page, req.PerPage)
	if err != nil {
		return nil, nil, err
	}

	pagination := &utils.PaginationResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: int((total + int64(req.PerPage) - 1) / int64(req.PerPage)),
	}

	return mapper.ToProductListResponse(products), pagination, nil
}

func (u *productUsecase) GetByID(ctx context.Context, userID string, id string) (*dto.ProductResponse, error) {
	supplierProfileID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	p, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if p.SupplierProfileID != supplierProfileID {
		return nil, ErrProductUnauthorized
	}

	resp := mapper.ToProductResponse(p)
	return &resp, nil
}

func (u *productUsecase) Create(ctx context.Context, userID string, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	supplierProfileID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	photos := make([]models.SupplierProductPhoto, len(req.Photos))
	for i, photo := range req.Photos {
		photos[i] = models.SupplierProductPhoto{
			FileURL:   photo.FileURL,
			Caption:   photo.Caption,
			SortOrder: photo.SortOrder,
		}
	}

	p := &models.SupplierProduct{
		SupplierProfileID: supplierProfileID,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       req.Description,
		MOQ:               req.MOQ,
		StartingPrice:     req.StartingPrice,
		Currency:          req.Currency,
		CapacityText:      req.CapacityText,
		IsFeatured:        req.IsFeatured,
		SortOrder:         req.SortOrder,
		Photos:            photos,
	}

	if err := u.productRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	// Fetch created product to return preloaded info
	created, err := u.productRepo.FindByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	resp := mapper.ToProductResponse(created)
	return &resp, nil
}

func (u *productUsecase) Update(ctx context.Context, userID string, id string, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	supplierProfileID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	p, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if p.SupplierProfileID != supplierProfileID {
		return nil, ErrProductUnauthorized
	}

	photos := make([]models.SupplierProductPhoto, len(req.Photos))
	for i, photo := range req.Photos {
		photos[i] = models.SupplierProductPhoto{
			SupplierProductID: id,
			FileURL:           photo.FileURL,
			Caption:           photo.Caption,
			SortOrder:         photo.SortOrder,
		}
	}

	p.CategoryID = req.CategoryID
	p.Name = req.Name
	p.Description = req.Description
	p.MOQ = req.MOQ
	p.StartingPrice = req.StartingPrice
	p.Currency = req.Currency
	p.CapacityText = req.CapacityText
	p.IsFeatured = req.IsFeatured
	p.SortOrder = req.SortOrder
	p.Photos = photos

	if err := u.productRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	// Fetch updated product with associations
	updated, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := mapper.ToProductResponse(updated)
	return &resp, nil
}

func (u *productUsecase) Delete(ctx context.Context, userID string, id string) error {
	supplierProfileID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return err
	}

	p, err := u.productRepo.FindByID(ctx, id)
	if err != nil {
		return ErrProductNotFound
	}

	if p.SupplierProfileID != supplierProfileID {
		return ErrProductUnauthorized
	}

	return u.productRepo.Delete(ctx, id)
}

func (u *productUsecase) ListCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := u.productRepo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.CategoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = *mapper.ToCategoryResponse(&c)
	}

	return resp, nil
}
