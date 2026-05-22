package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JournalLineListParams holds query parameters for listing journal lines.
type JournalLineListParams struct {
	CashBankJournalID string
	ChartOfAccountID  string
	AccountType       string
	ReferenceType     *string
	JournalStatus     string
	StartDate         *time.Time
	EndDate           *time.Time
	Search            string
	SortBy            string
	SortDir           string
	Limit             int
	Offset            int
}

// JournalLineWithEntry is a flat struct for JOIN result between journal_lines and journal_entries.
type JournalLineWithEntry struct {
	financeModels.JournalLine

	EntryDate          time.Time                   `gorm:"column:entry_date"`
	JournalDescription string                      `gorm:"column:journal_description"`
	JournalStatus      financeModels.JournalStatus `gorm:"column:journal_status"`
	ReferenceType      *string                     `gorm:"column:reference_type"`
	ReferenceID        *string                     `gorm:"column:reference_id"`
}

// JournalLineRepository defines data access for journal lines (sub-ledger view).
type JournalLineRepository interface {
	// List returns journal lines with their parent entry context, paginated and filtered.
	List(ctx context.Context, params JournalLineListParams) ([]JournalLineWithEntry, int64, error)

	// SumBeforeDate returns cumulative debit and credit for a specific COA before a given date.
	// Used for calculating the opening balance when computing running balance.
	SumBeforeDate(ctx context.Context, coaID string, beforeDate time.Time, journalStatus string) (debit float64, credit float64, err error)
}

type journalLineRepository struct {
	db *gorm.DB
}

// NewJournalLineRepository creates a new JournalLineRepository.
func NewJournalLineRepository(db *gorm.DB) JournalLineRepository {
	return &journalLineRepository{db: db}
}

var journalLineAllowedSort = map[string]string{
	"entry_date": "je.entry_date",
	"created_at": "jl.created_at",
	"debit":      "jl.debit",
	"credit":     "jl.credit",
	"coa_code":   "jl.chart_of_account_code_snapshot",
	"coa_name":   "jl.chart_of_account_name_snapshot",
}

