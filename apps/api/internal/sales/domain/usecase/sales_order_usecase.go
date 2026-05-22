package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	crmRepos "github.com/gilabs/gims/api/internal/crm/data/repositories"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	inventoryUsecase "github.com/gilabs/gims/api/internal/inventory/domain/usecase"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	organizationRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	salesQuotationRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrSalesOrderNotFound           = errors.New("sales order not found")
	ErrSalesOrderAlreadyExists      = errors.New("sales order with this code already exists")
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")
	ErrOrderProductNotFound         = errors.New("product not found in order")
	ErrInvalidOrderStatus           = errors.New("cannot modify order in current status")
	ErrOrderItemsLockedByQuotation  = errors.New("sales order items are locked because this order is sourced from a quotation")
	ErrQuotationNotFound            = errors.New("sales quotation not found")
	ErrQuotationNotApproved         = errors.New("quotation must be approved before converting to order")
	ErrInsufficientStock            = errors.New("insufficient stock available")
	ErrUnauthorizedAccess           = errors.New("unauthorized access to sales order")
	ErrCreditLimitExceeded          = errors.New("customer credit limit exceeded")
)

// SalesOrderUsecase defines the interface for sales order business logic
type SalesOrderUsecase interface {
	List(ctx context.Context, req *dto.ListSalesOrdersRequest) ([]dto.SalesOrderResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.SalesOrderResponse, error)
	ListItems(ctx context.Context, orderID string, req *dto.ListSalesOrderItemsRequest) ([]dto.SalesOrderItemResponse, *utils.PaginationResult, error)
	Create(ctx context.Context, req *dto.CreateSalesOrderRequest, createdBy *string) (*dto.SalesOrderResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSalesOrderRequest) (*dto.SalesOrderResponse, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, req *dto.UpdateSalesOrderStatusRequest, userID *string) (*dto.SalesOrderResponse, error)
	ConvertFromQuotation(ctx context.Context, req *dto.ConvertFromQuotationRequest, createdBy *string) (*dto.SalesOrderResponse, error)
	ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error)
}

type salesOrderUsecase struct {
	db                *gorm.DB
	orderRepo         salesRepos.SalesOrderRepository
	deliveryOrderRepo salesRepos.DeliveryOrderRepository
	quotationRepo     salesQuotationRepos.SalesQuotationRepository
	customerRepo      customerRepos.CustomerRepository
	contactRepo       crmRepos.ContactRepository
	productRepo       productRepos.ProductRepository
	inventoryUC       inventoryUsecase.InventoryUsecase
	employeeRepo      organizationRepos.EmployeeRepository
	auditService      audit.AuditService
}

// NewSalesOrderUsecase creates a new SalesOrderUsecase
func NewSalesOrderUsecase(
	db *gorm.DB,
	orderRepo salesRepos.SalesOrderRepository,
	deliveryOrderRepo salesRepos.DeliveryOrderRepository,
	quotationRepo salesQuotationRepos.SalesQuotationRepository,
	productRepo productRepos.ProductRepository,
	inventoryUC inventoryUsecase.InventoryUsecase,
	employeeRepo organizationRepos.EmployeeRepository,
) SalesOrderUsecase {
	return &salesOrderUsecase{
		db:                db,
		orderRepo:         orderRepo,
		deliveryOrderRepo: deliveryOrderRepo,
		quotationRepo:     quotationRepo,
		customerRepo:      customerRepos.NewCustomerRepository(db),
		contactRepo:       crmRepos.NewContactRepository(db),
		productRepo:       productRepo,
		inventoryUC:       inventoryUC,
		employeeRepo:      employeeRepo,
		auditService:      audit.NewAuditService(db),
	}
}

