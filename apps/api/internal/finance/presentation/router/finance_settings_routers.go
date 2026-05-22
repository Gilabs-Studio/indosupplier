package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterFinanceSettingsRoutes(
	group *gin.RouterGroup,
	h *handler.FinanceSettingsHandler,
) {
	settings := group.Group("/settings")
	settings.GET("", middleware.RequirePermission("finance_settings.read"), h.GetAll)
	settings.PUT("", middleware.RequirePermission("finance_settings.update"), h.BatchUpsert)

	settings.GET("/aging-buckets", middleware.RequirePermission("finance_settings.read"), h.GetAgingBucketConfig)
	settings.PUT("/aging-buckets", middleware.RequirePermission("finance_settings.update"), h.UpsertAgingBucketConfig)
}
