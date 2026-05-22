package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/google/uuid"
)

type FinancialReportEngine interface {
	GetGeneralLedger(ctx context.Context, params dto.GLParams) (*dto.GeneralLedgerResponse, error)
	GetTrialBalance(ctx context.Context, params dto.TBParams) (*dto.TrialBalanceResponse, error)
	GetBalanceSheet(ctx context.Context, params dto.BSParams) (*dto.BalanceSheetResponse, error)
	GetProfitLoss(ctx context.Context, params dto.PLParams) (*dto.ProfitAndLossResponse, error)
	GetCashFlow(ctx context.Context, params dto.CFParams) (*dto.CashFlowReport, error)
}

// GetAccountBalances is a shared report helper that retrieves account balances in one grouped query.
func (uc *financeReportUsecase) GetAccountBalances(
	ctx context.Context,
	companyID uuid.UUID,
	fiscalYearID uuid.UUID,
	accountIDs []uuid.UUID,
	fromDate, toDate time.Time,
) (map[uuid.UUID]dto.AccountBalance, error) {
	if companyID == uuid.Nil {
		return nil, errors.New("company_id is required")
	}
	if fiscalYearID == uuid.Nil {
		return nil, errors.New("fiscal_year_id is required")
	}

	accountIDStr := make([]string, 0, len(accountIDs))
	for _, id := range accountIDs {
		if id == uuid.Nil {
			continue
		}
		accountIDStr = append(accountIDStr, id.String())
	}

	companyIDStr := companyID.String()
	fiscalYearIDStr := fiscalYearID.String()
	raw, err := uc.reportRepo.GetAccountBalancesByAccounts(ctx, accountIDStr, fromDate, toDate, &companyIDStr, &fiscalYearIDStr)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]dto.AccountBalance, len(raw))
	for id, row := range raw {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		result[parsedID] = dto.AccountBalance{
			OpeningBalance: row.OpeningBalance,
			DebitTotal:     row.DebitTotal,
			CreditTotal:    row.CreditTotal,
			ClosingBalance: row.ClosingBalance,
		}
	}

	for _, id := range accountIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := result[id]; !ok {
			result[id] = dto.AccountBalance{}
		}
	}

	return result, nil
}
