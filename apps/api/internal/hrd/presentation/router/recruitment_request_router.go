package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// SetupRecruitmentRequestRoutes registers all recruitment request routes
func SetupRecruitmentRequestRoutes(router *gin.RouterGroup, h *handler.RecruitmentRequestHandler, ah *handler.RecruitmentApplicantHandler) {
	recruitment := router.Group("/recruitment-requests")
	{
		// CRITICAL: form-data before /:id to avoid route conflicts
		recruitment.GET("/form-data", middleware.RequirePermission("recruitment.read"), h.GetFormData)
		recruitment.GET("", middleware.RequirePermission("recruitment.read"), h.GetAll)
		recruitment.GET("/:id", middleware.RequirePermission("recruitment.read"), h.GetByID)
		recruitment.POST("", middleware.RequirePermission("recruitment.create"), h.Create)
		recruitment.PUT("/:id", middleware.RequirePermission("recruitment.update"), h.Update)
		recruitment.DELETE("/:id", middleware.RequirePermission("recruitment.delete"), h.Delete)

		// Status workflow
		recruitment.POST("/:id/status", middleware.RequirePermission("recruitment.update"), h.UpdateStatus)

		// Status action endpoints (convenience wrappers)
		recruitment.POST("/:id/submit", middleware.RequirePermission("recruitment.update"), h.Submit)
		recruitment.POST("/:id/approve", middleware.RequirePermission("recruitment.approve"), h.Approve)
		recruitment.POST("/:id/reject", middleware.RequirePermission("recruitment.approve"), h.Reject)
		recruitment.POST("/:id/open", middleware.RequirePermission("recruitment.update"), h.Open)
		recruitment.POST("/:id/close", middleware.RequirePermission("recruitment.update"), h.Close)
		recruitment.POST("/:id/cancel", middleware.RequirePermission("recruitment.update"), h.Cancel)

		// Filled count management
		recruitment.PUT("/:id/filled-count", middleware.RequirePermission("recruitment.update"), h.UpdateFilledCount)

		// Nested applicants route
		recruitment.GET("/:id/applicants", middleware.RequirePermission("recruitment.read"), ah.GetByRecruitmentRequest)
	}
}
