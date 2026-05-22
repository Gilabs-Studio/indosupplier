package usecase

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"gorm.io/gorm"
)

var (
	ErrCashBankNotFound        = errors.New("cash bank journal not found")
	ErrCashBankPostedImmutable = errors.New("posted cash bank journal cannot be modified")
	ErrCashBankInvalidLines    = errors.New("invalid cash bank journal lines")
	ErrControlAccountRestricted = errors.New("restricted: trade control accounts (AR/AP/Inventory) cannot be used in manual bank journals. please use the respective business modules (Sales/Purchase)")
)

type CashBankJournalUsecase interface {
	Create(ctx context.Context, req *dto.CreateCashBankJournalRequest) (*dto.CashBankJournalResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCashBankJournalRequest) (*dto.CashBankJournalResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.CashBankJournalResponse, error)
	List(ctx context.Context, req *dto.ListCashBankJournalsRequest) ([]dto.CashBankJournalResponse, int64, error)
	ListLines(ctx context.Context, cashBankJournalID string, page, perPage int) (*dto.ListJournalLinesResponse, int64, error)
	Post(ctx context.Context, id string) (*dto.CashBankJournalResponse, error)
	GetFormData(ctx context.Context) (*dto.CashBankFormDataResponse, error)
	// Sub-ledger: read-only access for Journal module
	ListPosted(ctx context.Context, req *dto.ListCashBankJournalsRequest) ([]dto.CashBankJournalResponse, int64, *dto.CashBankSubLedgerKPI, error)
}

type cashBankJournalUsecase struct {
	db               *gorm.DB
	coaRepo          repositories.ChartOfAccountRepository
	repo             repositories.CashBankJournalRepository
	journalUC        JournalEntryUsecase
	mapper           *mapper.CashBankJournalMapper
	settingsService  financesettings.SettingsService
	accountingEngine accounting.AccountingEngine
}

func NewCashBankJournalUsecase(
	db *gorm.DB,
	coaRepo repositories.ChartOfAccountRepository,
	repo repositories.CashBankJournalRepository,
	journalUC JournalEntryUsecase,
	mapper *mapper.CashBankJournalMapper,
	settingsService financesettings.SettingsService,
	accountingEngine accounting.AccountingEngine,
) CashBankJournalUsecase {
	return &cashBankJournalUsecase{
		db:               db,
		coaRepo:          coaRepo,
		repo:             repo,
		journalUC:        journalUC,
		mapper:           mapper,
		settingsService:  settingsService,
		accountingEngine: accountingEngine,
	}
}

func validateCashBankLines(lines []dto.CashBankJournalLineRequest) (float64, error) {
	if len(lines) < 1 {
		return 0, ErrCashBankInvalidLines
	}
	var sum float64
	for _, ln := range lines {
		if strings.TrimSpace(ln.ChartOfAccountID) == "" {
			return 0, ErrCashBankInvalidLines
		}
		if ln.Amount <= 0 {
			return 0, ErrCashBankInvalidLines
		}
		sum += ln.Amount
	}
	return sum, nil
}

