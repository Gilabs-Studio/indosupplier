package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/storage"
	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
	"github.com/gilabs/gims/api/internal/product/domain/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// mpnPattern enforces that MPN only contains safe printable characters (#353)
var mpnPattern = regexp.MustCompile(`^[A-Za-z0-9\-\.\/]+$`)

// ProductUsecase defines the interface for product business logic
type ProductUsecase interface {
	Create(ctx context.Context, req dto.CreateProductRequest, userID string) (dto.ProductResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductResponse, error)
	List(ctx context.Context, params repositories.ProductListParams) ([]dto.ProductResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProductRequest) (dto.ProductResponse, error)
	Delete(ctx context.Context, id string) error
	Submit(ctx context.Context, id string) (dto.ProductResponse, error)
	Approve(ctx context.Context, id string, userID string, req dto.ApproveProductRequest) (dto.ProductResponse, error)
	GetRecipe(ctx context.Context, id string) ([]dto.RecipeItemResponse, error)
	UpdateRecipe(ctx context.Context, id string, items []dto.RecipeItemRequest) ([]dto.RecipeItemResponse, error)
	ListRecipeVersions(ctx context.Context, id string) ([]dto.RecipeVersionResponse, error)
	CloneRecipeFromVersion(ctx context.Context, id string, req dto.CloneRecipeRequest) ([]dto.RecipeItemResponse, error)
	CompareRecipeVersions(ctx context.Context, id string, fromVersionID string, toVersionID string) (dto.RecipeVersionCompareResponse, error)
}

type productUsecase struct {
	db           *gorm.DB
	repo         repositories.ProductRepository
	categoryRepo repositories.ProductCategoryRepository
}

// NewProductUsecase creates a new ProductUsecase
func NewProductUsecase(db *gorm.DB, repo repositories.ProductRepository, categoryRepo repositories.ProductCategoryRepository) ProductUsecase {
	return &productUsecase{
		db:           db,
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

func (u *productUsecase) Create(ctx context.Context, req dto.CreateProductRequest, userID string) (dto.ProductResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Validate MPN format (#353)
	if req.ManufacturerPartNumber != "" && !mpnPattern.MatchString(req.ManufacturerPartNumber) {
		return dto.ProductResponse{}, errors.New("manufacturer_part_number may only contain letters, digits, hyphens, dots, or slashes")
	}

	// Generate Product Code
	prefix := "PRD"
	if req.CategoryID != nil {
		category, err := u.categoryRepo.FindByID(ctx, *req.CategoryID)
		if err == nil && category != nil {
			// Generate prefix from Category Name (uppercase, no spaces, max 3 chars)
			cleanedName := strings.ReplaceAll(strings.ToUpper(category.Name), " ", "")
			runes := []rune(cleanedName)
			if len(runes) >= 3 {
				prefix = string(runes[:3])
			} else if len(runes) > 0 {
				prefix = string(runes)
			}
		}
	}

	var imageURL *string
	if req.ImageURL != "" {
		imageURL = &req.ImageURL
	}

	// Format: PREFIX-YYYYMMDD-RAND4 (e.g., MED-20231023-A1B2)
	timestamp := apptime.Now().Format("20060102")
	randomSuffix := strings.ToUpper(uuid.New().String()[:4])
	generatedCode := fmt.Sprintf("%s-%s-%s", prefix, timestamp, randomSuffix)

	// Determine ProductKind defaults
	productKind := models.ProductKindStock
	if req.ProductKind != "" {
		productKind = req.ProductKind
	}

	// IsInventoryTracked defaults: true for STOCK, false for RECIPE/SERVICE
	isInventoryTracked := productKind == models.ProductKindStock
	if req.IsInventoryTracked != nil {
		isInventoryTracked = *req.IsInventoryTracked
	}

	product := &models.Product{
		ID:                     uuid.New().String(),
		Code:                   generatedCode,
		Name:                   req.Name,
		Description:            req.Description,
		ImageURL:               imageURL,
		ManufacturerPartNumber: req.ManufacturerPartNumber,
		CategoryID:             req.CategoryID,
		BrandID:                req.BrandID,
		SegmentID:              req.SegmentID,
		TypeID:                 req.TypeID,
		UomID:                  req.UomID,
		PurchaseUomID:          req.PurchaseUomID,
		PurchaseUomConversion:  req.PurchaseUomConversion,
		PackagingID:            req.PackagingID,
		ProcurementTypeID:      req.ProcurementTypeID,
		SupplierID:             req.SupplierID,
		BusinessUnitID:         req.BusinessUnitID,
		CostPrice:              req.CostPrice,
		SellingPrice:           req.SellingPrice,
		MinStock:               req.MinStock,
		MaxStock:               req.MaxStock,
		TaxType:                req.TaxType,
		IsTaxInclusive:         req.IsTaxInclusive,
		LeadTimeDays:           req.LeadTimeDays,
		Barcode:                req.Barcode,
		Sku:                    generatedCode,
		Weight:                 req.Weight,
		Volume:                 req.Volume,
		Notes:                  req.Notes,
		Status:                 models.ProductStatusApproved,
		IsApproved:             true,
		CreatedBy:              &userID,
		IsActive:               isActive,
		ProductKind:            productKind,
		IsIngredient:           req.IsIngredient,
		IsInventoryTracked:     isInventoryTracked,
		POSScope:               req.POSScope,
	}

	// Validate recipe items for RECIPE kind
	if productKind == models.ProductKindRecipe && len(req.RecipeItems) == 0 {
		return dto.ProductResponse{}, errors.New("RECIPE products must have at least one recipe item")
	}

	if err := u.repo.Create(ctx, product); err != nil {
		return dto.ProductResponse{}, err
	}

	// Create recipe items if present
	if len(req.RecipeItems) > 0 {
		if err := u.saveRecipeItems(product.ID, req.RecipeItems); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to save recipe items: %w", err)
		}
		if err := u.createRecipeVersion(product.ID, req.RecipeItems, nil, models.RecipeVersionChangeManual, "initial recipe version"); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to create initial recipe version: %w", err)
		}
	}

	// Sync outlet assignments if provided
	if req.OutletIDs != nil {
		if err := u.syncOutlets(ctx, product.ID, req.OutletIDs); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to sync outlets: %w", err)
		}
	}

	// Reload to get relations
	product, err := u.repo.FindByID(ctx, product.ID)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.ToProductResponse(product), nil
}

