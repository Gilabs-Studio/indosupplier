package models

import (
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DealStatus represents the current status of a deal
type DealStatus string

const (
	DealStatusOpen DealStatus = "open"
	DealStatusWon  DealStatus = "won"
	DealStatusLost DealStatus = "lost"
)

// Deal represents a sales opportunity in the CRM pipeline
type Deal struct {
	ID          string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string     `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_deals_tenant_code" json:"code"`
	Title       string     `gorm:"type:varchar(200);not null;index" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Status      DealStatus `gorm:"type:varchar(20);default:'open';index" json:"status"`

	// Pipeline
	PipelineStageID string         `gorm:"type:uuid;not null;index" json:"pipeline_stage_id"`
	PipelineStage   *PipelineStage `gorm:"foreignKey:PipelineStageID" json:"pipeline_stage,omitempty"`

	// Value & Probability
	Value             float64    `gorm:"type:decimal(15,2);default:0" json:"value"`
	Probability       int        `gorm:"type:int;default:0" json:"probability"`
	ExpectedCloseDate *time.Time `gorm:"type:date" json:"expected_close_date"`
	ActualCloseDate   *time.Time `gorm:"type:date" json:"actual_close_date"`
	CloseReason       string     `gorm:"type:text" json:"close_reason"`

	// Relationships
	CustomerID           *string                  `gorm:"type:uuid;index" json:"customer_id"`
	Customer             *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContactID            *string                  `gorm:"type:uuid;index" json:"contact_id"`
	Contact              *Contact                 `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	AssignedTo           *string                  `gorm:"type:uuid;index" json:"assigned_to"`
	AssignedEmployee     *orgModels.Employee      `gorm:"foreignKey:AssignedTo" json:"assigned_employee,omitempty"`
	LeadID               *string                  `gorm:"type:uuid;index" json:"lead_id"`
	Lead                 *Lead                    `gorm:"foreignKey:LeadID;constraint:false" json:"lead,omitempty"`

	// BANT (inherited from lead on conversion)
	BudgetConfirmed bool    `gorm:"default:false" json:"budget_confirmed"`
	BudgetAmount    float64 `gorm:"type:decimal(15,2);default:0" json:"budget_amount"`
	AuthConfirmed   bool    `gorm:"default:false" json:"auth_confirmed"`
	AuthPerson      string  `gorm:"type:varchar(200)" json:"auth_person"`
	NeedConfirmed   bool    `gorm:"default:false" json:"need_confirmed"`
	NeedDescription string  `gorm:"type:text" json:"need_description"`
	TimeConfirmed   bool    `gorm:"default:false" json:"time_confirmed"`

	// Conversion tracking (Deal → Sales Quotation)
	ConvertedToQuotationID *string    `gorm:"type:uuid;index" json:"converted_to_quotation_id"`
	ConvertedAt            *time.Time `json:"converted_at"`

	// Metadata
	Notes     string         `gorm:"type:text" json:"notes"`
	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Items   []DealProductItem `gorm:"foreignKey:DealID" json:"items,omitempty"`
	History []DealHistory     `gorm:"foreignKey:DealID" json:"history,omitempty"`
	Tasks   []Task            `gorm:"foreignKey:DealID" json:"tasks,omitempty"`
}

func (Deal) TableName() string {
	return "crm_deals"
}

func (d *Deal) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	if d.Code == "" {
		d.Code = generateDealCode(tx)
	}
	return nil
}

// generateDealCode creates auto-generated code: DEAL-YYYYMM-XXXXX
func generateDealCode(tx *gorm.DB) string {
	now := apptime.Now()
	prefix := fmt.Sprintf("DEAL-%s", now.Format("200601"))

	var count int64

	tenantID := ""
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			tenantID = tid
		}
	}

	tx.Model(&Deal{}).
		Where("code LIKE ?", prefix+"%").
		Where("COALESCE(tenant_id::text, '__global__') = COALESCE(?, '__global__')", tenantID).
		Count(&count)

	return fmt.Sprintf("%s-%05d", prefix, count+1)
}

// CalculateValue computes the total value from product items
func (d *Deal) CalculateValue() float64 {
	total := 0.0
	for _, item := range d.Items {
		total += item.Subtotal
	}
	return total
}

// DealProductItem represents a product line item within a deal
type DealProductItem struct {
	ID              string                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DealID          string                 `gorm:"type:uuid;not null;index" json:"deal_id"`
	TenantID        string                 `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	ProductID       *string                `gorm:"type:uuid;index" json:"product_id"`
	Product         *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductName     string                 `gorm:"type:varchar(200);not null" json:"product_name"`
	ProductSKU      string                 `gorm:"type:varchar(50)" json:"product_sku"`
	UnitPrice       float64                `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	Quantity        int                    `gorm:"type:int;default:1" json:"quantity"`
	DiscountPercent float64                `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	DiscountAmount  float64                `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	Subtotal        float64                `gorm:"type:decimal(15,2);default:0" json:"subtotal"`
	Notes           string                 `gorm:"type:text" json:"notes"`
	InterestLevel   int                    `gorm:"type:int;default:0" json:"interest_level"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (DealProductItem) TableName() string {
	return "crm_deal_product_items"
}

func (i *DealProductItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// CalculateSubtotal computes subtotal: (UnitPrice * Quantity) - DiscountAmount
func (i *DealProductItem) CalculateSubtotal() float64 {
	gross := i.UnitPrice * float64(i.Quantity)
	if i.DiscountPercent > 0 {
		i.DiscountAmount = gross * (i.DiscountPercent / 100)
	}
	return gross - i.DiscountAmount
}

// DealHistory records stage transitions for audit trail
type DealHistory struct {
	ID                string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DealID            string              `gorm:"type:uuid;not null;index" json:"deal_id"`
	TenantID          string              `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	FromStageID       *string             `gorm:"type:uuid" json:"from_stage_id"`
	FromStage         *PipelineStage      `gorm:"foreignKey:FromStageID" json:"from_stage,omitempty"`
	FromStageName     string              `gorm:"type:varchar(100)" json:"from_stage_name"`
	ToStageID         string              `gorm:"type:uuid;not null" json:"to_stage_id"`
	ToStage           *PipelineStage      `gorm:"foreignKey:ToStageID" json:"to_stage,omitempty"`
	ToStageName       string              `gorm:"type:varchar(100)" json:"to_stage_name"`
	FromProbability   int                 `gorm:"type:int;default:0" json:"from_probability"`
	ToProbability     int                 `gorm:"type:int;default:0" json:"to_probability"`
	DaysInPrevStage   int                 `gorm:"type:int;default:0" json:"days_in_prev_stage"`
	ChangedBy         *string             `gorm:"type:uuid" json:"changed_by"`
	ChangedByEmployee *orgModels.Employee `gorm:"foreignKey:ChangedBy" json:"changed_by_employee,omitempty"`
	ChangedAt         time.Time           `json:"changed_at"`
	Reason            string              `gorm:"type:text" json:"reason"`
	Notes             string              `gorm:"type:text" json:"notes"`
}

func (DealHistory) TableName() string {
	return "crm_deal_history"
}

func (h *DealHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	if h.ChangedAt.IsZero() {
		h.ChangedAt = apptime.Now()
	}
	return nil
}
