package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductRecipeItem defines a single ingredient line in a product's recipe.
// A product may have multiple recipe items, each pointing to an ingredient product
// (a Product with IsIngredient = true) and the quantity consumed per unit sold.
type ProductRecipeItem struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	// ProductID is the F&B product that owns this recipe (e.g. "Nasi Goreng")
	ProductID string   `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// IngredientProductID is the raw material product (Product.IsIngredient = true)
	IngredientProductID string   `gorm:"type:uuid;not null;index" json:"ingredient_product_id"`
	IngredientProduct   *Product `gorm:"foreignKey:IngredientProductID" json:"ingredient_product,omitempty"`

	// Quantity consumed per one unit of the parent product sold
	Quantity float64 `gorm:"type:decimal(15,4);not null;default:0" json:"quantity"`

	// UomID is the unit of measure for the consumed quantity (may differ from ingredient's base UOM)
	UomID *string        `gorm:"type:uuid;index" json:"uom_id"`
	Uom   *UnitOfMeasure `gorm:"foreignKey:UomID" json:"uom,omitempty"`

	Notes string `gorm:"type:text" json:"notes"`

	SortOrder int `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for ProductRecipeItem
func (ProductRecipeItem) TableName() string {
	return "product_recipe_items"
}

// BeforeCreate hook to generate UUID
func (p *ProductRecipeItem) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
