package handler

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/usecase"
)

type RFQHandler struct {
	usecase usecase.RFQUsecase
}

func NewRFQHandler(uc usecase.RFQUsecase) *RFQHandler {
	return &RFQHandler{usecase: uc}
}

func (h *RFQHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
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

func (h *RFQHandler) Create(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.CreateRFQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	res, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, res, &response.Meta{CreatedBy: userID})
}

func (h *RFQHandler) GetByID(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	res, err := h.usecase.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrRFQNotFound) {
			errors.ErrorResponse(c, "RFQ_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

type ListRFQRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=20"`
	Status  string `form:"status"`
}

func (h *RFQHandler) List(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var query ListRFQRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		errors.InvalidQueryParamResponse(c)
		return
	}

	page, perPage := utils.NormalizePagination(query.Page, query.PerPage, 10)

	items, total, err := h.usecase.List(c.Request.Context(), userID, query.Status, page, perPage)
	if err != nil {
		if stderrors.Is(err, usecase.ErrBuyerProfileNotFound) {
			errors.ErrorResponse(c, "BUYER_PROFILE_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := utils.TotalPages(total, perPage)

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Filters: map[string]interface{}{
			"status": query.Status,
		},
	}

	response.SuccessResponse(c, items, meta)
}

func (h *RFQHandler) GetBids(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	res, err := h.usecase.GetBids(c.Request.Context(), userID, id)
	if err != nil {
		if stderrors.Is(err, usecase.ErrRFQNotFound) {
			errors.ErrorResponse(c, "RFQ_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *RFQHandler) AcceptBid(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	bidID := c.Param("bidId")
	err := h.usecase.AcceptBid(c.Request.Context(), userID, id, bidID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrRFQNotFound) {
			errors.ErrorResponse(c, "RFQ_NOT_FOUND", map[string]interface{}{"id": id}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrBidNotFound) {
			errors.ErrorResponse(c, "BID_NOT_FOUND", map[string]interface{}{"bidId": bidID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
