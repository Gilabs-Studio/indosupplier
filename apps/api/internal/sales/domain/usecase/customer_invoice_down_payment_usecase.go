package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrCustomerInvoiceConflict = errors.New("customer invoice status conflict")
	ErrCustomerInvoiceInvalid  = errors.New("invalid operation for customer invoice")
	ErrInvalidStatus           = errors.New("invalid status")
)

const (
	customerInvoiceDPDateFormat = "2006-01-02"
	customerInvoiceDPByIDQuery  = "id = ?"
	errCustomerInvoiceDPDBNil   = "db is nil"
)

type CustomerInvoiceDownPaymentUsecase interface {
	AddData(ctx context.Context) (*dto.CustomerInvoiceDownPaymentAddResponse, error)
	List(ctx context.Context, params *dto.ListCustomerInvoicesRequest) ([]*dto.CustomerInvoiceDownPaymentListResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	Create(ctx context.Context, req *dto.CreateCustomerInvoiceDownPaymentRequest) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceDownPaymentRequest) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	Delete(ctx context.Context, id string) error
	Pending(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	Approve(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	Cancel(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
	TriggerJournalForInvoiceDP(ctx context.Context, invoice *models.CustomerInvoice) error
}

type customerInvoiceDownPaymentUsecase struct {
	db           *gorm.DB
	repo         repositories.CustomerInvoiceRepository
	soRepo       repositories.SalesOrderRepository
	auditService audit.AuditService
	journalUC    finUsecase.JournalEntryUsecase
	coaUC        finUsecase.ChartOfAccountUsecase
	engine       accounting.AccountingEngine
}

func NewCustomerInvoiceDownPaymentUsecase(db *gorm.DB, repo repositories.CustomerInvoiceRepository, soRepo repositories.SalesOrderRepository, auditService audit.AuditService, journalUC finUsecase.JournalEntryUsecase, coaUC finUsecase.ChartOfAccountUsecase, engine accounting.AccountingEngine) CustomerInvoiceDownPaymentUsecase {
	return &customerInvoiceDownPaymentUsecase{db: db, repo: repo, soRepo: soRepo, auditService: auditService, journalUC: journalUC, coaUC: coaUC, engine: engine}
}

func (uc *customerInvoiceDownPaymentUsecase) AddData(ctx context.Context) (*dto.CustomerInvoiceDownPaymentAddResponse, error) {
	req := &dto.ListSalesOrdersRequest{Status: string(models.SalesOrderStatusApproved), SortBy: "created_at", SortDir: "desc", PerPage: 100}
	sos, _, err := uc.soRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// Get IDs of SOs that already have an active down payment invoice
	var soIDsWithDP []string
	database.GetDB(ctx, uc.db).Model(&models.CustomerInvoice{}).
		Where("type = ? AND sales_order_id IS NOT NULL AND status NOT IN ?",
			models.CustomerInvoiceTypeDownPayment,
			[]string{"CANCELLED", "VOID"}).
		Distinct("sales_order_id").
		Pluck("sales_order_id", &soIDsWithDP)
	soIDsWithDPSet := make(map[string]struct{}, len(soIDsWithDP))
	for _, id := range soIDsWithDP {
		soIDsWithDPSet[id] = struct{}{}
	}

	// Get IDs of SOs that already have a paid regular invoice (#243)
	var soIDsPaid []string
	database.GetDB(ctx, uc.db).Model(&models.CustomerInvoice{}).
		Where("type = ? AND sales_order_id IS NOT NULL AND status = ?",
			models.CustomerInvoiceTypeRegular,
			models.CustomerInvoiceStatusPaid).
		Distinct("sales_order_id").
		Pluck("sales_order_id", &soIDsPaid)
	soIDsPaidSet := make(map[string]struct{}, len(soIDsPaid))
	for _, id := range soIDsPaid {
		soIDsPaidSet[id] = struct{}{}
	}

	soRes := make([]dto.CustomerInvoiceAddSalesOrder, 0, len(sos))
	for _, so := range sos {
		// Skip SOs that already have an active down payment invoice (#252)
		if _, hasDP := soIDsWithDPSet[so.ID]; hasDP {
			continue
		}
		// Skip SOs that already have a fully paid regular invoice (#243)
		if _, isPaid := soIDsPaidSet[so.ID]; isPaid {
			continue
		}
		items := make([]dto.CustomerInvoiceAddSalesOrderItem, 0, len(so.Items))
		for _, it := range so.Items {
			var prod *dto.CustomerInvoiceAddProductMini
			if it.Product != nil {
				img := ""
				if it.Product.ImageURL != nil {
					img = *it.Product.ImageURL
				}
				prod = &dto.CustomerInvoiceAddProductMini{ID: it.Product.ID, Name: it.Product.Name, Code: it.Product.Code, ImageURL: img}
			}
			items = append(items, dto.CustomerInvoiceAddSalesOrderItem{ID: it.ID, Product: prod, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal})
		}
		var cust *dto.CustomerInvoiceAddCustomerMini
		if so.Customer != nil {
			cust = &dto.CustomerInvoiceAddCustomerMini{ID: so.Customer.ID, Name: so.Customer.Name}
		}
		soRes = append(soRes, dto.CustomerInvoiceAddSalesOrder{ID: so.ID, Customer: cust, Code: so.Code, OrderDate: so.OrderDate, Status: string(so.Status), TotalAmount: so.TotalAmount, Items: items})
	}
	return &dto.CustomerInvoiceDownPaymentAddResponse{SalesOrders: soRes}, nil
}

func (uc *customerInvoiceDownPaymentUsecase) mapToDetail(ctx context.Context, ci *models.CustomerInvoice) *dto.CustomerInvoiceDownPaymentDetailResponse {
	n := ""
	if ci.Notes != "" {
		n = ci.Notes
	}
	var soDto *dto.CustomerInvoiceDownPaymentSalesOrder
	var custID string
	if ci.SalesOrder != nil {
		soDto = &dto.CustomerInvoiceDownPaymentSalesOrder{ID: ci.SalesOrder.ID, Code: ci.SalesOrder.Code}
		if ci.SalesOrder.CustomerID != nil {
			soDto.CustomerID = ci.SalesOrder.CustomerID
			custID = *ci.SalesOrder.CustomerID
		}
		if ci.SalesOrder.CustomerName != "" {
			name := ci.SalesOrder.CustomerName
			soDto.CustomerName = &name
		}
	}
	var dueDate *string
	if ci.DueDate != nil {
		d := ci.DueDate.Format(customerInvoiceDPDateFormat)
		dueDate = &d
	}
	salesOrderID := ""
	if ci.SalesOrderID != nil {
		salesOrderID = *ci.SalesOrderID
	}
	var relatedCode *string
	var regularInvoice models.CustomerInvoice
	if err := database.GetDB(ctx, uc.db).Where("down_payment_invoice_id = ?", ci.ID).Select("code").First(&regularInvoice).Error; err == nil {
		relatedCode = &regularInvoice.Code
	}

	return &dto.CustomerInvoiceDownPaymentDetailResponse{
		ID:                 ci.ID,
		SalesOrderID:       salesOrderID,
		SalesOrder:         soDto,
		CustomerID:         custID,
		Code:               ci.Code,
		RelatedInvoiceCode: relatedCode,
		InvoiceNumber:      ci.InvoiceNumber,
		InvoiceDate:        ci.InvoiceDate.Format(customerInvoiceDPDateFormat),
		DueDate:            dueDate,
		Amount:             ci.Amount,
		RemainingAmount:    ci.RemainingAmount,
		Status:             string(ci.Status),
		AttachmentURL:      ci.AttachmentURL,
		Notes:              &n,
		CreatedBy:          ci.CreatedBy,
		CreatedAt:          ci.CreatedAt,
		UpdatedAt:          ci.UpdatedAt,
	}
}

func (uc *customerInvoiceDownPaymentUsecase) mapToList(ctx context.Context, ci *models.CustomerInvoice) *dto.CustomerInvoiceDownPaymentListResponse {
	var soDto *dto.CustomerInvoiceDownPaymentSalesOrder
	if ci.SalesOrder != nil {
		soDto = &dto.CustomerInvoiceDownPaymentSalesOrder{ID: ci.SalesOrder.ID, Code: ci.SalesOrder.Code}
		if ci.SalesOrder.CustomerID != nil {
			soDto.CustomerID = ci.SalesOrder.CustomerID
		}
		if ci.SalesOrder.CustomerName != "" {
			name := ci.SalesOrder.CustomerName
			soDto.CustomerName = &name
		}
	}
	var dueDate *string
	if ci.DueDate != nil {
		d := ci.DueDate.Format(customerInvoiceDPDateFormat)
		dueDate = &d
	}
	salesOrderID := ""
	if ci.SalesOrderID != nil {
		salesOrderID = *ci.SalesOrderID
	}
	var relatedCode *string
	var regularInvoice models.CustomerInvoice
	if err := database.GetDB(ctx, uc.db).Where("down_payment_invoice_id = ?", ci.ID).Select("code").First(&regularInvoice).Error; err == nil {
		relatedCode = &regularInvoice.Code
	}

	return &dto.CustomerInvoiceDownPaymentListResponse{
		ID:                 ci.ID,
		SalesOrderID:       salesOrderID,
		SalesOrder:         soDto,
		Code:               ci.Code,
		RelatedInvoiceCode: relatedCode,
		InvoiceNumber:      ci.InvoiceNumber,
		InvoiceDate:        ci.InvoiceDate.Format(customerInvoiceDPDateFormat),
		DueDate:            dueDate,
		Amount:             ci.Amount,
		RemainingAmount:    ci.RemainingAmount,
		Status:             string(ci.Status),
		AttachmentURL:      ci.AttachmentURL,
		CreatedAt:          ci.CreatedAt,
	}
}

func (uc *customerInvoiceDownPaymentUsecase) List(ctx context.Context, params *dto.ListCustomerInvoicesRequest) ([]*dto.CustomerInvoiceDownPaymentListResponse, int64, error) {
	params.Type = string(models.CustomerInvoiceTypeDownPayment)
	items, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*dto.CustomerInvoiceDownPaymentListResponse, 0, len(items))
	for _, it := range items {
		i := it
		res = append(res, uc.mapToList(ctx, &i))
	}
	return res, total, nil
}

func (uc *customerInvoiceDownPaymentUsecase) GetByID(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.CustomerInvoice{}, id, security.SalesScopeQueryOptions()) {
		return nil, ErrCustomerInvoiceNotFound
	}
	ci, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	if ci.Type != models.CustomerInvoiceTypeDownPayment {
		return nil, ErrCustomerInvoiceNotFound
	}
	return uc.mapToDetail(ctx, ci), nil
}

func (uc *customerInvoiceDownPaymentUsecase) Create(ctx context.Context, req *dto.CreateCustomerInvoiceDownPaymentRequest) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New(errCustomerInvoiceDPDBNil)
	}
	invDate, dueDate, err := parseCustomerInvoiceDPDates(req.InvoiceDate, req.DueDate)
	if err != nil {
		return nil, err
	}

	var out *models.CustomerInvoice
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		created, createErr := uc.createDraftDownPaymentInTx(ctx, tx, req, invDate, dueDate)
		if createErr != nil {
			return createErr
		}
		out = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Reload after commit so relations (SalesOrder, Customer) are visible on the main connection.
	loaded, err := uc.repo.FindByID(ctx, out.ID)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.create", loaded.ID, map[string]interface{}{"after": loaded})
	return uc.mapToDetail(ctx, loaded), nil
}

