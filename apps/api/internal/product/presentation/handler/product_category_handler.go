package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/product/domain/dto"
	"github.com/gilabs/gims/api/internal/product/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ProductCategoryHandler handles HTTP requests for product categories
type ProductCategoryHandler struct {
	uc usecase.ProductCategoryUsecase
}

// NewProductCategoryHandler creates a new ProductCategoryHandler
func NewProductCategoryHandler(uc usecase.ProductCategoryUsecase) *ProductCategoryHandler {
	return &ProductCategoryHandler{uc: uc}
}

// Create handles POST /product-categories
func (h *ProductCategoryHandler) Create(c *gin.Context) {
	var req dto.CreateProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET /product-categories/:id
func (h *ProductCategoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "PRODUCT_CATEGORY_NOT_FOUND", map[string]interface{}{
			"product_category_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /product-categories
func (h *ProductCategoryHandler) List(c *gin.Context) {
	params := repositories.ListParams{
		Search:     c.Query("search"),
		SortBy:     c.DefaultQuery("sort_by", "name"),
		SortDir:    c.DefaultQuery("sort_dir", "asc"),
		ActiveOnly: c.Query("active_only") == "true",
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT /product-categories/:id
func (h *ProductCategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		errors.ErrorResponse(c, "PRODUCT_CATEGORY_NOT_FOUND", map[string]interface{}{
			"product_category_id": id,
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE /product-categories/:id
func (h *ProductCategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "PRODUCT_CATEGORY_NOT_FOUND", map[string]interface{}{
			"product_category_id": id,
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "product_category", id, meta)
}

// GetTree handles GET /product-categories/tree
// Returns hierarchical category tree structure with optional product counts
// Query params:
//   - parent_id: string (optional) - start from specific parent, nil for root
//   - depth: int (optional) - max depth to load, 0 = unlimited (default: 0)
//   - include_count: bool (optional) - include product count per category (default: true)
//   - only_active: bool (optional) - only include active categories (default: false)
func (h *ProductCategoryHandler) GetTree(c *gin.Context) {
	params := dto.CategoryTreeParams{
		IncludeCount: true, // default true
		OnlyActive:   false,
		MaxDepth:     0, // unlimited
	}

	// Parse parent_id
	if parentID := c.Query("parent_id"); parentID != "" {
		params.ParentID = &parentID
	}

	// Parse depth
	if depthStr := c.Query("depth"); depthStr != "" {
		if depth, err := strconv.Atoi(depthStr); err == nil && depth >= 0 {
			params.MaxDepth = depth
		}
	}

	// Parse include_count
	if includeCount := c.Query("include_count"); includeCount == "false" {
		params.IncludeCount = false
	}

	// Parse only_active
	if onlyActive := c.Query("only_active"); onlyActive == "true" {
		params.OnlyActive = true
	}

	results, err := h.uc.GetTree(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, results, nil)
}

// GetChildren handles GET /product-categories/:id/children
// Returns direct children of a category (for lazy loading)
// Query params:
//   - include_count: bool (optional) - include product count per category (default: true)
//   - only_active: bool (optional) - only include active categories (default: false)
func (h *ProductCategoryHandler) GetChildren(c *gin.Context) {
	parentID := c.Param("id")
	if parentID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Parent ID is required",
		}, nil)
		return
	}

	params := dto.CategoryTreeParams{
		IncludeCount: true,
		OnlyActive:   false,
		MaxDepth:     1, // Only direct children
	}

	// Parse include_count
	if includeCount := c.Query("include_count"); includeCount == "false" {
		params.IncludeCount = false
	}

	// Parse only_active
	if onlyActive := c.Query("only_active"); onlyActive == "true" {
		params.OnlyActive = true
	}

	results, err := h.uc.GetChildren(c.Request.Context(), parentID, params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, results, nil)
}

