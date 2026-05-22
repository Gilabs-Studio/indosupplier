package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNonTradePayableNotFound = errors.New("non-trade payable not found")
)

type NonTradePayableUsecase interface {
	Create(ctx context.Context, req *dto.CreateNonTradePayableRequest) (*dto.NonTradePayableResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateNonTradePayableRequest) (*dto.NonTradePayableResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)
	List(ctx context.Context, req *dto.ListNonTradePayablesRequest) ([]dto.NonTradePayableResponse, int64, error)
	Post(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)
	Cancel(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)

	// Backward-compatible aliases used by legacy clients.
	Submit(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)
	Approve(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)
	Reject(ctx context.Context, id string) (*dto.NonTradePayableResponse, error)

	Pay(ctx context.Context, id string, req *dto.PayNonTradePayableRequest) (*dto.NonTradePayableResponse, error)
	GetFormData(ctx context.Context) (*dto.NonTradePayableFormDataResponse, error)
}

type nonTradePayableUsecase struct {
	db               *gorm.DB
	coaRepo          repositories.ChartOfAccountRepository
	repo             repositories.NonTradePayableRepository
	journalUC        JournalEntryUsecase
	mapper           *mapper.NonTradePayableMapper
	settingsService  financesettings.SettingsService
	accountingEngine accounting.AccountingEngine
}

func NewNonTradePayableUsecase(
	db *gorm.DB,
	coaRepo repositories.ChartOfAccountRepository,
	repo repositories.NonTradePayableRepository,
	journalUC JournalEntryUsecase,
	mapper *mapper.NonTradePayableMapper,
	settingsService financesettings.SettingsService,
	accountingEngine accounting.AccountingEngine,
) NonTradePayableUsecase {
	return &nonTradePayableUsecase{
		db:               db,
		coaRepo:          coaRepo,
		repo:             repo,
		journalUC:        journalUC,
		mapper:           mapper,
		settingsService:  settingsService,
		accountingEngine: accountingEngine,
	}
}

func parseOptDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeNTPStatus(value *string) *financeModels.NonTradePayableStatus {
	if value == nil {
		return nil
	}
	v := strings.ToLower(strings.TrimSpace(*value))
	if v == "" {
		return nil
	}

	status := financeModels.NonTradePayableStatus(v)
	switch status {
	case financeModels.NTPStatusDraft,
		financeModels.NTPStatusPosted,
		financeModels.NTPStatusPartial,
		financeModels.NTPStatusPaid,
		financeModels.NTPStatusCancelled,
		financeModels.NTPStatusSubmitted,
		financeModels.NTPStatusApproved,
		financeModels.NTPStatusRejected:
		return &status
	default:
		return nil
	}
}

func (uc *nonTradePayableUsecase) getPaidAmount(ctx context.Context, payableID string) (float64, error) {
	payableID = strings.TrimSpace(payableID)
	if payableID == "" {
		return 0, nil
	}

	const query = `
		SELECT COALESCE(SUM(je.debit_total), 0)
		FROM journal_entries je
		WHERE je.deleted_at IS NULL
			AND je.status = 'posted'
			AND UPPER(COALESCE(je.reference_type, '')) = 'NTP_PAYMENT'
			AND (je.reference = ? OR je.reference_id = ?)
	`

	var paid float64
	if err := uc.db.WithContext(ctx).Raw(query, payableID, payableID).Scan(&paid).Error; err != nil {
		return 0, err
	}
	if math.IsNaN(paid) || paid < 0 {
		return 0, nil
	}
	return paid, nil
}

func (uc *nonTradePayableUsecase) resolvePaymentCOAID(ctx context.Context, req *dto.PayNonTradePayableRequest) (string, error) {
	if req == nil {
		return "", errors.New("request is required")
	}

	if coaID := strings.TrimSpace(req.ChartOfAccountID); coaID != "" {
		return coaID, nil
	}

	bankAccountID := strings.TrimSpace(req.BankAccountID)
	if bankAccountID == "" {
		return "", errors.New("chart_of_account_id or bank_account_id is required")
	}

	var bankAccount coreModels.BankAccount
	if err := uc.db.WithContext(ctx).First(&bankAccount, "id = ?", bankAccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("bank account not found")
		}
		return "", err
	}

	if bankAccount.ChartOfAccountID == nil || strings.TrimSpace(*bankAccount.ChartOfAccountID) == "" {
		return "", errors.New("bank account has no chart_of_account_id mapping")
	}

	return strings.TrimSpace(*bankAccount.ChartOfAccountID), nil
}

