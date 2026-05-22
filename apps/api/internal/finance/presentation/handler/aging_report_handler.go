package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type AgingReportHandler struct {
	uc usecase.AgingReportUsecase
}

func NewAgingReportHandler(uc usecase.AgingReportUsecase) *AgingReportHandler {
	return &AgingReportHandler{uc: uc}
}

func parseAsOfDate(c *gin.Context) (time.Time, error) {
	v := strings.TrimSpace(c.Query("as_of_date"))
	if v == "" {
		now := apptime.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	parsed, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}

func parseIncludeCurrent(c *gin.Context) (bool, error) {
	raw := strings.TrimSpace(c.Query("include_current"))
	if raw == "" {
		return true, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err == nil {
		return parsed, nil
	}
	if raw == "1" {
		return true, nil
	}
	if raw == "0" {
		return false, nil
	}
	return false, err
}

func parseMinAmount(c *gin.Context) (float64, error) {
	raw := strings.TrimSpace(c.Query("min_amount"))
	if raw == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, nil
	}
	return parsed, nil
}

func parseAgingFinanceQuery(c *gin.Context) (dto.AgingFinanceQuery, error) {
	asOf, err := parseAsOfDate(c)
	if err != nil {
		return dto.AgingFinanceQuery{}, err
	}
	includeCurrent, err := parseIncludeCurrent(c)
	if err != nil {
		return dto.AgingFinanceQuery{}, err
	}
	minAmount, err := parseMinAmount(c)
	if err != nil {
		return dto.AgingFinanceQuery{}, err
	}

	return dto.AgingFinanceQuery{
		AsOfDate:       asOf,
		Search:         strings.TrimSpace(c.Query("search")),
		MinAmount:      minAmount,
		IncludeCurrent: includeCurrent,
	}, nil
}

func (h *AgingReportHandler) ARAging(c *gin.Context) {
	asOf, err := parseAsOfDate(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid as_of_date", nil, nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	search := strings.TrimSpace(c.Query("search"))

	res, total, err := h.uc.ARAging(c.Request.Context(), asOf, search, page, perPage)
	if err != nil {
		log.Printf("[aging][ar] failed: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "AR_AGING_FAILED", "failed to fetch AR aging report", nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, res, meta)
}

func (h *AgingReportHandler) APAging(c *gin.Context) {
	asOf, err := parseAsOfDate(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid as_of_date", nil, nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	search := strings.TrimSpace(c.Query("search"))

	res, total, err := h.uc.APAging(c.Request.Context(), asOf, search, page, perPage)
	if err != nil {
		log.Printf("[aging][ap] failed: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "AP_AGING_FAILED", "failed to fetch AP aging report", nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, res, meta)
}

func (h *AgingReportHandler) ARAgingFinance(c *gin.Context) {
	query, err := parseAgingFinanceQuery(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid aging query parameters", nil, nil)
		return
	}
	partner := strings.TrimSpace(c.Query("customer_id"))
	if partner != "" {
		if _, err := uuid.Parse(partner); err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid customer_id", nil, nil)
			return
		}
	}
	query.PartnerID = partner

	res, err := h.uc.ARAgingFinance(c.Request.Context(), query)
	if err != nil {
		log.Printf("[aging][ar_finance] failed: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "AR_AGING_FAILED", "failed to fetch AR aging report", nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AgingReportHandler) APAgingFinance(c *gin.Context) {
	query, err := parseAgingFinanceQuery(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid aging query parameters", nil, nil)
		return
	}
	partner := strings.TrimSpace(c.Query("supplier_id"))
	if partner != "" {
		if _, err := uuid.Parse(partner); err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid supplier_id", nil, nil)
			return
		}
	}
	query.PartnerID = partner

	res, err := h.uc.APAgingFinance(c.Request.Context(), query)
	if err != nil {
		log.Printf("[aging][ap_finance] failed: %v", err)
		response.ErrorResponse(c, http.StatusInternalServerError, "AP_AGING_FAILED", "failed to fetch AP aging report", nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}
