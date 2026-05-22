package usecase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	financeRepositories "github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	financeDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/repositories"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/mapper"
	purchaseService "github.com/gilabs/gims/api/internal/purchase/domain/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPurchasePaymentNotFound     = errors.New("purchase payment not found")
	ErrPurchasePaymentConflict     = errors.New("purchase payment conflict")
	ErrInvalidPaymentDate          = errors.New("invalid payment date format")
	payableSupplierInvoiceStatuses = []models.SupplierInvoiceStatus{
		models.SupplierInvoiceStatusUnpaid,
		models.SupplierInvoiceStatusPartial,
		models.SupplierInvoiceStatusWaitingPayment,
	}
)

type PurchasePaymentUsecase interface {
	AddData(ctx context.Context) (*dto.PurchasePaymentAddResponse, error)
	List(ctx context.Context, params repositories.PurchasePaymentListParams) ([]*dto.PurchasePaymentListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error)
	Create(ctx context.Context, req *dto.CreatePurchasePaymentRequest) (*dto.PurchasePaymentDetailResponse, error)
	CreateBatch(ctx context.Context, req *dto.CreatePurchasePaymentBatchRequest) (*dto.PurchasePaymentBatchResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdatePurchasePaymentRequest) (*dto.PurchasePaymentDetailResponse, error)
	Delete(ctx context.Context, id string) error
	Confirm(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error)
	ConfirmBatch(ctx context.Context, req *dto.ConfirmPurchasePaymentBatchRequest) (*dto.PurchasePaymentBatchResponse, error)
	Reverse(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error)
	ReverseWithReason(ctx context.Context, id string, reason string) (*dto.PurchasePaymentDetailResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.PurchasePaymentAuditTrailEntry, int64, error)
	ExportCSV(ctx context.Context, params repositories.PurchasePaymentListParams) ([]byte, error)
	TriggerJournalForPayment(ctx context.Context, pay *models.PurchasePayment) error
}

type purchasePaymentUsecase struct {
	db                 *gorm.DB
	repo               repositories.PurchasePaymentRepository
	siRepo             repositories.SupplierInvoiceRepository
	auditService       audit.AuditService
	mapper             *mapper.PurchasePaymentMapper
	journalUC          finUsecase.JournalEntryUsecase
	coaUC              finUsecase.ChartOfAccountUsecase
	engine             accounting.AccountingEngine
	settingsUC         financesettings.SettingsService
	purchaseJournalSvc purchaseService.PurchaseJournalService
	cashBankTxnUC      finUsecase.CashBankTransactionUsecase
}

func NewPurchasePaymentUsecase(db *gorm.DB, repo repositories.PurchasePaymentRepository, siRepo repositories.SupplierInvoiceRepository, auditService audit.AuditService, journalUC finUsecase.JournalEntryUsecase, coaUC finUsecase.ChartOfAccountUsecase, engine accounting.AccountingEngine, settingsUC financesettings.SettingsService, cashBankTxnUC finUsecase.CashBankTransactionUsecase) PurchasePaymentUsecase {
	uc := &purchasePaymentUsecase{db: db, repo: repo, siRepo: siRepo, auditService: auditService, mapper: mapper.NewPurchasePaymentMapper(), journalUC: journalUC, coaUC: coaUC, engine: engine, settingsUC: settingsUC, cashBankTxnUC: cashBankTxnUC}
	uc.purchaseJournalSvc = purchaseService.NewPurchaseJournalService(db, journalUC, engine)
	return uc
}

func normalizePurchasePaymentMethod(raw string) (models.PurchasePaymentMethod, error) {
	method := strings.ToUpper(strings.TrimSpace(raw))
	if method != string(models.PurchasePaymentMethodBank) && method != string(models.PurchasePaymentMethodCash) {
		return "", ErrPurchasePaymentConflict
	}

	return models.PurchasePaymentMethod(method), nil
}

func dedupePaymentIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}

func trimOptionalString(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func resolvePurchasePaymentBankAccountForCreate(
	tx *gorm.DB,
	bankAccountID string,
	method models.PurchasePaymentMethod,
	actorID string,
	companyID string,
	ctx context.Context,
	settingsUC financesettings.SettingsService,
	coaUC finUsecase.ChartOfAccountUsecase,
) (*coreModels.BankAccount, error) {
	if method == models.PurchasePaymentMethodCash {
		resolvedCompanyID := strings.TrimSpace(companyID)
		if resolvedCompanyID == "" {
			derivedCompanyID, err := resolvePurchasePaymentActorCompanyID(tx, actorID)
			if err != nil {
				return nil, err
			}
			resolvedCompanyID = derivedCompanyID
		}

		cashBankAccount, err := resolvePurchaseCashBankAccountForCompany(ctx, tx, resolvedCompanyID, settingsUC, coaUC)
		if err != nil {
			cashBankAccount, err = lockAnyActivePurchasePaymentBankAccountForCompany(tx, resolvedCompanyID)
			if err != nil {
				return nil, err
			}
		}

		if bankAccountID != "" {
			selectedBankAccount, err := lockActivePurchasePaymentBankAccount(tx, bankAccountID)
			if err == nil && selectedBankAccount.ChartOfAccountID != nil && cashBankAccount.ChartOfAccountID != nil && strings.TrimSpace(*selectedBankAccount.ChartOfAccountID) == strings.TrimSpace(*cashBankAccount.ChartOfAccountID) {
				return selectedBankAccount, nil
			}
		}

		return cashBankAccount, nil
	}

	if bankAccountID != "" {
		return lockActivePurchasePaymentBankAccount(tx, bankAccountID)
	}

	resolvedCompanyID := strings.TrimSpace(companyID)
	if resolvedCompanyID == "" {
		derivedCompanyID, err := resolvePurchasePaymentActorCompanyID(tx, actorID)
		if err != nil {
			return nil, err
		}
		resolvedCompanyID = derivedCompanyID
	}

	return lockAnyActivePurchasePaymentBankAccountForCompany(tx, resolvedCompanyID)
}

func resolvePurchasePaymentActorCompanyID(tx *gorm.DB, actorID string) (string, error) {
	trimmedActorID := strings.TrimSpace(actorID)
	if trimmedActorID == "" {
		return "", errors.New("user not authenticated")
	}

	type employeeCompanyRow struct {
		CompanyID string
	}

	var row employeeCompanyRow
	if err := tx.Table("employees").
		Select("company_id").
		Where("user_id = ? AND deleted_at IS NULL", trimmedActorID).
		Limit(1).
		Scan(&row).Error; err != nil {
		return "", err
	}

	companyID := strings.TrimSpace(row.CompanyID)
	if companyID == "" {
		return "", errors.New("employee company not found")
	}

	return companyID, nil
}

func lockAnyActivePurchasePaymentBankAccountForCompany(tx *gorm.DB, companyID string) (*coreModels.BankAccount, error) {
	var ba coreModels.BankAccount
	if err := tx.Model(&coreModels.BankAccount{}).
		Where("company_id = ?", companyID).
		Where("is_active = ?", true).
		Order("updated_at DESC").
		Order("created_at DESC").
		First(&ba).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("no active cash/bank account configured for this company")
		}
		return nil, err
	}

	return &ba, nil
}

