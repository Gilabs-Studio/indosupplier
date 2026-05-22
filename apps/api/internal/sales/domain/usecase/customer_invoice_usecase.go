package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	finDto "github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUsecase "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	loyaltyDto "github.com/gilabs/gims/api/internal/loyalty/domain/dto"
	loyaltyUsecase "github.com/gilabs/gims/api/internal/loyalty/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/mapper"
	salesService "github.com/gilabs/gims/api/internal/sales/domain/service"
	"gorm.io/gorm"
)

// Date format constant
const dateFormat = "2006-01-02"

// Errors
var (
	ErrCustomerInvoiceNotFound   = errors.New("customer invoice not found")
	ErrInvalidInvoiceStatus      = errors.New("invalid invoice status for this operation")
	ErrInvoiceExceedsRemaining   = errors.New("invoice quantity exceeds remaining invoiceable quantity")
	ErrInvoiceDOMismatch         = errors.New("delivery order does not belong to the same sales order")
	ErrInvalidDownPaymentInvoice = errors.New("invalid down payment invoice")
)

// CustomerInvoiceUsecase defines the interface for customer invoice business logic
type CustomerInvoiceUsecase interface {
	List(ctx context.Context, req *dto.ListCustomerInvoicesRequest) ([]dto.CustomerInvoiceResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.CustomerInvoiceResponse, error)
	ListItems(ctx context.Context, invoiceID string, req *dto.ListCustomerInvoiceItemsRequest) ([]dto.CustomerInvoiceItemResponse, *utils.PaginationResult, error)
	Create(ctx context.Context, req *dto.CreateCustomerInvoiceRequest, createdBy *string) (*dto.CustomerInvoiceResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceRequest) (*dto.CustomerInvoiceResponse, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceStatusRequest, userID *string) (*dto.CustomerInvoiceResponse, error)
	Reverse(ctx context.Context, id string) (*dto.CustomerInvoiceResponse, error)
	ReverseWithReason(ctx context.Context, id string, reason string) (*dto.CustomerInvoiceResponse, error)
	TriggerJournalForInvoice(ctx context.Context, invoice *models.CustomerInvoice) error
	PreviewJournal(ctx context.Context, req *dto.CreateCustomerInvoiceRequest) (*dto.CustomerInvoiceJournalPreviewResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
	SetLoyaltyUsecase(luc loyaltyUsecase.LoyaltyUsecase)
}

type customerInvoiceUsecase struct {
	db              *gorm.DB
	invoiceRepo     repositories.CustomerInvoiceRepository
	productRepo     productRepos.ProductRepository
	salesOrderRepo  repositories.SalesOrderRepository
	journalUC       finUsecase.JournalEntryUsecase
	coaUC           finUsecase.ChartOfAccountUsecase
	mappingUC       finUsecase.SystemAccountMappingUsecase // Deprecated: use engine instead
	engine          accounting.AccountingEngine
	auditService    audit.AuditService
	loyaltyUC       loyaltyUsecase.LoyaltyUsecase // Optional: awards loyalty points when invoice is PAID
	salesJournalSvc salesService.SalesJournalService
}

// SetLoyaltyUsecase injects the loyalty usecase for awarding points on invoice payment.
// It is called after construction to avoid circular import issues during wiring.
func (uc *customerInvoiceUsecase) SetLoyaltyUsecase(luc loyaltyUsecase.LoyaltyUsecase) {
	uc.loyaltyUC = luc
}

// NewCustomerInvoiceUsecase creates a new CustomerInvoiceUsecase
func NewCustomerInvoiceUsecase(
	db *gorm.DB,
	invoiceRepo repositories.CustomerInvoiceRepository,
	productRepo productRepos.ProductRepository,
	salesOrderRepo repositories.SalesOrderRepository,
	journalUC finUsecase.JournalEntryUsecase,
	coaUC finUsecase.ChartOfAccountUsecase,
	auditService audit.AuditService,
	engine accounting.AccountingEngine,
	salesJournalSvc salesService.SalesJournalService,
) CustomerInvoiceUsecase {
	if salesJournalSvc == nil {
		salesJournalSvc = salesService.NewSalesJournalService(db, journalUC, engine)
	}

	return &customerInvoiceUsecase{
		db:              db,
		invoiceRepo:     invoiceRepo,
		productRepo:     productRepo,
		salesOrderRepo:  salesOrderRepo,
		journalUC:       journalUC,
		coaUC:           coaUC,
		auditService:    auditService,
		engine:          engine,
		salesJournalSvc: salesJournalSvc,
	}
}

func (uc *customerInvoiceUsecase) List(ctx context.Context, req *dto.ListCustomerInvoicesRequest) ([]dto.CustomerInvoiceResponse, *utils.PaginationResult, error) {
	req.Status = normalizeCustomerInvoiceListStatus(req.Status)

	invoices, total, err := uc.invoiceRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return mapper.MapCustomerInvoicesToResponse(invoices), pagination, nil
}

func normalizeCustomerInvoiceListStatus(status string) string {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	if normalized == "" {
		return ""
	}

	// Backward-compatible alias: older clients use "sent" for pending approval.
	if normalized == "SENT" {
		return "SUBMITTED"
	}

	if normalized == "CANCELED" {
		return "CANCELLED"
	}

	return normalized
}

func (uc *customerInvoiceUsecase) GetByID(ctx context.Context, id string) (*dto.CustomerInvoiceResponse, error) {
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.CustomerInvoice{}, id, security.DefaultScopeQueryOptions()) {
		return nil, ErrCustomerInvoiceNotFound
	}
	invoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrCustomerInvoiceNotFound
	}

	return mapper.MapCustomerInvoiceToResponse(invoice), nil
}

func (uc *customerInvoiceUsecase) ListItems(ctx context.Context, invoiceID string, req *dto.ListCustomerInvoiceItemsRequest) ([]dto.CustomerInvoiceItemResponse, *utils.PaginationResult, error) {
	// Verify invoice exists
	_, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, nil, ErrCustomerInvoiceNotFound
	}

	items, total, err := uc.invoiceRepo.ListItems(ctx, invoiceID, req)
	if err != nil {
		return nil, nil, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return mapper.MapCustomerInvoiceItemsToResponse(items), pagination, nil
}

