package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type inMemoryFiscalYearRepo struct {
	items   map[string]*financeModels.FiscalYear
	counter int
	hasTx   bool
}

type inMemoryOpeningBalanceRepo struct {
	hasPosted bool
}

func newInMemoryFiscalYearRepo(seed ...financeModels.FiscalYear) *inMemoryFiscalYearRepo {
	items := make(map[string]*financeModels.FiscalYear, len(seed))
	for i := range seed {
		item := seed[i]
		copied := item
		items[item.ID] = &copied
	}
	return &inMemoryFiscalYearRepo{items: items, counter: 100}
}

func (r *inMemoryFiscalYearRepo) Create(ctx context.Context, item *financeModels.FiscalYear) error {
	_ = ctx
	if item.ID == "" {
		r.counter++
		item.ID = fmt.Sprintf("fy-%d", r.counter)
	}
	copied := *item
	r.items[item.ID] = &copied
	return nil
}

func (r *inMemoryFiscalYearRepo) Update(ctx context.Context, item *financeModels.FiscalYear) error {
	_ = ctx
	if _, ok := r.items[item.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	copied := *item
	r.items[item.ID] = &copied
	return nil
}

func (r *inMemoryFiscalYearRepo) Delete(ctx context.Context, id string) error {
	_ = ctx
	if _, ok := r.items[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *inMemoryFiscalYearRepo) FindByID(ctx context.Context, id string) (*financeModels.FiscalYear, error) {
	_ = ctx
	item, ok := r.items[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *item
	return &copied, nil
}

func (r *inMemoryFiscalYearRepo) List(ctx context.Context, params repositories.FiscalYearListParams) ([]financeModels.FiscalYear, int64, error) {
	_ = ctx
	items := make([]financeModels.FiscalYear, 0)
	for _, item := range r.items {
		if item.CompanyID != params.CompanyID {
			continue
		}
		if params.Status != nil && item.Status != *params.Status {
			continue
		}
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (r *inMemoryFiscalYearRepo) FindActiveByCompany(ctx context.Context, companyID string) (*financeModels.FiscalYear, error) {
	_ = ctx
	for _, item := range r.items {
		if item.CompanyID == companyID && item.Status == financeModels.FiscalYearStatusActive {
			copied := *item
			return &copied, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *inMemoryFiscalYearRepo) HasPostedJournalInRange(ctx context.Context, companyID string, startDate, endDate time.Time) (bool, error) {
	_ = ctx
	_ = companyID
	_ = startDate
	_ = endDate
	return r.hasTx, nil
}

func (r *inMemoryFiscalYearRepo) GetDB(ctx context.Context) *gorm.DB {
	_ = ctx
	return nil
}

func (r *inMemoryOpeningBalanceRepo) ListLines(ctx context.Context, companyID, fiscalYearID string) ([]financeModels.OpeningBalanceLine, error) {
	_ = ctx
	_ = companyID
	_ = fiscalYearID
	return nil, nil
}

func (r *inMemoryOpeningBalanceRepo) ReplaceLines(ctx context.Context, companyID, fiscalYearID string, lines []financeModels.OpeningBalanceLine) error {
	_ = ctx
	_ = companyID
	_ = fiscalYearID
	_ = lines
	return nil
}

func (r *inMemoryOpeningBalanceRepo) DeleteLines(ctx context.Context, companyID, fiscalYearID string) error {
	_ = ctx
	_ = companyID
	_ = fiscalYearID
	return nil
}

func (r *inMemoryOpeningBalanceRepo) HasPostedOpeningJournal(ctx context.Context, companyID, fiscalYearID string) (bool, error) {
	_ = ctx
	_ = companyID
	_ = fiscalYearID
	return r.hasPosted, nil
}

func (r *inMemoryOpeningBalanceRepo) GetPostedOpeningJournalID(ctx context.Context, companyID, fiscalYearID string) (*string, error) {
	_ = ctx
	_ = companyID
	_ = fiscalYearID
	return nil, nil
}

func (r *inMemoryOpeningBalanceRepo) HasPostedOperationalJournalInRange(ctx context.Context, startDate, endDate string) (bool, error) {
	_ = ctx
	_ = startDate
	_ = endDate
	return false, nil
}

func (r *inMemoryOpeningBalanceRepo) GetDB(ctx context.Context) *gorm.DB {
	_ = ctx
	return nil
}

func TestFiscalYearCreate_ShouldRejectInvalidRange(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo()
	uc := NewFiscalYearUsecase(nil, repo, nil)

	_, err := uc.Create(context.Background(), &dto.CreateFiscalYearRequest{
		CompanyID: "11111111-1111-1111-1111-111111111111",
		Name:      "FY 2026",
		StartDate: "2026-01-01",
		EndDate:   "2025-12-31",
	}, nil)

	require.ErrorIs(t, err, ErrFiscalYearInvalidRange)
}

func TestFiscalYearActivate_ShouldAllowOnlySingleActivePerCompany(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo(
		financeModels.FiscalYear{
			ID:        "fy-active",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2025",
			StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusActive,
		},
		financeModels.FiscalYear{
			ID:        "fy-draft",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2026",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusDraft,
		},
	)
	uc := NewFiscalYearUsecase(nil, repo, nil)

	_, err := uc.Activate(context.Background(), "fy-draft", "11111111-1111-1111-1111-111111111111")
	require.ErrorIs(t, err, ErrFiscalYearActiveAlreadyUsed)
}

func TestFiscalYearLock_ShouldRejectWhenPostedTransactionExists(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo(
		financeModels.FiscalYear{
			ID:        "fy-active",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2026",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusActive,
		},
	)
	openingBalanceRepo := &inMemoryOpeningBalanceRepo{hasPosted: true}
	repo.hasTx = true
	uc := NewFiscalYearUsecase(nil, repo, openingBalanceRepo)

	_, err := uc.Lock(context.Background(), "fy-active", "11111111-1111-1111-1111-111111111111")
	require.ErrorIs(t, err, ErrFiscalYearHasTransactions)
}

func TestFiscalYearLock_ShouldRequireOpeningBalanceBeforeLocking(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo(
		financeModels.FiscalYear{
			ID:        "fy-active",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2026",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusActive,
		},
	)
	openingBalanceRepo := &inMemoryOpeningBalanceRepo{hasPosted: false}
	uc := NewFiscalYearUsecase(nil, repo, openingBalanceRepo)

	_, err := uc.Lock(context.Background(), "fy-active", "11111111-1111-1111-1111-111111111111")
	require.ErrorIs(t, err, ErrFiscalYearOpeningBalanceNotPosted)
}

func TestFiscalYearDelete_ShouldDeleteDraftOnly(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo(
		financeModels.FiscalYear{
			ID:        "fy-draft",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2026",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusDraft,
		},
		financeModels.FiscalYear{
			ID:        "fy-active",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2025",
			StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusActive,
		},
	)

	uc := NewFiscalYearUsecase(nil, repo, nil)

	err := uc.Delete(context.Background(), "fy-draft", "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)
	_, exists := repo.items["fy-draft"]
	require.False(t, exists)

	err = uc.Delete(context.Background(), "fy-active", "11111111-1111-1111-1111-111111111111")
	require.ErrorIs(t, err, ErrFiscalYearDeleteNotDraft)
}

func TestFiscalYearDelete_ShouldRejectCrossCompany(t *testing.T) {
	t.Parallel()

	repo := newInMemoryFiscalYearRepo(
		financeModels.FiscalYear{
			ID:        "fy-draft",
			CompanyID: "11111111-1111-1111-1111-111111111111",
			Name:      "FY 2026",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    financeModels.FiscalYearStatusDraft,
		},
	)

	uc := NewFiscalYearUsecase(nil, repo, nil)
	err := uc.Delete(context.Background(), "fy-draft", "22222222-2222-2222-2222-222222222222")
	require.ErrorIs(t, err, ErrFiscalYearCompanyMismatch)

	_, exists := repo.items["fy-draft"]
	require.True(t, exists)
}
