package handler

import (
	"net/http"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

// ARAPReconciliationHandler handles HTTP requests for AR/AP reconciliation.
type ARAPReconciliationHandler struct {
	uc usecase.ARAPReconciliationUsecase
}

// NewARAPReconciliationHandler creates a new instance of ARAPReconciliationHandler.
func NewARAPReconciliationHandler(uc usecase.ARAPReconciliationUsecase) *ARAPReconciliationHandler {
	return &ARAPReconciliationHandler{uc: uc}
}

// ReconcileAR reconciles Sales Accounts Receivable (Customer Invoices vs GL).
func (h *ARAPReconciliationHandler) ReconcileAR(c *gin.Context) {
	asOf, err := parseAsOfDateFromCtx(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid as_of_date", nil, nil)
		return
	}

	report, err := h.uc.ReconcileAR(c.Request.Context(), asOf)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "AR_RECONCILIATION_FAILED", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, report, nil)
}

// ReconcileAP reconciles Purchase Accounts Payable (Supplier Invoices vs GL).
func (h *ARAPReconciliationHandler) ReconcileAP(c *gin.Context) {
	asOf, err := parseAsOfDateFromCtx(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid as_of_date", nil, nil)
		return
	}

	report, err := h.uc.ReconcileAP(c.Request.Context(), asOf)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "AP_RECONCILIATION_FAILED", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, report, nil)
}

func parseAsOfDateFromCtx(c *gin.Context) (time.Time, error) {
	v := c.Query("as_of_date")
	if v == "" {
		now := apptime.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC), nil
	}
	parsed, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, time.UTC), nil
}
