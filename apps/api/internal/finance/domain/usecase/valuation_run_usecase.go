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
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	inventoryModels "github.com/gilabs/gims/api/internal/inventory/data/models"
	purchaseModels "github.com/gilabs/gims/api/internal/purchase/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"gorm.io/gorm"
)

var (
	ErrValuationConflict      = errors.New("a valuation run is already pending for this type and period")
	ErrValuationPeriod        = errors.New("invalid valuation period: start must be before or equal to end")
	ErrValuationTypeUnknown   = errors.New("unknown valuation type")
	ErrValuationStatus        = errors.New("valuation run is not in approvable status")
	ErrReconciliationFailed   = errors.New("inventory reconciliation mismatch: inventory GL is not equal to subledger")
	ErrPeriodLocked           = errors.New("cannot run valuation: period is locked (already posted)")
	ErrPeriodLockRequired      = errors.New("period locking required: use unlock endpoint to modify posted runs")
)

type ValuationItem struct {
	ReferenceID string
	ProductID   *string
	Qty         float64
	BookValue   float64
	ActualValue float64
	Delta       float64
	Direction   string
}

type ValuationResult struct {
	Items      []ValuationItem
	TotalDelta float64
}

type ValuationStrategy interface {
	Calculate(ctx context.Context, periodStart, periodEnd time.Time) (*ValuationResult, error)
}

type ValuationRunUsecase interface {
	Preview(ctx context.Context, req *dto.RunValuationRequest) (*dto.ValuationPreviewResponse, error)
	Run(ctx context.Context, req *dto.RunValuationRequest) (*dto.ValuationRunResponse, error)
	Approve(ctx context.Context, id string, req *dto.ApproveValuationRequest) (*dto.ValuationRunResponse, error)
	Unlock(ctx context.Context, id string, req *dto.UnlockValuationRequest) (*dto.ValuationRunResponse, error)
	BulkApprove(ctx context.Context, req *dto.BulkApproveValuationRequest) (*dto.BulkApproveValuationResponse, error)
	GetByID(ctx context.Context, id string) (*dto.ValuationRunResponse, error)
	List(ctx context.Context, req *dto.ListValuationRunsRequest) ([]dto.ValuationRunResponse, int64, *dto.ValuationKPIMeta, error)
}

type valuationRunUsecase struct {
	db                  *gorm.DB
	repo                repositories.ValuationRunRepository
	journalUC           JournalEntryUsecase
	accountingEngine    accounting.AccountingEngine
	settings            financesettings.SettingsService
	settingsValidator   financesettings.SettingsValidator
	strategyValidator   ValuationStrategyValidator
	strategies          map[string]ValuationStrategy
}

func NewValuationRunUsecase(
	db *gorm.DB,
	repo repositories.ValuationRunRepository,
	journalUC JournalEntryUsecase,
	settings financesettings.SettingsService,
	accountingEngine accounting.AccountingEngine,
) ValuationRunUsecase {
	uc := &valuationRunUsecase{
		db:                db,
		repo:              repo,
		journalUC:         journalUC,
		accountingEngine:  accountingEngine,
		settings:          settings,
		settingsValidator: financesettings.NewSettingsValidator(settings),
		strategyValidator: NewValuationStrategyValidator(db),
		strategies:        make(map[string]ValuationStrategy),
	}

	uc.strategies["inventory"] = &inventoryValuationStrategy{db: db}
	uc.strategies["fx"] = &fxValuationStrategy{db: db}
	uc.strategies["depreciation"] = &depreciationValuationStrategy{db: db}

	return uc
}

func (uc *valuationRunUsecase) Preview(ctx context.Context, req *dto.RunValuationRequest) (*dto.ValuationPreviewResponse, error) {
	// 1. Validate required finance settings exist
	if err := uc.settingsValidator.ValidateRequiredSettings(ctx, financesettings.ValidatorConfig{
		RequiredSettings: financesettings.DefaultRequiredSettings(),
		FailFast:         true,
	}); err != nil {
		return nil, fmt.Errorf("finance configuration incomplete: %w", err)
	}

	// 2. Resolve and validate strategy
	valType, periodStart, periodEnd, _, strategy, err := uc.resolveRunRequest(ctx, req, true)
	if err != nil {
		return nil, err
	}

	// 3. Strategy-specific validation
	if err := uc.validateStrategyDataRequirements(ctx, valType); err != nil {
		return nil, fmt.Errorf("valuation data validation failed: %w", err)
	}

	// 4. Calculate valuation
	result, err := strategy.Calculate(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("valuation calculation failed: %w", err)
	}

	previewLines, _, _, err := uc.buildJournalPreview(ctx, valType, result, "VAL-PREVIEW", periodEnd)
	if err != nil {
		return nil, err
	}

	debit, credit := 0.0, 0.0
	for _, line := range previewLines {
		debit += line.Debit
		credit += line.Credit
	}

	return &dto.ValuationPreviewResponse{
		ValuationType: valType,
		PeriodStart:   periodStart.Format("2006-01-02"),
		PeriodEnd:     periodEnd.Format("2006-01-02"),
		Items:         mapValuationItems(result.Items),
		TotalDelta:    result.TotalDelta,
		TotalGain:     totalGain(result.Items),
		TotalLoss:     totalLoss(result.Items),
		JournalLines:  previewLines,
		IsBalanced:    math.Abs(debit-credit) < 0.0001,
	}, nil
}

