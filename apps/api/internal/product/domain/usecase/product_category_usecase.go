package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
	"github.com/gilabs/gims/api/internal/product/domain/mapper"
	"github.com/google/uuid"
)

// ProductCategoryUsecase defines the interface for product category business logic
type ProductCategoryUsecase interface {
	Create(ctx context.Context, req dto.CreateProductCategoryRequest) (dto.ProductCategoryResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductCategoryResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ProductCategoryResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateProductCategoryRequest) (dto.ProductCategoryResponse, error)
	Delete(ctx context.Context, id string) error
	// Tree methods
	GetTree(ctx context.Context, params dto.CategoryTreeParams) ([]dto.CategoryTreeResponse, error)
	GetChildren(ctx context.Context, parentID string, params dto.CategoryTreeParams) ([]dto.CategoryTreeResponse, error)
}

type productCategoryUsecase struct {
	repo repositories.ProductCategoryRepository
}

// NewProductCategoryUsecase creates a new ProductCategoryUsecase
func NewProductCategoryUsecase(repo repositories.ProductCategoryRepository) ProductCategoryUsecase {
	return &productCategoryUsecase{repo: repo}
}

func (u *productCategoryUsecase) Create(ctx context.Context, req dto.CreateProductCategoryRequest) (dto.ProductCategoryResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	categoryType := req.CategoryType
	if categoryType == "" {
		categoryType = models.CategoryTypeGoods
	}
	category := &models.ProductCategory{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		CategoryType: categoryType,
		ParentID:     req.ParentID,
		IsActive:     isActive,
	}

	if err := u.repo.Create(ctx, category); err != nil {
		return dto.ProductCategoryResponse{}, err
	}

	return mapper.ToProductCategoryResponse(category), nil
}

func (u *productCategoryUsecase) GetByID(ctx context.Context, id string) (dto.ProductCategoryResponse, error) {
	category, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductCategoryResponse{}, err
	}
	return mapper.ToProductCategoryResponse(category), nil
}

func (u *productCategoryUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ProductCategoryResponse, int64, error) {
	categories, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToProductCategoryResponseList(categories), total, nil
}

func (u *productCategoryUsecase) Update(ctx context.Context, id string, req dto.UpdateProductCategoryRequest) (dto.ProductCategoryResponse, error) {
	category, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ProductCategoryResponse{}, err
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.CategoryType != "" {
		category.CategoryType = req.CategoryType
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, category); err != nil {
		return dto.ProductCategoryResponse{}, err
	}

	return mapper.ToProductCategoryResponse(category), nil
}

func (u *productCategoryUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("product category not found")
	}

	count, err := u.repo.CountProductsByCategory(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete category with associated products")
	}

	return u.repo.Delete(ctx, id)
}

// GetTree returns hierarchical category tree starting from root
func (u *productCategoryUsecase) GetTree(ctx context.Context, params dto.CategoryTreeParams) ([]dto.CategoryTreeResponse, error) {
	var categories []models.ProductCategory
	var err error

	if params.ParentID != nil {
		// Get children of specific parent
		categories, err = u.repo.GetChildrenByParentID(ctx, *params.ParentID, params.OnlyActive)
	} else {
		// Get root categories
		categories, err = u.repo.GetRootCategories(ctx, params.OnlyActive)
	}

	if err != nil {
		return nil, err
	}

	return u.buildTree(ctx, categories, params, 0)
}

// GetChildren returns direct children of a category
func (u *productCategoryUsecase) GetChildren(ctx context.Context, parentID string, params dto.CategoryTreeParams) ([]dto.CategoryTreeResponse, error) {
	categories, err := u.repo.GetChildrenByParentID(ctx, parentID, params.OnlyActive)
	if err != nil {
		return nil, err
	}

	return u.buildTree(ctx, categories, params, 1)
}

// buildTree recursively builds tree structure from categories
func (u *productCategoryUsecase) buildTree(ctx context.Context, categories []models.ProductCategory, params dto.CategoryTreeParams, level int) ([]dto.CategoryTreeResponse, error) {
	result := make([]dto.CategoryTreeResponse, 0, len(categories))

	for _, cat := range categories {
		categoryType := cat.CategoryType
		if categoryType == "" {
			categoryType = models.CategoryTypeGoods
		}
		node := dto.CategoryTreeResponse{
			ID:           cat.ID,
			Name:         cat.Name,
			Description:  cat.Description,
			CategoryType: categoryType,
			ParentID:     cat.ParentID,
			IsActive:     cat.IsActive,
			Level:        level,
			Children:     []dto.CategoryTreeResponse{}, // Always initialize as empty array for JSON
		}

		// Check if has children
		hasChildren, err := u.repo.HasChildren(ctx, cat.ID)
		if err != nil {
			return nil, err
		}
		node.HasChildren = hasChildren

		// Get product count if requested
		if params.IncludeCount {
			count, err := u.repo.CountProductsByCategory(ctx, cat.ID)
			if err != nil {
				return nil, err
			}
			node.ProductCount = count
		}

		// Recursively load children if depth allows
		// MaxDepth 0 = unlimited, otherwise stop at MaxDepth
		if hasChildren && (params.MaxDepth == 0 || level < params.MaxDepth-1) {
			children, err := u.repo.GetChildrenByParentID(ctx, cat.ID, params.OnlyActive)
			if err != nil {
				return nil, err
			}
			childNodes, err := u.buildTree(ctx, children, params, level+1)
			if err != nil {
				return nil, err
			}
			node.Children = childNodes
		}

		result = append(result, node)
	}

	return result, nil
}

