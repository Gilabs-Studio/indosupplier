package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/usecase"
)

type DiscoveryHandler struct {
	discoveryUC usecase.DiscoveryUsecase
}

func NewDiscoveryHandler(discoveryUC usecase.DiscoveryUsecase) *DiscoveryHandler {
	return &DiscoveryHandler{discoveryUC: discoveryUC}
}

func (h *DiscoveryHandler) List(c *gin.Context) {
	q := c.Query("q")
	category := c.Query("category")
	region := c.Query("region")
	verifiedOnly := c.Query("verified") == "true"

	suppliers, err := h.discoveryUC.List(c.Request.Context(), q, category, region, verifiedOnly)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, suppliers, nil)
}

func (h *DiscoveryHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	supplier, err := h.discoveryUC.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		errors.ErrorResponse(c, "SUPPLIER_PROFILE_NOT_FOUND", nil, nil)
		return
	}

	response.SuccessResponse(c, supplier, nil)
}

func (h *DiscoveryHandler) ListProducts(c *gin.Context) {
	q := c.Query("q")
	products, err := h.discoveryUC.ListProducts(c.Request.Context(), q)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, products, nil)
}

func (h *DiscoveryHandler) GetProductByID(c *gin.Context) {
	product, err := h.discoveryUC.GetProductByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		errors.NotFoundResponse(c, "product", c.Param("id"))
		return
	}

	response.SuccessResponse(c, product, nil)
}

func (h *DiscoveryHandler) LookupSuppliers(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	suppliers, err := h.discoveryUC.LookupSuppliers(c.Request.Context(), q, page, limit)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, suppliers, nil)
}

func (h *DiscoveryHandler) LookupProducts(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	products, err := h.discoveryUC.LookupProducts(c.Request.Context(), q, page, limit)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, products, nil)
}
