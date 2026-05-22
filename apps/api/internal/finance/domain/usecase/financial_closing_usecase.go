package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrFinancialClosingNotFound = errors.New("financial closing not found")
	// periodClosingNamespace is a deterministic namespace UUID for creating stable reference IDs
	periodClosingNamespace = uuid.MustParse("d9b5f22a-0000-0000-0000-000000000000")
)

type FinancialClosingUsecase interface {
	Create(ctx context.Context, req *dto.CreateFinancialClosingRequest) (*dto.FinancialClosingResponse, error)
	Approve(ctx context.Context, id string) (*dto.FinancialClosingResponse, error)
	Reopen(ctx context.Context, id string) (*dto.FinancialClosingResponse, error)
	YearEndClose(ctx context.Context, req *dto.YearEndCloseRequest) (*dto.FinancialClosingResponse, error)
	GetByID(ctx context.Context, id string) (*dto.FinancialClosingResponse, error)
	List(ctx context.Context, req *dto.ListFinancialClosingsRequest) ([]dto.FinancialClosingResponse, int64, error)
	GetAnalysis(ctx context.Context, id string) (*dto.FinancialClosingAnalysisResponse, error)
	Delete(ctx context.Context, id string) error
}

type financialClosingUsecase struct {
	db             *gorm.DB
	coaRepo        repositories.ChartOfAccountRepository
	repo           repositories.FinancialClosingRepository
	accountingRepo repositories.AccountingPeriodRepository
	snapshotRepo   repositories.FinancialClosingSnapshotRepository
	logRepo        repositories.FinancialClosingLogRepository
	journalUC      JournalEntryUsecase
	mapper         *mapper.FinancialClosingMapper
}

func NewFinancialClosingUsecase(
	db *gorm.DB,
	coaRepo repositories.ChartOfAccountRepository,
	repo repositories.FinancialClosingRepository,
	accountingRepo repositories.AccountingPeriodRepository,
	snapshotRepo repositories.FinancialClosingSnapshotRepository,
	logRepo repositories.FinancialClosingLogRepository,
	journalUC JournalEntryUsecase,
	mapper *mapper.FinancialClosingMapper,
) FinancialClosingUsecase {
	return &financialClosingUsecase{db: db, coaRepo: coaRepo, repo: repo, accountingRepo: accountingRepo, snapshotRepo: snapshotRepo, logRepo: logRepo, journalUC: journalUC, mapper: mapper}
}

func (uc *financialClosingUsecase) getClosingPeriodRange(ctx context.Context, endDate time.Time) (time.Time, time.Time, error) {
	start := time.Time{}
	latest, err := uc.repo.LatestApproved(ctx)
	if err == nil {
		// start is the day after the latest approved closing
		start = latest.PeriodEndDate.AddDate(0, 0, 1)
	} else if err != gorm.ErrRecordNotFound {
		return time.Time{}, time.Time{}, err
	}
	return start, endDate, nil
}

