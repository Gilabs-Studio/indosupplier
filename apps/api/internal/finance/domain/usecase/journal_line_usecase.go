package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

// JournalLineUsecase handles operations for the journal lines sub-ledger view.
type JournalLineUsecase interface {
	ListLines(ctx context.Context, req *dto.ListJournalLinesRequest) (*dto.ListJournalLinesResponse, int64, error)
	ExportLinesCSV(ctx context.Context, req *dto.ListJournalLinesRequest, writer io.Writer) error
}

type journalLineUsecase struct {
	lineRepo repositories.JournalLineRepository
}

// NewJournalLineUsecase creates a new JournalLineUsecase.
func NewJournalLineUsecase(lineRepo repositories.JournalLineRepository) JournalLineUsecase {
	return &journalLineUsecase{lineRepo: lineRepo}
}

// isDebitNormal returns true if the account type has a normal debit balance.
// Debit-normal: ASSET, CASH_BANK, CURRENT_ASSET, FIXED_ASSET, EXPENSE, COST_OF_GOODS_SOLD, SALARY_WAGES, OPERATIONAL
// Credit-normal: LIABILITY, TRADE_PAYABLE, EQUITY, REVENUE
func isDebitNormal(accountType string) bool {
	switch financeModels.AccountType(accountType) {
	case financeModels.AccountTypeAsset,
		financeModels.AccountTypeCashBank,
		financeModels.AccountTypeCurrentAsset,
		financeModels.AccountTypeFixedAsset,
		financeModels.AccountTypeExpense,
		financeModels.AccountTypeCOGS,
		financeModels.AccountTypeSalaryWages,
		financeModels.AccountTypeOperational:
		return true
	default:
		return false
	}
}

func (uc *journalLineUsecase) ListLines(ctx context.Context, req *dto.ListJournalLinesRequest) (*dto.ListJournalLinesResponse, int64, error) {
	if req == nil {
		req = &dto.ListJournalLinesRequest{}
	}

	page, perPage := normalizePagination(req.Page, req.PerPage)

	startDate, err := parseDateOptional(req.StartDate)
	if err != nil {
		return nil, 0, err
	}
	endDate, err := parseDateOptional(req.EndDate)
	if err != nil {
		return nil, 0, err
	}

	params := repositories.JournalLineListParams{
		CashBankJournalID: strings.TrimSpace(req.CashBankJournalID),
		ChartOfAccountID:  strings.TrimSpace(req.ChartOfAccountID),
		AccountType:       strings.TrimSpace(req.AccountType),
		ReferenceType:     req.ReferenceType,
		JournalStatus:     strings.TrimSpace(req.JournalStatus),
		StartDate:         startDate,
		EndDate:           endDate,
		Search:            strings.TrimSpace(req.Search),
		SortBy:            req.SortBy,
		SortDir:           req.SortDir,
		Limit:             perPage,
		Offset:            (page - 1) * perPage,
	}

	items, total, err := uc.lineRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Determine if we should calculate running balance (single COA filter only)
	coaID := strings.TrimSpace(req.ChartOfAccountID)
	hasRunningBalance := coaID != ""

	var openingDebit, openingCredit float64
	var accountType string

	if hasRunningBalance && len(items) > 0 {
		// Get account type from first item snapshot
		accountType = items[0].ChartOfAccountTypeSnapshot

		// Calculate opening balance: sum of all debits/credits before the first item's date
		// For paginated results, we need to calculate from the page offset
		if params.Offset == 0 && startDate != nil {
			// For first page with date filter: get opening balance before start date
			openingDebit, openingCredit, err = uc.lineRepo.SumBeforeDate(ctx, coaID, *startDate, params.JournalStatus)
			if err != nil {
				return nil, 0, err
			}
		} else if params.Offset > 0 {
			// For subsequent pages: we need to get sum of all lines before this page
			// Query all lines before this page's offset to get cumulative balance
			allParams := params
			allParams.Limit = params.Offset
			allParams.Offset = 0

			priorItems, _, err := uc.lineRepo.List(ctx, allParams)
			if err != nil {
				return nil, 0, err
			}

			// Add the opening balance before start date if we have a date filter
			if startDate != nil {
				openingDebit, openingCredit, err = uc.lineRepo.SumBeforeDate(ctx, coaID, *startDate, params.JournalStatus)
				if err != nil {
					return nil, 0, err
				}
			}

			// Add all prior page items
			for _, item := range priorItems {
				openingDebit += item.Debit
				openingCredit += item.Credit
			}
		} else if startDate == nil && params.Offset == 0 {
			// First page, no date filter — opening balance starts at 0
			openingDebit = 0
			openingCredit = 0
		}
	}

	// Build response
	lines := make([]dto.JournalLineDetailResponse, 0, len(items))
	var totalDebit, totalCredit float64
	var cumulativeDebit, cumulativeCredit float64

	cumulativeDebit = openingDebit
	cumulativeCredit = openingCredit

	for _, item := range items {
		totalDebit += item.Debit
		totalCredit += item.Credit

		var runningBalance float64
		if hasRunningBalance {
			cumulativeDebit += item.Debit
			cumulativeCredit += item.Credit

			if isDebitNormal(accountType) {
				runningBalance = cumulativeDebit - cumulativeCredit
			} else {
				runningBalance = cumulativeCredit - cumulativeDebit
			}
		}

		line := dto.JournalLineDetailResponse{
			ID:                 item.ID,
			JournalEntryID:     item.JournalEntryID,
			EntryDate:          item.EntryDate.Format("2006-01-02"),
			JournalDescription: item.JournalDescription,
			JournalStatus:      string(item.JournalStatus),
			ReferenceType:      item.ReferenceType,
			ReferenceID:        item.ReferenceID,
			ChartOfAccountID:   item.ChartOfAccountID,
			ChartOfAccountCode: item.ChartOfAccountCodeSnapshot,
			ChartOfAccountName: item.ChartOfAccountNameSnapshot,
			ChartOfAccountType: item.ChartOfAccountTypeSnapshot,
			Debit:              item.Debit,
			Credit:             item.Credit,
			Memo:               item.Memo,
			RunningBalance:     runningBalance,
			CreatedAt:          item.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		}
		lines = append(lines, line)
	}

	resp := &dto.ListJournalLinesResponse{
		Lines:       lines,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
	}

	return resp, total, nil
}