func (uc *customerInvoiceUsecase) Create(ctx context.Context, req *dto.CreateCustomerInvoiceRequest, createdBy *string) (*dto.CustomerInvoiceResponse, error) {
	invoiceDate, err := time.Parse(dateFormat, req.InvoiceDate)
	if err != nil {
		return nil, err
	}

	invoiceType := parseCustomerInvoiceType(req.Type)

	soItemMap, err := uc.buildSalesOrderItemMapForCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	invoice := &models.CustomerInvoice{
		Type:                 invoiceType,
		InvoiceDate:          invoiceDate,
		SalesOrderID:         req.SalesOrderID,
		DeliveryOrderID:      req.DeliveryOrderID,
		PaymentTermsID:       req.PaymentTermsID,
		DownPaymentInvoiceID: req.DownPaymentInvoiceID,
		TaxRate:              req.TaxRate,
		DeliveryCost:         req.DeliveryCost,
		OtherCost:            req.OtherCost,
		Notes:                req.Notes,
		Status:               models.CustomerInvoiceStatusDraft,
		CreatedBy:            createdBy,
	}
	if err := applyInvoiceDueDate(invoice, req.DueDate); err != nil {
		return nil, err
	}

	items, subtotal, err := uc.buildCreateInvoiceItems(ctx, req, soItemMap)
	if err != nil {
		return nil, err
	}

	invoice.Items = items
	invoice.Subtotal = subtotal
	invoice.TaxAmount = subtotal * (invoice.TaxRate / 100)
	invoice.Amount = subtotal + invoice.TaxAmount + invoice.DeliveryCost + invoice.OtherCost

	if err := uc.applyPaidDownPaymentsOnCreate(ctx, invoice, req.SalesOrderID, invoiceType, req.DownPaymentInvoiceID); err != nil {
		return nil, err
	}
	invoice.RemainingAmount = invoice.Amount

	err = uc.createInvoiceWithSOItemUpdates(ctx, invoice)
	if err != nil {
		return nil, err
	}

	// Fetch the created invoice with relations
	createdInvoice, err := uc.invoiceRepo.FindByID(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}

	logSalesAudit(uc.auditService, ctx, "customer_invoice.create", createdInvoice.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":           createdInvoice.Code,
			"status":         createdInvoice.Status,
			"type":           createdInvoice.Type,
			"invoice_date":   createdInvoice.InvoiceDate,
			"amount":         createdInvoice.Amount,
			"remaining":      createdInvoice.RemainingAmount,
			"sales_order_id": createdInvoice.SalesOrderID,
		},
	})

	return mapper.MapCustomerInvoiceToResponse(createdInvoice), nil
}

func parseCustomerInvoiceType(raw string) models.CustomerInvoiceType {
	if strings.TrimSpace(raw) == "" {
		return models.CustomerInvoiceTypeRegular
	}

	return models.CustomerInvoiceType(raw)
}

func applyInvoiceDueDate(invoice *models.CustomerInvoice, dueDateRaw *string) error {
	if invoice == nil || dueDateRaw == nil || strings.TrimSpace(*dueDateRaw) == "" {
		return fmt.Errorf("due_date is required")
	}

	dueDate, err := time.Parse(dateFormat, *dueDateRaw)
	if err != nil {
		return fmt.Errorf("invalid due_date format, expected YYYY-MM-DD")
	}
	invoice.DueDate = &dueDate
	return nil
}

func (uc *customerInvoiceUsecase) buildSalesOrderItemMapForCreate(
	ctx context.Context,
	req *dto.CreateCustomerInvoiceRequest,
) (map[string]*models.SalesOrderItem, error) {
	soItemMap := make(map[string]*models.SalesOrderItem)
	if req.SalesOrderID == nil {
		return soItemMap, nil
	}

	salesOrder, err := uc.salesOrderRepo.FindByID(ctx, *req.SalesOrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesOrderNotFound
		}
		return nil, err
	}

	if req.DeliveryOrderID != nil {
		deliveryOrderFound := false
		for _, do := range salesOrder.DeliveryOrders {
			if do.ID == *req.DeliveryOrderID {
				deliveryOrderFound = true
				break
			}
		}
		if !deliveryOrderFound {
			return nil, ErrInvoiceDOMismatch
		}
	}

	for i := range salesOrder.Items {
		soItemMap[salesOrder.Items[i].ID] = &salesOrder.Items[i]
	}

	return soItemMap, nil
}

func (uc *customerInvoiceUsecase) buildCreateInvoiceItems(
	ctx context.Context,
	req *dto.CreateCustomerInvoiceRequest,
	soItemMap map[string]*models.SalesOrderItem,
) ([]models.CustomerInvoiceItem, float64, error) {
	items := make([]models.CustomerInvoiceItem, len(req.Items))
	subtotal := 0.0

	for i, itemReq := range req.Items {
		resolvedProductID, product, err := uc.resolveInvoiceItemProduct(ctx, itemReq, soItemMap, nil)
		if err != nil {
			return nil, 0, err
		}

		if err := validateCreateInvoiceItemQuantity(itemReq.SalesOrderItemID, itemReq.Quantity, soItemMap, product.Name); err != nil {
			return nil, 0, err
		}

		itemSubtotal := (itemReq.Price * itemReq.Quantity) - itemReq.Discount
		subtotal += itemSubtotal

		items[i] = models.CustomerInvoiceItem{
			ProductID:           resolvedProductID,
			SalesOrderItemID:    itemReq.SalesOrderItemID,
			DeliveryOrderItemID: itemReq.DeliveryOrderItemID,
			Quantity:            itemReq.Quantity,
			Price:               itemReq.Price,
			Discount:            itemReq.Discount,
			Subtotal:            itemSubtotal,
			HPPAmount:           itemReq.HPPAmount,
		}

		if items[i].HPPAmount == 0 && product.CurrentHpp > 0 {
			items[i].HPPAmount = product.CurrentHpp
		}
	}

	return items, subtotal, nil
}

func (uc *customerInvoiceUsecase) resolveInvoiceItemProduct(
	ctx context.Context,
	itemReq dto.CreateCustomerInvoiceItemRequest,
	soItemMap map[string]*models.SalesOrderItem,
	existingItems []models.CustomerInvoiceItem,
) (string, *productModels.Product, error) {
	resolvedProductID := strings.TrimSpace(itemReq.ProductID)

	if itemReq.SalesOrderItemID != nil {
		if soItemMap != nil {
			if soItem, ok := soItemMap[*itemReq.SalesOrderItemID]; ok && soItem != nil && strings.TrimSpace(soItem.ProductID) != "" {
				resolvedProductID = soItem.ProductID
			}
		}

		if resolvedProductID == "" && len(existingItems) > 0 {
			for i := range existingItems {
				if existingItems[i].SalesOrderItemID != nil && *existingItems[i].SalesOrderItemID == *itemReq.SalesOrderItemID && strings.TrimSpace(existingItems[i].ProductID) != "" {
					resolvedProductID = existingItems[i].ProductID
					break
				}
			}
		}
	}

	if resolvedProductID == "" {
		return "", nil, ErrProductNotFound
	}

	product, err := uc.productRepo.FindByID(ctx, resolvedProductID)
	if err != nil {
		return "", nil, ErrProductNotFound
	}

	return resolvedProductID, product, nil
}