func (uc *financialClosingUsecase) validateClosingPeriod(ctx context.Context, start, end time.Time) ([]dto.FinancialClosingValidationResult, error) {
	results := []dto.FinancialClosingValidationResult{}

	// 1) All journal entries within the period must be posted
	var unpostedCount int64
	if err := uc.db.WithContext(ctx).
		Model(&financeModels.JournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ?", start, end).
		Where("status != ?", financeModels.JournalStatusPosted).
		Count(&unpostedCount).Error; err != nil {
		return nil, err
	}
	if unpostedCount == 0 {
		results = append(results, dto.FinancialClosingValidationResult{Name: "posted_journals", Passed: true, Message: "All journal entries are posted"})
	} else {
		results = append(results, dto.FinancialClosingValidationResult{Name: "posted_journals", Passed: false, Message: fmt.Sprintf("Found %d unposted journal entries", unpostedCount)})
	}

	// 2) Trial balance must be balanced
	type aggRow struct {
		DebitTotal  float64
		CreditTotal float64
	}
	var tb aggRow
	if err := uc.db.WithContext(ctx).
		Table("journal_lines").
		Select("COALESCE(SUM(debit),0) AS debit_total, COALESCE(SUM(credit),0) AS credit_total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.status = ?", financeModels.JournalStatusPosted).
		Where("journal_entries.entry_date >= ? AND journal_entries.entry_date <= ?", start, end).
		Scan(&tb).Error; err != nil {
		return nil, err
	}
	if math.Abs(tb.DebitTotal-tb.CreditTotal) < 0.0001 {
		results = append(results, dto.FinancialClosingValidationResult{Name: "trial_balance", Passed: true, Message: "Trial balance is balanced"})
	} else {
		results = append(results, dto.FinancialClosingValidationResult{Name: "trial_balance", Passed: false, Message: fmt.Sprintf("Trial balance mismatch: debit %.2f vs credit %.2f", tb.DebitTotal, tb.CreditTotal)})
	}

	// 3) No orphan journal entries (no lines)
	var orphanCount int64
	if err := uc.db.WithContext(ctx).
		Raw(`
		SELECT COUNT(*) FROM journal_entries je
		WHERE je.entry_date >= ? AND je.entry_date <= ?
		AND NOT EXISTS (SELECT 1 FROM journal_lines jl WHERE jl.journal_entry_id = je.id)
	`, start, end).
		Scan(&orphanCount).Error; err != nil {
		return nil, err
	}
	if orphanCount == 0 {
		results = append(results, dto.FinancialClosingValidationResult{Name: "orphan_entries", Passed: true, Message: "No orphan journal entries found"})
	} else {
		results = append(results, dto.FinancialClosingValidationResult{Name: "orphan_entries", Passed: false, Message: fmt.Sprintf("Found %d journal entries without lines", orphanCount)})
	}

	// 4) No reversed journal entries in period
	var reversedCount int64
	if err := uc.db.WithContext(ctx).
		Model(&financeModels.JournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ?", start, end).
		Where("status = ?", financeModels.JournalStatusReversed).
		Count(&reversedCount).Error; err != nil {
		return nil, err
	}
	if reversedCount == 0 {
		results = append(results, dto.FinancialClosingValidationResult{Name: "reversed_entries", Passed: true, Message: "No reversed journal entries in period"})
	} else {
		results = append(results, dto.FinancialClosingValidationResult{Name: "reversed_entries", Passed: false, Message: fmt.Sprintf("Found %d reversed journal entries", reversedCount)})
	}

	return results, nil
}

func (uc *financialClosingUsecase) createPeriodClosingJournal(ctx context.Context, start, end time.Time, periodEndDate time.Time) (float64, error) {
	// Find retained earnings account
	retainedCoA, err := uc.coaRepo.FindByCode(ctx, COACodeRetainedEarnings)
	if err != nil {
		return 0, fmt.Errorf("retained earnings account (code %s) not found: %w", COACodeRetainedEarnings, err)
	}

	// Get all Revenue and Expense COAs, so we can close them.
	allCOAs, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return 0, err
	}

	revenueIDs := make(map[string]bool)
	expenseIDs := make(map[string]bool)
	for _, coa := range allCOAs {
		t := strings.ToUpper(string(coa.Type))
		if t == "REVENUE" {
			revenueIDs[coa.ID] = true
		} else if t == "EXPENSE" || t == "COST_OF_GOODS_SOLD" || t == "SALARY_WAGES" || t == "OPERATIONAL" {
			expenseIDs[coa.ID] = true
		}
	}

	type aggRow struct {
		ChartOfAccountID string
		DebitTotal       float64
		CreditTotal      float64
	}

	var rows []aggRow
	if err := uc.db.WithContext(ctx).
		Table("journal_lines").
		Select("journal_lines.chart_of_account_id as chart_of_account_id, COALESCE(SUM(journal_lines.debit),0) as debit_total, COALESCE(SUM(journal_lines.credit),0) as credit_total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.status = ?", financeModels.JournalStatusPosted).
		Where("journal_entries.entry_date >= ? AND journal_entries.entry_date <= ?", start, end).
		Group("journal_lines.chart_of_account_id").
		Scan(&rows).Error; err != nil {
		return 0, err
	}

	var lines []dto.JournalLineRequest
	var totalRevenue, totalExpense float64
	for _, r := range rows {
		if revenueIDs[r.ChartOfAccountID] {
			balance := r.CreditTotal - r.DebitTotal
			if balance == 0 {
				continue
			}
			totalRevenue += balance
			if balance > 0 {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            balance,
					Credit:           0,
					Memo:             fmt.Sprintf("Period closing %s — close revenue", periodEndDate.Format("2006-01-02")),
				})
			} else {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            0,
					Credit:           -balance,
					Memo:             fmt.Sprintf("Period closing %s — close revenue (deficit)", periodEndDate.Format("2006-01-02")),
				})
			}
		} else if expenseIDs[r.ChartOfAccountID] {
			balance := r.DebitTotal - r.CreditTotal
			if balance == 0 {
				continue
			}
			totalExpense += balance
			if balance > 0 {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            0,
					Credit:           balance,
					Memo:             fmt.Sprintf("Period closing %s — close expense", periodEndDate.Format("2006-01-02")),
				})
			} else {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            -balance,
					Credit:           0,
					Memo:             fmt.Sprintf("Period closing %s — close expense (surplus)", periodEndDate.Format("2006-01-02")),
				})
			}
		}
	}

	netProfit := totalRevenue - totalExpense
	if netProfit > 0 {
		lines = append(lines, dto.JournalLineRequest{
			ChartOfAccountID: retainedCoA.ID,
			Debit:            0,
			Credit:           netProfit,
			Memo:             fmt.Sprintf("Period closing %s — net income to retained earnings", periodEndDate.Format("2006-01-02")),
		})
	} else if netProfit < 0 {
		lines = append(lines, dto.JournalLineRequest{
			ChartOfAccountID: retainedCoA.ID,
			Debit:            -netProfit,
			Credit:           0,
			Memo:             fmt.Sprintf("Period closing %s — net loss to retained earnings", periodEndDate.Format("2006-01-02")),
		})
	}

	if len(lines) == 0 {
		// Non-fatal: closing can continue even when there is no P&L movement in period.
		return 0, nil
	}

	// Create or update a single closing journal entry for this period.
	refType := "PERIOD_CLOSING"
	refID := uuid.NewSHA1(periodClosingNamespace, []byte(periodEndDate.Format("2006-01-02"))).String()
	journalReq := &dto.CreateJournalEntryRequest{
		EntryDate:         periodEndDate.Format("2006-01-02"),
		Description:       fmt.Sprintf("Period closing for %s", periodEndDate.Format("2006-01-02")),
		ReferenceType:     &refType,
		ReferenceID:       &refID,
		Lines:             lines,
		IsSystemGenerated: true,
	}
	_, err = uc.journalUC.PostOrUpdateJournal(ctx, journalReq)
	if err != nil {
		return 0, fmt.Errorf("failed to create period closing journal: %w", err)
	}

	return netProfit, nil
}

