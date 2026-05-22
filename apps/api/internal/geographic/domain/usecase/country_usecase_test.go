package usecase_test

import (
	"context"
	"testing"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase"
	"github.com/gilabs/gims/api/internal/geographic/domain/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCountryUsecase_GetByID(t *testing.T) {
	type fields struct {
		countryRepo *mocks.CountryRepository
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
		want    *dto.CountryResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				countryRepo: mocks.NewCountryRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "country-1",
			},
			mock: func(f fields) {
				country := &models.Country{
					ID:        "country-1",
					Name:      "Test Country",
					Code:      "TC",
					PhoneCode: "+62",
				}
				f.countryRepo.On("FindByID", mock.Anything, "country-1").Return(country, nil)
			},
			want: &dto.CountryResponse{
				ID:        "country-1",
				Name:      "Test Country",
				Code:      "TC",
				PhoneCode: "+62",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				countryRepo: mocks.NewCountryRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "country-99",
			},
			mock: func(f fields) {
				f.countryRepo.On("FindByID", mock.Anything, "country-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrCountryNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewCountryUsecase(tt.fields.countryRepo)

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

func TestCountryUsecase_Create(t *testing.T) {
	type fields struct {
		countryRepo *mocks.CountryRepository
	}
	type args struct {
		ctx context.Context
		req *dto.CreateCountryRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.CountryResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				countryRepo: mocks.NewCountryRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateCountryRequest{
					Name:      "New Country",
					Code:      "NC",
					PhoneCode: "+1",
				},
			},
			mock: func(f fields) {
				f.countryRepo.On("FindByCode", mock.Anything, "NC").Return(nil, gorm.ErrRecordNotFound)
				f.countryRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Country) bool {
					return c.Name == "New Country" && c.Code == "NC"
				})).Return(nil)
			},
			want: &dto.CountryResponse{
				Name:      "New Country",
				Code:      "NC",
				PhoneCode: "+1",
			},
			wantErr: false,
		},
		{
			name: "Code Exists",
			fields: fields{
				countryRepo: mocks.NewCountryRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateCountryRequest{
					Name:      "Existing Country",
					Code:      "EX",
					PhoneCode: "+1",
				},
			},
			mock: func(f fields) {
				f.countryRepo.On("FindByCode", mock.Anything, "EX").Return(&models.Country{ID: "existing-id"}, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewCountryUsecase(tt.fields.countryRepo)

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
