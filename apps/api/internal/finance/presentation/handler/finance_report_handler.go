package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type FinanceReportHandler struct {
	uc usecase.FinanceReportUsecase
}

func NewFinanceReportHandler(uc usecase.FinanceReportUsecase) *FinanceReportHandler {
	return &FinanceReportHandler{uc: uc}
}

func parseDateOrDefault(c *gin.Context, key string, def time.Time) time.Time {
	val := c.Query(key)
	if val == "" {
		return def
	}
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return def
	}
	return t
}

func parseDateWithAliases(c *gin.Context, keys []string, def time.Time) time.Time {
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		val := strings.TrimSpace(c.Query(key))
		if val == "" {
			continue
		}
		if parsed, err := time.Parse("2006-01-02", val); err == nil {
			return parsed
		}
	}
	return def
}

func parseOptionalCompanyID(c *gin.Context) *string {
	value := strings.TrimSpace(c.Query("company_id"))
	if value == "" {
		return nil
	}
	return &value
}

func parseOptionalFiscalYearID(c *gin.Context) *string {
	value := strings.TrimSpace(c.Query("fiscal_year_id"))
	if value == "" {
		return nil
	}
	return &value
}

func parseOptionalAccountID(c *gin.Context) *string {
	value := strings.TrimSpace(c.Query("account_id"))
	if value == "" {
		return nil
	}
	return &value
}

func parseIncludeZero(c *gin.Context) bool {
	raw := strings.TrimSpace(c.Query("include_zero"))
	if raw == "" {
		return false
	}
	parsed, err := strconv.ParseBool(raw)
	if err == nil {
		return parsed
	}
	return raw == "1"
}

func (h *FinanceReportHandler) GeneralLedger(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)
	accountID := parseOptionalAccountID(c)

	res, err := h.uc.GetGeneralLedger(c.Request.Context(), startDate, endDate, companyID, fiscalYearID, accountID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GENERAL_LEDGER_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinanceReportHandler) BalanceSheet(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)
	includeZero := parseIncludeZero(c)

	res, err := h.uc.GetBalanceSheet(c.Request.Context(), startDate, endDate, companyID, fiscalYearID, includeZero)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "BALANCE_SHEET_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinanceReportHandler) ProfitAndLoss(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)

	res, err := h.uc.GetProfitAndLoss(c.Request.Context(), startDate, endDate, companyID, fiscalYearID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "PROFIT_LOSS_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinanceReportHandler) TrialBalance(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)

	res, err := h.uc.GetTrialBalance(c.Request.Context(), startDate, endDate, companyID, fiscalYearID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "TRIAL_BALANCE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinanceReportHandler) CashFlow(c *gin.Context) {
	startDate := parseDateWithAliases(c, []string{"from_date", "start_date"}, apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateWithAliases(c, []string{"to_date", "end_date"}, apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)

	res, err := h.uc.GetCashFlow(c.Request.Context(), startDate, endDate, companyID, fiscalYearID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CASH_FLOW_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinanceReportHandler) ExportGeneralLedger(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)
	accountID := parseOptionalAccountID(c)
	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		bytes, err := h.uc.ExportGeneralLedger(ctx, startDate, endDate, companyID, fiscalYearID, accountID)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    "general_ledger.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Bytes:       bytes,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil, nil)
		return
	}
	exportjob.WriteSyncFile(c, file)
}

func (h *FinanceReportHandler) ExportBalanceSheet(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)
	includeZero := parseIncludeZero(c)

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		bytes, err := h.uc.ExportBalanceSheet(ctx, startDate, endDate, companyID, fiscalYearID, includeZero)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    "balance_sheet.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Bytes:       bytes,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil, nil)
		return
	}
	exportjob.WriteSyncFile(c, file)
}

func (h *FinanceReportHandler) ExportProfitAndLoss(c *gin.Context) {
	startDate := parseDateOrDefault(c, "start_date", apptime.Now().AddDate(0, -1, 0))
	endDate := parseDateOrDefault(c, "end_date", apptime.Now())
	companyID := parseOptionalCompanyID(c)
	fiscalYearID := parseOptionalFiscalYearID(c)

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		bytes, err := h.uc.ExportProfitAndLoss(ctx, startDate, endDate, companyID, fiscalYearID)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    "profit_and_loss.xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Bytes:       bytes,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil, nil)
		return
	}
	exportjob.WriteSyncFile(c, file)
}