func (uc *customerInvoiceDownPaymentUsecase) Update(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceDownPaymentRequest) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	if existing.Type != models.CustomerInvoiceTypeDownPayment {
		return nil, ErrCustomerInvoiceNotFound
	}
	if existing.Status != models.CustomerInvoiceStatusDraft {
		return nil, ErrCustomerInvoiceConflict
	}

	invDate, dueDate, err := parseCustomerInvoiceDPDates(req.InvoiceDate, req.DueDate)
	if err != nil {
		return nil, err
	}

	var out *models.CustomerInvoice
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		updated, updateErr := uc.updateDraftDownPaymentInTx(ctx, tx, id, req, invDate, dueDate)
		if updateErr != nil {
			return updateErr
		}
		out = updated
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.update", id, map[string]interface{}{"after": out})
	return uc.mapToDetail(ctx, out), nil
}

func (uc *customerInvoiceDownPaymentUsecase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrCustomerInvoiceNotFound
		}
		return err
	}
	if existing.Type != models.CustomerInvoiceTypeDownPayment {
		return ErrCustomerInvoiceNotFound
	}
	if existing.Status != models.CustomerInvoiceStatusDraft {
		return ErrCustomerInvoiceConflict
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.delete", id, map[string]interface{}{"before": existing})
	return nil
}

