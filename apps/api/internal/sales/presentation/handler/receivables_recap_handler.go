package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/exportjob"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
)

type ReceivablesRecapHandler struct {
	uc usecase.ReceivablesRecapUsecase
}

func NewReceivablesRecapHandler(uc usecase.ReceivablesRecapUsecase) *ReceivablesRecapHandler {
	return &ReceivablesRecapHandler{uc: uc}
}

// List handles GET /sales/receivables-recap
func (h *ReceivablesRecapHandler) List(c *gin.Context) {
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

	params := repositories.ReceivablesRecapListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", "outstanding_amount"),
		SortDir: c.DefaultQuery("sort_dir", "desc"),
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, perPage, int(total)),
		Filters:    map[string]interface{}{},
		Sort:       &response.SortMeta{Field: params.SortBy, Order: params.SortDir},
	}
	if strings.TrimSpace(params.Search) != "" {
		meta.Filters["search"] = params.Search
	}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// Summary handles GET /sales/receivables-recap/summary
func (h *ReceivablesRecapHandler) Summary(c *gin.Context) {
	summary, err := h.uc.GetSummary(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, summary, nil)
}

// Export handles GET /sales/receivables-recap/export
func (h *ReceivablesRecapHandler) Export(c *gin.Context) {
	params := repositories.ReceivablesRecapListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", "outstanding_amount"),
		SortDir: c.DefaultQuery("sort_dir", "desc"),
		Limit:   10000,
	}

	generator := func(ctx context.Context) (*exportjob.GeneratedFile, error) {
		data, err := h.uc.ExportCSV(ctx, params)
		if err != nil {
			return nil, err
		}
		return &exportjob.GeneratedFile{
			FileName:    "receivables_recap.csv",
			ContentType: "text/csv",
			Bytes:       data,
		}, nil
	}

	if exportjob.QueueIfRequested(c, generator) {
		return
	}

	file, err := generator(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	exportjob.WriteSyncFile(c, file)
}