func (uc *cashBankJournalUsecase) Create(ctx context.Context, req *dto.CreateCashBankJournalRequest) (*dto.CashBankJournalResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	trxDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.TransactionDate))
	if err != nil {
		return nil, errors.New("invalid transaction_date")
	}
	if req.Type != financeModels.CashBankTypeCashIn && req.Type != financeModels.CashBankTypeCashOut && req.Type != financeModels.CashBankTypeTransfer {
		return nil, errors.New("invalid type: must be cash_in, cash_out, or transfer")
	}

	sum, err := validateCashBankLines(req.Lines)
	if err != nil {
		return nil, err
	}

	if err := uc.validateControlAccounts(ctx, req.Lines); err != nil {
		return nil, err
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
		coaIDs := make([]string, 0, len(req.Lines))
		for _, ln := range req.Lines {
			coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, ln := range req.Lines {
			if coaByID[strings.TrimSpace(ln.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		cb := &financeModels.CashBankJournal{
			TransactionDate:             trxDate,
			Type:                        req.Type,
			Description:                 strings.TrimSpace(req.Description),
			BankAccountID:               bankAccountID,
			TotalAmount:                 sum,
			Status:                      financeModels.CashBankStatusDraft,
			CreatedBy:                   &actorID,
		}
		if err := tx.Create(cb).Error; err != nil {
			return err
		}
		for _, ln := range req.Lines {
			coa := coaByID[strings.TrimSpace(ln.ChartOfAccountID)]
			codeSnap := ""
			nameSnap := ""
			typeSnap := ""
			snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			item := &financeModels.CashBankJournalLine{
				CashBankJournalID:          cb.ID,
				ChartOfAccountID:           strings.TrimSpace(ln.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				ReferenceType:              ln.ReferenceType,
				ReferenceID:                ln.ReferenceID,
				Amount:                     ln.Amount,
				Memo:                       strings.TrimSpace(ln.Memo),
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		createdID = cb.ID
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

func (uc *cashBankJournalUsecase) Update(ctx context.Context, id string, req *dto.UpdateCashBankJournalRequest) (*dto.CashBankJournalResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	cb, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCashBankNotFound
		}
		return nil, err
	}
	if cb.Status == financeModels.CashBankStatusPosted {
		return nil, ErrCashBankPostedImmutable
	}

	trxDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.TransactionDate))
	if err != nil {
		return nil, errors.New("invalid transaction_date")
	}
	if req.Type != financeModels.CashBankTypeCashIn && req.Type != financeModels.CashBankTypeCashOut && req.Type != financeModels.CashBankTypeTransfer {
		return nil, errors.New("invalid type: must be cash_in, cash_out, or transfer")
	}

	sum, err := validateCashBankLines(req.Lines)
	if err != nil {
		return nil, err
	}
	if math.Abs(sum) < 0.000001 {
		return nil, ErrCashBankInvalidLines
	}

	if err := uc.validateControlAccounts(ctx, req.Lines); err != nil {
		return nil, err
	}

	bankAccountID := strings.TrimSpace(req.BankAccountID)
	if bankAccountID == "" {
		return nil, errors.New("bank_account_id is required")
	}
	var bank coreModels.BankAccount
	if err := uc.db.WithContext(ctx).First(&bank, "id = ?", bankAccountID).Error; err != nil {
		return nil, err
	}

	bankNameSnap := cb.BankAccountNameSnapshot
	bankNumberSnap := cb.BankAccountNumberSnapshot
	bankHolderSnap := cb.BankAccountHolderSnapshot
	bankCurrencySnap := cb.BankAccountCurrencySnapshot
	if strings.TrimSpace(bankAccountID) != strings.TrimSpace(cb.BankAccountID) {
		bankNameSnap = strings.TrimSpace(bank.Name)
		bankNumberSnap = strings.TrimSpace(bank.AccountNumber)
		bankHolderSnap = strings.TrimSpace(bank.AccountHolder)
		bankCurrencySnap = strings.TrimSpace(bank.Currency)
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		coaIDs := make([]string, 0, len(req.Lines))
		for _, ln := range req.Lines {
			coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, ln := range req.Lines {
			if coaByID[strings.TrimSpace(ln.ChartOfAccountID)] == nil {
				return errors.New("chart of account not found")
			}
		}

		if err := tx.Model(&financeModels.CashBankJournal{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"transaction_date":               trxDate,
				"type":                           req.Type,
				"description":                    strings.TrimSpace(req.Description),
				"bank_account_id":                bankAccountID,
				"bank_account_name_snapshot":     bankNameSnap,
				"bank_account_number_snapshot":   bankNumberSnap,
				"bank_account_holder_snapshot":   bankHolderSnap,
				"bank_account_currency_snapshot": bankCurrencySnap,
				"total_amount":                   sum,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("cash_bank_journal_id = ?", id).Delete(&financeModels.CashBankJournalLine{}).Error; err != nil {
			return err
		}
		for _, ln := range req.Lines {
			coa := coaByID[strings.TrimSpace(ln.ChartOfAccountID)]
			codeSnap := ""
			nameSnap := ""
			typeSnap := ""
			snapshotCOAIntoLine(&codeSnap, &nameSnap, &typeSnap, coa)
			item := &financeModels.CashBankJournalLine{
				CashBankJournalID:          id,
				ChartOfAccountID:           strings.TrimSpace(ln.ChartOfAccountID),
				ChartOfAccountCodeSnapshot: codeSnap,
				ChartOfAccountNameSnapshot: nameSnap,
				ChartOfAccountTypeSnapshot: typeSnap,
				ReferenceType:              ln.ReferenceType,
				ReferenceID:                ln.ReferenceID,
				Amount:                     ln.Amount,
				Memo:                       strings.TrimSpace(ln.Memo),
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

func (uc *cashBankJournalUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	cb, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrCashBankNotFound
		}
		return err
	}
	if cb.Status == financeModels.CashBankStatusPosted {
		return ErrCashBankPostedImmutable
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.CashBankJournal{}, "id = ?", id).Error
}

func (uc *cashBankJournalUsecase) GetByID(ctx context.Context, id string) (*dto.CashBankJournalResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.CashBankJournal{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrCashBankNotFound
	}
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCashBankNotFound
		}
		return nil, err
	}
	resp := uc.mapper.ToResponse(item)
	return &resp, nil
}

func (uc *cashBankJournalUsecase) List(ctx context.Context, req *dto.ListCashBankJournalsRequest) ([]dto.CashBankJournalResponse, int64, error) {
	if req == nil {
		req = &dto.ListCashBankJournalsRequest{}
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

	startDate, err := parseDateOptional(req.StartDate)
	if err != nil {
		return nil, 0, err
	}
	endDate, err := parseEndDateOptional(req.EndDate)
	if err != nil {
		return nil, 0, err
	}

	items, total, err := uc.repo.List(ctx, repositories.CashBankJournalListParams{
		Search:        req.Search,
		Type:          req.Type,
		Status:        req.Status,
		BankAccountID: req.BankAccountID,
		StartDate:     startDate,
		EndDate:       endDate,
		SortBy:        req.SortBy,
		SortDir:       req.SortDir,
		Limit:         perPage,
		Offset:        (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.CashBankJournalResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}

func (uc *cashBankJournalUsecase) ListLines(ctx context.Context, cashBankJournalID string, page, perPage int) (*dto.ListJournalLinesResponse, int64, error) {
	cashBankJournalID = strings.TrimSpace(cashBankJournalID)
	if cashBankJournalID == "" {
		return nil, 0, errors.New("cash_bank_journal_id is required")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	cb, err := uc.repo.FindByID(ctx, cashBankJournalID, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrCashBankNotFound
		}
		return nil, 0, err
	}

	total := int64(len(cb.Lines))
	start := (page - 1) * perPage
	end := start + perPage
	if int64(start) > total {
		start = int(total)
	}
	if int64(end) > total {
		end = int(total)
	}

	lines := cb.Lines[start:end]
	resLines := make([]dto.JournalLineDetailResponse, 0, len(lines))
	for _, ln := range lines {
		var debit, credit float64
		if cb.Type == financeModels.CashBankTypeCashIn {
			debit = ln.Amount
		} else {
			credit = ln.Amount
		}

		resLines = append(resLines, dto.JournalLineDetailResponse{
			ID:                 ln.ID,
			JournalEntryID:     cb.ID,
			EntryDate:          cb.TransactionDate.Format("2006-01-02"),
			JournalDescription: cb.Description,
			JournalStatus:      string(cb.Status),
			ReferenceType:      ln.ReferenceType,
			ReferenceID:        ln.ReferenceID,
			ChartOfAccountID:   ln.ChartOfAccountID,
			ChartOfAccountCode: ln.ChartOfAccountCodeSnapshot,
			ChartOfAccountName: ln.ChartOfAccountNameSnapshot,
			ChartOfAccountType: ln.ChartOfAccountTypeSnapshot,
			Debit:              debit,
			Credit:             credit,
			Memo:               ln.Memo,
			RunningBalance:     0,
			CreatedAt:          ln.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		})
	}

	resp := &dto.ListJournalLinesResponse{
		Lines:       resLines,
		TotalDebit:  0,
		TotalCredit: 0,
	}

	return resp, total, nil
}

func (uc *cashBankJournalUsecase) Post(ctx context.Context, id string) (*dto.CashBankJournalResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	cb, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCashBankNotFound
		}
		return nil, err
	}
	if cb.Status == financeModels.CashBankStatusPosted {
		resp := uc.mapper.ToResponse(cb)
		return &resp, nil
	}

	var bank coreModels.BankAccount
	if err := uc.db.WithContext(ctx).First(&bank, "id = ?", cb.BankAccountID).Error; err != nil {
		return nil, err
	}
	if bank.ChartOfAccountID == nil || strings.TrimSpace(*bank.ChartOfAccountID) == "" {
		return nil, errors.New("bank account is not linked to chart_of_account_id")
	}
	bankCOAID := strings.TrimSpace(*bank.ChartOfAccountID)

	var txData accounting.TransactionData
	txData.ReferenceType = reference.RefTypeCashBank
	txData.ReferenceID = cb.ID
	txData.EntryDate = cb.TransactionDate.Format("2006-01-02")
	txData.Description = strings.TrimSpace(cb.Description)
	txData.TotalAmount = cb.TotalAmount
	txData.BankAccountCOAID = bankCOAID
	txData.DescriptionArgs = []interface{}{txData.Description}

	for _, ln := range cb.Lines {
		txData.LineItems = append(txData.LineItems, accounting.TransactionLineItem{
			ChartOfAccountID: ln.ChartOfAccountID,
			Amount:           ln.Amount,
			Memo:             ln.Memo,
		})
	}

	var profile accounting.PostingProfile
	switch cb.Type {
	case financeModels.CashBankTypeCashIn:
		profile = accounting.ProfileCashBankCashIn
	case financeModels.CashBankTypeTransfer:
		profile = accounting.ProfileCashBankTransfer
	default:
		profile = accounting.ProfileCashBankCashOut
	}

	journalReq, err := uc.accountingEngine.GenerateJournal(ctx, profile, txData)
	if err != nil {
		return nil, err
	}

	journalRes, err := uc.journalUC.PostOrUpdateJournal(ctx, journalReq)
	if err != nil {
		return nil, err
	}

	now := apptime.Now()
	if err := uc.db.WithContext(ctx).Model(&financeModels.CashBankJournal{}).
		Where("id = ?", cb.ID).
		Updates(map[string]interface{}{
			"status":           financeModels.CashBankStatusPosted,
			"journal_entry_id": journalRes.ID,
			"posted_at":        now,
			"posted_by":        actorID,
		}).Error; err != nil {
		return nil, err
	}

	postedID := cb.ID

	full, err := uc.repo.FindByID(ctx, postedID, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

// ListPosted returns only posted cash bank journals for the Journal module sub-ledger.
// It also computes KPI: total inflow (cash_in), total outflow (cash_out + transfer), and net movement.
func (uc *cashBankJournalUsecase) ListPosted(ctx context.Context, req *dto.ListCashBankJournalsRequest) ([]dto.CashBankJournalResponse, int64, *dto.CashBankSubLedgerKPI, error) {
	if req == nil {
		req = &dto.ListCashBankJournalsRequest{}
	}

	// Force status to posted only
	postedStatus := financeModels.CashBankStatusPosted
	req.Status = &postedStatus

	items, total, err := uc.List(ctx, req)
	if err != nil {
		return nil, 0, nil, err
	}

	// Compute KPI from results
	var inflow, outflow float64
	for _, item := range items {
		switch item.Type {
		case financeModels.CashBankTypeCashIn:
			inflow += item.TotalAmount
		case financeModels.CashBankTypeCashOut, financeModels.CashBankTypeTransfer:
			outflow += item.TotalAmount
		}
	}

	kpi := &dto.CashBankSubLedgerKPI{
		TotalInflow:  inflow,
		TotalOutflow: outflow,
		NetMovement:  inflow - outflow,
		TotalRecords: total,
	}

	return items, total, kpi, nil
}

func (uc *cashBankJournalUsecase) validateControlAccounts(ctx context.Context, lines []dto.CashBankJournalLineRequest) error {
	restrictedKeys := []string{
		financeModels.SettingCOASalesReceivable,
		financeModels.SettingCOASalesAdvance,
		financeModels.SettingCOAPurchasePayable,
		financeModels.SettingCOAPurchaseAdvance,
		financeModels.SettingCOAPurchaseGRIR,
		financeModels.SettingCOAInventory,
	}

	restrictedCodes := make(map[string]bool)
	for _, key := range restrictedKeys {
		code, err := uc.settingsService.GetCOACode(ctx, key)
		if err == nil && code != "" {
			restrictedCodes[strings.TrimSpace(code)] = true
		}
	}

	coaIDs := make([]string, 0, len(lines))
	for _, ln := range lines {
		coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
	}

	var coas []financeModels.ChartOfAccount
	if err := uc.db.WithContext(ctx).Where("id IN ?", coaIDs).Find(&coas).Error; err != nil {
		return err
	}

	for _, coa := range coas {
		if restrictedCodes[strings.TrimSpace(coa.Code)] {
			return ErrControlAccountRestricted
		}
	}

	return nil
}
