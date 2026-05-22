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
	ErrCashBankTransactionNotFound  = errors.New("cash bank transaction not found")
	ErrCashBankTransactionImmutable = errors.New("posted transaction cannot be edited")
	ErrCashBankTransactionReversed  = errors.New("transaction already reversed")
)

type CashBankTransactionUsecase interface {
	Create(ctx context.Context, req *dto.CreateCashBankTransactionRequest) (*dto.CashBankTransactionResponse, error)
	List(ctx context.Context, req *dto.ListCashBankTransactionsRequest) ([]dto.CashBankTransactionResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.CashBankTransactionResponse, error)
	Reverse(ctx context.Context, id string, req *dto.ReverseCashBankTransactionRequest) (*dto.CashBankTransactionResponse, error)
	GetFormData(ctx context.Context, companyID string) (*dto.CashBankTransactionFormDataResponse, error)
}

type cashBankTransactionUsecase struct {
	db        *gorm.DB
	repo      repositories.CashBankTransactionRepository
	journalUC JournalEntryUsecase
}

func NewCashBankTransactionUsecase(db *gorm.DB, repo repositories.CashBankTransactionRepository, journalUC JournalEntryUsecase) CashBankTransactionUsecase {
	return &cashBankTransactionUsecase{db: db, repo: repo, journalUC: journalUC}
}

func (u *cashBankTransactionUsecase) Create(ctx context.Context, req *dto.CreateCashBankTransactionRequest) (*dto.CashBankTransactionResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	trxDate, err := parseDateRequired(req.Date)
	if err != nil {
		return nil, err
	}
	if trxDate.After(apptime.Now()) {
		return nil, errors.New("date cannot be in the future")
	}

	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var bankAccount coreModels.BankAccount
	if err := u.db.WithContext(ctx).First(&bankAccount, "id = ? AND company_id = ? AND is_active = true", strings.TrimSpace(req.BankAccountID), strings.TrimSpace(req.CompanyID)).Error; err != nil {
		return nil, err
	}
	if bankAccount.ChartOfAccountID == nil || strings.TrimSpace(*bankAccount.ChartOfAccountID) == "" {
		return nil, errors.New("bank account must be linked to chart_of_account")
	}

	contraAccountID := ""
	if req.ContraAccountID != nil {
		contraAccountID = strings.TrimSpace(*req.ContraAccountID)
	}
	if contraAccountID == "" {
		return nil, errors.New("contra_account_id is required")
	}

	referenceNumber, err := u.generateReferenceNumber(ctx, req.CompanyID, req.Type, trxDate)
	if err != nil {
		return nil, err
	}

	autoPost := true
	if req.AutoPost != nil {
		autoPost = *req.AutoPost
	}

	item := &financeModels.CashBankTransaction{
		CompanyID:       strings.TrimSpace(req.CompanyID),
		BankAccountID:   strings.TrimSpace(req.BankAccountID),
		ReferenceNumber: referenceNumber,
		Type:            req.Type,
		Date:            trxDate,
		Amount:          req.Amount,
		Reference:       strings.TrimSpace(req.Reference),
		Description:     strings.TrimSpace(req.Description),
		ContraAccountID: &contraAccountID,
		AttachmentURL:   req.AttachmentURL,
		Status:          financeModels.CashBankTransactionStatusDraft,
		CreatedBy:       &actorID,
		UpdatedBy:       &actorID,
	}

	if err := u.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	journalNumber := ""
	if autoPost {
		refType := reference.RefTypeCashBank
		refID := item.ID
		jReq := &dto.CreateJournalEntryRequest{
			CompanyID:         item.CompanyID,
			EntryDate:         item.Date.Format("2006-01-02"),
			Reference:         item.ReferenceNumber,
			Description:       item.Description,
			ReferenceType:     &refType,
			ReferenceID:       &refID,
			IsSystemGenerated: true,
		}

		bankCOA := strings.TrimSpace(*bankAccount.ChartOfAccountID)
		if req.Type == financeModels.CashBankTransactionTypePaymentIn {
			jReq.Lines = []dto.JournalLineRequest{
				{ChartOfAccountID: bankCOA, Debit: req.Amount, Credit: 0, Memo: item.Description},
				{ChartOfAccountID: contraAccountID, Debit: 0, Credit: req.Amount, Memo: item.Description},
			}
		} else {
			jReq.Lines = []dto.JournalLineRequest{
				{ChartOfAccountID: contraAccountID, Debit: req.Amount, Credit: 0, Memo: item.Description},
				{ChartOfAccountID: bankCOA, Debit: 0, Credit: req.Amount, Memo: item.Description},
			}
		}

		journalRes, err := u.journalUC.PostOrUpdateJournal(ctx, jReq)
		if err != nil {
			return nil, err
		}
		journalNumber = journalRes.JournalNumber
		item.JournalEntryID = &journalRes.ID
		item.Status = financeModels.CashBankTransactionStatusPosted
		if err := u.repo.Update(ctx, item); err != nil {
			return nil, err
		}
	}

	return u.toResponse(item, journalNumber), nil
}

