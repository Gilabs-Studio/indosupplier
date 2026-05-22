package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	purchaseModels "github.com/gilabs/gims/api/internal/purchase/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"gorm.io/gorm"
)

var (
	ErrTaxInvoiceNotFound = errors.New("tax invoice not found")
)

type TaxInvoiceUsecase interface {
	Create(ctx context.Context, req *dto.CreateTaxInvoiceRequest) (*dto.TaxInvoiceResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateTaxInvoiceRequest) (*dto.TaxInvoiceResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.TaxInvoiceResponse, error)
	List(ctx context.Context, req *dto.ListTaxInvoicesRequest) ([]dto.TaxInvoiceResponse, int64, error)
}

type taxInvoiceUsecase struct {
	db     *gorm.DB
	repo   repositories.TaxInvoiceRepository
	mapper *mapper.TaxInvoiceMapper
}

func NewTaxInvoiceUsecase(db *gorm.DB, repo repositories.TaxInvoiceRepository, mapper *mapper.TaxInvoiceMapper) TaxInvoiceUsecase {
	return &taxInvoiceUsecase{db: db, repo: repo, mapper: mapper}
}

func (uc *taxInvoiceUsecase) Create(ctx context.Context, req *dto.CreateTaxInvoiceRequest) (*dto.TaxInvoiceResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	d, err := time.Parse("2006-01-02", strings.TrimSpace(req.TaxInvoiceDate))
	if err != nil {
		return nil, errors.New("invalid tax_invoice_date")
	}

	if req.CustomerInvoiceID != nil && strings.TrimSpace(*req.CustomerInvoiceID) != "" && req.SupplierInvoiceID != nil && strings.TrimSpace(*req.SupplierInvoiceID) != "" {
		return nil, errors.New("only one of customer_invoice_id or supplier_invoice_id can be set")
	}

	item := &financeModels.TaxInvoice{
		TaxInvoiceNumber:  strings.TrimSpace(req.TaxInvoiceNumber),
		TaxInvoiceDate:    d,
		CustomerInvoiceID: req.CustomerInvoiceID,
		SupplierInvoiceID: req.SupplierInvoiceID,
		DPPAmount:         req.DPPAmount,
		VATAmount:         req.VATAmount,
		TotalAmount:       req.TotalAmount,
		Notes:             strings.TrimSpace(req.Notes),
		CreatedBy:         &actorID,
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.CustomerInvoiceID != nil && strings.TrimSpace(*req.CustomerInvoiceID) != "" {
			var inv salesModels.CustomerInvoice
			if err := tx.First(&inv, "id = ?", strings.TrimSpace(*req.CustomerInvoiceID)).Error; err != nil {
				return err
			}
		}
		if req.SupplierInvoiceID != nil && strings.TrimSpace(*req.SupplierInvoiceID) != "" {
			var inv purchaseModels.SupplierInvoice
			if err := tx.First(&inv, "id = ?", strings.TrimSpace(*req.SupplierInvoiceID)).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(item).Error; err != nil {
			return err
		}
		if req.CustomerInvoiceID != nil && strings.TrimSpace(*req.CustomerInvoiceID) != "" {
			if err := tx.Model(&salesModels.CustomerInvoice{}).Where("id = ?", strings.TrimSpace(*req.CustomerInvoiceID)).Updates(map[string]interface{}{
				"tax_invoice_id": item.ID,
			}).Error; err != nil {
				return err
			}
		}
		if req.SupplierInvoiceID != nil && strings.TrimSpace(*req.SupplierInvoiceID) != "" {
			if err := tx.Model(&purchaseModels.SupplierInvoice{}).Where("id = ?", strings.TrimSpace(*req.SupplierInvoiceID)).Updates(map[string]interface{}{
				"tax_invoice_number": item.TaxInvoiceNumber,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *taxInvoiceUsecase) Update(ctx context.Context, id string, req *dto.UpdateTaxInvoiceRequest) (*dto.TaxInvoiceResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTaxInvoiceNotFound
		}
		return nil, err
	}

	d, err := time.Parse("2006-01-02", strings.TrimSpace(req.TaxInvoiceDate))
	if err != nil {
		return nil, errors.New("invalid tax_invoice_date")
	}

	if err := uc.db.WithContext(ctx).Model(&financeModels.TaxInvoice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"tax_invoice_number": strings.TrimSpace(req.TaxInvoiceNumber),
		"tax_invoice_date":   d,
		"dpp_amount":         req.DPPAmount,
		"vat_amount":         req.VATAmount,
		"total_amount":       req.TotalAmount,
		"notes":              strings.TrimSpace(req.Notes),
	}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

func (uc *taxInvoiceUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTaxInvoiceNotFound
		}
		return err
	}

	return uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if item.CustomerInvoiceID != nil && strings.TrimSpace(*item.CustomerInvoiceID) != "" {
			if err := tx.Model(&salesModels.CustomerInvoice{}).Where("id = ?", strings.TrimSpace(*item.CustomerInvoiceID)).Updates(map[string]interface{}{
				"tax_invoice_id": nil,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&financeModels.TaxInvoice{}, "id = ?", id).Error
	})
}

func (uc *taxInvoiceUsecase) GetByID(ctx context.Context, id string) (*dto.TaxInvoiceResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.TaxInvoice{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrTaxInvoiceNotFound
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTaxInvoiceNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *taxInvoiceUsecase) List(ctx context.Context, req *dto.ListTaxInvoicesRequest) ([]dto.TaxInvoiceResponse, int64, error) {
	if req == nil {
		req = &dto.ListTaxInvoicesRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	var startDate *time.Time
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.StartDate))
		if err != nil {
			return nil, 0, errors.New("invalid start_date")
		}
		startDate = &parsed
	}
	var endDate *time.Time
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.EndDate))
		if err != nil {
			return nil, 0, errors.New("invalid end_date")
		}
		endDate = &parsed
	}

	items, total, err := uc.repo.List(ctx, repositories.TaxInvoiceListParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		SortBy:    req.SortBy,
		SortDir:   req.SortDir,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.TaxInvoiceResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}
