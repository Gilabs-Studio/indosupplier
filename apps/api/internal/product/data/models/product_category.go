package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryType defines the type of a product category
const (
	CategoryTypeGoods = "GOODS"
	CategoryTypeFnB   = "FNB"
)

// ProductCategory represents a category for products (supports hierarchy)
type ProductCategory struct {
	ID           string            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name         string            `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description  string            `gorm:"type:text" json:"description"`
	CategoryType string            `gorm:"type:varchar(20);not null;default:'GOODS'" json:"category_type"`
	ParentID     *string           `gorm:"type:uuid;index" json:"parent_id"`
	Parent       *ProductCategory  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []ProductCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	IsActive     bool              `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `gorm:"index" json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `gorm:"index" json:"-"`
}

// TableName specifies the table name for ProductCategory
func (ProductCategory) TableName() string {
	return "product_categories"
}

// BeforeCreate hook to generate UUID
func (p *ProductCategory) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
