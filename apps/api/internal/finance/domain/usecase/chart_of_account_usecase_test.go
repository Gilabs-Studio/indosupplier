package usecase

import (
	"context"
	"fmt"
	"testing"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type inMemoryCOARepo struct {
	items   map[string]*financeModels.ChartOfAccount
	used    map[string]bool
	counter int
}

func newInMemoryCOARepo(seed ...financeModels.ChartOfAccount) *inMemoryCOARepo {
	items := make(map[string]*financeModels.ChartOfAccount, len(seed))
	for i := range seed {
		item := seed[i]
		copied := item
		items[item.ID] = &copied
	}
	return &inMemoryCOARepo{items: items, used: map[string]bool{}, counter: 100}
}

func (r *inMemoryCOARepo) Create(ctx context.Context, item *financeModels.ChartOfAccount) error {
	_ = ctx
	if item.ID == "" {
		r.counter++
		item.ID = fmt.Sprintf("coa-%d", r.counter)
	}
	copied := *item
	r.items[item.ID] = &copied
	return nil
}

func (r *inMemoryCOARepo) FindByID(ctx context.Context, id string) (*financeModels.ChartOfAccount, error) {
	_ = ctx
	item, ok := r.items[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *item
	return &copied, nil
}

func (r *inMemoryCOARepo) FindAll(ctx context.Context, onlyActive bool) ([]financeModels.ChartOfAccount, error) {
	_ = ctx
	out := make([]financeModels.ChartOfAccount, 0, len(r.items))
	for _, item := range r.items {
		if onlyActive && !item.IsActive {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

func (r *inMemoryCOARepo) List(ctx context.Context, params repositories.ChartOfAccountListParams) ([]financeModels.ChartOfAccount, int64, error) {
	items, err := r.FindAll(ctx, false)
	if err != nil {
		return nil, 0, err
	}
	return items, int64(len(items)), nil
}

func (r *inMemoryCOARepo) Update(ctx context.Context, item *financeModels.ChartOfAccount) error {
	_ = ctx
	if _, ok := r.items[item.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	copied := *item
	r.items[item.ID] = &copied
	return nil
}

func (r *inMemoryCOARepo) Delete(ctx context.Context, id string) error {
	_ = ctx
	delete(r.items, id)
	return nil
}

func (r *inMemoryCOARepo) ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error) {
	_ = ctx
	for id, item := range r.items {
		if item.Code != code {
			continue
		}
		if excludeID != nil && id == *excludeID {
			continue
		}
		return true, nil
	}
	return false, nil
}

func (r *inMemoryCOARepo) FindByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	_ = ctx
	for _, item := range r.items {
		if item.Code == code {
			copied := *item
			return &copied, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *inMemoryCOARepo) GetByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	return r.FindByCode(ctx, code)
}

func (r *inMemoryCOARepo) FindOpeningBalanceEquity(ctx context.Context) (*financeModels.ChartOfAccount, error) {
	return r.FindByCode(ctx, "39999")
}

func (r *inMemoryCOARepo) HasChildren(ctx context.Context, id string) (bool, error) {
	_ = ctx
	for _, item := range r.items {
		if item.ParentID != nil && *item.ParentID == id {
			return true, nil
		}
	}
	return false, nil
}

func (r *inMemoryCOARepo) IsUsedInJournal(ctx context.Context, id string) (bool, error) {
	_ = ctx
	return r.used[id], nil
}

func (r *inMemoryCOARepo) HasJournalLines(ctx context.Context, id string) (bool, error) {
	return r.IsUsedInJournal(ctx, id)
}

func (r *inMemoryCOARepo) UpdateIsPostable(ctx context.Context, id string, isPostable bool) error {
	_ = ctx
	item, ok := r.items[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	item.IsPostable = isPostable
	return nil
}

func (r *inMemoryCOARepo) RecalculateAllIsPostable(ctx context.Context) error {
	_ = ctx
	childCount := map[string]int{}
	for _, item := range r.items {
		if item.ParentID != nil && *item.ParentID != "" {
			childCount[*item.ParentID]++
		}
	}
	for _, item := range r.items {
		if item.ParentID == nil || *item.ParentID == "" {
			item.IsPostable = false
			continue
		}
		item.IsPostable = childCount[item.ID] == 0
	}
	return nil
}

func (r *inMemoryCOARepo) GetDB(ctx context.Context) *gorm.DB {
	_ = ctx
	return nil
}

func TestChartOfAccountCreate_ShouldSetParentNonPostable_WhenChildAdded(t *testing.T) {
	t.Parallel()

	rootID := "root"
	parentID := "parent"
	repo := newInMemoryCOARepo(
		financeModels.ChartOfAccount{ID: rootID, Code: "1-0000", Name: "Assets", Type: financeModels.AccountTypeAsset, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: parentID, Code: "1-1000", Name: "Current Assets", Type: financeModels.AccountTypeAsset, ParentID: &rootID, IsActive: true, IsPostable: true},
	)
	uc := NewChartOfAccountUsecase(nil, repo, mapper.NewChartOfAccountMapper())

	created, err := uc.Create(context.Background(), &dto.CreateChartOfAccountRequest{
		Code:     "1-1101",
		Name:     "Cash",
		Type:     financeModels.AccountTypeCashBank,
		ParentID: &parentID,
	})
	require.NoError(t, err)
	require.True(t, created.IsPostable)

	updatedParent, err := repo.FindByID(context.Background(), parentID)
	require.NoError(t, err)
	require.False(t, updatedParent.IsPostable)
}

func TestChartOfAccountDelete_ShouldSetParentPostable_WhenLastChildDeleted(t *testing.T) {
	t.Parallel()

	rootID := "root"
	parentID := "parent"
	childID := "child"
	repo := newInMemoryCOARepo(
		financeModels.ChartOfAccount{ID: rootID, Code: "1-0000", Name: "Assets", Type: financeModels.AccountTypeAsset, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: parentID, Code: "1-1000", Name: "Current Assets", Type: financeModels.AccountTypeAsset, ParentID: &rootID, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: childID, Code: "1-1101", Name: "Cash", Type: financeModels.AccountTypeCashBank, ParentID: &parentID, IsActive: true, IsPostable: true},
	)
	uc := NewChartOfAccountUsecase(nil, repo, mapper.NewChartOfAccountMapper())

	err := uc.Delete(context.Background(), childID)
	require.NoError(t, err)

	updatedParent, err := repo.FindByID(context.Background(), parentID)
	require.NoError(t, err)
	require.True(t, updatedParent.IsPostable)
}

func TestChartOfAccountUpdate_ShouldRejectCircularParentReference(t *testing.T) {
	t.Parallel()

	rootID := "root"
	aID := "a"
	bID := "b"
	repo := newInMemoryCOARepo(
		financeModels.ChartOfAccount{ID: rootID, Code: "1-0000", Name: "Assets", Type: financeModels.AccountTypeAsset, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: aID, Code: "1-1000", Name: "A", Type: financeModels.AccountTypeAsset, ParentID: &rootID, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: bID, Code: "1-1100", Name: "B", Type: financeModels.AccountTypeAsset, ParentID: &aID, IsActive: true, IsPostable: true},
	)
	uc := NewChartOfAccountUsecase(nil, repo, mapper.NewChartOfAccountMapper())

	_, err := uc.Update(context.Background(), aID, &dto.UpdateChartOfAccountRequest{
		Code:     "1-1000",
		Name:     "A",
		Type:     financeModels.AccountTypeAsset,
		ParentID: &bID,
	})
	require.ErrorIs(t, err, ErrCOAInvalidParent)
}

func TestChartOfAccountCreate_ShouldAutoGenerateCode_FromParent(t *testing.T) {
	t.Parallel()

	rootID := "root"
	parentID := "parent"
	repo := newInMemoryCOARepo(
		financeModels.ChartOfAccount{ID: rootID, Code: "10000", Name: "Assets", Type: financeModels.AccountTypeAsset, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: parentID, Code: "11000", Name: "Current Assets", Type: financeModels.AccountTypeAsset, ParentID: &rootID, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: "child-1", Code: "11100", Name: "Cash Group", Type: financeModels.AccountTypeAsset, ParentID: &parentID, IsActive: true, IsPostable: true},
		financeModels.ChartOfAccount{ID: "child-2", Code: "11120", Name: "Receivable Group", Type: financeModels.AccountTypeAsset, ParentID: &parentID, IsActive: true, IsPostable: true},
	)
	uc := NewChartOfAccountUsecase(nil, repo, mapper.NewChartOfAccountMapper())

	created, err := uc.Create(context.Background(), &dto.CreateChartOfAccountRequest{
		Code:     "manual-code-should-be-ignored",
		Name:     "Generated Child",
		Type:     financeModels.AccountTypeAsset,
		ParentID: &parentID,
	})
	require.NoError(t, err)
	require.Equal(t, "11121", created.Code)
}

func TestChartOfAccountCreate_ShouldKeepNonPostable_WhenRequested(t *testing.T) {
	t.Parallel()

	rootID := "root"
	parentID := "parent"
	repo := newInMemoryCOARepo(
		financeModels.ChartOfAccount{ID: rootID, Code: "10000", Name: "Assets", Type: financeModels.AccountTypeAsset, IsActive: true, IsPostable: false},
		financeModels.ChartOfAccount{ID: parentID, Code: "11100", Name: "Cash and Cash Equivalent", Type: financeModels.AccountTypeCashBank, ParentID: &rootID, IsActive: true, IsPostable: false},
	)
	uc := NewChartOfAccountUsecase(nil, repo, mapper.NewChartOfAccountMapper())

	isPostable := false
	created, err := uc.Create(context.Background(), &dto.CreateChartOfAccountRequest{
		Name:       "Petty Cash",
		Type:       financeModels.AccountTypeCashBank,
		ParentID:   &parentID,
		IsPostable: &isPostable,
	})
	require.NoError(t, err)
	require.False(t, created.IsPostable)
}

