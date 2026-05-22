package handler

import (
	"math"
	"strconv"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// POSSessionHandler handles POS shift sessions
type POSSessionHandler struct {
	uc usecase.POSSessionUsecase
}

// NewPOSSessionHandler creates the handler
func NewPOSSessionHandler(uc usecase.POSSessionUsecase) *POSSessionHandler {
	return &POSSessionHandler{uc: uc}
}

// Open opens a new POS session for the authenticated cashier
func (h *POSSessionHandler) Open(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.OpenSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	session, err := h.uc.Open(c.Request.Context(), &req, uc.userID)
	if err != nil {
		handlePOSSessionError(c, err)
		return
	}
	response.SuccessResponse(c, session, nil)
}

// Close closes the given session
func (h *POSSessionHandler) Close(c *gin.Context) {
	_, ok := extractUserContext(c)
	if !ok {
		return
	}

	sessionID := c.Param("id")
	var req dto.CloseSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	session, err := h.uc.Close(c.Request.Context(), sessionID, &req, "")
	if err != nil {
		handlePOSSessionError(c, err)
		return
	}
	response.SuccessResponse(c, session, nil)
}

// GetByID returns a single session
func (h *POSSessionHandler) GetByID(c *gin.Context) {
	session, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		handlePOSSessionError(c, err)
		return
	}
	response.SuccessResponse(c, session, nil)
}

// GetActive returns the open session for the authenticated cashier
func (h *POSSessionHandler) GetActive(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	session, err := h.uc.GetActiveSession(c.Request.Context(), uc.userID)
	if err != nil {
		handlePOSSessionError(c, err)
		return
	}
	response.SuccessResponse(c, session, nil)
}

// List returns paginated sessions
func (h *POSSessionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	params := repositories.POSSessionListParams{
		OutletID:  c.Query("outlet_id"),
		CashierID: c.Query("cashier_id"),
		Status:    c.Query("status"),
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	sessions, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	hasNext := page < totalPages
	hasPrev := page > 1
	var nextPage, prevPage *int
	if hasNext {
		n := page + 1
		nextPage = &n
	}
	if hasPrev {
		p := page - 1
		prevPage = &p
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page: page, PerPage: perPage, Total: int(total),
			TotalPages: totalPages, HasNext: hasNext, HasPrev: hasPrev,
			NextPage: nextPage, PrevPage: prevPage,
		},
	}
	response.SuccessResponse(c, sessions, meta)
}

func handlePOSSessionError(c *gin.Context, err error) {
	switch err {
	case usecase.ErrPOSSessionNotFound:
		coreErrors.NotFoundResponse(c, "pos_session", "")
	case usecase.ErrPOSSessionAlreadyOpen:
		coreErrors.ErrorResponse(c, "POS_SESSION_ALREADY_OPEN", nil, nil)
	case usecase.ErrPOSOutletForbidden:
		coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
	default:
		coreErrors.InternalServerErrorResponse(c, "")
	}
}