func (uc *customerInvoiceDownPaymentUsecase) Pending(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New(errCustomerInvoiceDPDBNil)
	}
	var out *models.CustomerInvoice
	err := database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		var ci models.CustomerInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ci, customerInvoiceDPByIDQuery, id).Error; err != nil {
			return err
		}
		if ci.Type != models.CustomerInvoiceTypeDownPayment {
			return ErrCustomerInvoiceNotFound
		}
		if ci.Status != models.CustomerInvoiceStatusDraft {
			return ErrCustomerInvoiceConflict
		}
		if err := tx.Model(&ci).Update("status", models.CustomerInvoiceStatusSubmitted).Error; err != nil {
			return err
		}

		loaded, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		out = loaded
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.pending", id, map[string]interface{}{"after": out})
	actorUserID, _ := ctx.Value("user_id").(string)
	if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
		PermissionCode: "customer_invoice_dp.approve",
		EntityType:     "customer_invoice_dp",
		EntityID:       out.ID,
		Title:          "Customer Invoice Down Payment Approval",
		Message:        "A customer invoice down payment has been submitted and requires your approval.",
		ActorUserID:    actorUserID,
	}); err != nil {
		log.Printf("warning: failed to create customer invoice DP notification: %v", err)
	}
	return uc.mapToDetail(ctx, out), nil
}