func (u *cashBankTransactionUsecase) List(ctx context.Context, req *dto.ListCashBankTransactionsRequest) ([]dto.CashBankTransactionResponse, int64, error) {
	if req == nil {
		req = &dto.ListCashBankTransactionsRequest{}
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

	items, total, err := u.repo.List(ctx, repositories.CashBankTransactionListParams{
		CompanyID:     req.CompanyID,
		BankAccountID: req.BankAccountID,
		Type:          req.Type,
		Status:        req.Status,
		StartDate:     startDate,
		EndDate:       endDate,
		Search:        req.Search,
		SortBy:        req.SortBy,
		SortDir:       req.SortDir,
		Limit:         perPage,
		Offset:        (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.CashBankTransactionResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *u.toResponse(&items[i], ""))
	}

	return responses, total, nil
}

func (u *cashBankTransactionUsecase) GetByID(ctx context.Context, id string) (*dto.CashBankTransactionResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCashBankTransactionNotFound
		}
		return nil, err
	}

	return u.toResponse(item, ""), nil
}

func (u *cashBankTransactionUsecase) Reverse(ctx context.Context, id string, req *dto.ReverseCashBankTransactionRequest) (*dto.CashBankTransactionResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCashBankTransactionNotFound
		}
		return nil, err
	}

	if item.Status != financeModels.CashBankTransactionStatusPosted {
		return nil, errors.New("cannot reverse non-posted transaction")
	}
	if item.ReversedByID != nil && strings.TrimSpace(*item.ReversedByID) != "" {
		return nil, ErrCashBankTransactionReversed
	}

	reason := ""
	if req != nil {
		reason = strings.TrimSpace(req.Reason)
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if item.JournalEntryID == nil || strings.TrimSpace(*item.JournalEntryID) == "" {
		return nil, errors.New("journal entry not found for transaction")
	}

	reverseJournal, err := u.journalUC.ReverseWithReason(ctx, strings.TrimSpace(*item.JournalEntryID), reason)
	if err != nil {
		return nil, err
	}

	reversalType := financeModels.CashBankTransactionTypePaymentOut
	if item.Type == financeModels.CashBankTransactionTypePaymentOut {
		reversalType = financeModels.CashBankTransactionTypePaymentIn
	}

	referenceNumber, err := u.generateReferenceNumber(ctx, item.CompanyID, reversalType, apptime.Now())
	if err != nil {
		return nil, err
	}

	reverseOfID := item.ID
	reversalDescription := strings.TrimSpace(reason)
	if reversalDescription == "" {
		reversalDescription = "Reversal of " + item.ReferenceNumber
	}

	reversal := &financeModels.CashBankTransaction{
		CompanyID:       item.CompanyID,
		BankAccountID:   item.BankAccountID,
		ReferenceNumber: referenceNumber,
		Type:            reversalType,
		Date:            apptime.Now(),
		Amount:          item.Amount,
		Reference:       item.Reference,
		Description:     reversalDescription,
		ContraAccountID: item.ContraAccountID,
		JournalEntryID:  &reverseJournal.ID,
		Status:          financeModels.CashBankTransactionStatusPosted,
		ReverseOfID:     &reverseOfID,
		CreatedBy:       &actorID,
		UpdatedBy:       &actorID,
	}
	if err := u.repo.Create(ctx, reversal); err != nil {
		return nil, err
	}

	item.Status = financeModels.CashBankTransactionStatusReversed
	item.ReversedByID = &reversal.ID
	item.UpdatedBy = &actorID
	if err := u.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return u.toResponse(reversal, reverseJournal.JournalNumber), nil
}

