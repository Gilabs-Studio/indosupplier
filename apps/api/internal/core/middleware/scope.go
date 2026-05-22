package middleware

import (
	"context"
	"log"
	"time"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ScopeContext holds resolved user context for data-level authorization
type ScopeContext struct {
	UserID       string
	EmployeeID   string
	DivisionID   string
	AreaIDs      []string
	WarehouseIDs []string
	OutletIDs    []string
}

// GetScopeContext extracts the ScopeContext from Gin context
func GetScopeContext(c *gin.Context) *ScopeContext {
	if val, exists := c.Get("scope_context"); exists {
		if sc, ok := val.(*ScopeContext); ok {
			return sc
		}
	}
	return nil
}

// ScopeMiddleware resolves the current user's employee, division, and areas
// for downstream scope-based data filtering. It fetches once per request
// and stores the result in context.
func ScopeMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.Next()
			return
		}

		// Admin also gets their employee data resolved, but their scoping checks
		// are bypassed downstream in permission.go. We just need their ID.

		// Resolve employee data for the current user
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		type employeeRow struct {
			ID         string
			DivisionID *string
		}
		var emp employeeRow
		err := db.WithContext(ctx).
			Table("employees").
			Select("id, division_id").
			Where("user_id = ? AND deleted_at IS NULL", userIDStr).
			First(&emp).Error

		sc := &ScopeContext{
			UserID: userIDStr,
		}

		if err == nil {
			sc.EmployeeID = emp.ID
			if emp.DivisionID != nil {
				sc.DivisionID = *emp.DivisionID
			}

			// Resolve area IDs from employee_areas
			var areaIDs []string
			if err := db.WithContext(ctx).
				Table("employee_areas").
				Where("employee_id = ?", emp.ID).
				Pluck("area_id", &areaIDs).Error; err == nil {
				sc.AreaIDs = areaIDs
			}

			// Resolve outlet IDs from employee_outlets (explicit employee->outlet assignments)
			var outletIDs []string
			if err := db.WithContext(ctx).
				Table("employee_outlets").
				Where("employee_id = ? AND deleted_at IS NULL", emp.ID).
				Pluck("outlet_id", &outletIDs).Error; err == nil {
				sc.OutletIDs = outletIDs
			}

			// Resolve warehouse IDs from employee_warehouses so employee assignment
			// drives POS/inventory scope consistently across the app.
			var warehouseIDs []string
			if err := db.WithContext(ctx).
				Table("employee_warehouses").
				Where("employee_id = ? AND deleted_at IS NULL", emp.ID).
				Pluck("warehouse_id", &warehouseIDs).Error; err == nil {
				sc.WarehouseIDs = warehouseIDs
			}


			log.Printf(
				"[scope] resolved user_id=%s employee_id=%s division_id=%s area_ids=%d outlet_ids=%d warehouse_ids=%d",
				userIDStr,
				sc.EmployeeID,
				sc.DivisionID,
				len(sc.AreaIDs),
				len(sc.OutletIDs),
				len(sc.WarehouseIDs),
			)
		} else if err != gorm.ErrRecordNotFound {
			// Non-404 errors should fail gracefully — log but don't die
			errors.InternalServerErrorResponse(c, "failed to resolve user scope")
			c.Abort()
			return
		}
		// If employee not found, user has no employee profile — scope stays minimal
		if err == gorm.ErrRecordNotFound {
			log.Printf("[scope] no-employee user_id=%s", userIDStr)
		}

		c.Set("scope_context", sc)
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "scope_user_id", sc.UserID)
		reqCtx = context.WithValue(reqCtx, "scope_employee_id", sc.EmployeeID)
		reqCtx = context.WithValue(reqCtx, "scope_division_id", sc.DivisionID)
		reqCtx = context.WithValue(reqCtx, "scope_outlet_ids", sc.OutletIDs)
		reqCtx = context.WithValue(reqCtx, "scope_area_ids", sc.AreaIDs)
		reqCtx = context.WithValue(reqCtx, "scope_warehouse_ids", sc.WarehouseIDs)
		c.Request = c.Request.WithContext(reqCtx)
		c.Next()
	}
}
