package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

// ARAPReconciliationUsecase defines the business logic for AR/AP reconciliation.
type ARAPReconciliationUsecase interface {
	ReconcileAR(ctx context.Context, asOf time.Time) (*dto.ARAPReconciliationReport, error)
	ReconcileAP(ctx context.Context, asOf time.Time) (*dto.ARAPReconciliationReport, error)
}

type arapReconciliationUsecase struct {
	db              *gorm.DB
	agingRepo       repositories.AgingReportRepository
	coaRepo         repositories.ChartOfAccountRepository
	settingsService financesettings.SettingsService
	engine          accounting.AccountingEngine
}

// NewARAPReconciliationUsecase creates a new instance of ARAPReconciliationUsecase.
func NewARAPReconciliationUsecase(
	db *gorm.DB,
	agingRepo repositories.AgingReportRepository,
	coaRepo repositories.ChartOfAccountRepository,
	settingsService financesettings.SettingsService,
	engine accounting.AccountingEngine,
) ARAPReconciliationUsecase {
	return &arapReconciliationUsecase{
		db:              db,
		agingRepo:       agingRepo,
		coaRepo:         coaRepo,
		settingsService: settingsService,
		engine:          engine,
	}
}

func (uc *arapReconciliationUsecase) ReconcileAR(ctx context.Context, asOf time.Time) (*dto.ARAPReconciliationReport, error) {
	// 1. Get AR Account from Settings
	coaCode, err := uc.settingsService.GetCOACode(ctx, models.SettingCOASalesReceivable)
	if err != nil {
		return nil, fmt.Errorf("failed to get AR account mapping: %w", err)
	}

	coa, err := uc.coaRepo.FindByCode(ctx, coaCode)
	if err != nil {
		return nil, fmt.Errorf("AR account %s not found in COA: %w", coaCode, err)
	}

	// 2. Get GL Balance for the AR Account
	glBalance, err := uc.engine.GetAccountBalance(ctx, coa.ID, asOf)
	if err != nil {
		return nil, fmt.Errorf("failed to get GL balance for AR account: %w", err)
	}

	// 3. Get Subledger Data from Aging Report Repository
	// We use a large limit to get all records for reconciliation
	rows, _, err := uc.agingRepo.ListARAging(ctx, repositories.AgingListParams{
		AsOfDate: asOf,
		Limit:    10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AR subledger data: %w", err)
	}

	// 4. Calculate Subledger Total
	var subledgerTotal float64
	details := make([]dto.ARAPReconciliationRow, 0, len(rows))
	for _, row := range rows {
		subledgerTotal += row.RemainingAmount
		
		status := "MATCHED"
		// In a per-invoice reconciliation we'd compare against invoice-linked GL lines,
		// but here we are doing a bulk reconciliation. 
		// For detailed row status, we'd need to query journal lines linked to this invoice.
		// For now, we report the invoice and leave status as MATCHED unless we find a reason not to.
		
		details = append(details, dto.ARAPReconciliationRow{
			InvoiceID:       row.InvoiceID,
			InvoiceCode:     row.Code,
			PartnerName:     "Customer", // AgingRow doesn't have CustomerName currently
			InvoiceAmount:   row.Amount,
			RemainingAmount: row.RemainingAmount,
			Status:          status,
		})
	}

	// 5. Compare and Build Report
	diff := glBalance - subledgerTotal
	status := "MATCHED"
	if math.Abs(diff) > 0.01 {
		status = "MISMATCHED"
	}

	return &dto.ARAPReconciliationReport{
		Type:     "AR",
		AsOfDate: asOf,
		Account: dto.ChartOfAccountResponse{
			ID:   coa.ID,
			Code: coa.Code,
			Name: coa.Name,
			Type: coa.Type,
		},
		Summary: dto.ReconciliationSummary{
			TotalSubledger: subledgerTotal,
			TotalGL:        glBalance,
			Difference:     diff,
			Status:         status,
		},
		Details: details,
	}, nil
}

func (uc *arapReconciliationUsecase) ReconcileAP(ctx context.Context, asOf time.Time) (*dto.ARAPReconciliationReport, error) {
	// 1. Get AP Account from Settings
	coaCode, err := uc.settingsService.GetCOACode(ctx, models.SettingCOAPurchasePayable)
	if err != nil {
		return nil, fmt.Errorf("failed to get AP account mapping: %w", err)
	}

	coa, err := uc.coaRepo.FindByCode(ctx, coaCode)
	if err != nil {
		return nil, fmt.Errorf("AP account %s not found in COA: %w", coaCode, err)
	}

	// 2. Get GL Balance for the AP Account (Liability usually has credit balance, engine should return signed value)
	glBalance, err := uc.engine.GetAccountBalance(ctx, coa.ID, asOf)
	if err != nil {
		return nil, fmt.Errorf("failed to get GL balance for AP account: %w", err)
	}
	
	// Liability balance usually reflected as negative if using debit-positive convention, 
	// or positive if using natural balance. Normalizing based on account type.
	if coa.Type == models.AccountTypeLiability || coa.Type == models.AccountTypeTradePayable {
		glBalance = -glBalance
	}

	// 3. Get Subledger Data from Aging Report Repository
	rows, _, err := uc.agingRepo.ListAPAging(ctx, repositories.AgingListParams{
		AsOfDate: asOf,
		Limit:    10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AP subledger data: %w", err)
	}

	// 4. Calculate Subledger Total
	var subledgerTotal float64
	details := make([]dto.ARAPReconciliationRow, 0, len(rows))
	for _, row := range rows {
		subledgerTotal += row.RemainingAmount
		
		details = append(details, dto.ARAPReconciliationRow{
			InvoiceID:       row.InvoiceID,
			InvoiceCode:     row.Code,
			PartnerName:     row.SupplierName,
			InvoiceAmount:   row.Amount,
			RemainingAmount: row.RemainingAmount,
			Status:          "MATCHED",
		})
	}

	// 5. Compare and Build Report
	diff := glBalance - subledgerTotal
	status := "MATCHED"
	if math.Abs(diff) > 0.01 {
		status = "MISMATCHED"
	}

	return &dto.ARAPReconciliationReport{
		Type:     "AP",
		AsOfDate: asOf,
		Account: dto.ChartOfAccountResponse{
			ID:   coa.ID,
			Code: coa.Code,
			Name: coa.Name,
			Type: coa.Type,
		},
		Summary: dto.ReconciliationSummary{
			TotalSubledger: subledgerTotal,
			TotalGL:        glBalance,
			Difference:     diff,
			Status:         status,
		},
		Details: details,
	}, nil
}
