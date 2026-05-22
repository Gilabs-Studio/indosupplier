package models

import (
	"time"

	roleModels "github.com/gilabs/gims/api/internal/role/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity
type User struct {
	ID                   string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID             string           `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Email                string           `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password             string           `gorm:"type:varchar(255);not null" json:"-"` // Hidden from JSON
	Name                 string           `gorm:"type:varchar(255);not null;index" json:"name"`
	AvatarURL            string           `gorm:"type:text" json:"avatar_url"`
	RoleID               string           `gorm:"type:uuid;not null;index" json:"role_id"`
	Role                 *roleModels.Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Status               string           `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	PasswordResetPending bool             `gorm:"type:boolean;not null;default:false;index" json:"password_reset_pending"`
	CreatedAt            time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time        `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt            gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
