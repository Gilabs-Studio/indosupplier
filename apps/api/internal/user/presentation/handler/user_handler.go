package handler

import (
	stderrors "errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/core/storage"
	"github.com/gilabs/gims/api/internal/core/utils"
	domainDTO "github.com/gilabs/gims/api/internal/user/domain/dto"
	"github.com/gilabs/gims/api/internal/user/domain/usecase"
	presentationDTO "github.com/gilabs/gims/api/internal/user/presentation/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userUC usecase.UserUsecase
}

func NewUserHandler(userUC usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUC: userUC,
	}
}

// List handles list users request
func (h *UserHandler) List(c *gin.Context) {
	var req domainDTO.ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	users, pagination, err := h.userUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Map to Presentation DTO
	userDTOs := make([]presentationDTO.UserResponseDTO, len(users))
	for i, u := range users {
		userDTOs[i] = presentationDTO.UserResponseDTO{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			AvatarURL: u.AvatarURL,
			RoleID:    u.RoleID,
			Role: presentationDTO.RoleDTO{
				Code: u.Role.Code,
				Name: u.Role.Name,
			},
			Status:               u.Status,
			PasswordResetPending: u.PasswordResetPending,
			CreatedAt:            u.CreatedAt,
		}
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
	if req.RoleID != "" {
		meta.Filters["role_id"] = req.RoleID
	}

	response.SuccessResponse(c, userDTOs, meta)
}

// GetAvailable returns active users not yet linked to an employee.
// Query params: search (optional prefix), exclude_employee_id (optional, keeps current employee's user visible during edit).
func (h *UserHandler) GetAvailable(c *gin.Context) {
	search := c.Query("search")
	excludeEmployeeID := c.Query("exclude_employee_id")

	available, err := h.userUC.GetAvailable(c.Request.Context(), search, excludeEmployeeID)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, available, nil)
}

// GetByID handles get user by ID request
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	resp := presentationDTO.UserResponseDTO{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		RoleID:    user.RoleID,
		Role: presentationDTO.RoleDTO{
			Code: user.Role.Code,
			Name: user.Role.Name,
		},
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}

	response.SuccessResponse(c, resp, nil)
}

// Create handles create user request
func (h *UserHandler) Create(c *gin.Context) {
	var req domainDTO.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdUser, err := h.userUC.Create(c.Request.Context(), &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrUserAlreadyExists) {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "user",
				"field":    "email",
				"value":    req.Email,
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrRoleNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  req.RoleID,
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrUserLimitReached) {
			errors.ErrorResponse(c, "USER_LIMIT_REACHED", map[string]interface{}{
				"reason": "user limit for this tenant has been reached; upgrade your plan to add more users",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	resp := presentationDTO.UserResponseDTO{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		AvatarURL: createdUser.AvatarURL,
		RoleID:    createdUser.RoleID,
		Role: presentationDTO.RoleDTO{
			Code: createdUser.Role.Code,
			Name: createdUser.Role.Name,
		},
		Status:    createdUser.Status,
		CreatedAt: createdUser.CreatedAt,
	}

	response.SuccessResponseCreated(c, resp, meta)
}

// Update handles update user request
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req domainDTO.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	updatedUser, err := h.userUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrUserNotFound) {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrOwnerMutationForbidden) {
			errors.ForbiddenResponse(c, "owner user cannot be modified", nil)
			return
		}
		if stderrors.Is(err, usecase.ErrUserAlreadyExists) {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "user",
				"field":    "email",
			}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrRoleNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.UpdatedBy = id
		}
	}

	resp := presentationDTO.UserResponseDTO{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
		Email:     updatedUser.Email,
		AvatarURL: updatedUser.AvatarURL,
		RoleID:    updatedUser.RoleID,
		Role: presentationDTO.RoleDTO{
			Code: updatedUser.Role.Code,
			Name: updatedUser.Role.Name,
		},
		Status:    updatedUser.Status,
		CreatedAt: updatedUser.CreatedAt,
	}

	response.SuccessResponse(c, resp, meta)
}