func (uc *valuationRunUsecase) Run(ctx context.Context, req *dto.RunValuationRequest) (*dto.ValuationRunResponse, error) {
	// 1. Validate required finance settings exist
	if err := uc.settingsValidator.ValidateRequiredSettings(ctx, financesettings.ValidatorConfig{
		RequiredSettings: financesettings.DefaultRequiredSettings(),
		FailFast:         true,
	}); err != nil {
		return nil, fmt.Errorf("finance configuration incomplete: %w", err)
	}

	// 2. Resolve and validate strategy
	valType, periodStart, periodEnd, refID, strategy, err := uc.resolveRunRequest(ctx, req, false)
	if err != nil {
		return nil, err
	}

	// 3. Strategy-specific validation
	if err := uc.validateStrategyDataRequirements(ctx, valType); err != nil {
		return nil, fmt.Errorf("valuation data validation failed: %w", err)
	}

	// 4. Check for duplicate (idempotency)
	existing, err := uc.repo.FindByReferenceID(ctx, refID)
	if err == nil && existing != nil {
		// Check if existing run is locked (posted period)
		if existing.IsLocked {
			return nil, fmt.Errorf("%w (locked at %s) — use admin endpoint to unlock period", ErrPeriodLocked, existing.LockedAt)
		}
		return uc.toResponse(existing), nil
	}

	// 5. Check for period lock (prevent overwrite of posted runs)
	hasPending, err := uc.repo.HasPendingRun(ctx, valType, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to check valuation lock: %w", err)
	}
	if hasPending {
		return nil, ErrValuationConflict
	}

	// 6. Calculate valuation
	result, err := strategy.Calculate(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("valuation calculation failed: %w", err)
	}

	actorID := strings.TrimSpace(getActorID(ctx))
	run := &financeModels.ValuationRun{
		ReferenceID:   refID,
		ValuationType: financeModels.ValuationType(valType),
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Status:        financeModels.ValuationRunStatusDraft,
		TotalDelta:    result.TotalDelta,
		CreatedBy:     nullableString(actorID),
	}

	details := make([]financeModels.ValuationRunDetail, 0, len(result.Items))
	for _, item := range result.Items {
		details = append(details, financeModels.ValuationRunDetail{
			ValuationRunID: "",
			ReferenceID:    item.ReferenceID,
			ProductID:      item.ProductID,
			Qty:            item.Qty,
			BookValue:      item.BookValue,
			ActualValue:    item.ActualValue,
			Delta:          item.Delta,
			Direction:      financeModels.ValuationDirection(item.Direction),
		})
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(run).Error; err != nil {
			return err
		}
		for i := range details {
			details[i].ValuationRunID = run.ID
		}
		if err := uc.repo.CreateDetails(ctx, tx, details); err != nil {
			return err
		}
		if len(details) == 0 || math.Abs(result.TotalDelta) < 0.0001 {
			run.Status = financeModels.ValuationRunStatusNoDifference
			now := apptime.Now()
			run.CompletedAt = &now
		} else {
			run.Status = financeModels.ValuationRunStatusPendingApproval
		}
		if err := tx.Save(run).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist valuation run: %w", err)
	}

	run.Details = details
	return uc.toResponse(run), nil
}

