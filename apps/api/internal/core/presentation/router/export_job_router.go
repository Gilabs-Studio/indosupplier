package router

import (
	"github.com/gilabs/gims/api/internal/core/presentation/handler"
	"github.com/gin-gonic/gin"
)

func RegisterExportJobRoutes(group *gin.RouterGroup, h *handler.ExportJobHandler) {
	jobs := group.Group("/exports/jobs")
	jobs.GET("/:id", h.Get)
	jobs.GET("/:id/download", h.Download)
}
