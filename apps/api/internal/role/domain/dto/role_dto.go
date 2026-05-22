package dto

import (
	"time"

	permissionDto "github.com/gilabs/gims/api/internal/permission/domain/dto"
)

// RoleResponse represents role response DTO
type RoleResponse struct {
	ID          string                             `json:"id"`
	Name        string                             `json:"name"`
	Code        string                             `json:"code"`
	Description string                             `json:"description"`
	Status      string                             `json:"status"`
	IsProtected bool                               `json:"is_protected"`
	Permissions []permissionDto.PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time                          `json:"created_at"`
	UpdatedAt   time.Time                          `json:"updated_at"`
}

// CreateRoleRequest represents create role request DTO
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=3"`
	Code        string `json:"code" binding:"required,min=3"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
	IsProtected bool   `json:"is_protected"` // Only system can set this to true
}

// UpdateRoleRequest represents update role request DTO
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=3"`
	Code        string `json:"code" binding:"omitempty,min=3"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// AssignPermissionsRequest represents assign permissions to role request DTO
// Supports both full state (legacy) and differential modes (Sprint 21).
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"omitempty,dive,uuid"`
	// Assignments allows scope-aware permission assignment (Sprint 20)
	Assignments []PermissionAssignment `json:"assignments" binding:"omitempty,dive"`
	// Diff support: Instead of full state, send only changed IDs (Sprint 21)
	// When mode=diff, must provide Diff with Added/Removed arrays
	Mode string `json:"mode" binding:"omitempty,oneof=full diff"`
	Diff *DiffPermissionsRequest `json:"diff" binding:"omitempty"`
}

// PermissionAssignment represents a single permission-scope pair for assignment
type PermissionAssignment struct {
	PermissionID string `json:"permission_id" binding:"required,uuid"`
	Scope        string `json:"scope" binding:"required,oneof=OWN DIVISION AREA WAREHOUSE OUTLET ALL"`
}

// RoleMenuAccessRequest represents menu-level access assignment request.
type RoleMenuAccessRequest struct {
	Assignments []RoleMenuAccessAssignment `json:"assignments" binding:"omitempty,dive"`
}

// RoleMenuAccessAssignment represents one role-menu assignment.
type RoleMenuAccessAssignment struct {
	MenuID string `json:"menu_id" binding:"required,uuid"`
	Scope  string `json:"scope" binding:"required,oneof=OWN DIVISION AREA WAREHOUSE OUTLET ALL"`
}

// RoleMenuAccessResponse represents menu-level access entry.
type RoleMenuAccessResponse struct {
	MenuID    string `json:"menu_id"`
	Scope     string `json:"scope"`
	IsEnabled bool   `json:"is_enabled"`
}

// DiffPermissionsRequest represents differential permission update (only changed IDs).
// Instead of sending the full current state, send only added and removed permission IDs.
// This allows partial updates without triggering full validation on unchanged permissions.
type DiffPermissionsRequest struct {
	// Added permission IDs to assign (with optional scope, defaults to ALL if omitted)
	Added []string `json:"added" binding:"omitempty,dive,uuid"`
	// Added permissions with explicit scope-aware assignment
	AddedWithScope []PermissionAssignment `json:"added_with_scope" binding:"omitempty,dive"`
	// Removed permission IDs to unassign
	Removed []string `json:"removed" binding:"omitempty,dive,uuid"`
}
