package dto

type BankAccountResponse struct {
	ID                 string                           `json:"id"`
	CompanyID          string                           `json:"company_id,omitempty"`
	Code               string                           `json:"code,omitempty"`
	Name               string                           `json:"name"`
	AccountType        string                           `json:"account_type,omitempty"`
	BankID             *string                          `json:"bank_id,omitempty"`
	BankDetail         *BankMasterResponse              `json:"bank_detail,omitempty"`
	AccountNumber      string                           `json:"account_number"`
	AccountHolder      string                           `json:"account_holder"`
	CurrencyID         *string                          `json:"currency_id"`
	CurrencyDetail     *CurrencyResponse                `json:"currency_detail,omitempty"`
	Currency           string                           `json:"currency"`
	ChartOfAccountID   *string                          `json:"chart_of_account_id"`
	VillageID          *string                          `json:"village_id"`
	BankAddress        string                           `json:"bank_address"`
	BankPhone          string                           `json:"bank_phone"`
	CountryCode        string                           `json:"country_code,omitempty"`
	BankBranchCode     string                           `json:"bank_branch_code,omitempty"`
	OpeningBalance     float64                          `json:"opening_balance,omitempty"`
	CurrentBalance     float64                          `json:"current_balance,omitempty"`
	IsReconcilable     bool                             `json:"is_reconcilable,omitempty"`
	CreatedBy          string                           `json:"created_by,omitempty"`
	UpdatedBy          string                           `json:"updated_by,omitempty"`
	IsActive           bool                             `json:"is_active"`
	CreatedAt          string                           `json:"created_at"`
	UpdatedAt          string                           `json:"updated_at"`
	Warning            *BankAccountWarning              `json:"warning,omitempty"`
	RecentTransactions []BankAccountTransaction         `json:"recent_transactions,omitempty"`
	BalanceBreakdown   *BankAccountBalanceBreakdown     `json:"balance_breakdown,omitempty"`
	Metadata           *BankAccountMetadata             `json:"metadata,omitempty"`
	TransactionHistory []BankAccountTransactionResponse `json:"transaction_history,omitempty"`
}

type BankAccountTransactionResponse struct {
	ID                 string  `json:"id"`
	TransactionType    string  `json:"transaction_type"`
	TransactionDate    string  `json:"transaction_date"`
	ReferenceType      string  `json:"reference_type"`
	ReferenceID        string  `json:"reference_id"`
	ReferenceNumber    *string `json:"reference_number,omitempty"`
	RelatedEntityType  *string `json:"related_entity_type,omitempty"`
	RelatedEntityID    *string `json:"related_entity_id,omitempty"`
	RelatedEntityLabel *string `json:"related_entity_label,omitempty"`
	Amount             float64 `json:"amount"`
	Status             string  `json:"status"`
	Description        string  `json:"description"`
}

