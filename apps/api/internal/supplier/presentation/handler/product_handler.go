package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	domainDTO "github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/usecase"
)

type ProductHandler struct {
	productUC usecase.ProductUsecase
}

func NewProductHandler(productUC usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{productUC: productUC}
}

func (h *ProductHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return "", false
	}
	userID, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user ID structure in context")
		return "", false
	}
	return userID, true
}

func (h *ProductHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req domainDTO.ListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	products, pagination, err := h.productUC.List(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: &response.PaginationMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      pagination.Total,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.Page < pagination.TotalPages,
		HasPrev:    pagination.Page > 1,
	}, Filters: map[string]interface{}{}}

	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}
	if req.CategoryID != "" {
		meta.Filters["category_id"] = req.CategoryID
	}

	response.SuccessResponse(c, products, meta)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	product, err := h.productUC.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProductNotFound) {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrProductUnauthorized) {
			errors.ForbiddenResponse(c, "you do not own this product", nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, product, nil)
}

func (h *ProductHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req domainDTO.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	product, err := h.productUC.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, product, &response.Meta{CreatedBy: userID})
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var req domainDTO.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	product, err := h.productUC.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProductNotFound) {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrProductUnauthorized) {
			errors.ForbiddenResponse(c, "you do not own this product", nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, product, &response.Meta{UpdatedBy: userID})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	err := h.productUC.Delete(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProductNotFound) {
			errors.ErrorResponse(c, "PRODUCT_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrProductUnauthorized) {
			errors.ForbiddenResponse(c, "you do not own this product", nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "product", id, &response.Meta{DeletedBy: userID})
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	categories, err := h.productUC.ListCategories(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, categories, nil)
}
