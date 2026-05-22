package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"gorm.io/gorm"
)

var (
	ErrBankTransferNotFound = errors.New("bank transfer not found")
)

type BankTransferUsecase interface {
	Create(ctx context.Context, req *dto.CreateBankTransferRequest) (*dto.BankTransferResponse, error)
	List(ctx context.Context, req *dto.ListBankTransfersRequest) ([]dto.BankTransferResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.BankTransferResponse, error)
	Complete(ctx context.Context, id string) (*dto.BankTransferResponse, error)
	Cancel(ctx context.Context, id string, req *dto.CancelBankTransferRequest) (*dto.BankTransferResponse, error)
}

type bankTransferUsecase struct {
	db        *gorm.DB
	repo      repositories.BankTransferRepository
	journalUC JournalEntryUsecase
}

func NewBankTransferUsecase(db *gorm.DB, repo repositories.BankTransferRepository, journalUC JournalEntryUsecase) BankTransferUsecase {
	return &bankTransferUsecase{db: db, repo: repo, journalUC: journalUC}
}

func (u *bankTransferUsecase) Create(ctx context.Context, req *dto.CreateBankTransferRequest) (*dto.BankTransferResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if strings.TrimSpace(req.FromBankAccountID) == strings.TrimSpace(req.ToBankAccountID) {
		return nil, errors.New("from and to bank account must be different")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	transferDate, err := parseDateRequired(req.Date)
	if err != nil {
		return nil, err
	}
	if transferDate.After(apptime.Now()) {
		return nil, errors.New("date cannot be in the future")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var fromAcc coreModels.BankAccount
	if err := u.db.WithContext(ctx).First(&fromAcc, "id = ? AND company_id = ? AND is_active = true", strings.TrimSpace(req.FromBankAccountID), strings.TrimSpace(req.CompanyID)).Error; err != nil {
		return nil, err
	}
	var toAcc coreModels.BankAccount
	if err := u.db.WithContext(ctx).First(&toAcc, "id = ? AND company_id = ? AND is_active = true", strings.TrimSpace(req.ToBankAccountID), strings.TrimSpace(req.CompanyID)).Error; err != nil {
		return nil, err
	}
	if fromAcc.CurrencyID != nil && toAcc.CurrencyID != nil && strings.TrimSpace(*fromAcc.CurrencyID) != strings.TrimSpace(*toAcc.CurrencyID) {
		return nil, errors.New("accounts must have same currency")
	}

	transferNumber, err := u.generateTransferNumber(ctx, req.CompanyID, transferDate)
	if err != nil {
		return nil, err
	}

	item := &financeModels.BankTransfer{
		CompanyID:         strings.TrimSpace(req.CompanyID),
		TransferNumber:    transferNumber,
		FromBankAccountID: strings.TrimSpace(req.FromBankAccountID),
		ToBankAccountID:   strings.TrimSpace(req.ToBankAccountID),
		Amount:            req.Amount,
		Date:              transferDate,
		Reference:         strings.TrimSpace(req.Reference),
		Description:       strings.TrimSpace(req.Description),
		Status:            financeModels.BankTransferStatusPending,
		CreatedBy:         &actorID,
		UpdatedBy:         &actorID,
	}

	if err := u.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	return u.toResponse(item), nil
}

func (u *bankTransferUsecase) List(ctx context.Context, req *dto.ListBankTransfersRequest) ([]dto.BankTransferResponse, int64, error) {
	if req == nil {
		req = &dto.ListBankTransfersRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	startDate, err := parseDateOptional(req.DateFrom)
	if err != nil {
		return nil, 0, err
	}
	endDate, err := parseEndDateOptional(req.DateTo)
	if err != nil {
		return nil, 0, err
	}

	items, total, err := u.repo.List(ctx, repositories.BankTransferListParams{
		CompanyID:         req.CompanyID,
		FromBankAccountID: req.FromBankAccountID,
		ToBankAccountID:   req.ToBankAccountID,
		Status:            req.Status,
		StartDate:         startDate,
		EndDate:           endDate,
		Search:            req.Search,
		SortBy:            req.SortBy,
		SortDir:           req.SortDir,
		Limit:             perPage,
		Offset:            (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.BankTransferResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *u.toResponse(&items[i]))
	}

	return responses, total, nil
}

func (u *bankTransferUsecase) GetByID(ctx context.Context, id string) (*dto.BankTransferResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankTransferNotFound
		}
		return nil, err
	}

	return u.toResponse(item), nil
}

func (u *bankTransferUsecase) Complete(ctx context.Context, id string) (*dto.BankTransferResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankTransferNotFound
		}
		return nil, err
	}
	if item.Status != financeModels.BankTransferStatusPending {
		return nil, errors.New("only pending transfer can be completed")
	}

	var fromAcc coreModels.BankAccount
	if err := u.db.WithContext(ctx).First(&fromAcc, "id = ?", item.FromBankAccountID).Error; err != nil {
		return nil, err
	}
	var toAcc coreModels.BankAccount
	if err := u.db.WithContext(ctx).First(&toAcc, "id = ?", item.ToBankAccountID).Error; err != nil {
		return nil, err
	}
	if fromAcc.ChartOfAccountID == nil || strings.TrimSpace(*fromAcc.ChartOfAccountID) == "" {
		return nil, errors.New("source bank account must have chart_of_account")
	}
	if toAcc.ChartOfAccountID == nil || strings.TrimSpace(*toAcc.ChartOfAccountID) == "" {
		return nil, errors.New("destination bank account must have chart_of_account")
	}

	refType := reference.RefTypeCashBank
	refID := item.ID
	description := strings.TrimSpace(item.Description)
	if description == "" {
		description = "Bank transfer " + item.TransferNumber
	}
	jReq := &dto.CreateJournalEntryRequest{
		CompanyID:         item.CompanyID,
		EntryDate:         item.Date.Format("2006-01-02"),
		Reference:         item.TransferNumber,
		Description:       description,
		ReferenceType:     &refType,
		ReferenceID:       &refID,
		IsSystemGenerated: true,
		Lines: []dto.JournalLineRequest{
			{ChartOfAccountID: strings.TrimSpace(*toAcc.ChartOfAccountID), Debit: item.Amount, Credit: 0, Memo: item.Description},
			{ChartOfAccountID: strings.TrimSpace(*fromAcc.ChartOfAccountID), Debit: 0, Credit: item.Amount, Memo: item.Description},
		},
	}

	journalRes, err := u.journalUC.PostOrUpdateJournal(ctx, jReq)
	if err != nil {
		return nil, err
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID != "" {
		item.UpdatedBy = &actorID
	}
	item.Status = financeModels.BankTransferStatusCompleted
	item.JournalEntryID = &journalRes.ID
	if err := u.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return u.toResponse(item), nil
}

func (u *bankTransferUsecase) Cancel(ctx context.Context, id string, req *dto.CancelBankTransferRequest) (*dto.BankTransferResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankTransferNotFound
		}
		return nil, err
	}
	if item.Status != financeModels.BankTransferStatusPending {
		return nil, errors.New("cannot cancel completed transfer")
	}

	item.Status = financeModels.BankTransferStatusCancelled
	if req != nil && strings.TrimSpace(req.Reason) != "" {
		item.Description = strings.TrimSpace(item.Description + " | cancelled: " + strings.TrimSpace(req.Reason))
	}
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID != "" {
		item.UpdatedBy = &actorID
	}
	if err := u.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return u.toResponse(item), nil
}