func lockActivePurchasePaymentBankAccount(tx *gorm.DB, bankAccountID string) (*coreModels.BankAccount, error) {
	var ba coreModels.BankAccount
	if err := tx.First(&ba, "id = ?", bankAccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("bank account not found")
		}
		return nil, err
	}

	if !ba.IsActive {
		return nil, ErrPurchasePaymentConflict
	}

	return &ba, nil
}

func resolvePurchaseCashBankAccountForCompany(
	ctx context.Context,
	tx *gorm.DB,
	companyID string,
	settingsUC financesettings.SettingsService,
	coaUC finUsecase.ChartOfAccountUsecase,
) (*coreModels.BankAccount, error) {
	if settingsUC == nil || coaUC == nil {
		return nil, errors.New("finance settings/coA dependency is required for cash payments")
	}

	coaCode, err := settingsUC.GetCOAByKey(ctx, "finance.cash_default")
	if err != nil || strings.TrimSpace(coaCode) == "" {
		coaCode, err = settingsUC.GetCOACode(ctx, "coa.cash")
		if err != nil {
			return nil, err
		}
	}

	coa, err := coaUC.GetByCode(ctx, strings.TrimSpace(coaCode))
	if err != nil {
		return nil, err
	}

	var bankAccount coreModels.BankAccount
	if err := tx.Where("company_id = ? AND is_active = ? AND chart_of_account_id = ?", strings.TrimSpace(companyID), true, coa.ID).
		Order("updated_at DESC").
		Order("created_at DESC").
		First(&bankAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	return &bankAccount, nil
}

func isPayableSupplierInvoiceStatus(status models.SupplierInvoiceStatus) bool {
	for _, allowed := range payableSupplierInvoiceStatuses {
		if status == allowed {
			return true
		}
	}

	return false
}

func isSamePendingPaymentRequest(
	pending *models.PurchasePayment,
	bankAccountID string,
	paymentDate string,
	method models.PurchasePaymentMethod,
	amount float64,
) bool {
	if pending == nil {
		return false
	}

	pendingBA := ""
	if pending.BankAccountID != nil {
		pendingBA = strings.TrimSpace(*pending.BankAccountID)
	}
	if pendingBA != strings.TrimSpace(bankAccountID) {
		return false
	}
	if pending.PaymentDate.Format("2006-01-02") != strings.TrimSpace(paymentDate) {
		return false
	}
	if pending.Method != method {
		return false
	}

	return math.Abs(pending.Amount-amount) <= 0.0001
}

func (uc *purchasePaymentUsecase) AddData(ctx context.Context) (*dto.PurchasePaymentAddResponse, error) {
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	var bankAccounts []coreModels.BankAccount
	if err := database.GetDB(ctx, uc.db).
		Model(&coreModels.BankAccount{}).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&bankAccounts).Error; err != nil {
		return nil, err
	}

	baItems := make([]*dto.PurchasePaymentBankAccountSummary, 0, len(bankAccounts))
	for i := range bankAccounts {
		ba := bankAccounts[i]
		baItems = append(baItems, &dto.PurchasePaymentBankAccountSummary{ID: ba.ID, Name: ba.Name, AccountNumber: ba.AccountNumber, AccountHolder: ba.AccountHolder, Currency: ba.Currency})
	}

	var invoices []*models.SupplierInvoice
	if err := database.GetDB(ctx, uc.db).
		Model(&models.SupplierInvoice{}).
		Preload("PurchaseOrder").
		Where("status IN ?", payableSupplierInvoiceStatuses).
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, err
	}

	invItems := make([]*dto.PurchasePaymentAddInvoiceItem, 0, len(invoices))
	for _, inv := range invoices {
		var poObj *struct {
			ID   string `json:"id"`
			Code string `json:"code"`
		}
		if inv.PurchaseOrder != nil {
			poObj = &struct {
				ID   string `json:"id"`
				Code string `json:"code"`
			}{ID: inv.PurchaseOrder.ID, Code: inv.PurchaseOrder.Code}
		}

		invItems = append(invItems, &dto.PurchasePaymentAddInvoiceItem{
			ID:              inv.ID,
			PurchaseOrder:   poObj,
			Code:            inv.Code,
			InvoiceNumber:   inv.InvoiceNumber,
			Type:            string(inv.Type),
			InvoiceDate:     inv.InvoiceDate.Format("2006-01-02"),
			DueDate:         inv.DueDate.Format("2006-01-02"),
			Amount:          inv.Amount,
			PaidAmount:      inv.PaidAmount,
			RemainingAmount: inv.RemainingAmount,
			Status:          string(inv.Status),
		})
	}

	return &dto.PurchasePaymentAddResponse{BankAccounts: baItems, Invoices: invItems}, nil
}

func (uc *purchasePaymentUsecase) List(ctx context.Context, params repositories.PurchasePaymentListParams) ([]*dto.PurchasePaymentListResponse, int64, error) {
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return uc.mapper.ToListResponseList(items), total, nil
}

func (uc *purchasePaymentUsecase) GetByID(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.PurchasePayment{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, ErrPurchasePaymentNotFound
	}

	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchasePaymentNotFound
		}
		return nil, err
	}
	return uc.mapper.ToDetailResponse(p), nil
}

