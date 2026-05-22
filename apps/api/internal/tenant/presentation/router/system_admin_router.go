package router

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterSystemAdminRoutes registers all system admin routes under /internal.
// These routes are completely separate from tenant API routes.
func RegisterSystemAdminRoutes(r *gin.Engine, h *handler.SystemAdminHandler, jwtManager *jwt.JWTManager, sysAdminRepo repositories.SystemAdminRepository) {
	internal := r.Group("/internal")
	{
		// Public system admin auth (rate-limited)
		internal.POST("/sys-login", middleware.RateLimitMiddleware("login"), h.Login)

		// Protected system admin routes
		protected := internal.Group("")
		protected.Use(middleware.SystemAdminAuthMiddleware(jwtManager, sysAdminRepo))
		{
			protected.GET("/me", h.Me)
			protected.PATCH("/me", h.UpdateProfile)
			protected.PATCH("/me/password", h.ChangePassword)
			protected.POST("/sys-logout", h.Logout)
			protected.GET("/system-dashboard", h.Dashboard)

			// Tenant management
			protected.GET("/tenants", h.ListTenants)
			protected.GET("/tenants/:id", h.GetTenantDetail)
			protected.POST("/tenants/:id/recover-deletion", h.RecoverTenantDeletion)
			protected.GET("/tenants/:id/subscription", h.GetTenantSubscription)

			// Coupon management
			protected.POST("/coupons", h.CreateCoupon)
			protected.GET("/coupons", h.ListCoupons)
			protected.PUT("/coupons/:id", h.UpdateCoupon)
			protected.PATCH("/coupons/:id/status", h.SetCouponStatus)

			// Subscription management
			protected.GET("/subscriptions", h.ListSubscriptions)

			// Subscription plan config management
			protected.GET("/plans", h.ListPlans)
			protected.PUT("/plans", h.UpsertPlan)
			protected.PATCH("/plans/:slug/status", h.SetPlanActive)
			protected.PUT("/plans/:slug/entitlements", h.SyncPlanEntitlements)

			// Permission management
			protected.GET("/permissions", h.ListPermissions)
			protected.GET("/permissions/menus", h.ListPermissionMenus)
			protected.GET("/permissions/:id", h.GetPermission)
			protected.POST("/permissions", h.CreatePermission)
			protected.PUT("/permissions/:id", h.UpdatePermission)
			protected.DELETE("/permissions/:id", h.DeletePermission)
		}
	}
}
