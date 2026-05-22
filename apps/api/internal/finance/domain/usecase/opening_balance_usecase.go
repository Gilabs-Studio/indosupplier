package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrOpeningBalanceNotBalanced   = errors.New("opening balance is not balanced")
	ErrOpeningBalanceAlreadyPosted = errors.New("opening balance already posted for fiscal year")
	ErrOpeningBalanceNoLines       = errors.New("opening balance staging lines not found")
	ErrFiscalYearNotActive         = errors.New("fiscal year must be active before posting opening balance")
)

type OpeningBalanceUsecase interface {
	Get(ctx context.Context, companyID, fiscalYearID string) ([]dto.OpeningBalanceLineResponse, error)
	UpsertDraft(ctx context.Context, req *dto.UpsertOpeningBalanceRequest) error
	Validate(ctx context.Context, req *dto.ValidateOpeningBalanceRequest) (*dto.OpeningBalanceValidationResponse, error)
	Simulate(ctx context.Context, req *dto.ValidateOpeningBalanceRequest) (*dto.OpeningBalanceSimulationResponse, error)
	Summary(ctx context.Context, companyID, fiscalYearID string) (*dto.OpeningBalanceSummaryResponse, error)
	Post(ctx context.Context, req *dto.PostOpeningBalanceRequest, actorID string) (*dto.PostOpeningBalanceResponse, error)
}

type openingBalanceUsecase struct {
	repo                repositories.OpeningBalanceRepository
	fiscalYearRepo      repositories.FiscalYearRepository
	coaRepo             repositories.ChartOfAccountRepository
	inventorySettingsUC InventorySettingsUsecase
}

func NewOpeningBalanceUsecase(
	repo repositories.OpeningBalanceRepository,
	fiscalYearRepo repositories.FiscalYearRepository,
	coaRepo repositories.ChartOfAccountRepository,
	inventorySettingsUC InventorySettingsUsecase,
) OpeningBalanceUsecase {
	return &openingBalanceUsecase{
		repo:                repo,
		fiscalYearRepo:      fiscalYearRepo,
		coaRepo:             coaRepo,
		inventorySettingsUC: inventorySettingsUC,
	}
}

func (uc *openingBalanceUsecase) Get(ctx context.Context, companyID, fiscalYearID string) ([]dto.OpeningBalanceLineResponse, error) {
	lines, err := uc.repo.ListLines(ctx, companyID, fiscalYearID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.OpeningBalanceLineResponse, 0, len(lines))
	for _, line := range lines {
		res = append(res, dto.OpeningBalanceLineResponse{
			ID:             line.ID,
			AccountID:      line.AccountID,
			DebitAmount:    line.DebitAmount,
			CreditAmount:   line.CreditAmount,
			Description:    line.Description,
			ProductID:      line.ProductID,
			ProductQty:     line.ProductQty,
			ProductAvgCost: line.ProductAvgCost,
			UpdatedAt:      line.UpdatedAt,
		})
	}
	return res, nil
}

func (uc *openingBalanceUsecase) UpsertDraft(ctx context.Context, req *dto.UpsertOpeningBalanceRequest) error {
	items := make([]financeModels.OpeningBalanceLine, 0, len(req.Lines))
	for _, line := range req.Lines {
		if line.DebitAmount > 0 && line.CreditAmount > 0 {
			return errors.New("opening balance line cannot have both debit and credit")
		}
		if line.DebitAmount == 0 && line.CreditAmount == 0 {
			continue
		}
		items = append(items, financeModels.OpeningBalanceLine{
			CompanyID:      req.CompanyID,
			FiscalYearID:   req.FiscalYearID,
			AccountID:      strings.TrimSpace(line.AccountID),
			DebitAmount:    line.DebitAmount,
			CreditAmount:   line.CreditAmount,
			Description:    strings.TrimSpace(line.Description),
			ProductID:      line.ProductID,
			ProductQty:     line.ProductQty,
			ProductAvgCost: line.ProductAvgCost,
		})
	}
	return uc.repo.ReplaceLines(ctx, req.CompanyID, req.FiscalYearID, items)
}

