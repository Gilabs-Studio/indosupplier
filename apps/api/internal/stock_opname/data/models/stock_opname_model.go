package models

import (
	"time"

	employeeModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	warehouse "github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StockOpnameStatus string

const (
	StockOpnameStatusDraft    StockOpnameStatus = "draft"
	StockOpnameStatusPending  StockOpnameStatus = "pending"
	StockOpnameStatusApproved StockOpnameStatus = "approved"
	StockOpnameStatusRejected StockOpnameStatus = "rejected"
	StockOpnameStatusPosted   StockOpnameStatus = "posted"
)

type StockOpname struct {
	ID               string            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID         string            `gorm:"column:tenant_id;type:uuid;uniqueIndex:idx_stock_opnames_tenant_number,priority:1" json:"tenant_id,omitempty"`
	OpnameNumber     string            `gorm:"type:varchar(50);uniqueIndex:idx_stock_opnames_tenant_number,priority:2;not null"`
	WarehouseID      string            `gorm:"type:uuid;not null;index"`
	JournalID        *string           `gorm:"type:uuid;index"`
	Date             time.Time         `gorm:"type:date;not null"`
	Status           StockOpnameStatus `gorm:"type:varchar(20);not null;default:'draft'"`
	Description      string            `gorm:"type:text"`
	TotalItems       int               `gorm:"type:int;default:0"`
	TotalVarianceQty float64           `gorm:"type:decimal(15,2);default:0"`

	// Employee assignment
	OrderedByID  *string `gorm:"type:uuid;index" json:"ordered_by_id,omitempty"`
	AssignedToID *string `gorm:"type:uuid;index" json:"assigned_to_id,omitempty"`

	// Audit
	CreatedBy *string   `gorm:"type:uuid"`
	UpdatedBy *string   `gorm:"type:uuid"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	Items []StockOpnameItem `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE"`

	// Associations (used for joins)
	Warehouse  *warehouse.Warehouse          `gorm:"foreignKey:WarehouseID;references:ID" json:"warehouse,omitempty"`
	OrderedBy  *employeeModels.Employee      `gorm:"foreignKey:OrderedByID;references:ID" json:"ordered_by,omitempty"`
	AssignedTo *employeeModels.Employee      `gorm:"foreignKey:AssignedToID;references:ID" json:"assigned_to,omitempty"`
}

type StockOpnameItem struct {
	ID               string                    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID         string                    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	StockOpnameID    string                    `gorm:"type:uuid;not null;index"`
	Product          *productModels.Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductID        string                    `gorm:"type:uuid;not null;index"`
	InventoryBatchID *string                   `gorm:"type:uuid;index" json:"inventory_batch_id,omitempty"`
	BatchNumber      string                    `gorm:"type:varchar(100);default:''" json:"batch_number,omitempty"`
	BatchQty         float64                   `gorm:"type:decimal(15,2);default:0" json:"batch_qty,omitempty"`
	SystemQty        float64                   `gorm:"type:decimal(15,2);not null;default:0"`
	PhysicalQty      *float64                  `gorm:"type:decimal(15,2)"` // Nullable until counted
	VarianceQty      float64                   `gorm:"type:decimal(15,2);default:0"`
	Notes            string                    `gorm:"type:text"`
    
    // Audit
    CreatedAt     time.Time `gorm:"autoCreateTime"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (StockOpname) TableName() string {
	return "stock_opnames"
}

func (StockOpnameItem) TableName() string {
	return "stock_opname_items"
}

// BeforeCreate hook to generate UUID
func (s *StockOpname) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	if s.TenantID == "" {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			s.TenantID = tid
		}
	}
	return nil
}

// BeforeCreate hook to generate UUID
func (s *StockOpnameItem) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	if s.TenantID == "" {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			s.TenantID = tid
		}
	}
	return nil
}
