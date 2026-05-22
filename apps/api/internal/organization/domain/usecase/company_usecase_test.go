package usecase_test

import (
	"context"
	"testing"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	orgService "github.com/gilabs/gims/api/internal/organization/domain/service"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase/mocks"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	warehouseRepos "github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type noopOutletRepository struct{}

func (n *noopOutletRepository) Create(_ context.Context, _ *models.Outlet) error {
	return nil
}

func (n *noopOutletRepository) GetByID(_ context.Context, _ string) (*models.Outlet, error) {
	return nil, gorm.ErrRecordNotFound
}

func (n *noopOutletRepository) GetByCode(_ context.Context, _ string) (*models.Outlet, error) {
	return nil, gorm.ErrRecordNotFound
}

func (n *noopOutletRepository) GetNextCode(_ context.Context) (string, error) {
	return "", nil
}

func (n *noopOutletRepository) List(_ context.Context, _ orgRepos.OutletListParams) ([]*models.Outlet, int64, error) {
	return []*models.Outlet{}, 0, nil
}

func (n *noopOutletRepository) Update(_ context.Context, _ *models.Outlet) error {
	return nil
}

func (n *noopOutletRepository) Delete(_ context.Context, _ string) error {
	return nil
}

func (n *noopOutletRepository) FindByCompanyID(_ context.Context, _ string) ([]*models.Outlet, error) {
	return []*models.Outlet{}, nil
}

func (n *noopOutletRepository) UpdateIsActiveByCompanyID(_ context.Context, _ string, _ bool) error {
	return nil
}

func (n *noopOutletRepository) FindByWarehouseIDs(_ context.Context, _ []string) ([]*models.Outlet, error) {
	return []*models.Outlet{}, nil
}

type noopWarehouseRepository struct{}

func (n *noopWarehouseRepository) Create(_ context.Context, _ *warehouseModels.Warehouse) error {
	return nil
}

func (n *noopWarehouseRepository) GetByID(_ context.Context, _ string) (*warehouseModels.Warehouse, error) {
	return nil, gorm.ErrRecordNotFound
}

func (n *noopWarehouseRepository) GetByCode(_ context.Context, _ string) (*warehouseModels.Warehouse, error) {
	return nil, gorm.ErrRecordNotFound
}

func (n *noopWarehouseRepository) GetNextCode(_ context.Context) (string, error) {
	return "", nil
}

func (n *noopWarehouseRepository) List(_ context.Context, _ warehouseRepos.WarehouseListParams) ([]*warehouseModels.Warehouse, int64, error) {
	return []*warehouseModels.Warehouse{}, 0, nil
}

func (n *noopWarehouseRepository) Update(_ context.Context, _ *warehouseModels.Warehouse) error {
	return nil
}

func (n *noopWarehouseRepository) Delete(_ context.Context, _ string) error {
	return nil
}

func (n *noopWarehouseRepository) HasActiveStock(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (n *noopWarehouseRepository) FindByOutletIDs(_ context.Context, _ []string) ([]*warehouseModels.Warehouse, error) {
	return []*warehouseModels.Warehouse{}, nil
}

func (n *noopWarehouseRepository) UpdateIsActiveByOutletIDs(_ context.Context, _ []string, _ bool) error {
	return nil
}

type stubTimezoneService struct{}

func (s *stubTimezoneService) DetectTimezoneFromCoordinates(_ context.Context, _, _ float64) (*coreModels.TimezoneInfo, error) {
	return &coreModels.TimezoneInfo{ZoneName: "Asia/Jakarta"}, nil
}

func (s *stubTimezoneService) ValidateTimezone(_ context.Context, _ string) error {
	return nil
}

func (s *stubTimezoneService) GetTimezoneForCompany(_ context.Context, _, _ *float64, currentTimezone string) (string, error) {
	if currentTimezone != "" {
		return currentTimezone, nil
	}
	return "Asia/Jakarta", nil
}

type companyRepositoryAdapter struct {
	*mocks.CompanyRepository
}

func (a *companyRepositoryAdapter) FindAll(_ context.Context) ([]models.Company, error) {
	return []models.Company{}, nil
}

func newCompanyUsecaseForTest(companyRepo *mocks.CompanyRepository) usecase.CompanyUsecase {
	adaptedCompanyRepo := &companyRepositoryAdapter{CompanyRepository: companyRepo}
	var outletRepo orgRepos.OutletRepository = &noopOutletRepository{}
	var warehouseRepo warehouseRepos.WarehouseRepository = &noopWarehouseRepository{}
	var timezoneService orgService.TimezoneService = &stubTimezoneService{}

	return usecase.NewCompanyUsecase(adaptedCompanyRepo, outletRepo, warehouseRepo, timezoneService)
}

func TestCompanyUsecase_GetByID(t *testing.T) {
	type fields struct {
		companyRepo *mocks.CompanyRepository
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.CompanyResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				companyRepo: mocks.NewCompanyRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "company-1",
			},
			mock: func(f fields) {
				company := &models.Company{
					ID:   "company-1",
					Name: "Test Company",
				}
				f.companyRepo.On("FindByIDWithVillage", mock.Anything, "company-1").Return(company, nil)
			},
			want: &dto.CompanyResponse{
				ID:   "company-1",
				Name: "Test Company",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				companyRepo: mocks.NewCompanyRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "company-99",
			},
			mock: func(f fields) {
				f.companyRepo.On("FindByIDWithVillage", mock.Anything, "company-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrCompanyNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newCompanyUsecaseForTest(tt.fields.companyRepo)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.GetByID(tt.args.ctx, tt.args.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Name, got.Name)
			}
		})
	}
}