func (uc *valuationRunUsecase) Approve(ctx context.Context, id string, _ *dto.ApproveValuationRequest) (*dto.ValuationRunResponse, error) {
	actorID := strings.TrimSpace(getActorID(ctx))
	if actorID == "" {
		return nil, errors.New("user context is required")
	}

	var outRun *financeModels.ValuationRun
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		run, err := uc.repo.FindByIDForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}

		if run.Status == financeModels.ValuationRunStatusPosted {
			outRun = run
			return nil
		}
		if run.Status != financeModels.ValuationRunStatusPendingApproval && run.Status != financeModels.ValuationRunStatusApproved {
			return ErrValuationStatus
		}

		if err := uc.validateReconciliation(ctx, tx, run); err != nil {
			return err
		}

		result := detailsToResult(run.Details)
		previewLines, totalDebit, totalCredit, err := uc.buildJournalPreview(ctx, string(run.ValuationType), result, run.ReferenceID, run.PeriodEnd)
		if err != nil {
			return err
		}

		if len(previewLines) == 0 {
			run.Status = financeModels.ValuationRunStatusNoDifference
			now := apptime.Now()
			run.CompletedAt = &now
			if err := tx.Save(run).Error; err != nil {
				return err
			}
			outRun = run
			return nil
		}

		run.Status = financeModels.ValuationRunStatusApproved
		if err := tx.Save(run).Error; err != nil {
			return err
		}

		journalID, err := uc.createPostedJournalInTx(ctx, tx, run, previewLines, totalDebit, totalCredit, actorID)
		if err != nil {
			return err
		}

		run.Status = financeModels.ValuationRunStatusPosted
		run.TotalDebit = totalDebit
		run.TotalCredit = totalCredit
		run.JournalEntryID = &journalID
		run.ApprovedBy = &actorID
		run.ApprovedAt = timePtr(apptime.Now())
		now := apptime.Now()
		run.CompletedAt = &now
		
		// CRITICAL: Lock period after posting to prevent re-runs
		run.IsLocked = true
		run.LockedAt = timePtr(apptime.Now())
		
		if err := tx.Save(run).Error; err != nil {
			return err
		}

		outRun = run
		return nil
	})
	if err != nil {
		return nil, err
	}

	return uc.toResponse(outRun), nil
}

// Unlock removes the period lock for a posted valuation run (admin-only operation).
// This allows corrections to be made if errors are discovered post-posting.
// Requires explicit RBAC permission for audit trail.
func (uc *valuationRunUsecase) Unlock(ctx context.Context, id string, req *dto.UnlockValuationRequest) (*dto.ValuationRunResponse, error) {
	actorID := strings.TrimSpace(getActorID(ctx))
	if actorID == "" {
		return nil, errors.New("user context is required for unlock audit trail")
	}

	if req == nil {
		req = &dto.UnlockValuationRequest{}
	}

	var outRun *financeModels.ValuationRun
	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		run, err := uc.repo.FindByIDForUpdate(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("valuation run not found: %w", err)
		}

		// Only allow unlock of posted (locked) runs
		if !run.IsLocked {
			return errors.New("cannot unlock: valuation run is not currently locked")
		}

		if run.Status != financeModels.ValuationRunStatusPosted {
			return fmt.Errorf("cannot unlock: invalid status (%s); only posted runs can be unlocked", run.Status)
		}

		// Unlock the period
		run.IsLocked = false
		run.LockedAt = nil
		
		// Store unlock reason for audit trail (optional)
		if req.UnlockReason != "" {
			reasonNote := fmt.Sprintf("Unlocked by %s at %s: %s", actorID, apptime.Now().Format("2006-01-02 15:04:05"), req.UnlockReason)
			if run.ApprovalNotes != "" {
				run.ApprovalNotes = run.ApprovalNotes + "\n" + reasonNote
			} else {
				run.ApprovalNotes = reasonNote
			}
		}

		if err := tx.Save(run).Error; err != nil {
			return fmt.Errorf("failed to unlock valuation run: %w", err)
		}

		outRun = run
		return nil
	})

	if err != nil {
		return nil, err
	}

	return uc.toResponse(outRun), nil
}

// BulkApprove approves multiple valuation runs in a single transaction.
// Returns success/failure status per run for batch processing.
func (uc *valuationRunUsecase) BulkApprove(ctx context.Context, req *dto.BulkApproveValuationRequest) (*dto.BulkApproveValuationResponse, error) {
	if req == nil || len(req.RunIDs) == 0 {
		return nil, errors.New("no runs provided for bulk approve")
	}

	if len(req.RunIDs) > 100 {
		return nil, errors.New("bulk approve limited to 100 runs per request")
	}

	actorID := strings.TrimSpace(getActorID(ctx))
	if actorID == "" {
		return nil, errors.New("user context is required")
	}

	response := &dto.BulkApproveValuationResponse{
		Results:         make([]dto.BulkApproveResult, 0, len(req.RunIDs)),
		TotalProcessed:  0,
		SuccessCount:    0,
		FailureCount:    0,
	}

	// Process each run serially to capture individual errors
	for _, runID := range req.RunIDs {
		result := dto.BulkApproveResult{
			RunID: runID,
		}

		// Approve individually within a transaction
		_, err := uc.Approve(ctx, runID, &dto.ApproveValuationRequest{Notes: ""})

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			response.FailureCount++
		} else {
			result.Success = true
			result.Error = ""
			response.SuccessCount++
		}

		response.Results = append(response.Results, result)
		response.TotalProcessed++
	}

	response.ProcessedAt = apptime.Now().Format("2006-01-02T15:04:05Z07:00")
	return response, nil
}

