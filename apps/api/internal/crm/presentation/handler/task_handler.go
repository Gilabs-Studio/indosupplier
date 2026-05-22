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

// TaskHandler handles HTTP requests for CRM tasks
type TaskHandler struct {
	uc usecase.TaskUsecase
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(uc usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{uc: uc}
}

// Create handles POST request to create a task
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = id
		}
	}

	result, err := h.uc.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != "" {
		meta.CreatedBy = createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET request to get a task by ID
func (h *TaskHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_TASK_NOT_FOUND", map[string]interface{}{
			"task_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list tasks with filtering and pagination
func (h *TaskHandler) List(c *gin.Context) {
	params := repositories.TaskListParams{
		Search:     c.Query("search"),
		SortBy:     c.DefaultQuery("sort_by", "created_at"),
		SortDir:    c.DefaultQuery("sort_dir", "desc"),
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		Type:       c.Query("type"),
		AssignedTo: c.Query("assigned_to"),
		CustomerID: c.Query("customer_id"),
		DealID:     c.Query("deal_id"),
		LeadID:     c.Query("lead_id"),
		DueDateFrom: c.Query("due_date_from"),
		DueDateTo:   c.Query("due_date_to"),
	}

	// Overdue filter
	if isOverdue := c.Query("is_overdue"); isOverdue != "" {
		v := isOverdue == "true"
		params.IsOverdue = &v
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

	if params.Status != "" {
		meta.Filters["status"] = params.Status
	}
	if params.Priority != "" {
		meta.Filters["priority"] = params.Priority
	}
	if params.AssignedTo != "" {
		meta.Filters["assigned_to"] = params.AssignedTo
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT request to update a task
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateTaskRequest
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
		handleTaskError(c, err)
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

// Delete handles DELETE request to delete a task
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleTaskError(c, err)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_task", id, meta)
}

// Assign handles POST request to assign a task
func (h *TaskHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	assignedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			assignedBy = uid
		}
	}

	result, err := h.uc.Assign(c.Request.Context(), id, req, assignedBy)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Complete handles POST request to mark a task as completed
func (h *TaskHandler) Complete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.Complete(c.Request.Context(), id)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// MarkInProgress handles POST request to mark a task as in_progress
func (h *TaskHandler) MarkInProgress(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.MarkInProgress(c.Request.Context(), id)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Cancel handles POST request to cancel a task
func (h *TaskHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.Cancel(c.Request.Context(), id)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetFormData handles GET request to get form data for task forms
func (h *TaskHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// --- Reminder nested handlers ---

// ListReminders handles GET request to list reminders for a task
func (h *TaskHandler) ListReminders(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Task ID is required",
		}, nil)
		return
	}

	results, err := h.uc.ListReminders(c.Request.Context(), taskID)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, results, nil)
}

// GetReminderByID handles GET request to get a specific reminder
func (h *TaskHandler) GetReminderByID(c *gin.Context) {
	taskID := c.Param("id")
	reminderID := c.Param("reminderID")
	if taskID == "" || reminderID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Task ID and Reminder ID are required",
		}, nil)
		return
	}

	result, err := h.uc.GetReminderByID(c.Request.Context(), taskID, reminderID)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// CreateReminder handles POST request to create a reminder for a task
func (h *TaskHandler) CreateReminder(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Task ID is required",
		}, nil)
		return
	}

	var req dto.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = id
		}
	}

	result, err := h.uc.CreateReminder(c.Request.Context(), taskID, req, createdBy)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// UpdateReminder handles PUT request to update a reminder
func (h *TaskHandler) UpdateReminder(c *gin.Context) {
	taskID := c.Param("id")
	reminderID := c.Param("reminderID")
	if taskID == "" || reminderID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Task ID and Reminder ID are required",
		}, nil)
		return
	}

	var req dto.UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.UpdateReminder(c.Request.Context(), taskID, reminderID, req)
	if err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// DeleteReminder handles DELETE request to delete a reminder
func (h *TaskHandler) DeleteReminder(c *gin.Context) {
	taskID := c.Param("id")
	reminderID := c.Param("reminderID")
	if taskID == "" || reminderID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Task ID and Reminder ID are required",
		}, nil)
		return
	}

	if err := h.uc.DeleteReminder(c.Request.Context(), taskID, reminderID); err != nil {
		handleTaskError(c, err)
		return
	}

	response.SuccessResponseDeleted(c, "crm_task_reminder", reminderID, nil)
}

// handleTaskError maps business errors to appropriate HTTP responses
func handleTaskError(c *gin.Context, err error) {
	switch err.Error() {
	case "task not found":
		errors.ErrorResponse(c, "CRM_TASK_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "customer not found":
		errors.ErrorResponse(c, "CUSTOMER_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact not found":
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "deal not found":
		errors.ErrorResponse(c, "DEAL_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot update cancelled task":
		errors.ErrorResponse(c, "CRM_TASK_CANCELLED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "task is already completed":
		errors.ErrorResponse(c, "CRM_TASK_ALREADY_COMPLETED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot complete a cancelled task":
		errors.ErrorResponse(c, "CRM_TASK_CANCELLED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot reopen a completed or cancelled task":
		errors.ErrorResponse(c, "CRM_TASK_STATE_INVALID", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "task is already cancelled":
		errors.ErrorResponse(c, "CRM_TASK_ALREADY_CANCELLED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot cancel a completed task":
		errors.ErrorResponse(c, "CRM_TASK_ALREADY_COMPLETED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "reminder not found":
		errors.ErrorResponse(c, "CRM_REMINDER_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "reminder does not belong to this task":
		errors.ErrorResponse(c, "CRM_REMINDER_MISMATCH", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "invalid due_date format, use ISO 8601":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "due_date",
		}, nil)
	case "invalid remind_at format, use ISO 8601":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "remind_at",
		}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