func validateCreateInvoiceItemQuantity(
	salesOrderItemID *string,
	requestedQty float64,
	soItemMap map[string]*models.SalesOrderItem,
	productName string,
) error {
	if salesOrderItemID == nil {
		return nil
	}

	soItem, ok := soItemMap[*salesOrderItemID]
	if !ok {
		return fmt.Errorf("sales order item %s not found", *salesOrderItemID)
	}

	remainingQty := soItem.Quantity - soItem.InvoicedQuantity
	if requestedQty <= remainingQty {
		return nil
	}

	return fmt.Errorf("%w: product %s has %.3f remaining, requested %.3f",
		ErrInvoiceExceedsRemaining, productName, remainingQty, requestedQty)
}

func (uc *customerInvoiceUsecase) applyPaidDownPaymentsOnCreate(
	ctx context.Context,
	invoice *models.CustomerInvoice,
	salesOrderID *string,
	invoiceType models.CustomerInvoiceType,
	downPaymentInvoiceID *string,
) error {
	if salesOrderID == nil || invoiceType != models.CustomerInvoiceTypeRegular {
		return nil
	}

	normalizedSalesOrderID := strings.TrimSpace(*salesOrderID)
	if normalizedSalesOrderID == "" {
		return nil
	}

	dpID := ""
	if downPaymentInvoiceID != nil {
		dpID = strings.TrimSpace(*downPaymentInvoiceID)
	}

	if dpID == "" {
		autoDPID, err := uc.findAutoApplicableDownPaymentInvoiceID(ctx, normalizedSalesOrderID)
		if err != nil {
			return err
		}
		dpID = autoDPID
	}

	if dpID == "" {
		return nil
	}

	dp, err := uc.invoiceRepo.FindByID(ctx, dpID)
	if err != nil {
		return ErrInvalidDownPaymentInvoice
	}
	if dp == nil || dp.Type != models.CustomerInvoiceTypeDownPayment || !isPaidOrPartialDownPaymentStatus(dp.Status) {
		return ErrInvalidDownPaymentInvoice
	}
	if dp.SalesOrderID == nil || strings.TrimSpace(*dp.SalesOrderID) == "" || strings.TrimSpace(*dp.SalesOrderID) != normalizedSalesOrderID {
		return ErrInvalidDownPaymentInvoice
	}

	paidDownPayment := resolvePaidDownPaymentAmount(dp)
	if paidDownPayment <= 0 {
		return ErrInvalidDownPaymentInvoice
	}

	invoice.DownPaymentInvoiceID = &dpID
	invoice.DownPaymentAmount = paidDownPayment
	invoice.Amount -= paidDownPayment
	if invoice.Amount < 0 {
		invoice.Amount = 0
	}

	return nil
}

func isPaidOrPartialDownPaymentStatus(status models.CustomerInvoiceStatus) bool {
	return status == models.CustomerInvoiceStatusPaid || status == models.CustomerInvoiceStatusPartial
}

func resolvePaidDownPaymentAmount(dp *models.CustomerInvoice) float64 {
	if dp == nil {
		return 0
	}

	if dp.PaidAmount > 0 {
		return dp.PaidAmount
	}

	if dp.Amount > 0 && dp.Amount >= dp.RemainingAmount {
		return dp.Amount - dp.RemainingAmount
	}

	return 0
}

func (uc *customerInvoiceUsecase) findAutoApplicableDownPaymentInvoiceID(ctx context.Context, salesOrderID string) (string, error) {
	normalizedSalesOrderID := strings.TrimSpace(salesOrderID)
	if uc.db == nil || normalizedSalesOrderID == "" {
		return "", nil
	}

	var candidates []models.CustomerInvoice
	if err := database.GetDB(ctx, uc.db).
		Where("type = ?", models.CustomerInvoiceTypeDownPayment).
		Where("sales_order_id = ?", normalizedSalesOrderID).
		Where("status IN ?", []models.CustomerInvoiceStatus{models.CustomerInvoiceStatusPaid, models.CustomerInvoiceStatusPartial}).
		Where("paid_amount > 0").
		Order("invoice_date DESC").
		Order("created_at DESC").
		Find(&candidates).Error; err != nil {
		return "", err
	}

	for _, candidate := range candidates {
		candidateID := strings.TrimSpace(candidate.ID)
		if candidateID == "" {
			continue
		}

		var linkedRegularCount int64
		if err := database.GetDB(ctx, uc.db).
			Model(&models.CustomerInvoice{}).
			Where("type = ?", models.CustomerInvoiceTypeRegular).
			Where("down_payment_invoice_id = ?", candidateID).
			Where("status NOT IN ?", []models.CustomerInvoiceStatus{models.CustomerInvoiceStatusCancelled, models.CustomerInvoiceStatusReversed}).
			Count(&linkedRegularCount).Error; err != nil {
			return "", err
		}

		if linkedRegularCount == 0 {
			return candidateID, nil
		}
	}

	return "", nil
}

func (uc *customerInvoiceUsecase) applyLinkedDownPaymentOnUpdate(ctx context.Context, invoice *models.CustomerInvoice) error {
	if invoice == nil {
		return nil
	}

	if invoice.Type != models.CustomerInvoiceTypeRegular || invoice.SalesOrderID == nil || invoice.DownPaymentInvoiceID == nil {
		invoice.DownPaymentAmount = 0
		return nil
	}

	dpID := strings.TrimSpace(*invoice.DownPaymentInvoiceID)
	if dpID == "" {
		invoice.DownPaymentAmount = 0
		invoice.DownPaymentInvoiceID = nil
		return nil
	}

	dp, err := uc.invoiceRepo.FindByID(ctx, dpID)
	if err != nil {
		return ErrInvalidDownPaymentInvoice
	}
	if dp == nil || dp.Type != models.CustomerInvoiceTypeDownPayment || !isPaidOrPartialDownPaymentStatus(dp.Status) {
		return ErrInvalidDownPaymentInvoice
	}
	if dp.SalesOrderID == nil || invoice.SalesOrderID == nil || strings.TrimSpace(*dp.SalesOrderID) != strings.TrimSpace(*invoice.SalesOrderID) {
		return ErrInvalidDownPaymentInvoice
	}

	paidDownPayment := resolvePaidDownPaymentAmount(dp)
	if paidDownPayment <= 0 {
		return ErrInvalidDownPaymentInvoice
	}

	invoice.DownPaymentAmount = paidDownPayment
	invoice.Amount -= paidDownPayment
	if invoice.Amount < 0 {
		invoice.Amount = 0
	}

	return nil
}

