package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/travel_planner/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	travelVisitReadPermission  = "travel.visit.read"
	travelVisitCreatePermission = "travel.visit.create"
	travelPlannerReadPermission = "travel_planner.read"
	travelPlannerUpdatePermission = "travel_planner.update"
	travelPlannerDeletePermission = "travel_planner.delete"
	travelPlannerCreatePermission = "travel_planner.create"
)

func RegisterTravelPlanRoutes(r *gin.RouterGroup, h *handler.TravelPlanHandler) {
	r.GET("/visit-planner/form-data", middleware.RequirePermission(travelVisitReadPermission), h.GetVisitPlannerFormData)
	r.GET("/visit-planner/routes", middleware.RequirePermission(travelVisitReadPermission), h.ListVisitPlannerRoutes)
	r.POST("/visit-planner/plans", middleware.RequirePermission(travelVisitCreatePermission), h.CreateVisitPlannerPlan)
	r.POST("/navigation/optimize", middleware.RequirePermission(travelVisitReadPermission), h.OptimizeNavigationForVisit)
	r.POST("/visits", middleware.RequirePermission(travelVisitCreatePermission), h.UpsertVisitLog)
	r.POST("/locations", middleware.RequirePermission(travelVisitCreatePermission), h.UpsertLocation)

	// Navigation lifecycle: placed before the /travel-planner group to keep the
	// path structure flat and avoid nesting under travel-planner resources.
	navigation := r.Group("/locations/navigation")
	{
		navigation.POST("/start", middleware.RequirePermission(travelVisitCreatePermission), h.StartNavigation)
		navigation.POST("/stop", middleware.RequirePermission(travelVisitCreatePermission), h.StopNavigation)
	}

	planner := r.Group("/travel-planner")
	{
		// Fixed routes before /plans/:id to avoid route shadowing.
		planner.GET("/form-data", middleware.RequirePermission(travelPlannerReadPermission), h.GetFormData)
		planner.GET("/participants", middleware.RequirePermission(travelPlannerReadPermission), h.ListParticipants)
		planner.GET("/place-search", middleware.RequirePermission(travelPlannerReadPermission), h.SearchPlaces)
		planner.GET("/visits/available", middleware.RequirePermission(travelPlannerReadPermission), h.ListAvailableVisits)

		plans := planner.Group("/plans")
		{
			plans.GET("", middleware.RequirePermission(travelPlannerReadPermission), h.List)
			plans.POST("", middleware.RequirePermission(travelPlannerCreatePermission), h.Create)

			plans.GET("/:id", middleware.RequirePermission(travelPlannerReadPermission), h.GetByID)
			plans.PUT("/:id", middleware.RequirePermission(travelPlannerUpdatePermission), h.Update)
			plans.PATCH("/:id/participants", middleware.RequirePermission(travelPlannerUpdatePermission), h.UpdateParticipants)
			plans.DELETE("/:id", middleware.RequirePermission(travelPlannerDeletePermission), h.Delete)

			plans.POST("/:id/optimize-route", middleware.RequirePermission(travelPlannerUpdatePermission), h.OptimizeRoute)
			plans.GET("/:id/google-maps-links", middleware.RequirePermission(travelPlannerReadPermission), h.GetGoogleMapsLinks)
			plans.GET("/:id/export/pdf", middleware.RequirePermission(travelPlannerReadPermission), h.ExportPDF)
			plans.GET("/:id/export/report-html", middleware.RequirePermission(travelPlannerReadPermission), h.ExportReportHTML)

			plans.GET("/:id/expenses", middleware.RequirePermission(travelPlannerReadPermission), h.ListExpenses)
			plans.POST("/:id/expenses", middleware.RequirePermission(travelPlannerUpdatePermission), h.CreateExpense)
			plans.DELETE("/:id/expenses/:expenseId", middleware.RequirePermission(travelPlannerDeletePermission), h.DeleteExpense)

			plans.GET("/:id/visits", middleware.RequirePermission(travelPlannerReadPermission), h.ListVisits)
			plans.POST("/:id/visits", middleware.RequirePermission(travelPlannerCreatePermission), h.CreateVisitFromTrip)
			plans.POST("/:id/visits/link", middleware.RequirePermission(travelPlannerUpdatePermission), h.LinkVisits)
			plans.DELETE("/:id/visits/:visitId", middleware.RequirePermission(travelPlannerUpdatePermission), h.UnlinkVisit)
		}
	}
}

func RegisterTravelPlannerWebSocketRoutes(r *gin.RouterGroup, h *handler.TravelPlanHandler) {
	r.GET("/travel/locations", middleware.RequirePermission(travelVisitReadPermission), h.TravelLocationsWebSocket)
}