func (u *productUsecase) GetByID(ctx context.Context, id string) (dto.ProductResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	return mapper.ToProductResponse(product), nil
}

func (u *productUsecase) List(ctx context.Context, params repositories.ProductListParams) ([]dto.ProductResponse, int64, error) {
	products, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProductResponseList(products), total, nil
}

func (u *productUsecase) Update(ctx context.Context, id string, req dto.UpdateProductRequest) (dto.ProductResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	// Validate MPN format (#353)
	if req.ManufacturerPartNumber != "" && !mpnPattern.MatchString(req.ManufacturerPartNumber) {
		return dto.ProductResponse{}, errors.New("manufacturer_part_number may only contain letters, digits, hyphens, dots, or slashes")
	}

	if req.Code != "" {
		product.Code = req.Code
	}
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.ImageURL != nil {
		// If image is being deleted (set to empty string) or replaced with new URL
		if product.ImageURL != nil && *product.ImageURL != "" {
			// Check if image URL is different (being replaced) or being deleted
			if *req.ImageURL == "" || *product.ImageURL != *req.ImageURL {
				// Delete old image from R2 (best-effort, don't fail the update if deletion fails)
				key := storage.KeyFromURL(*product.ImageURL)
				if key != "" {
					_ = storage.Delete(ctx, key)
				}
			}
		}
		// Set new image URL
		if *req.ImageURL == "" {
			product.ImageURL = nil
		} else {
			product.ImageURL = req.ImageURL
		}
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.BrandID != nil {
		product.BrandID = req.BrandID
	}
	if req.SegmentID != nil {
		product.SegmentID = req.SegmentID
	}
	if req.TypeID != nil {
		product.TypeID = req.TypeID
	}
	if req.UomID != nil {
		product.UomID = req.UomID
	}
	if req.PackagingID != nil {
		product.PackagingID = req.PackagingID
	}
	if req.ProcurementTypeID != nil {
		product.ProcurementTypeID = req.ProcurementTypeID
	}
	if req.SupplierID != nil {
		product.SupplierID = req.SupplierID
	}
	if req.BusinessUnitID != nil {
		product.BusinessUnitID = req.BusinessUnitID
	}
	if req.CostPrice != nil {
		product.CostPrice = *req.CostPrice
	}
	if req.SellingPrice != nil {
		product.SellingPrice = *req.SellingPrice
	}
	if req.MinStock != nil {
		product.MinStock = *req.MinStock
	}
	if req.MaxStock != nil {
		product.MaxStock = *req.MaxStock
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.Sku != "" {
		product.Sku = req.Sku
	}
	if req.Weight != nil {
		product.Weight = *req.Weight
	}
	if req.Volume != nil {
		product.Volume = *req.Volume
	}
	if req.Notes != "" {
		product.Notes = req.Notes
	}
	if req.ManufacturerPartNumber != "" {
		product.ManufacturerPartNumber = req.ManufacturerPartNumber
	}
	if req.PurchaseUomID != nil {
		product.PurchaseUomID = req.PurchaseUomID
	}
	if req.PurchaseUomConversion != nil {
		product.PurchaseUomConversion = *req.PurchaseUomConversion
	}
	if req.TaxType != "" {
		product.TaxType = req.TaxType
	}
	if req.IsTaxInclusive != nil {
		product.IsTaxInclusive = *req.IsTaxInclusive
	}
	if req.LeadTimeDays != nil {
		product.LeadTimeDays = *req.LeadTimeDays
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if req.ProductKind != "" {
		product.ProductKind = req.ProductKind
	}
	if req.IsIngredient != nil {
		product.IsIngredient = *req.IsIngredient
	}
	if req.IsInventoryTracked != nil {
		product.IsInventoryTracked = *req.IsInventoryTracked
	}
	if req.POSScope != nil {
		product.POSScope = *req.POSScope
	}

	// Validate recipe items for RECIPE kind
	if product.ProductKind == models.ProductKindRecipe && len(req.RecipeItems) > 0 {
		// Replace pattern: delete existing + create new
		if err := u.replaceRecipeItems(product.ID, req.RecipeItems); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to update recipe items: %w", err)
		}
		if err := u.createRecipeVersion(product.ID, req.RecipeItems, nil, models.RecipeVersionChangeManual, "updated from product update"); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to create recipe version: %w", err)
		}
	}

	// Reset status if was rejected
	if product.Status == models.ProductStatusRejected {
		product.Status = models.ProductStatusDraft
	}

	if err := u.repo.Update(ctx, product); err != nil {
		return dto.ProductResponse{}, err
	}

	// Sync outlet assignments if provided in the update request
	if req.OutletIDs != nil {
		if err := u.syncOutlets(ctx, product.ID, req.OutletIDs); err != nil {
			return dto.ProductResponse{}, fmt.Errorf("failed to sync outlets: %w", err)
		}
	}

	// Reload to get relations
	product, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.ToProductResponse(product), nil
}

func (u *productUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("product not found")
	}

	// Soft delete product
	return u.repo.Delete(ctx, id)
}

