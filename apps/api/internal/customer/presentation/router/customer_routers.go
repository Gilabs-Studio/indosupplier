package router

import (
	"github.com/gilabs/gims/api/internal/customer/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCustomerRoutes registers customer routes
func RegisterCustomerRoutes(rg *gin.RouterGroup, h *handler.CustomerHandler) {
	g := rg.Group("/customers")
	{
		// Form data must be before /:id to avoid route conflict
		g.GET("/form-data", h.GetFormData)

		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)

		// Nested bank accounts
		g.POST("/:id/bank-accounts", h.AddBankAccount)
		g.PUT("/:id/bank-accounts/:bankId", h.UpdateBankAccount)
		g.DELETE("/:id/bank-accounts/:bankId", h.DeleteBankAccount)
	}
}
