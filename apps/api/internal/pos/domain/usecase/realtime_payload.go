package usecase

import (
	"context"
	"encoding/json"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

func posOrderRealtimePayload(order *dto.POSOrderResponse, paymentStatus string, cancelReason string) map[string]interface{} {
	orderMap := map[string]interface{}{}
	if order != nil {
		raw, err := json.Marshal(order)
		if err == nil {
			_ = json.Unmarshal(raw, &orderMap)
		}
	}

	payload := map[string]interface{}{
		"order": orderMap,
	}
	for key, value := range orderMap {
		payload[key] = value
	}

	if order != nil {
		payload["order_id"] = order.ID
		payload["order_number"] = order.OrderNumber
		payload["outlet_id"] = order.OutletID
		payload["status"] = order.Status
		if order.TableID != nil {
			payload["table_id"] = *order.TableID
		}
		if order.TableLabel != nil {
			payload["table_label"] = *order.TableLabel
		}
		if order.CustomerName != nil {
			payload["customer"] = *order.CustomerName
			payload["customer_name"] = *order.CustomerName
		}
		payload["total_amount"] = order.TotalAmount
	}
	if paymentStatus != "" {
		payload["payment_status"] = paymentStatus
	}
	if cancelReason != "" {
		payload["cancel_reason"] = cancelReason
	}
	return payload
}

func tenantIDForPOSOrder(ctx context.Context, order *models.PosOrder) string {
	if order != nil && order.TenantID != "" {
		return order.TenantID
	}
	return scopeString(ctx, "tenant_id")
}