func (u *productUsecase) Submit(ctx context.Context, id string) (dto.ProductResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	if product.Status != models.ProductStatusDraft && product.Status != models.ProductStatusRejected {
		return dto.ProductResponse{}, errors.New("can only submit draft or rejected products")
	}

	product.Status = models.ProductStatusPending

	if err := u.repo.Update(ctx, product); err != nil {
		return dto.ProductResponse{}, err
	}

	// Reload to get relations
	product, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.ToProductResponse(product), nil
}

func (u *productUsecase) Approve(ctx context.Context, id string, userID string, req dto.ApproveProductRequest) (dto.ProductResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	if product.Status != models.ProductStatusPending {
		return dto.ProductResponse{}, errors.New("can only approve/reject pending products")
	}

	now := apptime.Now()

	if req.Action == "approve" {
		product.Status = models.ProductStatusApproved
		product.IsApproved = true
		product.ApprovedBy = &userID
		product.ApprovedAt = &now
	} else {
		product.Status = models.ProductStatusRejected
		product.IsApproved = false
	}

	if err := u.repo.Update(ctx, product); err != nil {
		return dto.ProductResponse{}, err
	}

	// Reload to get relations
	product, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapper.ToProductResponse(product), nil
}

func (u *productUsecase) GetRecipe(ctx context.Context, id string) ([]dto.RecipeItemResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.ProductKind != models.ProductKindRecipe {
		return nil, errors.New("only RECIPE kind products have recipe items")
	}

	resp := mapper.ToProductResponse(product)
	return resp.RecipeItems, nil
}