func attachOutstanding(response *dto.NonTradePayableResponse, amount float64, paid float64) {
	if response == nil {
		return
	}
	if math.IsNaN(paid) || paid < 0 {
		paid = 0
	}
	remaining := amount - paid
	if remaining < 0 && math.Abs(remaining) < 0.01 {
		remaining = 0
	}
	if remaining < 0 {
		remaining = 0
	}
	response.PaidAmount = paid
	response.RemainingAmount = remaining
}

func (uc *nonTradePayableUsecase) Create(ctx context.Context, req *dto.CreateNonTradePayableRequest) (*dto.NonTradePayableResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	d, err := time.Parse("2006-01-02", strings.TrimSpace(req.TransactionDate))
	if err != nil {
		return nil, errors.New("invalid transaction_date")
	}
	due, err := parseOptDate(req.DueDate)
	if err != nil {
		return nil, errors.New("invalid due_date")
	}

	coa, err := uc.coaRepo.FindByID(ctx, strings.TrimSpace(req.ChartOfAccountID))
	if err != nil {
		return nil, err
	}

	code, err := uc.repo.GenerateCode(ctx, d)
	if err != nil {
		return nil, err
	}

	item := &financeModels.NonTradePayable{
		TransactionDate:            d,
		Code:                       code,
		Description:                strings.TrimSpace(req.Description),
		ChartOfAccountID:           strings.TrimSpace(req.ChartOfAccountID),
		ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
		ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
		ChartOfAccountTypeSnapshot: strings.TrimSpace(string(coa.Type)),
		Amount:                     req.Amount,
		VendorName:                 strings.TrimSpace(req.VendorName),
		DueDate:                    due,
		ReferenceNo:                strings.TrimSpace(req.ReferenceNo),
		Status:                     financeModels.NTPStatusDraft,
		CreatedBy:                  &actorID,
	}
	if err := uc.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(item)
	attachOutstanding(&res, item.Amount, 0)
	return &res, nil
}

func (uc *nonTradePayableUsecase) Update(ctx context.Context, id string, req *dto.UpdateNonTradePayableRequest) (*dto.NonTradePayableResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNonTradePayableNotFound
		}
		return nil, err
	}
	if existing.Status != financeModels.NTPStatusDraft {
		return nil, errors.New("only draft non-trade payable can be edited")
	}

	d, err := time.Parse("2006-01-02", strings.TrimSpace(req.TransactionDate))
	if err != nil {
		return nil, errors.New("invalid transaction_date")
	}
	due, err := parseOptDate(req.DueDate)
	if err != nil {
		return nil, errors.New("invalid due_date")
	}

	coaID := strings.TrimSpace(req.ChartOfAccountID)
	if coaID == "" {
		return nil, errors.New("chart_of_account_id is required")
	}
	coa, err := uc.coaRepo.FindByID(ctx, coaID)
	if err != nil {
		return nil, err
	}

	coaCodeSnap := strings.TrimSpace(existing.ChartOfAccountCodeSnapshot)
	coaNameSnap := strings.TrimSpace(existing.ChartOfAccountNameSnapshot)
	coaTypeSnap := strings.TrimSpace(existing.ChartOfAccountTypeSnapshot)
	if strings.TrimSpace(existing.ChartOfAccountID) != coaID || (coaCodeSnap == "" && coaNameSnap == "" && coaTypeSnap == "") {
		coaCodeSnap = strings.TrimSpace(coa.Code)
		coaNameSnap = strings.TrimSpace(coa.Name)
		coaTypeSnap = strings.TrimSpace(string(coa.Type))
	}

	if err := uc.db.WithContext(ctx).Model(&financeModels.NonTradePayable{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transaction_date":               d,
		"description":                    strings.TrimSpace(req.Description),
		"chart_of_account_id":            coaID,
		"chart_of_account_code_snapshot": coaCodeSnap,
		"chart_of_account_name_snapshot": coaNameSnap,
		"chart_of_account_type_snapshot": coaTypeSnap,
		"amount":                         req.Amount,
		"vendor_name":                    strings.TrimSpace(req.VendorName),
		"due_date":                       due,
		"reference_no":                   strings.TrimSpace(req.ReferenceNo),
	}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	paidAmount, err := uc.getPaidAmount(ctx, full.ID)
	if err != nil {
		return nil, err
	}
	attachOutstanding(&res, full.Amount, paidAmount)
	return &res, nil
}