func (u *salesOrderUsecase) List(ctx context.Context, req *dto.ListSalesOrdersRequest) ([]dto.SalesOrderResponse, *utils.PaginationResult, error) {
	// Scope filtering is now handled at the repository level via ApplyScopeFilter.
	// The ScopeMiddleware + RequirePermission middleware inject scope context values
	// that the repository reads to apply OWN/DIVISION/AREA/ALL filtering.

	orders, total, err := u.orderRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.SalesOrderResponse, len(orders))
	for i := range orders {
		// For approved orders, fetch pending delivery quantities
		var pendingQtyMap map[string]float64
		if orders[i].Status == models.SalesOrderStatusApproved {
			pendingQtyMap, _ = u.deliveryOrderRepo.GetPendingDeliveryQtyBySalesOrder(ctx, orders[i].ID)
		}
		responses[i] = mapper.ToSalesOrderResponse(&orders[i], pendingQtyMap)
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

func (u *salesOrderUsecase) ListItems(ctx context.Context, orderID string, req *dto.ListSalesOrderItemsRequest) ([]dto.SalesOrderItemResponse, *utils.PaginationResult, error) {
	// Verify order exists
	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrSalesOrderNotFound
		}
		return nil, nil, err
	}

	// Access Control
	if err := u.checkAccess(ctx, order); err != nil {
		return nil, nil, err
	}

	// Fetch paginated items
	items, total, err := u.orderRepo.ListItems(ctx, orderID, req)
	if err != nil {
		return nil, nil, err
	}

	// Map to response DTOs
	responses := make([]dto.SalesOrderItemResponse, len(items))
	for i := range items {
		responses[i] = mapper.ToSalesOrderItemResponse(&items[i], 0)
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

func (u *salesOrderUsecase) GetByID(ctx context.Context, id string) (*dto.SalesOrderResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesOrder{}, id, security.SalesScopeQueryOptions()) {
		return nil, ErrSalesOrderNotFound
	}
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesOrderNotFound
		}
		return nil, err
	}

	// Fetch pending delivery quantities for this order
	pendingQtyMap, _ := u.deliveryOrderRepo.GetPendingDeliveryQtyBySalesOrder(ctx, order.ID)

	response := mapper.ToSalesOrderResponse(order, pendingQtyMap)
	u.attachCustomerContactResponse(ctx, &response)
	return &response, nil
}

func (u *salesOrderUsecase) Create(ctx context.Context, req *dto.CreateSalesOrderRequest, createdBy *string) (*dto.SalesOrderResponse, error) {
	// Validate products exist and get default prices
	productMap := make(map[string]*productModels.Product)
	for i, item := range req.Items {
		product, err := u.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrOrderProductNotFound
			}
			return nil, err
		}

		// Use product selling price if price not provided
		if item.Price == 0 {
			req.Items[i].Price = product.SellingPrice
		}

		productMap[item.ProductID] = product
	}

	// Generate order number
	code, err := u.orderRepo.GetNextOrderNumber(ctx, "SO")
	if err != nil {
		return nil, err
	}

	// 2. Convert request to model
	order, err := mapper.ToSalesOrderModel(req, code, createdBy)
	if err != nil {
		return nil, err
	}

	if err := u.applyCustomerSnapshot(ctx, order); err != nil {
		return nil, err
	}

	// Calculate totals early for credit check
	u.calculateTotals(order)

	// 3. Perform Credit Control Check
	if err := u.checkCreditLimit(ctx, order); err != nil {
		return nil, err
	}

	// Populate snapshot fields
	for i := range order.Items {
		if p, ok := productMap[order.Items[i].ProductID]; ok {
			order.Items[i].ProductCode = p.Code
			order.Items[i].ProductName = p.Name
		}
	}

	// Create order
	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Update quotation status if linked
	if order.SalesQuotationID != nil {
		if err := u.quotationRepo.UpdateStatus(ctx, *order.SalesQuotationID, models.SalesQuotationStatusConverted, createdBy, nil); err != nil {
			// Log error but don't fail transaction as order is created
		}
	}

	// Fetch created order with relations
	created, err := u.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesOrderResponse(created, nil)
	u.attachCustomerContactResponse(ctx, &response)
	logSalesAudit(u.auditService, ctx, "sales_order.create", created.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":         created.Code,
			"status":       created.Status,
			"order_date":   created.OrderDate,
			"total_amount": created.TotalAmount,
		},
	})
	return &response, nil
}

