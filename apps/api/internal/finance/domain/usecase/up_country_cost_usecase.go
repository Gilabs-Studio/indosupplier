package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUpCountryCostNotFound = errors.New("up-country cost not found")
)

type UpCountryCostUsecase interface {
	Create(ctx context.Context, req *dto.CreateUpCountryCostRequest) (*dto.UpCountryCostResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateUpCountryCostRequest) (*dto.UpCountryCostResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	List(ctx context.Context, req *dto.ListUpCountryCostsRequest) ([]dto.UpCountryCostResponse, int64, error)
	GetStats(ctx context.Context) (*dto.UpCountryCostStatsResponse, error)
	Submit(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	ManagerApprove(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	ManagerReject(ctx context.Context, id string, comment string) (*dto.UpCountryCostResponse, error)
	FinanceApprove(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	MarkPaid(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	// Legacy - kept for backward compatibility
	Approve(ctx context.Context, id string) (*dto.UpCountryCostResponse, error)
	GetFormData(ctx context.Context) (*dto.UpCountryCostFormDataResponse, error)
}

type upCountryCostUsecase struct {
	db               *gorm.DB
	coaRepo          repositories.ChartOfAccountRepository
	repo             repositories.UpCountryCostRepository
	journalUC        JournalEntryUsecase
	mapper           *mapper.UpCountryCostMapper
	settingsService  financesettings.SettingsService
	accountingEngine accounting.AccountingEngine
}

func NewUpCountryCostUsecase(
	db *gorm.DB,
	coaRepo repositories.ChartOfAccountRepository,
	repo repositories.UpCountryCostRepository,
	journalUC JournalEntryUsecase,
	mapper *mapper.UpCountryCostMapper,
	settingsService financesettings.SettingsService,
	accountingEngine accounting.AccountingEngine,
) UpCountryCostUsecase {
	return &upCountryCostUsecase{
		db:               db,
		coaRepo:          coaRepo,
		repo:             repo,
		journalUC:        journalUC,
		mapper:           mapper,
		settingsService:  settingsService,
		accountingEngine: accountingEngine,
	}
}

func (uc *upCountryCostUsecase) Create(ctx context.Context, req *dto.CreateUpCountryCostRequest) (*dto.UpCountryCostResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)

	start, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, errors.New("invalid start_date")
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, errors.New("invalid end_date")
	}

	code, err := uc.repo.GenerateCode(ctx, apptime.Now())
	if err != nil {
		return nil, err
	}

	item := &financeModels.UpCountryCost{
		Code:      code,
		Purpose:   strings.TrimSpace(req.Purpose),
		Location:  strings.TrimSpace(req.Location),
		StartDate: start,
		EndDate:   end,
		Status:    financeModels.UpCountryCostStatusDraft,
		Notes:     strings.TrimSpace(req.Notes),
		CreatedBy: &actorID,
	}

	for _, e := range req.Employees {
		item.Employees = append(item.Employees, financeModels.UpCountryCostEmployee{
			EmployeeID: e.EmployeeID,
		})
	}

	for _, it := range req.Items {
		costItem := financeModels.UpCountryCostItem{
			CostType:    financeModels.CostType(it.CostType),
			Description: it.Description,
			Amount:      it.Amount,
		}
		if it.ExpenseDate != "" {
			d, parseErr := time.Parse("2006-01-02", strings.TrimSpace(it.ExpenseDate))
			if parseErr == nil {
				costItem.ExpenseDate = &d
			}
		}
		item.Items = append(item.Items, costItem)
	}

	if err := uc.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *upCountryCostUsecase) Update(ctx context.Context, id string, req *dto.UpdateUpCountryCostRequest) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	existing, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUpCountryCostNotFound
		}
		return nil, err
	}

	if existing.Status != financeModels.UpCountryCostStatusDraft {
		return nil, errors.New("only draft can be updated")
	}

	start, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		return nil, errors.New("invalid start_date")
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		return nil, errors.New("invalid end_date")
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&financeModels.UpCountryCost{}).Where("id = ?", id).Updates(map[string]interface{}{
			"purpose":    strings.TrimSpace(req.Purpose),
			"location":   strings.TrimSpace(req.Location),
			"start_date": start,
			"end_date":   end,
			"notes":      strings.TrimSpace(req.Notes),
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("up_country_cost_id = ?", id).Delete(&financeModels.UpCountryCostEmployee{}).Error; err != nil {
			return err
		}
		for _, e := range req.Employees {
			if err := tx.Create(&financeModels.UpCountryCostEmployee{UpCountryCostID: id, EmployeeID: e.EmployeeID}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("up_country_cost_id = ?", id).Delete(&financeModels.UpCountryCostItem{}).Error; err != nil {
			return err
		}
		for _, it := range req.Items {
			costItem := financeModels.UpCountryCostItem{
				UpCountryCostID: id,
				CostType:        financeModels.CostType(it.CostType),
				Description:     it.Description,
				Amount:          it.Amount,
			}
			if it.ExpenseDate != "" {
				d, parseErr := time.Parse("2006-01-02", strings.TrimSpace(it.ExpenseDate))
				if parseErr == nil {
					costItem.ExpenseDate = &d
				}
			}
			if err := tx.Create(&costItem).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	full, _ := uc.repo.FindByID(ctx, id, true)
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

func (uc *upCountryCostUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	existing, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		return err
	}
	if existing.Status != financeModels.UpCountryCostStatusDraft {
		return errors.New("only draft can be deleted")
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.UpCountryCost{}, "id = ?", id).Error
}

func (uc *upCountryCostUsecase) GetByID(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.UpCountryCost{}, id, security.MixedOwnershipScopeQueryOptions("employee_id")) {
		return nil, ErrUpCountryCostNotFound
	}
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *upCountryCostUsecase) List(ctx context.Context, req *dto.ListUpCountryCostsRequest) ([]dto.UpCountryCostResponse, int64, error) {
	if req == nil {
		req = &dto.ListUpCountryCostsRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}

	var status *financeModels.UpCountryCostStatus
	if req.Status != nil && *req.Status != "" {
		s := financeModels.UpCountryCostStatus(*req.Status)
		status = &s
	}

	var start *time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		p, _ := time.Parse("2006-01-02", *req.StartDate)
		start = &p
	}
	var end *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		p, _ := time.Parse("2006-01-02", *req.EndDate)
		end = &p
	}

	items, total, err := uc.repo.List(ctx, repositories.UpCountryCostListParams{
		Search:     req.Search,
		Status:     status,
		StartDate:  start,
		EndDate:    end,
		EmployeeID: req.EmployeeID,
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
		SortBy:     req.SortBy,
		SortDir:    req.SortDir,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.UpCountryCostResponse, 0, len(items))
	for i := range items {
		res = append(res, uc.mapper.ToResponse(&items[i]))
	}
	return res, total, nil
}

func (uc *upCountryCostUsecase) GetStats(ctx context.Context) (*dto.UpCountryCostStatsResponse, error) {
	stats, err := uc.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.UpCountryCostStatsResponse{
		TotalRequests:   stats.TotalRequests,
		PendingApproval: stats.PendingApproval,
		Approved:        stats.Approved,
		TotalAmount:     stats.TotalAmount,
	}, nil
}

func (uc *upCountryCostUsecase) Submit(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if item.Status != financeModels.UpCountryCostStatusDraft {
		return nil, errors.New("only draft can be submitted")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	now := apptime.Now()

	if err := uc.db.WithContext(ctx).Model(&financeModels.UpCountryCost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       financeModels.UpCountryCostStatusSubmitted,
		"submitted_at": &now,
		"submitted_by": &actorID,
	}).Error; err != nil {
		return nil, err
	}

	item.Status = financeModels.UpCountryCostStatusSubmitted
	item.SubmittedAt = &now
	item.SubmittedBy = &actorID
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "up_country_cost.approve",
		EntityType:     "up_country_cost",
		EntityID:       item.ID,
		Title:          "Up Country Cost Approval",
		Message:        "An up country cost request has been submitted and requires your approval.",
		ActorUserID:    actorID,
	}); err != nil {
		log.Printf("warning: failed to create up country cost notification: %v", err)
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *upCountryCostUsecase) ManagerApprove(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	now := apptime.Now()

	var item financeModels.UpCountryCost
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Items").
			First(&item, "id = ?", id).Error; err != nil {
			return err
		}

		if item.Status != financeModels.UpCountryCostStatusSubmitted {
			return errors.New("only submitted requests can be manager-approved")
		}

		var total float64
		for _, it := range item.Items {
			total += it.Amount
		}

		txData := accounting.TransactionData{
			ReferenceType: reference.RefTypeUpCountryCost,
			ReferenceID:   item.ID,
			EntryDate:     now.Format("2006-01-02"),
			Description:   "Up-Country Cost Approval: " + item.Code + " - " + item.Purpose,
			TotalAmount:   total,
			MemoArgs:      []interface{}{item.Purpose},
		}

		journalReq, err := uc.accountingEngine.GenerateJournal(txCtx, accounting.ProfileUpCountryApproval, txData)
		if err != nil {
			return err
		}
		if _, err := uc.journalUC.PostOrUpdateJournal(txCtx, journalReq); err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Model(&financeModels.UpCountryCost{}).
			Where("id = ? AND status = ?", id, financeModels.UpCountryCostStatusSubmitted).
			Updates(map[string]interface{}{
				"status":              financeModels.UpCountryCostStatusManagerApproved,
				"manager_approved_at": &now,
				"manager_approved_by": &actorID,
				"manager_comment":     "",
				// legacy fields
				"approved_at": &now,
				"approved_by": &actorID,
			}).Error; err != nil {
			return err
		}

		item.Status = financeModels.UpCountryCostStatusManagerApproved
		item.ManagerApprovedAt = &now
		item.ManagerApprovedBy = &actorID
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(&item)
	return &res, nil
}

func (uc *upCountryCostUsecase) ManagerReject(ctx context.Context, id string, comment string) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if item.Status != financeModels.UpCountryCostStatusSubmitted {
		return nil, errors.New("only submitted requests can be rejected by manager")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	now := apptime.Now()

	if err := uc.db.WithContext(ctx).Model(&financeModels.UpCountryCost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":              financeModels.UpCountryCostStatusRejected,
		"manager_approved_at": &now,
		"manager_approved_by": &actorID,
		"manager_comment":     strings.TrimSpace(comment),
	}).Error; err != nil {
		return nil, err
	}

	item.Status = financeModels.UpCountryCostStatusRejected
	item.ManagerComment = strings.TrimSpace(comment)
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *upCountryCostUsecase) FinanceApprove(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if item.Status != financeModels.UpCountryCostStatusManagerApproved {
		return nil, errors.New("only manager-approved requests can be finance-approved")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	now := apptime.Now()

	if err := uc.db.WithContext(ctx).Model(&financeModels.UpCountryCost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":              financeModels.UpCountryCostStatusFinanceApproved,
		"finance_approved_at": &now,
		"finance_approved_by": &actorID,
	}).Error; err != nil {
		return nil, err
	}

	item.Status = financeModels.UpCountryCostStatusFinanceApproved
	item.FinanceApprovedAt = &now
	item.FinanceApprovedBy = &actorID
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *upCountryCostUsecase) MarkPaid(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	id = strings.TrimSpace(id)
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if item.Status != financeModels.UpCountryCostStatusFinanceApproved {
		return nil, errors.New("only finance-approved requests can be marked as paid")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	now := apptime.Now()

	if err := uc.db.WithContext(ctx).Model(&financeModels.UpCountryCost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  financeModels.UpCountryCostStatusPaid,
		"paid_at": &now,
		"paid_by": &actorID,
	}).Error; err != nil {
		return nil, err
	}

	item.Status = financeModels.UpCountryCostStatusPaid
	item.PaidAt = &now
	item.PaidBy = &actorID
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

// Approve is kept for backward compatibility - delegates to ManagerApprove
func (uc *upCountryCostUsecase) Approve(ctx context.Context, id string) (*dto.UpCountryCostResponse, error) {
	return uc.ManagerApprove(ctx, id)
}