func (uc *purchasePaymentUsecase) Create(ctx context.Context, req *dto.CreatePurchasePaymentRequest) (*dto.PurchasePaymentDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	method, err := normalizePurchasePaymentMethod(req.Method)
	if err != nil {
		return nil, err
	}

	var createdID string
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ?", strings.TrimSpace(req.InvoiceID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrSupplierInvoiceNotFound
			}
			return err
		}
		if !isPayableSupplierInvoiceStatus(inv.Status) {
			return ErrPurchasePaymentConflict
		}

		companyID := ""
		if inv.CompanyID != nil {
			companyID = strings.TrimSpace(*inv.CompanyID)
		}

		ba, err := resolvePurchasePaymentBankAccountForCreate(
			tx,
			trimOptionalString(req.BankAccountID),
			method,
			actorID,
			companyID,
			ctx,
			uc.settingsUC,
			uc.coaUC,
		)
		if err != nil {
			return err
		}

		// Idempotency for retried submit: if the same pending payment already exists, return it.
		var existingPending models.PurchasePayment
		err = tx.Session(&gorm.Session{NewDB: true}).Model(&models.PurchasePayment{}).
			Where("supplier_invoice_id = ? AND status = ?", inv.ID, models.PurchasePaymentStatusPending).
			Order("created_at DESC").
			Take(&existingPending).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil {
			if isSamePendingPaymentRequest(&existingPending, ba.ID, req.PaymentDate, method, req.Amount) {
				createdID = existingPending.ID
				return nil
			}
			return ErrPurchasePaymentConflict
		}
		if req.Amount > inv.RemainingAmount+0.0001 {
			return ErrPurchasePaymentConflict
		}

		amount := math.Max(0, req.Amount)
		if amount <= 0 {
			return ErrPurchasePaymentConflict
		}

		// For CASH payments, do not store bank account ID (COA resolved from mapping at journal time)
		var bankAccountID *string
		if method != models.PurchasePaymentMethodCash {
			id := strings.TrimSpace(ba.ID)
			bankAccountID = &id
		}

		resolvedCompanyID := strings.TrimSpace(companyID)
		if resolvedCompanyID == "" && inv.CompanyID != nil {
			resolvedCompanyID = strings.TrimSpace(*inv.CompanyID)
		}

		paymentDateParsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.PaymentDate))
		if err != nil {
			return ErrInvalidPaymentDate
		}

		p := &models.PurchasePayment{
			SupplierInvoiceID: inv.ID,
			BankAccountID:     bankAccountID,
			PaymentDate:       paymentDateParsed,
			Amount:            amount,
			Method:            method,
			Status:            models.PurchasePaymentStatusPending,
			ReferenceNumber:   req.ReferenceNumber,
			Notes:             req.Notes,
			CreatedBy:         actorID,
		}
		p.CompanyID = resolvedCompanyID
		p.FiscalYearID = inv.FiscalYearID
		if p.CompanyID == "" || p.FiscalYearID == nil {
			companyID, fiscalYearID, err := newPurchaseFiscalYearResolver(uc.db, financeRepositories.NewFiscalYearRepository(uc.db)).Resolve(ctx, inv.InvoiceDate.Format("2006-01-02"))
			if err != nil {
				return err
			}
			p.CompanyID = companyID
			p.FiscalYearID = &fiscalYearID
		}
		// Only snapshot bank account for BANK method
		if method == models.PurchasePaymentMethodBank {
			snapshotPurchasePayment(p, ba)
		}

		transactionCOAID, err := uc.resolveTransactionCOAForJournal(ctx, p, ba)
		if err != nil {
			return err
		}
		p.SnapshotCOAID = &transactionCOAID

		if err := tx.Create(p).Error; err != nil {
			return err
		}

		if err := tx.Model(&inv).Updates(map[string]interface{}{
			"status":     models.SupplierInvoiceStatusWaitingPayment,
			"updated_at": apptime.Now(),
		}).Error; err != nil {
			return err
		}

		createdID = p.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	out, err := uc.repo.GetByID(ctx, createdID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_payment.create", out.ID, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *purchasePaymentUsecase) CreateBatch(ctx context.Context, req *dto.CreatePurchasePaymentBatchRequest) (*dto.PurchasePaymentBatchResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	method, err := normalizePurchasePaymentMethod(req.Method)
	if err != nil {
		return nil, err
	}

	paymentDate := strings.TrimSpace(req.PaymentDate)
	if paymentDate == "" {
		return nil, ErrPurchasePaymentConflict
	}

	paymentDateParsed, err := time.Parse("2006-01-02", paymentDate)
	if err != nil {
		return nil, ErrInvalidPaymentDate
	}

	normalizedItems := make([]dto.CreatePurchasePaymentBatchItemRequest, 0, len(req.Items))
	seenInvoices := make(map[string]struct{}, len(req.Items))
	for _, item := range req.Items {
		invoiceID := strings.TrimSpace(item.InvoiceID)
		if invoiceID == "" {
			return nil, ErrPurchasePaymentConflict
		}
		if _, exists := seenInvoices[invoiceID]; exists {
			return nil, ErrPurchasePaymentConflict
		}
		seenInvoices[invoiceID] = struct{}{}

		amount := math.Max(0, item.Amount)
		if amount <= 0 {
			return nil, ErrPurchasePaymentConflict
		}

		normalizedItems = append(normalizedItems, dto.CreatePurchasePaymentBatchItemRequest{
			InvoiceID: invoiceID,
			Amount:    amount,
		})
	}

	if len(normalizedItems) == 0 {
		return nil, ErrPurchasePaymentConflict
	}

	createdIDs := make([]string, 0, len(normalizedItems))
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ba, err := resolvePurchasePaymentBankAccountForCreate(
			tx,
			trimOptionalString(req.BankAccountID),
			method,
			actorID,
			"",
			ctx,
			uc.settingsUC,
			uc.coaUC,
		)
		if err != nil {
			return err
		}

		for _, item := range normalizedItems {
			var inv models.SupplierInvoice
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ?", item.InvoiceID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return ErrSupplierInvoiceNotFound
				}
				return err
			}

			if !isPayableSupplierInvoiceStatus(inv.Status) {
				return ErrPurchasePaymentConflict
			}

			var pendingCount int64
			if err := tx.Session(&gorm.Session{NewDB: true}).Model(&models.PurchasePayment{}).
				Where("supplier_invoice_id = ? AND status = ?", inv.ID, models.PurchasePaymentStatusPending).
				Count(&pendingCount).Error; err != nil {
				return err
			}
			if pendingCount > 0 {
				return ErrPurchasePaymentConflict
			}

			if item.Amount > inv.RemainingAmount+0.0001 {
				return ErrPurchasePaymentConflict
			}

			pay := &models.PurchasePayment{
				SupplierInvoiceID: inv.ID,
				PaymentDate:       paymentDateParsed,
				Amount:            item.Amount,
				Method:            method,
				Status:            models.PurchasePaymentStatusPending,
				ReferenceNumber:   req.ReferenceNumber,
				Notes:             req.Notes,
				CreatedBy:         actorID,
			}
			// assign bank account id pointer
			baID := strings.TrimSpace(ba.ID)
			pay.BankAccountID = &baID
			if inv.CompanyID != nil {
				pay.CompanyID = strings.TrimSpace(*inv.CompanyID)
			}
			pay.FiscalYearID = inv.FiscalYearID
			if pay.CompanyID == "" || pay.FiscalYearID == nil {
				companyID, fiscalYearID, err := newPurchaseFiscalYearResolver(uc.db, financeRepositories.NewFiscalYearRepository(uc.db)).Resolve(ctx, inv.InvoiceDate.Format("2006-01-02"))
				if err != nil {
					return err
				}
				pay.CompanyID = companyID
				pay.FiscalYearID = &fiscalYearID
			}
			// Only snapshot bank account for BANK method
			if method == models.PurchasePaymentMethodBank {
				snapshotPurchasePayment(pay, ba)
			}

			transactionCOAID, err := uc.resolveTransactionCOAForJournal(ctx, pay, ba)
			if err != nil {
				return err
			}
			pay.SnapshotCOAID = &transactionCOAID

			if err := tx.Create(pay).Error; err != nil {
				return err
			}

			if err := tx.Model(&inv).Updates(map[string]interface{}{
				"status":     models.SupplierInvoiceStatusWaitingPayment,
				"updated_at": apptime.Now(),
			}).Error; err != nil {
				return err
			}

			createdIDs = append(createdIDs, pay.ID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	payments := make([]*dto.PurchasePaymentDetailResponse, 0, len(createdIDs))
	totalAmount := 0.0
	for _, paymentID := range createdIDs {
		row, err := uc.repo.GetByID(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		mapped := uc.mapper.ToDetailResponse(row)
		payments = append(payments, mapped)
		totalAmount += mapped.Amount
	}

	auditTargetID := "batch"
	if len(createdIDs) > 0 {
		auditTargetID = createdIDs[0]
	}
	uc.auditService.Log(ctx, "purchase_payment.batch_create", auditTargetID, map[string]interface{}{
		"payment_ids":   createdIDs,
		"payment_count": len(createdIDs),
		"total_amount":  totalAmount,
	})

	return &dto.PurchasePaymentBatchResponse{
		Payments:    payments,
		TotalAmount: totalAmount,
		Count:       len(payments),
	}, nil
}

func (uc *purchasePaymentUsecase) Update(ctx context.Context, id string, req *dto.UpdatePurchasePaymentRequest) (*dto.PurchasePaymentDetailResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method != string(models.PurchasePaymentMethodBank) && method != string(models.PurchasePaymentMethodCash) {
		return nil, ErrPurchasePaymentConflict
	}

	var updatedID string
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pay models.PurchasePayment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pay, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrPurchasePaymentNotFound
			}
			return err
		}
		if pay.Status != models.PurchasePaymentStatusPending {
			return ErrPurchasePaymentConflict
		}

		var inv models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ?", pay.SupplierInvoiceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrSupplierInvoiceNotFound
			}
			return err
		}
		if inv.Status != models.SupplierInvoiceStatusUnpaid && inv.Status != models.SupplierInvoiceStatusPartial && inv.Status != models.SupplierInvoiceStatusWaitingPayment {
			return ErrPurchasePaymentConflict
		}
		available := inv.RemainingAmount + pay.Amount
		if req.Amount > available+0.0001 {
			return ErrPurchasePaymentConflict
		}

		var ba coreModels.BankAccount
		if err := tx.First(&ba, "id = ?", strings.TrimSpace(req.BankAccountID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("bank account not found")
			}
			return err
		}
		if !ba.IsActive {
			return ErrPurchasePaymentConflict
		}

		amount := math.Max(0, req.Amount)
		if amount <= 0 {
			return ErrPurchasePaymentConflict
		}

		updates := map[string]interface{}{
			"bank_account_id":                ba.ID,
			"payment_date":                   strings.TrimSpace(req.PaymentDate),
			"amount":                         amount,
			"method":                         models.PurchasePaymentMethod(method),
			"reference_number":               req.ReferenceNumber,
			"notes":                          req.Notes,
			"updated_at":                     apptime.Now(),
			"bank_account_name_snapshot":     ba.Name,
			"bank_account_number_snapshot":   ba.AccountNumber,
			"bank_account_holder_snapshot":   ba.AccountHolder,
			"bank_account_currency_snapshot": ba.Currency,
		}
		if err := tx.Model(&pay).Updates(updates).Error; err != nil {
			return err
		}

		updatedID = pay.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	out, err := uc.repo.GetByID(ctx, updatedID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_payment.update", out.ID, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *purchasePaymentUsecase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPurchasePaymentNotFound
		}
		return err
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pay models.PurchasePayment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pay, "id = ?", id).Error; err != nil {
			return err
		}
		if pay.Status != models.PurchasePaymentStatusPending {
			return ErrPurchasePaymentConflict
		}

		var inv models.SupplierInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ?", pay.SupplierInvoiceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrSupplierInvoiceNotFound
			}
			return err
		}

		type sumRow struct{ Total float64 }
		var row sumRow
		if err := tx.Model(&models.PurchasePayment{}).
			Select("COALESCE(SUM(amount),0) as total").
			Where("supplier_invoice_id = ?", inv.ID).
			Where("status = ?", models.PurchasePaymentStatusConfirmed).
			Scan(&row).Error; err != nil {
			return err
		}

		totalSettled := row.Total + inv.DownPaymentAmount
		restoredStatus := models.SupplierInvoiceStatusUnpaid
		if totalSettled > 0 {
			restoredStatus = models.SupplierInvoiceStatusPartial
		}
		if totalSettled >= inv.Amount-0.0001 {
			restoredStatus = models.SupplierInvoiceStatusPaid
		}

		updateData := map[string]interface{}{
			"status":           restoredStatus,
			"paid_amount":      row.Total,
			"remaining_amount": math.Max(0, inv.Amount-totalSettled),
			"updated_at":       apptime.Now(),
		}
		if restoredStatus == models.SupplierInvoiceStatusPaid {
			now := apptime.Now()
			updateData["payment_at"] = &now
		} else {
			updateData["payment_at"] = nil
		}

		if err := tx.Model(&inv).Updates(updateData).Error; err != nil {
			return err
		}

		if err := tx.Delete(&pay).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrPurchasePaymentNotFound
		}
		return err
	}

	uc.auditService.Log(ctx, "purchase_payment.delete", id, map[string]interface{}{"before": existing})
	return nil
}

