package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

// ToPOSOrderItemResponse maps a PosOrderItem to its DTO response
func ToPOSOrderItemResponse(item *models.PosOrderItem) dto.POSOrderItemResponse {
	return dto.POSOrderItemResponse{
		ID:          item.ID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		ProductCode: item.ProductCode,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Discount:    item.Discount,
		Subtotal:    item.Subtotal,
		Notes:       item.Notes,
		Status:      string(item.Status),
	}
}

// ToPOSOrderResponse maps a PosOrder (with preloaded Items) to its DTO response
func ToPOSOrderResponse(order *models.PosOrder) *dto.POSOrderResponse {
	items := make([]dto.POSOrderItemResponse, 0, len(order.Items))
	for i := range order.Items {
		items = append(items, ToPOSOrderItemResponse(&order.Items[i]))
	}

	return &dto.POSOrderResponse{
		ID:                order.ID,
		TenantID:          order.TenantID,
		OrderNumber:       order.OrderNumber,
		SessionID:         order.SessionID,
		OutletID:          order.OutletID,
		OrderType:         string(order.OrderType),
		TableID:           order.TableID,
		TableLabel:        order.TableLabel,
		CustomerID:        order.CustomerID,
		CustomerName:      order.CustomerName,
		GuestCount:        order.GuestCount,
		Subtotal:          order.Subtotal,
		DiscountAmount:    order.DiscountAmount,
		TaxAmount:         order.TaxAmount,
		ServiceCharge:     order.ServiceCharge,
		TotalAmount:       order.TotalAmount,
		Status:            string(order.Status),
		OrderSource:       order.OrderSource,
		VoidReason:        order.VoidReason,
		Notes:             order.Notes,
		SalesOrderID:      order.SalesOrderID,
		CustomerInvoiceID: order.CustomerInvoiceID,
		LoyaltyMemberID:   order.LoyaltyMemberID,
		LoyaltyRewardID:   order.LoyaltyRewardID,
		Items:             items,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}
}

// ToPOSSessionResponse maps a PosSession to its DTO response
func ToPOSSessionResponse(s *models.PosSession) *dto.POSSessionResponse {
	openedAt := s.OpenedAt.Format(time.RFC3339)
	createdAt := s.CreatedAt.Format(time.RFC3339)

	var closedAtStr *string
	if s.ClosedAt != nil {
		t := s.ClosedAt.Format(time.RFC3339)
		closedAtStr = &t
	}

	return &dto.POSSessionResponse{
		ID:          s.ID,
		Code:        s.Code,
		OutletID:    s.OutletID,
		WarehouseID: s.WarehouseID,
		CashierID:   s.CashierID,
		OpeningCash: s.OpeningCash,
		ClosingCash: s.ClosingCash,
		Status:      string(s.Status),
		TotalSales:  s.TotalSales,
		TotalOrders: s.TotalOrders,
		OpenedAt:    openedAt,
		ClosedAt:    closedAtStr,
		Notes:       s.Notes,
		CreatedAt:   createdAt,
	}
}

// ToPOSConfigResponse maps a POSConfig to its DTO response
func ToPOSConfigResponse(c *models.POSConfig) *dto.POSConfigResponse {
	return &dto.POSConfigResponse{
		ID:                      c.ID,
		OutletID:                c.OutletID,
		TaxRate:                 c.TaxRate,
		ServiceChargeRate:       c.ServiceChargeRate,
		AllowDiscount:           c.AllowDiscount,
		MaxDiscountPercent:      c.MaxDiscountPercent,
		PrintReceiptAuto:        c.PrintReceiptAuto,
		ReceiptFooter:           c.ReceiptFooter,
		ReceiptWhatsAppTemplate: c.ReceiptWhatsAppTemplate,
		Currency:                c.Currency,
	}
}