func (uc *journalLineUsecase) ExportLinesCSV(ctx context.Context, req *dto.ListJournalLinesRequest, writer io.Writer) error {
	// Remove pagination for export — fetch all matching lines
	exportReq := *req
	exportReq.Page = 0
	exportReq.PerPage = 0

	startDate, err := parseDateOptional(req.StartDate)
	if err != nil {
		return err
	}
	endDate, err := parseDateOptional(req.EndDate)
	if err != nil {
		return err
	}

	params := repositories.JournalLineListParams{
		CashBankJournalID: strings.TrimSpace(req.CashBankJournalID),
		ChartOfAccountID:  strings.TrimSpace(req.ChartOfAccountID),
		AccountType:       strings.TrimSpace(req.AccountType),
		ReferenceType:     req.ReferenceType,
		JournalStatus:     strings.TrimSpace(req.JournalStatus),
		StartDate:         startDate,
		EndDate:           endDate,
		Search:            strings.TrimSpace(req.Search),
		SortBy:            req.SortBy,
		SortDir:           req.SortDir,
		Limit:             10000, // Safety limit for export
		Offset:            0,
	}

	items, _, err := uc.lineRepo.List(ctx, params)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"Entry Date", "Journal Description", "Status", "Reference Type",
		"COA Code", "COA Name", "COA Type", "Memo",
		"Debit", "Credit", "Running Balance",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Calculate running balance if single COA
	coaID := strings.TrimSpace(req.ChartOfAccountID)
	hasRunningBalance := coaID != ""

	var openingDebit, openingCredit float64
	var accountType string

	if hasRunningBalance && len(items) > 0 {
		accountType = items[0].ChartOfAccountTypeSnapshot
		if startDate != nil {
			openingDebit, openingCredit, err = uc.lineRepo.SumBeforeDate(ctx, coaID, *startDate, params.JournalStatus)
			if err != nil {
				return err
			}
		}
	}

	cumulativeDebit := openingDebit
	cumulativeCredit := openingCredit

	for _, item := range items {
		var runningBalance float64
		if hasRunningBalance {
			cumulativeDebit += item.Debit
			cumulativeCredit += item.Credit
			if isDebitNormal(accountType) {
				runningBalance = cumulativeDebit - cumulativeCredit
			} else {
				runningBalance = cumulativeCredit - cumulativeDebit
			}
		}

		refType := ""
		if item.ReferenceType != nil {
			refType = *item.ReferenceType
		}

		rbStr := ""
		if hasRunningBalance {
			rbStr = fmt.Sprintf("%.2f", runningBalance)
		}

		row := []string{
			item.EntryDate.Format("2006-01-02"),
			item.JournalDescription,
			string(item.JournalStatus),
			refType,
			item.ChartOfAccountCodeSnapshot,
			item.ChartOfAccountNameSnapshot,
			item.ChartOfAccountTypeSnapshot,
			item.Memo,
			fmt.Sprintf("%.2f", item.Debit),
			fmt.Sprintf("%.2f", item.Credit),
			rbStr,
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