func TestCompanyUsecase_Create(t *testing.T) {
	type fields struct {
		companyRepo *mocks.CompanyRepository
	}
	type args struct {
		ctx       context.Context
		req       *dto.CreateCompanyRequest
		createdBy *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.CompanyResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				companyRepo: mocks.NewCompanyRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateCompanyRequest{
					Name: "New Company",
				},
				createdBy: nil,
			},
			mock: func(f fields) {
				f.companyRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
					return c.Name == "New Company"
				})).Return(nil)
				f.companyRepo.On("FindByIDWithVillage", mock.Anything, mock.Anything).Return(&models.Company{
					ID:   "new-id",
					Name: "New Company",
				}, nil)
			},
			want: &dto.CompanyResponse{
				Name: "New Company",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newCompanyUsecaseForTest(tt.fields.companyRepo)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Create(tt.args.ctx, tt.args.req, tt.args.createdBy)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Name, got.Name)
			}
		})
	}
}

func TestCompanyUsecase_Approve(t *testing.T) {
	type fields struct {
		companyRepo *mocks.CompanyRepository
	}
	type args struct {
		ctx        context.Context
		id         string
		req        *dto.ApproveCompanyRequest
		approvedBy string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.CompanyResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success Approve",
			fields: fields{
				companyRepo: mocks.NewCompanyRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "company-1",
				req: &dto.ApproveCompanyRequest{
					Action: "approve",
				},
				approvedBy: "admin",
			},
			mock: func(f fields) {
				f.companyRepo.On("FindByID", mock.Anything, "company-1").Return(&models.Company{
					ID:     "company-1",
					Status: models.CompanyStatusPending,
				}, nil)
				f.companyRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
					return c.Status == models.CompanyStatusApproved && c.IsApproved == true
				})).Return(nil)
				f.companyRepo.On("FindByIDWithVillage", mock.Anything, "company-1").Return(&models.Company{
					ID:         "company-1",
					Status:     models.CompanyStatusApproved,
					IsApproved: true,
				}, nil)
			},
			want: &dto.CompanyResponse{
				ID:         "company-1",
				Status:     string(models.CompanyStatusApproved),
				IsApproved: true,
			},
			wantErr: false,
		},
		{
			name: "Not Pending",
			fields: fields{
				companyRepo: mocks.NewCompanyRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "company-1",
				req: &dto.ApproveCompanyRequest{
					Action: "approve",
				},
				approvedBy: "admin",
			},
			mock: func(f fields) {
				f.companyRepo.On("FindByID", mock.Anything, "company-1").Return(&models.Company{
					ID:     "company-1",
					Status: models.CompanyStatusApproved,
				}, nil)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrCompanyNotPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newCompanyUsecaseForTest(tt.fields.companyRepo)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Approve(tt.args.ctx, tt.args.id, tt.args.req, tt.args.approvedBy)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.IsApproved, got.IsApproved)
			}
		})
	}
}