func (u *salesOrderUsecase) Update(ctx context.Context, id string, req *dto.UpdateSalesOrderRequest) (*dto.SalesOrderResponse, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesOrderNotFound
		}
		return nil, err
	}

	// Check if order can be modified
	if err := u.checkAccess(ctx, order); err != nil {
		return nil, err
	}

	// Check if order can be modified
	if order.Status != models.SalesOrderStatusDraft {
		return nil, ErrInvalidOrderStatus
	}

	beforeSnapshot := salesOrderAuditSnapshot(order)

	if order.SalesQuotationID != nil && len(req.Items) > 0 {
		return nil, ErrOrderItemsLockedByQuotation
	}

	// Validate products if items are being updated
	productMap := make(map[string]*productModels.Product)
	if len(req.Items) > 0 {
		for i, item := range req.Items {
			product, err := u.productRepo.FindByID(ctx, item.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrOrderProductNotFound
				}
				return nil, err
			}

			// Use product selling price if price not provided
			if item.Price == 0 {
				req.Items[i].Price = product.SellingPrice
			}

			productMap[item.ProductID] = product
		}
	}

	// Update model
	if err := mapper.UpdateSalesOrderModel(order, req); err != nil {
		return nil, err
	}

	if err := u.applyCustomerSnapshot(ctx, order); err != nil {
		return nil, err
	}

	// Recalculate totals
	u.calculateTotals(order)

	// Populate snapshot fields if items updated
	if len(req.Items) > 0 {
		for i := range order.Items {
			if p, ok := productMap[order.Items[i].ProductID]; ok {
				order.Items[i].ProductCode = p.Code
				order.Items[i].ProductName = p.Name
			}
		}
	}

	// Update order
	if err := u.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	// Fetch updated order with relations
	updated, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesOrderResponse(updated, nil)
	u.attachCustomerContactResponse(ctx, &response)
	afterSnapshot := salesOrderAuditSnapshot(updated)
	if shouldLogSnapshotChange(beforeSnapshot, afterSnapshot) {
		logSalesAudit(u.auditService, ctx, "sales_order.update", id, map[string]interface{}{
			"before": beforeSnapshot,
			"after":  afterSnapshot,
		})
	}
	return &response, nil
}

func (u *salesOrderUsecase) Delete(ctx context.Context, id string) error {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSalesOrderNotFound
		}
		return err
	}

	// Check if order can be modified
	if err := u.checkAccess(ctx, order); err != nil {
		return err
	}

	// Only allow deletion of draft orders
	if order.Status != models.SalesOrderStatusDraft {
		return ErrInvalidOrderStatus
	}

	// Release stock if reserved
	if order.ReservedStock {
		// Release stock in inventory
		for _, item := range order.Items {
			if err := u.inventoryUC.ReleaseStock(ctx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}

		if err := u.orderRepo.ReleaseStock(ctx, id); err != nil {
			return err
		}
	}

	return u.orderRepo.Delete(ctx, id)
}

