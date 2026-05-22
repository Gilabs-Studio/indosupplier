package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCompanyRoutes registers company routes.
// Company creation is disabled — tenants may only have one company (provisioned via seeder).
// To add a company, contact GiLabs sales: https://wa.me/6289607700028
func RegisterCompanyRoutes(rg *gin.RouterGroup, h *handler.CompanyHandler) {
	g := rg.Group("/companies")
	g.GET("", middleware.RequirePermission("company.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("company.read"), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission("company.update"), h.Update)
	// Approval workflow endpoints
	g.POST("/:id/submit", middleware.RequirePermission("company.update"), h.SubmitForApproval)
}