func (uc *valuationRunUsecase) GetByID(ctx context.Context, id string) (*dto.ValuationRunResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.ValuationRun{}, id, security.FinanceScopeQueryOptions()) {
		return nil, errors.New("valuation run not found")
	}
	run, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return uc.toResponse(run), nil
}

func (uc *valuationRunUsecase) List(ctx context.Context, req *dto.ListValuationRunsRequest) ([]dto.ValuationRunResponse, int64, *dto.ValuationKPIMeta, error) {
	page := req.Page
	perPage := req.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t, _ := time.Parse("2006-01-02", *req.StartDate)
		startDate = &t
	}
	if req.EndDate != nil {
		t, _ := time.Parse("2006-01-02", *req.EndDate)
		endDate = &t
	}

	items, total, err := uc.repo.List(ctx, repositories.ValuationRunListParams{
		ValuationType: req.ValuationType,
		Status:        req.Status,
		StartDate:     startDate,
		EndDate:       endDate,
		SortBy:        req.SortBy,
		SortDir:       req.SortDir,
		Limit:         perPage,
		Offset:        (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, nil, err
	}

	responses := make([]dto.ValuationRunResponse, 0, len(items))
	kpi := &dto.ValuationKPIMeta{TotalEntries: total}

	for i := range items {
		responses = append(responses, *uc.toResponse(&items[i]))
		kpi.TotalDebitSum += items[i].TotalDebit
		kpi.TotalCreditSum += items[i].TotalCredit
		switch items[i].Status {
		case financeModels.ValuationRunStatusPosted:
			kpi.CompletedRuns++
		case financeModels.ValuationRunStatusPendingApproval, financeModels.ValuationRunStatusApproved:
			kpi.ProcessingRuns++
		case financeModels.ValuationRunStatusFailed:
			kpi.FailedRuns++
		}
	}

	return responses, total, kpi, nil
}

func (uc *valuationRunUsecase) resolveRunRequest(ctx context.Context, req *dto.RunValuationRequest, allowEmptyRef bool) (string, time.Time, time.Time, string, ValuationStrategy, error) {
	if req == nil {
		return "", time.Time{}, time.Time{}, "", nil, errors.New("request is required")
	}

	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		return "", time.Time{}, time.Time{}, "", nil, fmt.Errorf("invalid period_start: %w", err)
	}
	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		return "", time.Time{}, time.Time{}, "", nil, fmt.Errorf("invalid period_end: %w", err)
	}
	if periodStart.After(periodEnd) {
		return "", time.Time{}, time.Time{}, "", nil, ErrValuationPeriod
	}

	valType := strings.ToLower(strings.TrimSpace(req.ValuationType))
	strategy, ok := uc.strategies[valType]
	if !ok {
		return "", time.Time{}, time.Time{}, "", nil, ErrValuationTypeUnknown
	}

	refID := strings.TrimSpace(req.ReferenceID)
	if refID == "" && !allowEmptyRef {
		refID = fmt.Sprintf("VAL-%s-%s", strings.ToUpper(valType), apptime.Now().Format("20060102150405"))
	}

	_ = ctx
	return valType, periodStart, periodEnd, refID, strategy, nil
}

