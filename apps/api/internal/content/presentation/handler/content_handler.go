package handler

import (
	stderrors "errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/content/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/content/domain/usecase"
	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
)

type ContentHandler struct {
	uc usecase.ContentUsecase
}

func NewContentHandler(uc usecase.ContentUsecase) *ContentHandler {
	return &ContentHandler{uc: uc}
}

func (h *ContentHandler) ListPublic(c *gin.Context) {
	req := listRequestFromQuery(c)
	articles, total, err := h.uc.ListPublic(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildListMeta(req, int(total))
	response.SuccessResponse(c, articles, meta)
}

func (h *ContentHandler) GetPublicBySlug(c *gin.Context) {
	article, err := h.uc.GetPublicBySlug(c.Request.Context(), c.Param("slug"), c.DefaultQuery("locale", "id"))
	if err != nil {
		if stderrors.Is(err, usecase.ErrContentArticleNotFound) {
			errors.NotFoundResponse(c, "content_article", c.Param("slug"))
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, article, nil)
}

func (h *ContentHandler) ListAdmin(c *gin.Context) {
	req := listRequestFromQuery(c)
	articles, total, err := h.uc.ListAdmin(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := buildListMeta(req, int(total))
	response.SuccessResponse(c, articles, meta)
}

func (h *ContentHandler) GetAdminByID(c *gin.Context) {
	article, err := h.uc.GetAdminByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if stderrors.Is(err, usecase.ErrContentArticleNotFound) {
			errors.NotFoundResponse(c, "content_article", c.Param("id"))
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, article, nil)
}

func (h *ContentHandler) Create(c *gin.Context) {
	var req dto.CreateContentArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	article, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	adminID := ""
	if value, ok := c.Get("admin_id"); ok {
		if id, ok := value.(string); ok {
			adminID = id
		}
	}
	response.SuccessResponseCreated(c, article, &response.Meta{CreatedBy: adminID})
}

func (h *ContentHandler) Update(c *gin.Context) {
	var req dto.UpdateContentArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	article, err := h.uc.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrContentArticleNotFound) {
			errors.NotFoundResponse(c, "content_article", c.Param("id"))
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	adminID := ""
	if value, ok := c.Get("admin_id"); ok {
		if id, ok := value.(string); ok {
			adminID = id
		}
	}
	response.SuccessResponse(c, article, &response.Meta{UpdatedBy: adminID})
}

func (h *ContentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if stderrors.Is(err, usecase.ErrContentArticleNotFound) {
			errors.NotFoundResponse(c, "content_article", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "content_article", id, nil)
}

func listRequestFromQuery(c *gin.Context) dto.ListContentArticlesRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))
	page, perPage = utils.NormalizePagination(page, perPage, 12)

	return dto.ListContentArticlesRequest{
		Type:    c.Query("type"),
		Locale:  c.Query("locale"),
		Status:  c.Query("status"),
		Search:  c.Query("search"),
		Page:    page,
		PerPage: perPage,
	}
}

func buildListMeta(req dto.ListContentArticlesRequest, total int) *response.Meta {
	return &response.Meta{
		Pagination: response.NewPaginationMeta(req.Page, req.PerPage, total),
		Filters: map[string]interface{}{
			"type":   req.Type,
			"locale": req.Locale,
			"status": req.Status,
			"search": req.Search,
		},
	}
}
