package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ContactHandler handles HTTP requests for contacts
type ContactHandler struct {
	uc usecase.ContactUsecase
}

// NewContactHandler creates a new contact handler
func NewContactHandler(uc usecase.ContactUsecase) *ContactHandler {
	return &ContactHandler{uc: uc}
}

// Create handles POST request to create a contact
func (h *ContactHandler) Create(c *gin.Context) {
	var req dto.CreateContactRequest
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
		handleContactError(c, err)
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

// GetByID handles GET request to get a contact by ID
func (h *ContactHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"contact_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list contacts
func (h *ContactHandler) List(c *gin.Context) {
	params := repositories.ContactListParams{
		ListParams: repositories.ListParams{
			Search:  c.Query("search"),
			SortBy:  c.DefaultQuery("sort_by", "name"),
			SortDir: c.DefaultQuery("sort_dir", "asc"),
		},
		CustomerID:    c.Query("customer_id"),
		ContactRoleID: c.Query("contact_role_id"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
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
	if params.CustomerID != "" {
		meta.Filters["customer_id"] = params.CustomerID
	}
	if params.ContactRoleID != "" {
		meta.Filters["contact_role_id"] = params.ContactRoleID
	}

	response.SuccessResponse(c, results, meta)
}

// ListByCustomer handles GET request to list contacts for a specific customer
func (h *ContactHandler) ListByCustomer(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Customer ID is required",
		}, nil)
		return
	}

	params := repositories.ListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", "name"),
		SortDir: c.DefaultQuery("sort_dir", "asc"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.ListByCustomerID(c.Request.Context(), customerID, params)
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

// GetFormData handles GET request to get form data for contacts
func (h *ContactHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// Update handles PUT request to update a contact
func (h *ContactHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
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

	result, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		handleContactError(c, err)
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

// Delete handles DELETE request to delete a contact
func (h *ContactHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"contact_id": id,
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_contact", id, meta)
}

// handleContactError maps business errors to appropriate HTTP responses
func handleContactError(c *gin.Context, err error) {
	switch err.Error() {
	case "customer not found":
		errors.ErrorResponse(c, "CUSTOMER_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact role not found":
		errors.ErrorResponse(c, "CRM_CONTACT_ROLE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact not found":
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact with this name already exists for this customer":
		errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
