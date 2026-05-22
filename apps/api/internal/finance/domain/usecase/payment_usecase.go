package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrPaymentPostedImmutable    = errors.New("posted payment cannot be modified")
	ErrPaymentInvalidAllocations = errors.New("invalid payment allocations")
	ErrPaymentTotalMismatch      = errors.New("total_amount must equal allocation sum")
)

type PaymentUsecase interface {
	Create(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.PaymentResponse, error)
	List(ctx context.Context, req *dto.ListPaymentsRequest) ([]dto.PaymentResponse, int64, error)
	Approve(ctx context.Context, id string) (*dto.PaymentResponse, error)
	GetFormData(ctx context.Context) (*dto.PaymentFormDataResponse, error)
	Reverse(ctx context.Context, id string, reason string) (*dto.PaymentResponse, error)
}

type paymentUsecase struct {
	db        *gorm.DB
	coaRepo   repositories.ChartOfAccountRepository
	repo      repositories.PaymentRepository
	journalUC JournalEntryUsecase
	mapper    *mapper.PaymentMapper
}

func NewPaymentUsecase(db *gorm.DB, coaRepo repositories.ChartOfAccountRepository, repo repositories.PaymentRepository, journalUC JournalEntryUsecase, mapper *mapper.PaymentMapper) PaymentUsecase {
	return &paymentUsecase{db: db, coaRepo: coaRepo, repo: repo, journalUC: journalUC, mapper: mapper}
}

// parseDateStrict is consolidated in helpers.go and is used throughout the finance package

func validateAllocations(allocs []dto.PaymentAllocationRequest) (float64, error) {
	if len(allocs) < 1 {
		return 0, ErrPaymentInvalidAllocations
	}
	var sum float64
	for _, a := range allocs {
		if strings.TrimSpace(a.ChartOfAccountID) == "" {
			return 0, ErrPaymentInvalidAllocations
		}
		if a.ReferenceType == nil || strings.TrimSpace(*a.ReferenceType) == "" {
			return 0, errors.New("reference_type is required for all allocations")
		}
		if a.ReferenceID == nil || strings.TrimSpace(*a.ReferenceID) == "" {
			return 0, errors.New("reference_id is required for all allocations")
		}
		if a.Amount <= 0 {
			return 0, ErrPaymentInvalidAllocations
		}
		sum += a.Amount
	}
	return sum, nil
}

func paymentAmountKey(v float64) string {
	return strconv.FormatFloat(v, 'f', 6, 64)
}

