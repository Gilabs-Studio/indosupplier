package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	crmRepos "github.com/gilabs/gims/api/internal/crm/data/repositories"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrSalesQuotationNotFound        = errors.New("sales quotation not found")
	ErrSalesQuotationAlreadyExists   = errors.New("sales quotation with this code already exists")
	ErrInvalidStatusTransition       = errors.New("invalid status transition")
	ErrQuotationAlreadyConverted     = errors.New("quotation already converted to sales order")
	ErrProductNotFound               = errors.New("product not found")
	ErrInvalidQuotationStatus        = errors.New("cannot modify quotation in current status")
	ErrValidUntilBeforeQuotationDate = errors.New("valid until date must not be earlier than quotation date")
)

// SalesQuotationUsecase defines the interface for sales quotation business logic
type SalesQuotationUsecase interface {
	List(ctx context.Context, req *dto.ListSalesQuotationsRequest) ([]dto.SalesQuotationResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.SalesQuotationResponse, error)
	ListItems(ctx context.Context, quotationID string, req *dto.ListSalesQuotationItemsRequest) ([]dto.SalesQuotationItemResponse, *utils.PaginationResult, error)
	Create(ctx context.Context, req *dto.CreateSalesQuotationRequest, createdBy *string) (*dto.SalesQuotationResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSalesQuotationRequest) (*dto.SalesQuotationResponse, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, req *dto.UpdateSalesQuotationStatusRequest, userID *string) (*dto.SalesQuotationResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
}

type salesQuotationUsecase struct {
	db            *gorm.DB
	quotationRepo salesRepos.SalesQuotationRepository
	customerRepo  customerRepos.CustomerRepository
	contactRepo   crmRepos.ContactRepository
	productRepo   productRepos.ProductRepository
	auditService  audit.AuditService
}

// NewSalesQuotationUsecase creates a new SalesQuotationUsecase
func NewSalesQuotationUsecase(
	db *gorm.DB,
	quotationRepo salesRepos.SalesQuotationRepository,
	productRepo productRepos.ProductRepository,
	auditService audit.AuditService,
) SalesQuotationUsecase {
	return &salesQuotationUsecase{
		db:            db,
		quotationRepo: quotationRepo,
		customerRepo:  customerRepos.NewCustomerRepository(db),
		contactRepo:   crmRepos.NewContactRepository(db),
		productRepo:   productRepo,
		auditService:  auditService,
	}
}

func (u *salesQuotationUsecase) List(ctx context.Context, req *dto.ListSalesQuotationsRequest) ([]dto.SalesQuotationResponse, *utils.PaginationResult, error) {
	quotations, total, err := u.quotationRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.SalesQuotationResponse, len(quotations))
	for i := range quotations {
		responses[i] = mapper.ToSalesQuotationResponse(&quotations[i])
		u.attachCustomerContactResponse(ctx, &responses[i])
	}

	// Calculate pagination
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

	return responses, pagination, nil
}

func (u *salesQuotationUsecase) ListItems(ctx context.Context, quotationID string, req *dto.ListSalesQuotationItemsRequest) ([]dto.SalesQuotationItemResponse, *utils.PaginationResult, error) {
	// Verify quotation exists
	_, err := u.quotationRepo.FindByID(ctx, quotationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSalesQuotationNotFound
		}
		return nil, nil, err
	}

	// Fetch paginated items
	items, total, err := u.quotationRepo.ListItems(ctx, quotationID, req)
	if err != nil {
		return nil, nil, err
	}

	// Map to response DTOs
	responses := make([]dto.SalesQuotationItemResponse, len(items))
	for i := range items {
		responses[i] = mapper.ToSalesQuotationItemResponse(&items[i])
	}

	// Calculate pagination
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

	return responses, pagination, nil
}

func (u *salesQuotationUsecase) GetByID(ctx context.Context, id string) (*dto.SalesQuotationResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesQuotation{}, id, security.SalesScopeQueryOptions()) {
		return nil, ErrSalesQuotationNotFound
	}

	quotation, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesQuotationNotFound
		}
		return nil, err
	}

	response := mapper.ToSalesQuotationResponse(quotation)
	u.attachCustomerContactResponse(ctx, &response)
	return &response, nil
}

