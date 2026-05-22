package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/crm/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	visitRead    = "crm_visit.read"
	visitCreate  = "crm_visit.create"
	visitUpdate  = "crm_visit.update"
	visitDelete  = "crm_visit.delete"
	visitApprove = "crm_visit.approve"
)

// RegisterVisitReportRoutes registers all visit report routes
func RegisterVisitReportRoutes(r *gin.RouterGroup, h *handler.VisitReportHandler, printH *handler.VisitReportPrintHandler) {
	g := r.Group("/visits")

	// Static routes first (before parameterized routes)
	g.GET("/form-data", middleware.RequirePermission(visitRead), h.GetFormData)
	// Team-level employee summary — for ALL/DIVISION/AREA scope views
	g.GET("/by-employee", middleware.RequirePermission(visitRead), h.ListByEmployee)

	// CRUD routes
	g.GET("", middleware.RequirePermission(visitRead), h.List)
	g.POST("", middleware.RequirePermission(visitCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(visitRead), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(visitUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(visitDelete), h.Delete)

	// Print PDF
	g.GET("/:id/print", middleware.RequirePermission(visitRead), printH.PrintVisitReport)

	// Workflow actions
	g.POST("/:id/submit", middleware.RequirePermission(visitUpdate), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(visitApprove), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(visitApprove), h.Reject)
	g.POST("/:id/check-in", middleware.RequirePermission(visitUpdate), h.CheckIn)
	g.POST("/:id/check-out", middleware.RequirePermission(visitUpdate), h.CheckOut)
	g.POST("/:id/photos", middleware.RequirePermission(visitUpdate), h.UploadPhotos)
}
