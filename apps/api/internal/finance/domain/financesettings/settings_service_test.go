package financesettings

import (
	"context"
	"errors"
	"testing"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type fakeFinanceSettingRepo struct{}

func (f *fakeFinanceSettingRepo) GetByKey(ctx context.Context, key string) (string, error) {
	return "", errors.New("not found")
}

func (f *fakeFinanceSettingRepo) FindByKey(ctx context.Context, key string) (*financeModels.FinanceSetting, error) {
	return nil, nil
}

func (f *fakeFinanceSettingRepo) GetAll(ctx context.Context) ([]financeModels.FinanceSetting, error) {
	return nil, nil
}

func (f *fakeFinanceSettingRepo) Upsert(ctx context.Context, key, value, description, category string) error {
	return nil
}

type fakeSystemAccountMappingRepo struct {
	data map[string]string
}

func (f *fakeSystemAccountMappingRepo) GetByKey(ctx context.Context, key string, companyID *string) (string, error) {
	if v, ok := f.data[key]; ok {
		return v, nil
	}
	return "", errors.New("not found")
}

func (f *fakeSystemAccountMappingRepo) GetMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	if v, ok := f.data[key]; ok {
		return &financeModels.SystemAccountMapping{Key: key, COACode: v, CompanyID: companyID}, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeSystemAccountMappingRepo) GetExactMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	return f.GetMappingByKey(ctx, key, companyID)
}

func (f *fakeSystemAccountMappingRepo) ListMappings(ctx context.Context, companyID *string) ([]financeModels.SystemAccountMapping, error) {
	result := make([]financeModels.SystemAccountMapping, 0, len(f.data))
	for key, code := range f.data {
		result = append(result, financeModels.SystemAccountMapping{Key: key, COACode: code, CompanyID: companyID})
	}
	return result, nil
}

func (f *fakeSystemAccountMappingRepo) DeleteByKey(ctx context.Context, key string, companyID *string) error {
	delete(f.data, key)
	return nil
}

func (f *fakeSystemAccountMappingRepo) Upsert(ctx context.Context, mapping *financeModels.SystemAccountMapping) error {
	if f.data == nil {
		f.data = make(map[string]string)
	}
	f.data[mapping.Key] = mapping.COACode
	return nil
}

func TestSettingsService_GetCOAByKey_ShouldReturnMappedCode(t *testing.T) {
	svc := NewSettingsService(&fakeFinanceSettingRepo{}, &fakeSystemAccountMappingRepo{
		data: map[string]string{
			"finance.cash_default": "1-1101",
		},
	})

	code, err := svc.GetCOAByKey(context.Background(), "finance.cash_default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != "1-1101" {
		t.Fatalf("expected mapped code 1-1101, got %s", code)
	}
}

func TestSettingsService_GetCOAByKey_ShouldReturnDescriptiveErrorWhenMissing(t *testing.T) {
	svc := NewSettingsService(&fakeFinanceSettingRepo{}, &fakeSystemAccountMappingRepo{
		data: map[string]string{},
	})

	_, err := svc.GetCOAByKey(context.Background(), "purchase.inventory_asset")
	if err == nil {
		t.Fatalf("expected error when mapping is missing")
	}
	if err.Error() != "system account mapping untuk 'purchase.inventory_asset' belum dikonfigurasi" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}
