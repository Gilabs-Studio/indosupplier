package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/mapper"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
)

var (
	ErrBankAccountNotFound = errors.New("bank account not found")
)

type BankAccountUsecase interface {
	Create(ctx context.Context, req *dto.CreateBankAccountRequest) (*dto.BankAccountResponse, error)
	List(ctx context.Context, params repositories.BankAccountListParams) ([]*dto.BankAccountResponse, int64, error)
	ListUnified(ctx context.Context, params repositories.BankAccountListParams) ([]*dto.UnifiedBankAccountResponse, int64, error)
	ListTransactionHistory(ctx context.Context, id string, limit, offset int) ([]dto.BankAccountTransactionResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.BankAccountResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateBankAccountRequest) (*dto.BankAccountResponse, error)
	Delete(ctx context.Context, id string) error
	// Phase 2 methods
	GetByIDWithBalance(ctx context.Context, id string) (*dto.BankAccountResponse, error)
	ListByCompanyWithBalance(ctx context.Context, companyID string, params repositories.BankAccountListParams) ([]*dto.BankAccountResponse, int64, error)
	ToggleStatus(ctx context.Context, bankAccountID string) (*dto.ToggleStatusResponse, error)
}

type bankAccountUsecase struct {
	db           *gorm.DB
	repo         repositories.BankAccountRepository
	currencyRepo repositories.CurrencyRepository
	mapper       *mapper.BankAccountMapper
}

func NewBankAccountUsecase(db *gorm.DB, repo repositories.BankAccountRepository) BankAccountUsecase {
	return &bankAccountUsecase{db: db, repo: repo, mapper: mapper.NewBankAccountMapper()}
}

func NewBankAccountUsecaseWithCurrency(db *gorm.DB, repo repositories.BankAccountRepository, currencyRepo repositories.CurrencyRepository) BankAccountUsecase {
	return &bankAccountUsecase{db: db, repo: repo, currencyRepo: currencyRepo, mapper: mapper.NewBankAccountMapper()}
}

func (u *bankAccountUsecase) resolveCurrency(ctx context.Context, currencyID *string) (*models.Currency, error) {
	if u.currencyRepo == nil || currencyID == nil || *currencyID == "" {
		return nil, nil
	}
	return u.currencyRepo.FindByID(ctx, *currencyID)
}

func (u *bankAccountUsecase) Create(ctx context.Context, req *dto.CreateBankAccountRequest) (*dto.BankAccountResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	resolvedCurrency, err := u.resolveCurrency(ctx, req.CurrencyID)
	if err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	currencyCode := strings.TrimSpace(req.Currency)
	if resolvedCurrency != nil {
		currencyCode = resolvedCurrency.Code
	}
	if currencyCode == "" {
		currencyCode = "IDR"
	}

	m := &models.BankAccount{
		CompanyID:        strings.TrimSpace(req.CompanyID),
		Code:             strings.TrimSpace(req.Code),
		AccountType:      strings.TrimSpace(req.AccountType),
		BankID:           req.BankID,
		Name:             strings.TrimSpace(req.Name),
		AccountNumber:    strings.TrimSpace(req.AccountNumber),
		AccountHolder:    strings.TrimSpace(req.AccountHolder),
		CurrencyID:       req.CurrencyID,
		CurrencyDetail:   resolvedCurrency,
		Currency:         currencyCode,
		ChartOfAccountID: req.ChartOfAccountID,
		VillageID:        req.VillageID,
		BankAddress:      strings.TrimSpace(req.BankAddress),
		BankPhone:        strings.TrimSpace(req.BankPhone),
		CountryCode:      strings.ToUpper(strings.TrimSpace(req.CountryCode)),
		BankBranchCode:   strings.TrimSpace(req.BankBranchCode),
		OpeningBalance:   req.OpeningBalance,
		IsActive:         isActive,
	}
	if m.AccountType == "" {
		m.AccountType = "operational"
	}
	if err := u.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return u.mapper.ToResponse(m), nil
}

func (u *bankAccountUsecase) List(ctx context.Context, params repositories.BankAccountListParams) ([]*dto.BankAccountResponse, int64, error) {
	items, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return u.mapper.ToResponseList(items), total, nil
}

