package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JournalEntryListParams struct {
	Search         string
	Status         *financeModels.JournalStatus
	CompanyID      string
	FiscalYearID   *string
	JournalType    *financeModels.JournalType
	StartDate      *time.Time
	EndDate        *time.Time
	SortBy         string
	SortDir        string
	Limit          int
	Offset         int
	ReferenceType  *string
	ReferenceTypes []string
}

type JournalEntryRepository interface {
	FindByID(ctx context.Context, id string, withLines bool) (*financeModels.JournalEntry, error)
	List(ctx context.Context, params JournalEntryListParams) ([]financeModels.JournalEntry, int64, error)
	FindByReferenceID(ctx context.Context, referenceType string, referenceID string) (*financeModels.JournalEntry, error)
	ExistsByReference(ctx context.Context, referenceType string, referenceID string) (bool, error)
}

type journalEntryRepository struct {
	db *gorm.DB
}

func NewJournalEntryRepository(db *gorm.DB) JournalEntryRepository {
	return &journalEntryRepository{db: db}
}

func (r *journalEntryRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *journalEntryRepository) FindByID(ctx context.Context, id string, withLines bool) (*financeModels.JournalEntry, error) {
	var item financeModels.JournalEntry
	q := r.getDB(ctx)
	// Apply tenant + permission scope filtering
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if withLines {
		q = q.Preload("Lines").Preload("Lines.ChartOfAccount")
	}
	q = q.Preload("CreatedByUser").Preload("PostedByUser").Preload("ReversedByUser")
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var journalAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "journal_entries", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "journal_entries", Name: "updated_at"},
	},
	"entry_date": {
		Column: clause.Column{Table: "journal_entries", Name: "entry_date"},
	},
	"status": {
		Column: clause.Column{Table: "journal_entries", Name: "status"},
	},
}

func (r *journalEntryRepository) List(ctx context.Context, params JournalEntryListParams) ([]financeModels.JournalEntry, int64, error) {
	var items []financeModels.JournalEntry
	var total int64

	q := r.getDB(ctx).Model(&financeModels.JournalEntry{}).
		Preload("Lines").
		Preload("Lines.ChartOfAccount").
		Preload("CreatedByUser").
		Preload("PostedByUser").
		Preload("ReversedByUser")

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("journal_entries.description ILIKE ?", like)
	}
	if params.Status != nil {
		q = q.Where("journal_entries.status = ?", *params.Status)
	}
	if strings.TrimSpace(params.CompanyID) != "" {
		q = q.Where("journal_entries.company_id = ?", strings.TrimSpace(params.CompanyID))
	}
	if params.FiscalYearID != nil && strings.TrimSpace(*params.FiscalYearID) != "" {
		q = q.Where("journal_entries.fiscal_year_id = ?", strings.TrimSpace(*params.FiscalYearID))
	}
	if params.JournalType != nil {
		q = q.Where("journal_entries.journal_type = ?", *params.JournalType)
	}
	if params.StartDate != nil {
		q = q.Where("journal_entries.entry_date >= ?", params.StartDate.Format("2006-01-02"))
	}
	if params.EndDate != nil {
		q = q.Where("journal_entries.entry_date <= ?", params.EndDate.Format("2006-01-02"))
	}
	if params.ReferenceType != nil {
		q = q.Where("UPPER(journal_entries.reference_type) = UPPER(?)", *params.ReferenceType)
	}
	if len(params.ReferenceTypes) > 0 {
		types := make([]string, len(params.ReferenceTypes))
		for i, t := range params.ReferenceTypes {
			types[i] = strings.ToUpper(t)
		}
		q = q.Where("UPPER(journal_entries.reference_type) IN ?", types)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := journalAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = journalAllowedSort["entry_date"]
	}
	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	if sortDir == "asc" {
		sortCol.Desc = false
	} else {
		sortCol.Desc = true
	}
	q = q.Order(sortCol)

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

// FindByReferenceID returns the journal entry for a given reference type and ID.
// This supports idempotency - prevents duplicate journals for the same transaction.
func (r *journalEntryRepository) FindByReferenceID(ctx context.Context, referenceType string, referenceID string) (*financeModels.JournalEntry, error) {
	var item financeModels.JournalEntry
	q := r.getDB(ctx).
		Preload("Lines").
		Preload("Lines.ChartOfAccount").
		Preload("CreatedByUser").
		Preload("PostedByUser").
		Preload("ReversedByUser")

	// Apply tenant + permission scope filtering
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if err := q.Where("reference_type = ? AND reference_id = ?", strings.ToUpper(referenceType), referenceID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Explicitly return nil for no record found
		}
		return nil, err
	}
	return &item, nil
}

// ExistsByReference checks if a journal entry already exists for the given reference type and ID.
// Returns true if exists, false if not found.
func (r *journalEntryRepository) ExistsByReference(ctx context.Context, referenceType string, referenceID string) (bool, error) {
	var count int64
	q := r.getDB(ctx).Model(&financeModels.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", strings.ToUpper(referenceType), referenceID)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
