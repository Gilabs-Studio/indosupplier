package handler

import (
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/usecase"
)

type SupplierReviewHandler struct {
	reviewUC usecase.SupplierReviewUsecase
}

func NewSupplierReviewHandler(reviewUC usecase.SupplierReviewUsecase) *SupplierReviewHandler {
	return &SupplierReviewHandler{reviewUC: reviewUC}
}

func (h *SupplierReviewHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return "", false
	}
	userID, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user ID structure in context")
		return "", false
	}
	return userID, true
}

func (h *SupplierReviewHandler) GetReviews(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	search := c.Query("search")

	var ratingFilter *int
	if ratingStr := c.Query("rating"); ratingStr != "" {
		if r, err := strconv.Atoi(ratingStr); err == nil && r >= 1 && r <= 5 {
			ratingFilter = &r
		}
	}

	res, err := h.reviewUC.GetReviews(c.Request.Context(), userID, page, limit, ratingFilter, status, search)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierReviewProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *SupplierReviewHandler) ReplyReview(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_PARAMS", "message": "review ID is required"}})
		return
	}

	var req dto.SupplierReplyReviewRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.reviewUC.ReplyReview(c.Request.Context(), userID, reviewID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrSupplierReviewProfileNotFound) {
			errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierReviewNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"review_id": reviewID}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierReviewNotApproved) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "REVIEW_NOT_APPROVED",
					"message": "Only approved reviews can be replied to",
				},
			})
			return
		}
		if stderrors.Is(err, usecase.ErrSupplierReviewEmptyReply) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "Reply text must be between 3 and 1000 characters",
				},
			})
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, &response.Meta{UpdatedBy: userID})
}
