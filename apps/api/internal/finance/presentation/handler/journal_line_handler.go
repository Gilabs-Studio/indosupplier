package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

// JournalLineHandler handles HTTP requests for the journal lines sub-ledger view.
type JournalLineHandler struct {
	uc usecase.JournalLineUsecase
}

// NewJournalLineHandler creates a new JournalLineHandler.
func NewJournalLineHandler(uc usecase.JournalLineUsecase) *JournalLineHandler {
	return &JournalLineHandler{uc: uc}
}

// ListLines handles GET /finance/journal-lines
func (h *JournalLineHandler) ListLines(c *gin.Context) {
	var req dto.ListJournalLinesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	req.Page = page
	req.PerPage = perPage

	result, total, err := h.uc.ListLines(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "JOURNAL_LINES_LIST_FAILED", err.Error(), nil, nil)
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, result, meta)
}

// ExportLines handles GET /finance/journal-lines/export (CSV download)
func (h *JournalLineHandler) ExportLines(c *gin.Context) {
	var req dto.ListJournalLinesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	// Generate filename with date range
	filename := "journal_lines"
	if req.StartDate != nil {
		filename += "_" + *req.StartDate
	}
	if req.EndDate != nil {
		filename += "_" + *req.EndDate
	}
	if filename == "journal_lines" {
		filename += "_" + time.Now().Format("2006-01-02")
	}
	filename += ".csv"

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		var buffer bytes.Buffer
		if err := h.uc.ExportLinesCSV(ctx, &req, &buffer); err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    filename,
			ContentType: "text/csv",
			Bytes:       buffer.Bytes(),
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

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
	exportjob.WriteSyncFile(c, file)
}
