package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleMenuAccess stores menu-level access assignments for a role.
type RoleMenuAccess struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RoleID    string         `gorm:"type:uuid;not null;index:idx_role_menu_access_role,priority:1;index:idx_role_menu_access_unique,priority:1,unique" json:"role_id"`
	MenuID    string         `gorm:"type:uuid;not null;index:idx_role_menu_access_menu,priority:1;index:idx_role_menu_access_unique,priority:2,unique" json:"menu_id"`
	Scope     string         `gorm:"type:varchar(20);not null;default:'ALL'" json:"scope"`
	IsEnabled bool           `gorm:"type:boolean;not null;default:true" json:"is_enabled"`
	TenantID  *string        `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RoleMenuAccess) TableName() string {
	return "role_menu_access"
}

func (r *RoleMenuAccess) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Scope == "" {
		r.Scope = ScopeAll
	}
	return nil
}
