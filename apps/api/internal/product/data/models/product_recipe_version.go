package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RecipeVersionChangeManual = "MANUAL"
	RecipeVersionChangeClone  = "CLONE"
)

// ProductRecipeVersion stores immutable snapshots of a product recipe.
type ProductRecipeVersion struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	ProductID string   `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	VersionNumber int    `gorm:"not null" json:"version_number"`
	ChangeType    string `gorm:"type:varchar(20);not null;default:'MANUAL'" json:"change_type"`
	Notes         string `gorm:"type:text" json:"notes"`

	SourceVersionID *string               `gorm:"type:uuid;index" json:"source_version_id"`
	SourceVersion   *ProductRecipeVersion `gorm:"foreignKey:SourceVersionID" json:"source_version,omitempty"`

	Items []ProductRecipeVersionItem `gorm:"foreignKey:RecipeVersionID" json:"items,omitempty"`

	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductRecipeVersion) TableName() string {
	return "product_recipe_versions"
}

func (v *ProductRecipeVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
