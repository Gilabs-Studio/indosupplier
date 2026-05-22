package models

import (
	permissionModels "github.com/gilabs/gims/api/internal/permission/data/models"
)

// Valid permission scope values
const (
	ScopeOwn       = "OWN"
	ScopeWarehouse = "WAREHOUSE"
	ScopeDivision  = "DIVISION"
	ScopeArea      = "AREA"
	ScopeOutlet    = "OUTLET"
	ScopeAll       = "ALL"
)

// ValidScopes returns all valid scope values
func ValidScopes() []string {
	return []string{ScopeOwn, ScopeWarehouse, ScopeDivision, ScopeArea, ScopeOutlet, ScopeAll}
}

// Valid data scope values for role-level data isolation
const (
	DataScopeAll      = "ALL"
	DataScopeDivision = "DIVISION"
	DataScopeArea     = "AREA"
	DataScopeOutlet   = "OUTLET"
	DataScopeOwn      = "OWN"
)

// ValidDataScopes returns all valid data scope values for roles
func ValidDataScopes() []string {
	return []string{DataScopeAll, DataScopeDivision, DataScopeArea, DataScopeOutlet, DataScopeOwn}
}

// IsValidDataScope checks if a data scope value is valid
func IsValidDataScope(scope string) bool {
	for _, s := range ValidDataScopes() {
		if s == scope {
			return true
		}
	}
	return false
}

// IsValidScope checks if a scope value is valid
func IsValidScope(scope string) bool {
	for _, s := range ValidScopes() {
		if s == scope {
			return true
		}
	}
	return false
}

// RolePermission represents the explicit junction table between Role and Permission with scope
type RolePermission struct {
	RoleID       string                      `gorm:"type:uuid;not null;primaryKey" json:"role_id"`
	PermissionID string                      `gorm:"type:uuid;not null;primaryKey" json:"permission_id"`
	Scope        string                      `gorm:"type:varchar(20);not null;default:'ALL'" json:"scope"`
	Role         *Role                       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission   *permissionModels.Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// TableName specifies the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}