func (uc *purchasePaymentUsecase) Confirm(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}
	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	confirmedID := ""
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		confirmedID, err = uc.confirmPaymentTx(ctx, tx, strings.TrimSpace(id))
		return err
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchasePaymentNotFound
		}
		return nil, err
	}

	out, err := uc.repo.GetByID(ctx, confirmedID)
	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_payment.confirm", id, map[string]interface{}{"after": out})
	return uc.mapper.ToDetailResponse(out), nil
}

func (uc *purchasePaymentUsecase) ConfirmBatch(ctx context.Context, req *dto.ConfirmPurchasePaymentBatchRequest) (*dto.PurchasePaymentBatchResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if uc.db == nil {
		return nil, errors.New("db is nil")
	}

	paymentIDs := dedupePaymentIDs(req.PaymentIDs)
	if len(paymentIDs) == 0 {
		return nil, ErrPurchasePaymentConflict
	}

	confirmedIDs := make([]string, 0, len(paymentIDs))
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, paymentID := range paymentIDs {
			confirmedID, err := uc.confirmPaymentTx(ctx, tx, paymentID)
			if err != nil {
				return err
			}
			confirmedIDs = append(confirmedIDs, confirmedID)
		}

		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPurchasePaymentNotFound
		}
		return nil, err
	}

	payments := make([]*dto.PurchasePaymentDetailResponse, 0, len(confirmedIDs))
	totalAmount := 0.0
	for _, paymentID := range confirmedIDs {
		row, err := uc.repo.GetByID(ctx, paymentID)
		if err != nil {
			return nil, err
		}
		mapped := uc.mapper.ToDetailResponse(row)
		payments = append(payments, mapped)
		totalAmount += mapped.Amount
	}

	auditTargetID := "batch"
	if len(confirmedIDs) > 0 {
		auditTargetID = confirmedIDs[0]
	}
	uc.auditService.Log(ctx, "purchase_payment.batch_confirm", auditTargetID, map[string]interface{}{
		"payment_ids":   confirmedIDs,
		"payment_count": len(confirmedIDs),
		"total_amount":  totalAmount,
	})

	return &dto.PurchasePaymentBatchResponse{
		Payments:    payments,
		TotalAmount: totalAmount,
		Count:       len(payments),
	}, nil
}