func (uc *customerInvoiceUsecase) createInvoiceWithSOItemUpdates(ctx context.Context, invoice *models.CustomerInvoice) error {
	return database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := uc.invoiceRepo.Create(txCtx, invoice); err != nil {
			return err
		}

		for _, item := range invoice.Items {
			if item.SalesOrderItemID == nil {
				continue
			}

			if err := uc.salesOrderRepo.UpdateItemInvoicedQty(txCtx, *item.SalesOrderItemID, item.Quantity); err != nil {
				return fmt.Errorf("failed to update invoiced qty for SO item %s: %w", *item.SalesOrderItemID, err)
			}
		}

		return nil
	})
}

func (uc *customerInvoiceUsecase) Update(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceRequest) (*dto.CustomerInvoiceResponse, error) {
	invoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrCustomerInvoiceNotFound
	}

	// Only allow updates on draft invoices
	if invoice.Status != models.CustomerInvoiceStatusDraft {
		return nil, ErrInvalidInvoiceStatus
	}

	beforeSnapshot := customerInvoiceAuditSnapshot(invoice)

	// Update fields
	if req.InvoiceDate != nil {
		invoiceDate, err := time.Parse(dateFormat, *req.InvoiceDate)
		if err == nil {
			invoice.InvoiceDate = invoiceDate
		}
	}

	if req.DueDate != nil {
		if *req.DueDate != "" {
			dueDate, err := time.Parse(dateFormat, *req.DueDate)
			if err == nil {
				invoice.DueDate = &dueDate
			}
		} else {
			invoice.DueDate = nil
		}
	}

	if req.Type != nil {
		invoice.Type = models.CustomerInvoiceType(*req.Type)
	}

	if req.PaymentTermsID != nil {
		invoice.PaymentTermsID = req.PaymentTermsID
	}

	if req.TaxRate != nil {
		invoice.TaxRate = *req.TaxRate
	}

	if req.DeliveryCost != nil {
		invoice.DeliveryCost = *req.DeliveryCost
	}

	if req.OtherCost != nil {
		invoice.OtherCost = *req.OtherCost
	}

	if req.Notes != nil {
		invoice.Notes = *req.Notes
	}

	// Update items if provided
	if req.Items != nil {
		// Prevent editing items if invoice is sourced from a Sales Order
		if invoice.SalesOrderID != nil && strings.TrimSpace(*invoice.SalesOrderID) != "" {
			return nil, fmt.Errorf("cannot modify items on a Sales Order-linked invoice")
		}
		// Rollback old invoiced quantities on SO items before replacing
		for _, item := range invoice.Items {
			if item.SalesOrderItemID != nil {
				if err := uc.salesOrderRepo.UpdateItemInvoicedQty(ctx, *item.SalesOrderItemID, -item.Quantity); err != nil {
					return nil, fmt.Errorf("failed to rollback invoiced qty for SO item %s: %w", *item.SalesOrderItemID, err)
				}
			}
		}

		var subtotal float64
		items := make([]models.CustomerInvoiceItem, len(*req.Items))
		for i, itemReq := range *req.Items {
			resolvedProductID, product, err := uc.resolveInvoiceItemProduct(ctx, itemReq, nil, invoice.Items)
			if err != nil {
				return nil, err
			}

			itemSubtotal := (itemReq.Price * itemReq.Quantity) - itemReq.Discount
			subtotal += itemSubtotal

			items[i] = models.CustomerInvoiceItem{
				ProductID:        resolvedProductID,
				SalesOrderItemID: itemReq.SalesOrderItemID,
				Quantity:         itemReq.Quantity,
				Price:            itemReq.Price,
				Discount:         itemReq.Discount,
				Subtotal:         itemSubtotal,
				HPPAmount:        itemReq.HPPAmount,
			}

			if items[i].HPPAmount == 0 && product.CurrentHpp > 0 {
				items[i].HPPAmount = product.CurrentHpp
			}
		}

		invoice.Items = items
		invoice.Subtotal = subtotal
	}

	// Recalculate totals
	invoice.TaxAmount = invoice.Subtotal * (invoice.TaxRate / 100)
	invoice.Amount = invoice.Subtotal + invoice.TaxAmount + invoice.DeliveryCost + invoice.OtherCost

	if err := uc.applyLinkedDownPaymentOnUpdate(ctx, invoice); err != nil {
		return nil, err
	}

	invoice.RemainingAmount = invoice.Amount - invoice.PaidAmount

	if err := uc.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, err
	}

	updatedInvoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply new invoiced quantities on SO items after successful update
	if req.Items != nil {
		for _, item := range updatedInvoice.Items {
			if item.SalesOrderItemID != nil {
				if err := uc.salesOrderRepo.UpdateItemInvoicedQty(ctx, *item.SalesOrderItemID, item.Quantity); err != nil {
					return nil, fmt.Errorf("failed to update invoiced qty for SO item %s: %w", *item.SalesOrderItemID, err)
				}
			}
		}
	}

	logSalesAudit(uc.auditService, ctx, "customer_invoice.update", id, map[string]interface{}{
		"before": beforeSnapshot,
		"after":  customerInvoiceAuditSnapshot(updatedInvoice),
	})

	return mapper.MapCustomerInvoiceToResponse(updatedInvoice), nil
}

func (uc *customerInvoiceUsecase) Delete(ctx context.Context, id string) error {
	invoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return ErrCustomerInvoiceNotFound
	}

	beforeSnapshot := customerInvoiceAuditSnapshot(invoice)

	// Allow deletion of draft invoices only
	if invoice.Status != models.CustomerInvoiceStatusDraft {
		return ErrInvalidInvoiceStatus
	}

	// Rollback InvoicedQuantity on SO items and delete in a transaction
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		for _, item := range invoice.Items {
			if item.SalesOrderItemID != nil {
				// Negative qty to decrement InvoicedQuantity
				if err := uc.salesOrderRepo.UpdateItemInvoicedQty(txCtx, *item.SalesOrderItemID, -item.Quantity); err != nil {
					return fmt.Errorf("failed to rollback invoiced qty for SO item %s: %w", *item.SalesOrderItemID, err)
				}
			}
		}

		return uc.invoiceRepo.Delete(txCtx, id)
	})
	if err != nil {
		return err
	}

	logSalesAudit(uc.auditService, ctx, "customer_invoice.delete", id, map[string]interface{}{
		"before": beforeSnapshot,
	})

	return nil
}

