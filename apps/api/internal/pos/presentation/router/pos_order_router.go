package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/pos/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPOSOrderRoutes registers POS order and catalog routes
func RegisterPOSOrderRoutes(rg *gin.RouterGroup, h *handler.POSOrderHandler, receiptH *handler.POSReceiptHandler) {
	// POS catalog endpoint (read-only, per outlet)
	rg.GET("/catalog/outlet/:outletID", middleware.RequirePermission("pos.order.create"), h.GetCatalog)
	// POS available outlets (filtered by scope)
	rg.GET("/outlets", middleware.RequirePermission("pos.order.read"), h.ListOutlets)

	orders := rg.Group("/orders")
	orders.Use(middleware.RequirePermission("pos.order.create"))

	// Static routes before parameterized
	orders.POST("", middleware.IdempotentRequest(), h.Create)
	orders.GET("", middleware.RequirePermission("pos.order.read"), h.List)
	orders.GET("/:id", h.GetByID)
	orders.POST("/:id/confirm", middleware.IdempotentRequest(), h.Confirm)
	orders.POST("/:id/void", h.Void)
	orders.POST("/:id/assign-table", h.AssignTable)
	orders.POST("/:id/serve", h.Serve)
	orders.POST("/:id/complete", h.Complete)

	// HTML receipt — requires pos.order.read; returns text/html for thermal printing
	orders.GET("/:id/receipt", middleware.RequirePermission("pos.order.read"), receiptH.GetReceipt)

	// Order item management
	orders.POST("/:id/items", middleware.IdempotentRequest(), h.AddItem)
	orders.PUT("/:id/items/:itemID", h.UpdateItem)
	orders.DELETE("/:id/items/:itemID", h.RemoveItem)
	// Per-item serve — place before the generic item routes to avoid routing ambiguity
	orders.POST("/:id/items/:itemID/serve", h.ServeItem)
}
