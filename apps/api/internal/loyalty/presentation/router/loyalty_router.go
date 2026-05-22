package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/loyalty/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterLoyaltyRoutes mounts all protected loyalty routes.
// Note: static sub-paths (lookup, enroll, by-customer) must precede /:id.
func RegisterLoyaltyRoutes(group *gin.RouterGroup, h *handler.LoyaltyHandler) {
	loyalty := group.Group("/loyalty")

	// Program management
	programs := loyalty.Group("/programs")
	programs.GET("", middleware.RequirePermission("loyalty.read"), h.ListPrograms)
	programs.POST("", middleware.RequirePermission("loyalty.create"), h.CreateProgram)
	programs.GET("/:id", middleware.RequirePermission("loyalty.read"), h.GetProgram)
	programs.PUT("/:id", middleware.RequirePermission("loyalty.update"), h.UpdateProgram)
	programs.DELETE("/:id", middleware.RequirePermission("loyalty.delete"), h.DeleteProgram)
	programs.PATCH("/:id/toggle", middleware.RequirePermission("loyalty.update"), h.ToggleProgramActive)

	// Member management — static routes before /:id
	members := loyalty.Group("/members")
	members.GET("", middleware.RequirePermission("loyalty.read"), h.ListMembers)
	members.POST("/enroll", middleware.RequirePermission("loyalty.create"), h.EnrollMember)
	members.GET("/lookup", middleware.RequirePermission("loyalty.read"), h.LookupMember)
	members.GET("/by-customer/:customer_id", middleware.RequirePermission("loyalty.read"), h.GetMemberByCustomer)
	members.GET("/:id", middleware.RequirePermission("loyalty.read"), h.GetMember)
	members.PUT("/:id/program", middleware.RequirePermission("loyalty.update"), h.ChangeProgram)

	// Points ledger
	loyalty.GET("/ledger/:member_id", middleware.RequirePermission("loyalty.read"), h.ListLedger)
	loyalty.POST("/redeem", middleware.RequirePermission("loyalty.create"), h.RedeemPoints)
	loyalty.POST("/adjust", middleware.RequirePermission("loyalty.update"), h.AdjustPoints)
}

// RegisterLoyaltyPublicRoutes mounts unauthenticated public loyalty routes.
func RegisterLoyaltyPublicRoutes(group *gin.RouterGroup, h *handler.LoyaltyHandler) {
	pub := group.Group("/public/loyalty")
	pub.POST("/register", h.PublicSelfRegister)
}
