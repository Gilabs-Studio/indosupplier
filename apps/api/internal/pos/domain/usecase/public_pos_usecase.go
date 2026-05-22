package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	posModels "github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/provider"
	"gorm.io/gorm"
)

// ErrPublicTokenInvalid is returned when the scanned QR token does not match any active record.
var ErrPublicTokenInvalid = errors.New("invalid or expired QR token")

// PublicPOSUsecase handles customer-facing self-order operations.
type PublicPOSUsecase interface {
	// GetTableInfo resolves a QR token and returns the table + menu payload.
	GetTableInfo(ctx context.Context, token string) (*dto.PublicTableInfoResponse, error)
	// CreateCustomerOrder places a new order on behalf of a customer.
	// Returns the simplified order response and publishes a WS event.
	CreateCustomerOrder(ctx context.Context, token string, req *dto.CreateCustomerOrderRequest) (*dto.CustomerOrderResponse, error)
	// GetOrderStatus returns the current status and totals for a customer order.
	GetOrderStatus(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error)
	// InitiateDigitalPayment creates a Xendit payment invoice for a customer order.
	InitiateDigitalPayment(ctx context.Context, token, orderID string, req *dto.InitiateCustomerPaymentRequest) (*dto.POSPaymentResponse, error)
	// MarkPayAtCashier marks the order as waiting cashier payment and broadcasts status updates.
	MarkPayAtCashier(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error)
	// CancelCustomerOrder cancels a self-order if it is not yet processed.
	CancelCustomerOrder(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error)
	// CancelDigitalPayment cancels a pending digital payment and falls back to cashier waiting mode.
	CancelDigitalPayment(ctx context.Context, token, orderID, paymentID string) (*dto.CustomerOrderResponse, error)
}

// POSHubPublisher is a narrow interface used by this usecase to broadcast WS events.
// It avoids an import cycle with the ws infrastructure package.
type POSHubPublisher interface {
	Publish(tenantID string, outletID string, eventType string, payload map[string]interface{})
}

type publicPOSUsecase struct {
	qrTokenRepo     repositories.TableQRTokenRepository
	orderUC         POSOrderUsecase
	paymentUC       POSPaymentUsecase
	configRepo      repositories.POSConfigRepository
	outletRepo      orgRepos.OutletRepository
	productRepo     repositories.POSProductRepository
	hub             POSHubPublisher
	deviceTokenRepo repositories.POSDeviceTokenRepository
	pushNotifier    provider.PushNotifier
	db              *gorm.DB
}

// NewPublicPOSUsecase creates a PublicPOSUsecase.
// hub may be nil to disable WebSocket broadcasts (useful in tests).
func NewPublicPOSUsecase(
	qrTokenRepo repositories.TableQRTokenRepository,
	orderUC POSOrderUsecase,
	paymentUC POSPaymentUsecase,
	configRepo repositories.POSConfigRepository,
	outletRepo orgRepos.OutletRepository,
	productRepo repositories.POSProductRepository,
	hub POSHubPublisher,
	deviceTokenRepo repositories.POSDeviceTokenRepository,
	pushNotifier provider.PushNotifier,
	db *gorm.DB,
) PublicPOSUsecase {
	return &publicPOSUsecase{
		qrTokenRepo:     qrTokenRepo,
		orderUC:         orderUC,
		paymentUC:       paymentUC,
		configRepo:      configRepo,
		outletRepo:      outletRepo,
		productRepo:     productRepo,
		hub:             hub,
		deviceTokenRepo: deviceTokenRepo,
		pushNotifier:    pushNotifier,
		db:              db,
	}
}

// ─── GetTableInfo ─────────────────────────────────────────────────────────────