func (uc *financialClosingUsecase) createSnapshot(ctx context.Context, start, end time.Time, netProfit float64, periodEnd time.Time, createdBy string) (*financeModels.FinancialClosingSnapshot, error) {
	// Calculate retained earnings balance at end of period
	retainedCoA, err := uc.coaRepo.FindByCode(ctx, COACodeRetainedEarnings)
	if err != nil {
		return nil, fmt.Errorf("retained earnings account (code %s) not found: %w", COACodeRetainedEarnings, err)
	}
	balances, err := uc.calculateBalances(ctx, nil, &end)
	if err != nil {
		return nil, err
	}
	retainedBalance := balances[retainedCoA.ID]

	snapshotData := map[string]any{
		"net_profit":                netProfit,
		"retained_earnings_balance": retainedBalance,
		"period_start":              start.Format("2006-01-02"),
		"period_end":                end.Format("2006-01-02"),
	}
	snapshotJSON, _ := json.Marshal(snapshotData)

	snapshot := &financeModels.FinancialClosingSnapshot{
		PeriodEndDate:          periodEnd,
		NetProfit:              netProfit,
		RetainedEarningsBalance: retainedBalance,
		SnapshotJSON:           string(snapshotJSON),
		CreatedBy:              &createdBy,
	}
	if err := uc.snapshotRepo.Create(ctx, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (uc *financialClosingUsecase) logAction(ctx context.Context, periodEnd time.Time, action financeModels.FinancialClosingLogAction, reason string) error {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	logItem := &financeModels.FinancialClosingLog{
		PeriodEndDate: periodEnd,
		Action:        action,
		Reason:        reason,
		CreatedBy:     &actorID,
	}
	return uc.logRepo.Create(ctx, logItem)
}

func (uc *financialClosingUsecase) Create(ctx context.Context, req *dto.CreateFinancialClosingRequest) (*dto.FinancialClosingResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	periodEnd, err := time.Parse("2006-01-02", strings.TrimSpace(req.PeriodEndDate))
	if err != nil {
		return nil, errors.New("invalid period_end_date")
	}

	item := &financeModels.FinancialClosing{
		PeriodEndDate: periodEnd,
		Status:        financeModels.FinancialClosingStatusDraft,
		Notes:         strings.TrimSpace(req.Notes),
		CreatedBy:     &actorID,
	}
	if err := uc.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	// Create an open accounting period spanning from the last closed period to this period end.
	start, end, err := uc.getClosingPeriodRange(ctx, item.PeriodEndDate)
	if err == nil {
		periodName := fmt.Sprintf("%04d-%02d", item.PeriodEndDate.Year(), item.PeriodEndDate.Month())
		period := &financeModels.AccountingPeriod{
			PeriodName: periodName,
			StartDate:  start,
			EndDate:    end,
			Status:     financeModels.AccountingPeriodStatusOpen,
		}
		_ = uc.accountingRepo.Create(ctx, period)
	}

	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *financialClosingUsecase) Approve(ctx context.Context, id string) (*dto.FinancialClosingResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFinancialClosingNotFound
		}
		return nil, err
	}
	if item.Status == financeModels.FinancialClosingStatusApproved {
		res := uc.mapper.ToResponse(item)
		return &res, nil
	}

	// disallow approving older/equal periods than latest approved
	latest, err := uc.repo.LatestApproved(ctx)
	if err == nil {
		if !item.PeriodEndDate.After(latest.PeriodEndDate) {
			return nil, errors.New("cannot approve closing period on/before latest approved period")
		}
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Determine the period range we are closing
	start, end, err := uc.getClosingPeriodRange(ctx, item.PeriodEndDate)
	if err != nil {
		return nil, err
	}

	validations, err := uc.validateClosingPeriod(ctx, start, end)
	if err != nil {
		return nil, err
	}
	failed := make([]string, 0)
	for _, v := range validations {
		if !v.Passed {
			failed = append(failed, fmt.Sprintf("%s: %s", v.Name, v.Message))
		}
	}
	if len(failed) > 0 {
		// Log validation attempt
		_ = uc.logAction(ctx, item.PeriodEndDate, financeModels.FinancialClosingLogActionValidate, strings.Join(failed, "; "))
		return nil, fmt.Errorf("validation failed: %s", strings.Join(failed, "; "))
	}

	// Create closing journal and snapshot (net profit -> retained earnings)
	netProfit, err := uc.createPeriodClosingJournal(ctx, start, end, item.PeriodEndDate)
	if err != nil {
		return nil, err
	}

	if _, err := uc.createSnapshot(ctx, start, end, netProfit, item.PeriodEndDate, actorID); err != nil {
		return nil, err
	}

	// Lock the accounting period
	now := apptime.Now()
	period, err := uc.accountingRepo.FindByDate(ctx, item.PeriodEndDate)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		periodName := fmt.Sprintf("%04d-%02d", item.PeriodEndDate.Year(), item.PeriodEndDate.Month())
		period = &financeModels.AccountingPeriod{
			PeriodName: periodName,
			StartDate:  start,
			EndDate:    end,
			Status:     financeModels.AccountingPeriodStatusClosed,
			LockedAt:   &now,
			LockedBy:   &actorID,
		}
		if err := uc.accountingRepo.Create(ctx, period); err != nil {
			return nil, err
		}
	} else {
		if err := uc.accountingRepo.UpdateStatus(ctx, period.ID, financeModels.AccountingPeriodStatusClosed, &now, &actorID); err != nil {
			return nil, err
		}
	}

	_ = uc.logAction(ctx, item.PeriodEndDate, financeModels.FinancialClosingLogActionClose, "approved")

	now = apptime.Now()
	if err := uc.db.WithContext(ctx).Model(&financeModels.FinancialClosing{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      financeModels.FinancialClosingStatusApproved,
		"approved_at": now,
		"approved_by": actorID,
	}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

func (uc *financialClosingUsecase) GetByID(ctx context.Context, id string) (*dto.FinancialClosingResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &financeModels.FinancialClosing{}, id, security.FinanceScopeQueryOptions()) {
		return nil, ErrFinancialClosingNotFound
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFinancialClosingNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *financialClosingUsecase) List(ctx context.Context, req *dto.ListFinancialClosingsRequest) ([]dto.FinancialClosingResponse, int64, error) {
	if req == nil {
		req = &dto.ListFinancialClosingsRequest{}
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

	items, total, err := uc.repo.List(ctx, repositories.FinancialClosingListParams{
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.FinancialClosingResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}

func (uc *financialClosingUsecase) GetAnalysis(ctx context.Context, id string) (*dto.FinancialClosingAnalysisResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Determine the period range for this closing
	start, end, err := uc.getClosingPeriodRange(ctx, item.PeriodEndDate)
	if err != nil {
		return nil, err
	}

	validations, _ := uc.validateClosingPeriod(ctx, start, end)
	_ = uc.logAction(ctx, item.PeriodEndDate, financeModels.FinancialClosingLogActionValidate, "analysis_requested")

	// Calculate Opening Balance (Start of Year)
	startOfYear := time.Date(item.PeriodEndDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	// Balances up to Period End Date
	closingBalances, err := uc.calculateBalances(ctx, nil, &item.PeriodEndDate)
	if err != nil {
		return nil, err
	}

	// Balances before Start of Year (to get Opening Balance)
	// Actually "Saldo Awal Tahun" as shown in UI usually means the balance at Jan 1st (including Jan 1st transactions maybe? No, usually before Jan 1st).
	// But in standard ERP, Saldo Awal Tahun is balance as of 01-01 00:00.
	openingDate := startOfYear.AddDate(0, 0, -1) // Balance as of Dec 31 previous year
	openingBalances, err := uc.calculateBalances(ctx, nil, &openingDate)
	if err != nil {
		return nil, err
	}

	coas := []struct {
		ID   string
		Code string
		Name string
	}{}
	if err := uc.db.WithContext(ctx).Table("chart_of_accounts").Select("id, code, name").Order("code ASC").Scan(&coas).Error; err != nil {
		return nil, err
	}

	rows := make([]dto.FinancialClosingAnalysisRow, 0, len(coas))
	for _, c := range coas {
		cb := closingBalances[c.ID]
		ob := openingBalances[c.ID]
		rows = append(rows, dto.FinancialClosingAnalysisRow{
			AccountID:      c.ID,
			AccountCode:    c.Code,
			AccountName:    c.Name,
			ClosingBalance: cb,
			OpeningBalance: ob,
			Difference:     cb - ob,
		})
	}

	var snapshot *dto.FinancialClosingSnapshotResponse
	if snap, err := uc.snapshotRepo.FindByPeriodEndDate(ctx, item.PeriodEndDate.Format("2006-01-02")); err == nil {
		snapshot = &dto.FinancialClosingSnapshotResponse{
			NetProfit:               snap.NetProfit,
			RetainedEarningsBalance: snap.RetainedEarningsBalance,
			PeriodEndDate:           snap.PeriodEndDate.Format("2006-01-02"),
			SnapshotJSON:            snap.SnapshotJSON,
		}
	}

	return &dto.FinancialClosingAnalysisResponse{
		Closing:     uc.mapper.ToResponse(item),
		Rows:        rows,
		Validations: validations,
		Snapshot:    snapshot,
	}, nil
}

func (uc *financialClosingUsecase) calculateBalances(ctx context.Context, start, end *time.Time) (map[string]float64, error) {
	type aggRow struct {
		ChartOfAccountID string
		DebitTotal       float64
		CreditTotal      float64
	}

	q := uc.db.WithContext(ctx).
		Table("journal_lines").
		Select("journal_lines.chart_of_account_id as chart_of_account_id, COALESCE(SUM(journal_lines.debit),0) as debit_total, COALESCE(SUM(journal_lines.credit),0) as credit_total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.status = ?", "posted")

	if start != nil {
		q = q.Where("journal_entries.entry_date >= ?", *start)
	}
	if end != nil {
		q = q.Where("journal_entries.entry_date <= ?", *end)
	}
	q = q.Group("journal_lines.chart_of_account_id")

	var results []aggRow
	if err := q.Scan(&results).Error; err != nil {
		return nil, err
	}

	// We need to know the account type to determine if Balance is Debit - Credit or Credit - Debit.
	// For simplicity, let's just return Debit - Credit for now, or fetch types.
	// In the UI "Saldo Akhir Akun" is usually Absolute or type-aware.

	coas := []struct {
		ID   string
		Type string
	}{}
	if err := uc.db.WithContext(ctx).Table("chart_of_accounts").Select("id, type").Scan(&coas).Error; err != nil {
		return nil, err
	}
	coaTypes := make(map[string]string)
	for _, c := range coas {
		coaTypes[c.ID] = c.Type
	}

	balances := make(map[string]float64)
	for _, r := range results {
		actorType := strings.ToLower(coaTypes[r.ChartOfAccountID])
		balance := r.DebitTotal - r.CreditTotal
		if actorType == "liability" || actorType == "equity" || actorType == "revenue" {
			balance = r.CreditTotal - r.DebitTotal
		}
		balances[r.ChartOfAccountID] = balance
	}

	return balances, nil
}

// Reopen reverts an approved financial closing back to draft.
// This allows corrections to journal entries in a previously closed period.
func (uc *financialClosingUsecase) Reopen(ctx context.Context, id string) (*dto.FinancialClosingResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFinancialClosingNotFound
		}
		return nil, err
	}

	if item.Status != financeModels.FinancialClosingStatusApproved {
		return nil, errors.New("only approved financial closings can be reopened")
	}

	// Prevent reopening if there is a later approved closing
	// (that would leave a gap in the timeline)
	latest, err := uc.repo.LatestApproved(ctx)
	if err == nil && latest.ID != item.ID {
		if latest.PeriodEndDate.After(item.PeriodEndDate) {
			return nil, errors.New("cannot reopen this period — a later period is already closed; reopen the latest first")
		}
	}

	// Reverse the closing journal entry (if it exists)
	refID := uuid.NewSHA1(periodClosingNamespace, []byte(item.PeriodEndDate.Format("2006-01-02"))).String()
	var closingEntry financeModels.JournalEntry
	if err := uc.db.WithContext(ctx).Where("reference_type = ? AND reference_id = ?", "PERIOD_CLOSING", refID).First(&closingEntry).Error; err == nil {
		// Only reverse posted entries
		if closingEntry.Status == financeModels.JournalStatusPosted {
			_, _ = uc.journalUC.Reverse(ctx, closingEntry.ID)
		}
	}

	// Unlock the accounting period (if any)
	period, err := uc.accountingRepo.FindByDate(ctx, item.PeriodEndDate)
	if err == nil {
		_ = uc.accountingRepo.UpdateStatus(ctx, period.ID, financeModels.AccountingPeriodStatusOpen, nil, nil)
	}

	_ = uc.logAction(ctx, item.PeriodEndDate, financeModels.FinancialClosingLogActionReopen, "reopened")

	if err := uc.db.WithContext(ctx).Model(&financeModels.FinancialClosing{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      financeModels.FinancialClosingStatusDraft,
		"approved_at": nil,
		"approved_by": nil,
	}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

// YearEndClose performs the fiscal year-end closing process:
// 1. Calculates net income (Revenue − Expense) for the fiscal year
// 2. Creates a closing journal entry that zeros out Revenue & Expense accounts
// 3. Transfers net income to Retained Earnings
// 4. Creates and approves a financial closing record for Dec 31 of the year
func (uc *financialClosingUsecase) YearEndClose(ctx context.Context, req *dto.YearEndCloseRequest) (*dto.FinancialClosingResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	year := req.FiscalYear
	if year < 2000 || year > 2100 {
		return nil, errors.New("invalid fiscal year")
	}

	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	// Look up retained earnings account by well-known COA code
	retainedCoA, err := uc.coaRepo.FindByCode(ctx, COACodeRetainedEarnings)
	if err != nil {
		return nil, fmt.Errorf("retained earnings account (code %s) not found: %w", COACodeRetainedEarnings, err)
	}

	// Get all Revenue and Expense COAs
	allCOAs, err := uc.coaRepo.FindAll(ctx, true)
	if err != nil {
		return nil, err
	}

	revenueIDs := make(map[string]bool)
	expenseIDs := make(map[string]bool)
	for _, coa := range allCOAs {
		t := strings.ToUpper(string(coa.Type))
		if t == "REVENUE" {
			revenueIDs[coa.ID] = true
		} else if t == "EXPENSE" || t == "COST_OF_GOODS_SOLD" || t == "SALARY_WAGES" || t == "OPERATIONAL" {
			expenseIDs[coa.ID] = true
		}
	}

	// Calculate balances for Revenue & Expense accounts during the fiscal year
	type aggRow struct {
		ChartOfAccountID string
		DebitTotal       float64
		CreditTotal      float64
	}

	var rows []aggRow
	if err := uc.db.WithContext(ctx).
		Table("journal_lines").
		Select("journal_lines.chart_of_account_id as chart_of_account_id, COALESCE(SUM(journal_lines.debit),0) as debit_total, COALESCE(SUM(journal_lines.credit),0) as credit_total").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.status = ?", financeModels.JournalStatusPosted).
		Where("journal_entries.entry_date >= ? AND journal_entries.entry_date <= ?", startOfYear, endOfYear).
		Group("journal_lines.chart_of_account_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Build closing journal lines — zero out each Revenue/Expense account
	var lines []dto.JournalLineRequest
	var totalRevenue, totalExpense float64

	for _, r := range rows {
		if revenueIDs[r.ChartOfAccountID] {
			// Revenue accounts have credit-normal balance: close by debiting
			balance := r.CreditTotal - r.DebitTotal
			if balance == 0 {
				continue
			}
			totalRevenue += balance
			if balance > 0 {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            balance,
					Credit:           0,
					Memo:             fmt.Sprintf("Year-end closing FY%d — close revenue", year),
				})
			} else {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            0,
					Credit:           -balance,
					Memo:             fmt.Sprintf("Year-end closing FY%d — close revenue (deficit)", year),
				})
			}
		} else if expenseIDs[r.ChartOfAccountID] {
			// Expense accounts have debit-normal balance: close by crediting
			balance := r.DebitTotal - r.CreditTotal
			if balance == 0 {
				continue
			}
			totalExpense += balance
			if balance > 0 {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            0,
					Credit:           balance,
					Memo:             fmt.Sprintf("Year-end closing FY%d — close expense", year),
				})
			} else {
				lines = append(lines, dto.JournalLineRequest{
					ChartOfAccountID: r.ChartOfAccountID,
					Debit:            -balance,
					Credit:           0,
					Memo:             fmt.Sprintf("Year-end closing FY%d — close expense (surplus)", year),
				})
			}
		}
	}

	netIncome := totalRevenue - totalExpense
	if len(lines) > 0 {
		if netIncome > 0 {
			// Profit → Credit Retained Earnings
			lines = append(lines, dto.JournalLineRequest{
				ChartOfAccountID: retainedCoA.ID,
				Debit:            0,
				Credit:           netIncome,
				Memo:             fmt.Sprintf("Year-end closing FY%d — net income to retained earnings", year),
			})
		} else if netIncome < 0 {
			// Loss → Debit Retained Earnings
			lines = append(lines, dto.JournalLineRequest{
				ChartOfAccountID: retainedCoA.ID,
				Debit:            -netIncome,
				Credit:           0,
				Memo:             fmt.Sprintf("Year-end closing FY%d — net loss to retained earnings", year),
			})
		}

		// Create the closing journal entry via PostOrUpdateJournal (auto-posts).
		// reference_id must be a UUID — use a deterministic UUID v5 derived from the
		// fiscal year so that re-running year-end close for the same year performs an
		// upsert rather than creating a duplicate.
		var yearEndClosingNS = uuid.MustParse("23d6b1a2-0000-0000-0000-000000000000")
		refType := "year_end_closing"
		refIDStr := uuid.NewSHA1(yearEndClosingNS, []byte(fmt.Sprintf("%d", year))).String()
		journalReq := &dto.CreateJournalEntryRequest{
			EntryDate:         endOfYear.Format("2006-01-02"),
			Description:       fmt.Sprintf("Year-End Closing Journal FY%d", year),
			ReferenceType:     &refType,
			ReferenceID:       &refIDStr,
			Lines:             lines,
			IsSystemGenerated: true,
		}

		_, err = uc.journalUC.PostOrUpdateJournal(ctx, journalReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create year-end closing journal: %w", err)
		}
	}

	// Create and auto-approve a financial closing record for Dec 31
	closingItem := &financeModels.FinancialClosing{
		PeriodEndDate: time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:        financeModels.FinancialClosingStatusApproved,
		Notes:         fmt.Sprintf("Year-end closing for FY%d. Net income: %.2f", year, netIncome),
		CreatedBy:     &actorID,
	}
	now := apptime.Now()
	closingItem.ApprovedAt = &now
	closingItem.ApprovedBy = &actorID

	if err := uc.db.WithContext(ctx).Create(closingItem).Error; err != nil {
		return nil, fmt.Errorf("failed to create year-end closing record: %w", err)
	}

	res := uc.mapper.ToResponse(closingItem)
	return &res, nil
}
func (uc *financialClosingUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}

	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrFinancialClosingNotFound
		}
		return err
	}

	if item.Status != financeModels.FinancialClosingStatusDraft {
		return errors.New("only draft financial closings can be deleted")
	}

	return uc.repo.Delete(ctx, id)
}