func (r *journalLineRepository) List(ctx context.Context, params JournalLineListParams) ([]JournalLineWithEntry, int64, error) {
	// If filtering by cash bank journal, use cash_bank_journal_lines directly (does not always have a journal_entry_id).
	if id := strings.TrimSpace(params.CashBankJournalID); id != "" {
		baseQuery := r.db.WithContext(ctx).
			Table("cash_bank_journal_lines AS cbl").
			Joins("JOIN cash_bank_journals cbj ON cbj.id = cbl.cash_bank_journal_id AND cbj.deleted_at IS NULL").
			Joins("LEFT JOIN chart_of_accounts coa ON coa.id = cbl.chart_of_account_id").
			Where("cbl.deleted_at IS NULL").
			Where("cbj.id = ?", id)

		// Apply scope-based data filtering (based on cash bank journal scope)
		// Ensure tenant qualification to avoid ambiguous tenant_id in JOINs
		baseQuery = applyQualifiedTenantFilter(ctx, baseQuery, "cbl.tenant_id", "cbj.tenant_id", "coa.tenant_id")
		baseQuery = security.ApplyScopeFilter(baseQuery, ctx, security.FinanceScopeQueryOptions())

		// Count total
		var total int64
		if err := baseQuery.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		// Sort
		sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
		isDesc := sortDir == "desc"

		selectFields := `
			cbl.id, cbl.cash_bank_journal_id AS journal_entry_id, cbl.chart_of_account_id,
			coa.code AS chart_of_account_code_snapshot, coa.name AS chart_of_account_name_snapshot,
			coa.type AS chart_of_account_type_snapshot,
			cbl.debit, cbl.credit, cbl.memo, cbl.created_at, cbl.updated_at,
			cbj.transaction_date AS entry_date, cbj.description AS journal_description,
			cbj.status AS journal_status, NULL::text AS reference_type, NULL::text AS reference_id
		`

		var items []JournalLineWithEntry
		q := baseQuery.Select(selectFields).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "cbj", Name: "transaction_date"}, Desc: isDesc}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "cbl", Name: "created_at"}, Desc: false}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "cbl", Name: "id"}, Desc: false})
		if params.Limit > 0 {
			q = q.Limit(params.Limit)
		}
		if params.Offset > 0 {
			q = q.Offset(params.Offset)
		}

		if err := q.Find(&items).Error; err != nil {
			return nil, 0, err
		}

		return items, total, nil
	}

	baseQuery := r.db.WithContext(ctx).
		Table("journal_lines AS jl").
		Joins("JOIN journal_entries AS je ON je.id = jl.journal_entry_id AND je.deleted_at IS NULL").
		Where("jl.deleted_at IS NULL")

	// Apply scope-based data filtering
	baseQuery = applyQualifiedTenantFilter(ctx, baseQuery, "jl.tenant_id", "je.tenant_id")
	baseQuery = security.ApplyScopeFilter(baseQuery, ctx, security.FinanceScopeQueryOptions())

	// Apply filters
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		baseQuery = baseQuery.Where(
			"(jl.chart_of_account_code_snapshot ILIKE ? OR jl.chart_of_account_name_snapshot ILIKE ? OR jl.memo ILIKE ?)",
			like, like, like,
		)
	}
	if id := strings.TrimSpace(params.ChartOfAccountID); id != "" {
		baseQuery = baseQuery.Where("jl.chart_of_account_id = ?", id)
	}
	if at := strings.TrimSpace(params.AccountType); at != "" {
		baseQuery = baseQuery.Where("jl.chart_of_account_type_snapshot = ?", at)
	}
	if params.ReferenceType != nil {
		if rt := strings.TrimSpace(*params.ReferenceType); rt != "" {
			baseQuery = baseQuery.Where("je.reference_type = ?", rt)
		}
	}
	if js := strings.TrimSpace(params.JournalStatus); js != "" {
		baseQuery = baseQuery.Where("je.status = ?", js)
	}
	if params.StartDate != nil {
		baseQuery = baseQuery.Where("je.entry_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		baseQuery = baseQuery.Where("je.entry_date <= ?", *params.EndDate)
	}

	// Count total
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort — default is entry_date ASC for running balance correctness
	sortCol := journalLineAllowedSort[params.SortBy]
	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	isDesc := sortDir == "desc"

	// Determine table and column for primary sort
	var primarySort clause.Column
	if sortCol != "" {
		parts := strings.Split(sortCol, ".")
		if len(parts) == 2 {
			primarySort = clause.Column{Table: parts[0], Name: parts[1]}
		} else {
			primarySort = clause.Column{Name: parts[0]}
		}
	} else {
		primarySort = clause.Column{Table: "je", Name: "entry_date"}
	}

	// Select fields
	selectFields := `
		jl.id, jl.journal_entry_id, jl.chart_of_account_id,
		jl.chart_of_account_code_snapshot, jl.chart_of_account_name_snapshot,
		jl.chart_of_account_type_snapshot,
		jl.debit, jl.credit, jl.memo, jl.created_at, jl.updated_at,
		je.entry_date, je.description AS journal_description,
		je.status AS journal_status, je.reference_type, je.reference_id
	`

	var items []JournalLineWithEntry
	q := baseQuery.Select(selectFields).
		Order(clause.OrderByColumn{Column: primarySort, Desc: isDesc}).
		Order(clause.OrderByColumn{Column: clause.Column{Table: "je", Name: "created_at"}, Desc: false}).
		Order(clause.OrderByColumn{Column: clause.Column{Table: "jl", Name: "id"}, Desc: false})

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *journalLineRepository) SumBeforeDate(ctx context.Context, coaID string, beforeDate time.Time, journalStatus string) (float64, float64, error) {
	type result struct {
		DebitSum  float64
		CreditSum float64
	}
	var res result

	q := r.db.WithContext(ctx).
		Table("journal_lines AS jl").
		Select("COALESCE(SUM(jl.debit), 0) AS debit_sum, COALESCE(SUM(jl.credit), 0) AS credit_sum").
		Joins("JOIN journal_entries AS je ON je.id = jl.journal_entry_id AND je.deleted_at IS NULL").
		Where("jl.deleted_at IS NULL").
		Where("jl.chart_of_account_id = ?", coaID).
		Where("je.entry_date < ?", beforeDate)

	if js := strings.TrimSpace(journalStatus); js != "" {
		q = q.Where("je.status = ?", js)
	}

	// Apply tenant qualification
	q = applyQualifiedTenantFilter(ctx, q, "jl.tenant_id", "je.tenant_id")
	// Apply permission-based scope filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if err := q.Scan(&res).Error; err != nil {
		return 0, 0, err
	}
	return res.DebitSum, res.CreditSum, nil
}