func (uc *customerInvoiceUsecase) UpdateStatus(ctx context.Context, id string, req *dto.UpdateCustomerInvoiceStatusRequest, userID *string) (*dto.CustomerInvoiceResponse, error) {
	invoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrCustomerInvoiceNotFound
	}

	newStatus := models.CustomerInvoiceStatus(strings.ToUpper(strings.TrimSpace(req.Status)))
	previousStatus := invoice.Status

	// Validate status transition
	if !isValidStatusTransition(invoice.Status, newStatus) {
		log.Printf("[UpdateStatus] Invalid transition for invoice %s: %s -> %s", id, invoice.Status, newStatus)
		return nil, ErrInvalidStatusTransition
	}

	var paymentAt *time.Time
	if req.PaymentAt != nil && *req.PaymentAt != "" {
		t, err := time.Parse(time.RFC3339, *req.PaymentAt)
		if err == nil {
			paymentAt = &t
		}
	}

	var updatedInvoice *models.CustomerInvoice
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		if err := uc.invoiceRepo.UpdateStatus(txCtx, id, newStatus, req.PaidAmount, paymentAt, userID); err != nil {
			return err
		}

		loaded, err := uc.invoiceRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}
		updatedInvoice = loaded

		if shouldTriggerSalesInvoiceJournal(previousStatus, newStatus, updatedInvoice.Type) {
			triggerCtx := withActorContext(txCtx, userID, updatedInvoice.CreatedBy)
			if err := uc.triggerSalesInvoiceJournal(triggerCtx, updatedInvoice); err != nil {
				return fmt.Errorf("failed to trigger sales invoice journal: %w", err)
			}

			if updatedInvoice.JournalEntryID == nil || strings.TrimSpace(*updatedInvoice.JournalEntryID) == "" {
				var journal financeModels.JournalEntry
				if err := tx.WithContext(txCtx).
					Where("reference_type = ? AND reference_id = ? AND status = ?", reference.RefTypeSalesInvoice, updatedInvoice.ID, financeModels.JournalStatusPosted).
					Order("created_at DESC").First(&journal).Error; err == nil {
					journalID := journal.ID
					if err := tx.WithContext(txCtx).Model(&models.CustomerInvoice{}).
						Where("id = ?", updatedInvoice.ID).
						Updates(map[string]interface{}{
							"journal_entry_id": journalID,
							"is_posted":        true,
						}).Error; err != nil {
						return err
					}
					updatedInvoice.JournalEntryID = &journalID
					updatedInvoice.IsPosted = true
				}
			}
		}

		if shouldTriggerSalesDPApplicationJournal(previousStatus, newStatus, updatedInvoice) {
			triggerCtx := withActorContext(txCtx, userID, updatedInvoice.CreatedBy)
			if err := uc.triggerSalesDPApplicationJournal(triggerCtx, updatedInvoice); err != nil {
				return fmt.Errorf("failed to trigger sales DP application journal: %w", err)
			}
		}

		if shouldTriggerSalesInvoiceReversal(previousStatus, newStatus, updatedInvoice.Type) {
			triggerCtx := withActorContext(txCtx, userID, updatedInvoice.CreatedBy)
			if err := uc.triggerSalesInvoiceJournalReversal(triggerCtx, updatedInvoice); err != nil {
				return fmt.Errorf("failed to reverse sales invoice journal: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	logSalesAudit(uc.auditService, ctx, "customer_invoice.status_change", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  updatedInvoice.Status,
		"paid_amount":   req.PaidAmount,
		"payment_at":    req.PaymentAt,
	})

	// Award loyalty points when a non-POS invoice transitions to PAID.
	if newStatus == models.CustomerInvoiceStatusPaid {
		go uc.tryEarnLoyaltyPointsForInvoice(context.Background(), updatedInvoice)
	}

	if newStatus == models.CustomerInvoiceStatusSubmitted {
		actorUserID, _ := ctx.Value("user_id").(string)
		if err := notificationService.CreateApprovalNotification(ctx, uc.db, notificationService.ApprovalNotificationParams{
			PermissionCode: "customer_invoice.approve",
			EntityType:     "customer_invoice",
			EntityID:       updatedInvoice.ID,
			Title:          "Customer Invoice Approval",
			Message:        "A customer invoice has been submitted and requires your approval.",
			ActorUserID:    actorUserID,
		}); err != nil {
			log.Printf("warning: failed to create customer invoice notification: %v", err)
		}
	}

	return mapper.MapCustomerInvoiceToResponse(updatedInvoice), nil
}

func (uc *customerInvoiceUsecase) ReverseWithReason(ctx context.Context, id string, reason string) (*dto.CustomerInvoiceResponse, error) {
	return uc.reverse(ctx, id, reason)
}

func (uc *customerInvoiceUsecase) Reverse(ctx context.Context, id string) (*dto.CustomerInvoiceResponse, error) {
	return uc.reverse(ctx, id, "Manual reversal")
}

