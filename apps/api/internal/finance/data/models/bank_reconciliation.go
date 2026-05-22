package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankReconciliationStatus string

const (
	BankReconciliationStatusInProgress BankReconciliationStatus = "in_progress"
	BankReconciliationStatusReconciled BankReconciliationStatus = "reconciled"
	BankReconciliationStatusLocked     BankReconciliationStatus = "locked"
)

type BankStatementLineStatus string

const (
	BankStatementLineStatusUnmatched     BankStatementLineStatus = "unmatched"
	BankStatementLineStatusAutoMatched   BankStatementLineStatus = "auto_matched"
	BankStatementLineStatusManualMatched BankStatementLineStatus = "manual_matched"
	BankStatementLineStatusExcluded      BankStatementLineStatus = "excluded"
)

type BankStatementLineDirection string

const (
	BankStatementLineDirectionDebit  BankStatementLineDirection = "debit"
	BankStatementLineDirectionCredit BankStatementLineDirection = "credit"
)

type BankReconciliation struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID     string `gorm:"type:uuid;not null;index" json:"company_id"`
	BankAccountID string `gorm:"type:uuid;not null;index" json:"bank_account_id"`

	StatementDate    time.Time `gorm:"type:date;not null;index" json:"statement_date"`
	StatementBalance float64   `gorm:"type:numeric(20,4);not null" json:"statement_balance"`
	BookBalance      float64   `gorm:"type:numeric(20,4);not null;default:0" json:"book_balance"`
	Difference       float64   `gorm:"type:numeric(20,4);not null;default:0" json:"difference"`

	FileFormat string `gorm:"type:varchar(10);not null" json:"file_format"`
	FileName   string `gorm:"type:varchar(255)" json:"file_name"`

	Status BankReconciliationStatus `gorm:"type:varchar(20);not null;default:'in_progress';index" json:"status"`

	CreatedBy    *string    `gorm:"type:uuid" json:"created_by,omitempty"`
	ReconciledBy *string    `gorm:"type:uuid" json:"reconciled_by,omitempty"`
	LockedBy     *string    `gorm:"type:uuid" json:"locked_by,omitempty"`
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
	LockedAt     *time.Time `json:"locked_at,omitempty"`

	Lines []BankStatementLine `gorm:"foreignKey:BankReconciliationID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BankReconciliation) TableName() string {
	return "bank_reconciliations"
}

func (r *BankReconciliation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

type BankStatementLine struct {
	ID string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	BankReconciliationID string `gorm:"type:uuid;not null;index" json:"bank_reconciliation_id"`

	Date        time.Time                  `gorm:"type:date;not null;index" json:"date"`
	Reference   string                     `gorm:"type:varchar(120)" json:"reference"`
	Description string                     `gorm:"type:text" json:"description"`
	Amount      float64                    `gorm:"type:numeric(20,4);not null" json:"amount"`
	Direction   BankStatementLineDirection `gorm:"type:varchar(10);not null" json:"direction"`
	Status      BankStatementLineStatus    `gorm:"type:varchar(20);not null;default:'unmatched';index" json:"status"`

	MatchedWithTransactionID *string `gorm:"type:uuid;index" json:"matched_with_transaction_id,omitempty"`
	ExcludeReason            string  `gorm:"type:text" json:"exclude_reason"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

func (BankStatementLine) TableName() string {
	return "bank_statement_lines"
}

func (l *BankStatementLine) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