func (uc *purchasePaymentUsecase) confirmPaymentTx(ctx context.Context, tx *gorm.DB, paymentID string) (string, error) {
	var pay models.PurchasePayment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pay, "id = ?", paymentID).Error; err != nil {
		return "", err
	}
	if pay.Status != models.PurchasePaymentStatusPending {
		return "", ErrPurchasePaymentConflict
	}

	var inv models.SupplierInvoice
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inv, "id = ?", pay.SupplierInvoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", ErrSupplierInvoiceNotFound
		}
		return "", err
	}
	if !isPayableSupplierInvoiceStatus(inv.Status) {
		return "", ErrPurchasePaymentConflict
	}

	// Sum already confirmed cash payments.
	type sumRow struct{ Total float64 }
	var row sumRow
	if err := tx.Model(&models.PurchasePayment{}).
		Select("COALESCE(SUM(amount),0) as total").
		Where("supplier_invoice_id = ?", inv.ID).
		Where("status = ?", models.PurchasePaymentStatusConfirmed).
		Scan(&row).Error; err != nil {
		return "", err
	}

	// Total settled amount = cash payments + this payment + down payment.
	totalSettled := row.Total + pay.Amount + inv.DownPaymentAmount
	if totalSettled > inv.Amount+0.0001 {
		return "", ErrPurchasePaymentConflict
	}

	if err := tx.Model(&pay).Updates(map[string]interface{}{"status": models.PurchasePaymentStatusConfirmed, "updated_at": apptime.Now()}).Error; err != nil {
		return "", err
	}

	newStatus := models.SupplierInvoiceStatusPartial
	if totalSettled >= inv.Amount-0.0001 {
		newStatus = models.SupplierInvoiceStatusPaid
	}

	updateData := map[string]interface{}{
		"status":           newStatus,
		"paid_amount":      row.Total + pay.Amount,
		"remaining_amount": math.Max(0, inv.Amount-totalSettled),
		"updated_at":       apptime.Now(),
	}
	if newStatus == models.SupplierInvoiceStatusPaid {
		now := apptime.Now()
		updateData["payment_at"] = &now
	}

	if err := tx.Model(&inv).Updates(updateData).Error; err != nil {
		return "", err
	}

	if newStatus == models.SupplierInvoiceStatusPaid && inv.Type == models.SupplierInvoiceTypeDownPayment && inv.PurchaseOrderID != "" {
		type dpSumRow struct{ Total float64 }
		var dpSum dpSumRow
		if err := tx.Model(&models.SupplierInvoice{}).
			Select("COALESCE(SUM(paid_amount),0) as total").
			Where("purchase_order_id = ?", inv.PurchaseOrderID).
			Where("type = ?", models.SupplierInvoiceTypeDownPayment).
			Where("status = ?", models.SupplierInvoiceStatusPaid).
			Where("deleted_at IS NULL").
			Scan(&dpSum).Error; err != nil {
			return "", err
		}

		var regularInvoices []models.SupplierInvoice
		if err := tx.Where("purchase_order_id = ?", inv.PurchaseOrderID).
			Where("type = ?", models.SupplierInvoiceTypeNormal).
			Where("deleted_at IS NULL").
			Find(&regularInvoices).Error; err != nil {
			return "", err
		}

		for _, regInv := range regularInvoices {
			totalDeductions := regInv.PaidAmount + dpSum.Total
			newRemaining := math.Max(0, regInv.Amount-totalDeductions)

			regStatus := models.SupplierInvoiceStatusUnpaid
			if newRemaining <= 0.0001 && regInv.Amount > 0 {
				regStatus = models.SupplierInvoiceStatusPaid
			} else if totalDeductions > 0 {
				regStatus = models.SupplierInvoiceStatusPartial
			}

			dpInvID := inv.ID
			regUpdates := map[string]interface{}{
				"status":                  regStatus,
				"down_payment_invoice_id": &dpInvID,
				"down_payment_amount":     dpSum.Total,
				"remaining_amount":        newRemaining,
				"updated_at":              apptime.Now(),
			}
			if regStatus == models.SupplierInvoiceStatusPaid && regInv.PaymentAt == nil {
				now := apptime.Now()
				regUpdates["payment_at"] = &now
			}

			if err := tx.Model(&models.SupplierInvoice{}).Where("id = ?", regInv.ID).Updates(regUpdates).Error; err != nil {
				return "", err
			}
			fmt.Printf("✅ Updated Regular Invoice %s: Status %s, DP deducted %.2f, remaining %.2f\n", regInv.Code, regStatus, dpSum.Total, newRemaining)
		}
	}

	// Auto-close the PO when all invoices are paid and all items are fully received.
	// Trigger for both normal invoices paid directly, and when a DP payment causes the linked
	// normal invoice to be fully settled via deduction.
	if newStatus == models.SupplierInvoiceStatusPaid && inv.PurchaseOrderID != "" {
		closePurchaseOrderWhenSettledAndFulfilled(ctx, tx, inv.PurchaseOrderID)
	}

	// Journal posting is part of the same DB transaction.
	txCtx := database.WithTx(ctx, tx)
	loadedPay, loadErr := uc.repo.GetByID(txCtx, pay.ID)
	if loadErr != nil {
		log.Printf("[JOURNAL_ERROR] Failed to load payment %s for journal: %v", pay.ID, loadErr)
		return "", fmt.Errorf("failed to load payment for journal: %w", loadErr)
	}
	log.Printf("[JOURNAL_DEBUG] Posting accounting for purchase payment %s (amount=%.2f, method=%s)", loadedPay.ID, loadedPay.Amount, loadedPay.Method)
	if err := uc.postPaymentAccounting(txCtx, loadedPay); err != nil {
		log.Printf("[JOURNAL_ERROR] Failed to post purchase payment accounting for %s: %v", loadedPay.ID, err)
		return "", fmt.Errorf("failed to post purchase payment accounting: %w", err)
	}
	log.Printf("[PURCHASE_PAYMENT_CONFIRM] Accounting posted successfully for payment %s", loadedPay.ID)

	return pay.ID, nil
}

// closePurchaseOrderWhenSettledAndFulfilled checks if a PO is both fully invoiced/settled
// and fully received, then updates its status to CLOSED. This mirrors the sales order pattern.
func closePurchaseOrderWhenSettledAndFulfilled(ctx context.Context, tx *gorm.DB, orderID string) {
	trimmedOrderID := strings.TrimSpace(orderID)
	if trimmedOrderID == "" {
		return
	}

	if !isPurchaseOrderFullySettled(tx, trimmedOrderID) || !isPurchaseOrderFullyFulfilled(ctx, tx, trimmedOrderID) {
		return
	}

	var po models.PurchaseOrder
	if err := tx.First(&po, "id = ?", trimmedOrderID).Error; err == nil {
		if po.Status != models.PurchaseOrderStatusClosed {
			now := apptime.Now()
			_ = tx.Model(&po).Updates(map[string]interface{}{
				"status":     models.PurchaseOrderStatusClosed,
				"closed_at":  &now,
				"updated_at": now,
			}).Error
		}
	}
}

