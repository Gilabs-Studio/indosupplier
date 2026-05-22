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
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"gorm.io/gorm"
)

const referenceTypeSalesDPApplication = "SALES_DP_APPLICATION"

var profileSalesDelivery = accounting.PostingProfile{
	ReferenceType:       reference.RefTypeDeliveryOrder,
	DescriptionTemplate: "Delivery Order %s: %s",
	Rules: []accounting.PostingRule{
		{
			COASettingKey: "coa.sales_cogs",
			Side:          "debit",
			AmountSource:  "cogs_total",
			MemoTemplate:  "COGS from delivery",
		},
		{
			COASettingKey: "coa.sales_inventory",
			Side:          "credit",
			AmountSource:  "cogs_total",
			MemoTemplate:  "Inventory release from delivery",
		},
	},
}

var profileSalesDPApplication = accounting.PostingProfile{
	ReferenceType:       referenceTypeSalesDPApplication,
	DescriptionTemplate: "Apply Sales DP %s: %s",
	Rules: []accounting.PostingRule{
		{
			COASettingKey: "coa.sales_advance",
			Side:          "debit",
			AmountSource:  "total",
			MemoTemplate:  "Sales Down Payment Application",
		},
		{
			COASettingKey: "coa.sales_receivable",
			Side:          "credit",
			AmountSource:  "total",
			MemoTemplate:  "Trade Receivables Reduction from DP",
		},
	},
}

// SalesJournalService centralizes sales module posting and reversal logic.
type SalesJournalService interface {
	GenerateSalesJournal(ctx context.Context, deliveryOrder *salesModels.DeliveryOrder) (*financeModels.JournalEntry, error)
	GenerateDPJournal(ctx context.Context, downPaymentInvoice *salesModels.CustomerInvoice) (*financeModels.JournalEntry, error)
	GenerateDPApplicationJournal(ctx context.Context, invoiceID string) (*financeModels.JournalEntry, error)
	ReverseSalesJournal(ctx context.Context, referenceType, referenceID, reason string) error
}

type salesJournalService struct {
	db        *gorm.DB
	journalUC finUsecase.JournalEntryUsecase
	engine    accounting.AccountingEngine
}

func NewSalesJournalService(db *gorm.DB, journalUC finUsecase.JournalEntryUsecase, engine accounting.AccountingEngine) SalesJournalService {
	return &salesJournalService{
		db:        db,
		journalUC: journalUC,
		engine:    engine,
	}
}

func (s *salesJournalService) GenerateSalesJournal(ctx context.Context, deliveryOrder *salesModels.DeliveryOrder) (*financeModels.JournalEntry, error) {
	if deliveryOrder == nil {
		return nil, errors.New("delivery order is required")
	}

	companyID, err := s.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return nil, err
	}

	cogsTotal := 0.0
	for _, item := range deliveryOrder.Items {
		if item.COGSAmount > 0 {
			cogsTotal += item.COGSAmount
		}
	}

	total := round2(cogsTotal)

	data := accounting.TransactionData{
		ReferenceType:   reference.RefTypeDeliveryOrder,
		ReferenceID:     strings.TrimSpace(deliveryOrder.ID),
		CompanyID:       companyID,
		EntryDate:       deliveryOrder.DeliveryDate.Format("2006-01-02"),
		Description:     fmt.Sprintf("Delivery Order %s", deliveryOrder.Code),
		TotalAmount:     total,
		COGSTotal:       round2(cogsTotal),
		DescriptionArgs: []interface{}{deliveryOrder.Code, deliveryOrder.Code},
	}

	return s.generateAndPost(ctx, profileSalesDelivery, data)
}

