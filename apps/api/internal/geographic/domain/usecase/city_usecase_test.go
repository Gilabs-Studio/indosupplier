package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCityUsecase_GetByID(t *testing.T) {
	type fields struct {
		cityRepo     *mocks.CityRepository
		provinceRepo *mocks.ProvinceRepository
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
		want    *dto.CityResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				cityRepo: mocks.NewCityRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "city-1",
			},
			mock: func(f fields) {
				city := &models.City{
					ID:   "city-1",
					Name: "Test City",
					Code: "TC",
				}
				f.cityRepo.On("FindByID", mock.Anything, "city-1").Return(city, nil)
			},
			want: &dto.CityResponse{
				ID:   "city-1",
				Name: "Test City",
				Code: "TC",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				cityRepo: mocks.NewCityRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "city-99",
			},
			mock: func(f fields) {
				f.cityRepo.On("FindByID", mock.Anything, "city-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrCityNotFound,
		},
		{
			name: "Internal Error",
			fields: fields{
				cityRepo: mocks.NewCityRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "city-error",
			},
			mock: func(f fields) {
				f.cityRepo.On("FindByID", mock.Anything, "city-error").Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
			err:     errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pRepo := mocks.NewProvinceRepository(t) // Unused in GetByID but required for constructor
			u := usecase.NewCityUsecase(tt.fields.cityRepo, pRepo)

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

func TestCityUsecase_Create(t *testing.T) {
	type fields struct {
		cityRepo     *mocks.CityRepository
		provinceRepo *mocks.ProvinceRepository
	}
	type args struct {
		ctx context.Context
		req *dto.CreateCityRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.CityResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				cityRepo:     mocks.NewCityRepository(t),
				provinceRepo: mocks.NewProvinceRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateCityRequest{
					Name:       "New City",
					Code:       "NC",
					ProvinceID: "prov-1",
				},
			},
			mock: func(f fields) {
				f.provinceRepo.On("FindByID", mock.Anything, "prov-1").Return(&models.Province{ID: "prov-1"}, nil)
				f.cityRepo.On("FindByCode", mock.Anything, "NC").Return(nil, gorm.ErrRecordNotFound)
				f.cityRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.City) bool {
					return c.Name == "New City" && c.Code == "NC"
				})).Return(nil)
				f.cityRepo.On("FindByID", mock.Anything, mock.Anything).Return(&models.City{
					ID:         "new-id",
					Name:       "New City",
					Code:       "NC",
					ProvinceID: "prov-1",
				}, nil)
			},
			want: &dto.CityResponse{
				Name: "New City",
				Code: "NC",
			},
			wantErr: false,
		},
		{
			name: "Province Not Found",
			fields: fields{
				cityRepo:     mocks.NewCityRepository(t),
				provinceRepo: mocks.NewProvinceRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateCityRequest{
					Name:       "New City",
					Code:       "NC",
					ProvinceID: "prov-99",
				},
			},
			mock: func(f fields) {
				f.provinceRepo.On("FindByID", mock.Anything, "prov-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewCityUsecase(tt.fields.cityRepo, tt.fields.provinceRepo)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Create(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Code, got.Code)
			}
		})
	}
}
