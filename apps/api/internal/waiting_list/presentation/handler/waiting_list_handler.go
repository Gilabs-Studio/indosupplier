package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	coreErrors "github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/domain/usecase"
)

type WaitingListHandler struct {
	uc usecase.WaitingListUsecase
}

func NewWaitingListHandler(uc usecase.WaitingListUsecase) *WaitingListHandler {
	return &WaitingListHandler{uc: uc}
}

// Join handles public signup for the waiting list
func (h *WaitingListHandler) Join(c *gin.Context) {
	var req dto.JoinWaitingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	res, err := h.uc.Join(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
			coreErrors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"email": "email already registered in waiting list",
			}, nil)
			return
		}
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", nil, nil)
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

// List retrieves all waiting list entries (admin only)
func (h *WaitingListHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	items, total, err := h.uc.List(c.Request.Context(), limit, offset, status)
	if err != nil {
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", nil, nil)
		return
	}

	pagination := response.NewPaginationMeta(page, limit, int(total))
	meta := &response.Meta{Pagination: pagination}

	response.SuccessResponse(c, items, meta)
}

// UpdateStatus changes the status of a waiting list entry (admin only)
func (h *WaitingListHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateWaitingListStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	res, err := h.uc.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, usecase.ErrEntryNotFound) {
			coreErrors.NotFoundResponse(c, "waiting_list", id)
			return
		}
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", nil, nil)
		return
	}

	response.SuccessResponse(c, res, nil)
}

// Delete removes an entry from the waiting list (admin only)
func (h *WaitingListHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.uc.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrEntryNotFound) {
			coreErrors.NotFoundResponse(c, "waiting_list", id)
			return
		}
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", nil, nil)
		return
	}

	response.SuccessResponseDeleted(c, "waiting_list", id, nil)
}
