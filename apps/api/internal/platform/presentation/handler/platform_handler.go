package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/platform/domain/usecase"
)

type PlatformHandler struct {
	uc usecase.PlatformUsecase
}

func NewPlatformHandler(uc usecase.PlatformUsecase) *PlatformHandler {
	return &PlatformHandler{uc: uc}
}

func (h *PlatformHandler) FeatureStatus(c *gin.Context) {
	response.SuccessResponse(c, h.uc.GetFeatureStatus(), nil)
}

func (h *PlatformHandler) Catalog(c *gin.Context) {
	response.SuccessResponse(c, h.uc.GetCatalog(), nil)
}

func (h *PlatformHandler) Dashboard(c *gin.Context) {
	persona := c.Param("persona")
	response.SuccessResponse(c, h.uc.GetDashboard(persona), nil)
}