func (u *publicPOSUsecase) GetTableInfo(ctx context.Context, token string) (*dto.PublicTableInfoResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	outlet, err := u.outletRepo.GetByID(ctx, qr.OutletID)
	if err != nil || outlet == nil {
		return nil, ErrPublicTokenInvalid
	}

	// Fetch catalog directly (bypasses RBAC auth-scope used by staff GetCatalog).
	warehouseID := ""
	if outlet.WarehouseID != nil {
		warehouseID = *outlet.WarehouseID
	}
	rawProducts, _ := u.productRepo.FindPOSAvailable(ctx, warehouseID, qr.OutletID)
	catalog := make([]dto.POSCatalogItem, 0, len(rawProducts))
	for _, p := range rawProducts {
		catalog = append(catalog, dto.POSCatalogItem{
			ProductID:   p.ProductID,
			ProductCode: p.ProductCode,
			ProductName: p.ProductName,
			ProductKind: p.ProductKind,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
			Category:    p.Category,
			IsAvailable: p.IsAvailable,
		})
	}

	config, _ := u.configRepo.FindByOutletID(ctx, qr.OutletID)
	pubConfig := dto.PublicPOSConfig{Currency: "IDR"}
	if config != nil {
		pubConfig.TaxRate = config.TaxRate
		pubConfig.ServiceChargeRate = config.ServiceChargeRate
		pubConfig.Currency = config.Currency
	}

	return &dto.PublicTableInfoResponse{
		OutletID:   qr.OutletID,
		OutletName: outlet.Name,
		TableLabel: qr.TableLabel,
		Token:      token,
		Catalog:    catalog,
		Config:     pubConfig,
	}, nil
}

// ─── CreateCustomerOrder ──────────────────────────────────────────────────────

func (u *publicPOSUsecase) CreateCustomerOrder(ctx context.Context, token string, req *dto.CreateCustomerOrderRequest) (*dto.CustomerOrderResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	// Build context for unauthenticated customer operations (RBAC + tenant scoping).
	orderCtx, tenantID, _, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	// Build a staff-compatible CreateOrderRequest with OrderSource overridden inside usecase.
	tableObjectID := qr.TableObjectID
	tableLabel := qr.TableLabel
	var tableID *string
	if tableObjectID != "" {
		tableID = &tableObjectID
	}
	createReq := &dto.CreateOrderRequest{
		OutletID:     qr.OutletID,
		OrderType:    "DINE_IN",
		TableID:      tableID,
		TableLabel:   &tableLabel,
		CustomerName: &req.CustomerName,
		Notes:        req.Notes,
	}
	order, err := u.orderUC.Create(orderCtx, createReq, "customer")
	if err != nil {
		return nil, err
	}

	// Attach each item from the cart.
	for _, item := range req.Items {
		pid := item.ProductID
		qty := item.Quantity
		addReq := &dto.AddOrderItemRequest{ProductID: pid, Quantity: qty, Notes: item.Notes}
		if _, err := u.orderUC.AddItem(orderCtx, order.ID, addReq); err != nil {
			return nil, err
		}
	}

	// Re-fetch to get updated totals after all items.
	updated, err := u.orderUC.GetByID(orderCtx, order.ID)
	if err != nil {
		return nil, err
	}

	// Broadcast WS event so staff dashboard refreshes instantly.
	if u.hub != nil {
		u.hub.Publish(tenantID, qr.OutletID, "pos.new_customer_order", posOrderRealtimePayload(updated, "WAITING_CASHIER", ""))
	}
	u.pushNewOrderNotification(ctx, tenantID, qr.OutletID, updated)

	// Fan-out in-app notification to all staff with pos.order.read access for this outlet.
	if u.db != nil {
		outletName := ""
		if outlet, err := u.outletRepo.GetByID(ctx, qr.OutletID); err == nil && outlet != nil {
			outletName = outlet.Name
		}
		notifParams := notificationService.POSOrderNotificationParams{
			TenantID:   tenantID,
			OutletID:   qr.OutletID,
			OutletName: outletName,
			TableLabel: tableLabel,
			OrderID:    updated.ID,
		}
		go func() {
			_ = notificationService.CreatePOSOrderNotification(context.Background(), u.db, notifParams)
		}()
	}

	return u.toCustomerOrderResponse(orderCtx, updated, tableLabel), nil
}

func (u *publicPOSUsecase) pushNewOrderNotification(ctx context.Context, tenantID, outletID string, order *dto.POSOrderResponse) {
	if u.deviceTokenRepo == nil || u.pushNotifier == nil || order == nil {
		return
	}
	tokens, err := u.deviceTokenRepo.FindByScope(ctx, tenantID, outletID)
	if err != nil {
		return
	}
	tableLabel := "Meja"
	if order.TableLabel != nil && strings.TrimSpace(*order.TableLabel) != "" {
		tableLabel = strings.TrimSpace(*order.TableLabel)
	}
	_ = u.pushNotifier.Send(ctx, tokens, provider.PushPayload{
		Title: "Order Baru",
		Body:  tableLabel,
		Data: map[string]string{
			"outlet_id": outletID,
			"tenant_id": tenantID,
			"order_id":  order.ID,
		},
	})
}