func (u *salesQuotationUsecase) Create(ctx context.Context, req *dto.CreateSalesQuotationRequest, createdBy *string) (*dto.SalesQuotationResponse, error) {
	// Validate products exist and get default prices
	for _, item := range req.Items {
		product, err := u.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, err
		}

		// Use product selling price if price not provided
		if item.Price == 0 {
			item.Price = product.SellingPrice
		}
	}

	// Generate quotation number
	code, err := u.quotationRepo.GetNextQuotationNumber(ctx, "SQ")
	if err != nil {
		return nil, err
	}

	// Convert request to model
	quotation, err := mapper.ToSalesQuotationModel(req, code, createdBy)
	if err != nil {
		return nil, err
	}

	// Validate date relationship
	if err := validateQuotationDates(quotation.QuotationDate, quotation.ValidUntil); err != nil {
		return nil, err
	}

	if err := u.applyCustomerSnapshot(ctx, quotation); err != nil {
		return nil, err
	}

	// Calculate totals
	u.calculateTotals(quotation)

	// Set valid until based on payment terms if not provided
	if quotation.ValidUntil == nil && quotation.PaymentTermsID != nil {
		// This would require fetching payment terms, for now we'll set a default
		// In a real implementation, you'd fetch payment terms and add days
		validUntil := quotation.QuotationDate.AddDate(0, 0, 30) // Default 30 days
		quotation.ValidUntil = &validUntil
	}

	// Create quotation
	if err := u.quotationRepo.Create(ctx, quotation); err != nil {
		return nil, err
	}

	// Fetch created quotation with relations
	created, err := u.quotationRepo.FindByID(ctx, quotation.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesQuotationResponse(created)
	u.attachCustomerContactResponse(ctx, &response)
	logSalesAudit(u.auditService, ctx, "sales_quotation.create", created.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":           created.Code,
			"status":         created.Status,
			"quotation_date": created.QuotationDate,
			"total_amount":   created.TotalAmount,
		},
	})
	return &response, nil
}

func (u *salesQuotationUsecase) Update(ctx context.Context, id string, req *dto.UpdateSalesQuotationRequest) (*dto.SalesQuotationResponse, error) {
	quotation, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesQuotationNotFound
		}
		return nil, err
	}

	// Check if quotation can be modified
	if quotation.Status != models.SalesQuotationStatusDraft {
		return nil, ErrInvalidQuotationStatus
	}

	beforeSnapshot := salesQuotationAuditSnapshot(quotation)

	// Validate products if items are being updated
	if req.Items != nil && len(*req.Items) > 0 {
		for i := range *req.Items {
			item := &(*req.Items)[i]
			product, err := u.productRepo.FindByID(ctx, item.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrProductNotFound
				}
				return nil, err
			}

			// Use product selling price if price not provided
			if item.Price == 0 {
				item.Price = product.SellingPrice
			}
		}
	}

	// Update model
	if err := mapper.UpdateSalesQuotationModel(quotation, req); err != nil {
		return nil, err
	}

	// Validate date relationship
	if err := validateQuotationDates(quotation.QuotationDate, quotation.ValidUntil); err != nil {
		return nil, err
	}

	if err := u.applyCustomerSnapshot(ctx, quotation); err != nil {
		return nil, err
	}

	// Recalculate totals
	u.calculateTotals(quotation)

	// Update quotation
	if err := u.quotationRepo.Update(ctx, quotation); err != nil {
		return nil, err
	}

	// Fetch updated quotation with relations
	updated, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesQuotationResponse(updated)
	u.attachCustomerContactResponse(ctx, &response)
	afterSnapshot := salesQuotationAuditSnapshot(updated)
	if shouldLogSnapshotChange(beforeSnapshot, afterSnapshot) {
		logSalesAudit(u.auditService, ctx, "sales_quotation.update", id, map[string]interface{}{
			"before": beforeSnapshot,
			"after":  afterSnapshot,
		})
	}
	return &response, nil
}

func (u *salesQuotationUsecase) Delete(ctx context.Context, id string) error {
	quotation, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSalesQuotationNotFound
		}
		return err
	}

	// Only allow deletion of draft quotations
	if quotation.Status != models.SalesQuotationStatusDraft {
		return ErrInvalidQuotationStatus
	}

	return u.quotationRepo.Delete(ctx, id)
}