func (u *salesOrderUsecase) UpdateStatus(ctx context.Context, id string, req *dto.UpdateSalesOrderStatusRequest, userID *string) (*dto.SalesOrderResponse, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesOrderNotFound
		}
		return nil, err
	}

	// Check if order can be modified
	if err := u.checkAccess(ctx, order); err != nil {
		return nil, err
	}

	newStatus := models.SalesOrderStatus(req.Status)
	previousStatus := order.Status

	// Validate status transition
	if !u.isValidStatusTransition(order.Status, newStatus) {
		return nil, ErrInvalidOrderStatusTransition
	}

	// Check credit control limit on approval
	if newStatus == models.SalesOrderStatusApproved {
		if err := u.checkCreditLimit(ctx, order); err != nil {
			return nil, err
		}
	}

	// Handle stock reservation on approval (wrapped in transaction for atomicity)
	if newStatus == models.SalesOrderStatusApproved && !order.ReservedStock {
		err := database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
			txCtx := database.WithTx(ctx, tx)

			// Reserve stock at product level for each item
			for _, item := range order.Items {
				if err := u.inventoryUC.ReserveStock(txCtx, item.ProductID, item.Quantity); err != nil {
					return err
				}
			}

			// Mark as reserved in SO
			if err := u.orderRepo.ReserveStock(txCtx, id); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Handle stock release on cancellation (wrapped in transaction for atomicity)
	if newStatus == models.SalesOrderStatusCancelled && order.ReservedStock {
		err := database.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
			txCtx := database.WithTx(ctx, tx)

			// Release stock at product level for each item
			for _, item := range order.Items {
				if err := u.inventoryUC.ReleaseStock(txCtx, item.ProductID, item.Quantity); err != nil {
					return err
				}
			}

			// Mark as released in SO
			if err := u.orderRepo.ReleaseStock(txCtx, id); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Update status
	var reason *string
	if req.CancellationReason != nil {
		reason = req.CancellationReason
	}

	if err := u.orderRepo.UpdateStatus(ctx, id, newStatus, userID, reason); err != nil {
		return nil, err
	}

	// Fetch updated order
	updated, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesOrderResponse(updated, nil)
	u.attachCustomerContactResponse(ctx, &response)

	if newStatus == models.SalesOrderStatusSubmitted {
		if err := notificationService.CreateApprovalNotification(ctx, u.db, notificationService.ApprovalNotificationParams{
			PermissionCode: "sales_order.approve",
			EntityType:     "sales_order",
			EntityID:       updated.ID,
			Title:          "Sales order approval required",
			Message:        fmt.Sprintf("Sales order %s requires approval and review.", updated.Code),
			ActorUserID:    stringValue(userID),
		}); err != nil {
			fmt.Printf("failed to create sales order approval notification: %v\n", err)
		}

		logSalesAudit(u.auditService, ctx, "sales_order.submit", id, map[string]interface{}{
			"before_status": previousStatus,
			"after_status":  updated.Status,
		})
	}

	logSalesAudit(u.auditService, ctx, "sales_order.status_change", id, map[string]interface{}{
		"before_status": previousStatus,
		"after_status":  updated.Status,
		"reason":        req.CancellationReason,
	})
	return &response, nil
}

func (u *salesOrderUsecase) ConvertFromQuotation(ctx context.Context, req *dto.ConvertFromQuotationRequest, createdBy *string) (*dto.SalesOrderResponse, error) {
	// Fetch quotation
	quotation, err := u.quotationRepo.FindByID(ctx, req.QuotationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuotationNotFound
		}
		return nil, err
	}

	// Validate quotation status
	if quotation.Status != models.SalesQuotationStatusApproved {
		return nil, ErrQuotationNotApproved
	}

	// Generate order number
	code, err := u.orderRepo.GetNextOrderNumber(ctx, "SO")
	if err != nil {
		return nil, err
	}

	// Convert quotation to order (pass customer info — use request fields with quotation fallback)
	customerName := req.CustomerName
	if customerName == "" {
		customerName = quotation.CustomerName
	}
	customerContact := req.CustomerContact
	if customerContact == "" {
		customerContact = quotation.CustomerContact
	}
	customerPhone := req.CustomerPhone
	if customerPhone == "" {
		customerPhone = quotation.CustomerPhone
	}
	customerEmail := req.CustomerEmail
	if customerEmail == "" {
		customerEmail = quotation.CustomerEmail
	}
	order, err := mapper.ConvertQuotationToOrderModel(quotation, req.DeliveryAreaID, req.CustomerContactID, customerName, customerContact, customerPhone, customerEmail, req.Notes, code, createdBy)
	if err != nil {
		return nil, err
	}

	// Create order
	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Update quotation status to converted
	if err := u.quotationRepo.UpdateStatus(ctx, quotation.ID, models.SalesQuotationStatusConverted, createdBy, nil); err != nil {
		log.Printf("Warning: failed to update quotation %s status to converted: %v", quotation.ID, err)
	}

	// Fetch created order with relations
	created, err := u.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		log.Printf("Warning: failed to fetch created order %s after conversion: %v", order.ID, err)
		response := mapper.ToSalesOrderResponse(order, nil)
		u.attachCustomerContactResponse(ctx, &response)
		logSalesAudit(u.auditService, ctx, "sales_order.create", order.ID, map[string]interface{}{
			"after": map[string]interface{}{
				"code":               order.Code,
				"status":             order.Status,
				"order_date":         order.OrderDate,
				"total_amount":       order.TotalAmount,
				"sales_quotation_id": req.QuotationID,
			},
		})
		return &response, nil
	}

	response := mapper.ToSalesOrderResponse(created, nil)
	u.attachCustomerContactResponse(ctx, &response)
	logSalesAudit(u.auditService, ctx, "sales_order.create", created.ID, map[string]interface{}{
		"after": map[string]interface{}{
			"code":               created.Code,
			"status":             created.Status,
			"order_date":         created.OrderDate,
			"total_amount":       created.TotalAmount,
			"sales_quotation_id": req.QuotationID,
		},
	})
	return &response, nil
}