// isPurchaseOrderFullySettled returns true when the PO has at least one active
// NORMAL supplier invoice and all active NORMAL invoices are in terminal states.
// Down payment invoices alone must never mark a PO as financially settled.
func isPurchaseOrderFullySettled(tx *gorm.DB, orderID string) bool {
	var totalInvoiceCount int64
	if err := tx.Model(&models.SupplierInvoice{}).
		Where("purchase_order_id = ?", orderID).
		Where("type = ?", models.SupplierInvoiceTypeNormal).
		Where("deleted_at IS NULL").
		Count(&totalInvoiceCount).Error; err != nil {
		return false
	}

	if totalInvoiceCount == 0 {
		return false
	}

	var pendingInvoiceCount int64
	if err := tx.Model(&models.SupplierInvoice{}).
		Where("purchase_order_id = ?", orderID).
		Where("type = ?", models.SupplierInvoiceTypeNormal).
		Where("status NOT IN ?", []models.SupplierInvoiceStatus{
			models.SupplierInvoiceStatusPaid,
			models.SupplierInvoiceStatusCancelled,
			models.SupplierInvoiceStatusRejected,
		}).
		Where("deleted_at IS NULL").
		Count(&pendingInvoiceCount).Error; err != nil {
		return false
	}

	return pendingInvoiceCount == 0
}

// isPurchaseOrderFullyFulfilled returns true when every PO item has been fully received
// via confirmed goods receipts.
func isPurchaseOrderFullyFulfilled(ctx context.Context, tx *gorm.DB, orderID string) bool {
	var po models.PurchaseOrder
	if err := tx.Preload("Items").First(&po, "id = ?", orderID).Error; err != nil {
		return false
	}

	if po.Status != models.PurchaseOrderStatusApproved || len(po.Items) == 0 {
		return false
	}

	for _, item := range po.Items {
		if item.Quantity <= 0 {
			return false
		}

		var totalReceived float64
		if err := tx.Table("goods_receipt_items").
			Select("COALESCE(SUM(goods_receipt_items.quantity_received), 0)").
			Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.goods_receipt_id").
			Where("goods_receipts.purchase_order_id = ?", orderID).
			Where("goods_receipts.status IN ?", []string{
				string(models.GoodsReceiptStatusConfirmed),
				string(models.GoodsReceiptStatusClosed),
			}).
			Where("goods_receipt_items.purchase_order_item_id = ?", item.ID).
			Where("goods_receipts.deleted_at IS NULL").
			Scan(&totalReceived).Error; err != nil {
			return false
		}

		if totalReceived+0.0001 < item.Quantity {
			return false
		}
	}

	return true
}

func (uc *purchasePaymentUsecase) ReverseWithReason(ctx context.Context, id string, reason string) (*dto.PurchasePaymentDetailResponse, error) {
	return uc.reverse(ctx, id, reason)
}

func (uc *purchasePaymentUsecase) Reverse(ctx context.Context, id string) (*dto.PurchasePaymentDetailResponse, error) {
	return uc.reverse(ctx, id, "Manual reversal")
}

func (uc *purchasePaymentUsecase) reverse(ctx context.Context, id string, reason string) (*dto.PurchasePaymentDetailResponse, error) {
	var reversed *models.PurchasePayment
	err := uc.db.Transaction(func(tx *gorm.DB) error {
		pay, err := uc.repo.GetByID(database.WithTx(ctx, tx), id)
		if err != nil {
			return err
		}

		if pay.Status != models.PurchasePaymentStatusConfirmed {
			return fmt.Errorf("only confirmed payments can be reversed")
		}

		// Update payment status
		if err := tx.Model(pay).Update("status", models.PurchasePaymentStatusReversed).Error; err != nil {
			return err
		}

		// Update invoice
		var inv models.SupplierInvoice
		if err := tx.First(&inv, "id = ?", pay.SupplierInvoiceID).Error; err != nil {
			return err
		}

		// Recalculate invoice status
		type paySumRow struct{ Total float64 }
		var row paySumRow
		if err := tx.Model(&models.PurchasePayment{}).
			Select("COALESCE(SUM(amount),0) as total").
			Where("supplier_invoice_id = ?", inv.ID).
			Where("status = ?", models.PurchasePaymentStatusConfirmed).
			Where("deleted_at IS NULL").
			Scan(&row).Error; err != nil {
			return err
		}

		// DP total already tracked in inv.DownPaymentAmount
		totalSettled := row.Total + inv.DownPaymentAmount
		newRemaining := math.Max(0, inv.Amount-totalSettled)

		newStatus := models.SupplierInvoiceStatusUnpaid
		if newRemaining <= 0.0001 && inv.Amount > 0 {
			newStatus = models.SupplierInvoiceStatusPaid
		} else if totalSettled > 0 {
			newStatus = models.SupplierInvoiceStatusPartial
		}

		updateData := map[string]interface{}{
			"status":           newStatus,
			"paid_amount":      row.Total,
			"remaining_amount": newRemaining,
			"updated_at":       apptime.Now(),
		}
		if newStatus != models.SupplierInvoiceStatusPaid {
			updateData["payment_at"] = nil
		}

		if err := tx.Model(&inv).Updates(updateData).Error; err != nil {
			return err
		}

		// Trigger journal reversal
		if err := uc.triggerJournalReversed(database.WithTx(ctx, tx), pay, reason); err != nil {
			return fmt.Errorf("failed to reverse journal: %w", err)
		}

		reversed = pay
		return nil
	})

	if err != nil {
		return nil, err
	}

	uc.auditService.Log(ctx, "purchase_payment.reverse", id, map[string]interface{}{
		"status": "REVERSED",
		"reason": reason,
	})

	return uc.mapper.ToDetailResponse(reversed), nil
}

func (uc *purchasePaymentUsecase) triggerJournalReversed(ctx context.Context, pay *models.PurchasePayment, reason string) error {
	if pay == nil || uc.journalUC == nil {
		return nil
	}

	if uc.cashBankTxnUC != nil && pay.CashBankTransactionID != nil {
		cbID := strings.TrimSpace(*pay.CashBankTransactionID)
		if cbID != "" {
			_, err := uc.cashBankTxnUC.Reverse(ctx, cbID, &financeDto.ReverseCashBankTransactionRequest{Reason: reason})
			return err
		}
	}

	if uc.purchaseJournalSvc != nil {
		return uc.purchaseJournalSvc.ReversePurchaseJournal(ctx, reference.RefTypePurchasePayment, pay.ID, reason)
	}

	refType := reference.RefTypePurchasePayment
	var existing financeModels.JournalEntry
	err := database.GetDB(ctx, uc.db).
		Where("reference_type = ? AND reference_id = ?", refType, pay.ID).
		Where("status = ?", financeModels.JournalStatusPosted).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = uc.journalUC.ReverseWithReason(ctx, existing.ID, reason)
	return err
}

func (uc *purchasePaymentUsecase) postPaymentAccounting(ctx context.Context, pay *models.PurchasePayment) error {
	if pay == nil {
		return nil
	}

	if uc.cashBankTxnUC != nil && pay.Method != models.PurchasePaymentMethodCash {
		log.Printf("[ACCOUNTING_DEBUG] postPaymentAccounting: Using Cash/Bank Transaction for payment %s", pay.ID)
		if err := uc.createCashBankTransactionForPayment(ctx, pay); err != nil {
			return err
		}
		return nil
	}

	log.Printf("[ACCOUNTING_DEBUG] postPaymentAccounting: Using Journal Entry trigger for payment %s", pay.ID)
	return uc.triggerJournalEntry(ctx, pay)
}