func (u *salesQuotationUsecase) UpdateStatus(ctx context.Context, id string, req *dto.UpdateSalesQuotationStatusRequest, userID *string) (*dto.SalesQuotationResponse, error) {
	quotation, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesQuotationNotFound
		}
		return nil, err
	}

	newStatus := models.SalesQuotationStatus(req.Status)
	previousStatus := quotation.Status

	// Validate status transition
	if !u.isValidStatusTransition(quotation.Status, newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	// Update status
	var reason *string
	if req.RejectionReason != nil {
		reason = req.RejectionReason
	}

	if err := u.quotationRepo.UpdateStatus(ctx, id, newStatus, userID, reason); err != nil {
		return nil, err
	}

	// Fetch updated quotation
	updated, err := u.quotationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesQuotationResponse(updated)
	u.attachCustomerContactResponse(ctx, &response)

	if newStatus == models.SalesQuotationStatusSent {
		if err := notificationService.CreateApprovalNotification(ctx, u.db, notificationService.ApprovalNotificationParams{
			PermissionCode: "sales_quotation.approve",
			EntityType:     "sales_quotation",
			EntityID:       updated.ID,
			Title:          "Sales quotation approval required",
			Message:        fmt.Sprintf("Sales quotation %s requires approval and review.", updated.Code),
			ActorUserID:    stringValue(userID),
		}); err != nil {
			fmt.Printf("failed to create sales quotation approval notification: %v\n", err)
		}
	}

	logSalesAudit(u.auditService, ctx, "sales_quotation.status_change", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  updated.Status,
		"reason":        req.RejectionReason,
	})
	return &response, nil
}

func (u *salesQuotationUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesQuotation{}, id, security.SalesScopeQueryOptions()) {
		return nil, 0, ErrSalesQuotationNotFound
	}

	return listAuditTrailEntries(ctx, u.db, id, "sales_quotation.", page, perPage)
}

// calculateTotals calculates all financial totals for the quotation
func (u *salesQuotationUsecase) calculateTotals(quotation *models.SalesQuotation) {
	// Calculate subtotal from items
	subtotal := 0.0
	for i := range quotation.Items {
		quotation.Items[i].CalculateSubtotal()
		subtotal += quotation.Items[i].Subtotal
	}

	quotation.Subtotal = subtotal

	// Apply discount
	subtotalAfterDiscount := quotation.Subtotal - quotation.DiscountAmount
	if subtotalAfterDiscount < 0 {
		subtotalAfterDiscount = 0
	}

	// Calculate tax (on subtotal after discount)
	if quotation.TaxRate == 0 {
		quotation.TaxRate = 11.00 // Default 11% PPN
	}
	quotation.TaxAmount = subtotalAfterDiscount * (quotation.TaxRate / 100.0)

	// Calculate total: Subtotal - Discount + Tax + Delivery + Other
	quotation.TotalAmount = subtotalAfterDiscount + quotation.TaxAmount + quotation.DeliveryCost + quotation.OtherCost
}

// isValidStatusTransition validates if status transition is allowed
func (u *salesQuotationUsecase) isValidStatusTransition(current, new models.SalesQuotationStatus) bool {
	validTransitions := map[models.SalesQuotationStatus][]models.SalesQuotationStatus{
		models.SalesQuotationStatusDraft: {
			models.SalesQuotationStatusSent,
			models.SalesQuotationStatusApproved,
			models.SalesQuotationStatusRejected,
		},
		models.SalesQuotationStatusSent: {
			models.SalesQuotationStatusApproved,
			models.SalesQuotationStatusRejected,
		},
		models.SalesQuotationStatusApproved: {
			models.SalesQuotationStatusConverted,
		},
		models.SalesQuotationStatusRejected: {
			models.SalesQuotationStatusDraft, // Can be revised
		},
		models.SalesQuotationStatusConverted: {
			// Cannot transition from converted
		},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}

	return false
}