func (uc *valuationRunUsecase) buildJournalPreview(
	ctx context.Context,
	valuationType string,
	result *ValuationResult,
	referenceID string,
	periodEnd time.Time,
) ([]dto.ValuationPreviewJournalLine, float64, float64, error) {
	if result == nil {
		return nil, 0, 0, nil
	}

	gain := totalGain(result.Items)
	loss := totalLoss(result.Items)
	preview := make([]dto.ValuationPreviewJournalLine, 0)
	totalDebit := 0.0
	totalCredit := 0.0

	appendLines := func(profile accounting.PostingProfile, amount float64, desc string) error {
		if amount <= 0 {
			return nil
		}
		req, err := uc.accountingEngine.GenerateJournal(ctx, profile, accounting.TransactionData{
			ReferenceType: profile.ReferenceType,
			ReferenceID:   referenceID,
			EntryDate:     periodEnd.Format("2006-01-02"),
			Description:   desc,
			TotalAmount:   amount,
		})
		if err != nil {
			return err
		}
		for _, l := range req.Lines {
			preview = append(preview, dto.ValuationPreviewJournalLine{
				ChartOfAccountID: l.ChartOfAccountID,
				Debit:            l.Debit,
				Credit:           l.Credit,
				Memo:             l.Memo,
			})
			totalDebit += l.Debit
			totalCredit += l.Credit
		}
		return nil
	}

	switch valuationType {
	case "inventory":
		if err := appendLines(accounting.ProfileInventoryValuation, gain, "Inventory valuation gain"); err != nil {
			return nil, 0, 0, err
		}
		if err := appendLines(accounting.ProfileInventoryValuationLoss, loss, "Inventory valuation loss"); err != nil {
			return nil, 0, 0, err
		}
	case "fx":
		if err := appendLines(accounting.ProfileFXValuation, gain, "FX valuation gain"); err != nil {
			return nil, 0, 0, err
		}
		if err := appendLines(accounting.ProfileFXValuationLoss, loss, "FX valuation loss"); err != nil {
			return nil, 0, 0, err
		}
	case "depreciation":
		if err := appendLines(accounting.ProfileDepreciationGain, gain, "Depreciation valuation gain"); err != nil {
			return nil, 0, 0, err
		}
		if err := appendLines(accounting.ProfileDepreciation, loss, "Depreciation valuation loss"); err != nil {
			return nil, 0, 0, err
		}
	default:
		return nil, 0, 0, ErrValuationTypeUnknown
	}

	return preview, totalDebit, totalCredit, nil
}

func (uc *valuationRunUsecase) createPostedJournalInTx(
	ctx context.Context,
	tx *gorm.DB,
	run *financeModels.ValuationRun,
	lines []dto.ValuationPreviewJournalLine,
	totalDebit float64,
	totalCredit float64,
	actorID string,
) (string, error) {
	entryDate, err := time.Parse("2006-01-02", run.PeriodEnd.Format("2006-01-02"))
	if err != nil {
		return "", err
	}

	refType := mapValuationToRefType(string(run.ValuationType))
	journal := &financeModels.JournalEntry{
		EntryDate:         entryDate,
		Description:       fmt.Sprintf("Valuation run %s", run.ReferenceID),
		ReferenceType:     &refType,
		ReferenceID:       &run.ReferenceID,
		Status:            financeModels.JournalStatusPosted,
		PostedBy:          &actorID,
		PostedAt:          timePtr(apptime.Now()),
		CreatedBy:         &actorID,
		IsSystemGenerated: true,
		IsValuation:       true,
		Source:            financeModels.JournalSourceValuation,
		ValuationRunID:    &run.ID,
		DebitTotal:        totalDebit,
		CreditTotal:       totalCredit,
	}
	if err := tx.WithContext(ctx).Create(journal).Error; err != nil {
		return "", err
	}

	journalLines := make([]financeModels.JournalLine, 0, len(lines))
	for _, line := range lines {
		journalLines = append(journalLines, financeModels.JournalLine{
			JournalEntryID:   journal.ID,
			ChartOfAccountID: line.ChartOfAccountID,
			Debit:            line.Debit,
			Credit:           line.Credit,
			Memo:             line.Memo,
		})
	}
	if len(journalLines) > 0 {
		if err := tx.WithContext(ctx).Create(&journalLines).Error; err != nil {
			return "", err
		}
	}

	return journal.ID, nil
}

func (uc *valuationRunUsecase) validateReconciliation(ctx context.Context, tx *gorm.DB, run *financeModels.ValuationRun) error {
	if run.ValuationType != financeModels.ValuationTypeInventory {
		return nil
	}

	coaID, err := uc.accountingEngine.ResolveCOAID(ctx, financeModels.SettingCOAInventoryAsset)
	if err != nil {
		return err
	}

	var glBalance float64
	if err := tx.WithContext(ctx).
		Table("journal_lines jl").
		Select("COALESCE(SUM(jl.debit - jl.credit), 0)").
		Joins("JOIN journal_entries je ON je.id = jl.journal_entry_id").
		Where("je.status = ? AND jl.chart_of_account_id = ? AND je.entry_date <= ?", financeModels.JournalStatusPosted, coaID, run.PeriodEnd).
		Scan(&glBalance).Error; err != nil {
		return err
	}

	var subledger float64
	if err := tx.WithContext(ctx).
		Model(&inventoryModels.InventoryBatch{}).
		Select("COALESCE(SUM(current_quantity * cost_price), 0)").
		Where("is_active = ?", true).
		Scan(&subledger).Error; err != nil {
		return err
	}

	tolerance := uc.reconciliationTolerance(ctx)
	if math.Abs(glBalance-subledger) > tolerance {
		return fmt.Errorf("%w (gl=%.2f subledger=%.2f tolerance=%.2f)", ErrReconciliationFailed, glBalance, subledger, tolerance)
	}

	return nil
}

