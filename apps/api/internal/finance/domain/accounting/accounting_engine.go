package accounting

import (
	"context"
	"fmt"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/service"
	"time"
)

// TransactionData holds the data needed by the AccountingEngine to generate journal lines.
// Each transaction type populates the fields it needs.
type TransactionData struct {
	// ReferenceType canonical reference type constant
	ReferenceType string

	// ReferenceID the source document ID
	ReferenceID string

	// CompanyID optional company scope for journal posting
	CompanyID string

	// FiscalYearID optional fiscal year scope for journal posting
	FiscalYearID *string

	// EntryDate in "2006-01-02" format
	EntryDate string

	// Description for the journal entry header
	Description string

	// TotalAmount is the primary amount for the transaction
	TotalAmount       float64
	SubTotal          float64
	TaxTotal          float64
	COGSTotal         float64
	DepositTotal      float64
	OtherTotal        float64
	GRIRVarianceTotal float64

	// TransactionCOAID is the COA selected by the user on the transaction (e.g. NTP expense account)
	TransactionCOAID string

	// BankAccountCOAID is the COA linked to the bank account (for cash/bank and payment modules)
	BankAccountCOAID string

	// PaymentAccountCOAID is the COA for the payment target account
	PaymentAccountCOAID string

	// LineItems are individual line-level amounts (for cash bank, payment allocations)
	LineItems []TransactionLineItem

	// DescriptionArgs used to format the profile's DescriptionTemplate via fmt.Sprintf
	DescriptionArgs []interface{}

	// MemoArgs used for memo templates
	MemoArgs []interface{}
}

// TransactionLineItem represents a single line within a multi-line transaction.
type TransactionLineItem struct {
	ChartOfAccountID string
	Amount           float64
	Memo             string
}

// AccountingEngine is the central engine for generating journal entries from
// finance transactions. It uses posting profiles and settings to resolve COA
// mappings at runtime, ensuring consistent journal generation across all modules.
type AccountingEngine interface {
	// GenerateJournal creates a CreateJournalEntryRequest from the given profile and transaction data.
	// The caller can then pass the request to JournalEntryUsecase.PostOrUpdateJournal.
	GenerateJournal(ctx context.Context, profile PostingProfile, data TransactionData) (*dto.CreateJournalEntryRequest, error)

	// ResolveCOAID resolves a COA ID from a settings key (e.g. "coa.expense").
	// This is useful for budget checks and manual account lookups.
	ResolveCOAID(ctx context.Context, settingKey string) (string, error)

	// GetAccountBalance calculates the running balance of a COA account as of a specific date.
	GetAccountBalance(ctx context.Context, coaID string, asOf time.Time) (float64, error)

	// GetAccountBalanceInRange calculates net movement (debit-credit) for a COA account within an inclusive date range.
	GetAccountBalanceInRange(ctx context.Context, coaID string, fromDate, toDate time.Time) (float64, error)
}

type accountingEngine struct {
	settingsService  financesettings.SettingsService
	coaRepo          repositories.ChartOfAccountRepository
	coaValidationSvc service.COAValidationService
}

// NewAccountingEngine creates a new central accounting engine with COA validation.
func NewAccountingEngine(settingsService financesettings.SettingsService, coaRepo repositories.ChartOfAccountRepository, coaValidationSvc service.COAValidationService) AccountingEngine {
	return &accountingEngine{
		settingsService:  settingsService,
		coaRepo:          coaRepo,
		coaValidationSvc: coaValidationSvc,
	}
}

