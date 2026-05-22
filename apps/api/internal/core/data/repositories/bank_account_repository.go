package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"sort"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BankAccountListParams struct {
	CompanyID   string // Phase 2: required for company scoping
	Search      string
	IsActive    *bool
	OwnerType   string
	CurrencyID  string
	AccountType string // Phase 2: optional filter
	SortBy      string
	SortDir     string
	Limit       int
	Offset      int
}

// Phase 2: Bank account response with computed balance and metadata
type BankAccountDetail struct {
	*models.BankAccount
	CurrentBalance     float64                      `json:"current_balance"`     // computed from GL
	IsReconcilable     bool                         `json:"is_reconcilable"`     // derived from COA
	BalanceBreakdown   *BankAccountBalanceBreakdown `json:"balance_breakdown"`   // detailed balance composition
	RecentTransactions []BankAccountTransaction     `json:"recent_transactions"` // last n transactions
	Metadata           *BankAccountMetadata         `json:"metadata"`            // reconciliation status, etc.
}

type BankAccountBalanceBreakdown struct {
	OpeningJournalBalance  float64 `json:"opening_journal_balance"`
	TransactionDebitTotal  float64 `json:"transaction_debit_total"`
	TransactionCreditTotal float64 `json:"transaction_credit_total"`
	CurrentBalance         float64 `json:"current_balance"`
}

type BankAccountMetadata struct {
	LastReconciledAt     *time.Time `json:"last_reconciled_at"`
	ReconciliationStatus string     `json:"reconciliation_status"` // unreconciled|in_progress|reconciled|locked
	StatementDate        *time.Time `json:"statement_date"`
	BookDifference       float64    `json:"book_difference"`
	WarningMessage       *string    `json:"warning_message,omitempty"` // warning if account registered during operations
}

type BankAccountRepository interface {
	Create(ctx context.Context, bankAccount *models.BankAccount) error
	FindByID(ctx context.Context, id string) (*models.BankAccount, error)
	List(ctx context.Context, params BankAccountListParams) ([]models.BankAccount, int64, error)
	ListUnified(ctx context.Context, params BankAccountListParams) ([]UnifiedBankAccount, int64, error)
	ListTransactionHistory(ctx context.Context, bankAccountID string, limit int) ([]BankAccountTransaction, error)
	ListTransactionHistoryPaginated(ctx context.Context, bankAccountID string, limit, offset int) ([]BankAccountTransaction, int64, error)
	Update(ctx context.Context, bankAccount *models.BankAccount) error
	Delete(ctx context.Context, id string) error
	// Phase 2 methods
	FindByIDWithBalance(ctx context.Context, id string) (*BankAccountDetail, error)
	ListByCompanyWithBalance(ctx context.Context, companyID string, params BankAccountListParams) ([]BankAccountDetail, int64, error)
	ComputeCurrentBalance(ctx context.Context, bankAccountID string) (float64, error)
	ToggleStatus(ctx context.Context, bankAccountID string) error
}

type UnifiedBankAccount struct {
	ID                    string
	SourceType            string
	Name                  string
	BankName              *string
	BankCode              *string
	AccountNumber         string
	AccountHolder         string
	CurrencyID            *string
	CurrencyCode          string
	CurrencyName          *string
	CurrencySymbol        *string
	CurrencyDecimalPlaces *int
	OwnerType             string
	OwnerID               *string
	OwnerName             string
	OwnerCode             *string
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type BankAccountTransaction struct {
	ID                 string
	TransactionType    string
	TransactionDate    time.Time
	ReferenceType      string
	ReferenceID        string
	ReferenceNumber    *string
	RelatedEntityType  *string
	RelatedEntityID    *string
	RelatedEntityLabel *string
	Amount             float64
	Status             string
	Description        string
}

type bankAccountRepository struct {
	db *gorm.DB
}

func NewBankAccountRepository(db *gorm.DB) BankAccountRepository {
	return &bankAccountRepository{db: db}
}

func applyQualifiedTenantFilter(ctx context.Context, query *gorm.DB, qualifiedColumns ...string) *gorm.DB {
	if query == nil {
		return query
	}

	if middleware.IsSystemAdmin(ctx) {
		return query
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return query
	}

	for _, col := range qualifiedColumns {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		query = query.Where(col+" = ?", tenantID)
	}

	return query
}

func (r *bankAccountRepository) Create(ctx context.Context, bankAccount *models.BankAccount) error {
	return database.GetDB(ctx, r.db).Create(bankAccount).Error
}

func (r *bankAccountRepository) FindByID(ctx context.Context, id string) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions()).Preload("CurrencyDetail")
	if err := q.First(&bankAccount, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &bankAccount, nil
}