func (s *salesJournalService) GenerateDPJournal(ctx context.Context, downPaymentInvoice *salesModels.CustomerInvoice) (*financeModels.JournalEntry, error) {
	if downPaymentInvoice == nil {
		return nil, errors.New("down payment invoice is required")
	}

	if downPaymentInvoice.Amount <= 0 {
		return nil, nil
	}

	companyID, err := s.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return nil, err
	}

	data := accounting.TransactionData{
		ReferenceType:   reference.RefTypeSalesInvoiceDP,
		ReferenceID:     downPaymentInvoice.ID,
		CompanyID:       companyID,
		EntryDate:       downPaymentInvoice.InvoiceDate.Format("2006-01-02"),
		Description:     fmt.Sprintf("Customer DP Invoice %s", downPaymentInvoice.Code),
		TotalAmount:     downPaymentInvoice.Amount,
		DescriptionArgs: []interface{}{safeInvoiceNumber(downPaymentInvoice.InvoiceNumber), downPaymentInvoice.Code},
	}

	return s.generateAndPost(ctx, accounting.ProfileSalesInvoiceDP, data)
}

func (s *salesJournalService) GenerateDPApplicationJournal(ctx context.Context, invoiceID string) (*financeModels.JournalEntry, error) {
	invoiceID = strings.TrimSpace(invoiceID)
	if invoiceID == "" {
		return nil, errors.New("invoice id is required")
	}

	if s.db == nil {
		return nil, errors.New("db is not configured")
	}

	var invoice salesModels.CustomerInvoice
	err := database.GetDB(ctx, s.db).
		Where("id = ?", invoiceID).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}

	if invoice.DownPaymentAmount <= 0 {
		return nil, nil
	}

	companyID, err := s.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return nil, err
	}

	data := accounting.TransactionData{
		ReferenceType:   referenceTypeSalesDPApplication,
		ReferenceID:     invoice.ID,
		CompanyID:       companyID,
		EntryDate:       invoice.InvoiceDate.Format("2006-01-02"),
		Description:     fmt.Sprintf("Sales DP Application %s", invoice.Code),
		TotalAmount:     round2(invoice.DownPaymentAmount),
		DescriptionArgs: []interface{}{invoice.Code, safeInvoiceNumber(invoice.InvoiceNumber)},
	}

	return s.generateAndPost(ctx, profileSalesDPApplication, data)
}

func (s *salesJournalService) ReverseSalesJournal(ctx context.Context, referenceType, referenceID, reason string) error {
	if s == nil || s.db == nil || s.journalUC == nil {
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
		return fmt.Errorf("failed to find sales journal: %w", err)
	}

	reversalReason := strings.TrimSpace(reason)
	if reversalReason == "" {
		reversalReason = "Manual reversal"
	}

	_, err = s.journalUC.ReverseWithReason(finUsecase.WithReversalFlag(ctx), existing.ID, reversalReason)
	if err != nil {
		return fmt.Errorf("failed to reverse sales journal: %w", err)
	}

	return nil
}

func (s *salesJournalService) generateAndPost(
	ctx context.Context,
	profile accounting.PostingProfile,
	data accounting.TransactionData,
) (*financeModels.JournalEntry, error) {
	if s == nil || s.journalUC == nil || s.engine == nil {
		return nil, nil
	}

	if strings.TrimSpace(data.ReferenceType) == "" {
		data.ReferenceType = profile.ReferenceType
	}
	if strings.TrimSpace(data.ReferenceID) == "" {
		return nil, errors.New("reference id is required")
	}

	req, err := s.engine.GenerateJournal(ctx, profile, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate journal: %w", err)
	}

	if !isJournalBalanced(req) {
		return nil, errors.New("generated journal is unbalanced")
	}

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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func safeInvoiceNumber(invoiceNumber *string) string {
	if invoiceNumber == nil || strings.TrimSpace(*invoiceNumber) == "" {
		return "-"
	}

	return strings.TrimSpace(*invoiceNumber)
}

func (s *salesJournalService) resolveCompanyIDFromActor(ctx context.Context) (string, error) {
	if s == nil || s.db == nil {
		return "", errors.New("db is not configured")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return "", errors.New("user not authenticated")
	}

	var companyID string
	if err := database.GetDB(ctx, s.db).
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
