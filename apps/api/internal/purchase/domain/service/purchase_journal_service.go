package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"gorm.io/gorm"
)

var profileSupplierDPApplication = accounting.PostingProfile{
	ReferenceType:       reference.RefTypeSupplierInvoice,
	DescriptionTemplate: "Apply Supplier DP %s: %s",
	Rules: []accounting.PostingRule{
		{
			COASettingKey: "coa.purchase_payable",
			Side:          "debit",
			AmountSource:  "total",
			MemoTemplate:  "Supplier DP Application",
		},
		{
			COASettingKey: "coa.purchase_advance",
			Side:          "credit",
			AmountSource:  "total",
			MemoTemplate:  "Supplier DP Application",
		},
	},
}

// PurchaseJournalTxn wraps posting profile and transaction payload.
type PurchaseJournalTxn struct {
	Profile accounting.PostingProfile
	Data    accounting.TransactionData
}

// PurchaseJournalService centralizes purchase module journal posting and reversal logic.
type PurchaseJournalService interface {
	GeneratePurchaseJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error)
	GenerateDPJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error)
	GenerateDPApplicationJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error)
	ReversePurchaseJournal(ctx context.Context, referenceType, referenceID, reason string) error
}

type purchaseJournalService struct {
	db        *gorm.DB
	journalUC finUsecase.JournalEntryUsecase
	engine    accounting.AccountingEngine
}

func NewPurchaseJournalService(db *gorm.DB, journalUC finUsecase.JournalEntryUsecase, engine accounting.AccountingEngine) PurchaseJournalService {
	return &purchaseJournalService{
		db:        db,
		journalUC: journalUC,
		engine:    engine,
	}
}

func (s *purchaseJournalService) GeneratePurchaseJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error) {
	if strings.TrimSpace(txn.Profile.ReferenceType) == "" {
		return nil, errors.New("posting profile reference type is required")
	}
	return s.generateAndPost(ctx, txn.Profile, txn.Data)
}

func (s *purchaseJournalService) GenerateDPJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error) {
	profile := txn.Profile
	if strings.TrimSpace(profile.ReferenceType) == "" {
		profile = accounting.ProfileSupplierInvoiceDP
	}
	return s.generateAndPost(ctx, profile, txn.Data)
}

func (s *purchaseJournalService) GenerateDPApplicationJournal(ctx context.Context, txn PurchaseJournalTxn) (*financeModels.JournalEntry, error) {
	profile := txn.Profile
	if strings.TrimSpace(profile.ReferenceType) == "" {
		profile = profileSupplierDPApplication
	}
	return s.generateAndPost(ctx, profile, txn.Data)
}

func (s *purchaseJournalService) ReversePurchaseJournal(ctx context.Context, referenceType, referenceID, reason string) error {
	if s == nil || s.journalUC == nil || s.db == nil {
		return nil
	}

	refType := reference.Normalize(referenceType)
	refID := strings.TrimSpace(referenceID)
	if refType == "" || refID == "" {
		return nil
	}

	var existing financeModels.JournalEntry
	err := database.GetDB(ctx, s.db).
		Where("reference_type = ? AND reference_id = ?", refType, refID).
		Where("status = ?", financeModels.JournalStatusPosted).
		Order("created_at DESC").
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find purchase journal: %w", err)
	}

	reversalReason := strings.TrimSpace(reason)
	if reversalReason == "" {
		reversalReason = "Manual reversal"
	}

	revCtx := finUsecase.WithReversalFlag(ctx)
	if _, err := s.journalUC.ReverseWithReason(revCtx, existing.ID, reversalReason); err != nil {
		return fmt.Errorf("failed to reverse purchase journal: %w", err)
	}

	return nil
}

func (s *purchaseJournalService) generateAndPost(
	ctx context.Context,
	profile accounting.PostingProfile,
	data accounting.TransactionData,
) (*financeModels.JournalEntry, error) {
	if s == nil || s.journalUC == nil || s.engine == nil {
		return nil, nil
	}

	if strings.TrimSpace(profile.ReferenceType) == "" {
		return nil, errors.New("posting profile reference type is required")
	}

	if strings.TrimSpace(data.ReferenceType) == "" {
		data.ReferenceType = profile.ReferenceType
	}
	if strings.TrimSpace(data.ReferenceID) == "" {
		return nil, errors.New("reference_id is required")
	}

	req, err := s.engine.GenerateJournal(ctx, profile, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate journal: %w", err)
	}

	if !isJournalBalanced(req) {
		return nil, errors.New("generated journal is unbalanced")
	}

	req.CompanyID = strings.TrimSpace(req.CompanyID)
	if req.CompanyID == "" {
		return nil, errors.New("company_id is required")
	}

	if req.FiscalYearID != nil {
		trimmedFiscalYearID := strings.TrimSpace(*req.FiscalYearID)
		if trimmedFiscalYearID == "" {
			req.FiscalYearID = nil
		} else {
			req.FiscalYearID = &trimmedFiscalYearID
		}
	}

	req.IsSystemGenerated = true
	res, err := s.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to post journal: %w", err)
	}

	if res == nil || strings.TrimSpace(res.ID) == "" || s.db == nil {
		return nil, nil
	}

	var posted financeModels.JournalEntry
	if err := database.GetDB(ctx, s.db).First(&posted, "id = ?", res.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load posted journal: %w", err)
	}

	return &posted, nil
}

func isJournalBalanced(req *finDto.CreateJournalEntryRequest) bool {
	if req == nil {
		return false
	}

	var debitTotal float64
	var creditTotal float64
	for _, line := range req.Lines {
		debitTotal += line.Debit
		creditTotal += line.Credit
	}

	return math.Abs(debitTotal-creditTotal) <= 0.001
}
