package handler

import (
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
)

// AreaMappingHandler handles area mapping requests
type AreaMappingHandler struct {
	usecase usecase.AreaMappingUsecase
}

// NewAreaMappingHandler creates a new area mapping handler
func NewAreaMappingHandler(uc usecase.AreaMappingUsecase) *AreaMappingHandler {
	return &AreaMappingHandler{
		usecase: uc,
	}
}

// GetAreaMapping returns customers and leads for map visualization
// @Summary Get area mapping data
// @Description Returns all customers and leads with their activity metrics for map visualization
// @Success 200 {object} response.SuccessResponse
// @Router /crm/area-mapping/map [get]
func (h *AreaMappingHandler) GetAreaMapping(c *gin.Context) {
	var req dto.GetAreaMappingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, 400, "INVALID_QUERY", "Invalid query parameters", nil, nil)
		return
	}

	mappingData, err := h.usecase.GetAreaMapping(c.Request.Context(), &req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "INVALID_MONTH") || strings.HasPrefix(err.Error(), "INVALID_YEAR") {
			response.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, 500, "AREA_MAPPING_FETCH_ERROR", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, mappingData, nil)
}