func (u *productUsecase) UpdateRecipe(ctx context.Context, id string, items []dto.RecipeItemRequest) ([]dto.RecipeItemResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.ProductKind != models.ProductKindRecipe {
		return nil, errors.New("only RECIPE kind products can have recipe items")
	}

	if len(items) == 0 {
		return nil, errors.New("RECIPE products must have at least one recipe item")
	}

	if err := u.replaceRecipeItems(id, items); err != nil {
		return nil, fmt.Errorf("failed to update recipe items: %w", err)
	}

	if err := u.createRecipeVersion(id, items, nil, models.RecipeVersionChangeManual, "updated recipe"); err != nil {
		return nil, fmt.Errorf("failed to save recipe version: %w", err)
	}

	// Reload to get full recipe with ingredient details
	product, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := mapper.ToProductResponse(product)
	return resp.RecipeItems, nil
}

func (u *productUsecase) ListRecipeVersions(ctx context.Context, id string) ([]dto.RecipeVersionResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.ProductKind != models.ProductKindRecipe {
		return nil, errors.New("only RECIPE kind products have recipe versions")
	}

	var versions []models.ProductRecipeVersion
	if err := u.db.WithContext(ctx).
		Where("product_id = ?", id).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		Preload("Items.IngredientProduct").
		Preload("Items.Uom").
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, err
	}

	result := make([]dto.RecipeVersionResponse, 0, len(versions))
	for _, version := range versions {
		result = append(result, mapRecipeVersionToDTO(version))
	}

	return result, nil
}

func (u *productUsecase) CloneRecipeFromVersion(ctx context.Context, id string, req dto.CloneRecipeRequest) ([]dto.RecipeItemResponse, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.ProductKind != models.ProductKindRecipe {
		return nil, errors.New("only RECIPE kind products can clone recipe")
	}

	if req.SourceVersionID == nil || *req.SourceVersionID == "" {
		return nil, errors.New("source_version_id is required")
	}

	var sourceVersion models.ProductRecipeVersion
	if err := u.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", *req.SourceVersionID, id).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		First(&sourceVersion).Error; err != nil {
		return nil, errors.New("recipe version not found")
	}

	if len(sourceVersion.Items) == 0 {
		return nil, errors.New("selected version has no recipe items")
	}

	items := make([]dto.RecipeItemRequest, 0, len(sourceVersion.Items))
	for _, item := range sourceVersion.Items {
		items = append(items, dto.RecipeItemRequest{
			IngredientProductID: item.IngredientProductID,
			Quantity:            item.Quantity,
			UomID:               item.UomID,
			Notes:               item.Notes,
			SortOrder:           item.SortOrder,
		})
	}

	if err := u.replaceRecipeItems(id, items); err != nil {
		return nil, fmt.Errorf("failed to clone recipe items: %w", err)
	}

	note := req.Notes
	if strings.TrimSpace(note) == "" {
		note = fmt.Sprintf("cloned from version %d", sourceVersion.VersionNumber)
	}

	if err := u.createRecipeVersion(id, items, req.SourceVersionID, models.RecipeVersionChangeClone, note); err != nil {
		return nil, fmt.Errorf("failed to save cloned recipe version: %w", err)
	}

	product, err = u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := mapper.ToProductResponse(product)
	return resp.RecipeItems, nil
}