func (uc *openingBalanceUsecase) Validate(ctx context.Context, req *dto.ValidateOpeningBalanceRequest) (*dto.OpeningBalanceValidationResponse, error) {
	lines, err := uc.repo.ListLines(ctx, req.CompanyID, req.FiscalYearID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return &dto.OpeningBalanceValidationResponse{Warnings: []string{"No opening balance lines found"}}, nil
	}

	fy, err := uc.fiscalYearRepo.FindByID(ctx, req.FiscalYearID)
	if err != nil {
		return nil, err
	}

	hasExistingTxns, err := uc.repo.HasPostedOperationalJournalInRange(ctx, fy.StartDate.Format("2006-01-02"), fy.EndDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	totalDebit, totalCredit := sumOpeningLines(lines)
	diff := round2OB(totalDebit - totalCredit)
	isBalanced := math.Abs(diff) < 0.005
	warnings := make([]string, 0)
	if hasExistingTxns {
		warnings = append(warnings, "Operational posted journals already exist in this fiscal year period")
	}
	if !isBalanced {
		warnings = append(warnings, fmt.Sprintf("Opening balance difference detected: %.2f", diff))
	}

	return &dto.OpeningBalanceValidationResponse{
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		Difference:      diff,
		IsBalanced:      isBalanced,
		HasExistingTxns: hasExistingTxns,
		Warnings:        warnings,
	}, nil
}

func (uc *openingBalanceUsecase) Summary(ctx context.Context, companyID, fiscalYearID string) (*dto.OpeningBalanceSummaryResponse, error) {
	lines, err := uc.repo.ListLines(ctx, companyID, fiscalYearID)
	if err != nil {
		return nil, err
	}
	postedJournalID, err := uc.repo.GetPostedOpeningJournalID(ctx, companyID, fiscalYearID)
	if err != nil {
		return nil, err
	}
	totalDebit, totalCredit := sumOpeningLines(lines)
	inventoryLines := 0
	for _, line := range lines {
		if line.ProductID != nil {
			inventoryLines++
		}
	}
	diff := round2OB(totalDebit - totalCredit)
	return &dto.OpeningBalanceSummaryResponse{
		CompanyID:       companyID,
		FiscalYearID:    fiscalYearID,
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		Difference:      diff,
		IsBalanced:      math.Abs(diff) < 0.005,
		IsPosted:        postedJournalID != nil,
		PostedJournalID: postedJournalID,
		TotalLines:      len(lines),
		InventoryLines:  inventoryLines,
	}, nil
}

func (uc *openingBalanceUsecase) Simulate(ctx context.Context, req *dto.ValidateOpeningBalanceRequest) (*dto.OpeningBalanceSimulationResponse, error) {
	lines, err := uc.repo.ListLines(ctx, req.CompanyID, req.FiscalYearID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.OpeningBalanceSimulationLineResponse, 0, len(lines))
	var warningCount int
	for _, line := range lines {
		status := "READY"
		action := "Line is valid for posting"

		coa, coaErr := uc.coaRepo.FindByID(ctx, line.AccountID)
		if coaErr != nil {
			status = "ACTION_REQUIRED"
			action = "Chart of account not found"
			warningCount++
			responses = append(responses, dto.OpeningBalanceSimulationLineResponse{
				AccountID:    line.AccountID,
				AccountCode:  "",
				AccountName:  "Unknown account",
				DebitAmount:  line.DebitAmount,
				CreditAmount: line.CreditAmount,
				Status:       status,
				Action:       action,
			})
			continue
		}

		if !coa.IsActive || !coa.IsPostable {
			status = "ACTION_REQUIRED"
			action = "Use an active and postable account"
			warningCount++
		}
		if line.DebitAmount > 0 && line.CreditAmount > 0 {
			status = "ACTION_REQUIRED"
			action = "Use either debit or credit, not both"
			warningCount++
		}
		if line.DebitAmount == 0 && line.CreditAmount == 0 {
			status = "ACTION_REQUIRED"
			action = "Provide debit or credit amount"
			warningCount++
		}

		responses = append(responses, dto.OpeningBalanceSimulationLineResponse{
			AccountID:    line.AccountID,
			AccountCode:  coa.Code,
			AccountName:  coa.Name,
			DebitAmount:  line.DebitAmount,
			CreditAmount: line.CreditAmount,
			Status:       status,
			Action:       action,
		})
	}

	totalDebit, totalCredit := sumOpeningLines(lines)
	difference := round2OB(totalDebit - totalCredit)
	isBalanced := math.Abs(difference) < 0.005

	valuationStatus := "READY_TO_POST"
	recommendation := "Opening balance is ready to post"
	if !isBalanced || warningCount > 0 {
		valuationStatus = "REVIEW_REQUIRED"
		recommendation = "Fix line issues and ensure total debit equals total credit"
	}

	return &dto.OpeningBalanceSimulationResponse{
		CompanyID:       req.CompanyID,
		FiscalYearID:    req.FiscalYearID,
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		Difference:      difference,
		IsBalanced:      isBalanced,
		ValuationStatus: valuationStatus,
		Recommendation:  recommendation,
		Lines:           responses,
	}, nil
}

func (uc *openingBalanceUsecase) Post(ctx context.Context, req *dto.PostOpeningBalanceRequest, actorID string) (*dto.PostOpeningBalanceResponse, error) {
	actorID = strings.TrimSpace(actorID)
	var actorIDPtr *string
	if actorID != "" {
		actorIDPtr = &actorID
	}

	companyID := strings.TrimSpace(req.CompanyID)
	fiscalYearID := strings.TrimSpace(req.FiscalYearID)
	if companyID == "" || fiscalYearID == "" {
		return nil, errors.New("company_id and fiscal_year_id are required")
	}
	fiscalYearIDPtr := &fiscalYearID

	validation, err := uc.Validate(ctx, &dto.ValidateOpeningBalanceRequest{
		CompanyID:    req.CompanyID,
		FiscalYearID: req.FiscalYearID,
	})
	if err != nil {
		return nil, err
	}
	if !validation.IsBalanced {
		return nil, ErrOpeningBalanceNotBalanced
	}

	lines, err := uc.repo.ListLines(ctx, req.CompanyID, req.FiscalYearID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrOpeningBalanceNoLines
	}
	if err := uc.validatePostableLines(ctx, lines); err != nil {
		return nil, err
	}

	journalID := ""
	postedAt := apptime.Now()
	referenceType := string(financeModels.RefOpeningBalance)
	referenceID := req.CompanyID + ":" + req.FiscalYearID

	err = database.GetDB(ctx, uc.repo.GetDB(ctx)).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", referenceID).Error; err != nil {
			return err
		}

		txCtx := context.WithValue(ctx, "tx", tx)
		fy, fyErr := uc.fiscalYearRepo.FindByID(txCtx, req.FiscalYearID)
		if fyErr != nil {
			return fyErr
		}
		if fy.Status != financeModels.FiscalYearStatusActive {
			return ErrFiscalYearNotActive
		}

		alreadyPosted, postedErr := uc.repo.HasPostedOpeningJournal(txCtx, req.CompanyID, req.FiscalYearID)
		if postedErr != nil {
			return postedErr
		}
		if alreadyPosted {
			return ErrOpeningBalanceAlreadyPosted
		}

		entry := &financeModels.JournalEntry{
			CompanyID:         companyID,
			FiscalYearID:      fiscalYearIDPtr,
			EntryDate:         fy.StartDate,
			Description:       "Opening balance posting",
			ReferenceType:     &referenceType,
			ReferenceID:       &referenceID,
			Status:            financeModels.JournalStatusPosted,
			JournalType:       financeModels.JournalTypeOpeningBalance,
			PostedBy:          actorIDPtr,
			PostedAt:          &postedAt,
			CreatedBy:         actorIDPtr,
			IsSystemGenerated: true,
			DebitTotal:        validation.TotalDebit,
			CreditTotal:       validation.TotalCredit,
		}
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		journalID = entry.ID

		for _, line := range lines {
			coa, coaErr := uc.coaRepo.FindByID(txCtx, line.AccountID)
			if coaErr != nil {
				return coaErr
			}
			journalLine := &financeModels.JournalLine{
				JournalEntryID:             entry.ID,
				ChartOfAccountID:           line.AccountID,
				ChartOfAccountCodeSnapshot: strings.TrimSpace(coa.Code),
				ChartOfAccountNameSnapshot: strings.TrimSpace(coa.Name),
				ChartOfAccountTypeSnapshot: string(coa.Type),
				Debit:                      line.DebitAmount,
				Credit:                     line.CreditAmount,
				Memo:                       strings.TrimSpace(line.Description),
			}
			if err := tx.Create(journalLine).Error; err != nil {
				return err
			}

			if line.ProductID != nil && line.ProductQty != nil && line.ProductAvgCost != nil {
				_, recErr := uc.inventorySettingsUC.RecalculateOnReceive(txCtx, req.CompanyID, *line.ProductID, *line.ProductQty, *line.ProductAvgCost)
				if recErr != nil {
					return recErr
				}
			}

			var linkedBankAccounts []coreModels.BankAccount
			if err := tx.WithContext(txCtx).
				Where("chart_of_account_id = ?", line.AccountID).
				Where("is_active = ?", true).
				Find(&linkedBankAccounts).Error; err != nil {
				return err
			}

			for _, bankAccount := range linkedBankAccounts {
				entryType := financeModels.CashBankTypeCashIn
				amount := line.DebitAmount
				if amount <= 0 {
					entryType = financeModels.CashBankTypeCashOut
					amount = line.CreditAmount
				}
				if amount <= 0 {
					continue
				}

				cashBankJournal := &financeModels.CashBankJournal{
					TransactionDate:             entry.EntryDate,
					Type:                        entryType,
					Description:                 "Opening balance bank impact: " + strings.TrimSpace(line.Description),
					BankAccountID:               bankAccount.ID,
					BankAccountNameSnapshot:     bankAccount.Name,
					BankAccountNumberSnapshot:   bankAccount.AccountNumber,
					BankAccountHolderSnapshot:   bankAccount.AccountHolder,
					BankAccountCurrencySnapshot: bankAccount.Currency,
					TotalAmount:                 amount,
					Status:                      financeModels.CashBankStatusPosted,
					JournalEntryID:              &entry.ID,
					PostedAt:                    &postedAt,
					PostedBy:                    actorIDPtr,
					CreatedBy:                   actorIDPtr,
				}
				if err := tx.Create(cashBankJournal).Error; err != nil {
					return err
				}
			}
		}

		if err := uc.repo.DeleteLines(txCtx, req.CompanyID, req.FiscalYearID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.PostOpeningBalanceResponse{
		JournalID:     journalID,
		JournalStatus: string(financeModels.JournalStatusPosted),
		JournalType:   string(financeModels.JournalTypeOpeningBalance),
		PostedAt:      postedAt.Format(timeLayoutRFC3339),
	}, nil
}

func (uc *openingBalanceUsecase) validatePostableLines(ctx context.Context, lines []financeModels.OpeningBalanceLine) error {
	for index, line := range lines {
		lineNumber := index + 1
		if strings.TrimSpace(line.AccountID) == "" {
			return fmt.Errorf("opening balance line %d: account_id is required", lineNumber)
		}
		if line.DebitAmount > 0 && line.CreditAmount > 0 {
			return fmt.Errorf("opening balance line %d: cannot have both debit and credit", lineNumber)
		}
		if line.DebitAmount == 0 && line.CreditAmount == 0 {
			return fmt.Errorf("opening balance line %d: debit or credit amount is required", lineNumber)
		}

		coa, err := uc.coaRepo.FindByID(ctx, line.AccountID)
		if err != nil {
			return fmt.Errorf("opening balance line %d: chart of account not found", lineNumber)
		}
		if !coa.IsActive || !coa.IsPostable {
			return fmt.Errorf("opening balance line %d: account must be active and postable", lineNumber)
		}
	}
	return nil
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"

func sumOpeningLines(lines []financeModels.OpeningBalanceLine) (float64, float64) {
	totalDebit := 0.0
	totalCredit := 0.0
	for _, line := range lines {
		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
	}
	return round2OB(totalDebit), round2OB(totalCredit)
}

func round2OB(value float64) float64 {
	return math.Round(value*100) / 100
}
