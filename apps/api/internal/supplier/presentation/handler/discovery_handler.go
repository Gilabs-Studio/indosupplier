package handler

import (
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