func (uc *valuationRunUsecase) reconciliationTolerance(ctx context.Context) float64 {
	value, err := uc.settings.GetValue(ctx, "valuation.reconciliation_tolerance")
	if err != nil || strings.TrimSpace(value) == "" {
		return 0.01
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 {
		return 0.01
	}
	return parsed
}

func (uc *valuationRunUsecase) toResponse(run *financeModels.ValuationRun) *dto.ValuationRunResponse {
	resp := &dto.ValuationRunResponse{
		ID:             run.ID,
		ReferenceID:    run.ReferenceID,
		ValuationType:  string(run.ValuationType),
		PeriodStart:    run.PeriodStart.Format("2006-01-02"),
		PeriodEnd:      run.PeriodEnd.Format("2006-01-02"),
		Status:         string(run.Status),
		TotalDebit:     run.TotalDebit,
		TotalCredit:    run.TotalCredit,
		TotalDelta:     run.TotalDelta,
		JournalEntryID: run.JournalEntryID,
		ErrorMessage:   run.ErrorMessage,
		IsLocked:       run.IsLocked,
		ApprovedBy:     run.ApprovedBy,
		ApprovalNotes:  run.ApprovalNotes,
		CreatedBy:      run.CreatedBy,
		Items:          mapValuationDetails(run.Details),
		CreatedAt:      run.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      run.UpdatedAt.Format(time.RFC3339),
	}
	if run.LockedAt != nil {
		s := run.LockedAt.Format(time.RFC3339)
		resp.LockedAt = &s
	}
	if run.ApprovedAt != nil {
		s := run.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &s
	}
	if run.CompletedAt != nil {
		s := run.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &s
	}
	return resp
}

func mapValuationItems(items []ValuationItem) []dto.ValuationItemResponse {
	out := make([]dto.ValuationItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ValuationItemResponse{
			ReferenceID: item.ReferenceID,
			ProductID:   item.ProductID,
			Qty:         item.Qty,
			BookValue:   item.BookValue,
			ActualValue: item.ActualValue,
			Delta:       item.Delta,
			Direction:   item.Direction,
		})
	}
	return out
}

func mapValuationDetails(items []financeModels.ValuationRunDetail) []dto.ValuationItemResponse {
	out := make([]dto.ValuationItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ValuationItemResponse{
			ReferenceID: item.ReferenceID,
			ProductID:   item.ProductID,
			Qty:         item.Qty,
			BookValue:   item.BookValue,
			ActualValue: item.ActualValue,
			Delta:       item.Delta,
			Direction:   string(item.Direction),
		})
	}
	return out
}

func detailsToResult(items []financeModels.ValuationRunDetail) *ValuationResult {
	res := &ValuationResult{Items: make([]ValuationItem, 0, len(items))}
	for _, item := range items {
		res.Items = append(res.Items, ValuationItem{
			ReferenceID: item.ReferenceID,
			ProductID:   item.ProductID,
			Qty:         item.Qty,
			BookValue:   item.BookValue,
			ActualValue: item.ActualValue,
			Delta:       item.Delta,
			Direction:   string(item.Direction),
		})
		res.TotalDelta += item.Delta
	}
	return res
}

func totalGain(items []ValuationItem) float64 {
	sum := 0.0
	for _, item := range items {
		if item.Delta > 0 {
			sum += item.Delta
		}
	}
	return sum
}

func totalLoss(items []ValuationItem) float64 {
	sum := 0.0
	for _, item := range items {
		if item.Delta < 0 {
			sum += math.Abs(item.Delta)
		}
	}
	return sum
}

func mapValuationToRefType(valuationType string) string {
	switch valuationType {
	case "inventory":
		return reference.RefTypeInventoryValuation
	case "fx":
		return reference.RefTypeCurrencyRevaluation
	case "depreciation":
		return reference.RefTypeDepreciationValuation
	default:
		return reference.RefTypeInventoryValuation
	}
}

