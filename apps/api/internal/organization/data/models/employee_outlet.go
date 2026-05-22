package models

import (
	"time"

	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmployeeOutlet represents the M:N relationship between Employee and Outlet.
// Associates an employee with outlets (e.g., for POS or operational purposes).
type EmployeeOutlet struct {
	ID        string   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string   `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID string   `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_outlet" json:"employee_id"`
	Employee  *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	OutletID  string   `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_outlet" json:"outlet_id"`
	Outlet    *Outlet  `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for EmployeeOutlet
func (EmployeeOutlet) TableName() string {
	return "employee_outlets"
}

// BeforeCreate hook to generate UUID
func (eo *EmployeeOutlet) BeforeCreate(tx *gorm.DB) error {
	if eo.ID == "" {
		eo.ID = uuid.New().String()
	}
	return nil
}

// EmployeeWarehouse represents the M:N relationship between Employee and Warehouse.
// Associates an employee with warehouses for inventory and operational scope.
// When an employee is assigned to outlets, their warehouses are auto-selected based on outlet.warehouse_id.
type EmployeeWarehouse struct {
	ID          string   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string   `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID  string   `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_warehouse" json:"employee_id"`
	Employee    *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	WarehouseID string   `gorm:"type:uuid;not null;index;uniqueIndex:idx_employee_warehouse" json:"warehouse_id"`
Warehouse   *warehouseModels.Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	// IsAuto=true means this warehouse assignment was auto-created from outlet assignment
	IsAuto    bool           `gorm:"default:true;index" json:"is_auto"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for EmployeeWarehouse
func (EmployeeWarehouse) TableName() string {
	return "employee_warehouses"
}

// BeforeCreate hook to generate UUID
func (ew *EmployeeWarehouse) BeforeCreate(tx *gorm.DB) error {
	if ew.ID == "" {
		ew.ID = uuid.New().String()
	}
	return nil
}
