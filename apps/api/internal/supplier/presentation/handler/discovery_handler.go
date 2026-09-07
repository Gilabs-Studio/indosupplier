package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
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
	supplierID := c.Query("supplier_id")
	category := c.Query("category")
	location := c.Query("location")
	if location == "" {
		location = c.Query("region")
	}
	sort := c.Query("sort")
	verifiedOnly := c.Query("verified") == "true"
	powerSupplierOnly := c.Query("power_supplier") == "true"
	readyStockOnly := c.Query("ready_stock") == "true"

	var minPricePtr, maxPricePtr *float64
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPricePtr = &val
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPricePtr = &val
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("limit", "12")))

	params := dto.ListPublicProductsParams{
		Query:             q,
		Category:          category,
		SupplierID:        supplierID,
		Location:          location,
		MinPrice:          minPricePtr,
		MaxPrice:          maxPricePtr,
		MinOrder:          c.Query("min_order"),
		VerifiedOnly:      verifiedOnly,
		PowerSupplierOnly: powerSupplierOnly,
		ReadyStockOnly:    readyStockOnly,
		Sort:              sort,
		Page:              page,
		Limit:             perPage,
	}

	products, err := h.discoveryUC.ListProducts(c.Request.Context(), params)
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
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))

	suppliers, err := h.discoveryUC.LookupSuppliers(c.Request.Context(), q, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, suppliers, nil)
}

func (h *DiscoveryHandler) LookupProducts(c *gin.Context) {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))

	products, err := h.discoveryUC.LookupProducts(c.Request.Context(), q, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, products, nil)
}