func getActorID(ctx context.Context) string {
	if v, ok := ctx.Value("user_id").(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func nullableString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// validateStrategyDataRequirements checks that the strategy-specific data requirements are met.
// This fails fast with explicit errors rather than silent failures.
func (uc *valuationRunUsecase) validateStrategyDataRequirements(ctx context.Context, valType string) error {
	switch valType {
	case "inventory":
		return uc.strategyValidator.ValidateInventoryData(ctx)
	case "fx":
		return uc.strategyValidator.ValidateFXData(ctx)
	case "depreciation":
		return uc.strategyValidator.ValidateDepreciationData(ctx)
	default:
		return fmt.Errorf("unknown valuation type: %s", valType)
	}
}

// inventoryValuationStrategy computes itemized inventory valuation from inventory_batches and stock_movements.
type inventoryValuationStrategy struct {
	db *gorm.DB
}

func (s *inventoryValuationStrategy) Calculate(ctx context.Context, periodStart, periodEnd time.Time) (*ValuationResult, error) {
	type row struct {
		ProductID string
		Qty       float64
		BookValue float64
		AvgCost   float64
	}

	rows := make([]row, 0)
	queryDB := database.GetDB(ctx, s.db)
	query := `
		SELECT
			ib.product_id,
			COALESCE(SUM(ib.current_quantity), 0) AS qty,
			COALESCE(SUM(ib.current_quantity * ib.cost_price), 0) AS book_value,
			COALESCE((
				SELECT AVG(sm.cost)
				FROM stock_movements sm
				WHERE sm.product_id = ib.product_id
					AND DATE(sm.date) BETWEEN DATE(?) AND DATE(?)
					AND sm.cost > 0
			), 0) AS avg_cost
		FROM inventory_batches ib
		WHERE ib.is_active = true
			AND ib.current_quantity > 0
	`
	args := []interface{}{periodStart, periodEnd}
	if tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx)); tenantID != "" {
		query += " AND ib.tenant_id = ?"
		args = append(args, tenantID)
	}
	query += " GROUP BY ib.product_id"
	err := queryDB.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]ValuationItem, 0, len(rows))
	totalDelta := 0.0
	for _, r := range rows {
		actualValue := r.BookValue
		if r.AvgCost > 0 {
			actualValue = r.Qty * r.AvgCost
		}
		delta := actualValue - r.BookValue
		if math.Abs(delta) < 0.0001 {
			continue
		}

		pid := r.ProductID
		direction := "loss"
		if delta > 0 {
			direction = "gain"
		}

		items = append(items, ValuationItem{
			ReferenceID: r.ProductID,
			ProductID:   &pid,
			Qty:         r.Qty,
			BookValue:   r.BookValue,
			ActualValue: actualValue,
			Delta:       delta,
			Direction:   direction,
		})
		totalDelta += delta
	}

	return &ValuationResult{Items: items, TotalDelta: totalDelta}, nil
}

// fxValuationStrategy computes FX valuation if exchange_rate support exists.
type fxValuationStrategy struct {
	db *gorm.DB
}