// Delete handles delete user request
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.userUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrOwnerMutationForbidden {
			errors.ForbiddenResponse(c, "owner user cannot be modified", nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Get user ID for meta
	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(string); ok {
			meta.DeletedBy = id
		}
	}

	response.SuccessResponseDeleted(c, "user", id, meta)
}

// UpdateProfile handles update profile request (self update)
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return
	}
	id, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user id in context")
		return
	}

	var req domainDTO.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	updatedUser, err := h.userUC.UpdateProfile(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrUserAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "user",
				"field":    "email",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		UpdatedBy: id,
	}

	// Map to Presentation DTO
	resp := presentationDTO.UserResponseDTO{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
		Email:     updatedUser.Email,
		AvatarURL: updatedUser.AvatarURL,
		Role: presentationDTO.RoleDTO{
			Code: updatedUser.Role.Code,
			Name: updatedUser.Role.Name,
		},
		Status:    updatedUser.Status,
		CreatedAt: updatedUser.CreatedAt,
	}

	response.SuccessResponse(c, resp, meta)
}

// ChangePassword handles change password request
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return
	}
	id, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user id in context")
		return
	}

	var req domainDTO.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.userUC.ChangePassword(c.Request.Context(), id, &req); err != nil {
		if err.Error() == "invalid old password" {
			errors.ErrorResponse(c, "INVALID_PASSWORD", map[string]interface{}{
				"field": "old_password",
			}, nil)
			return
		}
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		UpdatedBy: id,
	}

	response.SuccessResponse(c, map[string]string{"message": "password updated successfully"}, meta)
}

// UploadAvatar handles avatar upload request
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return
	}
	id, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user id in context")
		return
	}

	// 1. Get file from request
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}
	defer file.Close()

	// 2. Prepare upload config
	uploadConfig := utils.FileUploadConfig{
		MaxSize: 5 * 1024 * 1024,               // 5MB limit
		Folder:  fmt.Sprintf("avatars/%s", id), // Store each user's avatar under: avatars/{userId}/
	}

	// 3. Save file
	uploadedFile, err := utils.SaveUploadedFile(file, header, uploadConfig)
	if err != nil {
		if err == utils.ErrFileTooLarge {
			errors.ErrorResponse(c, "FILE_TOO_LARGE", map[string]interface{}{
				"max_size": "5MB",
			}, nil)
			return
		}
		if err == utils.ErrInvalidFileType || err == utils.ErrInvalidImage {
			errors.ErrorResponse(c, "INVALID_FILE_TYPE", map[string]interface{}{
				"allowed_types": "jpg, png, gif, webp",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// 4. Update user avatar URL
	if err := h.userUC.UpdateAvatar(c.Request.Context(), id, uploadedFile.URL); err != nil {
		// Best-effort cleanup: remove the uploaded file if DB update fails
		_ = storage.Delete(c.Request.Context(), uploadedFile.Path)

		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		UpdatedBy: id,
	}

	response.SuccessResponse(c, map[string]string{
		"avatar_url": uploadedFile.URL,
		"filename":   uploadedFile.Filename,
	}, meta)
}

// GetLimit returns the current user count and the maximum allowed for the active tenant.
func (h *UserHandler) GetLimit(c *gin.Context) {
	limit, err := h.userUC.GetLimit(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, limit, nil)
}

// RequestAccountDeletion schedules tenant deletion in 30 days (owner only).
func (h *UserHandler) RequestAccountDeletion(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return
	}
	id, ok := userIDVal.(string)
	if !ok || id == "" {
		errors.InternalServerErrorResponse(c, "invalid user id in context")
		return
	}

	schedule, err := h.userUC.RequestAccountDeletion(c.Request.Context(), id)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
		case usecase.ErrDeleteAccountForbidden:
			errors.ErrorResponse(c, "FORBIDDEN", map[string]interface{}{
				"reason": "only tenant owner can request account deletion",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, schedule, nil)
}
