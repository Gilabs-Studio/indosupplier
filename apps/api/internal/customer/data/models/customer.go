package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	geographic "github.com/gilabs/gims/api/internal/geographic/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Customer represents a customer entity with approval workflow
type Customer struct {
	ID             string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code           string              `gorm:"type:varchar(50);index" json:"code"`
	Name           string              `gorm:"type:varchar(200);not null;index" json:"name"`
	CustomerTypeID *string             `gorm:"type:uuid;index" json:"customer_type_id"`
	CustomerType   *CustomerType       `gorm:"foreignKey:CustomerTypeID" json:"customer_type,omitempty"`
	Address        string              `gorm:"type:text" json:"address"`
	ProvinceID     *string             `gorm:"type:uuid;index" json:"province_id"`
	Province       *geographic.Province `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	CityID         *string             `gorm:"type:uuid;index" json:"city_id"`
	City           *geographic.City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	DistrictID     *string             `gorm:"type:uuid;index" json:"district_id"`
	District       *geographic.District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	VillageID      *string             `gorm:"type:uuid;index" json:"village_id"`
	VillageName    *string             `gorm:"type:varchar(255)" json:"village_name,omitempty"`
	Village        *geographic.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	Email          string              `gorm:"type:varchar(100)" json:"email"`
	Website        string              `gorm:"type:varchar(200)" json:"website"`
	NPWP           string              `gorm:"type:varchar(30)" json:"npwp"`
	ContactPerson  string              `gorm:"type:varchar(100)" json:"contact_person"`
	Notes          string              `gorm:"type:text" json:"notes"`
	// Location coordinates for map display
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude"`
	// Sales defaults — auto-filled to sales documents when this customer is selected
	DefaultBusinessTypeID  *string                    `gorm:"type:uuid;index" json:"default_business_type_id"`
	DefaultBusinessType    *orgModels.BusinessType    `gorm:"foreignKey:DefaultBusinessTypeID" json:"default_business_type,omitempty"`
	DefaultAreaID          *string                    `gorm:"type:uuid;index" json:"default_area_id"`
	DefaultArea            *orgModels.Area            `gorm:"foreignKey:DefaultAreaID" json:"default_area,omitempty"`
	DefaultSalesRepID      *string                    `gorm:"type:uuid;index" json:"default_sales_rep_id"`
	DefaultSalesRep        *orgModels.Employee        `gorm:"foreignKey:DefaultSalesRepID" json:"default_sales_rep,omitempty"`
	DefaultPaymentTermsID  *string                    `gorm:"type:uuid;index" json:"default_payment_terms_id"`
	DefaultPaymentTerms    *coreModels.PaymentTerms   `gorm:"foreignKey:DefaultPaymentTermsID" json:"default_payment_terms,omitempty"`
	DefaultTaxRate         *float64                   `gorm:"type:decimal(5,2)" json:"default_tax_rate"`
	CreditLimit            float64                    `gorm:"type:decimal(15,2);default:0" json:"credit_limit"`
	CreditIsActive         bool                       `gorm:"default:false" json:"credit_is_active"`

	// Approval workflow
	CreatedBy  *string        `gorm:"type:uuid" json:"created_by"`
	IsActive   bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	// Nested relations
	BankAccounts []CustomerBank        `gorm:"foreignKey:CustomerID" json:"bank_accounts,omitempty"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// BeforeCreate hook to generate UUID
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
