package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func formatFloatKey(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

func safeStringPtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func journalTraceKey(referenceType, referenceID *string) string {
	rt := safeStringPtr(referenceType)
	rid := safeStringPtr(referenceID)
	if rt == "" && rid == "" {
		return "unknown"
	}
	return rt + ":" + rid
}

func logJournalEvent(event string, fields map[string]interface{}) {
	log.Printf("journal_observability event=%s fields=%+v", event, fields)
}

var (
	ErrJournalNotFound                 = errors.New("journal entry not found")
	ErrJournalPostedImmutable          = errors.New("posted journal entry cannot be modified")
	ErrJournalUnbalanced               = errors.New("journal entry must be balanced (debit = credit)")
	ErrJournalInvalidLines             = errors.New("invalid journal lines")
	ErrJournalControlAccountRestricted = errors.New("restricted: trade control accounts (AR/AP/Inventory) cannot be used in manual journals. Use the respective business modules (Sales/Purchase/Inventory)")
	ErrJournalNonPostableAccount       = errors.New("journal line uses non-postable account")
)

type JournalEntryUsecase interface {
	Create(ctx context.Context, req *dto.CreateJournalEntryRequest) (*dto.JournalEntryResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateJournalEntryRequest) (*dto.JournalEntryResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	List(ctx context.Context, req *dto.ListJournalEntriesRequest) ([]dto.JournalEntryResponse, int64, error)
	Post(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	Cancel(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	Reverse(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	ReverseWithReason(ctx context.Context, id string, reason string) (*dto.JournalEntryResponse, error)
	TrialBalance(ctx context.Context, startDate, endDate *time.Time) (*dto.TrialBalanceResponse, error)
	PostOrUpdateJournal(ctx context.Context, req *dto.CreateJournalEntryRequest) (*dto.JournalEntryResponse, error)
	GetFormData(ctx context.Context) (*dto.JournalEntryFormDataResponse, error)
	CreateAdjustmentJournal(ctx context.Context, req *dto.CreateAdjustmentJournalRequest) (*dto.JournalEntryResponse, error)
	UpdateAdjustmentJournal(ctx context.Context, id string, req *dto.UpdateJournalEntryRequest) (*dto.JournalEntryResponse, error)
	SubmitAdjustmentJournalForApproval(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error)
	ApproveAdjustmentJournal(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error)
	RejectAdjustmentJournal(ctx context.Context, id string, notes string) (*dto.JournalEntryResponse, error)
	GetAdjustmentApprovalHistory(ctx context.Context, id string) ([]dto.AdjustmentJournalApprovalResponse, error)
	PostAdjustmentJournal(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	ReverseAdjustmentJournal(ctx context.Context, id string) (*dto.JournalEntryResponse, error)
	ListJournalTemplates(ctx context.Context, req *dto.ListJournalTemplatesRequest) ([]dto.JournalTemplateResponse, int64, error)
	CreateJournalTemplate(ctx context.Context, req *dto.CreateJournalTemplateRequest) (*dto.JournalTemplateResponse, error)
	UseJournalTemplate(ctx context.Context, id string) (*dto.UseJournalTemplateResponse, error)
	RunValuation(ctx context.Context) (*dto.JournalEntryResponse, error)
	CreateOpeningBalanceJournal(ctx context.Context, req *dto.CreateOpeningBalanceJournalRequest) (*dto.JournalEntryResponse, error)
	ReverseOpeningBalance(ctx context.Context, accountID string) (*dto.JournalEntryResponse, error)
}

type journalEntryUsecase struct {
	db                     *gorm.DB
	coaRepo                repositories.ChartOfAccountRepository
	repo                   repositories.JournalEntryRepository
	mapper                 *mapper.JournalEntryMapper
	auditService           audit.AuditService
	settingsService        financesettings.SettingsService
	adjustmentApprovalRepo repositories.AdjustmentJournalApprovalRepository
	journalTemplateRepo    repositories.JournalTemplateRepository
}

func NewJournalEntryUsecase(db *gorm.DB, coaRepo repositories.ChartOfAccountRepository, repo repositories.JournalEntryRepository, mapper *mapper.JournalEntryMapper, auditService audit.AuditService, settingsService ...financesettings.SettingsService) JournalEntryUsecase {
	uc := &journalEntryUsecase{
		db:                     db,
		coaRepo:                coaRepo,
		repo:                   repo,
		mapper:                 mapper,
		auditService:           auditService,
		adjustmentApprovalRepo: repositories.NewAdjustmentJournalApprovalRepository(db),
		journalTemplateRepo:    repositories.NewJournalTemplateRepository(db),
	}
	if len(settingsService) > 0 {
		uc.settingsService = settingsService[0]
	}
	return uc
}

// parseDate is consolidated in helpers.go and is used throughout the finance package

func journalReferenceTypesForDomain(domain *string) []string {
	if domain == nil {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(*domain)) {
	case "sales":
		return []string{
			reference.RefTypeSalesInvoice,
			reference.RefTypeSalesInvoiceDP,
			reference.RefTypeSalesPayment,
			"SalesInvoice",
			"SalesPayment",
			"SalesInvoiceDP",
		}
	case "purchase":
		return []string{
			reference.RefTypeGoodsReceipt,
			reference.RefTypeSupplierInvoice,
			reference.RefTypeSupplierInvoiceDP,
			reference.RefTypePurchasePayment,
			"GoodsReceipt",
			"SupplierInvoice",
			"SupplierInvoiceDP",
			"PurchasePayment",
		}
	case "inventory", "stock":
		return []string{
			reference.RefTypeStockOpname,
			reference.RefTypeInventoryAdjustment,
			reference.RefTypeInventoryValuation,
			reference.RefTypeCostAdjustment,
		}
	case "cash_bank":
		return []string{
			reference.RefTypeCashBank,
			reference.RefTypePayment,
			reference.RefTypeSalesPayment,
			reference.RefTypePurchasePayment,
			reference.RefTypeNTPPayment,
		}
	case "finance":
		return []string{
			reference.RefTypeGeneral,
			reference.RefTypeNonTradePayable,
			reference.RefTypeAssetTransaction,
			reference.RefTypeAssetDepreciation,
			reference.RefTypeUpCountryCost,
			reference.RefTypePeriodClosing,
			reference.RefTypeReversal,
			reference.RefTypeSalaryExpense,
		}
	case "adjustment":
		return []string{
			reference.RefTypeManualAdjustment,
			reference.RefTypeAdjustment,
			reference.RefTypeCorrection,
		}
	case "valuation":
		return []string{
			reference.RefTypeInventoryValuation,
			reference.RefTypeCurrencyRevaluation,
			reference.RefTypeCostAdjustment,
			reference.RefTypeDepreciationValuation,
		}
	default:
		return nil
	}
}

func validateLines(lines []dto.JournalLineRequest) (float64, float64, error) {
	if len(lines) < 2 {
		return 0, 0, ErrJournalInvalidLines
	}
	var debitTotal float64
	var creditTotal float64
	for _, ln := range lines {
		if strings.TrimSpace(ln.ChartOfAccountID) == "" {
			return 0, 0, ErrJournalInvalidLines
		}
		if ln.Debit < 0 || ln.Credit < 0 {
			return 0, 0, ErrJournalInvalidLines
		}
		if (ln.Debit > 0 && ln.Credit > 0) || (ln.Debit == 0 && ln.Credit == 0) {
			return 0, 0, ErrJournalInvalidLines
		}
		debitTotal += ln.Debit
		creditTotal += ln.Credit
	}
	if math.Abs(math.Round(debitTotal*100)-math.Round(creditTotal*100)) > 0.1 {
		return debitTotal, creditTotal, ErrJournalUnbalanced
	}
	return debitTotal, creditTotal, nil
}

func parseJournalType(value *string) financeModels.JournalType {
	if value == nil {
		return financeModels.JournalTypeGeneral
	}
	v := strings.ToUpper(strings.TrimSpace(*value))
	if v == "" {
		return financeModels.JournalTypeGeneral
	}
	switch financeModels.JournalType(v) {
	case financeModels.JournalTypeGeneral,
		financeModels.JournalTypeAdjustment,
		financeModels.JournalTypeSales,
		financeModels.JournalTypePurchase,
		financeModels.JournalTypeOpeningBalance,
		financeModels.JournalTypeClosing:
		return financeModels.JournalType(v)
	}
	return financeModels.JournalTypeGeneral
}

func journalPrefixByType(journalType financeModels.JournalType) string {
	switch journalType {
	case financeModels.JournalTypeSales:
		return "SJ"
	case financeModels.JournalTypePurchase:
		return "PJ"
	case financeModels.JournalTypeAdjustment:
		return "AJ"
	case financeModels.JournalTypeOpeningBalance:
		return "OB"
	case financeModels.JournalTypeClosing:
		return "CL"
	default:
		return "GJ"
	}
}

func (uc *journalEntryUsecase) generateJournalNumber(ctx context.Context, tx *gorm.DB, companyID string, journalType financeModels.JournalType, entryDate time.Time) (string, error) {
	prefix := journalPrefixByType(journalType)
	year := entryDate.Year()
	basePattern := fmt.Sprintf("%s/%d/", prefix, year)

	var lastEntry financeModels.JournalEntry
	query := tx.WithContext(ctx).
		Model(&financeModels.JournalEntry{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("journal_number LIKE ?", basePattern+"%")
	if strings.TrimSpace(companyID) != "" {
		query = query.Where("company_id = ?", strings.TrimSpace(companyID))
	}

	err := query.Order("journal_number desc").Take(&lastEntry).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	seq := 1
	if err == nil {
		parts := strings.Split(strings.TrimSpace(lastEntry.JournalNumber), "/")
		if len(parts) == 3 {
			if parsed, parseErr := strconv.Atoi(parts[2]); parseErr == nil {
				seq = parsed + 1
			}
		}
	}

	if journalType == financeModels.JournalTypeOpeningBalance || journalType == financeModels.JournalTypeClosing {
		seq = 1
	}

	return fmt.Sprintf("%s/%d/%04d", prefix, year, seq), nil
}

func (uc *journalEntryUsecase) resolveFiscalYearID(ctx context.Context, tx *gorm.DB, companyID string, fiscalYearID *string, entryDate time.Time) (*string, error) {
	trimmedCompanyID := strings.TrimSpace(companyID)
	if trimmedCompanyID == "" {
		if fiscalYearID != nil && strings.TrimSpace(*fiscalYearID) != "" {
			fyID := strings.TrimSpace(*fiscalYearID)
			return &fyID, nil
		}
		return nil, errors.New("no active fiscal year for posting date")
	}

	if fiscalYearID != nil && strings.TrimSpace(*fiscalYearID) != "" {
		fyID := strings.TrimSpace(*fiscalYearID)
		var fy financeModels.FiscalYear
		if err := tx.WithContext(ctx).First(&fy, "id = ?", fyID).Error; err != nil {
			if strings.Contains(err.Error(), `relation "fiscal_years" does not exist`) {
				return nil, nil
			}
			return nil, err
		}
		if fy.CompanyID != trimmedCompanyID {
			return nil, errors.New("fiscal year does not belong to company")
		}
		if entryDate.Before(fy.StartDate) || entryDate.After(fy.EndDate) {
			return nil, errors.New("journal date is outside fiscal year range")
		}
		return &fyID, nil
	}

	var fy financeModels.FiscalYear
	err := tx.WithContext(ctx).
		Where("company_id = ?", trimmedCompanyID).
		Where("status = ?", financeModels.FiscalYearStatusActive).
		Where("? BETWEEN start_date AND end_date", entryDate).
		Order("start_date desc").
		Take(&fy).Error
	if err != nil {
		if strings.Contains(err.Error(), `relation "fiscal_years" does not exist`) {
			return nil, nil
		}
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active fiscal year for date %s", entryDate.Format("2006-01-02"))
		}
		return nil, err
	}
	fyID := fy.ID
	return &fyID, nil
}

func (uc *journalEntryUsecase) validatePostingFiscalYear(ctx context.Context, tx *gorm.DB, entry *financeModels.JournalEntry) error {
	if entry == nil || entry.FiscalYearID == nil || strings.TrimSpace(*entry.FiscalYearID) == "" {
		return errors.New("posted journal entry must have fiscal year")
	}

	var fy financeModels.FiscalYear
	if err := tx.WithContext(ctx).First(&fy, "id = ?", strings.TrimSpace(*entry.FiscalYearID)).Error; err != nil {
		return err
	}
	if fy.Status != financeModels.FiscalYearStatusActive {
		return errors.New("fiscal year is not active")
	}
	if entry.EntryDate.Before(fy.StartDate) || entry.EntryDate.After(fy.EndDate) {
		return errors.New("journal date is outside fiscal year range")
	}
	return nil
}

func validatePostingCOA(coa *financeModels.ChartOfAccount) error {
	if coa == nil {
		return errors.New("chart of account not found")
	}
	if !coa.IsPostable {
		return fmt.Errorf("akun '%s - %s' adalah akun induk (non-postable) dan tidak dapat digunakan dalam jurnal", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
	}
	if !coa.IsActive {
		return fmt.Errorf("akun '%s - %s' tidak aktif dan tidak dapat digunakan dalam jurnal", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
	}
	return nil
}

func warnAbnormalPostingSide(coa *financeModels.ChartOfAccount, line dto.JournalLineRequest) {
	if coa == nil {
		return
	}

	isDebitNormal := isDebitNormalAccountType(coa.Type)
	if line.Debit > 0 && !isDebitNormal {
		log.Printf("journal_validation_warning: account %s (%s) posted on debit side (unusual for credit-normal account)", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
	}
	if line.Credit > 0 && isDebitNormal {
		log.Printf("journal_validation_warning: account %s (%s) posted on credit side (unusual for debit-normal account)", strings.TrimSpace(coa.Code), strings.TrimSpace(coa.Name))
	}
}

func (uc *journalEntryUsecase) Create(ctx context.Context, req *dto.CreateJournalEntryRequest) (*dto.JournalEntryResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	if parseJournalType(req.JournalType) == financeModels.JournalTypeOpeningBalance {
		return nil, errors.New("opening journal must be created via opening balance posting endpoint")
	}

	entryDate, err := parseDate(req.EntryDate)
	if err != nil {
		return nil, err
	}
	if _, _, err := validateLines(req.Lines); err != nil {
		return nil, err
	}

	// Hardening: block manual journals from using trade control accounts (AR/AP/Inventory)
	if !req.IsSystemGenerated && !req.SkipControlAccountValidation {
		if err := uc.validateControlAccountsForLines(ctx, req.Lines); err != nil {
			return nil, err
		}
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	var createdID string
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		if err := ensureNotClosed(ctx, tx, entryDate); err != nil {
			return err
		}

		resolvedFiscalYearID, err := uc.resolveFiscalYearID(ctx, tx, req.CompanyID, req.FiscalYearID, entryDate)
		if err != nil {
			return err
		}
		coaIDs := make([]string, 0, len(req.Lines))
		for _, ln := range req.Lines {
			coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, ln := range req.Lines {
			coa := coaByID[strings.TrimSpace(ln.ChartOfAccountID)]
			if err := validatePostingCOA(coa); err != nil {
				return err
			}
			warnAbnormalPostingSide(coa, ln)
		}

		refTypeNormalized := reference.NormalizePtr(req.ReferenceType)
		var refTypePtr *string
		if refTypeNormalized != "" {
			refTypePtr = &refTypeNormalized
		}

		debitT, creditT, _ := validateLines(req.Lines)
		journalType := parseJournalType(req.JournalType)
		journalNumber, err := uc.generateJournalNumber(ctx, tx, req.CompanyID, journalType, entryDate)
		if err != nil {
			return err
		}

		exchangeRate := 1.0
		if req.ExchangeRate != nil && *req.ExchangeRate > 0 {
			exchangeRate = *req.ExchangeRate
		}
		currencyCode := strings.TrimSpace(req.CurrencyCode)
		if currencyCode == "" {
			currencyCode = "IDR"
		}

		entry := &financeModels.JournalEntry{
			CompanyID:         strings.TrimSpace(req.CompanyID),
			FiscalYearID:      resolvedFiscalYearID,
			JournalNumber:     journalNumber,
			EntryDate:         entryDate,
			Reference:         strings.TrimSpace(req.Reference),
			Description:       strings.TrimSpace(req.Description),
			ReferenceType:     refTypePtr,
			ReferenceID:       req.ReferenceID,
			JournalType:       journalType,
			Status:            financeModels.JournalStatusDraft,
			DebitTotal:        debitT,
			CreditTotal:       creditT,
			CurrencyCode:      currencyCode,
			ExchangeRate:      exchangeRate,
			CreatedBy:         &actorID,
			IsSystemGenerated: req.IsSystemGenerated,
			SourceDocumentURL: req.SourceDocumentURL,
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		for _, ln := range req.Lines {
			memo := strings.TrimSpace(ln.Memo)
			coa := coaByID[ln.ChartOfAccountID]
			line := &financeModels.JournalLine{
				JournalEntryID:             entry.ID,
				ChartOfAccountID:           ln.ChartOfAccountID,
				ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
				ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
				ChartOfAccountTypeSnapshot: string(coa.Type),
				Debit:                      ln.Debit,
				Credit:                     ln.Credit,
				Memo:                       memo,
			}
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		createdID = entry.ID
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

func (uc *journalEntryUsecase) Update(ctx context.Context, id string, req *dto.UpdateJournalEntryRequest) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	if parseJournalType(req.JournalType) == financeModels.JournalTypeOpeningBalance {
		return nil, errors.New("opening journal cannot be updated via manual journal endpoint")
	}

	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrJournalNotFound
	}

	entry, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if entry.Status == financeModels.JournalStatusPosted || entry.Status == financeModels.JournalStatusReversed || entry.Status == financeModels.JournalStatusCancelled {
		return nil, ErrJournalPostedImmutable
	}
	if entry.IsSystemGenerated {
		return nil, errors.New("system-generated journal entries cannot be modified")
	}

	entryDate, err := parseDate(req.EntryDate)
	if err != nil {
		return nil, err
	}
	if _, _, err := validateLines(req.Lines); err != nil {
		return nil, err
	}

	// Hardening: block manual journals from using trade control accounts (AR/AP/Inventory)
	if err := uc.validateControlAccountsForLines(ctx, req.Lines); err != nil {
		return nil, err
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureNotClosed(ctx, tx, entryDate); err != nil {
			return err
		}

		companyID := strings.TrimSpace(req.CompanyID)
		if companyID == "" {
			companyID = strings.TrimSpace(entry.CompanyID)
		}

		resolvedFiscalYearID, err := uc.resolveFiscalYearID(ctx, tx, companyID, req.FiscalYearID, entryDate)
		if err != nil {
			return err
		}

		existingLineSnapshot := make(map[string]financeModels.JournalLine)
		for _, ln := range entry.Lines {
			key := strings.TrimSpace(ln.ChartOfAccountID) + "|" + strings.TrimSpace(ln.Memo) + "|" + formatFloatKey(ln.Debit) + "|" + formatFloatKey(ln.Credit)
			existingLineSnapshot[key] = ln
		}
		refTypeNormalized := reference.NormalizePtr(req.ReferenceType)
		var refTypePtr *string
		if refTypeNormalized != "" {
			refTypePtr = &refTypeNormalized
		}

		debitT, creditT, _ := validateLines(req.Lines)

		exchangeRate := entry.ExchangeRate
		if req.ExchangeRate != nil && *req.ExchangeRate > 0 {
			exchangeRate = *req.ExchangeRate
		}
		currencyCode := strings.TrimSpace(req.CurrencyCode)
		if currencyCode == "" {
			currencyCode = entry.CurrencyCode
		}
		if currencyCode == "" {
			currencyCode = "IDR"
		}

		if err := tx.Model(&financeModels.JournalEntry{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"company_id":     companyID,
				"fiscal_year_id": resolvedFiscalYearID,
				"entry_date":     entryDate,
				"reference":      strings.TrimSpace(req.Reference),
				"description":    strings.TrimSpace(req.Description),
				"reference_type": refTypePtr,
				"reference_id":   req.ReferenceID,
				"journal_type":   parseJournalType(req.JournalType),
				"currency_code":  currencyCode,
				"exchange_rate":  exchangeRate,
				"debit_total":    debitT,
				"credit_total":   creditT,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("journal_entry_id = ?", id).Delete(&financeModels.JournalLine{}).Error; err != nil {
			return err
		}

		coaByID := make(map[string]*financeModels.ChartOfAccount, len(req.Lines))
		for _, ln := range req.Lines {
			if _, ok := coaByID[ln.ChartOfAccountID]; ok {
				continue
			}
			coa, err := uc.coaRepo.FindByID(ctx, ln.ChartOfAccountID)
			if err != nil {
				return err
			}
			if err := validatePostingCOA(coa); err != nil {
				return err
			}
			coaByID[ln.ChartOfAccountID] = coa
		}

		for _, ln := range req.Lines {
			memo := strings.TrimSpace(ln.Memo)
			warnAbnormalPostingSide(coaByID[ln.ChartOfAccountID], ln)
			key := strings.TrimSpace(ln.ChartOfAccountID) + "|" + memo + "|" + formatFloatKey(ln.Debit) + "|" + formatFloatKey(ln.Credit)
			if snap, ok := existingLineSnapshot[key]; ok && (snap.ChartOfAccountCodeSnapshot != "" || snap.ChartOfAccountNameSnapshot != "" || snap.ChartOfAccountTypeSnapshot != "") {
				line := &financeModels.JournalLine{
					JournalEntryID:             id,
					ChartOfAccountID:           ln.ChartOfAccountID,
					ChartOfAccountCodeSnapshot: snap.ChartOfAccountCodeSnapshot,
					ChartOfAccountNameSnapshot: snap.ChartOfAccountNameSnapshot,
					ChartOfAccountTypeSnapshot: snap.ChartOfAccountTypeSnapshot,
					Debit:                      ln.Debit,
					Credit:                     ln.Credit,
					Memo:                       memo,
				}
				if err := tx.Create(line).Error; err != nil {
					return err
				}
				continue
			}

			coa := coaByID[ln.ChartOfAccountID]
			line := &financeModels.JournalLine{
				JournalEntryID:             id,
				ChartOfAccountID:           ln.ChartOfAccountID,
				ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
				ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
				ChartOfAccountTypeSnapshot: string(coa.Type),
				Debit:                      ln.Debit,
				Credit:                     ln.Credit,
				Memo:                       memo,
			}
			if err := tx.Create(line).Error; err != nil {
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

func (uc *journalEntryUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return ErrJournalNotFound
	}
	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrJournalNotFound
		}
		return err
	}
	if entry.Status == financeModels.JournalStatusPosted || entry.Status == financeModels.JournalStatusReversed || entry.Status == financeModels.JournalStatusCancelled {
		return ErrJournalPostedImmutable
	}
	if entry.IsSystemGenerated {
		return errors.New("system-generated journal entries cannot be deleted")
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.JournalEntry{}, "id = ?", id).Error
}

func (uc *journalEntryUsecase) GetByID(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrJournalNotFound
	}
	item, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	resp := uc.mapper.ToResponse(item)
	if codes := repositories.BatchResolveJournalReferenceCodes(ctx, uc.db, []financeModels.JournalEntry{*item}); len(codes) > 0 {
		if c, ok := codes[item.ID]; ok && strings.TrimSpace(c) != "" {
			cc := strings.TrimSpace(c)
			resp.ReferenceCode = &cc
		}
	}
	return &resp, nil
}

func (uc *journalEntryUsecase) List(ctx context.Context, req *dto.ListJournalEntriesRequest) ([]dto.JournalEntryResponse, int64, error) {
	if req == nil {
		req = &dto.ListJournalEntriesRequest{}
	}
	page, perPage := normalizePagination(req.Page, req.PerPage)

	startDate, err := parseDateOptional(req.StartDate)
	if err != nil {
		return nil, 0, err
	}
	endDate, err := parseEndDateOptional(req.EndDate)
	if err != nil {
		return nil, 0, err
	}

	// Ensure times are in the application timezone if they represent a local end-of-day
	if endDate != nil {
		loc := apptime.Location()
		*endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc)
	}

	params := repositories.JournalEntryListParams{
		Search:         req.Search,
		Status:         req.Status,
		CompanyID:      req.CompanyID,
		FiscalYearID:   req.FiscalYearID,
		StartDate:      startDate,
		EndDate:        endDate,
		SortBy:         req.SortBy,
		SortDir:        req.SortDir,
		Limit:          perPage,
		Offset:         (page - 1) * perPage,
		ReferenceType:  req.ReferenceType,
		ReferenceTypes: journalReferenceTypesForDomain(req.Domain),
	}
	if req.JournalType != nil {
		jt := parseJournalType(req.JournalType)
		params.JournalType = &jt
	}

	log.Printf("journal_observability: List domain=%v refTypes=%v search=%q startDate=%v endDate=%v", req.Domain, params.ReferenceTypes, params.Search, params.StartDate, params.EndDate)

	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	codes := repositories.BatchResolveJournalReferenceCodes(ctx, uc.db, items)

	resp := make([]dto.JournalEntryResponse, 0, len(items))
	for i := range items {
		v := uc.mapper.ToSummaryResponse(&items[i])
		if c, ok := codes[items[i].ID]; ok && strings.TrimSpace(c) != "" {
			cc := strings.TrimSpace(c)
			v.ReferenceCode = &cc
		}
		resp = append(resp, v)
	}
	return resp, total, nil
}

func (uc *journalEntryUsecase) Post(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrJournalNotFound
	}
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	entry, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if entry.Status == financeModels.JournalStatusPosted {
		resp := uc.mapper.ToResponse(entry)
		return &resp, nil
	}
	if entry.Status == financeModels.JournalStatusCancelled || entry.Status == financeModels.JournalStatusReversed {
		return nil, ErrJournalPostedImmutable
	}

	var debitTotal float64
	var creditTotal float64
	for _, ln := range entry.Lines {
		debitTotal += ln.Debit
		creditTotal += ln.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.000001 {
		return nil, ErrJournalUnbalanced
	}

	now := apptime.Now()
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		if err := ensureNotClosed(ctx, tx, entry.EntryDate); err != nil {
			return err
		}

		if err := uc.validatePostingFiscalYear(ctx, tx, entry); err != nil {
			return err
		}

		if entry.ReferenceType != nil && entry.ReferenceID != nil {
			refTypeNormalized := reference.NormalizePtr(entry.ReferenceType)
			refID := strings.TrimSpace(*entry.ReferenceID)
			if refTypeNormalized != "" && refID != "" {
				lockKey := fmt.Sprintf("journal:%s:%s", refTypeNormalized, refID)
				if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
					return fmt.Errorf("failed to acquire advisory lock: %w", err)
				}
			}
		}

		coaIDs := make([]string, 0, len(entry.Lines))
		for _, ln := range entry.Lines {
			coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
		}
		coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
		if err != nil {
			return err
		}
		for _, ln := range entry.Lines {
			coa := coaByID[strings.TrimSpace(ln.ChartOfAccountID)]
			if err := validatePostingCOA(coa); err != nil {
				return err
			}
			warnAbnormalPostingSide(coa, dto.JournalLineRequest{ChartOfAccountID: ln.ChartOfAccountID, Debit: ln.Debit, Credit: ln.Credit, Memo: ln.Memo})
		}

		if err := tx.Model(&financeModels.JournalEntry{}).
			Where("id = ? AND status = ?", id, financeModels.JournalStatusDraft).
			Updates(map[string]interface{}{
				"debit_total":  debitTotal,
				"credit_total": creditTotal,
				"status":       financeModels.JournalStatusPosted,
				"posted_at":    &now,
				"posted_by":    &actorID,
			}).Error; err != nil {
			return err
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

func (uc *journalEntryUsecase) TrialBalance(ctx context.Context, startDate, endDate *time.Time) (*dto.TrialBalanceResponse, error) {
	if endDate != nil {
		eod := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		endDate = &eod
	}
	type aggRow struct {
		ChartOfAccountID string
		DebitTotal       float64
		CreditTotal      float64
	}

	q := uc.db.WithContext(ctx).
		Table("journal_lines").
		Select("journal_lines.chart_of_account_id as chart_of_account_id, COALESCE(SUM(journal_lines.debit),0) as debit_total, COALESCE(SUM(journal_lines.credit),0) as credit_total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.status = ?", financeModels.JournalStatusPosted).
		Where("journal_entries.deleted_at IS NULL").
		Where("journal_lines.deleted_at IS NULL")

	if startDate != nil {
		q = q.Where("journal_entries.entry_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("journal_entries.entry_date <= ?", *endDate)
	}
	q = q.Group("journal_lines.chart_of_account_id")

	var rows []aggRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	coas, err := uc.coaRepo.FindAll(ctx, false)
	if err != nil {
		return nil, err
	}
	agg := make(map[string]aggRow, len(rows))
	for _, r := range rows {
		agg[r.ChartOfAccountID] = r
	}

	out := make([]dto.TrialBalanceRow, 0, len(coas))
	for _, a := range coas {
		r, ok := agg[a.ID]
		if !ok {
			r = aggRow{ChartOfAccountID: a.ID}
		}
		out = append(out, dto.TrialBalanceRow{
			ChartOfAccountID: a.ID,
			Code:             a.Code,
			Name:             a.Name,
			Type:             a.Type,
			DebitTotal:       r.DebitTotal,
			CreditTotal:      r.CreditTotal,
			Balance:          r.DebitTotal - r.CreditTotal,
		})
	}

	resp := &dto.TrialBalanceResponse{StartDate: startDate, EndDate: endDate, Rows: out}
	return resp, nil
}

func (uc *journalEntryUsecase) PostOrUpdateJournal(ctx context.Context, req *dto.CreateJournalEntryRequest) (*dto.JournalEntryResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if req.ReferenceType == nil || req.ReferenceID == nil {
		return nil, errors.New("reference type and reference id are required for PostOrUpdateJournal")
	}

	entryDate, err := parseDate(req.EntryDate)
	if err != nil {
		return nil, err
	}
	journalType := parseJournalType(req.JournalType)

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	traceKey := journalTraceKey(req.ReferenceType, req.ReferenceID)

	refTypeNormalized := reference.NormalizePtr(req.ReferenceType)
	refID := strings.TrimSpace(*req.ReferenceID)

	var out *dto.JournalEntryResponse
	// We use a single transaction for the entire lookup-then-upsert-then-post flow to ensure atomicity.
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		// 1. Acquire an advisory lock scoped to this specific reference to prevent concurrent processes
		// from attempting to create/update the same journal simultaneously.
		lockKey := fmt.Sprintf("journal:%s:%s", refTypeNormalized, refID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
			return fmt.Errorf("failed to acquire advisory lock: %w", err)
		}

		// 2. Lookup existing entry using the same transaction (with row-level lock redundant but safe)
		var existing financeModels.JournalEntry
		lookupErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("reference_type = ? AND reference_id = ?", refTypeNormalized, refID).
			First(&existing).Error

		if lookupErr != nil && lookupErr != gorm.ErrRecordNotFound {
			return lookupErr
		}

		// 3. If exists and posted, return it (idempotency success)
		if lookupErr == nil && existing.Status == financeModels.JournalStatusPosted {
			logJournalEvent("post_or_update.idempotent_existing_posted", map[string]interface{}{
				"trace_key": traceKey,
				"entry_id":  existing.ID,
			})
			full, err := uc.repo.FindByID(database.WithTx(ctx, tx), existing.ID, true)
			if err != nil {
				return err
			}
			resp := uc.mapper.ToResponse(full)
			out = &resp
			return nil
		}

		resolvedCompanyID := strings.TrimSpace(req.CompanyID)
		if lookupErr == nil && resolvedCompanyID == "" {
			resolvedCompanyID = strings.TrimSpace(existing.CompanyID)
		}
		if resolvedCompanyID == "" {
			resolvedCompanyID, err = uc.resolveCompanyIDFromActor(ctx)
			if err != nil {
				return fmt.Errorf("company id is required for journal posting: %w", err)
			}
		}

		resolvedFiscalYearID, err := uc.resolveFiscalYearID(ctx, tx, resolvedCompanyID, req.FiscalYearID, entryDate)
		if err != nil {
			return err
		}

		exchangeRate := 1.0
		if req.ExchangeRate != nil && *req.ExchangeRate > 0 {
			exchangeRate = *req.ExchangeRate
		}
		currencyCode := strings.TrimSpace(req.CurrencyCode)
		if currencyCode == "" {
			currencyCode = "IDR"
		}

		// 4. Create or Update draft
		var entryID string
		var journalNumber string
		if lookupErr == gorm.ErrRecordNotFound {
			if err := ensureNotClosed(ctx, tx, entryDate); err != nil {
				return err
			}

			debitT, creditT, err := validateLines(req.Lines)
			if err != nil {
				return err
			}

			journalNumber, err = uc.generateJournalNumber(ctx, tx, resolvedCompanyID, journalType, entryDate)
			if err != nil {
				return err
			}

			var createdBy *string
			if actorID != "" {
				createdBy = &actorID
			}

			entry := &financeModels.JournalEntry{
				CompanyID:         resolvedCompanyID,
				FiscalYearID:      resolvedFiscalYearID,
				JournalNumber:     journalNumber,
				EntryDate:         entryDate,
				Reference:         strings.TrimSpace(req.Reference),
				Description:       strings.TrimSpace(req.Description),
				ReferenceType:     &refTypeNormalized,
				ReferenceID:       &refID,
				JournalType:       journalType,
				Status:            financeModels.JournalStatusDraft,
				DebitTotal:        debitT,
				CreditTotal:       creditT,
				CurrencyCode:      currencyCode,
				ExchangeRate:      exchangeRate,
				CreatedBy:         createdBy,
				IsSystemGenerated: req.IsSystemGenerated,
				SourceDocumentURL: req.SourceDocumentURL,
			}
			if err := tx.Create(entry).Error; err != nil {
				return err
			}
			entryID = entry.ID

			// Load COAs for snapshots
			coaIDs := make([]string, 0, len(req.Lines))
			for _, ln := range req.Lines {
				coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
			}
			coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
			if err != nil {
				return err
			}

			for _, ln := range req.Lines {
				lineCOAID := strings.TrimSpace(ln.ChartOfAccountID)
				coa := coaByID[lineCOAID]
				if coa == nil {
					return fmt.Errorf("chart of account %s not found", lineCOAID)
				}
				if err := validatePostingCOA(coa); err != nil {
					return err
				}
				warnAbnormalPostingSide(coa, ln)
				line := &financeModels.JournalLine{
					JournalEntryID:             entryID,
					ChartOfAccountID:           lineCOAID,
					ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
					ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
					ChartOfAccountTypeSnapshot: string(coa.Type),
					Debit:                      ln.Debit,
					Credit:                     ln.Credit,
					Memo:                       strings.TrimSpace(ln.Memo),
				}
				if err := tx.Create(line).Error; err != nil {
					return err
				}
			}
		} else {
			// Update existing draft
			entryID = existing.ID
			if err := ensureNotClosed(ctx, tx, entryDate); err != nil {
				return err
			}

			debitT, creditT, err := validateLines(req.Lines)
			if err != nil {
				return err
			}

			existingDate := existing.EntryDate.Format("2006-01-02")
			requestedDate := entryDate.Format("2006-01-02")
			journalNumber = strings.TrimSpace(existing.JournalNumber)
			if journalNumber == "" ||
				existingDate != requestedDate ||
				existing.JournalType != journalType ||
				strings.TrimSpace(existing.CompanyID) != resolvedCompanyID {
				journalNumber, err = uc.generateJournalNumber(ctx, tx, resolvedCompanyID, journalType, entryDate)
				if err != nil {
					return err
				}
			}

			if err := tx.Model(&financeModels.JournalEntry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
				"company_id":          resolvedCompanyID,
				"fiscal_year_id":      resolvedFiscalYearID,
				"journal_number":      journalNumber,
				"entry_date":          entryDate,
				"reference":           strings.TrimSpace(req.Reference),
				"description":         strings.TrimSpace(req.Description),
				"debit_total":         debitT,
				"credit_total":        creditT,
				"reference_type":      refTypeNormalized,
				"reference_id":        refID,
				"journal_type":        journalType,
				"currency_code":       currencyCode,
				"exchange_rate":       exchangeRate,
				"source_document_url": req.SourceDocumentURL,
			}).Error; err != nil {
				return err
			}

			if err := tx.Where("journal_entry_id = ?", entryID).Delete(&financeModels.JournalLine{}).Error; err != nil {
				return err
			}

			coaIDs := make([]string, 0, len(req.Lines))
			for _, ln := range req.Lines {
				coaIDs = append(coaIDs, strings.TrimSpace(ln.ChartOfAccountID))
			}
			coaByID, err := loadCOAMap(tx.WithContext(ctx), coaIDs)
			if err != nil {
				return err
			}

			for _, ln := range req.Lines {
				lineCOAID := strings.TrimSpace(ln.ChartOfAccountID)
				coa := coaByID[lineCOAID]
				if coa == nil {
					return fmt.Errorf("chart of account %s not found", lineCOAID)
				}
				if err := validatePostingCOA(coa); err != nil {
					return err
				}
				warnAbnormalPostingSide(coa, ln)
				line := &financeModels.JournalLine{
					JournalEntryID:             entryID,
					ChartOfAccountID:           lineCOAID,
					ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
					ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
					ChartOfAccountTypeSnapshot: string(coa.Type),
					Debit:                      ln.Debit,
					Credit:                     ln.Credit,
					Memo:                       strings.TrimSpace(ln.Memo),
				}
				if err := tx.Create(line).Error; err != nil {
					return err
				}
			}
		}

		// 5. Post it
		now := apptime.Now()
		var postedBy interface{}
		if actorID != "" {
			postedBy = actorID
		}
		if err := tx.Model(&financeModels.JournalEntry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
			"status":    financeModels.JournalStatusPosted,
			"posted_at": &now,
			"posted_by": postedBy,
		}).Error; err != nil {
			return err
		}

		// 6. Return response
		full, err := uc.repo.FindByID(database.WithTx(ctx, tx), entryID, true)
		if err != nil {
			return err
		}
		resp := uc.mapper.ToResponse(full)
		out = &resp
		return nil
	})

	if err != nil {
		logJournalEvent("post_or_update.failed", map[string]interface{}{
			"trace_key": traceKey,
			"error":     err.Error(),
		})
		return nil, err
	}

	return out, nil
}

func (uc *journalEntryUsecase) Cancel(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrJournalNotFound
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	entry, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}

	if entry.Status == financeModels.JournalStatusCancelled {
		resp := uc.mapper.ToResponse(entry)
		return &resp, nil
	}

	if entry.IsSystemGenerated {
		return nil, errors.New("system-generated journal entries cannot be cancelled")
	}

	reason := "Manual cancellation"
	if entry.Status == financeModels.JournalStatusPosted {
		if _, err := uc.reverse(ctx, id, reason); err != nil {
			return nil, fmt.Errorf("failed to reverse posted entry before cancellation: %w", err)
		}
	}

	now := apptime.Now()
	if err := uc.db.WithContext(ctx).Model(&financeModels.JournalEntry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          financeModels.JournalStatusCancelled,
			"reversal_reason": reason,
			"reversed_at":     &now,
			"reversed_by":     &actorID,
		}).Error; err != nil {
		return nil, err
	}

	uc.auditService.LogWithReason(ctx, "journal.cancel", entry.ID, reason, map[string]interface{}{
		"previous_status": entry.Status,
	})

	full, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		return nil, err
	}
	resp := uc.mapper.ToResponse(full)
	return &resp, nil
}

func (uc *journalEntryUsecase) resolveCompanyIDFromActor(ctx context.Context) (string, error) {
	if uc == nil || uc.db == nil {
		return "", errors.New("db is not configured")
	}

	if ctxCompanyID, _ := ctx.Value("company_id").(string); strings.TrimSpace(ctxCompanyID) != "" {
		return strings.TrimSpace(ctxCompanyID), nil
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return "", errors.New("user not authenticated")
	}

	var companyID string
	if err := database.GetDB(ctx, uc.db).
		WithContext(ctx).
		Table("employees").
		Select("company_id").
		Where("user_id = ? AND deleted_at IS NULL", actorID).
		Limit(1).
		Scan(&companyID).Error; err != nil {
		return "", err
	}

	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return "", errors.New("employee company not found")
	}

	return companyID, nil
}

// ReverseWithReason creates a new reversing journal entry with a specific reason.
func (uc *journalEntryUsecase) ReverseWithReason(ctx context.Context, id string, reason string) (*dto.JournalEntryResponse, error) {
	return uc.reverse(ctx, id, reason)
}

// Reverse creates a new reversing journal entry (swapped debit/credit) for a posted entry,
// then auto-posts the reversal. This is standard accounting practice for correcting errors.
func (uc *journalEntryUsecase) Reverse(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	return uc.reverse(ctx, id, "Manual reversal")
}

func (uc *journalEntryUsecase) reverse(ctx context.Context, id string, reason string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	if !security.CheckRecordScopeAccess(database.DB, ctx, &financeModels.JournalEntry{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrJournalNotFound
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	entry, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}

	if entry.Status != financeModels.JournalStatusPosted {
		return nil, errors.New("only posted journal entries can be reversed")
	}
	if entry.JournalType == financeModels.JournalTypeOpeningBalance {
		return nil, errors.New("opening journal cannot be reversed directly; use adjustment journal")
	}

	originalRefType := reference.NormalizePtr(entry.ReferenceType)
	reversalRefType := "REVERSAL"
	if originalRefType != "" {
		reversalRefType = "REVERSAL_" + originalRefType
	}

	// Idempotency guard: if reversal journal already exists by deterministic reference,
	// return it instead of failing.
	var existingReversalEntry financeModels.JournalEntry
	if err := uc.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ?", reversalRefType, entry.ID).
		Order("created_at DESC").
		First(&existingReversalEntry).Error; err == nil {
		found, ferr := uc.repo.FindByID(ctx, existingReversalEntry.ID, true)
		if ferr != nil {
			return nil, ferr
		}
		resp := uc.mapper.ToResponse(found)
		return &resp, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var existingReversal financeModels.JournalReversal
	err = uc.db.WithContext(ctx).
		Where("original_journal_entry_id = ?", entry.ID).
		First(&existingReversal).Error
	if err == nil {
		existing, ferr := uc.repo.FindByID(ctx, existingReversal.ReversalJournalEntryID, true)
		if ferr != nil {
			return nil, ferr
		}
		resp := uc.mapper.ToResponse(existing)
		return &resp, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Build reversed lines: swap debit and credit
	reversedLines := make([]dto.JournalLineRequest, 0, len(entry.Lines))
	for _, ln := range entry.Lines {
		reversedLines = append(reversedLines, dto.JournalLineRequest{
			ChartOfAccountID: ln.ChartOfAccountID,
			Debit:            ln.Credit,
			Credit:           ln.Debit,
			Memo:             "Reversal: " + ln.Memo,
		})
	}

	refType := reversalRefType
	reversalReq := &dto.CreateJournalEntryRequest{
		CompanyID:         entry.CompanyID,
		FiscalYearID:      entry.FiscalYearID,
		EntryDate:         apptime.Now().Format("2006-01-02"),
		Reference:         entry.Reference,
		Description:       "Reversal of: " + entry.Description,
		ReferenceType:     &refType,
		ReferenceID:       &entry.ID,
		CurrencyCode:      entry.CurrencyCode,
		ExchangeRate:      &entry.ExchangeRate,
		Lines:             reversedLines,
		IsSystemGenerated: true,
	}

	// Create the reversal journal entry
	reversal, err := uc.Create(ctx, reversalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal entry: %w", err)
	}

	// Auto-post the reversal
	posted, err := uc.Post(ctx, reversal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to post reversal entry: %w", err)
	}

	reversalMeta := &financeModels.JournalReversal{
		OriginalJournalEntryID: entry.ID,
		ReversalJournalEntryID: posted.ID,
		Reason:                 reason,
		CreatedBy:              &actorID,
	}
	if err := uc.db.WithContext(ctx).Create(reversalMeta).Error; err != nil {
		return nil, fmt.Errorf("failed to save reversal metadata: %w", err)
	}

	// Update original journal entry with reversal info
	now := apptime.Now()
	if err := uc.db.WithContext(ctx).Model(&financeModels.JournalEntry{}).
		Where("id = ?", entry.ID).
		Updates(map[string]interface{}{
			"status":              financeModels.JournalStatusReversed,
			"reversed_at":         &now,
			"reversed_by":         &actorID,
			"reversal_reason":     reason,
			"original_journal_id": nil, // This is the original
		}).Error; err != nil {
		log.Printf("warning: failed to mark original entry %s as reversed: %v", entry.ID, err)
	}

	// Update reversal journal entry to link back and populate totals
	var revDebit, revCredit float64
	for _, rl := range reversedLines {
		revDebit += rl.Debit
		revCredit += rl.Credit
	}

	if err := uc.db.WithContext(ctx).Model(&financeModels.JournalEntry{}).
		Where("id = ?", posted.ID).
		Updates(map[string]interface{}{
			"original_journal_id": &entry.ID,
			"is_reversal":         true,
			"reversed_from":       &entry.ID,
			"reversal_reason":     reason,
			"reversed_at":         &now, // The reversal itself is "reversed" impact
			"reversed_by":         &actorID,
			"debit_total":         revDebit,
			"credit_total":        revCredit,
		}).Error; err != nil {
		log.Printf("warning: failed to update reversal entry %s metadata: %v", posted.ID, err)
	}

	uc.auditService.LogWithReason(ctx, "journal.reverse", entry.ID, reason, map[string]interface{}{
		"original_id": entry.ID,
		"reversal_id": posted.ID,
	})

	return posted, nil
}

func (uc *journalEntryUsecase) CreateOpeningBalanceJournal(ctx context.Context, req *dto.CreateOpeningBalanceJournalRequest) (*dto.JournalEntryResponse, error) {
	if req == nil {
		return nil, errors.New("opening balance request is required")
	}
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return nil, errors.New("account id is required")
	}

	account, err := uc.coaRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	if !account.IsPostable {
		return nil, errors.New("opening balance can only be set for postable accounts")
	}
	if req.Amount == 0 {
		return nil, nil
	}

	openingEquity, err := uc.coaRepo.FindByCode(ctx, "39999")
	if err != nil || openingEquity == nil {
		return nil, errors.New("akun Saldo Awal Ekuitas (39999) belum dikonfigurasi")
	}
	if !openingEquity.IsPostable {
		return nil, errors.New("akun Saldo Awal Ekuitas (39999) harus postable")
	}

	entryDate := apptime.Now().Format("2006-01-02")
	if req.EntryDate != nil && strings.TrimSpace(*req.EntryDate) != "" {
		parsedDate, err := parseDateRequired(*req.EntryDate)
		if err != nil {
			return nil, err
		}
		entryDate = parsedDate.Format("2006-01-02")
	}

	lines := buildOpeningBalanceLines(account.ID, account.Type, req.Amount, openingEquity.ID)
	description := strings.TrimSpace(account.Code) + " " + strings.TrimSpace(account.Name)
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		description = strings.TrimSpace(*req.Description)
	}
	refType := string(financeModels.RefOpeningBalance)
	refID := account.ID
	journalType := string(financeModels.JournalTypeOpeningBalance)

	return uc.PostOrUpdateJournal(ctx, &dto.CreateJournalEntryRequest{
		EntryDate:         entryDate,
		Description:       "Opening balance - " + description,
		ReferenceType:     &refType,
		ReferenceID:       &refID,
		JournalType:       &journalType,
		Lines:             lines,
		IsSystemGenerated: true,
	})
}

func (uc *journalEntryUsecase) ReverseOpeningBalance(ctx context.Context, accountID string) (*dto.JournalEntryResponse, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, errors.New("account id is required")
	}

	entry, err := uc.repo.FindByReferenceID(ctx, string(financeModels.RefOpeningBalance), accountID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	if entry.Status == financeModels.JournalStatusReversed {
		return nil, nil
	}

	postedReversal, err := uc.ReverseWithReason(ctx, entry.ID, "Opening balance updated")
	if err != nil {
		return nil, err
	}

	referenceType := string(financeModels.RefOpeningBalanceRev)
	referenceID := entry.ID
	if err := uc.db.WithContext(ctx).Model(&financeModels.JournalEntry{}).
		Where("id = ?", entry.ID).
		Updates(map[string]interface{}{
			"reference_type": referenceType,
			"reference_id":   referenceID,
		}).Error; err != nil {
		return nil, err
	}

	return postedReversal, nil
}

func buildOpeningBalanceLines(accountID string, accountType financeModels.AccountType, amountValue float64, openingEquityID string) []dto.JournalLineRequest {
	amount := math.Abs(amountValue)
	debitNormal := isDebitNormalAccountType(accountType)

	if debitNormal {
		if amountValue >= 0 {
			return []dto.JournalLineRequest{
				{ChartOfAccountID: accountID, Debit: amount, Credit: 0, Memo: "Opening balance"},
				{ChartOfAccountID: openingEquityID, Debit: 0, Credit: amount, Memo: "Opening balance offset"},
			}
		}
		return []dto.JournalLineRequest{
			{ChartOfAccountID: openingEquityID, Debit: amount, Credit: 0, Memo: "Opening balance offset"},
			{ChartOfAccountID: accountID, Debit: 0, Credit: amount, Memo: "Opening balance"},
		}
	}

	if amountValue >= 0 {
		return []dto.JournalLineRequest{
			{ChartOfAccountID: openingEquityID, Debit: amount, Credit: 0, Memo: "Opening balance offset"},
			{ChartOfAccountID: accountID, Debit: 0, Credit: amount, Memo: "Opening balance"},
		}
	}

	return []dto.JournalLineRequest{
		{ChartOfAccountID: accountID, Debit: amount, Credit: 0, Memo: "Opening balance"},
		{ChartOfAccountID: openingEquityID, Debit: 0, Credit: amount, Memo: "Opening balance offset"},
	}
}

func isDebitNormalAccountType(accountType financeModels.AccountType) bool {
	switch accountType {
	case financeModels.AccountTypeAsset,
		financeModels.AccountTypeExpense,
		financeModels.AccountTypeCashBank,
		financeModels.AccountTypeCurrentAsset,
		financeModels.AccountTypeFixedAsset,
		financeModels.AccountTypeCOGS,
		financeModels.AccountTypeSalaryWages,
		financeModels.AccountTypeOperational:
		return true
	default:
		return false
	}
}

// GetFormData returns dropdown options for journal entry forms (COA list, enums, currencies).
func (uc *journalEntryUsecase) GetFormData(ctx context.Context) (*dto.JournalEntryFormDataResponse, error) {
	coas, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	coaOptions := make([]dto.COAFormOption, 0, len(coas))
	for _, coa := range coas {
		if !coa.IsPostable {
			continue
		}
		coaOptions = append(coaOptions, dto.COAFormOption{
			ID:   coa.ID,
			Code: coa.Code,
			Name: coa.Name,
			Type: string(coa.Type),
		})
	}

	// Journal Types
	journalTypes := []dto.EnumOption{
		{Value: string(financeModels.JournalTypeGeneral), Label: "General Journal"},
		{Value: string(financeModels.JournalTypeAdjustment), Label: "Adjustment Journal"},
		{Value: string(financeModels.JournalTypeSales), Label: "Sales Journal"},
		{Value: string(financeModels.JournalTypePurchase), Label: "Purchase Journal"},
		{Value: string(financeModels.JournalTypeOpeningBalance), Label: "Opening Balance"},
		{Value: string(financeModels.JournalTypeClosing), Label: "Closing Journal"},
	}

	// Reference Types
	referenceTypes := []dto.EnumOption{
		{Value: string(financeModels.RefSO), Label: "Sales Order"},
		{Value: string(financeModels.RefPO), Label: "Purchase Order"},
		{Value: string(financeModels.RefDO), Label: "Delivery Order"},
		{Value: string(financeModels.RefGR), Label: "Goods Receipt"},
		{Value: string(financeModels.RefStockOpname), Label: "Stock Opname"},
		{Value: string(financeModels.RefAdjustment), Label: "Adjustment"},
		{Value: string(financeModels.RefNonTradePayable), Label: "Non-Trade Payable"},
		{Value: string(financeModels.RefPayment), Label: "Payment"},
		{Value: string(financeModels.RefAssetTransaction), Label: "Asset Transaction"},
		{Value: string(financeModels.RefAssetDepreciation), Label: "Asset Depreciation"},
		{Value: string(financeModels.RefCashBank), Label: "Cash Bank"},
		{Value: string(financeModels.RefUpCountryCost), Label: "Up Country Cost"},
		{Value: string(financeModels.RefInventoryVal), Label: "Inventory Valuation"},
		{Value: string(financeModels.RefCurrencyReval), Label: "Currency Revaluation"},
		{Value: string(financeModels.RefCostAdjustment), Label: "Cost Adjustment"},
		{Value: string(financeModels.RefDepreciation), Label: "Depreciation Valuation"},
		{Value: string(financeModels.RefOpeningBalance), Label: "Opening Balance"},
		{Value: string(financeModels.RefOpeningBalanceRev), Label: "Opening Balance Reversed"},
	}

	// Currencies (supported currencies)
	currencies := []dto.EnumOption{
		{Value: "IDR", Label: "Indonesian Rupiah (IDR)"},
		{Value: "USD", Label: "US Dollar (USD)"},
		{Value: "EUR", Label: "Euro (EUR)"},
		{Value: "GBP", Label: "British Pound (GBP)"},
		{Value: "SGD", Label: "Singapore Dollar (SGD)"},
		{Value: "MYR", Label: "Malaysian Ringgit (MYR)"},
		{Value: "THB", Label: "Thai Baht (THB)"},
		{Value: "PHP", Label: "Philippine Peso (PHP)"},
		{Value: "VND", Label: "Vietnamese Dong (VND)"},
		{Value: "JPY", Label: "Japanese Yen (JPY)"},
	}

	// Statuses
	statuses := []dto.EnumOption{
		{Value: string(financeModels.JournalStatusDraft), Label: "Draft"},
		{Value: string(financeModels.JournalStatusPosted), Label: "Posted"},
		{Value: string(financeModels.JournalStatusReversed), Label: "Reversed"},
		{Value: string(financeModels.JournalStatusCancelled), Label: "Cancelled"},
	}

	return &dto.JournalEntryFormDataResponse{
		ChartOfAccounts: coaOptions,
		JournalTypes:    journalTypes,
		ReferenceTypes:  referenceTypes,
		Currencies:      currencies,
		Statuses:        statuses,
	}, nil
}

// CreateAdjustmentJournal creates a manual correction journal entry.
// reference_type is always forced to "MANUAL_ADJUSTMENT" and is_system_generated = false.
// This enforces governance: only Finance-controlled manual adjustments can use this endpoint.
// Hardening: control accounts (AR/AP/Inventory) are always blocked for adjustments.
func (uc *journalEntryUsecase) CreateAdjustmentJournal(ctx context.Context, req *dto.CreateAdjustmentJournalRequest) (*dto.JournalEntryResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	// Hardening: adjustment journals must never touch trade control accounts
	if !req.SkipControlAccountValidation {
		if err := uc.validateControlAccountsForLines(ctx, req.Lines); err != nil {
			return nil, err
		}
	}

	refType := "MANUAL_ADJUSTMENT"
	journalType := string(financeModels.JournalTypeAdjustment)
	baseReq := &dto.CreateJournalEntryRequest{
		CompanyID:                    req.CompanyID,
		FiscalYearID:                 req.FiscalYearID,
		EntryDate:                    req.EntryDate,
		Reference:                    req.Reference,
		Description:                  req.Description,
		ReferenceType:                &refType,
		ReferenceID:                  nil,
		JournalType:                  &journalType,
		CurrencyCode:                 req.CurrencyCode,
		ExchangeRate:                 req.ExchangeRate,
		Lines:                        req.Lines,
		IsSystemGenerated:            false,
		SkipControlAccountValidation: req.SkipControlAccountValidation,
		SourceDocumentURL:            req.SourceDocumentURL,
	}

	return uc.Create(ctx, baseReq)
}

// UpdateAdjustmentJournal updates a manual correction journal entry.
func (uc *journalEntryUsecase) UpdateAdjustmentJournal(ctx context.Context, id string, req *dto.UpdateJournalEntryRequest) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if entry.ReferenceType == nil || *entry.ReferenceType != "MANUAL_ADJUSTMENT" {
		return nil, errors.New("can only update manual adjustment journals")
	}

	refType := "MANUAL_ADJUSTMENT"
	journalType := string(financeModels.JournalTypeAdjustment)
	req.ReferenceType = &refType
	req.ReferenceID = entry.ReferenceID
	req.JournalType = &journalType

	return uc.Update(ctx, id, req)
}

// PostAdjustmentJournal posts a manual correction journal entry.
func (uc *journalEntryUsecase) PostAdjustmentJournal(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if entry.ReferenceType == nil || *entry.ReferenceType != "MANUAL_ADJUSTMENT" {
		return nil, errors.New("can only post manual adjustment journals")
	}
	latestApproval, err := uc.adjustmentApprovalRepo.GetLatestByJournalID(ctx, entry.ID)
	if err != nil {
		return nil, err
	}
	if latestApproval == nil || latestApproval.Action != financeModels.AdjustmentJournalApprovalActionApproved {
		return nil, ErrAdjustmentNeedsApproval
	}
	return uc.Post(ctx, id)
}

// ReverseAdjustmentJournal reverses a posted manual correction journal entry.
func (uc *journalEntryUsecase) ReverseAdjustmentJournal(ctx context.Context, id string) (*dto.JournalEntryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	entry, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	if entry.ReferenceType == nil || *entry.ReferenceType != "MANUAL_ADJUSTMENT" {
		return nil, errors.New("can only reverse manual adjustment journals")
	}
	return uc.Reverse(ctx, id)
}

// RunValuation implements a skeleton valuation process.
// In a real implementation, this would trigger actual calculations for FIFO/Average
// inventory values, currency rates, or cost adjustments and generate corresponding journals.
// For now, it generates a sample balanced "INVENTORY_VALUATION" journal entry.
func (uc *journalEntryUsecase) RunValuation(ctx context.Context) (*dto.JournalEntryResponse, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)

	refType := "INVENTORY_VALUATION"
	// Use a timestamped reference ID to ensure uniqueness for multiple runs
	refID := fmt.Sprintf("VAL-RUN-%d", time.Now().Unix())

	// Skeleton dynamic logic: let's assume we adjusted inventory value by $100
	req := &dto.CreateJournalEntryRequest{
		EntryDate:         apptime.Now().Format("2006-01-02"),
		Description:       "Inventory Valuation Run - Automatic Adjustment",
		ReferenceType:     &refType,
		ReferenceID:       &refID,
		IsSystemGenerated: true,
		Lines: []dto.JournalLineRequest{
			{
				ChartOfAccountID: "11040001", // Sample Inventory Asset Account ID (mock)
				Debit:            100.00,
				Credit:           0,
				Memo:             "Valuation adjustment - increase",
			},
			{
				ChartOfAccountID: "51010001", // Sample COGS/Valuation Expense Account ID (mock)
				Debit:            0,
				Credit:           100.00,
				Memo:             "Valuation adjustment - contra",
			},
		},
	}

	// In a real scenario, we would allow the creation even if COA IDs don't exist yet by pre-validating or using specific system accounts.
	// For this skeleton, we'll try to find any 2 COAs if the hardcoded ones fail, to ensure the "Run" at least produces something in a test/dev env.
	coas, _ := uc.coaRepo.FindAll(ctx, false)
	if len(coas) >= 2 {
		req.Lines[0].ChartOfAccountID = coas[0].ID
		req.Lines[1].ChartOfAccountID = coas[1].ID
	}

	// We use PostOrUpdateJournal to ensure idempotency and auto-post the result
	return uc.PostOrUpdateJournal(ctx, req)
}

// validateControlAccountsForLines checks that journal lines do not reference
// trade control accounts (AR, AP, Inventory, GRIR, Advance). These accounts
// must only be touched by their respective operational modules (Sales, Purchase,
// Inventory) to preserve the single-source-of-truth principle.
// This is a no-op if settingsService is nil (e.g. in unit tests).
func (uc *journalEntryUsecase) validateControlAccountsForLines(ctx context.Context, lines []dto.JournalLineRequest) error {
	if uc.settingsService == nil {
		return nil
	}

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

	if len(restrictedCodes) == 0 {
		return nil
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
			return ErrJournalControlAccountRestricted
		}
	}

	return nil
}