func (u *bankAccountUsecase) ListUnified(ctx context.Context, params repositories.BankAccountListParams) ([]*dto.UnifiedBankAccountResponse, int64, error) {
	items, total, err := u.repo.ListUnified(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*dto.UnifiedBankAccountResponse, 0, len(items))
	for _, item := range items {
		response := &dto.UnifiedBankAccountResponse{
			ID:            item.ID,
			SourceType:    item.SourceType,
			Name:          item.Name,
			BankName:      item.BankName,
			BankCode:      item.BankCode,
			AccountNumber: item.AccountNumber,
			AccountHolder: item.AccountHolder,
			CurrencyID:    item.CurrencyID,
			Currency:      item.CurrencyCode,
			OwnerType:     item.OwnerType,
			OwnerID:       item.OwnerID,
			OwnerName:     item.OwnerName,
			OwnerCode:     item.OwnerCode,
			IsActive:      item.IsActive,
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     item.UpdatedAt.Format(time.RFC3339),
		}
		if item.CurrencyID != nil {
			decimalPlaces := 0
			if item.CurrencyDecimalPlaces != nil {
				decimalPlaces = *item.CurrencyDecimalPlaces
			}
			response.CurrencyDetail = &dto.CurrencyResponse{
				ID:            *item.CurrencyID,
				Code:          item.CurrencyCode,
				Name:          derefString(item.CurrencyName),
				Symbol:        derefString(item.CurrencySymbol),
				DecimalPlaces: decimalPlaces,
			}
		}
		responses = append(responses, response)
	}
	return responses, total, nil
}

func (u *bankAccountUsecase) GetByID(ctx context.Context, id string) (*dto.BankAccountResponse, error) {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankAccountNotFound
		}
		return nil, err
	}

	res := u.mapper.ToResponse(item)
	history, err := u.repo.ListTransactionHistory(ctx, id, 100)
	if err != nil {
		return nil, err
	}
	if len(history) > 0 {
		res.TransactionHistory = make([]dto.BankAccountTransactionResponse, 0, len(history))
		for _, h := range history {
			res.TransactionHistory = append(res.TransactionHistory, dto.BankAccountTransactionResponse{
				ID:                 h.ID,
				TransactionType:    h.TransactionType,
				TransactionDate:    h.TransactionDate.Format("2006-01-02T15:04:05+07:00"),
				ReferenceType:      h.ReferenceType,
				ReferenceID:        h.ReferenceID,
				ReferenceNumber:    h.ReferenceNumber,
				RelatedEntityType:  h.RelatedEntityType,
				RelatedEntityID:    h.RelatedEntityID,
				RelatedEntityLabel: h.RelatedEntityLabel,
				Amount:             h.Amount,
				Status:             h.Status,
				Description:        h.Description,
			})
		}
	}

	return res, nil
}

func (u *bankAccountUsecase) ListTransactionHistory(ctx context.Context, id string, limit, offset int) ([]dto.BankAccountTransactionResponse, int64, error) {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrBankAccountNotFound
		}
		return nil, 0, err
	}

	history, total, err := u.repo.ListTransactionHistoryPaginated(ctx, id, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.BankAccountTransactionResponse, 0, len(history))
	for _, h := range history {
		responses = append(responses, dto.BankAccountTransactionResponse{
			ID:                 h.ID,
			TransactionType:    h.TransactionType,
			TransactionDate:    h.TransactionDate.Format("2006-01-02T15:04:05+07:00"),
			ReferenceType:      h.ReferenceType,
			ReferenceID:        h.ReferenceID,
			ReferenceNumber:    h.ReferenceNumber,
			RelatedEntityType:  h.RelatedEntityType,
			RelatedEntityID:    h.RelatedEntityID,
			RelatedEntityLabel: h.RelatedEntityLabel,
			Amount:             h.Amount,
			Status:             h.Status,
			Description:        h.Description,
		})
	}

	return responses, total, nil
}

func (u *bankAccountUsecase) Update(ctx context.Context, id string, req *dto.UpdateBankAccountRequest) (*dto.BankAccountResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankAccountNotFound
		}
		return nil, err
	}
	resolvedCurrency, err := u.resolveCurrency(ctx, req.CurrencyID)
	if err != nil {
		return nil, err
	}
	currencyCode := strings.TrimSpace(req.Currency)
	if resolvedCurrency != nil {
		currencyCode = resolvedCurrency.Code
	}
	if currencyCode == "" {
		currencyCode = "IDR"
	}
	item.Name = strings.TrimSpace(req.Name)
	if code := strings.TrimSpace(req.Code); code != "" {
		item.Code = code
	}
	if accountType := strings.TrimSpace(req.AccountType); accountType != "" {
		item.AccountType = accountType
	}
	item.BankID = req.BankID
	item.AccountNumber = strings.TrimSpace(req.AccountNumber)
	item.AccountHolder = strings.TrimSpace(req.AccountHolder)
	item.CurrencyID = req.CurrencyID
	item.CurrencyDetail = resolvedCurrency
	item.Currency = currencyCode
	item.ChartOfAccountID = req.ChartOfAccountID
	item.VillageID = req.VillageID
	item.BankAddress = strings.TrimSpace(req.BankAddress)
	item.BankPhone = strings.TrimSpace(req.BankPhone)
	item.CountryCode = strings.ToUpper(strings.TrimSpace(req.CountryCode))
	item.BankBranchCode = strings.TrimSpace(req.BankBranchCode)
	item.OpeningBalance = req.OpeningBalance
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return u.mapper.ToResponse(item), nil
}

