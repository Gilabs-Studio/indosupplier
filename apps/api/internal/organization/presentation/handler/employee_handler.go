package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/core/storage"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// EmployeeHandler handles HTTP requests for employee operations
type EmployeeHandler struct {
	employeeUC usecase.EmployeeUsecase
}

// NewEmployeeHandler creates a new EmployeeHandler instance
func NewEmployeeHandler(uc usecase.EmployeeUsecase) *EmployeeHandler {
	return &EmployeeHandler{employeeUC: uc}
}

// Create handles POST /employees
func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	resp, err := h.employeeUC.Create(c.Request.Context(), req, userID)
	if err != nil {
		switch err {
		case usecase.ErrEmployeeCodeExists:
			errors.ErrorResponse(c, "EMPLOYEE_CODE_EXISTS", map[string]interface{}{
				"message": err.Error(),
			}, nil)
		case usecase.ErrReplacementNotFound:
			errors.ErrorResponse(c, "REPLACEMENT_NOT_FOUND", map[string]interface{}{
				"message": err.Error(),
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	meta := &response.Meta{}
	if userID != "" {
		meta.CreatedBy = userID
	}
	response.SuccessResponseCreated(c, resp, meta)
}

// GetByID handles GET /employees/:id
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.employeeUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// List handles GET /employees
func (h *EmployeeHandler) List(c *gin.Context) {
	var params dto.EmployeeListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	employees, total, err := h.employeeUC.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Calculate pagination meta
	perPage := params.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	page := params.Page
	if page <= 0 {
		page = 1
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
	if params.DivisionID != "" {
		meta.Filters["division_id"] = params.DivisionID
	}
	if params.JobPositionID != "" {
		meta.Filters["job_position_id"] = params.JobPositionID
	}
	if params.AreaID != "" {
		meta.Filters["area_id"] = params.AreaID
	}

	response.SuccessResponse(c, employees, meta)
}

// Update handles PUT /employees/:id
func (h *EmployeeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.Update(c.Request.Context(), id, req)
	if err != nil {
		switch err {
		case usecase.ErrEmployeeNotFound:
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
		case usecase.ErrEmployeeCodeExists:
			errors.ErrorResponse(c, "EMPLOYEE_CODE_EXISTS", map[string]interface{}{
				"message": err.Error(),
			}, nil)
		case usecase.ErrReplacementNotFound:
			errors.ErrorResponse(c, "REPLACEMENT_NOT_FOUND", map[string]interface{}{
				"message": err.Error(),
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}
	response.SuccessResponse(c, resp, meta)
}

// Delete handles DELETE /employees/:id
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.employeeUC.Delete(c.Request.Context(), id); err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.DeletedBy = uid
		}
	}
	response.SuccessResponseDeleted(c, "employee", id, meta)
}

// Approval Handlers removed

func (h *EmployeeHandler) AssignAreas(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignEmployeeAreasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.AssignAreas(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// AssignSupervisorAreas handles POST /employees/:id/supervisor-areas
func (h *EmployeeHandler) AssignSupervisorAreas(c *gin.Context) {
	id := c.Param("id")

	var req dto.AssignEmployeeSupervisorAreasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.AssignSupervisorAreas(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// BulkUpdateAreas handles PUT /employees/:id/areas — replaces all area assignments atomically.
func (h *EmployeeHandler) BulkUpdateAreas(c *gin.Context) {
	id := c.Param("id")

	var req dto.BulkUpdateEmployeeAreasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.BulkUpdateAreas(c.Request.Context(), id, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// RemoveAreaAssignment handles DELETE /employees/:id/areas/:area_id
func (h *EmployeeHandler) RemoveAreaAssignment(c *gin.Context) {
	id := c.Param("id")
	areaID := c.Param("area_id")

	if err := h.employeeUC.RemoveAreaAssignment(c.Request.Context(), id, areaID); err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, nil, nil)
}

// GetFormData handles GET /employees/form-data — returns dropdown options for the employee form.
func (h *EmployeeHandler) GetFormData(c *gin.Context) {
	formData, err := h.employeeUC.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, formData, nil)
}

// GetEmployeeContracts handles GET /employees/:id/contracts
func (h *EmployeeHandler) GetEmployeeContracts(c *gin.Context) {
	id := c.Param("id")

	contracts, err := h.employeeUC.GetEmployeeContracts(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, contracts, nil)
}

// CreateEmployeeContract handles POST /employees/:id/contracts
func (h *EmployeeHandler) CreateEmployeeContract(c *gin.Context) {
	id := c.Param("id")

	var req dto.CreateEmployeeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	resp, err := h.employeeUC.CreateEmployeeContract(c.Request.Context(), id, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, resp, nil)
}

// GetActiveContract handles GET /employees/:id/contracts/active
func (h *EmployeeHandler) GetActiveContract(c *gin.Context) {
	id := c.Param("id")

	contract, err := h.employeeUC.GetActiveContract(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	if contract == nil {
		response.SuccessResponse(c, nil, nil)
		return
	}

	response.SuccessResponse(c, contract, nil)
}

// UpdateEmployeeContract handles PUT /employees/:id/contracts/:contract_id
func (h *EmployeeHandler) UpdateEmployeeContract(c *gin.Context) {
	id := c.Param("id")
	contractID := c.Param("contract_id")

	var req dto.UpdateEmployeeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.UpdateEmployeeContract(c.Request.Context(), id, contractID, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// DeleteEmployeeContract handles DELETE /employees/:id/contracts/:contract_id
func (h *EmployeeHandler) DeleteEmployeeContract(c *gin.Context) {
	id := c.Param("id")
	contractID := c.Param("contract_id")

	if err := h.employeeUC.DeleteEmployeeContract(c.Request.Context(), id, contractID); err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, nil, nil)
}

// TerminateEmployeeContract handles POST /employees/:id/contracts/:contract_id/terminate
func (h *EmployeeHandler) TerminateEmployeeContract(c *gin.Context) {
	id := c.Param("id")
	contractID := c.Param("contract_id")

	var req dto.TerminateEmployeeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	resp, err := h.employeeUC.TerminateEmployeeContract(c.Request.Context(), id, contractID, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// RenewEmployeeContract handles POST /employees/:id/contracts/:contract_id/renew
func (h *EmployeeHandler) RenewEmployeeContract(c *gin.Context) {
	id := c.Param("id")
	contractID := c.Param("contract_id")

	var req dto.RenewEmployeeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	resp, err := h.employeeUC.RenewEmployeeContract(c.Request.Context(), id, contractID, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// CorrectActiveEmployeeContract handles PATCH /employees/:id/contracts/active
func (h *EmployeeHandler) CorrectActiveEmployeeContract(c *gin.Context) {
	id := c.Param("id")

	var req dto.CorrectEmployeeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	resp, err := h.employeeUC.CorrectActiveEmployeeContract(c.Request.Context(), id, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// GetEmployeeEducationHistories handles GET /employees/:id/education-histories
func (h *EmployeeHandler) GetEmployeeEducationHistories(c *gin.Context) {
	id := c.Param("id")

	educations, err := h.employeeUC.GetEmployeeEducationHistories(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, educations, nil)
}

// CreateEmployeeEducationHistory handles POST /employees/:id/education-histories
func (h *EmployeeHandler) CreateEmployeeEducationHistory(c *gin.Context) {
	id := c.Param("id")

	var req dto.CreateEmployeeEducationHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	resp, err := h.employeeUC.CreateEmployeeEducationHistory(c.Request.Context(), id, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, resp, nil)
}

// UpdateEmployeeEducationHistory handles PUT /employees/:id/education-histories/:education_id
func (h *EmployeeHandler) UpdateEmployeeEducationHistory(c *gin.Context) {
	id := c.Param("id")
	educationID := c.Param("education_id")

	var req dto.UpdateEmployeeEducationHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.UpdateEmployeeEducationHistory(c.Request.Context(), id, educationID, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrEducationNotFound {
			errors.ErrorResponse(c, "EDUCATION_NOT_FOUND", map[string]interface{}{
				"education_id": educationID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// DeleteEmployeeEducationHistory handles DELETE /employees/:id/education-histories/:education_id
func (h *EmployeeHandler) DeleteEmployeeEducationHistory(c *gin.Context) {
	id := c.Param("id")
	educationID := c.Param("education_id")

	err := h.employeeUC.DeleteEmployeeEducationHistory(c.Request.Context(), id, educationID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrEducationNotFound {
			errors.ErrorResponse(c, "EDUCATION_NOT_FOUND", map[string]interface{}{
				"education_id": educationID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"message": "Education history deleted successfully"}, nil)
}

// GetEmployeeCertifications handles GET /employees/:id/certifications
func (h *EmployeeHandler) GetEmployeeCertifications(c *gin.Context) {
	id := c.Param("id")

	certs, err := h.employeeUC.GetEmployeeCertifications(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, certs, nil)
}

// CreateEmployeeCertification handles POST /employees/:id/certifications
func (h *EmployeeHandler) CreateEmployeeCertification(c *gin.Context) {
	id := c.Param("id")

	var req dto.CreateEmployeeCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	resp, err := h.employeeUC.CreateEmployeeCertification(c.Request.Context(), id, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidCertificationDates {
			errors.ErrorResponse(c, "INVALID_CERTIFICATION_DATES", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, resp, nil)
}

// UpdateEmployeeCertification handles PUT /employees/:id/certifications/:certification_id
func (h *EmployeeHandler) UpdateEmployeeCertification(c *gin.Context) {
	id := c.Param("id")
	certID := c.Param("certification_id")

	var req dto.UpdateEmployeeCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.UpdateEmployeeCertification(c.Request.Context(), id, certID, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCertificationNotFound {
			errors.ErrorResponse(c, "CERTIFICATION_NOT_FOUND", map[string]interface{}{
				"certification_id": certID,
			}, nil)
			return
		}
		if err == usecase.ErrInvalidCertificationDates {
			errors.ErrorResponse(c, "INVALID_CERTIFICATION_DATES", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// DeleteEmployeeCertification handles DELETE /employees/:id/certifications/:certification_id
func (h *EmployeeHandler) DeleteEmployeeCertification(c *gin.Context) {
	id := c.Param("id")
	certID := c.Param("certification_id")

	err := h.employeeUC.DeleteEmployeeCertification(c.Request.Context(), id, certID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCertificationNotFound {
			errors.ErrorResponse(c, "CERTIFICATION_NOT_FOUND", map[string]interface{}{
				"certification_id": certID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"message": "Certification deleted successfully"}, nil)
}

// GetEmployeeAssets handles GET /employees/:id/assets
func (h *EmployeeHandler) GetEmployeeAssets(c *gin.Context) {
	id := c.Param("id")

	assets, err := h.employeeUC.GetEmployeeAssets(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, assets, nil)
}

// CreateEmployeeAsset handles POST /employees/:id/assets
func (h *EmployeeHandler) CreateEmployeeAsset(c *gin.Context) {
	id := c.Param("id")

	var req dto.CreateEmployeeAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	resp, err := h.employeeUC.CreateEmployeeAsset(c.Request.Context(), id, req, userID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrDuplicateAssetCode {
			errors.ErrorResponse(c, "DUPLICATE_ASSET_CODE", map[string]interface{}{
				"asset_code": req.AssetCode,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, resp, nil)
}

// UpdateEmployeeAsset handles PUT /employees/:id/assets/:asset_id
func (h *EmployeeHandler) UpdateEmployeeAsset(c *gin.Context) {
	id := c.Param("id")
	assetID := c.Param("asset_id")

	var req dto.UpdateEmployeeAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.UpdateEmployeeAsset(c.Request.Context(), id, assetID, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrAssetNotFound {
			errors.ErrorResponse(c, "ASSET_NOT_FOUND", map[string]interface{}{
				"asset_id": assetID,
			}, nil)
			return
		}
		if err == usecase.ErrAssetAlreadyReturned {
			errors.ErrorResponse(c, "ASSET_ALREADY_RETURNED", nil, nil)
			return
		}
		if err == usecase.ErrDuplicateAssetCode {
			errors.ErrorResponse(c, "DUPLICATE_ASSET_CODE", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// ReturnEmployeeAsset handles POST /employees/:id/assets/:asset_id/return
func (h *EmployeeHandler) ReturnEmployeeAsset(c *gin.Context) {
	id := c.Param("id")
	assetID := c.Param("asset_id")

	var req dto.ReturnEmployeeAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	resp, err := h.employeeUC.ReturnEmployeeAsset(c.Request.Context(), id, assetID, req)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrAssetNotFound {
			errors.ErrorResponse(c, "ASSET_NOT_FOUND", map[string]interface{}{
				"asset_id": assetID,
			}, nil)
			return
		}
		if err == usecase.ErrAssetAlreadyReturned {
			errors.ErrorResponse(c, "ASSET_ALREADY_RETURNED", nil, nil)
			return
		}
		if err == usecase.ErrInvalidReturnDate {
			errors.ErrorResponse(c, "INVALID_RETURN_DATE", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// DeleteEmployeeAsset handles DELETE /employees/:id/assets/:asset_id
func (h *EmployeeHandler) DeleteEmployeeAsset(c *gin.Context) {
	id := c.Param("id")
	assetID := c.Param("asset_id")

	err := h.employeeUC.DeleteEmployeeAsset(c.Request.Context(), id, assetID)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrAssetNotFound {
			errors.ErrorResponse(c, "ASSET_NOT_FOUND", map[string]interface{}{
				"asset_id": assetID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"message": "Asset deleted successfully"}, nil)
}

// GetEmployeeSignature handles GET /employees/:id/signature
func (h *EmployeeHandler) GetEmployeeSignature(c *gin.Context) {
	id := c.Param("id")

	signature, err := h.employeeUC.GetEmployeeSignature(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Return success even if signature is nil (employee doesn't have signature yet)
	// Frontend will handle the null case
	response.SuccessResponse(c, signature, nil)
}

// UploadEmployeeSignature handles POST /employees/:id/signature
func (h *EmployeeHandler) UploadEmployeeSignature(c *gin.Context) {
	id := c.Param("id")

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		errors.ErrorResponse(c, "FILE_REQUIRED", map[string]interface{}{
			"message": "Signature file is required",
		}, nil)
		return
	}
	defer file.Close()

	// Get user ID from context
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	// Get employee to retrieve name for filename
	employee, err := h.employeeUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, "Failed to retrieve employee details")
		return
	}

	// Save file to storage using the existing utility
	uploadConfig := utils.FileUploadConfig{
		MaxSize:      5 * 1024 * 1024, // 5MB
		OriginalName: employee.Name + " signature",
	}

	uploadedFile, err := utils.SaveSignatureFile(file, header, uploadConfig)
	if err != nil {
		switch err {
		case utils.ErrInvalidFileType:
			errors.ErrorResponse(c, "INVALID_FILE_TYPE", map[string]interface{}{
				"message": "Only PNG and JPG files are allowed",
			}, nil)
		case utils.ErrFileTooLarge:
			errors.ErrorResponse(c, "FILE_TOO_LARGE", map[string]interface{}{
				"message": "File size must be less than 5MB",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, fmt.Sprintf("Failed to save file: %v", err))
		}
		return
	}

	// Calculate file hash for integrity
	hash := sha256.New()
	file.Seek(0, 0) // Reset file pointer to beginning
	if _, err := io.Copy(hash, file); err != nil {
		// Non-critical error, use placeholder
		// Continue with upload
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Get image dimensions
	file.Seek(0, 0) // Reset file pointer
	img, _, err := image.Decode(file)
	width, height := 0, 0
	if err == nil {
		bounds := img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	signature, err := h.employeeUC.UploadEmployeeSignature(c.Request.Context(), id, uploadedFile.URL, uploadedFile.Filename, uploadedFile.Size, fileHash, uploadedFile.MimeType, width, height, userID)
			// Best-effort cleanup: remove the uploaded file if DB update fails
			_ = storage.Delete(c.Request.Context(), uploadedFile.Path)

	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, signature, nil)
}

// DeleteEmployeeSignature handles DELETE /employees/:id/signature
func (h *EmployeeHandler) DeleteEmployeeSignature(c *gin.Context) {
	id := c.Param("id")

	err := h.employeeUC.DeleteEmployeeSignature(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrEmployeeNotFound {
			errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
				"employee_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"message": "Signature deleted successfully"}, nil)
}
