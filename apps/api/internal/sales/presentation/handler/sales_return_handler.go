package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SalesReturnHandler struct {
	uc usecase.SalesReturnUsecase
}

const (
	salesReturnIDRequiredMessage = "ID is required"
	invalidIDFormatMessage       = "Invalid ID format"
)

func NewSalesReturnHandler(uc usecase.SalesReturnUsecase) *SalesReturnHandler {
	return &SalesReturnHandler{uc: uc}
}

func (h *SalesReturnHandler) GetFormData(c *gin.Context) {
	data, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, data, nil)
}

func (h *SalesReturnHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	params := repositories.SalesReturnListParams{
		Search:     c.Query("search"),
		Status:     c.Query("status"),
		Action:     c.Query("action"),
		InvoiceID:  c.Query("invoice_id"),
		DeliveryID: c.Query("delivery_id"),
		SortBy:     c.DefaultQuery("sort_by", "created_at"),
		SortDir:    c.DefaultQuery("sort_dir", "desc"),
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
		Sort: &response.SortMeta{
			Field: params.SortBy,
			Order: params.SortDir,
		},
	}
	if strings.TrimSpace(params.Search) != "" {
		meta.Filters["search"] = params.Search
	}
	if strings.TrimSpace(params.Status) != "" {
		meta.Filters["status"] = params.Status
	}
	if strings.TrimSpace(params.Action) != "" {
		meta.Filters["action"] = params.Action
	}
	if strings.TrimSpace(params.InvoiceID) != "" {
		meta.Filters["invoice_id"] = params.InvoiceID
	}
	if strings.TrimSpace(params.DeliveryID) != "" {
		meta.Filters["delivery_id"] = params.DeliveryID
	}

	response.SuccessResponse(c, items, meta)
}

func (h *SalesReturnHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": salesReturnIDRequiredMessage}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": invalidIDFormatMessage}, nil)
		return
	}

	item, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesReturnNotFound {
			errors.NotFoundResponse(c, "sales_return", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, item, nil)
}

func (h *SalesReturnHandler) Create(c *gin.Context) {
	var req dto.CreateSalesReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	item, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrSalesReturnInvalid {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, item, nil)
}

func (h *SalesReturnHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": salesReturnIDRequiredMessage}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": invalidIDFormatMessage}, nil)
		return
	}

	var req dto.UpdateSalesReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	item, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrSalesReturnNotFound {
			errors.NotFoundResponse(c, "sales_return", id)
			return
		}
		if err == usecase.ErrSalesReturnInvalid {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, item, nil)
}

func (h *SalesReturnHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": salesReturnIDRequiredMessage}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": invalidIDFormatMessage}, nil)
		return
	}

	var req dto.UpdateSalesReturnStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	item, err := h.uc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		if err == usecase.ErrSalesReturnNotFound {
			errors.NotFoundResponse(c, "sales_return", id)
			return
		}
		if err == usecase.ErrSalesReturnInvalid {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, item, nil)
}

func (h *SalesReturnHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": salesReturnIDRequiredMessage}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": invalidIDFormatMessage}, nil)
		return
	}

	err := h.uc.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesReturnNotFound {
			errors.NotFoundResponse(c, "sales_return", id)
			return
		}
		if err == usecase.ErrSalesReturnInvalid {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"id": id}, nil)
}

// AuditTrail handles GET /sales/returns/:id/audit-trail.
func (h *SalesReturnHandler) AuditTrail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": salesReturnIDRequiredMessage}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": invalidIDFormatMessage}, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	items, total, err := h.uc.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}
