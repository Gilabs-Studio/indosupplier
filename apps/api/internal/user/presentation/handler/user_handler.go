package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	domainDTO "github.com/gilabs/indosupplier/api/internal/user/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/user/domain/usecase"
	presentationDTO "github.com/gilabs/indosupplier/api/internal/user/presentation/dto"
)

type UserHandler struct {
	userUC usecase.UserUsecase
}

func NewUserHandler(userUC usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

func toPresentationUser(u domainDTO.UserResponse) presentationDTO.UserResponseDTO {
	resp := presentationDTO.UserResponseDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Status:    u.Status,
		Capabilities: presentationDTO.AccountCapabilitiesDTO{
			Buyer:    u.Capabilities.Buyer,
			Supplier: u.Capabilities.Supplier,
		},
		CreatedAt: u.CreatedAt,
	}
	if u.BuyerProfile != nil {
		resp.BuyerProfile = &presentationDTO.AccountProfileRefDTO{
			ID:     u.BuyerProfile.ID,
			Status: u.BuyerProfile.Status,
		}
	}
	if u.SupplierProfile != nil {
		resp.SupplierProfile = &presentationDTO.AccountProfileRefDTO{
			ID:     u.SupplierProfile.ID,
			Status: u.SupplierProfile.Status,
		}
	}
	return resp
}

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

	userDTOs := make([]presentationDTO.UserResponseDTO, len(users))
	for i, u := range users {
		userDTOs[i] = toPresentationUser(u)
	}

	meta := &response.Meta{Pagination: &response.PaginationMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      pagination.Total,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.Page < pagination.TotalPages,
		HasPrev:    pagination.Page > 1,
	}, Filters: map[string]interface{}{}}

	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}
	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}
	response.SuccessResponse(c, userDTOs, meta)
}

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

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrUserNotFound) {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, toPresentationUser(*user), nil)
}

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
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{"resource": "user", "field": "email", "value": req.Email}, nil)
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

	response.SuccessResponseCreated(c, toPresentationUser(*createdUser), meta)
}

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
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrUserAlreadyExists) {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{"resource": "user", "field": "email"}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, toPresentationUser(*updatedUser), meta)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.userUC.Delete(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrUserNotFound) {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "user", id, meta)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
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
		if stderrors.Is(err, usecase.ErrUserNotFound) {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrUserAlreadyExists) {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{"resource": "user", "field": "email"}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, toPresentationUser(*updatedUser), &response.Meta{UpdatedBy: id})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
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
		if stderrors.Is(err, usecase.ErrUserNotFound) {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{"user_id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"message": "Password changed successfully"}, &response.Meta{UpdatedBy: id})
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
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

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}
	defer file.Close()

	if err := utils.ValidateImageFile(file, header, 5*1024*1024); err != nil {
		errors.ErrorResponse(c, "INVALID_IMAGE", map[string]interface{}{"reason": err.Error()}, nil)
		return
	}

	uploaded, err := utils.SaveUploadedFile(file, header, utils.FileUploadConfig{MaxSize: 5 * 1024 * 1024, Folder: "avatars/" + id})
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	avatarURL := uploaded.URL

	if err := h.userUC.UpdateAvatar(c.Request.Context(), id, avatarURL); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]string{"avatar_url": avatarURL}, &response.Meta{UpdatedBy: id})
}

func (h *UserHandler) RequestAccountDeletion(c *gin.Context) {
	errors.ForbiddenResponse(c, "account deletion is not available in core mode", nil)
}

func (h *UserHandler) GetLimit(c *gin.Context) {
	limit, err := h.userUC.GetLimit(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"current": limit.Current, "max": limit.Max, "remaining": maxInt(limit.Max-limit.Current, 0)}, nil)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
