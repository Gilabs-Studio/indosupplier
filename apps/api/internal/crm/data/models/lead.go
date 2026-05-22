package models

import (
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Lead represents a sales prospect with BANT qualification scoring
type Lead struct {
	ID          string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_leads_tenant_code" json:"tenant_id,omitempty"`
	Code        string  `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_leads_tenant_code" json:"code"`
	FirstName   string  `gorm:"type:varchar(100);not null;index" json:"first_name"`
	LastName    string  `gorm:"type:varchar(100);index" json:"last_name"`
	CompanyName string  `gorm:"type:varchar(200);index" json:"company_name"`
	Email       string  `gorm:"type:varchar(100)" json:"email"`
	Phone       string  `gorm:"type:varchar(30)" json:"phone"`
	ContactRoleID *string `gorm:"type:uuid;index" json:"contact_role_id"`
	ContactRole   *ContactRole `gorm:"foreignKey:ContactRoleID" json:"contact_role,omitempty"`
	JobTitle      string       `gorm:"type:varchar(100)" json:"job_title"`
	Address       string       `gorm:"type:text" json:"address"`
	City          string       `gorm:"type:varchar(100)" json:"city"`
	Province      string       `gorm:"type:varchar(100)" json:"province"`
	ProvinceID    *string      `gorm:"type:uuid;index" json:"province_id"`
	CityID        *string      `gorm:"type:uuid;index" json:"city_id"`
	DistrictID    *string      `gorm:"type:uuid;index" json:"district_id"`
	VillageName   string       `gorm:"type:varchar(200)" json:"village_name"`

	// Classification
	LeadSourceID *string     `gorm:"type:uuid;index" json:"lead_source_id"`
	LeadSource   *LeadSource `gorm:"foreignKey:LeadSourceID" json:"lead_source,omitempty"`
	LeadStatusID *string     `gorm:"type:uuid;index" json:"lead_status_id"`
	LeadStatus   *LeadStatus `gorm:"foreignKey:LeadStatusID" json:"lead_status,omitempty"`

	// Scoring
	LeadScore      int     `gorm:"type:int;default:0" json:"lead_score"`
	Probability    int     `gorm:"type:int;default:0" json:"probability"`
	EstimatedValue float64 `gorm:"type:decimal(15,2);default:0" json:"estimated_value"`

	// BANT Qualification
	BudgetConfirmed bool       `gorm:"default:false" json:"budget_confirmed"`
	BudgetAmount    float64    `gorm:"type:decimal(15,2);default:0" json:"budget_amount"`
	AuthConfirmed   bool       `gorm:"default:false" json:"auth_confirmed"`
	AuthPerson      string     `gorm:"type:varchar(200)" json:"auth_person"`
	NeedConfirmed   bool       `gorm:"default:false" json:"need_confirmed"`
	NeedDescription string     `gorm:"type:text" json:"need_description"`
	TimeConfirmed   bool       `gorm:"default:false" json:"time_confirmed"`
	TimeExpected    *time.Time `gorm:"type:date" json:"time_expected"`

	// Assignment
	AssignedTo       *string             `gorm:"type:uuid;index" json:"assigned_to"`
	AssignedEmployee *orgModels.Employee `gorm:"foreignKey:AssignedTo" json:"assigned_employee,omitempty"`

	// Conversion (populated after conversion)
	CustomerID  *string                  `gorm:"type:uuid;index" json:"customer_id"`
	Customer    *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContactID   *string                  `gorm:"type:uuid;index" json:"contact_id"`
	Contact     *Contact                 `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	DealID      *string                  `gorm:"type:uuid;index" json:"deal_id"`
	Deal        *Deal                    `gorm:"foreignKey:DealID;constraint:false" json:"deal,omitempty"`
	ConvertedAt *time.Time               `json:"converted_at"`
	ConvertedBy *string                  `gorm:"type:uuid" json:"converted_by"`

	// Sales defaults for customer conversion
	BusinessTypeID *string                  `gorm:"type:uuid;index" json:"business_type_id"`
	BusinessType   *orgModels.BusinessType  `gorm:"foreignKey:BusinessTypeID" json:"business_type,omitempty"`
	AreaID         *string                  `gorm:"type:uuid;index" json:"area_id"`
	Area           *orgModels.Area          `gorm:"foreignKey:AreaID" json:"area,omitempty"`
	PaymentTermsID *string                  `gorm:"type:uuid;index" json:"payment_terms_id"`
	PaymentTerms   *coreModels.PaymentTerms `gorm:"foreignKey:PaymentTermsID" json:"payment_terms,omitempty"`

	// NPWP / Tax ID
	NPWP string `gorm:"type:varchar(30)" json:"npwp"`

	// External Source Data (Google Maps, etc)
	Latitude             *float64 `gorm:"type:numeric(10,8)" json:"latitude"`
	Longitude            *float64 `gorm:"type:numeric(11,8)" json:"longitude"`
	Rating               *float64 `gorm:"type:numeric(3,1)" json:"rating"`
	RatingCount          *int     `gorm:"type:int" json:"rating_count"`
	Types                string   `gorm:"type:text" json:"types"`
	OpeningHours         string   `gorm:"type:text" json:"opening_hours"`
	ThumbnailURL         string   `gorm:"type:text" json:"thumbnail_url"`
	CID                  string   `gorm:"type:varchar(200)" json:"cid"`
	PlaceID              string   `gorm:"type:varchar(200)" json:"place_id"`
	Website              string   `gorm:"type:varchar(255)" json:"website"`
	BankAccountID        *string  `gorm:"type:uuid;index" json:"bank_account_id"`
	BankAccountReference string   `gorm:"type:varchar(255)" json:"bank_account_reference"`

	// Metadata
	Notes            string         `gorm:"type:text" json:"notes"`
	ProcessedFromN8N bool           `gorm:"type:boolean;default:false;index" json:"processed_from_n8n"`
	ProcessedAt      *time.Time     `gorm:"type:timestamptz" json:"processed_at"`
	CreatedBy        *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Activities   []Activity        `gorm:"foreignKey:LeadID" json:"activities,omitempty"`
	Tasks        []Task            `gorm:"foreignKey:LeadID" json:"tasks,omitempty"`
	ProductItems []LeadProductItem `gorm:"foreignKey:LeadID" json:"product_items,omitempty"`
}

func (Lead) TableName() string {
	return "crm_leads"
}

func (l *Lead) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.Code == "" {
		code, err := generateLeadCode(tx, l.TenantID)
		if err != nil {
			return err
		}
		l.Code = code
	}
	return nil
}

// generateLeadCode creates auto-generated code: LEAD-YYYYMM-XXXXX
func generateLeadCode(tx *gorm.DB, tenantID string) (string, error) {
	now := apptime.Now()
	prefix := fmt.Sprintf("LEAD-%s", now.Format("200601"))
	lockKey := fmt.Sprintf("crm_lead_code:%s:%s", tenantID, prefix)

	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
		return "", err
	}

	var maxSeq int64
	query := tx.Unscoped().Model(&Lead{}).Where("code LIKE ?", prefix+"%")
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Select("COALESCE(MAX(CAST(substring(code from '[0-9]+$') AS integer)), 0)").Scan(&maxSeq).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%05d", prefix, maxSeq+1), nil
}

// CalculateLeadScore computes score from BANT fields + status score
func (l *Lead) CalculateLeadScore() int {
	score := 0

	if l.LeadStatus != nil {
		score += l.LeadStatus.Score
	}

	if l.BudgetConfirmed {
		score += 15
	}
	if l.AuthConfirmed {
		score += 15
	}
	if l.NeedConfirmed {
		score += 15
	}
	if l.TimeConfirmed {
		score += 15
	}

	if score > 100 {
		score = 100
	}

	return score
}

// IsConverted returns true if lead has been converted
func (l *Lead) IsConverted() bool {
	return l.ConvertedAt != nil
}

// LeadProductItem represents a product of interest associated with a lead
type LeadProductItem struct {
	ID                  string                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LeadID              string                 `gorm:"type:uuid;not null;index" json:"lead_id"`
	ProductID           *string                `gorm:"type:uuid;index" json:"product_id"`
	Product             *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductName         string                 `gorm:"type:varchar(200);not null" json:"product_name"`
	ProductSKU          string                 `gorm:"type:varchar(50)" json:"product_sku"`
	InterestLevel       int                    `gorm:"type:int;default:0" json:"interest_level"`
	Quantity            int                    `gorm:"type:int;default:1" json:"quantity"`
	UnitPrice           float64                `gorm:"type:decimal(15,2);default:0" json:"unit_price"`
	Notes               string                 `gorm:"type:text" json:"notes"`
	SourceVisitReportID *string                `gorm:"type:uuid;index" json:"source_visit_report_id"`
	LastSurveyAnswers   *string                `gorm:"type:jsonb" json:"last_survey_answers"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LeadProductItem) TableName() string {
	return "crm_lead_product_items"
}

func (i *LeadProductItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}
