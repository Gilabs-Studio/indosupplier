package usecase_test

import (
	"context"
	"testing"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestDivisionUsecase_GetByID(t *testing.T) {
	type fields struct {
		divisionRepo *mocks.DivisionRepository
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
		want    *dto.DivisionResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				divisionRepo: mocks.NewDivisionRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "div-1",
			},
			mock: func(f fields) {
				div := &models.Division{
					ID:   "div-1",
					Name: "IT Division",
				}
				f.divisionRepo.On("FindByID", mock.Anything, "div-1").Return(div, nil)
			},
			want: &dto.DivisionResponse{
				ID:   "div-1",
				Name: "IT Division",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				divisionRepo: mocks.NewDivisionRepository(t),
			},
			args: args{
				ctx: context.Background(),
				id:  "div-99",
			},
			mock: func(f fields) {
				f.divisionRepo.On("FindByID", mock.Anything, "div-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrDivisionNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewDivisionUsecase(tt.fields.divisionRepo)

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

func TestDivisionUsecase_Create(t *testing.T) {
	type fields struct {
		divisionRepo *mocks.DivisionRepository
	}
	type args struct {
		ctx context.Context
		req *dto.CreateDivisionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.DivisionResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				divisionRepo: mocks.NewDivisionRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateDivisionRequest{
					Name: "New Division",
				},
			},
			mock: func(f fields) {
				f.divisionRepo.On("FindByName", mock.Anything, "New Division").Return(nil, gorm.ErrRecordNotFound)
				f.divisionRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *models.Division) bool {
					return d.Name == "New Division"
				})).Return(nil)
			},
			want: &dto.DivisionResponse{
				Name: "New Division",
			},
			wantErr: false,
		},
		{
			name: "Name Exists",
			fields: fields{
				divisionRepo: mocks.NewDivisionRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateDivisionRequest{
					Name: "Existing Div",
				},
			},
			mock: func(f fields) {
				f.divisionRepo.On("FindByName", mock.Anything, "Existing Div").Return(&models.Division{ID: "exist-id"}, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewDivisionUsecase(tt.fields.divisionRepo)

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
			}
		})
	}
}