func (u *bankAccountUsecase) Delete(ctx context.Context, id string) error {
	item, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrBankAccountNotFound
		}
		return err
	}
	return u.repo.Delete(ctx, item.ID)
}

// ========== PHASE 2 METHODS ==========

// GetByIDWithBalance retrieves bank account with computed balance from GL
func (u *bankAccountUsecase) GetByIDWithBalance(ctx context.Context, id string) (*dto.BankAccountResponse, error) {
	detail, err := u.repo.FindByIDWithBalance(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankAccountNotFound
		}
		return nil, err
	}

	// Map to response DTO
	res := &dto.BankAccountResponse{
		ID:               detail.BankAccount.ID,
		CompanyID:        detail.BankAccount.CompanyID,
		Code:             detail.BankAccount.Code,
		Name:             detail.BankAccount.Name,
		AccountType:      detail.BankAccount.AccountType,
		BankID:           detail.BankAccount.BankID,
		AccountNumber:    detail.BankAccount.AccountNumber,
		AccountHolder:    detail.BankAccount.AccountHolder,
		Currency:         detail.BankAccount.Currency,
		CurrencyID:       detail.BankAccount.CurrencyID,
		ChartOfAccountID: detail.BankAccount.ChartOfAccountID,
		CountryCode:      detail.BankAccount.CountryCode,
		BankBranchCode:   detail.BankAccount.BankBranchCode,
		OpeningBalance:   detail.BankAccount.OpeningBalance,
		CurrentBalance:   detail.CurrentBalance,
		IsReconcilable:   detail.IsReconcilable,
		IsActive:         detail.BankAccount.IsActive,
		CreatedBy:        detail.BankAccount.CreatedBy,
		UpdatedBy:        detail.BankAccount.UpdatedBy,
		CreatedAt:        detail.BankAccount.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        detail.BankAccount.UpdatedAt.Format(time.RFC3339),
	}

	// Add balance breakdown if available
	if detail.BalanceBreakdown != nil {
		res.BalanceBreakdown = &dto.BankAccountBalanceBreakdown{
			OpeningJournalBalance:  detail.BalanceBreakdown.OpeningJournalBalance,
			TransactionDebitTotal:  detail.BalanceBreakdown.TransactionDebitTotal,
			TransactionCreditTotal: detail.BalanceBreakdown.TransactionCreditTotal,
			CurrentBalance:         detail.BalanceBreakdown.CurrentBalance,
		}
	}

	// Add metadata if available
	if detail.Metadata != nil {
		var lastReconciledAtStr *string
		if detail.Metadata.LastReconciledAt != nil {
			lastReconciledStr := detail.Metadata.LastReconciledAt.Format(time.RFC3339)
			lastReconciledAtStr = &lastReconciledStr
		}
		var statementDateStr *string
		if detail.Metadata.StatementDate != nil {
			statementStr := detail.Metadata.StatementDate.Format(time.RFC3339)
			statementDateStr = &statementStr
		}
		res.Metadata = &dto.BankAccountMetadata{
			LastReconciledAt:     lastReconciledAtStr,
			ReconciliationStatus: detail.Metadata.ReconciliationStatus,
			StatementDate:        statementDateStr,
			BookDifference:       detail.Metadata.BookDifference,
		}
		if detail.Metadata.WarningMessage != nil && strings.TrimSpace(*detail.Metadata.WarningMessage) != "" {
			res.Warning = &dto.BankAccountWarning{
				Type:    "ACCOUNT_CREATED_DURING_OPERATIONS",
				Message: strings.TrimSpace(*detail.Metadata.WarningMessage),
				Level:   "info",
			}
		}
	}

	if len(detail.RecentTransactions) > 0 {
		res.RecentTransactions = make([]dto.BankAccountTransaction, 0, len(detail.RecentTransactions))
		for _, tx := range detail.RecentTransactions {
			res.RecentTransactions = append(res.RecentTransactions, dto.BankAccountTransaction{
				ID:              tx.ID,
				ReferenceNumber: derefString(tx.ReferenceNumber),
				Type:            tx.TransactionType,
				Date:            tx.TransactionDate.Format(time.RFC3339),
				Amount:          tx.Amount,
				Status:          tx.Status,
				Description:     tx.Description,
			})
		}
	}

	bankDetail, err := u.loadBankMaster(ctx, detail.BankAccount.BankID)
	if err == nil && bankDetail != nil {
		res.BankDetail = &dto.BankMasterResponse{
			ID:        bankDetail.ID,
			Code:      bankDetail.Code,
			Name:      bankDetail.Name,
			SwiftCode: bankDetail.SwiftCode,
		}
	}

	return res, nil
}

