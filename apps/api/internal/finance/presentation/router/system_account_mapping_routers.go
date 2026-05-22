package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterSystemAccountMappingRoutes(group *gin.RouterGroup, h *handler.SystemAccountMappingHandler) {
	registerSystemAccountMappingRoutesInGroup(group.Group("/settings/account-mappings"), h)
	registerSystemAccountMappingRoutesInGroup(group.Group("/settings/accounting-mapping"), h)
}

func registerSystemAccountMappingRoutesInGroup(mappings *gin.RouterGroup, h *handler.SystemAccountMappingHandler) {
	mappings.GET("",
		middleware.RequirePermission("account_mappings.read"),
		h.List,
	)
	mappings.GET("/:key",
		middleware.RequirePermission("account_mappings.read"),
		h.GetByKey,
	)
	mappings.PUT("/:key",
		middleware.RequirePermission("account_mappings.update"),
		h.Upsert,
	)
	mappings.DELETE("/:key",
		middleware.RequirePermission("account_mappings.delete"),
		h.Delete,
	)
}