var bankAccountAllowedSort = map[string]string{
	"created_at":     "bank_accounts.created_at",
	"updated_at":     "bank_accounts.updated_at",
	"name":           "bank_accounts.name",
	"code":           "bank_accounts.code",
	"account_type":   "bank_accounts.account_type",
	"currency":       "bank_accounts.currency",
	"account_number": "bank_accounts.account_number",
}

func (r *bankAccountRepository) List(ctx context.Context, params BankAccountListParams) ([]models.BankAccount, int64, error) {
	var items []models.BankAccount
	var total int64

	q := database.GetDB(ctx, r.db).Model(&models.BankAccount{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	// Phase 2: Company scoping
	if params.CompanyID != "" {
		q = q.Where("company_id = ?", params.CompanyID)
	}

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where(
			"bank_accounts.name ILIKE ? OR bank_accounts.code ILIKE ? OR bank_accounts.account_number ILIKE ? OR bank_accounts.account_holder ILIKE ?",
			like, like, like, like,
		)
	}
	if params.IsActive != nil {
		q = q.Where("bank_accounts.is_active = ?", *params.IsActive)
	}
	// Phase 2: Account type filter
	if params.AccountType != "" {
		q = q.Where("bank_accounts.account_type = ?", params.AccountType)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := bankAccountAllowedSort[params.SortBy]
	if sortCol == "" {
		sortCol = bankAccountAllowedSort["created_at"]
	}
	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	if sortDir != "asc" {
		sortDir = "desc"
	}
	q = q.Order(clause.OrderByColumn{
		Column: clause.Column{Name: sortCol},
		Desc:   sortDir == "desc",
	})

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Preload("CurrencyDetail").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *bankAccountRepository) Update(ctx context.Context, bankAccount *models.BankAccount) error {
	return database.GetDB(ctx, r.db).Save(bankAccount).Error
}

func (r *bankAccountRepository) ListUnified(ctx context.Context, params BankAccountListParams) ([]UnifiedBankAccount, int64, error) {
	items := make([]UnifiedBankAccount, 0)

	var companyAccounts []models.BankAccount
	companyQuery := database.GetDB(ctx, r.db).Preload("CurrencyDetail").Model(&models.BankAccount{})
	companyQuery = security.ApplyScopeFilter(companyQuery, ctx, security.FinanceScopeQueryOptions())
	if params.IsActive != nil {
		companyQuery = companyQuery.Where("is_active = ?", *params.IsActive)
	}
	if err := companyQuery.Find(&companyAccounts).Error; err != nil {
		return nil, 0, err
	}
	for _, account := range companyAccounts {
		var decimalPlaces *int
		var currencyName *string
		var currencySymbol *string
		if account.CurrencyDetail != nil {
			decimalPlaces = &account.CurrencyDetail.DecimalPlaces
			currencyName = &account.CurrencyDetail.Name
			currencySymbol = &account.CurrencyDetail.Symbol
		}
		items = append(items, UnifiedBankAccount{
			ID:                    account.ID,
			SourceType:            "company",
			Name:                  account.Name,
			AccountNumber:         account.AccountNumber,
			AccountHolder:         account.AccountHolder,
			CurrencyID:            account.CurrencyID,
			CurrencyCode:          account.Currency,
			CurrencyName:          currencyName,
			CurrencySymbol:        currencySymbol,
			CurrencyDecimalPlaces: decimalPlaces,
			OwnerType:             "company",
			OwnerName:             "Company",
			IsActive:              account.IsActive,
			CreatedAt:             account.CreatedAt,
			UpdatedAt:             account.UpdatedAt,
		})
	}

	var customerBanks []customerModels.CustomerBank
	customerBankQuery := r.db.WithContext(ctx).
		Preload("Bank").
		Preload("Currency").
		Joins("JOIN customers ON customers.id = customer_banks.customer_id AND customers.deleted_at IS NULL")
	customerBankQuery = applyQualifiedTenantFilter(ctx, customerBankQuery, "customer_banks.tenant_id", "customers.tenant_id")
	if err := customerBankQuery.Find(&customerBanks).Error; err != nil {
		return nil, 0, err
	}
	for _, bank := range customerBanks {
		var owner struct {
			ID   string
			Code string
			Name string
		}
		if err := database.GetDB(ctx, r.db).Table("customers").Select("id, code, name").Where("id = ?", bank.CustomerID).Scan(&owner).Error; err != nil {
			return nil, 0, err
		}
		bankName := ""
		bankCode := ""
		if bank.Bank != nil {
			bankName = bank.Bank.Name
			bankCode = bank.Bank.Code
		}
		var decimalPlaces *int
		var currencyName *string
		var currencySymbol *string
		var currencyCode string
		if bank.Currency != nil {
			decimalPlaces = &bank.Currency.DecimalPlaces
			currencyName = &bank.Currency.Name
			currencySymbol = &bank.Currency.Symbol
			currencyCode = bank.Currency.Code
		}
		ownerID := owner.ID
		ownerCode := owner.Code
		items = append(items, UnifiedBankAccount{
			ID:                    bank.ID,
			SourceType:            "customer",
			Name:                  firstNonEmpty(bankName, bank.AccountName),
			BankName:              stringPtrIfNotEmpty(bankName),
			BankCode:              stringPtrIfNotEmpty(bankCode),
			AccountNumber:         bank.AccountNumber,
			AccountHolder:         bank.AccountName,
			CurrencyID:            bank.CurrencyID,
			CurrencyCode:          currencyCode,
			CurrencyName:          currencyName,
			CurrencySymbol:        currencySymbol,
			CurrencyDecimalPlaces: decimalPlaces,
			OwnerType:             "customer",
			OwnerID:               &ownerID,
			OwnerName:             owner.Name,
			OwnerCode:             &ownerCode,
			IsActive:              true,
			CreatedAt:             bank.CreatedAt,
			UpdatedAt:             bank.UpdatedAt,
		})
	}

	var supplierBanks []supplierModels.SupplierBank
	supplierBankQuery := r.db.WithContext(ctx).
		Preload("Bank").
		Preload("Currency").
		Joins("JOIN suppliers ON suppliers.id = supplier_banks.supplier_id AND suppliers.deleted_at IS NULL")
	supplierBankQuery = applyQualifiedTenantFilter(ctx, supplierBankQuery, "supplier_banks.tenant_id", "suppliers.tenant_id")
	if err := supplierBankQuery.Find(&supplierBanks).Error; err != nil {
		return nil, 0, err
	}
	for _, bank := range supplierBanks {
		var owner struct {
			ID   string
			Code string
			Name string
		}
		if err := database.GetDB(ctx, r.db).Table("suppliers").Select("id, code, name").Where("id = ?", bank.SupplierID).Scan(&owner).Error; err != nil {
			return nil, 0, err
		}
		bankName := ""
		bankCode := ""
		if bank.Bank != nil {
			bankName = bank.Bank.Name
			bankCode = bank.Bank.Code
		}
		var decimalPlaces *int
		var currencyName *string
		var currencySymbol *string
		var currencyCode string
		if bank.Currency != nil {
			decimalPlaces = &bank.Currency.DecimalPlaces
			currencyName = &bank.Currency.Name
			currencySymbol = &bank.Currency.Symbol
			currencyCode = bank.Currency.Code
		}
		ownerID := owner.ID
		ownerCode := owner.Code
		items = append(items, UnifiedBankAccount{
			ID:                    bank.ID,
			SourceType:            "supplier",
			Name:                  firstNonEmpty(bankName, bank.AccountName),
			BankName:              stringPtrIfNotEmpty(bankName),
			BankCode:              stringPtrIfNotEmpty(bankCode),
			AccountNumber:         bank.AccountNumber,
			AccountHolder:         bank.AccountName,
			CurrencyID:            bank.CurrencyID,
			CurrencyCode:          currencyCode,
			CurrencyName:          currencyName,
			CurrencySymbol:        currencySymbol,
			CurrencyDecimalPlaces: decimalPlaces,
			OwnerType:             "supplier",
			OwnerID:               &ownerID,
			OwnerName:             owner.Name,
			OwnerCode:             &ownerCode,
			IsActive:              true,
			CreatedAt:             bank.CreatedAt,
			UpdatedAt:             bank.UpdatedAt,
		})
	}

	if search := strings.ToLower(strings.TrimSpace(params.Search)); search != "" {
		filtered := make([]UnifiedBankAccount, 0, len(items))
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Name), search) ||
				strings.Contains(strings.ToLower(item.AccountNumber), search) ||
				strings.Contains(strings.ToLower(item.AccountHolder), search) ||
				strings.Contains(strings.ToLower(item.OwnerName), search) ||
				(item.BankName != nil && strings.Contains(strings.ToLower(*item.BankName), search)) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if ownerType := strings.ToLower(strings.TrimSpace(params.OwnerType)); ownerType != "" {
		filtered := make([]UnifiedBankAccount, 0, len(items))
		for _, item := range items {
			if strings.EqualFold(item.OwnerType, ownerType) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if currencyID := strings.TrimSpace(params.CurrencyID); currencyID != "" {
		filtered := make([]UnifiedBankAccount, 0, len(items))
		for _, item := range items {
			if item.CurrencyID != nil && *item.CurrencyID == currencyID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	if sortDir != "asc" {
		sortDir = "desc"
	}
	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		var result bool
		switch params.SortBy {
		case "name":
			result = left.Name < right.Name
		case "owner_name":
			result = left.OwnerName < right.OwnerName
		case "currency":
			result = left.CurrencyCode < right.CurrencyCode
		default:
			result = left.CreatedAt.Before(right.CreatedAt)
		}
		if sortDir == "asc" {
			return result
		}
		return !result
	})

	total := int64(len(items))
	start := params.Offset
	if start > len(items) {
		start = len(items)
	}
	end := len(items)
	if params.Limit > 0 && start+params.Limit < end {
		end = start + params.Limit
	}

	return items[start:end], total, nil
}

func (r *bankAccountRepository) ListTransactionHistory(ctx context.Context, bankAccountID string, limit int) ([]BankAccountTransaction, error) {
	items, _, err := r.ListTransactionHistoryPaginated(ctx, bankAccountID, limit, 0)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *bankAccountRepository) ListTransactionHistoryPaginated(ctx context.Context, bankAccountID string, limit, offset int) ([]BankAccountTransaction, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	if _, err := r.FindByID(ctx, bankAccountID); err != nil {
		return nil, 0, err
	}

	const baseQuery = `
		FROM (
			SELECT
				p.id,
				'payment' AS transaction_type,
				p.payment_date::timestamp AS transaction_date,
				'PAYMENT' AS reference_type,
				p.id AS reference_id,
				NULL::text AS reference_number,
				NULL::text AS related_entity_type,
				NULL::text AS related_entity_id,
				NULL::text AS related_entity_label,
				p.total_amount AS amount,
				p.status::text AS status,
				COALESCE(p.description, '') AS description
			FROM payments p
			WHERE p.bank_account_id = ? AND p.deleted_at IS NULL

			UNION ALL

			SELECT
				cbj.id,
				'cash_bank_journal' AS transaction_type,
				cbj.transaction_date::timestamp AS transaction_date,
				'CASH_BANK_JOURNAL' AS reference_type,
				cbj.id AS reference_id,
				NULL::text AS reference_number,
				NULL::text AS related_entity_type,
				NULL::text AS related_entity_id,
				NULL::text AS related_entity_label,
				cbj.total_amount AS amount,
				cbj.status::text AS status,
				COALESCE(cbj.description, '') AS description
			FROM cash_bank_journals cbj
			WHERE cbj.bank_account_id = ? AND cbj.deleted_at IS NULL

			UNION ALL

			SELECT
				sp.id,
				'sales_payment' AS transaction_type,
				sp.created_at AS transaction_date,
				'SALES_PAYMENT' AS reference_type,
				sp.id AS reference_id,
				COALESCE(sp.reference_number, ci.invoice_number, ci.code) AS reference_number,
				'customer' AS related_entity_type,
				so.customer_id::text AS related_entity_id,
				COALESCE(so.customer_name, '') AS related_entity_label,
				sp.amount AS amount,
				sp.status::text AS status,
				COALESCE(sp.notes, '') AS description
			FROM sales_payments sp
			LEFT JOIN customer_invoices ci ON ci.id = sp.customer_invoice_id
			LEFT JOIN sales_orders so ON so.id = ci.sales_order_id
			WHERE sp.bank_account_id = ? AND sp.deleted_at IS NULL

			UNION ALL

			SELECT
				pp.id,
				'purchase_payment' AS transaction_type,
				pp.created_at AS transaction_date,
				'PURCHASE_PAYMENT' AS reference_type,
				pp.id AS reference_id,
				COALESCE(pp.reference_number, si.invoice_number, si.code) AS reference_number,
				'supplier' AS related_entity_type,
				si.supplier_id::text AS related_entity_id,
				COALESCE(si.supplier_name_snapshot, '') AS related_entity_label,
				pp.amount AS amount,
				pp.status::text AS status,
				COALESCE(pp.notes, '') AS description
			FROM purchase_payments pp
			LEFT JOIN supplier_invoices si ON si.id = pp.supplier_invoice_id
			WHERE pp.bank_account_id = ? AND pp.deleted_at IS NULL
		) t
	`

	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int64
	if err := database.GetDB(ctx, r.db).Raw(countQuery, bankAccountID, bankAccountID, bankAccountID, bankAccountID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			t.id,
			t.transaction_type,
			t.transaction_date,
			t.reference_type,
			t.reference_id,
			t.reference_number,
			t.related_entity_type,
			t.related_entity_id,
			t.related_entity_label,
			t.amount,
			t.status,
			t.description
		` + baseQuery + `
		ORDER BY t.transaction_date DESC
		LIMIT ?
		OFFSET ?
	`

	items := make([]BankAccountTransaction, 0)
	if err := database.GetDB(ctx, r.db).Raw(query, bankAccountID, bankAccountID, bankAccountID, bankAccountID, limit, offset).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *bankAccountRepository) Delete(ctx context.Context, id string) error {
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions())
	return q.Delete(&models.BankAccount{}, "id = ?", id).Error
}

// ========== PHASE 2 METHODS ==========

// FindByIDWithBalance retrieves bank account with computed current balance from GL
func (r *bankAccountRepository) FindByIDWithBalance(ctx context.Context, id string) (*BankAccountDetail, error) {
	var bankAccount models.BankAccount
	if err := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions()).
		Preload("CurrencyDetail").
		First(&bankAccount, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Compute current balance from GL journal entries for this bank's COA account
	currentBalance, err := r.ComputeCurrentBalance(ctx, id)
	if err != nil {
		// Log error but continue; balance will be estimated
		currentBalance = 0
	}

	// Get recent transactions
	recentTxns, _ := r.ListTransactionHistory(ctx, id, 5)

	// Determine if account is reconcilable based on COA postability
	isReconcilable := bankAccount.ChartOfAccountID != nil

	// Balance breakdown
	balanceBreakdown := &BankAccountBalanceBreakdown{
		OpeningJournalBalance:  bankAccount.OpeningBalance, // Phase 0 opening balance
		TransactionDebitTotal:  0,                          // TODO: aggregate from GL
		TransactionCreditTotal: 0,                          // TODO: aggregate from GL
		CurrentBalance:         currentBalance,
	}

	detail := &BankAccountDetail{
		BankAccount:        &bankAccount,
		CurrentBalance:     currentBalance,
		IsReconcilable:     isReconcilable,
		BalanceBreakdown:   balanceBreakdown,
		RecentTransactions: recentTxns,
		Metadata: &BankAccountMetadata{
			ReconciliationStatus: "unreconciled", // TODO: fetch from reconciliation table
		},
	}

	return detail, nil
}

// ListByCompanyWithBalance retrieves bank accounts for a company with computed balances
func (r *bankAccountRepository) ListByCompanyWithBalance(ctx context.Context, companyID string, params BankAccountListParams) ([]BankAccountDetail, int64, error) {
	var bankAccounts []models.BankAccount
	var total int64

	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions()).
		Where("company_id = ?", companyID).
		Preload("CurrencyDetail")

	// Apply filters
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where(
			"bank_accounts.name ILIKE ? OR bank_accounts.code ILIKE ? OR bank_accounts.account_number ILIKE ?",
			like, like, like,
		)
	}
	if params.IsActive != nil {
		q = q.Where("bank_accounts.is_active = ?", *params.IsActive)
	}
	if params.AccountType != "" {
		q = q.Where("bank_accounts.account_type = ?", params.AccountType)
	}
	if params.CurrencyID != "" {
		q = q.Where("bank_accounts.currency_id = ?", params.CurrencyID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortCol := "bank_accounts.created_at"
	if params.SortBy == "name" {
		sortCol = "bank_accounts.name"
	} else if params.SortBy == "code" {
		sortCol = "bank_accounts.code"
	} else if params.SortBy == "account_number" {
		sortCol = "bank_accounts.account_number"
	}
	sortDir := "desc"
	if strings.ToLower(params.SortDir) == "asc" {
		sortDir = "asc"
	}
	q = q.Order(clause.OrderByColumn{
		Column: clause.Column{Name: sortCol},
		Desc:   sortDir == "desc",
	})

	// Apply pagination
	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&bankAccounts).Error; err != nil {
		return nil, 0, err
	}

	// Enrich with computed balances
	details := make([]BankAccountDetail, 0, len(bankAccounts))
	for _, acc := range bankAccounts {
		currentBalance, _ := r.ComputeCurrentBalance(ctx, acc.ID)
		recentTxns, _ := r.ListTransactionHistory(ctx, acc.ID, 5)

		detail := BankAccountDetail{
			BankAccount:        &acc,
			CurrentBalance:     currentBalance,
			IsReconcilable:     acc.ChartOfAccountID != nil,
			RecentTransactions: recentTxns,
			Metadata: &BankAccountMetadata{
				ReconciliationStatus: "unreconciled",
			},
		}
		details = append(details, detail)
	}

	return details, total, nil
}

// ComputeCurrentBalance calculates current balance from posted GL journal entries
// Returns: SUM(debit - credit) for the bank account's COA account
func (r *bankAccountRepository) ComputeCurrentBalance(ctx context.Context, bankAccountID string) (float64, error) {
	var bankAccount models.BankAccount
	if err := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions()).
		Select("chart_of_account_id").
		First(&bankAccount, "id = ?", bankAccountID).Error; err != nil {
		return 0, err
	}

	if bankAccount.ChartOfAccountID == nil || *bankAccount.ChartOfAccountID == "" {
		return 0, nil // No COA linked, balance is 0
	}

	// Query GL balance: SUM(debit) - SUM(credit) for posted journals
	var balance struct {
		Balance float64
	}

	query := `
		SELECT COALESCE(
			SUM(COALESCE(jel.debit_amount, 0)) - SUM(COALESCE(jel.credit_amount, 0)),
			0
		) AS balance
		FROM journal_entry_lines jel
		JOIN journal_entries je ON je.id = jel.journal_entry_id
		WHERE jel.account_id = ?
		  AND je.status = 'posted'
		  AND je.deleted_at IS NULL
	`

	if err := database.GetDB(ctx, r.db).Raw(query, *bankAccount.ChartOfAccountID).Scan(&balance).Error; err != nil {
		return 0, err
	}

	return balance.Balance, nil
}

// ToggleStatus toggles the is_active flag of a bank account
func (r *bankAccountRepository) ToggleStatus(ctx context.Context, bankAccountID string) error {
	var bankAccount models.BankAccount
	if err := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&models.BankAccount{}), ctx, security.FinanceScopeQueryOptions()).First(&bankAccount, "id = ?", bankAccountID).Error; err != nil {
		return err
	}

	bankAccount.IsActive = !bankAccount.IsActive
	return database.GetDB(ctx, r.db).Save(&bankAccount).Error
}

func stringPtrIfNotEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