func (uc *purchasePaymentUsecase) createCashBankTransactionForPayment(ctx context.Context, pay *models.PurchasePayment) error {
	if pay == nil || uc.cashBankTxnUC == nil {
		return nil
	}

	if pay.CashBankTransactionID != nil && strings.TrimSpace(*pay.CashBankTransactionID) != "" {
		return nil
	}

	// Cash payments do not require a bank account; only BANK method payments need one
	if pay.Method == models.PurchasePaymentMethodCash {
		return nil
	}

	var bankAccount coreModels.BankAccount
	if pay.BankAccount != nil {
		bankAccount = *pay.BankAccount
	} else {
		if pay.BankAccountID == nil || strings.TrimSpace(*pay.BankAccountID) == "" {
			return errors.New("bank account is required for purchase payment approval")
		}
		if err := database.GetDB(ctx, uc.db).First(&bankAccount, "id = ?", strings.TrimSpace(*pay.BankAccountID)).Error; err != nil {
			return err
		}
	}
	if strings.TrimSpace(bankAccount.ID) == "" {
		return errors.New("bank account is required for purchase payment approval")
	}

	contraAccountID, err := uc.resolvePaymentContraAccountID(ctx, pay)
	if err != nil {
		return err
	}

	referenceValue := ""
	if pay.ReferenceNumber != nil {
		referenceValue = strings.TrimSpace(*pay.ReferenceNumber)
	}

	invoiceCode := ""
	if pay.SupplierInvoice != nil {
		invoiceCode = strings.TrimSpace(pay.SupplierInvoice.Code)
	}
	description := fmt.Sprintf("Supplier Payment %s", invoiceCode)
	if referenceValue != "" {
		description = fmt.Sprintf("Supplier Payment %s (Ref: %s)", invoiceCode, referenceValue)
	}

	autoPost := true
	req := &financeDto.CreateCashBankTransactionRequest{
		CompanyID:       strings.TrimSpace(bankAccount.CompanyID),
		BankAccountID:   strings.TrimSpace(bankAccount.ID),
		Type:            financeModels.CashBankTransactionTypePaymentOut,
		Date:            pay.PaymentDate.Format("2006-01-02"),
		Amount:          pay.Amount,
		Reference:       referenceValue,
		Description:     description,
		ContraAccountID: &contraAccountID,
		AutoPost:        &autoPost,
	}

	cbTxn, err := uc.cashBankTxnUC.Create(ctx, req)
	if err != nil {
		return err
	}
	if cbTxn == nil || strings.TrimSpace(cbTxn.ID) == "" {
		return errors.New("cash bank transaction response is empty")
	}

	cbTxnID := strings.TrimSpace(cbTxn.ID)
	if err := database.GetDB(ctx, uc.db).Model(&models.PurchasePayment{}).
		Where("id = ?", pay.ID).
		Updates(map[string]interface{}{
			"cash_bank_transaction_id": cbTxnID,
			"updated_at":               apptime.Now(),
		}).Error; err != nil {
		return err
	}

	pay.CashBankTransactionID = &cbTxnID
	return nil
}

func (uc *purchasePaymentUsecase) resolvePaymentContraAccountID(ctx context.Context, pay *models.PurchasePayment) (string, error) {
	if uc.settingsUC == nil || uc.coaUC == nil {
		return "", errors.New("finance settings/coA dependency is required")
	}

	settingKey := "coa.purchase_payable"
	if pay.SupplierInvoice != nil && pay.SupplierInvoice.Type == models.SupplierInvoiceTypeDownPayment {
		settingKey = "coa.purchase_advance"
	}

	keysToTry := purchasePaymentMappingKeys(settingKey)
	keysToTry = append(keysToTry, settingKey)

	var coaCode string
	for _, key := range keysToTry {
		resolvedKey := strings.TrimSpace(key)
		if resolvedKey == "" {
			continue
		}

		code, err := uc.settingsUC.GetCOAByKey(ctx, resolvedKey)
		if err == nil && strings.TrimSpace(code) != "" {
			coaCode = strings.TrimSpace(code)
			break
		}
	}

	if coaCode == "" {
		legacyCode, err := uc.settingsUC.GetCOACode(ctx, settingKey)
		if err != nil {
			return "", err
		}
		coaCode = strings.TrimSpace(legacyCode)
	}

	coa, err := uc.coaUC.GetByCode(ctx, coaCode)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(coa.ID), nil
}

func purchasePaymentMappingKeys(settingKey string) []string {
	switch strings.TrimSpace(settingKey) {
	case "coa.purchase_payable", "coa.accounts_payable":
		return []string{"purchase.accounts_payable"}
	case "coa.purchase_advance", "coa.purchase_advances":
		return []string{"purchase.advance", "purchase.purchase_advance"}
	default:
		return nil
	}
}

