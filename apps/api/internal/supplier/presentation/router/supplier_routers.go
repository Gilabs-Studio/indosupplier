package router

import (
	"github.com/gilabs/gims/api/internal/supplier/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSupplierRoutes registers supplier routes
func RegisterSupplierRoutes(rg *gin.RouterGroup, h *handler.SupplierHandler) {
	g := rg.Group("/suppliers")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.GetByID)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
		
		// Approval workflow
		g.POST("/:id/submit", h.Submit)
		g.POST("/:id/approve", h.Approve)
		
		// Nested contacts
		g.POST("/:id/contacts", h.AddContact)
		g.PUT("/:id/contacts/:contactId", h.UpdateContact)
		g.DELETE("/:id/contacts/:contactId", h.DeleteContact)
		
		// Nested bank accounts
		g.POST("/:id/bank-accounts", h.AddBankAccount)
		g.PUT("/:id/bank-accounts/:bankId", h.UpdateBankAccount)
		g.DELETE("/:id/bank-accounts/:bankId", h.DeleteBankAccount)
	}
}