// GenerateJournal builds a balanced journal entry request from a posting profile.
// It validates that all required COA settings are configured before processing.
func (e *accountingEngine) GenerateJournal(ctx context.Context, profile PostingProfile, data TransactionData) (*dto.CreateJournalEntryRequest, error) {
	// VALIDATION: Ensure all required COA settings are configured
	if err := e.validateProfileCOAs(ctx, profile); err != nil {
		return nil, fmt.Errorf("accounting engine validation failed: %w", err)
	}

	if strings.TrimSpace(data.TransactionCOAID) != "" {
		coa, err := e.coaRepo.FindByID(ctx, strings.TrimSpace(data.TransactionCOAID))
		if err != nil {
			return nil, fmt.Errorf("TransactionCOAID %s not found: %w", strings.TrimSpace(data.TransactionCOAID), err)
		}
		if !coa.IsActive {
			return nil, fmt.Errorf("COA %s (%s) is not active", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
		}
		if !coa.IsPostable {
			return nil, fmt.Errorf("COA %s (%s) is not postable", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
		}
		if coa.Type != financeModels.AccountTypeAsset && coa.Type != financeModels.AccountTypeCashBank {
			return nil, fmt.Errorf("COA %s (%s) for payment must be ASSET/CASH_BANK type, got %s", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name), coa.Type)
		}
	}

	var lines []dto.JournalLineRequest

	for _, rule := range profile.Rules {
		coaID, err := e.resolveRuleCOA(ctx, rule, data)
		if err != nil {
			return nil, fmt.Errorf("accounting engine: resolve COA for rule (setting=%s, source=%s): %w", rule.COASettingKey, rule.COASource, err)
		}

		amount := e.resolveAmount(rule, data)
		if amount == 0 {
			continue
		}
		memo := e.resolveMemo(rule, data)

		var debit, credit float64
		switch rule.Side {
		case "debit":
			debit = amount
		case "credit":
			credit = amount
		case "dynamic":
			// Dynamic side — determined by caller in the TransactionData
			// For now, default to debit for positive, credit for negative
			if amount >= 0 {
				debit = amount
			} else {
				credit = -amount
			}
		default:
			return nil, fmt.Errorf("accounting engine: unknown side '%s'", rule.Side)
		}

		lines = append(lines, dto.JournalLineRequest{
			ChartOfAccountID: coaID,
			Debit:            debit,
			Credit:           credit,
			Memo:             memo,
		})
	}

	// Append line items if present (for multi-line transactions like cash bank, payment)
	for _, item := range data.LineItems {
		var debit, credit float64

		// Determine side based on the profile's first rule side:
		// If the main rule is debit for total (cash_in), line items are credit
		// If the main rule is credit for total (cash_out), line items are debit
		if len(profile.Rules) > 0 {
			switch profile.Rules[0].Side {
			case "debit":
				credit = item.Amount
			case "credit":
				debit = item.Amount
			default:
				credit = item.Amount
			}
		} else {
			credit = item.Amount
		}

		lines = append(lines, dto.JournalLineRequest{
			ChartOfAccountID: item.ChartOfAccountID,
			Debit:            debit,
			Credit:           credit,
			Memo:             strings.TrimSpace(item.Memo),
		})
	}

	// Build description
	desc := data.Description
	if profile.DescriptionTemplate != "" && len(data.DescriptionArgs) > 0 {
		desc = fmt.Sprintf(profile.DescriptionTemplate, data.DescriptionArgs...)
	}

	var fiscalYearID *string
	if data.FiscalYearID != nil {
		trimmedFiscalYearID := strings.TrimSpace(*data.FiscalYearID)
		if trimmedFiscalYearID != "" {
			fiscalYearID = &trimmedFiscalYearID
		}
	}

	journalTypeValue := strings.TrimSpace(string(profile.JournalType))

	refType := data.ReferenceType
	req := &dto.CreateJournalEntryRequest{
		CompanyID:         strings.TrimSpace(data.CompanyID),
		FiscalYearID:      fiscalYearID,
		EntryDate:         data.EntryDate,
		Description:       desc,
		ReferenceType:     &refType,
		ReferenceID:       &data.ReferenceID,
		JournalType:       &journalTypeValue,
		Lines:             lines,
		IsSystemGenerated: true,
	}

	return req, nil
}

func (e *accountingEngine) ResolveCOAID(ctx context.Context, settingKey string) (string, error) {
	coaCode, err := e.resolveCOACode(ctx, settingKey)
	if err != nil {
		return "", err
	}
	coa, err := e.coaRepo.FindByCode(ctx, coaCode)
	if err != nil {
		return "", fmt.Errorf("COA with code '%s' for setting '%s' not found: %w", coaCode, settingKey, err)
	}
	if !coa.IsPostable {
		return "", fmt.Errorf("COA with code '%s' for setting '%s' is non-postable", coaCode, settingKey)
	}
	if !coa.IsActive {
		return "", fmt.Errorf("COA with code '%s' for setting '%s' is inactive", coaCode, settingKey)
	}
	return coa.ID, nil
}

func (e *accountingEngine) GetAccountBalance(ctx context.Context, coaID string, asOf time.Time) (float64, error) {
	var balance float64
	// We query the database directly here to avoid circular dependency with JournalEntryUsecase
	// AccountingEngine is a low-level service that can talk to DB for primitive queries.
	err := e.coaRepo.GetDB(ctx).Table("journal_lines jl").
		Select("COALESCE(SUM(jl.debit - jl.credit), 0)").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where("je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date <= ?",
			financeModels.JournalStatusPosted, coaID, asOf.Format("2006-01-02")).
		Scan(&balance).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate account balance: %w", err)
	}
	return balance, nil
}