func (uc *customerInvoiceDownPaymentUsecase) Approve(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New(errCustomerInvoiceDPDBNil)
	}
	err := database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		return uc.approveDownPaymentInTx(ctx, tx, id)
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	out, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.approve", id, nil)
	return uc.mapToDetail(ctx, out), nil
}

func (uc *customerInvoiceDownPaymentUsecase) Cancel(ctx context.Context, id string) (*dto.CustomerInvoiceDownPaymentDetailResponse, error) {
	if uc.db == nil {
		return nil, errors.New(errCustomerInvoiceDPDBNil)
	}
	var out *models.CustomerInvoice
	err := database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		var ci models.CustomerInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ci, customerInvoiceDPByIDQuery, id).Error; err != nil {
			return err
		}
		if ci.Type != models.CustomerInvoiceTypeDownPayment {
			return ErrCustomerInvoiceNotFound
		}
		if ci.Status != models.CustomerInvoiceStatusDraft && ci.Status != models.CustomerInvoiceStatusSubmitted {
			return ErrCustomerInvoiceConflict
		}
		if err := tx.Model(&ci).Update("status", models.CustomerInvoiceStatusCancelled).Error; err != nil {
			return err
		}

		loaded, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		out = loaded
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCustomerInvoiceNotFound
		}
		return nil, err
	}
	uc.auditService.Log(ctx, "customer_invoice_dp.cancel", id, map[string]interface{}{"after": out})
	return uc.mapToDetail(ctx, out), nil
}

