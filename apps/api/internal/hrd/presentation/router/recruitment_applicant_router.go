package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

// SetupRecruitmentApplicantRoutes registers all recruitment applicant routes
func SetupRecruitmentApplicantRoutes(router *gin.RouterGroup, h *handler.RecruitmentApplicantHandler) {
	// Applicant stages (public read, admin write)
	router.GET("/applicant-stages", middleware.RequirePermission("recruitment.read"), h.GetStages)

	// Applicants
	applicants := router.Group("/applicants")
	{
		applicants.GET("", middleware.RequirePermission("recruitment.read"), h.GetAll)
		applicants.GET("/by-stage", middleware.RequirePermission("recruitment.read"), h.GetByStage)
		applicants.GET("/:id", middleware.RequirePermission("recruitment.read"), h.GetByID)
		applicants.POST("", middleware.RequirePermission("recruitment.create"), h.Create)
		applicants.PUT("/:id", middleware.RequirePermission("recruitment.update"), h.Update)
		applicants.DELETE("/:id", middleware.RequirePermission("recruitment.delete"), h.Delete)

		// Stage movement
		applicants.POST("/:id/move-stage", middleware.RequirePermission("recruitment.update"), h.MoveStage)

		// Activities
		applicants.GET("/:id/activities", middleware.RequirePermission("recruitment.read"), h.GetActivities)
		applicants.POST("/:id/activities", middleware.RequirePermission("recruitment.update"), h.AddActivity)

		// Convert to employee
		applicants.GET("/:id/can-convert", middleware.RequirePermission("recruitment.read"), h.CanConvertToEmployee)
		applicants.POST("/:id/convert-to-employee", middleware.RequirePermission("recruitment.update"), h.ConvertToEmployee)
	}

	// Applicants by recruitment request (nested under recruitment-requests)
	// This route is registered in the recruitment request router
}
