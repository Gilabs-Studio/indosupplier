package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

type fakeMappingRepo struct {
	rows map[string]financeModels.SystemAccountMapping
}

func (f *fakeMappingRepo) key(mappingKey string, companyID *string) string {
	if companyID == nil || *companyID == "" {
		return mappingKey + "::global"
	}
	return fmt.Sprintf("%s::%s", mappingKey, *companyID)
}

func (f *fakeMappingRepo) GetByKey(ctx context.Context, key string, companyID *string) (string, error) {
	m, err := f.GetMappingByKey(ctx, key, companyID)
	if err != nil {
		return "", err
	}
	return m.COACode, nil
}

func (f *fakeMappingRepo) GetMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	if companyID != nil && *companyID != "" {
		if row, ok := f.rows[f.key(key, companyID)]; ok {
			copy := row
			return &copy, nil
		}
	}

	if row, ok := f.rows[f.key(key, nil)]; ok {
		copy := row
		return &copy, nil
	}

	return nil, gorm.ErrRecordNotFound
}

func (f *fakeMappingRepo) GetExactMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	row, ok := f.rows[f.key(key, companyID)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := row
	return &copy, nil
}

func (f *fakeMappingRepo) ListMappings(ctx context.Context, companyID *string) ([]financeModels.SystemAccountMapping, error) {
	result := make([]financeModels.SystemAccountMapping, 0)
	if companyID == nil || *companyID == "" {
		for composite, row := range f.rows {
			if strings.HasSuffix(composite, "::global") {
				result = append(result, row)
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
		return result, nil
	}

	seen := map[string]bool{}
	for _, row := range f.rows {
		if row.CompanyID != nil && *row.CompanyID == *companyID {
			result = append(result, row)
			seen[row.Key] = true
		}
	}
	for _, row := range f.rows {
		if row.CompanyID == nil && !seen[row.Key] {
			result = append(result, row)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result, nil
}

func (f *fakeMappingRepo) DeleteByKey(ctx context.Context, key string, companyID *string) error {
	k := f.key(key, companyID)
	if _, ok := f.rows[k]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.rows, k)
	return nil
}

func (f *fakeMappingRepo) Upsert(ctx context.Context, mapping *financeModels.SystemAccountMapping) error {
	if f.rows == nil {
		f.rows = map[string]financeModels.SystemAccountMapping{}
	}
	k := f.key(mapping.Key, mapping.CompanyID)
	row := *mapping
	if existing, ok := f.rows[k]; ok {
		row.ID = existing.ID
	}
	if row.ID == "" {
		row.ID = "row-" + mapping.Key
	}
	f.rows[k] = row
	return nil
}

type fakeCOAUsecase struct {
	items map[string]dto.ChartOfAccountResponse
}

func (f *fakeCOAUsecase) Create(ctx context.Context, req *dto.CreateChartOfAccountRequest) (*dto.ChartOfAccountResponse, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeCOAUsecase) Update(ctx context.Context, id string, req *dto.UpdateChartOfAccountRequest) (*dto.ChartOfAccountResponse, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeCOAUsecase) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (f *fakeCOAUsecase) GetByID(ctx context.Context, id string) (*dto.ChartOfAccountResponse, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeCOAUsecase) List(ctx context.Context, req *dto.ListChartOfAccountsRequest) ([]dto.ChartOfAccountResponse, int64, error) {
	return nil, 0, errors.New("not implemented")
}
func (f *fakeCOAUsecase) Tree(ctx context.Context, onlyActive bool) ([]dto.ChartOfAccountTreeNode, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeCOAUsecase) GetByCode(ctx context.Context, code string) (*dto.ChartOfAccountResponse, error) {
	row, ok := f.items[code]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := row
	return &copy, nil
}
func (f *fakeCOAUsecase) RecalculateAllIsPostable(ctx context.Context) error {
	return nil
}

type fakeAuditLogger struct {
	changesCount int
}

func (f *fakeAuditLogger) Log(ctx context.Context, action string, targetID string, metadata map[string]interface{}) {
}
func (f *fakeAuditLogger) LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}) {
}
func (f *fakeAuditLogger) LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{}) {
	f.changesCount++
}
func (f *fakeAuditLogger) LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{}) {
	f.changesCount++
}

func TestSystemAccountMappingUsecase_ShouldRejectNonPostableAccount_WhenUpserting(t *testing.T) {
	repo := &fakeMappingRepo{rows: map[string]financeModels.SystemAccountMapping{}}
	coa := &fakeCOAUsecase{items: map[string]dto.ChartOfAccountResponse{
		"1-1000": {Code: "1-1000", IsPostable: false, IsActive: true},
	}}
	uc := NewSystemAccountMappingUsecase(repo, coa)

	_, err := uc.Upsert(context.Background(), "sales.revenue", "1-1000", "Sales Revenue", nil)
	if !errors.Is(err, ErrAccountNotPostable) {
		t.Fatalf("expected ErrAccountNotPostable, got %v", err)
	}
}

func TestSystemAccountMappingUsecase_ShouldFallbackToGlobal_WhenCompanyOverrideDeleted(t *testing.T) {
	companyID := "11111111-1111-1111-1111-111111111111"
	repo := &fakeMappingRepo{rows: map[string]financeModels.SystemAccountMapping{
		"sales.revenue::global": {
			ID:      "global-1",
			Key:     "sales.revenue",
			COACode: "4-1000",
			Label:   "Global Sales Revenue",
		},
		"sales.revenue::11111111-1111-1111-1111-111111111111": {
			ID:        "company-1",
			Key:       "sales.revenue",
			CompanyID: &companyID,
			COACode:   "4-1001",
			Label:     "Company Sales Revenue",
		},
	}}
	coa := &fakeCOAUsecase{items: map[string]dto.ChartOfAccountResponse{
		"4-1000": {ID: "coa-1", Code: "4-1000", Name: "Sales Revenue Global", IsPostable: true, IsActive: true},
		"4-1001": {ID: "coa-2", Code: "4-1001", Name: "Sales Revenue Company", IsPostable: true, IsActive: true},
	}}
	audit := &fakeAuditLogger{}
	uc := NewSystemAccountMappingUsecase(repo, coa, audit)

	beforeDelete, err := uc.GetByKey(context.Background(), "sales.revenue", &companyID)
	if err != nil {
		t.Fatalf("expected no error before delete, got %v", err)
	}
	if beforeDelete.COACode != "4-1001" {
		t.Fatalf("expected company override coa_code 4-1001, got %s", beforeDelete.COACode)
	}

	if err := uc.Delete(context.Background(), "sales.revenue", &companyID); err != nil {
		t.Fatalf("expected no error deleting company override, got %v", err)
	}
	if audit.changesCount == 0 {
		t.Fatalf("expected audit log for delete operation")
	}

	afterDelete, err := uc.GetByKey(context.Background(), "sales.revenue", &companyID)
	if err != nil {
		t.Fatalf("expected no error after delete fallback, got %v", err)
	}
	if afterDelete.COACode != "4-1000" {
		t.Fatalf("expected global fallback coa_code 4-1000, got %s", afterDelete.COACode)
	}
}
