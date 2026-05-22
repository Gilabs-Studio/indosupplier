package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// CompanyHandler handles company HTTP requests
type CompanyHandler struct {
	companyUC usecase.CompanyUsecase
}

// NewCompanyHandler creates a new CompanyHandler
func NewCompanyHandler(companyUC usecase.CompanyUsecase) *CompanyHandler {
	return &CompanyHandler{companyUC: companyUC}
}

func (h *CompanyHandler) List(c *gin.Context) {
	var req dto.ListCompaniesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	companies, pagination, err := h.companyUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
			HasNext:    pagination.Page < pagination.TotalPages,
			HasPrev:    pagination.Page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}
	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}

	response.SuccessResponse(c, companies, meta)
}

func (h *CompanyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	company, err := h.companyUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			errors.ErrorResponse(c, "COMPANY_NOT_FOUND", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, company, nil)
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var createdBy *string
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = &id
		}
	}

	company, err := h.companyUC.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if createdBy != nil {
		meta.CreatedBy = *createdBy
	}

	response.SuccessResponseCreated(c, company, meta)
}

func (h *CompanyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	company, err := h.companyUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			errors.ErrorResponse(c, "COMPANY_NOT_FOUND", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, company, nil)
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.companyUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			errors.ErrorResponse(c, "COMPANY_NOT_FOUND", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "company", id, nil)
}

// SubmitForApproval handles submit company for approval
func (h *CompanyHandler) SubmitForApproval(c *gin.Context) {
	id := c.Param("id")

	company, err := h.companyUC.SubmitForApproval(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			errors.ErrorResponse(c, "COMPANY_NOT_FOUND", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCompanyAlreadyApproved {
			errors.ErrorResponse(c, "COMPANY_ALREADY_APPROVED", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, company, nil)
}

// Approve handles approve/reject company
func (h *CompanyHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	var req dto.ApproveCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	approvedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			approvedBy = uid
		}
	}

	company, err := h.companyUC.Approve(c.Request.Context(), id, &req, approvedBy)
	if err != nil {
		if err == usecase.ErrCompanyNotFound {
			errors.ErrorResponse(c, "COMPANY_NOT_FOUND", map[string]interface{}{
				"company_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCompanyNotPending {
			errors.ErrorResponse(c, "COMPANY_NOT_PENDING", map[string]interface{}{
				"company_id": id,
				"reason":     "company is not pending approval",
			}, nil)
			return
		}
		if err == usecase.ErrInvalidApprovalAction {
			errors.ErrorResponse(c, "INVALID_APPROVAL_ACTION", map[string]interface{}{
				"valid_actions": []string{"approve", "reject"},
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, company, nil)
}