func (u *cashBankTransactionUsecase) GetFormData(ctx context.Context, companyID string) (*dto.CashBankTransactionFormDataResponse, error) {
	accounts := make([]dto.CashBankAccountFormOption, 0)
	var bankAccounts []coreModels.BankAccount
	q := u.db.WithContext(ctx).Where("is_active = true")
	if strings.TrimSpace(companyID) != "" {
		q = q.Where("company_id = ?", strings.TrimSpace(companyID))
	}
	if err := q.Order("code asc").Find(&bankAccounts).Error; err != nil {
		return nil, err
	}
	for _, acc := range bankAccounts {
		accounts = append(accounts, dto.CashBankAccountFormOption{
			ID:            acc.ID,
			Code:          acc.Code,
			Name:          acc.Name,
			AccountNumber: acc.AccountNumber,
		})
	}

	contraAccounts := make([]dto.CashBankCOAFormOption, 0)
	var coaRows []financeModels.ChartOfAccount
	if err := u.db.WithContext(ctx).
		Where("is_active = true AND is_postable = true").
		Order("code asc").
		Find(&coaRows).Error; err != nil {
		return nil, err
	}
	for _, coa := range coaRows {
		contraAccounts = append(contraAccounts, dto.CashBankCOAFormOption{ID: coa.ID, Code: coa.Code, Name: coa.Name, Type: string(coa.Type)})
	}

	return &dto.CashBankTransactionFormDataResponse{
		BankAccounts: accounts,
		TransactionTypes: []dto.ValueLabelOption{
			{Value: string(financeModels.CashBankTransactionTypePaymentIn), Label: "Payment In (Cash/Transfer In)"},
			{Value: string(financeModels.CashBankTransactionTypePaymentOut), Label: "Payment Out (Withdrawal)"},
		},
		ContraAccounts: contraAccounts,
	}, nil
}

func (u *cashBankTransactionUsecase) generateReferenceNumber(ctx context.Context, companyID string, txType financeModels.CashBankTransactionType, date time.Time) (string, error) {
	prefix := "BNK"
	if txType == financeModels.CashBankTransactionTypePaymentIn {
		prefix = "KAS"
	}

	startOfYear := time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(date.Year(), time.December, 31, 23, 59, 59, 0, time.UTC)

	var count int64
	if err := u.db.WithContext(ctx).
		Model(&financeModels.CashBankTransaction{}).
		Where("company_id = ?", strings.TrimSpace(companyID)).
		Where("date >= ? AND date <= ?", startOfYear, endOfYear).
		Count(&count).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%d/%04d", prefix, date.Year(), count+1), nil
}

func (u *cashBankTransactionUsecase) toResponse(item *financeModels.CashBankTransaction, journalNumber string) *dto.CashBankTransactionResponse {
	return &dto.CashBankTransactionResponse{
		ID:                 item.ID,
		CompanyID:          item.CompanyID,
		ReferenceNumber:    item.ReferenceNumber,
		BankAccountID:      item.BankAccountID,
		Type:               item.Type,
		Date:               item.Date.Format("2006-01-02"),
		Amount:             item.Amount,
		Reference:          item.Reference,
		Description:        item.Description,
		ContraAccountID:    item.ContraAccountID,
		JournalEntryID:     item.JournalEntryID,
		JournalEntryNumber: journalNumber,
		Status:             item.Status,
		AttachmentURL:      item.AttachmentURL,
		ReverseOfID:        item.ReverseOfID,
		ReversedByID:       item.ReversedByID,
		CreatedBy:          item.CreatedBy,
		UpdatedBy:          item.UpdatedBy,
		CreatedAt:          item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
	}
}