type UnifiedBankAccountResponse struct {
	ID             string            `json:"id"`
	SourceType     string            `json:"source_type"`
	Name           string            `json:"name"`
	BankName       *string           `json:"bank_name,omitempty"`
	BankCode       *string           `json:"bank_code,omitempty"`
	AccountNumber  string            `json:"account_number"`
	AccountHolder  string            `json:"account_holder"`
	CurrencyID     *string           `json:"currency_id"`
	Currency       string            `json:"currency"`
	CurrencyDetail *CurrencyResponse `json:"currency_detail,omitempty"`
	OwnerType      string            `json:"owner_type"`
	OwnerID        *string           `json:"owner_id,omitempty"`
	OwnerName      string            `json:"owner_name"`
	OwnerCode      *string           `json:"owner_code,omitempty"`
	IsActive       bool              `json:"is_active"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type CreateBankAccountRequest struct {
	CompanyID        string  `json:"company_id" binding:"required,uuid"`
	Code             string  `json:"code" binding:"required,max=50"`
	AccountType      string  `json:"account_type" binding:"omitempty,oneof=operational suspense transit"`
	BankID           *string `json:"bank_id" binding:"omitempty,uuid"`
	Name             string  `json:"name" binding:"required"`
	AccountNumber    string  `json:"account_number" binding:"required"`
	AccountHolder    string  `json:"account_holder" binding:"required"`
	CurrencyID       *string `json:"currency_id" binding:"required,uuid"`
	Currency         string  `json:"currency"`
	ChartOfAccountID *string `json:"chart_of_account_id"`
	VillageID        *string `json:"village_id" binding:"omitempty,uuid"`
	BankAddress      string  `json:"bank_address"`
	BankPhone        string  `json:"bank_phone"`
	CountryCode      string  `json:"country_code" binding:"omitempty,max=2"`
	BankBranchCode   string  `json:"bank_branch_code" binding:"omitempty,max=20"`
	OpeningBalance   float64 `json:"opening_balance" binding:"omitempty,min=0"`
	IsActive         *bool   `json:"is_active"`
}

// ========== PHASE 2 RESPONSE/REQUEST TYPES ==========

// BankAccountBalanceBreakdown breaks down GL balance composition
type BankAccountBalanceBreakdown struct {
	OpeningJournalBalance  float64 `json:"opening_journal_balance"`
	TransactionDebitTotal  float64 `json:"transaction_debit_total"`
	TransactionCreditTotal float64 `json:"transaction_credit_total"`
	CurrentBalance         float64 `json:"current_balance"`
}

// BankAccountMetadata provides reconciliation and account metadata
type BankAccountMetadata struct {
	LastReconciledAt     *string `json:"last_reconciled_at,omitempty"`
	ReconciliationStatus string  `json:"reconciliation_status,omitempty"`
	StatementDate        *string `json:"statement_date,omitempty"`
	BookDifference       float64 `json:"book_difference,omitempty"`
}

// BankAccountWarning provides warning messages for bank accounts
type BankAccountWarning struct {
	Type    string `json:"type"` // "ACCOUNT_CREATED_DURING_OPERATIONS"
	Message string `json:"message"`
	Level   string `json:"level"` // "info" | "warning"
}

// BankAccountTransaction represents a recent transaction for a bank account
type BankAccountTransaction struct {
	ID              string  `json:"id"`
	ReferenceNumber string  `json:"reference_number"`
	Type            string  `json:"type"`
	Date            string  `json:"date"`
	Amount          float64 `json:"amount"`
	JournalEntryID  *string `json:"journal_entry_id,omitempty"`
	JournalEntryNum string  `json:"journal_entry_number,omitempty"`
	Status          string  `json:"status"`
	Description     string  `json:"description,omitempty"`
}

// ToggleStatusResponse is returned when toggling bank account status
type ToggleStatusResponse struct {
	ID       string `json:"id"`
	IsActive bool   `json:"is_active"`
}

type UpdateBankAccountRequest struct {
	Name             string  `json:"name" binding:"required"`
	Code             string  `json:"code" binding:"omitempty,max=50"`
	AccountType      string  `json:"account_type" binding:"omitempty,oneof=operational suspense transit"`
	BankID           *string `json:"bank_id" binding:"omitempty,uuid"`
	AccountNumber    string  `json:"account_number" binding:"required"`
	AccountHolder    string  `json:"account_holder" binding:"required"`
	CurrencyID       *string `json:"currency_id" binding:"required,uuid"`
	Currency         string  `json:"currency"`
	ChartOfAccountID *string `json:"chart_of_account_id"`
	VillageID        *string `json:"village_id" binding:"omitempty,uuid"`
	BankAddress      string  `json:"bank_address"`
	BankPhone        string  `json:"bank_phone"`
	CountryCode      string  `json:"country_code" binding:"omitempty,max=2"`
	BankBranchCode   string  `json:"bank_branch_code" binding:"omitempty,max=20"`
	OpeningBalance   float64 `json:"opening_balance" binding:"omitempty,min=0"`
	IsActive         *bool   `json:"is_active"`
}

type BankMasterResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SwiftCode string `json:"swift_code,omitempty"`
}
