package models

import (
	"time"

	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JournalStatus string

const (
	JournalStatusDraft     JournalStatus = "draft"
	JournalStatusPosted    JournalStatus = "posted"
	JournalStatusReversed  JournalStatus = "reversed"
	JournalStatusCancelled JournalStatus = "cancelled"
)

type ReferenceType string

const (
	RefSO                ReferenceType = "SO"                     // Sales Order
	RefPO                ReferenceType = "PO"                     // Purchase Order
	RefDO                ReferenceType = "DO"                     // Delivery Order
	RefGR                ReferenceType = "GR"                     // Goods Receipt
	RefStockOpname       ReferenceType = "STOCK_OP"               // Stock Opname
	RefAdjustment        ReferenceType = "ADJUSTMENT"             // Adjustment
	RefNonTradePayable   ReferenceType = "NTP"                    // Non-Trade Payable
	RefPayment           ReferenceType = "PAYMENT"                // Payment
	RefAssetTransaction  ReferenceType = "ASSET_TXN"              // Asset Transaction
	RefAssetDepreciation ReferenceType = "ASSET_DEP"              // Asset Depreciation
	RefCashBank          ReferenceType = "CASH_BANK"              // Cash Bank
	RefUpCountryCost     ReferenceType = "UP_COUNTRY"             // Up Country Cost
	RefInventoryVal      ReferenceType = "INVENTORY_VALUATION"    // Inventory Valuation
	RefCurrencyReval     ReferenceType = "CURRENCY_REVALUATION"   // Currency Revaluation
	RefCostAdjustment    ReferenceType = "COST_ADJUSTMENT"        // Cost Adjustment
	RefDepreciation      ReferenceType = "DEPRECIATION_VALUATION" // Depreciation Valuation
	RefOpeningBalance    ReferenceType = "OPENING_BALANCE"        // Opening Balance
	RefOpeningBalanceRev ReferenceType = "OPENING_BALANCE_REV"    // Opening Balance Reversed Marker
)

// JournalSource distinguishes operational journals from valuation-generated journals.
type JournalSource string

const (
	JournalSourceOperational JournalSource = "OPERATIONAL"
	JournalSourceValuation   JournalSource = "VALUATION"
)

type JournalType string

const (
	JournalTypeGeneral        JournalType = "GENERAL"
	JournalTypeAdjustment     JournalType = "ADJUSTMENT"
	JournalTypeSales          JournalType = "SALES"
	JournalTypePurchase       JournalType = "PURCHASE"
	JournalTypeOpeningBalance JournalType = "OPENING_BALANCE"
	JournalTypeClosing        JournalType = "CLOSING"
)

type JournalEntry struct {
	ID            string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CompanyID     string    `gorm:"type:uuid;index" json:"company_id,omitempty"`
	FiscalYearID  *string   `gorm:"type:uuid;index" json:"fiscal_year_id,omitempty"`
	JournalNumber string    `gorm:"type:varchar(32);index" json:"journal_number,omitempty"`
	EntryDate     time.Time `gorm:"type:date;not null;index" json:"entry_date"`
	Reference     string    `gorm:"type:varchar(255)" json:"reference,omitempty"`
	Description   string    `gorm:"type:text" json:"description"`

	ReferenceType *string `gorm:"type:varchar(50);index;uniqueIndex:idx_journal_entry_reference" json:"reference_type"`
	ReferenceID   *string `gorm:"type:varchar(255);index;uniqueIndex:idx_journal_entry_reference" json:"reference_id"`

	Status      JournalStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	JournalType JournalType   `gorm:"type:varchar(30);default:'GENERAL';index" json:"journal_type"`
	PostedBy    *string       `gorm:"type:uuid" json:"posted_by"`
	PostedAt    *time.Time    `json:"posted_at"`

	ReversedBy *string    `gorm:"type:uuid" json:"reversed_by,omitempty"`
	ReversedAt *time.Time `json:"reversed_at,omitempty"`

	DebitTotal   float64 `gorm:"type:decimal(18,2);default:0" json:"debit_total"`
	CreditTotal  float64 `gorm:"type:decimal(18,2);default:0" json:"credit_total"`
	CurrencyCode string  `gorm:"type:varchar(10);default:'IDR'" json:"currency_code"`
	ExchangeRate float64 `gorm:"type:decimal(15,6);default:1" json:"exchange_rate"`

	OriginalJournalID *string `gorm:"type:uuid;index" json:"original_journal_id,omitempty"`
	ReversalReason    string  `gorm:"type:text" json:"reversal_reason,omitempty"`
	IsReversal        bool    `gorm:"default:false;index" json:"is_reversal"`
	ReversedFrom      *string `gorm:"type:uuid;index" json:"reversed_from,omitempty"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	IsSystemGenerated bool          `gorm:"default:false;index" json:"is_system_generated"`
	IsValuation       bool          `gorm:"default:false;index" json:"is_valuation"`
	Source            JournalSource `gorm:"type:varchar(20);default:'OPERATIONAL';index" json:"source"`
	ValuationRunID    *string       `gorm:"type:uuid;index" json:"valuation_run_id,omitempty"`
	SourceDocumentURL *string       `gorm:"type:varchar(500)" json:"source_document_url,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Lines       []JournalLine       `gorm:"foreignKey:JournalEntryID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
	Attachments []JournalAttachment `gorm:"foreignKey:JournalEntryID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`

	CreatedByUser  *userModels.User `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	PostedByUser   *userModels.User `gorm:"foreignKey:PostedBy" json:"posted_by_user,omitempty"`
	ReversedByUser *userModels.User `gorm:"foreignKey:ReversedBy" json:"reversed_by_user,omitempty"`
}

func (JournalEntry) TableName() string {
	return "journal_entries"
}

func (je *JournalEntry) BeforeCreate(tx *gorm.DB) error {
	if je.ID == "" {
		je.ID = uuid.New().String()
	}
	return nil
}