func (u *productUsecase) CompareRecipeVersions(ctx context.Context, id string, fromVersionID string, toVersionID string) (dto.RecipeVersionCompareResponse, error) {
	if strings.TrimSpace(fromVersionID) == "" || strings.TrimSpace(toVersionID) == "" {
		return dto.RecipeVersionCompareResponse{}, errors.New("from_version_id and to_version_id are required")
	}

	var fromVersion models.ProductRecipeVersion
	if err := u.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", fromVersionID, id).
		Preload("Items.IngredientProduct").
		Preload("Items.Uom").
		First(&fromVersion).Error; err != nil {
		return dto.RecipeVersionCompareResponse{}, errors.New("from recipe version not found")
	}

	var toVersion models.ProductRecipeVersion
	if err := u.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", toVersionID, id).
		Preload("Items.IngredientProduct").
		Preload("Items.Uom").
		First(&toVersion).Error; err != nil {
		return dto.RecipeVersionCompareResponse{}, errors.New("to recipe version not found")
	}

	fromMap := make(map[string]models.ProductRecipeVersionItem, len(fromVersion.Items))
	toMap := make(map[string]models.ProductRecipeVersionItem, len(toVersion.Items))

	for _, item := range fromVersion.Items {
		fromMap[item.IngredientProductID] = item
	}
	for _, item := range toVersion.Items {
		toMap[item.IngredientProductID] = item
	}

	diffs := make([]dto.RecipeVersionCompareDiff, 0)
	summary := map[string]int{"added": 0, "removed": 0, "changed": 0, "unchanged": 0}

	seen := make(map[string]struct{}, len(fromMap)+len(toMap))
	for ingredientID := range fromMap {
		seen[ingredientID] = struct{}{}
	}
	for ingredientID := range toMap {
		seen[ingredientID] = struct{}{}
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, ingredientID := range keys {
		fromItem, hasFrom := fromMap[ingredientID]
		toItem, hasTo := toMap[ingredientID]

		diffType := ""
		switch {
		case !hasFrom && hasTo:
			diffType = "ADDED"
			summary["added"]++
		case hasFrom && !hasTo:
			diffType = "REMOVED"
			summary["removed"]++
		case hasFrom && hasTo && fromItem.Quantity != toItem.Quantity:
			diffType = "CHANGED"
			summary["changed"]++
		default:
			diffType = "UNCHANGED"
			summary["unchanged"]++
		}

		ingredient := toItem.IngredientProduct
		if ingredient == nil {
			ingredient = fromItem.IngredientProduct
		}

		uom := toItem.Uom
		if uom == nil {
			uom = fromItem.Uom
		}

		entry := dto.RecipeVersionCompareDiff{
			IngredientProductID: ingredientID,
			FromQuantity:        fromItem.Quantity,
			ToQuantity:          toItem.Quantity,
			DeltaQuantity:       toItem.Quantity - fromItem.Quantity,
			Type:                diffType,
		}

		if ingredient != nil {
			entry.Ingredient = &dto.RecipeIngredientBasic{
				ID:           ingredient.ID,
				Code:         ingredient.Code,
				Name:         ingredient.Name,
				CostPrice:    ingredient.CostPrice,
				CurrentStock: ingredient.CurrentStock,
			}
		}

		if uom != nil {
			entry.Uom = &dto.UnitOfMeasureBasic{
				ID:     uom.ID,
				Name:   uom.Name,
				Symbol: uom.Symbol,
			}
		}

		diffs = append(diffs, entry)
	}

	return dto.RecipeVersionCompareResponse{
		FromVersionID: fromVersionID,
		ToVersionID:   toVersionID,
		FromVersion:   fromVersion.VersionNumber,
		ToVersion:     toVersion.VersionNumber,
		Summary:       summary,
		Diffs:         diffs,
	}, nil
}

func mapRecipeVersionToDTO(version models.ProductRecipeVersion) dto.RecipeVersionResponse {
	items := make([]dto.RecipeVersionItemResponse, 0, len(version.Items))
	for _, item := range version.Items {
		mappedItem := dto.RecipeVersionItemResponse{
			IngredientProductID: item.IngredientProductID,
			Quantity:            item.Quantity,
			UomID:               item.UomID,
			Notes:               item.Notes,
			SortOrder:           item.SortOrder,
		}

		if item.IngredientProduct != nil {
			mappedItem.Ingredient = &dto.RecipeIngredientBasic{
				ID:           item.IngredientProduct.ID,
				Code:         item.IngredientProduct.Code,
				Name:         item.IngredientProduct.Name,
				CostPrice:    item.IngredientProduct.CostPrice,
				CurrentStock: item.IngredientProduct.CurrentStock,
			}
		}

		if item.Uom != nil {
			mappedItem.Uom = &dto.UnitOfMeasureBasic{
				ID:     item.Uom.ID,
				Name:   item.Uom.Name,
				Symbol: item.Uom.Symbol,
			}
		}

		items = append(items, mappedItem)
	}

	return dto.RecipeVersionResponse{
		ID:              version.ID,
		VersionNumber:   version.VersionNumber,
		ChangeType:      version.ChangeType,
		Notes:           version.Notes,
		SourceVersionID: version.SourceVersionID,
		CreatedBy:       version.CreatedBy,
		CreatedAt:       version.CreatedAt,
		Items:           items,
	}
}

func (u *productUsecase) createRecipeVersion(productID string, items []dto.RecipeItemRequest, sourceVersionID *string, changeType string, notes string) error {
	return database.RetryTx(u.db, func(tx *gorm.DB) error {
		var maxVersion int
		if err := tx.Model(&models.ProductRecipeVersion{}).
			Where("product_id = ?", productID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVersion).Error; err != nil {
			return err
		}

		version := models.ProductRecipeVersion{
			ID:              uuid.New().String(),
			ProductID:       productID,
			VersionNumber:   maxVersion + 1,
			ChangeType:      changeType,
			Notes:           notes,
			SourceVersionID: sourceVersionID,
		}

		if err := tx.Create(&version).Error; err != nil {
			return err
		}

		for _, item := range items {
			versionItem := models.ProductRecipeVersionItem{
				ID:                  uuid.New().String(),
				RecipeVersionID:     version.ID,
				IngredientProductID: item.IngredientProductID,
				Quantity:            item.Quantity,
				UomID:               item.UomID,
				Notes:               item.Notes,
				SortOrder:           item.SortOrder,
			}

			if err := tx.Create(&versionItem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// saveRecipeItems creates recipe items for a product
func (u *productUsecase) saveRecipeItems(productID string, items []dto.RecipeItemRequest) error {
	return database.RetryTx(u.db, func(tx *gorm.DB) error {
		for _, item := range items {
			ri := models.ProductRecipeItem{
				ID:                  uuid.New().String(),
				ProductID:           productID,
				IngredientProductID: item.IngredientProductID,
				Quantity:            item.Quantity,
				UomID:               item.UomID,
				Notes:               item.Notes,
				SortOrder:           item.SortOrder,
			}
			if err := tx.Create(&ri).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// replaceRecipeItems deletes existing recipe items and creates new ones
func (u *productUsecase) replaceRecipeItems(productID string, items []dto.RecipeItemRequest) error {
	return database.RetryTx(u.db, func(tx *gorm.DB) error {
		// Hard-delete existing recipe items (subordinate records, not soft-deleted)
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&models.ProductRecipeItem{}).Error; err != nil {
			return err
		}

		for _, item := range items {
			ri := models.ProductRecipeItem{
				ID:                  uuid.New().String(),
				ProductID:           productID,
				IngredientProductID: item.IngredientProductID,
				Quantity:            item.Quantity,
				UomID:               item.UomID,
				Notes:               item.Notes,
				SortOrder:           item.SortOrder,
			}
			if err := tx.Create(&ri).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// syncOutlets replaces outlet assignments for a product using a delete-then-insert pattern.
// Passing an empty outletIDs slice removes all outlet assignments.
func (u *productUsecase) syncOutlets(ctx context.Context, productID string, outletIDs []string) error {
	return database.RetryTx(u.db, func(tx *gorm.DB) error {
		// Soft-delete all existing outlet assignments for this product
		if err := tx.Where("product_id = ?", productID).Delete(&models.ProductOutlet{}).Error; err != nil {
			return err
		}
		for _, outletID := range outletIDs {
			po := models.ProductOutlet{
				ProductID: productID,
				OutletID:  outletID,
			}
			if err := tx.Create(&po).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