// ─── GetOrderStatus ───────────────────────────────────────────────────────────

func (u *publicPOSUsecase) GetOrderStatus(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	orderCtx, _, _, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	order, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	// Ensure the order belongs to this outlet (IDOR guard).
	if order.OutletID != qr.OutletID {
		return nil, ErrPublicTokenInvalid
	}

	tableLabel := qr.TableLabel
	if order.TableLabel != nil {
		tableLabel = *order.TableLabel
	}

	return u.toCustomerOrderResponse(orderCtx, order, tableLabel), nil
}

// ─── InitiateDigitalPayment ───────────────────────────────────────────────────

func (u *publicPOSUsecase) InitiateDigitalPayment(ctx context.Context, token, orderID string, req *dto.InitiateCustomerPaymentRequest) (*dto.POSPaymentResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	orderCtx, tenantID, companyID, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	order, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	// IDOR guard: order must belong to the outlet encoded in the token.
	if order.OutletID != qr.OutletID {
		return nil, ErrPublicTokenInvalid
	}

	payReq := &dto.ProcessPaymentRequest{
		Method:      req.Method,
		Amount:      req.Amount,
		ChannelCode: req.ChannelCode,
	}

	// Use a stable public actor label; payment usecase normalizes this into a valid UUID.
	payment, err := u.paymentUC.InitiateDigitalPayment(orderCtx, orderID, payReq, "customer", companyID)
	if err != nil {
		return nil, err
	}

	if u.hub != nil {
		tableLabel := qr.TableLabel
		if order.TableLabel != nil {
			tableLabel = *order.TableLabel
		}
		payload := posOrderRealtimePayload(order, "WAITING_DIGITAL", "")
		payload["table_label"] = tableLabel
		u.hub.Publish(tenantID, qr.OutletID, "pos.order_status_changed", payload)
	}

	return payment, nil
}

func (u *publicPOSUsecase) MarkPayAtCashier(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	orderCtx, tenantID, _, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	order, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	if order.OutletID != qr.OutletID {
		return nil, ErrPublicTokenInvalid
	}

	if order.Status == string(posModels.PosOrderStatusPaid) || order.Status == string(posModels.PosOrderStatusCompleted) {
		return nil, ErrPOSOrderAlreadyPaid
	}
	if order.Status == string(posModels.PosOrderStatusVoided) {
		return nil, ErrPOSOrderCannotModify
	}

	// Move draft orders into active kitchen/cashier workflow.
	if order.Status == string(posModels.PosOrderStatusDraft) {
		if _, err := u.orderUC.Confirm(orderCtx, orderID, &dto.ConfirmOrderRequest{}, "customer"); err != nil {
			return nil, err
		}
	}

	updated, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	tableLabel := qr.TableLabel
	if updated.TableLabel != nil {
		tableLabel = *updated.TableLabel
	}

	if u.hub != nil {
		payload := posOrderRealtimePayload(updated, "WAITING_CASHIER", "")
		payload["table_label"] = tableLabel
		u.hub.Publish(tenantID, qr.OutletID, "pos.order_status_changed", payload)
	}

	return u.toCustomerOrderResponse(orderCtx, updated, tableLabel), nil
}