func (u *salesOrderUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesOrder{}, id, security.SalesScopeQueryOptions()) {
		return nil, 0, ErrSalesOrderNotFound
	}

	return listAuditTrailEntries(ctx, u.db, id, "sales_order.", page, perPage)
}

// checkAccess verifies if the current user has access to the order
func (u *salesOrderUsecase) checkAccess(ctx context.Context, order *models.SalesOrder) error {
	userRole, _ := ctx.Value("user_role").(string)
	userID, _ := ctx.Value("user_id").(string)

	// Admin and Manager bypass checks
	if userRole == "admin" || userRole == "manager" {
		return nil
	}

	// If no user context, assume internal call or unsecured?
	if userID == "" {
		// Secure by default: if we don't know who you are, you can't touch it.
		// Unless it's a system process (which might not have user_id but should be handled via different context or role)
		return ErrSalesOrderNotFound
	}

	// Check if user is the creator
	if order.CreatedBy != nil && *order.CreatedBy == userID {
		return nil
	}

	// Check if user is the assigned Sales Rep
	// Optimization: Avoid DB call if SalesRepID matches directly
	if order.SalesRepID != nil {
		// First check if the user IS the sales rep directly (if user ID matches employee ID logic, but usually they are different)
		// We need to fetch employee record for the user to compare with SalesRepID
		employee, err := u.employeeRepo.FindByUserID(ctx, userID)
		if err != nil {
			// If user is not an employee, they probably shouldn't see orders
			return ErrSalesOrderNotFound
		}

		if *order.SalesRepID == employee.ID {
			return nil
		}
	}

	return ErrSalesOrderNotFound // Access Denied (Obfuscated)
}

