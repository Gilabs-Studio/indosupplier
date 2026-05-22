package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JournalLine struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	JournalEntryID string        `gorm:"type:uuid;not null;index" json:"journal_entry_id"`
	JournalEntry   *JournalEntry `gorm:"foreignKey:JournalEntryID" json:"journal_entry,omitempty"`

	ChartOfAccountID           string          `gorm:"type:uuid;not null;index" json:"chart_of_account_id"`
	ChartOfAccount             *ChartOfAccount `gorm:"foreignKey:ChartOfAccountID" json:"chart_of_account,omitempty"`
	ChartOfAccountCodeSnapshot string          `gorm:"type:varchar(50)" json:"chart_of_account_code_snapshot,omitempty"`
	ChartOfAccountNameSnapshot string          `gorm:"type:varchar(200)" json:"chart_of_account_name_snapshot,omitempty"`
	ChartOfAccountTypeSnapshot string          `gorm:"type:varchar(20)" json:"chart_of_account_type_snapshot,omitempty"`

	Debit  float64 `gorm:"type:decimal(18,2);default:0" json:"debit"`
	Credit float64 `gorm:"type:decimal(18,2);default:0" json:"credit"`
	Memo   string  `gorm:"type:text" json:"memo"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JournalLine) TableName() string {
	return "journal_lines"
}

func (jl *JournalLine) BeforeCreate(tx *gorm.DB) error {
	if jl.ID == "" {
		jl.ID = uuid.New().String()
	}
	return nil
}