func (u *publicPOSUsecase) CancelCustomerOrder(ctx context.Context, token, orderID string) (*dto.CustomerOrderResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	orderCtx, tenantID, _, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	order, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	if order.OutletID != qr.OutletID {
		return nil, ErrPOSOrderNotFound
	}

	// Customer can only cancel if it's not yet being prepared and not paid.
	if order.Status != string(posModels.PosOrderStatusDraft) && order.Status != "WAITING_PAYMENT" && order.Status != "WAITING_CASHIER" && order.Status != "PENDING" {
		return nil, ErrPOSOrderCannotModify
	}

	// Use Void from POSOrderUsecase
	_, err = u.orderUC.Void(orderCtx, orderID, &dto.VoidOrderRequest{
		Reason: "Cancelled by customer via self-order UI",
	}, "customer")
	if err != nil {
		return nil, err
	}

	updated, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	tableLabel := qr.TableLabel
	if updated.TableLabel != nil {
		tableLabel = *updated.TableLabel
	}

	if u.hub != nil {
		payload := posOrderRealtimePayload(updated, "", "")
		payload["table_label"] = tableLabel
		u.hub.Publish(tenantID, qr.OutletID, "pos.order_status_changed", payload)
	}

	return u.toCustomerOrderResponse(orderCtx, updated, tableLabel), nil
}

func (u *publicPOSUsecase) CancelDigitalPayment(ctx context.Context, token, orderID, paymentID string) (*dto.CustomerOrderResponse, error) {
	qr, err := u.qrTokenRepo.FindByToken(ctx, token)
	if err != nil || qr == nil || !qr.IsActive {
		return nil, ErrPublicTokenInvalid
	}

	orderCtx, tenantID, _, err := u.buildPublicOrderContext(ctx, qr.OutletID)
	if err != nil {
		return nil, ErrPublicTokenInvalid
	}

	order, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	if order.OutletID != qr.OutletID {
		return nil, ErrPublicTokenInvalid
	}

	reason := "Cancelled by customer via self-order"
	if err := u.paymentUC.CancelPendingPayment(orderCtx, orderID, paymentID, "customer", &reason); err != nil {
		return nil, err
	}

	updated, err := u.orderUC.GetByID(orderCtx, orderID)
	if err != nil {
		return nil, err
	}

	tableLabel := qr.TableLabel
	if updated.TableLabel != nil {
		tableLabel = *updated.TableLabel
	}

	if u.hub != nil {
		payload := posOrderRealtimePayload(updated, "WAITING_CASHIER", "")
		payload["table_label"] = tableLabel
		u.hub.Publish(tenantID, qr.OutletID, "pos.order_status_changed", payload)
	}

	return u.toCustomerOrderResponse(orderCtx, updated, tableLabel), nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (u *publicPOSUsecase) buildPublicOrderContext(ctx context.Context, outletID string) (context.Context, string, string, error) {
	outlet, err := u.outletRepo.GetByID(ctx, outletID)
	if err != nil || outlet == nil {
		return nil, "", "", ErrPublicTokenInvalid
	}

	publicCtx := context.WithValue(ctx, "permission_scope", "ALL")
	publicCtx = context.WithValue(publicCtx, "tenant_id", outlet.TenantID)

	companyID := ""
	if outlet.CompanyID != nil {
		companyID = *outlet.CompanyID
	}

	return publicCtx, outlet.TenantID, companyID, nil
}

func (u *publicPOSUsecase) toCustomerOrderResponse(ctx context.Context, order *dto.POSOrderResponse, tableLabel string) *dto.CustomerOrderResponse {
	var paymentStatus *string
	var cancelReason *string
	payments, err := u.paymentUC.GetByOrderID(ctx, order.ID)
	if err == nil && len(payments) > 0 {
		latest := payments[0].Status
		paymentStatus = &latest
		for _, payment := range payments {
			if payment.Status != "CANCELLED" {
				continue
			}
			if payment.Notes == nil {
				continue
			}
			note := strings.TrimSpace(*payment.Notes)
			if note == "" {
				continue
			}
			cancelReason = &note
			break
		}
	}

	// If no cancel reason from payment notes, check order's VoidReason (for direct void operations)
	if cancelReason == nil && order.VoidReason != nil {
		reason := strings.TrimSpace(*order.VoidReason)
		if reason != "" {
			cancelReason = &reason
		}
	}

	return &dto.CustomerOrderResponse{
		OrderID:       order.ID,
		OrderNumber:   order.OrderNumber,
		TableLabel:    tableLabel,
		Status:        order.Status,
		PaymentStatus: paymentStatus,
		CancelReason:  cancelReason,
		Subtotal:      order.Subtotal,
		TaxAmount:     order.TaxAmount,
		TotalAmount:   order.TotalAmount,
		Items:         order.Items,
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
	}
}