func (e *accountingEngine) GetAccountBalanceInRange(ctx context.Context, coaID string, fromDate, toDate time.Time) (float64, error) {
	if fromDate.After(toDate) {
		return 0, fmt.Errorf("invalid date range: from_date is after to_date")
	}

	var balance float64
	err := e.coaRepo.GetDB(ctx).Table("journal_lines jl").
		Select("COALESCE(SUM(jl.debit - jl.credit), 0)").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where("je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date BETWEEN ? AND ?",
			financeModels.JournalStatusPosted,
			coaID,
			fromDate.Format("2006-01-02"),
			toDate.Format("2006-01-02")).
		Scan(&balance).Error
	if err != nil {
		return 0, fmt.Errorf("failed to calculate account balance in range: %w", err)
	}

	return balance, nil
}

func (e *accountingEngine) resolveRuleCOA(ctx context.Context, rule PostingRule, data TransactionData) (string, error) {
	if rule.UseTransactionCOA {
		if data.TransactionCOAID == "" {
			return "", fmt.Errorf("transaction COA ID is required by profile but not provided")
		}
		return data.TransactionCOAID, nil
	}

	if rule.COASettingKey != "" {
		// Resolve from settings
		coaCode, err := e.resolveCOACode(ctx, rule.COASettingKey)
		if err != nil {
			return "", err
		}
		// Look up COA by code to get the ID
		coa, err := e.coaRepo.FindByCode(ctx, coaCode)
		if err != nil {
			return "", fmt.Errorf("COA with code '%s' (from setting '%s') not found: %w", coaCode, rule.COASettingKey, err)
		}
		if !coa.IsPostable {
			return "", fmt.Errorf("COA with code '%s' (from setting '%s') is non-postable", coaCode, rule.COASettingKey)
		}
		if !coa.IsActive {
			return "", fmt.Errorf("COA with code '%s' (from setting '%s') is inactive", coaCode, rule.COASettingKey)
		}
		return coa.ID, nil
	}

	switch rule.COASource {
	case "transaction":
		if data.TransactionCOAID == "" {
			return "", fmt.Errorf("transaction COA ID is required but not provided")
		}
		coa, err := e.coaRepo.FindByID(ctx, data.TransactionCOAID)
		if err != nil {
			return "", fmt.Errorf("transaction COA ID '%s' not found: %w", data.TransactionCOAID, err)
		}
		if !coa.IsPostable {
			return "", fmt.Errorf("transaction COA ID '%s' is non-postable", data.TransactionCOAID)
		}
		if !coa.IsActive {
			return "", fmt.Errorf("transaction COA ID '%s' is inactive", data.TransactionCOAID)
		}
		return data.TransactionCOAID, nil
	case "bank_account":
		if data.BankAccountCOAID == "" {
			return "", fmt.Errorf("bank account COA ID is required but not provided")
		}
		coa, err := e.coaRepo.FindByID(ctx, data.BankAccountCOAID)
		if err != nil {
			return "", fmt.Errorf("bank account COA ID '%s' not found: %w", data.BankAccountCOAID, err)
		}
		if !coa.IsPostable {
			return "", fmt.Errorf("bank account COA ID '%s' is non-postable", data.BankAccountCOAID)
		}
		if !coa.IsActive {
			return "", fmt.Errorf("bank account COA ID '%s' is inactive", data.BankAccountCOAID)
		}
		return data.BankAccountCOAID, nil
	case "payment_account":
		if data.PaymentAccountCOAID == "" {
			return "", fmt.Errorf("payment account COA ID is required but not provided")
		}
		coa, err := e.coaRepo.FindByID(ctx, data.PaymentAccountCOAID)
		if err != nil {
			return "", fmt.Errorf("payment account COA ID '%s' not found: %w", data.PaymentAccountCOAID, err)
		}
		if !coa.IsPostable {
			return "", fmt.Errorf("payment account COA ID '%s' is non-postable", data.PaymentAccountCOAID)
		}
		if !coa.IsActive {
			return "", fmt.Errorf("payment account COA ID '%s' is inactive", data.PaymentAccountCOAID)
		}
		return data.PaymentAccountCOAID, nil
	default:
		return "", fmt.Errorf("unknown COA source: '%s'", rule.COASource)
	}
}