func (uc *customerInvoiceUsecase) reverse(ctx context.Context, id string, reason string) (*dto.CustomerInvoiceResponse, error) {
	invoice, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrCustomerInvoiceNotFound
	}

	previousStatus := invoice.Status
	newStatus := models.CustomerInvoiceStatusReversed

	if !isValidStatusTransition(previousStatus, newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	var updatedInvoice *models.CustomerInvoice
	err = database.GetDB(ctx, uc.db).Transaction(func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)

		now := time.Now()
		if err := tx.Model(invoice).Updates(map[string]interface{}{
			"status":       newStatus,
			"cancelled_at": &now,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		loaded, err := uc.invoiceRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}
		updatedInvoice = loaded

		// Trigger journal reversal
		if shouldTriggerSalesInvoiceReversal(previousStatus, newStatus, updatedInvoice.Type) {
			if err := uc.triggerSalesInvoiceJournalReversed(txCtx, updatedInvoice, reason); err != nil {
				return fmt.Errorf("failed to reverse sales invoice journal: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logSalesAudit(uc.auditService, ctx, "customer_invoice.reverse", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  newStatus,
		"reason":        reason,
	})

	return mapper.MapCustomerInvoiceToResponse(updatedInvoice), nil
}

func (uc *customerInvoiceUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if uc.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if !security.CheckRecordScopeAccess(uc.db, ctx, &models.CustomerInvoice{}, id, security.DefaultScopeQueryOptions()) {
		return nil, 0, ErrCustomerInvoiceNotFound
	}

	return listAuditTrailEntries(ctx, uc.db, id, "customer_invoice.", page, perPage)
}

func shouldTriggerSalesInvoiceJournal(previousStatus, currentStatus models.CustomerInvoiceStatus, invoiceType models.CustomerInvoiceType) bool {
	if invoiceType != models.CustomerInvoiceTypeRegular {
		return false
	}

	isCurrentPostApproved := currentStatus == models.CustomerInvoiceStatusApproved ||
		currentStatus == models.CustomerInvoiceStatusUnpaid ||
		currentStatus == models.CustomerInvoiceStatusWaitingPayment ||
		currentStatus == models.CustomerInvoiceStatusPartial ||
		currentStatus == models.CustomerInvoiceStatusPaid

	isPreviousPostApproved := previousStatus == models.CustomerInvoiceStatusApproved ||
		previousStatus == models.CustomerInvoiceStatusUnpaid ||
		previousStatus == models.CustomerInvoiceStatusWaitingPayment ||
		previousStatus == models.CustomerInvoiceStatusPartial ||
		previousStatus == models.CustomerInvoiceStatusPaid

	return isCurrentPostApproved && !isPreviousPostApproved
}

func shouldTriggerSalesDPApplicationJournal(previousStatus, currentStatus models.CustomerInvoiceStatus, invoice *models.CustomerInvoice) bool {
	if invoice == nil || invoice.Type != models.CustomerInvoiceTypeRegular {
		return false
	}

	if invoice.DownPaymentAmount <= 0 {
		return false
	}

	if !hasLinkedDeliverySalesJournal(invoice) {
		return false
	}

	isCurrentPostApproved := currentStatus == models.CustomerInvoiceStatusApproved ||
		currentStatus == models.CustomerInvoiceStatusUnpaid ||
		currentStatus == models.CustomerInvoiceStatusWaitingPayment ||
		currentStatus == models.CustomerInvoiceStatusPartial ||
		currentStatus == models.CustomerInvoiceStatusPaid

	isPreviousPostApproved := previousStatus == models.CustomerInvoiceStatusApproved ||
		previousStatus == models.CustomerInvoiceStatusUnpaid ||
		previousStatus == models.CustomerInvoiceStatusWaitingPayment ||
		previousStatus == models.CustomerInvoiceStatusPartial ||
		previousStatus == models.CustomerInvoiceStatusPaid

	return isCurrentPostApproved && !isPreviousPostApproved
}

func hasLinkedDeliverySalesJournal(invoice *models.CustomerInvoice) bool {
	if invoice == nil || invoice.DeliveryOrderID == nil {
		return false
	}

	if invoice.DeliveryOrder == nil {
		return false
	}

	if invoice.DeliveryOrder.JournalEntryID == nil {
		return false
	}

	return strings.TrimSpace(*invoice.DeliveryOrder.JournalEntryID) != ""
}

func customerInvoiceAuditSnapshot(invoice *models.CustomerInvoice) map[string]interface{} {
	if invoice == nil {
		return nil
	}

	return map[string]interface{}{
		"code":                    invoice.Code,
		"invoice_number":          invoice.InvoiceNumber,
		"status":                  invoice.Status,
		"type":                    invoice.Type,
		"invoice_date":            invoice.InvoiceDate,
		"due_date":                invoice.DueDate,
		"sales_order_id":          invoice.SalesOrderID,
		"delivery_order_id":       invoice.DeliveryOrderID,
		"payment_terms_id":        invoice.PaymentTermsID,
		"tax_rate":                invoice.TaxRate,
		"tax_amount":              invoice.TaxAmount,
		"delivery_cost":           invoice.DeliveryCost,
		"other_cost":              invoice.OtherCost,
		"subtotal":                invoice.Subtotal,
		"down_payment_amount":     invoice.DownPaymentAmount,
		"down_payment_invoice_id": invoice.DownPaymentInvoiceID,
		"amount":                  invoice.Amount,
		"paid_amount":             invoice.PaidAmount,
		"remaining_amount":        invoice.RemainingAmount,
		"payment_at":              invoice.PaymentAt,
		"notes":                   invoice.Notes,
		"items":                   customerInvoiceAuditItems(invoice.Items),
	}
}

func customerInvoiceAuditItems(items []models.CustomerInvoiceItem) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"id":                     item.ID,
			"product_id":             item.ProductID,
			"sales_order_item_id":    item.SalesOrderItemID,
			"delivery_order_item_id": item.DeliveryOrderItemID,
			"quantity":               item.Quantity,
			"price":                  item.Price,
			"discount":               item.Discount,
			"subtotal":               item.Subtotal,
			"hpp_amount":             item.HPPAmount,
		})
	}

	return out
}

func shouldTriggerSalesInvoiceReversal(previousStatus, currentStatus models.CustomerInvoiceStatus, invoiceType models.CustomerInvoiceType) bool {
	if invoiceType != models.CustomerInvoiceTypeRegular {
		return false
	}

	if currentStatus != models.CustomerInvoiceStatusCancelled && currentStatus != models.CustomerInvoiceStatusReversed {
		return false
	}

	return previousStatus == models.CustomerInvoiceStatusApproved ||
		previousStatus == models.CustomerInvoiceStatusUnpaid ||
		previousStatus == models.CustomerInvoiceStatusWaitingPayment ||
		previousStatus == models.CustomerInvoiceStatusPartial ||
		previousStatus == models.CustomerInvoiceStatusPaid
}

func withActorContext(ctx context.Context, preferredUserID, fallbackUserID *string) context.Context {
	if actor, _ := ctx.Value("user_id").(string); strings.TrimSpace(actor) != "" {
		return ctx
	}

	if preferredUserID != nil && strings.TrimSpace(*preferredUserID) != "" {
		return context.WithValue(ctx, "user_id", strings.TrimSpace(*preferredUserID))
	}

	if fallbackUserID != nil && strings.TrimSpace(*fallbackUserID) != "" {
		return context.WithValue(ctx, "user_id", strings.TrimSpace(*fallbackUserID))
	}

	return ctx
}