func (uc *nonTradePayableUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNonTradePayableNotFound
		}
		return err
	}

	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.Status != financeModels.NTPStatusDraft && existing.Status != financeModels.NTPStatusCancelled {
		return errors.New("only draft or cancelled non-trade payable can be deleted")
	}

	return uc.db.WithContext(ctx).Delete(&financeModels.NonTradePayable{}, "id = ?", id).Error
}

func (uc *nonTradePayableUsecase) GetByID(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.NonTradePayable{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrNonTradePayableNotFound
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNonTradePayableNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	paidAmount, err := uc.getPaidAmount(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	attachOutstanding(&res, item.Amount, paidAmount)
	return &res, nil
}

func (uc *nonTradePayableUsecase) List(ctx context.Context, req *dto.ListNonTradePayablesRequest) ([]dto.NonTradePayableResponse, int64, error) {
	if req == nil {
		req = &dto.ListNonTradePayablesRequest{}
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

	items, total, err := uc.repo.List(ctx, repositories.NonTradePayableListParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    normalizeNTPStatus(req.Status),
		SortBy:    req.SortBy,
		SortDir:   req.SortDir,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.NonTradePayableResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		paidAmount, paidErr := uc.getPaidAmount(ctx, items[i].ID)
		if paidErr != nil {
			return nil, 0, paidErr
		}
		attachOutstanding(&mapped, items[i].Amount, paidAmount)
		res = append(res, mapped)
	}
	return res, total, nil
}

func (uc *nonTradePayableUsecase) Post(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var item financeModels.NonTradePayable
	var journalID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("ChartOfAccount").
			First(&item, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNonTradePayableNotFound
			}
			return err
		}

		switch item.Status {
		case financeModels.NTPStatusCancelled, financeModels.NTPStatusRejected:
			return errors.New("cancelled non-trade payable cannot be posted")
		case financeModels.NTPStatusPaid:
			return errors.New("paid non-trade payable cannot be posted")
		case financeModels.NTPStatusPosted, financeModels.NTPStatusPartial:
			return nil
		}

		if err := ensureNotClosed(txCtx, tx, item.TransactionDate); err != nil {
			return err
		}

		txData := accounting.TransactionData{
			ReferenceType:    reference.RefTypeNonTradePayable,
			ReferenceID:      item.ID,
			EntryDate:        item.TransactionDate.Format("2006-01-02"),
			Description:      "NTP Posted: " + item.Code + " - " + item.Description,
			TotalAmount:      item.Amount,
			TransactionCOAID: item.ChartOfAccountID,
			MemoArgs:         []interface{}{item.Description},
		}

		journalReq, err := uc.accountingEngine.GenerateJournal(txCtx, accounting.ProfileNonTradePayableApproval, txData)
		if err != nil {
			return err
		}
		journalReq.Reference = item.ID

		journalRes, err := uc.journalUC.PostOrUpdateJournal(txCtx, journalReq)
		if err != nil {
			return err
		}
		journalID = journalRes.ID

		if err := tx.WithContext(ctx).Model(&financeModels.NonTradePayable{}).
			Where("id = ?", id).
			Update("status", financeModels.NTPStatusPosted).Error; err != nil {
			return err
		}
		item.Status = financeModels.NTPStatusPosted
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(&item)
	if journalID != "" {
		res.JournalID = &journalID
	}
	paidAmount, err := uc.getPaidAmount(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	attachOutstanding(&res, item.Amount, paidAmount)
	return &res, nil
}

func (uc *nonTradePayableUsecase) Cancel(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNonTradePayableNotFound
		}
		return nil, err
	}

	if item.Status == financeModels.NTPStatusPaid {
		return nil, errors.New("paid non-trade payable cannot be cancelled")
	}

	if item.Status != financeModels.NTPStatusCancelled {
		if err := uc.db.WithContext(ctx).Model(&financeModels.NonTradePayable{}).
			Where("id = ?", id).
			Update("status", financeModels.NTPStatusCancelled).Error; err != nil {
			return nil, err
		}
		item.Status = financeModels.NTPStatusCancelled
	}

	res := uc.mapper.ToResponse(item)
	paidAmount, err := uc.getPaidAmount(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	attachOutstanding(&res, item.Amount, paidAmount)
	return &res, nil
}

func (uc *nonTradePayableUsecase) Submit(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	return uc.Post(ctx, id)
}

func (uc *nonTradePayableUsecase) Approve(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	return uc.Post(ctx, id)
}

func (uc *nonTradePayableUsecase) Reject(ctx context.Context, id string) (*dto.NonTradePayableResponse, error) {
	return uc.Cancel(ctx, id)
}

func (uc *nonTradePayableUsecase) Pay(ctx context.Context, id string, req *dto.PayNonTradePayableRequest) (*dto.NonTradePayableResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}
	if strings.TrimSpace(req.PaymentDate) == "" {
		return nil, errors.New("payment_date is required")
	}
	paymentDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.PaymentDate))
	if err != nil {
		return nil, errors.New("invalid payment_date")
	}

	paymentCOAID, err := uc.resolvePaymentCOAID(ctx, req)
	if err != nil {
		return nil, err
	}

	var item financeModels.NonTradePayable
	var journalID string
	var paidAmount float64
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("ChartOfAccount").
			First(&item, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNonTradePayableNotFound
			}
			return err
		}

		if item.Status == financeModels.NTPStatusCancelled || item.Status == financeModels.NTPStatusRejected {
			return errors.New("cancelled non-trade payable cannot be paid")
		}
		if item.Status == financeModels.NTPStatusDraft || item.Status == financeModels.NTPStatusSubmitted {
			return errors.New("non-trade payable must be posted before payment")
		}

		if err := ensureNotClosed(txCtx, tx, paymentDate); err != nil {
			return err
		}

		paidBefore, err := uc.getPaidAmount(txCtx, item.ID)
		if err != nil {
			return err
		}
		remaining := item.Amount - paidBefore
		if remaining < 0 && math.Abs(remaining) < 0.01 {
			remaining = 0
		}
		if remaining <= 0 {
			return errors.New("non-trade payable is already fully paid")
		}
		if req.Amount > remaining+0.01 {
			return fmt.Errorf("payment amount exceeds outstanding amount %.2f", remaining)
		}

		txData := accounting.TransactionData{
			ReferenceType:       reference.RefTypeNTPPayment,
			ReferenceID:         uuid.NewString(),
			EntryDate:           req.PaymentDate,
			Description:         "NTP Payment: " + item.Code + " - Ref: " + req.BankReference,
			TotalAmount:         req.Amount,
			PaymentAccountCOAID: paymentCOAID,
			MemoArgs:            []interface{}{req.BankReference},
		}

		journalReq, err := uc.accountingEngine.GenerateJournal(txCtx, accounting.ProfileNonTradePayablePayment, txData)
		if err != nil {
			return err
		}
		journalReq.Reference = item.ID

		journalRes, err := uc.journalUC.PostOrUpdateJournal(txCtx, journalReq)
		if err != nil {
			return err
		}
		journalID = journalRes.ID

		paidAmount = paidBefore + req.Amount
		nextStatus := financeModels.NTPStatusPartial
		if paidAmount >= item.Amount-0.01 {
			nextStatus = financeModels.NTPStatusPaid
		}

		if err := tx.WithContext(ctx).Model(&financeModels.NonTradePayable{}).
			Where("id = ?", id).
			Update("status", nextStatus).Error; err != nil {
			return err
		}
		item.Status = nextStatus
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(&item)
	res.JournalID = &journalID
	attachOutstanding(&res, item.Amount, paidAmount)
	return &res, nil
}