func shouldTriggerSalesInvoiceDPJournal(previousStatus, currentStatus models.CustomerInvoiceStatus, invoiceType models.CustomerInvoiceType) bool {
	if invoiceType != models.CustomerInvoiceTypeDownPayment {
		return false
	}

	if currentStatus != models.CustomerInvoiceStatusUnpaid {
		return false
	}

	return previousStatus != models.CustomerInvoiceStatusUnpaid
}

func (uc *customerInvoiceDownPaymentUsecase) TriggerJournalForInvoiceDP(ctx context.Context, invoice *models.CustomerInvoice) error {
	return uc.triggerSalesInvoiceDPJournal(ctx, invoice)
}

func (uc *customerInvoiceDownPaymentUsecase) triggerSalesInvoiceDPJournal(ctx context.Context, invoice *models.CustomerInvoice) error {
	if invoice == nil || uc.journalUC == nil || uc.engine == nil {
		return nil
	}

	if invoice.Amount <= 0 {
		return nil
	}

	data := accounting.TransactionData{
		ReferenceType:   "SALES_INVOICE_DP",
		ReferenceID:     invoice.ID,
		EntryDate:       invoice.InvoiceDate.Format(customerInvoiceDPDateFormat),
		Description:     fmt.Sprintf("Customer DP Invoice %s (%s)", safeInvoiceNumber(invoice.InvoiceNumber), invoice.Code),
		TotalAmount:     invoice.Amount,
		DescriptionArgs: []interface{}{safeInvoiceNumber(invoice.InvoiceNumber), invoice.Code},
	}

	req, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSalesInvoiceDP, data)
	if err != nil {
		return fmt.Errorf("failed to generate customer invoice DP journal: %w", err)
	}

	// Balance check
	var debitTotal, creditTotal float64
	for _, l := range req.Lines {
		debitTotal += l.Debit
		creditTotal += l.Credit
	}
	if math.Abs(debitTotal-creditTotal) > 0.001 {
		return fmt.Errorf("generated customer invoice DP journal is unbalanced: debit=%.2f credit=%.2f", debitTotal, creditTotal)
	}

	_, err = uc.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post customer invoice DP journal: %w", err)
	}

	log.Printf("journal_observability event=trigger.success module=sales_customer_invoice_dp reference_id=%s", invoice.ID)
	return nil
}

func safeInvoiceNumber(invoiceNumber *string) string {
	if invoiceNumber == nil || strings.TrimSpace(*invoiceNumber) == "" {
		return "-"
	}

	return strings.TrimSpace(*invoiceNumber)
}

func (uc *customerInvoiceDownPaymentUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if uc.db == nil {
		return nil, 0, errors.New(errCustomerInvoiceDPDBNil)
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.CustomerInvoice{}, id, security.SalesScopeQueryOptions()) {
		return nil, 0, ErrCustomerInvoiceNotFound
	}

	return listAuditTrailEntries(ctx, uc.db, id, "customer_invoice_dp.", page, perPage)
}

func parseCustomerInvoiceDPDates(invoiceDateRaw, dueDateRaw string) (time.Time, time.Time, error) {
	invoiceDate, err := time.Parse(customerInvoiceDPDateFormat, invoiceDateRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid invoice date")
	}

	dueDate, err := time.Parse(customerInvoiceDPDateFormat, dueDateRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid due date")
	}

	return invoiceDate, dueDate, nil
}