func (u *bankTransferUsecase) generateTransferNumber(ctx context.Context, companyID string, date time.Time) (string, error) {
	startOfYear := time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(date.Year(), time.December, 31, 23, 59, 59, 0, time.UTC)

	var count int64
	if err := u.db.WithContext(ctx).
		Model(&financeModels.BankTransfer{}).
		Where("company_id = ?", strings.TrimSpace(companyID)).
		Where("date >= ? AND date <= ?", startOfYear, endOfYear).
		Count(&count).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("TRF/%d/%04d", date.Year(), count+1), nil
}

func (u *bankTransferUsecase) toResponse(item *financeModels.BankTransfer) *dto.BankTransferResponse {
	return &dto.BankTransferResponse{
		ID:                item.ID,
		CompanyID:         item.CompanyID,
		TransferNumber:    item.TransferNumber,
		FromBankAccountID: item.FromBankAccountID,
		ToBankAccountID:   item.ToBankAccountID,
		Amount:            item.Amount,
		Date:              item.Date.Format("2006-01-02"),
		Reference:         item.Reference,
		Description:       item.Description,
		Status:            item.Status,
		JournalEntryID:    item.JournalEntryID,
		CreatedBy:         item.CreatedBy,
		UpdatedBy:         item.UpdatedBy,
		CreatedAt:         item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         item.UpdatedAt.Format(time.RFC3339),
	}
}