func (uc *purchasePaymentUsecase) triggerJournalEntry(ctx context.Context, pay *models.PurchasePayment) error {
	if pay == nil || uc.journalUC == nil || uc.engine == nil {
		return nil
	}

	companyID := strings.TrimSpace(pay.CompanyID)

	// prepare bank account id for resolver (may be nil)
	var payBankAccountID string
	if pay.BankAccountID != nil {
		payBankAccountID = strings.TrimSpace(*pay.BankAccountID)
	}

	resolvedBankAccount, err := resolvePurchasePaymentBankAccountForCreate(
		uc.db.WithContext(ctx),
		payBankAccountID,
		pay.Method,
		pay.CreatedBy,
		companyID,
		ctx,
		uc.settingsUC,
		uc.coaUC,
	)
	if err != nil {
		return err
	}

	baCOAID := ""
	if pay.SnapshotCOAID != nil {
		baCOAID = strings.TrimSpace(*pay.SnapshotCOAID)
	}
	if baCOAID == "" {
		baCOAID, err = uc.resolveTransactionCOAForJournal(ctx, pay, resolvedBankAccount)
		if err != nil {
			return fmt.Errorf("failed to resolve transaction COA for purchase payment journal: %w", err)
		}
		pay.SnapshotCOAID = &baCOAID
		if err := database.GetDB(ctx, uc.db).Model(&models.PurchasePayment{}).
			Where("id = ?", pay.ID).
			Updates(map[string]interface{}{"snapshot_coa_id": baCOAID, "updated_at": apptime.Now()}).Error; err != nil {
			return fmt.Errorf("failed to persist snapshot COA for purchase payment: %w", err)
		}
	}

	reqRefNum := ""
	if pay.ReferenceNumber != nil {
		reqRefNum = *pay.ReferenceNumber
	}

	invoiceCode := ""
	isDP := false
	if pay.SupplierInvoice != nil {
		invoiceCode = pay.SupplierInvoice.Code
		isDP = (pay.SupplierInvoice.Type == models.SupplierInvoiceTypeDownPayment)
	}

	profile := accounting.ProfilePurchasePayment
	if isDP {
		profile = accounting.ProfilePurchasePaymentDP
	}

	companyID, fiscalYearID, err := resolvePurchaseJournalScope(ctx, uc.db, &pay.CompanyID, pay.FiscalYearID, pay.PaymentDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("failed to resolve purchase payment journal scope: %w", err)
	}

	data := accounting.TransactionData{
		ReferenceType:    "PURCHASE_PAYMENT",
		ReferenceID:      pay.ID,
		CompanyID:        companyID,
		FiscalYearID:     fiscalYearID,
		EntryDate:        pay.PaymentDate.Format("2006-01-02"),
		Description:      fmt.Sprintf("Supplier Payment %s (Ref: %s)", invoiceCode, reqRefNum),
		TotalAmount:      pay.Amount,
		TransactionCOAID: baCOAID,
		DescriptionArgs:  []interface{}{invoiceCode, reqRefNum},
	}

	if uc.purchaseJournalSvc != nil {
		if _, err := uc.purchaseJournalSvc.GeneratePurchaseJournal(ctx, purchaseService.PurchaseJournalTxn{
			Profile: profile,
			Data:    data,
		}); err != nil {
			return fmt.Errorf("failed to post purchase payment journal: %w", err)
		}

		log.Printf("journal_observability event=trigger.success module=purchase_payment reference_id=%s", pay.ID)
		return nil
	}

	req, err := uc.engine.GenerateJournal(ctx, profile, data)
	if err != nil {
		return fmt.Errorf("failed to generate purchase payment journal: %w", err)
	}

	// Balance check
	var debitTotal, creditTotal float64
	for _, l := range req.Lines {
		debitTotal += l.Debit
		creditTotal += l.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.001 {
		return fmt.Errorf("generated purchase payment journal is unbalanced: debit=%.2f credit=%.2f", debitTotal, creditTotal)
	}

	req.IsSystemGenerated = true
	_, err = uc.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post purchase payment journal: %w", err)
	}

	pay.ResolvedCOAID = &baCOAID
	if err := database.GetDB(ctx, uc.db).Model(&models.PurchasePayment{}).
		Where("id = ?", pay.ID).
		Updates(map[string]interface{}{"resolved_coa_id": baCOAID, "updated_at": apptime.Now()}).Error; err != nil {
		return fmt.Errorf("failed to persist resolved COA for purchase payment: %w", err)
	}

	log.Printf("journal_observability event=trigger.success module=purchase_payment reference_id=%s", pay.ID)
	return nil
}

func (uc *purchasePaymentUsecase) TriggerJournalForPayment(ctx context.Context, pay *models.PurchasePayment) error {
	if pay == nil {
		return nil
	}

	// When payment is already linked to a cash-bank transaction, accounting is handled
	// through CASH_BANK reference. Re-triggering PURCHASE_PAYMENT journal would duplicate entries.
	if pay.CashBankTransactionID != nil && strings.TrimSpace(*pay.CashBankTransactionID) != "" {
		return nil
	}

	return uc.triggerJournalEntry(ctx, pay)
}

func (uc *purchasePaymentUsecase) resolveTransactionCOAForJournal(ctx context.Context, pay *models.PurchasePayment, resolvedBankAccount *coreModels.BankAccount) (string, error) {
	if pay == nil {
		return "", errors.New("purchase payment is nil")
	}
	if uc.settingsUC == nil || uc.coaUC == nil {
		return "", errors.New("finance settings/coA dependency is required for purchase payment posting")
	}

	if strings.EqualFold(string(pay.Method), string(models.PurchasePaymentMethodCash)) {
		defaultCode, err := uc.settingsUC.GetCOAByKey(ctx, "finance.cash_default")
		if err != nil || strings.TrimSpace(defaultCode) == "" {
			return "", fmt.Errorf("failed to resolve finance.cash_default mapping: %w", err)
		}

		def, err := uc.coaUC.GetByCode(ctx, defaultCode)
		if err != nil {
			return "", fmt.Errorf("failed to fetch COA for finance.cash_default (%s): %w", defaultCode, err)
		}
		return def.ID, nil
	}

	if resolvedBankAccount != nil && resolvedBankAccount.ChartOfAccountID != nil {
		chartOfAccountID := strings.TrimSpace(*resolvedBankAccount.ChartOfAccountID)
		if chartOfAccountID != "" {
			return chartOfAccountID, nil
		}
	}

	return "", errors.New("bank account must have ChartOfAccountID linked for BANK payments")
}

func (uc *purchasePaymentUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.PurchasePaymentAuditTrailEntry, int64, error) {
	if uc.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.PurchasePayment{}, id, security.PurchaseScopeQueryOptions()) {
		return nil, 0, ErrPurchasePaymentNotFound
	}

	tx := uc.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("audit_logs.permission_code LIKE ?", "purchase_payment.%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type auditRow struct {
		ID             string    `gorm:"column:id"`
		ActorID        string    `gorm:"column:actor_id"`
		PermissionCode string    `gorm:"column:permission_code"`
		TargetID       string    `gorm:"column:target_id"`
		Action         string    `gorm:"column:action"`
		Metadata       string    `gorm:"column:metadata"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		ActorEmail     *string   `gorm:"column:actor_email"`
		ActorName      *string   `gorm:"column:actor_name"`
	}

	rows := make([]auditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	entries := make([]dto.PurchasePaymentAuditTrailEntry, 0, len(rows))
	refCache := make(map[string]string)
	for _, r := range rows {
		metaMap := parsePurchaseAuditMetadata(ctx, uc.db, r.Metadata, refCache)

		var usr *dto.AuditTrailUser
		if r.ActorID != "" {
			email := ""
			name := ""
			if r.ActorEmail != nil {
				email = *r.ActorEmail
			}
			if r.ActorName != nil {
				name = *r.ActorName
			}
			usr = &dto.AuditTrailUser{ID: r.ActorID, Email: email, Name: name}
		}

		entries = append(entries, dto.PurchasePaymentAuditTrailEntry{
			ID:             r.ID,
			PermissionCode: r.PermissionCode,
			Action:         r.Action,
			TargetID:       r.TargetID,
			Metadata:       metaMap,
			User:           usr,
			CreatedAt:      r.CreatedAt,
		})
	}

	return entries, total, nil
}

func (uc *purchasePaymentUsecase) ExportCSV(ctx context.Context, params repositories.PurchasePaymentListParams) ([]byte, error) {
	// export uses list params but ignores pagination (caller should cap limit)
	items, _, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"id", "invoice_code", "bank_account", "payment_date", "amount", "method", "status", "created_at"})

	for _, it := range items {
		invCode := ""
		if it.SupplierInvoice != nil {
			invCode = it.SupplierInvoice.Code
		}
		bankName := ""
		if it.BankAccount != nil {
			bankName = it.BankAccount.Name
		}
		_ = w.Write([]string{
			it.ID,
			invCode,
			bankName,
			it.PaymentDate.Format("2006-01-02"),
			strconv.FormatFloat(it.Amount, 'f', 2, 64),
			string(it.Method),
			string(it.Status),
			it.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}