func (uc *customerInvoiceDownPaymentUsecase) createDraftDownPaymentInTx(
	ctx context.Context,
	tx *gorm.DB,
	req *dto.CreateCustomerInvoiceDownPaymentRequest,
	invDate time.Time,
	dueDate time.Time,
) (*models.CustomerInvoice, error) {
	var so models.SalesOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&so, customerInvoiceDPByIDQuery, req.SalesOrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSalesOrderNotFound
		}
		return nil, err
	}
	if so.Status != models.SalesOrderStatusApproved {
		return nil, ErrCustomerInvoiceInvalid
	}

	// Validate amount does not exceed SO total (#241)
	if req.Amount > so.TotalAmount {
		return nil, errors.New("down payment amount cannot exceed sales order total")
	}

	code, err := uc.repo.GetNextInvoiceNumber(ctx, "CIDP")
	if err != nil {
		return nil, err
	}

	invNo := fmt.Sprintf("CUS-DP-%s-%s", apptime.Now().Format("20060102"), strings.TrimPrefix(code, "CIDP-"))
	creatorID, _ := ctx.Value("user_id").(string)

	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	ci := &models.CustomerInvoice{
		Type:          models.CustomerInvoiceTypeDownPayment,
		SalesOrderID:  &so.ID,
		Code:          code,
		InvoiceNumber: &invNo,
		InvoiceDate:   invDate,
		DueDate:       &dueDate,
		Amount:        req.Amount,
		Subtotal:      req.Amount,
		Status:        models.CustomerInvoiceStatusDraft,
		AttachmentURL: req.AttachmentURL,
		Notes:         notes,
		CreatedBy:     &creatorID,
	}

	if err := tx.Create(ci).Error; err != nil {
		return nil, err
	}

	return ci, nil
}

func (uc *customerInvoiceDownPaymentUsecase) updateDraftDownPaymentInTx(
	ctx context.Context,
	tx *gorm.DB,
	id string,
	req *dto.UpdateCustomerInvoiceDownPaymentRequest,
	invDate time.Time,
	dueDate time.Time,
) (*models.CustomerInvoice, error) {
	var ci models.CustomerInvoice
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ci, customerInvoiceDPByIDQuery, id).Error; err != nil {
		return nil, err
	}
	if ci.Status != models.CustomerInvoiceStatusDraft || ci.Type != models.CustomerInvoiceTypeDownPayment {
		return nil, ErrCustomerInvoiceConflict
	}

	var so models.SalesOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&so, customerInvoiceDPByIDQuery, req.SalesOrderID).Error; err != nil {
		return nil, err
	}
	if so.Status != models.SalesOrderStatusApproved {
		return nil, ErrCustomerInvoiceInvalid
	}

	if req.Amount > so.TotalAmount {
		return nil, fmt.Errorf("down payment amount cannot exceed sales order total")
	}

	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	if err := tx.Model(&ci).Updates(map[string]interface{}{
		"sales_order_id": so.ID,
		"invoice_date":   invDate,
		"due_date":       dueDate,
		"amount":         req.Amount,
		"subtotal":       req.Amount,
		"attachment_url": req.AttachmentURL,
		"notes":          notes,
		"updated_at":     apptime.Now(),
	}).Error; err != nil {
		return nil, err
	}

	loaded, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return loaded, nil
}

func (uc *customerInvoiceDownPaymentUsecase) approveDownPaymentInTx(ctx context.Context, tx *gorm.DB, id string) error {
	var ci models.CustomerInvoice
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ci, customerInvoiceDPByIDQuery, id).Error; err != nil {
		return err
	}
	if ci.Type != models.CustomerInvoiceTypeDownPayment {
		return ErrCustomerInvoiceNotFound
	}
	if ci.Status != models.CustomerInvoiceStatusSubmitted {
		return ErrCustomerInvoiceConflict
	}

	previousStatus := ci.Status
	newStatus := models.CustomerInvoiceStatusUnpaid

	now := apptime.Now()
	if err := tx.Model(&ci).Updates(map[string]interface{}{
		"status":           newStatus,
		"approved_at":      &now,
		"remaining_amount": ci.Amount,
	}).Error; err != nil {
		return err
	}

	ci.Status = newStatus
	ci.RemainingAmount = ci.Amount

	if shouldTriggerSalesInvoiceDPJournal(previousStatus, newStatus, ci.Type) {
		triggerCtx := withActorContext(database.WithTx(ctx, tx), ci.CreatedBy, ci.CreatedBy)
		if err := uc.triggerSalesInvoiceDPJournal(triggerCtx, &ci); err != nil {
			return fmt.Errorf("failed to trigger customer invoice down payment journal: %w", err)
		}
	}

	return nil
}
