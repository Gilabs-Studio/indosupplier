package handler

import (
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type OpeningBalanceHandler struct {
	uc usecase.OpeningBalanceUsecase
}

func NewOpeningBalanceHandler(uc usecase.OpeningBalanceUsecase) *OpeningBalanceHandler {
	return &OpeningBalanceHandler{uc: uc}
}

func (h *OpeningBalanceHandler) Get(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	fiscalYearID := strings.TrimSpace(c.Query("fiscal_year_id"))
	if companyID == "" || fiscalYearID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id and fiscal_year_id are required", nil)
		return
	}
	res, err := h.uc.Get(c.Request.Context(), companyID, fiscalYearID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *OpeningBalanceHandler) Upsert(c *gin.Context) {
	var req dto.UpsertOpeningBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	if err := h.uc.UpsertDraft(c.Request.Context(), &req); err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, gin.H{"saved": true}, nil)
}

func (h *OpeningBalanceHandler) Validate(c *gin.Context) {
	var req dto.ValidateOpeningBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.Validate(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *OpeningBalanceHandler) Simulate(c *gin.Context) {
	var req dto.ValidateOpeningBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.Simulate(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *OpeningBalanceHandler) Post(c *gin.Context) {
	var req dto.PostOpeningBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	actorID := ""
	if actor, exists := c.Get("user_id"); exists {
		if actorStr, ok := actor.(string); ok {
			actorID = strings.TrimSpace(actorStr)
		}
	}

	res, err := h.uc.Post(c.Request.Context(), &req, actorID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *OpeningBalanceHandler) Summary(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	fiscalYearID := strings.TrimSpace(c.Query("fiscal_year_id"))
	if companyID == "" || fiscalYearID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id and fiscal_year_id are required", nil)
		return
	}
	res, err := h.uc.Summary(c.Request.Context(), companyID, fiscalYearID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}