func (u *salesQuotationUsecase) applyCustomerSnapshot(ctx context.Context, quotation *models.SalesQuotation) error {
	if quotation != nil && quotation.CustomerContactID != nil && *quotation.CustomerContactID != "" && u.contactRepo != nil {
		contact, err := u.contactRepo.FindByID(ctx, *quotation.CustomerContactID)
		if err == nil {
			if quotation.CustomerID == nil || *quotation.CustomerID == "" || contact.CustomerID == *quotation.CustomerID {
				quotation.CustomerContact = salesQuotationFirstNonEmpty(quotation.CustomerContact, contact.Name)
				quotation.CustomerEmail = salesQuotationFirstNonEmpty(quotation.CustomerEmail, contact.Email)
				quotation.CustomerPhone = salesQuotationFirstNonEmpty(quotation.CustomerPhone, contact.Phone)
			}
		}
	}

	if quotation == nil || quotation.CustomerID == nil || *quotation.CustomerID == "" || u.customerRepo == nil {
		return nil
	}

	customer, err := u.customerRepo.FindByID(ctx, *quotation.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	quotation.CustomerName = customer.Name
	quotation.CustomerContact = salesQuotationFirstNonEmpty(quotation.CustomerContact, customer.ContactPerson)
	quotation.CustomerEmail = salesQuotationFirstNonEmpty(quotation.CustomerEmail, customer.Email)
	quotation.CustomerPhone = salesQuotationFirstNonEmpty(quotation.CustomerPhone, salesQuotationResolvePrimaryPhone(customer))

	return nil
}

func (u *salesQuotationUsecase) attachCustomerContactResponse(ctx context.Context, response *dto.SalesQuotationResponse) {
	if response == nil || response.CustomerContactID == nil || *response.CustomerContactID == "" || u.contactRepo == nil {
		return
	}

	contact, err := u.contactRepo.FindByID(ctx, *response.CustomerContactID)
	if err != nil {
		return
	}

	response.CustomerContactRef = &dto.CustomerContactResponse{
		ID:    contact.ID,
		Name:  contact.Name,
		Phone: contact.Phone,
		Email: contact.Email,
	}
}

// validateQuotationDates ensures valid_until is not earlier than quotation_date
func validateQuotationDates(quotationDate time.Time, validUntil *time.Time) error {
	if validUntil == nil {
		return nil
	}
	if validUntil.Before(quotationDate) {
		return ErrValidUntilBeforeQuotationDate
	}
	return nil
}

func salesQuotationFirstNonEmpty(current string, fallback string) string {
	if current != "" {
		return current
	}
	return fallback
}

func salesQuotationResolvePrimaryPhone(customer *customerModels.Customer) string {
	_ = customer
	return ""
}

func salesQuotationPaymentTermsName(quotation *models.SalesQuotation) string {
	if quotation == nil || quotation.PaymentTerms == nil {
		return ""
	}
	return quotation.PaymentTerms.Name
}

func salesQuotationBusinessUnitName(quotation *models.SalesQuotation) string {
	if quotation == nil || quotation.BusinessUnit == nil {
		return ""
	}
	return quotation.BusinessUnit.Name
}

func salesQuotationBusinessTypeName(quotation *models.SalesQuotation) string {
	if quotation == nil || quotation.BusinessType == nil {
		return ""
	}
	return quotation.BusinessType.Name
}

func salesQuotationSalesRepName(quotation *models.SalesQuotation) string {
	if quotation == nil || quotation.SalesRep == nil {
		return ""
	}
	return quotation.SalesRep.Name
}

func salesQuotationAuditSnapshot(quotation *models.SalesQuotation) map[string]interface{} {
	if quotation == nil {
		return nil
	}

	return map[string]interface{}{
		"code":               quotation.Code,
		"status":             quotation.Status,
		"quotation_date":     quotation.QuotationDate,
		"valid_until":        quotation.ValidUntil,
		"customer_id":        quotation.CustomerID,
		"customer_name":      quotation.CustomerName,
		"customer_contact":   quotation.CustomerContact,
		"customer_phone":     quotation.CustomerPhone,
		"customer_email":     quotation.CustomerEmail,
		"payment_terms_id":   quotation.PaymentTermsID,
		"payment_terms_name": salesQuotationPaymentTermsName(quotation),
		"sales_rep_id":       quotation.SalesRepID,
		"sales_rep_name":     salesQuotationSalesRepName(quotation),
		"business_unit_id":   quotation.BusinessUnitID,
		"business_unit_name": salesQuotationBusinessUnitName(quotation),
		"business_type_id":   quotation.BusinessTypeID,
		"business_type_name": salesQuotationBusinessTypeName(quotation),
		"subtotal":           quotation.Subtotal,
		"discount_amount":    quotation.DiscountAmount,
		"tax_rate":           quotation.TaxRate,
		"tax_amount":         quotation.TaxAmount,
		"delivery_cost":      quotation.DeliveryCost,
		"other_cost":         quotation.OtherCost,
		"total_amount":       quotation.TotalAmount,
		"notes":              quotation.Notes,
		"items":              salesQuotationAuditItems(quotation.Items),
	}
}

func salesQuotationAuditItems(items []models.SalesQuotationItem) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := ""
		code := ""
		if item.Product != nil {
			name = item.Product.Name
			code = item.Product.Code
		}

		out = append(out, map[string]interface{}{
			"id":           item.ID,
			"product_id":   item.ProductID,
			"product_code": code,
			"product_name": name,
			"quantity":     item.Quantity,
			"price":        item.Price,
			"discount":     item.Discount,
			"subtotal":     item.Subtotal,
		})
	}

	return out
}