func (uc *customerInvoiceUsecase) TriggerJournalForInvoice(ctx context.Context, invoice *models.CustomerInvoice) error {
	return uc.triggerSalesInvoiceJournal(ctx, invoice)
}

func (uc *customerInvoiceUsecase) PreviewJournal(ctx context.Context, req *dto.CreateCustomerInvoiceRequest) (*dto.CustomerInvoiceJournalPreviewResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if uc.engine == nil {
		return nil, errors.New("accounting engine is not configured")
	}

	soItemMap, err := uc.buildSalesOrderItemMapForCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	items, subtotal, err := uc.buildCreateInvoiceItems(ctx, req, soItemMap)
	if err != nil {
		return nil, err
	}

	taxRate := req.TaxRate
	if taxRate <= 0 {
		taxRate = 11
	}
	taxAmount := subtotal * (taxRate / 100)
	totalAmount := subtotal + taxAmount + req.DeliveryCost + req.OtherCost

	tempInvoice := &models.CustomerInvoice{
		Code:              "PREVIEW",
		InvoiceDate:       parsePreviewInvoiceDate(req.InvoiceDate),
		Type:              models.CustomerInvoiceType(strings.ToLower(strings.TrimSpace(req.Type))),
		SalesOrderID:      req.SalesOrderID,
		DeliveryOrderID:   req.DeliveryOrderID,
		TaxRate:           taxRate,
		TaxAmount:         taxAmount,
		DeliveryCost:      req.DeliveryCost,
		OtherCost:         req.OtherCost,
		Subtotal:          subtotal,
		Amount:            totalAmount,
		RemainingAmount:   totalAmount,
		DownPaymentAmount: 0,
		Items:             items,
	}
	if tempInvoice.Type == "" {
		tempInvoice.Type = models.CustomerInvoiceTypeRegular
	}
	if err := uc.applyPaidDownPaymentsOnCreate(ctx, tempInvoice, req.SalesOrderID, tempInvoice.Type, req.DownPaymentInvoiceID); err != nil {
		return nil, err
	}

	data := accounting.TransactionData{
		ReferenceType:   reference.RefTypeSalesInvoice,
		ReferenceID:     "preview",
		EntryDate:       tempInvoice.InvoiceDate.Format(dateFormat),
		Description:     fmt.Sprintf("Sales Invoice Preview %s", tempInvoice.Code),
		TotalAmount:     tempInvoice.Amount,
		SubTotal:        tempInvoice.Subtotal,
		TaxTotal:        tempInvoice.TaxAmount,
		DepositTotal:    tempInvoice.DownPaymentAmount,
		DescriptionArgs: []interface{}{tempInvoice.Code, safeInvoiceNumber(tempInvoice.InvoiceNumber)},
	}

	reqJournal, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSalesInvoice, data)
	if err != nil {
		return nil, err
	}

	preview := &dto.CustomerInvoiceJournalPreviewResponse{
		ReferenceType: reference.RefTypeSalesInvoice,
		ReferenceID:   "preview",
		InvoiceDate:   tempInvoice.InvoiceDate.Format(dateFormat),
		Subtotal:      tempInvoice.Subtotal,
		TaxAmount:     tempInvoice.TaxAmount,
		DownPayment:   tempInvoice.DownPaymentAmount,
		TotalAmount:   tempInvoice.Amount,
		IsBalanced:    isPreviewJournalBalanced(reqJournal),
		Lines:         make([]dto.CustomerInvoiceJournalPreviewLine, 0, len(reqJournal.Lines)),
	}
	for _, line := range reqJournal.Lines {
		preview.Lines = append(preview.Lines, dto.CustomerInvoiceJournalPreviewLine{
			ChartOfAccountID: line.ChartOfAccountID,
			Debit:            line.Debit,
			Credit:           line.Credit,
			Memo:             line.Memo,
		})
	}

	return preview, nil
}

func (uc *customerInvoiceUsecase) triggerSalesInvoiceJournal(ctx context.Context, invoice *models.CustomerInvoice) error {
	if invoice != nil && invoice.DeliveryOrderID != nil && invoice.DeliveryOrder == nil && uc.invoiceRepo != nil {
		if hydrated, err := uc.invoiceRepo.FindByID(ctx, invoice.ID); err == nil && hydrated != nil {
			invoice = hydrated
		}
	}

	if hasLinkedDeliverySalesJournal(invoice) {
		log.Printf("journal_observability event=trigger.skipped reason=delivery_order_already_posted reference_id=%s", invoice.ID)
		return nil
	}

	if invoice == nil || uc.journalUC == nil || uc.engine == nil {
		return nil
	}

	companyID, err := uc.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve company for sales invoice journal: %w", err)
	}

	revenueBase := invoice.Subtotal + invoice.DeliveryCost + invoice.OtherCost
	totalInvoiceAmount := revenueBase + invoice.TaxAmount
	cogsTotal := calculateInvoiceCOGSTotal(invoice.Items)
	journalAmount := totalInvoiceAmount
	if journalAmount <= 0 {
		journalAmount = invoice.Amount
	}
	if journalAmount <= 0 {
		journalAmount = revenueBase + invoice.TaxAmount
	}

	data := accounting.TransactionData{
		ReferenceType:   reference.RefTypeSalesInvoice,
		ReferenceID:     invoice.ID,
		CompanyID:       companyID,
		EntryDate:       invoice.InvoiceDate.Format(dateFormat),
		Description:     fmt.Sprintf("Sales Invoice %s", invoice.Code),
		TotalAmount:     journalAmount,
		SubTotal:        revenueBase,
		TaxTotal:        invoice.TaxAmount,
		DepositTotal:    invoice.DownPaymentAmount,
		COGSTotal:       cogsTotal,
		DescriptionArgs: []interface{}{invoice.Code, safeInvoiceNumber(invoice.InvoiceNumber)},
	}

	req, err := uc.engine.GenerateJournal(ctx, accounting.ProfileSalesInvoice, data)
	if err != nil {
		return fmt.Errorf("failed to generate sales invoice journal: %w", err)
	}

	postedJournal, err := uc.journalUC.PostOrUpdateJournal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to post sales invoice journal: %w", err)
	}

	if postedJournal != nil && strings.TrimSpace(postedJournal.ID) != "" && uc.db != nil {
		if err := database.GetDB(ctx, uc.db).
			Model(&models.CustomerInvoice{}).
			Where("id = ?", invoice.ID).
			Updates(map[string]interface{}{
				"journal_entry_id": postedJournal.ID,
				"is_posted":        true,
			}).Error; err != nil {
			return fmt.Errorf("failed to link sales invoice journal: %w", err)
		}
		invoice.JournalEntryID = &postedJournal.ID
		invoice.IsPosted = true
	}

	log.Printf("journal_observability event=trigger.success reference_type=%s reference_id=%s", reference.RefTypeSalesInvoice, invoice.ID)
	return nil
}

