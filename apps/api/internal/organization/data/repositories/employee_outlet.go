package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmployeeOutletRepository defines the interface for employee outlet assignment operations.
type EmployeeOutletRepository interface {
	// AssignOutlets assigns outlets to an employee.
	AssignOutlets(ctx context.Context, employeeID string, outletIDs []string) error
	// RemoveOutlets removes specific outlet assignments for an employee.
	RemoveOutlets(ctx context.Context, employeeID string, outletIDs []string) error
	// RemoveAllOutlets removes all outlet assignments for an employee.
	RemoveAllOutlets(ctx context.Context, employeeID string) error
	// GetByEmployeeID returns all outlet assignments for an employee.
	GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeOutlet, error)
	// GetByOutletID returns all employee assignments for an outlet.
	GetByOutletID(ctx context.Context, outletID string) ([]models.EmployeeOutlet, error)
}

type employeeOutletRepository struct {
	db *gorm.DB
}

// NewEmployeeOutletRepository creates a new EmployeeOutletRepository instance.
func NewEmployeeOutletRepository(db *gorm.DB) EmployeeOutletRepository {
	return &employeeOutletRepository{db: db}
}

func (r *employeeOutletRepository) AssignOutlets(ctx context.Context, employeeID string, outletIDs []string) error {
	if len(outletIDs) == 0 {
		return nil
	}

	employeeOutlets := make([]models.EmployeeOutlet, 0, len(outletIDs))
	tenantID := middleware.TenantFromContext(ctx)
	for _, outletID := range outletIDs {
		employeeOutlets = append(employeeOutlets, models.EmployeeOutlet{
			TenantID:   tenantID,
			EmployeeID: employeeID,
			OutletID:   outletID,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&employeeOutlets).Error
}

func (r *employeeOutletRepository) RemoveOutlets(ctx context.Context, employeeID string, outletIDs []string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ? AND outlet_id IN ?", employeeID, outletIDs).
		Delete(&models.EmployeeOutlet{}).Error
}

func (r *employeeOutletRepository) RemoveAllOutlets(ctx context.Context, employeeID string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Delete(&models.EmployeeOutlet{}).Error
}

func (r *employeeOutletRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeOutlet, error) {
	var outlets []models.EmployeeOutlet
	return outlets, database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Find(&outlets).Error
}

func (r *employeeOutletRepository) GetByOutletID(ctx context.Context, outletID string) ([]models.EmployeeOutlet, error) {
	var outlets []models.EmployeeOutlet
	return outlets, database.GetDB(ctx, r.db).
		Where("outlet_id = ?", outletID).
		Find(&outlets).Error
}

// EmployeeWarehouseRepository defines the interface for employee warehouse assignment operations.
type EmployeeWarehouseRepository interface {
	// AssignWarehouses assigns warehouses to an employee.
	AssignWarehouses(ctx context.Context, employeeID string, warehouseIDs []string) error
	// AssignWarehousesWithAutoFlag assigns warehouses with an auto flag (used for outlet-based auto-selection).
	AssignWarehousesWithAutoFlag(ctx context.Context, employeeID string, warehouseIDs []string, isAuto bool) error
	// RemoveWarehouses removes specific warehouse assignments for an employee.
	RemoveWarehouses(ctx context.Context, employeeID string, warehouseIDs []string) error
	// RemoveAllWarehouses removes all warehouse assignments for an employee.
	RemoveAllWarehouses(ctx context.Context, employeeID string) error
	// RemoveAutoWarehouses removes only auto-created warehouse assignments for an employee.
	RemoveAutoWarehouses(ctx context.Context, employeeID string) error
	// GetByEmployeeID returns all warehouse assignments for an employee.
	GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeWarehouse, error)
	// GetByWarehouseID returns all employee assignments for a warehouse.
	GetByWarehouseID(ctx context.Context, warehouseID string) ([]models.EmployeeWarehouse, error)
	// GetAutoWarehouses returns only auto-created warehouse assignments for an employee.
	GetAutoWarehouses(ctx context.Context, employeeID string) ([]models.EmployeeWarehouse, error)
}

type employeeWarehouseRepository struct {
	db *gorm.DB
}

// NewEmployeeWarehouseRepository creates a new EmployeeWarehouseRepository instance.
func NewEmployeeWarehouseRepository(db *gorm.DB) EmployeeWarehouseRepository {
	return &employeeWarehouseRepository{db: db}
}

func (r *employeeWarehouseRepository) AssignWarehouses(ctx context.Context, employeeID string, warehouseIDs []string) error {
	return r.AssignWarehousesWithAutoFlag(ctx, employeeID, warehouseIDs, false)
}

func (r *employeeWarehouseRepository) AssignWarehousesWithAutoFlag(ctx context.Context, employeeID string, warehouseIDs []string, isAuto bool) error {
	if len(warehouseIDs) == 0 {
		return nil
	}

	employeeWarehouses := make([]models.EmployeeWarehouse, 0, len(warehouseIDs))
	tenantID := middleware.TenantFromContext(ctx)
	for _, warehouseID := range warehouseIDs {
		employeeWarehouses = append(employeeWarehouses, models.EmployeeWarehouse{
			TenantID:    tenantID,
			EmployeeID:  employeeID,
			WarehouseID: warehouseID,
			IsAuto:      isAuto,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&employeeWarehouses).Error
}

func (r *employeeWarehouseRepository) RemoveWarehouses(ctx context.Context, employeeID string, warehouseIDs []string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ? AND warehouse_id IN ?", employeeID, warehouseIDs).
		Delete(&models.EmployeeWarehouse{}).Error
}

func (r *employeeWarehouseRepository) RemoveAllWarehouses(ctx context.Context, employeeID string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Delete(&models.EmployeeWarehouse{}).Error
}

func (r *employeeWarehouseRepository) RemoveAutoWarehouses(ctx context.Context, employeeID string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ? AND is_auto = true", employeeID).
		Delete(&models.EmployeeWarehouse{}).Error
}

func (r *employeeWarehouseRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeWarehouse, error) {
	var warehouses []models.EmployeeWarehouse
	return warehouses, database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Find(&warehouses).Error
}

func (r *employeeWarehouseRepository) GetByWarehouseID(ctx context.Context, warehouseID string) ([]models.EmployeeWarehouse, error) {
	var warehouses []models.EmployeeWarehouse
	return warehouses, database.GetDB(ctx, r.db).
		Where("warehouse_id = ?", warehouseID).
		Find(&warehouses).Error
}

func (r *employeeWarehouseRepository) GetAutoWarehouses(ctx context.Context, employeeID string) ([]models.EmployeeWarehouse, error) {
	var warehouses []models.EmployeeWarehouse
	return warehouses, database.GetDB(ctx, r.db).
		Where("employee_id = ? AND is_auto = true", employeeID).
		Find(&warehouses).Error
}