func (u *salesOrderUsecase) applyCustomerSnapshot(ctx context.Context, order *models.SalesOrder) error {
	if order != nil && order.CustomerContactID != nil && *order.CustomerContactID != "" && u.contactRepo != nil {
		contact, err := u.contactRepo.FindByID(ctx, *order.CustomerContactID)
		if err == nil {
			if order.CustomerID == nil || *order.CustomerID == "" || contact.CustomerID == *order.CustomerID {
				order.CustomerContact = salesOrderFirstNonEmpty(order.CustomerContact, contact.Name)
				order.CustomerEmail = salesOrderFirstNonEmpty(order.CustomerEmail, contact.Email)
				order.CustomerPhone = salesOrderFirstNonEmpty(order.CustomerPhone, contact.Phone)
			}
		}
	}

	if order == nil || order.CustomerID == nil || *order.CustomerID == "" || u.customerRepo == nil {
		return nil
	}

	customer, err := u.customerRepo.FindByID(ctx, *order.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	order.CustomerName = customer.Name
	order.CustomerContact = salesOrderFirstNonEmpty(order.CustomerContact, customer.ContactPerson)
	order.CustomerEmail = salesOrderFirstNonEmpty(order.CustomerEmail, customer.Email)
	order.CustomerPhone = salesOrderFirstNonEmpty(order.CustomerPhone, salesOrderResolvePrimaryPhone(customer))

	return nil
}

func (u *salesOrderUsecase) attachCustomerContactResponse(ctx context.Context, response *dto.SalesOrderResponse) {
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

func salesOrderFirstNonEmpty(current string, fallback string) string {
	if current != "" {
		return current
	}
	return fallback
}

func salesOrderResolvePrimaryPhone(customer *customerModels.Customer) string {
	_ = customer
	return ""
}

// calculateTotals calculates all financial totals for the order
func (u *salesOrderUsecase) calculateTotals(order *models.SalesOrder) {
	// Calculate subtotal from items
	subtotal := 0.0
	for i := range order.Items {
		order.Items[i].CalculateSubtotal()
		subtotal += order.Items[i].Subtotal
	}

	order.Subtotal = subtotal

	// Apply discount
	subtotalAfterDiscount := order.Subtotal - order.DiscountAmount
	if subtotalAfterDiscount < 0 {
		subtotalAfterDiscount = 0
	}

	// Calculate tax (on subtotal after discount)
	if order.TaxRate == 0 {
		order.TaxRate = 11.00 // Default 11% PPN
	}
	order.TaxAmount = subtotalAfterDiscount * (order.TaxRate / 100.0)

	// Calculate total: Subtotal - Discount + Tax + Delivery + Other
	order.TotalAmount = subtotalAfterDiscount + order.TaxAmount + order.DeliveryCost + order.OtherCost
}

// isValidStatusTransition validates if status transition is allowed
func (u *salesOrderUsecase) isValidStatusTransition(current, new models.SalesOrderStatus) bool {
	validTransitions := map[models.SalesOrderStatus][]models.SalesOrderStatus{
		models.SalesOrderStatusDraft: {
			models.SalesOrderStatusSubmitted,
			models.SalesOrderStatusCancelled,
		},
		models.SalesOrderStatusSubmitted: {
			models.SalesOrderStatusApproved,
			models.SalesOrderStatusRejected,
		},
		models.SalesOrderStatusApproved: {
			models.SalesOrderStatusClosed,
			models.SalesOrderStatusCancelled,
		},
		models.SalesOrderStatusClosed: {
			// Cannot transition from closed
		},
		models.SalesOrderStatusRejected: {
			models.SalesOrderStatusDraft,
		},
		models.SalesOrderStatusCancelled: {
			// Cannot transition from cancelled
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

func salesOrderPaymentTermsName(order *models.SalesOrder) string {
	if order == nil || order.PaymentTerms == nil {
		return ""
	}
	return order.PaymentTerms.Name
}

func salesOrderBusinessUnitName(order *models.SalesOrder) string {
	if order == nil || order.BusinessUnit == nil {
		return ""
	}
	return order.BusinessUnit.Name
}

func salesOrderBusinessTypeName(order *models.SalesOrder) string {
	if order == nil || order.BusinessType == nil {
		return ""
	}
	return order.BusinessType.Name
}

func salesOrderDeliveryAreaName(order *models.SalesOrder) string {
	if order == nil || order.DeliveryArea == nil {
		return ""
	}
	return order.DeliveryArea.Name
}

func salesOrderSalesRepName(order *models.SalesOrder) string {
	if order == nil || order.SalesRep == nil {
		return ""
	}
	return order.SalesRep.Name
}

func (u *salesOrderUsecase) checkCreditLimit(ctx context.Context, order *models.SalesOrder) error {
	if strings.EqualFold(strings.TrimSpace(order.SourceType), "POS") {
		return nil
	}

	if order.CustomerID == nil || *order.CustomerID == "" {
		return nil
	}

	customer, err := u.customerRepo.FindByID(ctx, *order.CustomerID)
	if err != nil {
		return nil // Customer not found, skip credit check or handle as error?
	}

	if !customer.CreditIsActive || customer.CreditLimit <= 0 {
		return nil
	}

	// Get outstanding balance (Unpaid Invoices)
	var outstanding float64
	err = u.db.Table("customer_invoices").
		Joins("JOIN sales_orders ON sales_orders.id = customer_invoices.sales_order_id").
		Where("sales_orders.customer_id = ?", *order.CustomerID).
		Where("customer_invoices.status IN ?", []string{
			"UNPAID", "PARTIAL", "WAITING_PAYMENT",
		}).
		Where("customer_invoices.deleted_at IS NULL").
		Select("COALESCE(SUM(customer_invoices.remaining_amount), 0)").
		Scan(&outstanding).Error

	if err != nil {
		return fmt.Errorf("failed to check customer credit: %w", err)
	}

	if outstanding+order.TotalAmount > customer.CreditLimit {
		// Check for credit override permission
		perms, ok := ctx.Value("user_permissions").(map[string]bool)
		if ok && perms["sales_order.credit_override"] {
			return nil // Override allowed
		}

		return fmt.Errorf("%w: limit %.2f, outstanding %.2f, new order %.2f",
			ErrCreditLimitExceeded, customer.CreditLimit, outstanding, order.TotalAmount)
	}

	return nil
}

func salesOrderAuditSnapshot(order *models.SalesOrder) map[string]interface{} {
	if order == nil {
		return nil
	}

	return map[string]interface{}{
		"code":               order.Code,
		"status":             order.Status,
		"order_date":         order.OrderDate,
		"sales_quotation_id": order.SalesQuotationID,
		"customer_id":        order.CustomerID,
		"customer_name":      order.CustomerName,
		"customer_contact":   order.CustomerContact,
		"customer_phone":     order.CustomerPhone,
		"customer_email":     order.CustomerEmail,
		"payment_terms_id":   order.PaymentTermsID,
		"payment_terms_name": salesOrderPaymentTermsName(order),
		"sales_rep_id":       order.SalesRepID,
		"sales_rep_name":     salesOrderSalesRepName(order),
		"business_unit_id":   order.BusinessUnitID,
		"business_unit_name": salesOrderBusinessUnitName(order),
		"business_type_id":   order.BusinessTypeID,
		"business_type_name": salesOrderBusinessTypeName(order),
		"delivery_area_id":   order.DeliveryAreaID,
		"delivery_area_name": salesOrderDeliveryAreaName(order),
		"subtotal":           order.Subtotal,
		"discount_amount":    order.DiscountAmount,
		"tax_rate":           order.TaxRate,
		"tax_amount":         order.TaxAmount,
		"delivery_cost":      order.DeliveryCost,
		"other_cost":         order.OtherCost,
		"total_amount":       order.TotalAmount,
		"notes":              order.Notes,
		"reserved_stock":     order.ReservedStock,
		"items":              salesOrderAuditItems(order.Items),
	}
}

func salesOrderAuditItems(items []models.SalesOrderItem) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := item.ProductName
		if name == "" && item.Product != nil {
			name = item.Product.Name
		}

		code := item.ProductCode
		if code == "" && item.Product != nil {
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
