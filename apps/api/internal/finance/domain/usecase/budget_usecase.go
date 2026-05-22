package usecase

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrBudgetNotFound          = errors.New("budget not found")
	ErrBudgetApprovedImmutable = errors.New("approved budget cannot be modified")
	ErrBudgetInvalidItems      = errors.New("invalid budget items")
)

type BudgetUsecase interface {
	Create(ctx context.Context, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateBudgetRequest) (*dto.BudgetResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.BudgetResponse, error)
	List(ctx context.Context, req *dto.ListBudgetsRequest) ([]dto.BudgetResponse, int64, error)
	Approve(ctx context.Context, id string) (*dto.BudgetResponse, error)
	SyncActuals(ctx context.Context, id string) (*dto.BudgetResponse, error)
	GetFormData(ctx context.Context) (*dto.BudgetFormDataResponse, error)
}

type budgetUsecase struct {
	db      *gorm.DB
	coaRepo repositories.ChartOfAccountRepository
	repo    repositories.BudgetRepository
	mapper  *mapper.BudgetMapper
}

func NewBudgetUsecase(db *gorm.DB, coaRepo repositories.ChartOfAccountRepository, repo repositories.BudgetRepository, mapper *mapper.BudgetMapper) BudgetUsecase {
	return &budgetUsecase{db: db, coaRepo: coaRepo, repo: repo, mapper: mapper}
}

func validateBudgetItems(items []dto.BudgetItemRequest) (float64, error) {
	if len(items) < 1 {
		return 0, ErrBudgetInvalidItems
	}
	var sum float64
	for _, it := range items {
		if strings.TrimSpace(it.ChartOfAccountID) == "" {
			return 0, ErrBudgetInvalidItems
		}
		if it.Amount <= 0 {
			return 0, ErrBudgetInvalidItems
		}
		sum += it.Amount
	}
	return sum, nil
}

func (uc *budgetUsecase) Create(ctx context.Context, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, errors.New("invalid start_date")
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, errors.New("invalid end_date")
	}
	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be >= start_date")
	}

	sum, err := validateBudgetItems(req.Items)
	if err != nil {
		return nil, err
	}
	if sum <= 0 {
		return nil, ErrBudgetInvalidItems
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var createdID string
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coaIDs := make([]string, 0, len(req.Items))
		for _, it := range req.Items {
			coaIDs = append(coaIDs, strings.TrimSpace(it.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, it := range req.Items {
			if coaByID[strings.TrimSpace(it.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		b := &financeModels.Budget{
			Name:        strings.TrimSpace(req.Name),
			Description: strings.TrimSpace(req.Description),
			StartDate:   startDate,
			EndDate:     endDate,
			TotalAmount: sum,
			Status:      financeModels.BudgetStatusDraft,
			CreatedBy:   &actorID,
		}
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		for _, it := range req.Items {
			coa := coaByID[strings.TrimSpace(it.ChartOfAccountID)]
			codeSnap := ""
			nameSnap := ""
			typeSnap := ""
			snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			item := &financeModels.BudgetItem{
				BudgetID:                   b.ID,
				ChartOfAccountID:           strings.TrimSpace(it.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				Amount:                     it.Amount,
				Memo:                       strings.TrimSpace(it.Memo),
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		createdID = b.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, createdID, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

func (uc *budgetUsecase) Update(ctx context.Context, id string, req *dto.UpdateBudgetRequest) (*dto.BudgetResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	b, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}
	if b.Status == financeModels.BudgetStatusApproved {
		return nil, ErrBudgetApprovedImmutable
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, errors.New("invalid start_date")
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, errors.New("invalid end_date")
	}
	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be >= start_date")
	}

	sum, err := validateBudgetItems(req.Items)
	if err != nil {
		return nil, err
	}
	if math.Abs(sum) < 0.000001 {
		return nil, ErrBudgetInvalidItems
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coaIDs := make([]string, 0, len(req.Items))
		for _, it := range req.Items {
			coaIDs = append(coaIDs, strings.TrimSpace(it.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, it := range req.Items {
			if coaByID[strings.TrimSpace(it.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		if err := tx.Model(&financeModels.Budget{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"name":         strings.TrimSpace(req.Name),
				"description":  strings.TrimSpace(req.Description),
				"start_date":   startDate,
				"end_date":     endDate,
				"total_amount": sum,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("budget_id = ?", id).Delete(&financeModels.BudgetItem{}).Error; err != nil {
			return err
		}
		for _, it := range req.Items {
			coa := coaByID[strings.TrimSpace(it.ChartOfAccountID)]
			codeSnap := ""
			nameSnap := ""
			typeSnap := ""
			snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			item := &financeModels.BudgetItem{
				BudgetID:                   id,
				ChartOfAccountID:           strings.TrimSpace(it.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				Amount:                     it.Amount,
				Memo:                       strings.TrimSpace(it.Memo),
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

func (uc *budgetUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}

	b, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrBudgetNotFound
		}
		return err
	}
	if b.Status == financeModels.BudgetStatusApproved {
		return ErrBudgetApprovedImmutable
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.Budget{}, "id = ?", id).Error
}

func (uc *budgetUsecase) GetByID(ctx context.Context, id string) (*dto.BudgetResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.Budget{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrBudgetNotFound
	}
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}
	resp := uc.mapper.ToResponse(item)
	return &resp, nil
}

func (uc *budgetUsecase) List(ctx context.Context, req *dto.ListBudgetsRequest) ([]dto.BudgetResponse, int64, error) {
	if req == nil {
		req = &dto.ListBudgetsRequest{}
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

	items, total, err := uc.repo.List(ctx, repositories.BudgetListParams{
		Search:    req.Search,
		Status:    req.Status,
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

	res := make([]dto.BudgetResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}

func (uc *budgetUsecase) Approve(ctx context.Context, id string) (*dto.BudgetResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	b, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}
	if b.Status == financeModels.BudgetStatusApproved {
		full, err := uc.repo.FindByID(ctx, id, true)
		if err != nil {
			return nil, err
		}
		resp := uc.mapper.ToResponse(full)
		return &resp, nil
	}

	now := apptime.Now()
	if err := uc.db.WithContext(ctx).Model(&financeModels.Budget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      financeModels.BudgetStatusApproved,
			"approved_at": now,
			"approved_by": actorID,
		}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

func (uc *budgetUsecase) SyncActuals(ctx context.Context, id string) (*dto.BudgetResponse, error) {
	id = strings.TrimSpace(id)
	if _, err := uc.repo.FindByID(ctx, id, false); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}

	if err := uc.repo.SyncActuals(ctx, id); err != nil {
		return nil, err
	}

	full, _ := uc.repo.FindByID(ctx, id, true)
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}
