package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/customer/data/repositories"
	"github.com/gilabs/gims/api/internal/customer/domain/dto"
	"github.com/gilabs/gims/api/internal/customer/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CustomerHandler handles HTTP requests for customers
type CustomerHandler struct {
	uc usecase.CustomerUsecase
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(uc usecase.CustomerUsecase) *CustomerHandler {
	return &CustomerHandler{uc: uc}
}

// getUserID extracts user ID from gin context
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// Create handles POST /customers
func (h *CustomerHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := getUserID(c)
	if userID == "" {
		errors.ErrorResponse(c, "UNAUTHORIZED", map[string]interface{}{
			"message": "User not authenticated",
		}, nil)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), userID, req)
	if err != nil {
		errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
			"resource": "customer",
			"field":    "code",
			"message":  err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{CreatedBy: userID}
	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET /customers/:id
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CUSTOMER_NOT_FOUND", map[string]interface{}{
			"customer_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /customers
func (h *CustomerHandler) List(c *gin.Context) {
	params := repositories.CustomerListParams{
		ListParams: repositories.ListParams{
			Search:  c.Query("search"),
			SortBy:  c.DefaultQuery("sort_by", "name"),
			SortDir: c.DefaultQuery("sort_dir", "asc"),
		},
		CustomerTypeID: c.Query("customer_type_id"),
	}

	if isLoyaltyStr := c.Query("is_loyalty_member"); isLoyaltyStr != "" {
		val := isLoyaltyStr == "true"
		params.IsLoyaltyMember = &val
	}

	// Parse pagination with max enforcement
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
	if params.CustomerTypeID != "" {
		meta.Filters["customer_type_id"] = params.CustomerTypeID
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT /customers/:id
func (h *CustomerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateCustomerRequest
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
		errors.ErrorResponse(c, "CUSTOMER_UPDATE_FAILED", map[string]interface{}{
			"customer_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{UpdatedBy: getUserID(c)}
	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE /customers/:id
func (h *CustomerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "CUSTOMER_DELETE_FAILED", map[string]interface{}{
			"customer_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{DeletedBy: getUserID(c)}
	response.SuccessResponseDeleted(c, "customer", id, meta)
}

// GetFormData handles GET /customers/form-data
func (h *CustomerHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// === Nested Bank Account Handlers ===

// AddBankAccount handles POST /customers/:id/bank-accounts
func (h *CustomerHandler) AddBankAccount(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Customer ID is required",
		}, nil)
		return
	}

	var req dto.CreateCustomerBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.AddBankAccount(c.Request.Context(), customerID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// UpdateBankAccount handles PUT /customers/:id/bank-accounts/:bankId
func (h *CustomerHandler) UpdateBankAccount(c *gin.Context) {
	bankID := c.Param("bankId")
	if bankID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Bank account ID is required",
		}, nil)
		return
	}

	var req dto.UpdateCustomerBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.UpdateBankAccount(c.Request.Context(), bankID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// DeleteBankAccount handles DELETE /customers/:id/bank-accounts/:bankId
func (h *CustomerHandler) DeleteBankAccount(c *gin.Context) {
	bankID := c.Param("bankId")
	if bankID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Bank account ID is required",
		}, nil)
		return
	}

	if err := h.uc.DeleteBankAccount(c.Request.Context(), bankID); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "bank_account", bankID, nil)
}