// ListByCompanyWithBalance retrieves bank accounts for a company with computed balances
func (u *bankAccountUsecase) ListByCompanyWithBalance(ctx context.Context, companyID string, params repositories.BankAccountListParams) ([]*dto.BankAccountResponse, int64, error) {
	params.CompanyID = companyID

	details, total, err := u.repo.ListByCompanyWithBalance(ctx, companyID, params)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.BankAccountResponse, 0, len(details))
	for _, detail := range details {
		res := &dto.BankAccountResponse{
			ID:               detail.BankAccount.ID,
			CompanyID:        detail.BankAccount.CompanyID,
			Code:             detail.BankAccount.Code,
			Name:             detail.BankAccount.Name,
			AccountType:      detail.BankAccount.AccountType,
			BankID:           detail.BankAccount.BankID,
			AccountNumber:    detail.BankAccount.AccountNumber,
			AccountHolder:    detail.BankAccount.AccountHolder,
			Currency:         detail.BankAccount.Currency,
			CurrencyID:       detail.BankAccount.CurrencyID,
			ChartOfAccountID: detail.BankAccount.ChartOfAccountID,
			CountryCode:      detail.BankAccount.CountryCode,
			BankBranchCode:   detail.BankAccount.BankBranchCode,
			OpeningBalance:   detail.BankAccount.OpeningBalance,
			CurrentBalance:   detail.CurrentBalance,
			IsReconcilable:   detail.IsReconcilable,
			IsActive:         detail.BankAccount.IsActive,
			CreatedBy:        detail.BankAccount.CreatedBy,
			UpdatedBy:        detail.BankAccount.UpdatedBy,
			CreatedAt:        detail.BankAccount.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        detail.BankAccount.UpdatedAt.Format(time.RFC3339),
		}

		if detail.Metadata != nil {
			res.Metadata = &dto.BankAccountMetadata{
				ReconciliationStatus: detail.Metadata.ReconciliationStatus,
			}
		}

		bankDetail, err := u.loadBankMaster(ctx, detail.BankAccount.BankID)
		if err == nil && bankDetail != nil {
			res.BankDetail = &dto.BankMasterResponse{
				ID:        bankDetail.ID,
				Code:      bankDetail.Code,
				Name:      bankDetail.Name,
				SwiftCode: bankDetail.SwiftCode,
			}
		}

		responses = append(responses, res)
	}

	return responses, total, nil
}

// ToggleStatus toggles the is_active flag of a bank account
func (u *bankAccountUsecase) ToggleStatus(ctx context.Context, bankAccountID string) (*dto.ToggleStatusResponse, error) {
	// Verify account exists
	_, err := u.repo.FindByID(ctx, bankAccountID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankAccountNotFound
		}
		return nil, err
	}

	// Toggle status
	if err := u.repo.ToggleStatus(ctx, bankAccountID); err != nil {
		return nil, err
	}

	// Fetch updated account
	updated, err := u.repo.FindByID(ctx, bankAccountID)
	if err != nil {
		return nil, err
	}

	return &dto.ToggleStatusResponse{
		ID:       updated.ID,
		IsActive: updated.IsActive,
	}, nil
}

func (u *bankAccountUsecase) loadBankMaster(ctx context.Context, bankID *string) (*supplierModels.Bank, error) {
	if bankID == nil || strings.TrimSpace(*bankID) == "" {
		return nil, nil
	}

	var bank supplierModels.Bank
	if err := u.db.WithContext(ctx).First(&bank, "id = ?", strings.TrimSpace(*bankID)).Error; err != nil {
		return nil, err
	}

	return &bank, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