func (uc *paymentUsecase) Create(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	payDate, err := parseDate(req.PaymentDate)
	if err != nil {
		return nil, err
	}

	sum, err := validateAllocations(req.Allocations)
	if err != nil {
		return nil, err
	}
	if math.Abs(sum-req.TotalAmount) > 0.000001 {
		return nil, ErrPaymentTotalMismatch
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	bankAccountID := strings.TrimSpace(req.BankAccountID)
	if bankAccountID == "" {
		return nil, errors.New("bank_account_id is required")
	}
	var bank coreModels.BankAccount
	if err := uc.db.WithContext(ctx).First(&bank, "id = ?", bankAccountID).Error; err != nil {
		return nil, err
	}

	var createdID string
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coaIDs := make([]string, 0, len(req.Allocations))
		for _, al := range req.Allocations {
			coaIDs = append(coaIDs, strings.TrimSpace(al.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, al := range req.Allocations {
			if coaByID[strings.TrimSpace(al.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		p := &financeModels.Payment{
			PaymentDate:                 payDate,
			Description:                 strings.TrimSpace(req.Description),
			BankAccountID:               bankAccountID,
			BankAccountNameSnapshot:     strings.TrimSpace(bank.Name),
			BankAccountNumberSnapshot:   strings.TrimSpace(bank.AccountNumber),
			BankAccountHolderSnapshot:   strings.TrimSpace(bank.AccountHolder),
			BankAccountCurrencySnapshot: strings.TrimSpace(bank.Currency),
			TotalAmount:                 req.TotalAmount,
			Status:                      financeModels.PaymentStatusDraft,
			CreatedBy:                   &actorID,
		}
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for _, al := range req.Allocations {
			coa := coaByID[strings.TrimSpace(al.ChartOfAccountID)]
			codeSnap := ""
			nameSnap := ""
			typeSnap := ""
			snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			item := &financeModels.PaymentAllocation{
				PaymentID:                  p.ID,
				ChartOfAccountID:           strings.TrimSpace(al.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				ReferenceType:              al.ReferenceType,
				ReferenceID:                al.ReferenceID,
				Amount:                     al.Amount,
				Memo:                       strings.TrimSpace(al.Memo),
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		createdID = p.ID
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

func (uc *paymentUsecase) Update(ctx context.Context, id string, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	p, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if p.Status == financeModels.PaymentStatusPosted || p.Status == financeModels.PaymentStatusReversed {
		return nil, ErrPaymentPostedImmutable
	}

	payDate, err := parseDate(req.PaymentDate)
	if err != nil {
		return nil, err
	}

	sum, err := validateAllocations(req.Allocations)
	if err != nil {
		return nil, err
	}
	if math.Abs(sum-req.TotalAmount) > 0.000001 {
		return nil, ErrPaymentTotalMismatch
	}

	bankAccountID := strings.TrimSpace(req.BankAccountID)
	if bankAccountID == "" {
		return nil, errors.New("bank_account_id is required")
	}
	var bank coreModels.BankAccount
	if err := uc.db.WithContext(ctx).First(&bank, "id = ?", bankAccountID).Error; err != nil {
		return nil, err
	}

	bankNameSnap := p.BankAccountNameSnapshot
	bankNumberSnap := p.BankAccountNumberSnapshot
	bankHolderSnap := p.BankAccountHolderSnapshot
	bankCurrencySnap := p.BankAccountCurrencySnapshot
	if strings.TrimSpace(bankAccountID) != strings.TrimSpace(p.BankAccountID) {
		bankNameSnap = strings.TrimSpace(bank.Name)
		bankNumberSnap = strings.TrimSpace(bank.AccountNumber)
		bankHolderSnap = strings.TrimSpace(bank.AccountHolder)
		bankCurrencySnap = strings.TrimSpace(bank.Currency)
	}

	existingAllocSnap := map[string]financeModels.PaymentAllocation{}
	for _, al := range p.Allocations {
		refType := ""
		if al.ReferenceType != nil {
			refType = strings.TrimSpace(*al.ReferenceType)
		}
		refID := ""
		if al.ReferenceID != nil {
			refID = strings.TrimSpace(*al.ReferenceID)
		}
		k := strings.TrimSpace(al.ChartOfAccountID) + "|" + refType + "|" + refID + "|" + paymentAmountKey(al.Amount) + "|" + strings.TrimSpace(al.Memo)
		existingAllocSnap[k] = al
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coaIDs := make([]string, 0, len(req.Allocations))
		for _, al := range req.Allocations {
			coaIDs = append(coaIDs, strings.TrimSpace(al.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, al := range req.Allocations {
			if coaByID[strings.TrimSpace(al.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		if err := tx.Model(&financeModels.Payment{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"payment_date":                   payDate,
				"description":                    strings.TrimSpace(req.Description),
				"bank_account_id":                bankAccountID,
				"bank_account_name_snapshot":     bankNameSnap,
				"bank_account_number_snapshot":   bankNumberSnap,
				"bank_account_holder_snapshot":   bankHolderSnap,
				"bank_account_currency_snapshot": bankCurrencySnap,
				"total_amount":                   req.TotalAmount,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("payment_id = ?", id).Delete(&financeModels.PaymentAllocation{}).Error; err != nil {
			return err
		}
		for _, al := range req.Allocations {
			refType := ""
			if al.ReferenceType != nil {
				refType = strings.TrimSpace(*al.ReferenceType)
			}
			refID := ""
			if al.ReferenceID != nil {
				refID = strings.TrimSpace(*al.ReferenceID)
			}
			k := strings.TrimSpace(al.ChartOfAccountID) + "|" + refType + "|" + refID + "|" + paymentAmountKey(al.Amount) + "|" + strings.TrimSpace(al.Memo)
			existing := existingAllocSnap[k]

			codeSnap := strings.TrimSpace(existing.ChartOfAccountCodeSnapshot)
			nameSnap := strings.TrimSpace(existing.ChartOfAccountNameSnapshot)
			typeSnap := strings.TrimSpace(existing.ChartOfAccountTypeSnapshot)
			if codeSnap == "" && nameSnap == "" && typeSnap == "" {
				coa := coaByID[strings.TrimSpace(al.ChartOfAccountID)]
				snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			}

			item := &financeModels.PaymentAllocation{
				PaymentID:                  id,
				ChartOfAccountID:           strings.TrimSpace(al.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				ReferenceType:              al.ReferenceType,
				ReferenceID:                al.ReferenceID,
				Amount:                     al.Amount,
				Memo:                       strings.TrimSpace(al.Memo),
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

func (uc *paymentUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}

	p, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPaymentNotFound
		}
		return err
	}
	if p.Status == financeModels.PaymentStatusPosted || p.Status == financeModels.PaymentStatusReversed {
		return ErrPaymentPostedImmutable
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.Payment{}, "id = ?", id).Error
}

func (uc *paymentUsecase) GetByID(ctx context.Context, id string) (*dto.PaymentResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.Payment{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrPaymentNotFound
	}
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	resp := uc.mapper.ToResponse(item)
	return &resp, nil
}

func (uc *paymentUsecase) List(ctx context.Context, req *dto.ListPaymentsRequest) ([]dto.PaymentResponse, int64, error) {
	if req == nil {
		req = &dto.ListPaymentsRequest{}
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

	items, total, err := uc.repo.List(ctx, repositories.PaymentListParams{
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

	res := make([]dto.PaymentResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}

func (uc *paymentUsecase) Approve(ctx context.Context, id string) (*dto.PaymentResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var postedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p financeModels.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Allocations").
			First(&p, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPaymentNotFound
			}
			return err
		}

		if p.Status == financeModels.PaymentStatusPosted {
			postedID = p.ID
			return nil
		}

		var bank coreModels.BankAccount
		if err := tx.First(&bank, "id = ?", p.BankAccountID).Error; err != nil {
			return err
		}
		if bank.ChartOfAccountID == nil || strings.TrimSpace(*bank.ChartOfAccountID) == "" {
			return errors.New("bank account is not linked to chart_of_account_id")
		}
		bankCOAID := strings.TrimSpace(*bank.ChartOfAccountID)

		if err := ensureNotClosed(ctx, tx, p.PaymentDate); err != nil {
			return err
		}
		now := apptime.Now()

		refType := reference.RefTypePayment

		journalLines := make([]dto.JournalLineRequest, 0, len(p.Allocations)+1)
		for _, al := range p.Allocations {
			if al.Amount > 0 {
				if err := EnsureWithinBudget(ctx, tx, al.ChartOfAccountID, p.PaymentDate, al.Amount); err != nil {
					return fmt.Errorf("budget check failed for account %s: %w", al.ChartOfAccountID, err)
				}
			}
			journalLines = append(journalLines, dto.JournalLineRequest{
				ChartOfAccountID: al.ChartOfAccountID,
				Debit:            al.Amount,
				Credit:           0,
				Memo:             strings.TrimSpace(al.Memo),
			})
		}
		journalLines = append(journalLines, dto.JournalLineRequest{
			ChartOfAccountID: bankCOAID,
			Debit:            0,
			Credit:           p.TotalAmount,
			Memo:             "Payment bank outflow",
		})

		reqJournal := &dto.CreateJournalEntryRequest{
			EntryDate:         p.PaymentDate.Format("2006-01-02"),
			Description:       strings.TrimSpace(p.Description),
			ReferenceType:     &refType,
			ReferenceID:       &p.ID,
			Lines:             journalLines,
			IsSystemGenerated: true,
		}

		txCtx := database.WithTx(ctx, tx)
		journalResp, err := uc.journalUC.PostOrUpdateJournal(txCtx, reqJournal)
		if err != nil {
			return fmt.Errorf("failed to generate journal: %w", err)
		}

		if err := tx.Model(&financeModels.Payment{}).
			Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status":           financeModels.PaymentStatusPosted,
				"journal_entry_id": journalResp.ID,
				"approved_at":      now,
				"approved_by":      actorID,
				"posted_at":        now,
				"posted_by":        actorID,
			}).Error; err != nil {
			return err
		}

		postedID = p.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, postedID, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

func (uc *paymentUsecase) Reverse(ctx context.Context, id string, reason string) (*dto.PaymentResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var reversedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p financeModels.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&p, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPaymentNotFound
			}
			return err
		}

		if p.Status == financeModels.PaymentStatusReversed {
			reversedID = p.ID
			return nil
		}

		if p.Status != financeModels.PaymentStatusPosted {
			return errors.New("only posted payments can be reversed")
		}

		if p.JournalEntryID == nil {
			return errors.New("no associated journal entry found to reverse")
		}

		if err := ensureNotClosed(ctx, tx, p.PaymentDate); err != nil {
			return err
		}

		txCtx := database.WithTx(ctx, tx)
		_, err := uc.journalUC.ReverseWithReason(txCtx, *p.JournalEntryID, reason)
		if err != nil {
			return fmt.Errorf("failed to reverse journal: %w", err)
		}

		if err := tx.Model(&financeModels.Payment{}).
			Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status": financeModels.PaymentStatusReversed,
			}).Error; err != nil {
			return err
		}

		reversedID = p.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, reversedID, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}
