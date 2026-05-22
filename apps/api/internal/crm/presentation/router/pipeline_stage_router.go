package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	pipelineStageRead   = "crm_pipeline_stage.read"
	pipelineStageCreate = "crm_pipeline_stage.create"
	pipelineStageUpdate = "crm_pipeline_stage.update"
	pipelineStageDelete = "crm_pipeline_stage.delete"
)

// RegisterPipelineStageRoutes registers pipeline stage routes
func RegisterPipelineStageRoutes(r *gin.RouterGroup, h *handler.PipelineStageHandler) {
	g := r.Group("/pipeline-stages")
	g.GET("", middleware.RequirePermission(pipelineStageRead), h.List)
	g.POST("", middleware.RequirePermission(pipelineStageCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(pipelineStageRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(pipelineStageUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(pipelineStageDelete), h.Delete)
}
