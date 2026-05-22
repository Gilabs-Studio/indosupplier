package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
)

func listScopedActiveEmployees(ctx context.Context, db *gorm.DB) ([]orgModels.Employee, error) {
	var employees []orgModels.Employee

	query := database.GetDB(ctx, db).
		Model(&orgModels.Employee{}).
		Where("is_active = ?", true)

	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OwnerUserIDColumn:     "user_id",
		OwnerEmployeeIDColumn: "id",
		DivisionJoinSQL:       "division_id = ?",
		AreaJoinSQL:           "id IN (SELECT employee_id FROM employee_areas WHERE area_id IN ? AND deleted_at IS NULL)",
		OutletJoinSQL:         "id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL)",
		WarehouseJoinSQL:      "id IN (SELECT employee_id FROM employee_warehouses WHERE warehouse_id IN ? AND deleted_at IS NULL)",
	})

	if err := query.Order("name ASC").Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}