func parsePreviewInvoiceDate(value string) time.Time {
	parsed, err := time.Parse(dateFormat, value)
	if err != nil {
		return time.Now()
	}
	return parsed
}

func isPreviewJournalBalanced(req *finDto.CreateJournalEntryRequest) bool {
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

func (uc *customerInvoiceUsecase) resolveCompanyIDFromActor(ctx context.Context) (string, error) {
	if uc == nil || uc.db == nil {
		return "", errors.New("db is not configured")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return "", errors.New("user not authenticated")
	}

	var companyID string
	if err := database.GetDB(ctx, uc.db).
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

func (uc *customerInvoiceUsecase) triggerSalesDPApplicationJournal(ctx context.Context, invoice *models.CustomerInvoice) error {
	if invoice == nil || uc.salesJournalSvc == nil {
		return nil
	}

	_, err := uc.salesJournalSvc.GenerateDPApplicationJournal(ctx, invoice.ID)
	if err != nil {
		return err
	}

	log.Printf("journal_observability event=trigger.success reference_type=SALES_DP_APPLICATION reference_id=%s", invoice.ID)
	return nil
}

func calculateInvoiceCOGSTotal(items []models.CustomerInvoiceItem) float64 {
	total := 0.0
	for _, item := range items {
		if item.HPPAmount <= 0 || item.Quantity <= 0 {
			continue
		}
		total += item.HPPAmount * item.Quantity
	}

	return total
}

func (uc *customerInvoiceUsecase) triggerSalesInvoiceJournalReversed(ctx context.Context, invoice *models.CustomerInvoice, reason string) error {
	if invoice != nil && uc.salesJournalSvc != nil {
		if err := uc.salesJournalSvc.ReverseSalesJournal(ctx, "SALES_DP_APPLICATION", invoice.ID, reason); err != nil {
			return err
		}
	}

	if hasLinkedDeliverySalesJournal(invoice) {
		return nil
	}

	if invoice == nil || uc.journalUC == nil {
		return nil
	}

	refType := reference.RefTypeSalesInvoice
	var existing financeModels.JournalEntry
	err := database.GetDB(ctx, uc.db).
		Where("reference_type = ? AND reference_id = ?", refType, invoice.ID).
		Where("status = ?", financeModels.JournalStatusPosted).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = uc.journalUC.ReverseWithReason(ctx, existing.ID, reason)
	return err
}

func (uc *customerInvoiceUsecase) triggerSalesInvoiceJournalReversal(ctx context.Context, invoice *models.CustomerInvoice) error {
	return uc.triggerSalesInvoiceJournalReversed(ctx, invoice, "Manual reversal via status change")
}

// tryEarnLoyaltyPointsForInvoice awards loyalty points for a fully-paid non-POS customer invoice.
// It runs in a background goroutine — errors are logged and never surface to the caller.
func (uc *customerInvoiceUsecase) tryEarnLoyaltyPointsForInvoice(ctx context.Context, invoice *models.CustomerInvoice) {
	if uc.loyaltyUC == nil || invoice == nil {
		return
	}

	// Only process invoices that are linked to a Sales Order.
	if invoice.SalesOrderID == nil {
		return
	}

	// Fetch the linked Sales Order to check its source type and customer.
	so, err := uc.salesOrderRepo.FindByID(ctx, *invoice.SalesOrderID)
	if err != nil || so == nil {
		return
	}

	// Skip POS-sourced orders — loyalty points are already awarded by the POS payment flow.
	if so.SourceType == "POS" {
		return
	}

	if so.CustomerID == nil {
		return
	}

	// Lookup the loyalty member for this customer.
	memberResp, err := uc.loyaltyUC.GetMemberByCustomerID(ctx, *so.CustomerID)
	if err != nil || memberResp == nil {
		// Customer is not a loyalty member — nothing to do.
		return
	}

	_, err = uc.loyaltyUC.EarnPoints(ctx, &loyaltyDto.EarnPointsRequest{
		MemberID:        memberResp.ID,
		TransactionID:   invoice.ID,
		TransactionType: "sales_order",
		TotalAmount:     invoice.Amount,
	})
	if err != nil {
		log.Printf("[loyalty] failed to earn points for invoice %s (member %s): %v", invoice.ID, memberResp.ID, err)
	}
}

// isValidStatusTransition checks if the status transition is valid
func isValidStatusTransition(from, to models.CustomerInvoiceStatus) bool {
	validTransitions := map[models.CustomerInvoiceStatus][]models.CustomerInvoiceStatus{
		models.CustomerInvoiceStatusDraft:          {models.CustomerInvoiceStatusSubmitted, models.CustomerInvoiceStatusApproved, models.CustomerInvoiceStatusCancelled},
		models.CustomerInvoiceStatusSubmitted:      {models.CustomerInvoiceStatusApproved, models.CustomerInvoiceStatusUnpaid, models.CustomerInvoiceStatusRejected},
		models.CustomerInvoiceStatusApproved:       {models.CustomerInvoiceStatusUnpaid, models.CustomerInvoiceStatusPartial, models.CustomerInvoiceStatusPaid, models.CustomerInvoiceStatusCancelled, models.CustomerInvoiceStatusReversed},
		models.CustomerInvoiceStatusRejected:       {models.CustomerInvoiceStatusDraft},
		models.CustomerInvoiceStatusUnpaid:         {models.CustomerInvoiceStatusWaitingPayment, models.CustomerInvoiceStatusPartial, models.CustomerInvoiceStatusPaid, models.CustomerInvoiceStatusCancelled, models.CustomerInvoiceStatusReversed},
		models.CustomerInvoiceStatusWaitingPayment: {models.CustomerInvoiceStatusUnpaid, models.CustomerInvoiceStatusPartial, models.CustomerInvoiceStatusPaid, models.CustomerInvoiceStatusCancelled, models.CustomerInvoiceStatusReversed},
		models.CustomerInvoiceStatusPartial:        {models.CustomerInvoiceStatusWaitingPayment, models.CustomerInvoiceStatusPaid, models.CustomerInvoiceStatusCancelled, models.CustomerInvoiceStatusReversed},
		models.CustomerInvoiceStatusPaid:           {models.CustomerInvoiceStatusReversed},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