func (e *accountingEngine) resolveCOACode(ctx context.Context, settingKey string) (string, error) {
	if e.settingsService == nil {
		return "", fmt.Errorf("settings service is not configured")
	}

	mappingKeys := make([]string, 0, 2)
	if mapped := mapLegacySettingKeyToSystemMapping(settingKey); mapped != "" {
		mappingKeys = append(mappingKeys, mapped)
	}
	mappingKeys = append(mappingKeys, settingKey)

	seen := make(map[string]struct{}, len(mappingKeys))
	for _, key := range mappingKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		coaCode, err := e.settingsService.GetCOAByKey(ctx, key)
		if err == nil && strings.TrimSpace(coaCode) != "" {
			return strings.TrimSpace(coaCode), nil
		}
	}

	// Backward-compatible fallback to finance_settings.
	return e.settingsService.GetCOACode(ctx, settingKey)
}

func mapLegacySettingKeyToSystemMapping(settingKey string) string {
	switch strings.TrimSpace(settingKey) {
	case "coa.inventory", "coa.inventory_asset":
		return "purchase.inventory_asset"
	case "coa.gr_ir", "coa.purchase_gr_ir":
		return "purchase.gr_ir_clearing"
	case "coa.purchase_vat_in", "coa.vat_in":
		return "purchase.tax_input"
	case "coa.purchase_payable", "coa.accounts_payable":
		return "purchase.accounts_payable"
	case "coa.sales_receivable":
		return "sales.accounts_receivable"
	case "coa.sales_revenue":
		return "sales.revenue"
	case "coa.sales_vat_out":
		return "sales.tax_output"
	case "coa.sales_cogs", "coa.cogs":
		return "sales.cogs"
	case "coa.sales_return":
		return "sales.sales_return"
	case "coa.inventory_gain":
		return "inventory.adjustment_gain"
	case "coa.inventory_loss":
		return "inventory.adjustment_loss"
	case "coa.depreciation_accumulated":
		return "asset.accumulated_depreciation"
	case "coa.depreciation_expense":
		return "asset.depreciation_expense"
	case "coa.bank":
		return "finance.bank_default"
	case "coa.cash":
		return "finance.cash_default"
	default:
		return ""
	}
}

// resolveAmount determines the amount for a posting rule.
func (e *accountingEngine) resolveAmount(rule PostingRule, data TransactionData) float64 {
	switch rule.AmountSource {
	case "total":
		return data.TotalAmount
	case "sub_total":
		return data.SubTotal
	case "tax_total":
		return data.TaxTotal
	case "cogs_total":
		return data.COGSTotal
	case "deposit_total":
		return data.DepositTotal
	case "net_total":
		return data.TotalAmount - data.DepositTotal
	case "other_total":
		return data.OtherTotal
	case "gr_ir_variance":
		return data.GRIRVarianceTotal
	case "calculated":
		// For dynamic calculations (e.g. period closing), the amount is in TotalAmount
		return data.TotalAmount
	default:
		return data.TotalAmount
	}
}

// resolveMemo builds the memo text for a posting rule.
func (e *accountingEngine) resolveMemo(rule PostingRule, data TransactionData) string {
	if rule.MemoTemplate == "" {
		return ""
	}
	if len(data.MemoArgs) > 0 && strings.Contains(rule.MemoTemplate, "%s") {
		return fmt.Sprintf(rule.MemoTemplate, data.MemoArgs...)
	}
	return rule.MemoTemplate
}

// validateProfileCOAs ensures all COA settings required by the posting profile are configured.
// Fails fast with a clear error message listing missing settings.
func (e *accountingEngine) validateProfileCOAs(ctx context.Context, profile PostingProfile) error {
	// Extract all COA setting keys from the posting profile rules
	var requiredKeys []string
	for _, rule := range profile.Rules {
		// Only validate rules that reference settings (not user-provided COAs)
		if rule.COASettingKey != "" {
			requiredKeys = append(requiredKeys, rule.COASettingKey)
		}
	}

	// If no settings keys required, no validation needed
	if len(requiredKeys) == 0 {
		return nil
	}

	// Validate all required settings exist and have values
	return e.coaValidationSvc.ValidateRequiredSettings(ctx, requiredKeys...)
}
