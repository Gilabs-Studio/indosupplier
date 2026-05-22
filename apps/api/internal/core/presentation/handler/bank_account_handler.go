package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BankAccountHandler struct {
	uc usecase.BankAccountUsecase
}

func NewBankAccountHandler(uc usecase.BankAccountUsecase) *BankAccountHandler {
	return &BankAccountHandler{uc: uc}
}

// List handles GET /finance/bank-accounts
func (h *BankAccountHandler) List(c *gin.Context) {
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

	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	isActive := c.Query("is_active")
	var activePtr *bool
	if isActive == "true" {
		v := true
		activePtr = &v
	} else if isActive == "false" {
		v := false
		activePtr = &v
	}

	params := repositories.BankAccountListParams{
		CompanyID:   c.Query("company_id"),
		Search:      search,
		IsActive:    activePtr,
		CurrencyID:  c.Query("currency_id"),
		AccountType: c.Query("account_type"),
		SortBy:      sortBy,
		SortDir:     sortDir,
		Limit:       perPage,
		Offset:      (page - 1) * perPage,
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
		Sort: &response.SortMeta{
			Field: params.SortBy,
			Order: params.SortDir,
		},
	}
	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	if params.CompanyID != "" {
		meta.Filters["company_id"] = params.CompanyID
	}
	if params.CurrencyID != "" {
		meta.Filters["currency_id"] = params.CurrencyID
	}
	if params.AccountType != "" {
		meta.Filters["account_type"] = params.AccountType
	}
	if activePtr != nil {
		meta.Filters["is_active"] = *activePtr
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// ListUnified handles GET /finance/bank-accounts/unified
func (h *BankAccountHandler) ListUnified(c *gin.Context) {
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

	params := repositories.BankAccountListParams{
		Search:     c.Query("search"),
		OwnerType:  c.Query("owner_type"),
		CurrencyID: c.Query("currency_id"),
		SortBy:     c.DefaultQuery("sort_by", "created_at"),
		SortDir:    c.DefaultQuery("sort_dir", "desc"),
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
	}

	items, total, err := h.uc.ListUnified(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
		Sort: &response.SortMeta{
			Field: params.SortBy,
			Order: params.SortDir,
		},
	}
	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	if params.OwnerType != "" {
		meta.Filters["owner_type"] = params.OwnerType
	}
	if params.CurrencyID != "" {
		meta.Filters["currency_id"] = params.CurrencyID
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// ListTransactionHistory handles GET /finance/bank-accounts/:id/transaction-history
func (h *BankAccountHandler) ListTransactionHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}

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

	items, total, err := h.uc.ListTransactionHistory(c.Request.Context(), id, perPage, (page-1)*perPage)
	if err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// GetByID handles GET /finance/bank-accounts/:id
func (h *BankAccountHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Create handles POST /finance/bank-accounts
func (h *BankAccountHandler) Create(c *gin.Context) {
	var req dto.CreateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

// Update handles PUT /finance/bank-accounts/:id
func (h *BankAccountHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	var req dto.UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, res, nil)
}

// Delete handles DELETE /finance/bank-accounts/:id
func (h *BankAccountHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, map[string]interface{}{"id": id}, nil)
}

// ========== PHASE 2 HANDLERS ==========

// GetByIDWithBalance handles GET /finance/bank-accounts/phase2/:id/detail
// Returns bank account with computed GL balance, breakdown, and metadata
func (h *BankAccountHandler) GetByIDWithBalance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}

	res, err := h.uc.GetByIDWithBalance(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

// ListByCompanyWithBalance handles GET /finance/bank-accounts/company/:company_id/with-balance
// Returns all bank accounts for a company with computed GL balances and metadata
func (h *BankAccountHandler) ListByCompanyWithBalance(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		errors.ErrorResponse(c, "INVALID_PARAMS", map[string]interface{}{"message": "company_id is required"}, nil)
		return
	}

	// Parse pagination and filters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	params := repositories.BankAccountListParams{
		CompanyID:   companyID,
		Search:      c.Query("search"),
		AccountType: c.Query("account_type"),
		CurrencyID:  c.Query("currency_id"),
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortDir:     c.DefaultQuery("sort_dir", "desc"),
		Limit:       perPage,
		Offset:      (page - 1) * perPage,
	}

	// Parse is_active filter (optional)
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		params.IsActive = &isActive
	}

	items, total, err := h.uc.ListByCompanyWithBalance(c.Request.Context(), companyID, params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{"company_id": companyID},
		Sort: &response.SortMeta{
			Field: params.SortBy,
			Order: params.SortDir,
		},
	}
	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	if params.AccountType != "" {
		meta.Filters["account_type"] = params.AccountType
	}
	if params.CurrencyID != "" {
		meta.Filters["currency_id"] = params.CurrencyID
	}
	if params.IsActive != nil {
		meta.Filters["is_active"] = *params.IsActive
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// ToggleStatus handles POST /finance/bank-accounts/:id/toggle-status
// Toggles the is_active flag of a bank account
func (h *BankAccountHandler) ToggleStatus(c *gin.Context) {
	bankAccountID := c.Param("id")
	if bankAccountID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}

	res, err := h.uc.ToggleStatus(c.Request.Context(), bankAccountID)
	if err != nil {
		if err == usecase.ErrBankAccountNotFound {
			errors.NotFoundResponse(c, "bank_account", bankAccountID)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}
