package router

import (
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/stock_opname/presentation/handler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionService interface matches what middleware needs
type PermissionService interface {
	GetPermissions(roleCode string) ([]string, error)
	GetPermissionsWithScope(roleCode string) (map[string]string, error)
}

func requireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, exists := c.Get("user_permissions")
		if !exists {
			errors.ForbiddenResponse(c, "permission check failed", nil)
			c.Abort()
			return
		}

		permMap, ok := perms.(map[string]bool)
		if !ok {
			errors.ForbiddenResponse(c, "permission format error", nil)
			c.Abort()
			return
		}

		for _, permission := range permissions {
			if permMap[permission] {
				c.Next()
				return
			}
		}

		errors.ForbiddenResponse(c, "Missing one of permissions: "+strings.Join(permissions, ", "), nil)
		c.Abort()
	}
}

func RegisterStockOpnameRoutes(
	r *gin.RouterGroup,
	h *handler.StockOpnameHandler,
	jwtManager *jwt.JWTManager,
	permService PermissionService,
	db *gorm.DB,
) {
	group := r.Group("/stock-opnames")
	group.Use(middleware.AuthMiddleware(jwtManager, permService))
	group.Use(middleware.ScopeMiddleware(db))

	group.GET("", middleware.RequirePermission("stock_opname.read"), h.List)
	group.POST("", middleware.RequirePermission("stock_opname.create"), h.Create)
	group.GET("/:id", middleware.RequirePermission("stock_opname.read"), h.GetByID)
	group.PUT("/:id", middleware.RequirePermission("stock_opname.update"), h.Update)
	group.DELETE("/:id", middleware.RequirePermission("stock_opname.delete"), h.Delete)
	
	// Items
	group.GET("/:id/items", middleware.RequirePermission("stock_opname.read"), h.ListItems)
	group.PUT("/:id/items", middleware.RequirePermission("stock_opname.update"), h.SaveItems)

	// Status
	group.POST("/:id/status", requireAnyPermission("stock_opname.update", "stock_opname.approve", "stock_opname.reject", "stock_opname.post"), h.UpdateStatus)

	// Inventory namespace aliases for phase inventory workflow
	inventoryOpname := r.Group("/inventory/opname")
	inventoryOpname.Use(middleware.AuthMiddleware(jwtManager, permService))
	inventoryOpname.Use(middleware.ScopeMiddleware(db))

	inventoryOpname.GET("", middleware.RequirePermission("stock_opname.read"), h.List)
	inventoryOpname.POST("", middleware.RequirePermission("stock_opname.create"), h.Create)
	// Static route must be declared before parameterized route to avoid /:id capturing "my-warehouses".
	inventoryOpname.GET("/my-warehouses", middleware.RequirePermission("stock_opname.read"), h.GetMyWarehouses)
	inventoryOpname.GET("/:id", middleware.RequirePermission("stock_opname.read"), h.GetByID)
	inventoryOpname.PUT("/:id", middleware.RequirePermission("stock_opname.update"), h.Update)

	// Items aliases
	inventoryOpname.GET("/:id/items", middleware.RequirePermission("stock_opname.read"), h.ListItems)
	inventoryOpname.PUT("/:id/items", middleware.RequirePermission("stock_opname.update"), h.SaveItems)
	// Backward-compatible path introduced during initial phase rollout
	inventoryOpname.PUT("/:id/lines", middleware.RequirePermission("stock_opname.update"), h.SaveLines)
	inventoryOpname.POST("/:id/generate-journal", middleware.RequirePermission("stock_opname.update"), h.GenerateJournal)
	inventoryOpname.POST("/:id/submit-approval", middleware.RequirePermission("stock_opname.update"), h.SubmitApproval)
	inventoryOpname.DELETE("/:id", middleware.RequirePermission("stock_opname.delete"), h.Delete)
}
