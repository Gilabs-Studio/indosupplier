package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
	"github.com/gilabs/gims/api/internal/supplier/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SupplierHandler handles HTTP requests for suppliers
type SupplierHandler struct {
	uc usecase.SupplierUsecase
}

// NewSupplierHandler creates a new SupplierHandler
func NewSupplierHandler(uc usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{uc: uc}
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

// Create handles POST /suppliers
func (h *SupplierHandler) Create(c *gin.Context) {
	var req dto.CreateSupplierRequest
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
			"resource": "supplier",
			"field":    "code",
			"message":  err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{CreatedBy: userID}
	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET /suppliers/:id
func (h *SupplierHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "SUPPLIER_NOT_FOUND", map[string]interface{}{
			"supplier_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /suppliers
func (h *SupplierHandler) List(c *gin.Context) {
	params := repositories.SupplierListParams{
		ListParams: repositories.ListParams{
			Search:  c.Query("search"),
			SortBy:  c.DefaultQuery("sort_by", "name"),
			SortDir: c.DefaultQuery("sort_dir", "asc"),
		},
		SupplierTypeID: c.Query("supplier_type_id"),
		Status:         c.Query("status"),
	}

	// Parse is_approved filter
	if isApproved := c.Query("is_approved"); isApproved != "" {
		val := isApproved == "true"
		params.IsApproved = &val
	}

	// Parse pagination
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
	if params.SupplierTypeID != "" {
		meta.Filters["supplier_type_id"] = params.SupplierTypeID
	}
	if params.Status != "" {
		meta.Filters["status"] = params.Status
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT /suppliers/:id
func (h *SupplierHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateSupplierRequest
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
		errors.ErrorResponse(c, "SUPPLIER_UPDATE_FAILED", map[string]interface{}{
			"supplier_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{UpdatedBy: getUserID(c)}
	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE /suppliers/:id
func (h *SupplierHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "SUPPLIER_DELETE_FAILED", map[string]interface{}{
			"supplier_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{DeletedBy: getUserID(c)}
	response.SuccessResponseDeleted(c, "supplier", id, meta)
}

// Submit handles POST /suppliers/:id/submit
func (h *SupplierHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.Submit(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "SUPPLIER_SUBMIT_FAILED", map[string]interface{}{
			"supplier_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Approve handles POST /suppliers/:id/approve
func (h *SupplierHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	userID := getUserID(c)
	if userID == "" {
		errors.ErrorResponse(c, "UNAUTHORIZED", map[string]interface{}{
			"message": "User not authenticated",
		}, nil)
		return
	}

	var req dto.ApproveSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Approve(c.Request.Context(), id, userID, req)
	if err != nil {
		errors.ErrorResponse(c, "SUPPLIER_APPROVE_FAILED", map[string]interface{}{
			"supplier_id": id,
			"message":     err.Error(),
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// === Nested Contact Handlers ===

// AddContact handles POST /suppliers/:id/contacts
func (h *SupplierHandler) AddContact(c *gin.Context) {
	supplierID := c.Param("id")
	if supplierID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Supplier ID is required",
		}, nil)
		return
	}

	var req dto.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.AddContact(c.Request.Context(), supplierID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// UpdateContact handles PUT /suppliers/:id/contacts/:contactId
func (h *SupplierHandler) UpdateContact(c *gin.Context) {
	contactID := c.Param("contactId")
	if contactID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Contact ID is required",
		}, nil)
		return
	}

	var req dto.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.UpdateContact(c.Request.Context(), contactID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// DeleteContact handles DELETE /suppliers/:id/contacts/:contactId
func (h *SupplierHandler) DeleteContact(c *gin.Context) {
	contactID := c.Param("contactId")
	if contactID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Contact ID is required",
		}, nil)
		return
	}

	if err := h.uc.DeleteContact(c.Request.Context(), contactID); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "contact", contactID, nil)
}

// === Nested Bank Account Handlers ===

// AddBankAccount handles POST /suppliers/:id/bank-accounts
func (h *SupplierHandler) AddBankAccount(c *gin.Context) {
	supplierID := c.Param("id")
	if supplierID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Supplier ID is required",
		}, nil)
		return
	}

	var req dto.CreateSupplierBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.AddBankAccount(c.Request.Context(), supplierID, req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// UpdateBankAccount handles PUT /suppliers/:id/bank-accounts/:bankId
func (h *SupplierHandler) UpdateBankAccount(c *gin.Context) {
	bankID := c.Param("bankId")
	if bankID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Bank account ID is required",
		}, nil)
		return
	}

	var req dto.UpdateSupplierBankRequest
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

// DeleteBankAccount handles DELETE /suppliers/:id/bank-accounts/:bankId
func (h *SupplierHandler) DeleteBankAccount(c *gin.Context) {
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