func (s *fxValuationStrategy) Calculate(ctx context.Context, periodStart, periodEnd time.Time) (*ValuationResult, error) {
	_ = periodStart
	_ = periodEnd

	queryDB := database.GetDB(ctx, s.db)
	if !queryDB.Migrator().HasTable("exchange_rates") {
		return &ValuationResult{Items: nil, TotalDelta: 0}, nil
	}

	items := make([]ValuationItem, 0)
	totalDelta := 0.0
	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))

	if queryDB.Migrator().HasColumn(&salesModels.CustomerInvoice{}, "exchange_rate") {
		type arRow struct {
			ID           string
			Remaining    float64
			ExchangeRate float64
			CurrentRate  float64
		}
		arRows := make([]arRow, 0)
		query := `
			SELECT
				ci.id,
				COALESCE(ci.remaining_amount, 0) AS remaining,
				COALESCE(ci.exchange_rate, 0) AS exchange_rate,
				COALESCE((
					SELECT er.rate FROM exchange_rates er
					WHERE er.currency_code = ci.currency_code
					ORDER BY er.rate_date DESC
					LIMIT 1
				), 0) AS current_rate
			FROM customer_invoices ci
			WHERE ci.status IN (?, ?, ?)
				AND COALESCE(ci.remaining_amount, 0) > 0
		`
		args := []interface{}{salesModels.CustomerInvoiceStatusUnpaid, salesModels.CustomerInvoiceStatusPartial, salesModels.CustomerInvoiceStatusWaitingPayment}
		if tenantID != "" {
			query += " AND ci.tenant_id = ?"
			args = append(args, tenantID)
		}
		_ = queryDB.WithContext(ctx).Raw(query, args...).Scan(&arRows).Error

		for _, row := range arRows {
			if row.ExchangeRate <= 0 || row.CurrentRate <= 0 {
				continue
			}
			book := row.Remaining * row.ExchangeRate
			actual := row.Remaining * row.CurrentRate
			delta := actual - book
			if math.Abs(delta) < 0.0001 {
				continue
			}
			direction := "loss"
			if delta > 0 {
				direction = "gain"
			}
			items = append(items, ValuationItem{
				ReferenceID: row.ID,
				Qty:         row.Remaining,
				BookValue:   book,
				ActualValue: actual,
				Delta:       delta,
				Direction:   direction,
			})
			totalDelta += delta
		}
	}

	if queryDB.Migrator().HasColumn(&purchaseModels.SupplierInvoice{}, "exchange_rate") {
		type apRow struct {
			ID           string
			Remaining    float64
			ExchangeRate float64
			CurrentRate  float64
		}
		apRows := make([]apRow, 0)
		query := `
			SELECT
				si.id,
				COALESCE(si.remaining_amount, 0) AS remaining,
				COALESCE(si.exchange_rate, 0) AS exchange_rate,
				COALESCE((
					SELECT er.rate FROM exchange_rates er
					WHERE er.currency_code = si.currency_code
					ORDER BY er.rate_date DESC
					LIMIT 1
				), 0) AS current_rate
			FROM supplier_invoices si
			WHERE si.status IN (?, ?, ?)
				AND COALESCE(si.remaining_amount, 0) > 0
		`
		args := []interface{}{purchaseModels.SupplierInvoiceStatusUnpaid, purchaseModels.SupplierInvoiceStatusPartial, purchaseModels.SupplierInvoiceStatusWaitingPayment}
		if tenantID != "" {
			query += " AND si.tenant_id = ?"
			args = append(args, tenantID)
		}
		_ = queryDB.WithContext(ctx).Raw(query, args...).Scan(&apRows).Error

		for _, row := range apRows {
			if row.ExchangeRate <= 0 || row.CurrentRate <= 0 {
				continue
			}
			book := row.Remaining * row.ExchangeRate
			actual := row.Remaining * row.CurrentRate
			delta := book - actual
			if math.Abs(delta) < 0.0001 {
				continue
			}
			direction := "loss"
			if delta > 0 {
				direction = "gain"
			}
			items = append(items, ValuationItem{
				ReferenceID: row.ID,
				Qty:         row.Remaining,
				BookValue:   book,
				ActualValue: actual,
				Delta:       delta,
				Direction:   direction,
			})
			totalDelta += delta
		}
	}

	return &ValuationResult{Items: items, TotalDelta: totalDelta}, nil
}

// depreciationValuationStrategy computes depreciation deltas from approved depreciation schedules.
type depreciationValuationStrategy struct {
	db *gorm.DB
}

func (s *depreciationValuationStrategy) Calculate(ctx context.Context, periodStart, periodEnd time.Time) (*ValuationResult, error) {
	type row struct {
		ID        string
		AssetID   string
		Amount    float64
		BookValue float64
	}
	rows := make([]row, 0)
	queryDB := database.GetDB(ctx, s.db)
	query := `
		SELECT ad.id, ad.asset_id, COALESCE(ad.amount, 0) AS amount, COALESCE(ad.book_value, 0) AS book_value
		FROM asset_depreciations ad
		WHERE ad.status = ?
			AND DATE(ad.depreciation_date) BETWEEN DATE(?) AND DATE(?)
	`
	args := []interface{}{financeModels.AssetDepreciationStatusApproved, periodStart, periodEnd}
	if tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx)); tenantID != "" {
		query += " AND ad.tenant_id = ?"
		args = append(args, tenantID)
	}
	err := queryDB.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]ValuationItem, 0, len(rows))
	totalDelta := 0.0
	for _, r := range rows {
		book := r.BookValue + r.Amount
		actual := r.BookValue
		delta := actual - book
		if math.Abs(delta) < 0.0001 {
			continue
		}
		assetID := r.AssetID
		direction := "loss"
		if delta > 0 {
			direction = "gain"
		}
		items = append(items, ValuationItem{
			ReferenceID: r.ID,
			ProductID:   &assetID,
			Qty:         1,
			BookValue:   book,
			ActualValue: actual,
			Delta:       delta,
			Direction:   direction,
		})
		totalDelta += delta
	}

	return &ValuationResult{Items: items, TotalDelta: totalDelta}, nil
}
