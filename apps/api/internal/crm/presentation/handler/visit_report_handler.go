package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// VisitReportHandler handles HTTP requests for visit reports
type VisitReportHandler struct {
	uc usecase.VisitReportUsecase
}

// NewVisitReportHandler creates a new visit report handler
func NewVisitReportHandler(uc usecase.VisitReportUsecase) *VisitReportHandler {
	return &VisitReportHandler{uc: uc}
}

// List handles GET request to list visit reports with filtering and pagination
func (h *VisitReportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	req := &dto.ListVisitReportsRequest{
		Page:              page,
		PerPage:           perPage,
		Search:            c.Query("search"),
		CustomerID:        c.Query("customer_id"),
		EmployeeID:        c.Query("employee_id"),
		ContactID:         c.Query("contact_id"),
		DealID:            c.Query("deal_id"),
		LeadID:            c.Query("lead_id"),
		TravelPlanID:      c.Query("travel_plan_id"),
		WithoutTravelPlan: c.Query("without_travel_plan") == "true",
		Outcome:           c.Query("outcome"),
		DateFrom:          c.Query("date_from"),
		DateTo:            c.Query("date_to"),
		SortBy:            c.DefaultQuery("sort_by", "created_at"),
		SortDir:           c.DefaultQuery("sort_dir", "desc"),
	}

	results, pagination, err := h.uc.List(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if pagination != nil {
		meta.Pagination = &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      int(pagination.Total),
			TotalPages: int(pagination.TotalPages),
			HasNext:    pagination.Page < int(pagination.TotalPages),
			HasPrev:    pagination.Page > 1,
		}
	}

	response.SuccessResponse(c, results, meta)
}

// GetByID handles GET request to get a visit report by ID
func (h *VisitReportHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Create handles POST request to create a visit report
func (h *VisitReportHandler) Create(c *gin.Context) {
	var req dto.CreateVisitReportRequest
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

	result, err := h.uc.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != nil {
		meta.CreatedBy = *createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

// Update handles PUT request to update a visit report
func (h *VisitReportHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateVisitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleVisitReportError(c, err)
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

// Delete handles DELETE request to delete a visit report
func (h *VisitReportHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleVisitReportError(c, err)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_visit_report", id, meta)
}

// CheckIn handles POST request to check in to a visit
func (h *VisitReportHandler) CheckIn(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.CheckInVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	result, err := h.uc.CheckIn(c.Request.Context(), id, &req, userID)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// CheckOut handles POST request to check out from a visit
func (h *VisitReportHandler) CheckOut(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.CheckOutVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	result, err := h.uc.CheckOut(c.Request.Context(), id, &req, userID)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Submit handles POST request to submit a visit report for approval
func (h *VisitReportHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.SubmitVisitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = dto.SubmitVisitReportRequest{}
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	result, err := h.uc.Submit(c.Request.Context(), id, &req, userID)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Approve handles POST request to approve a visit report
func (h *VisitReportHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.ApproveVisitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = dto.ApproveVisitReportRequest{}
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	result, err := h.uc.Approve(c.Request.Context(), id, &req, userID)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Reject handles POST request to reject a visit report
func (h *VisitReportHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.RejectVisitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = &id
		}
	}

	result, err := h.uc.Reject(c.Request.Context(), id, &req, userID)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// UploadPhotos handles POST request to upload photos for a visit report
func (h *VisitReportHandler) UploadPhotos(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req struct {
		PhotoURLs []string `json:"photo_urls" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.UploadPhotos(c.Request.Context(), id, req.PhotoURLs)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetFormData handles GET request to get form data for visit report forms
func (h *VisitReportHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// ListProgressHistory handles GET request to list progress history for a visit report
func (h *VisitReportHandler) ListProgressHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	results, pagination, err := h.uc.ListProgressHistory(c.Request.Context(), id, page, perPage)
	if err != nil {
		handleVisitReportError(c, err)
		return
	}

	meta := &response.Meta{}
	if pagination != nil {
		meta.Pagination = &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      int(pagination.Total),
			TotalPages: int(pagination.TotalPages),
			HasNext:    pagination.Page < int(pagination.TotalPages),
			HasPrev:    pagination.Page > 1,
		}
	}

	response.SuccessResponse(c, results, meta)
}

// handleVisitReportError maps business errors to appropriate HTTP responses
func handleVisitReportError(c *gin.Context, err error) {
	switch err.Error() {
	case usecase.ErrVisitReportNotFound.Error():
		errors.ErrorResponse(c, "VISIT_NOT_FOUND", map[string]interface{}{
			"message": "Visit report not found",
		}, nil)
	case usecase.ErrVisitReportNotDraft.Error():
		errors.ErrorResponse(c, "VISIT_NOT_DRAFT", map[string]interface{}{
			"message": "Visit report can only be modified in draft or rejected status",
		}, nil)
	case usecase.ErrVisitReportAlreadyCheckedIn.Error():
		errors.ErrorResponse(c, "VISIT_ALREADY_CHECKED_IN", map[string]interface{}{
			"message": "Visit report has already been checked in",
		}, nil)
	case usecase.ErrVisitReportNotCheckedIn.Error():
		errors.ErrorResponse(c, "VISIT_NOT_CHECKED_IN", map[string]interface{}{
			"message": "Visit report has not been checked in yet",
		}, nil)
	case usecase.ErrVisitReportCannotApproveOwn.Error():
		errors.ErrorResponse(c, "VISIT_CANNOT_APPROVE_OWN", map[string]interface{}{
			"message": "Cannot approve your own visit report",
		}, nil)
	case usecase.ErrVisitReportRejectionRequired.Error():
		errors.ErrorResponse(c, "VISIT_REJECTION_REASON_REQUIRED", map[string]interface{}{
			"message": "Rejection reason is required",
		}, nil)
	case usecase.ErrVisitReportMaxPhotosExceeded.Error():
		errors.ErrorResponse(c, "VISIT_MAX_PHOTOS_EXCEEDED", map[string]interface{}{
			"message": "Maximum 5 photos allowed per visit report",
		}, nil)
	case usecase.ErrVisitReportNotSubmitted.Error():
		errors.ErrorResponse(c, "VISIT_NOT_SUBMITTED", map[string]interface{}{
			"message": "Visit report must be submitted before approval or rejection",
		}, nil)
	case usecase.ErrVisitReportImmutable.Error():
		errors.ErrorResponse(c, "VISIT_APPROVED_IMMUTABLE", map[string]interface{}{
			"message": "Approved visit reports cannot be modified",
		}, nil)
	case "customer not found":
		errors.ErrorResponse(c, "CUSTOMER_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact not found":
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "assigned employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "deal not found":
		errors.ErrorResponse(c, "DEAL_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "lead not found":
		errors.ErrorResponse(c, "CRM_LEAD_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}

// ListByEmployee handles GET /visits/by-employee — returns per-employee visit report metrics.
// Intended for ALL/DIVISION/AREA scope views on the frontend.
func (h *VisitReportHandler) ListByEmployee(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	req := &dto.ListByEmployeeRequest{
		Page:    page,
		PerPage: perPage,
		Search:  c.Query("search"),
	}

	summaries, pagination, err := h.uc.ListByEmployee(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if pagination != nil {
		meta.Pagination = &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      int(pagination.Total),
			TotalPages: int(pagination.TotalPages),
			HasNext:    pagination.Page < int(pagination.TotalPages),
			HasPrev:    pagination.Page > 1,
		}
	}

	response.SuccessResponse(c, summaries, meta)
}
