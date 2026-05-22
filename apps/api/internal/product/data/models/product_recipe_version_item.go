package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRecipeVersionItem stores item snapshot rows for each recipe version.
type ProductRecipeVersionItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	RecipeVersionID string                `gorm:"type:uuid;not null;index" json:"recipe_version_id"`
	RecipeVersion   *ProductRecipeVersion `gorm:"foreignKey:RecipeVersionID" json:"recipe_version,omitempty"`

	IngredientProductID string   `gorm:"type:uuid;not null;index" json:"ingredient_product_id"`
	IngredientProduct   *Product `gorm:"foreignKey:IngredientProductID" json:"ingredient_product,omitempty"`

	Quantity float64 `gorm:"type:decimal(15,4);not null;default:0" json:"quantity"`

	UomID *string        `gorm:"type:uuid;index" json:"uom_id"`
	Uom   *UnitOfMeasure `gorm:"foreignKey:UomID" json:"uom,omitempty"`

	Notes     string `gorm:"type:text" json:"notes"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductRecipeVersionItem) TableName() string {
	return "product_recipe_version_items"
}

func (v *ProductRecipeVersionItem) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
