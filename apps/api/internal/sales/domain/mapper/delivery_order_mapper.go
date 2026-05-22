package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

// ToDeliveryOrderResponse converts a DeliveryOrder model to response DTO
func ToDeliveryOrderResponse(m *salesModels.DeliveryOrder) dto.DeliveryOrderResponse {
	var warehouseID string
	if m.WarehouseID != nil {
		warehouseID = *m.WarehouseID
	}

	response := dto.DeliveryOrderResponse{
		ID:                m.ID,
		Code:              m.Code,
		DeliveryDate:      m.DeliveryDate.Format("2006-01-02"),
		WarehouseID:       warehouseID,
		SalesOrderID:      m.SalesOrderID,
		TrackingNumber:    m.TrackingNumber,
		ReceiverName:      m.ReceiverName,
		ReceiverPhone:     m.ReceiverPhone,
		DeliveryAddress:   m.DeliveryAddress,
		ReceiverSignature: m.ReceiverSignature,
		Status:            string(m.Status),
		Notes:             m.Notes,
		IsPartialDelivery: m.IsPartialDelivery,
		CreatedAt:         m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         m.UpdatedAt.Format(time.RFC3339),
	}

	if m.Warehouse != nil {
		response.Warehouse = &dto.WarehouseResponse{
			ID:   m.Warehouse.ID,
			Code: m.Warehouse.Code,
			Name: m.Warehouse.Name,
		}
	}

	if m.SalesOrder != nil {
		salesOrderResp := ToSalesOrderResponse(m.SalesOrder, nil)
		response.SalesOrder = &salesOrderResp
	}

	if m.DeliveredByID != nil {
		response.DeliveredByID = m.DeliveredByID
		if m.DeliveredBy != nil {
			response.DeliveredBy = &dto.EmployeeResponse{
				ID:           m.DeliveredBy.ID,
				EmployeeCode: m.DeliveredBy.EmployeeCode,
				Name:         m.DeliveredBy.Name,
				Email:        m.DeliveredBy.Email,
				Phone:        m.DeliveredBy.Phone,
			}
		}
	}

	if m.CourierAgencyID != nil {
		response.CourierAgencyID = m.CourierAgencyID
		if m.CourierAgency != nil {
			response.CourierAgency = &dto.CourierAgencyResponse{
				ID:          m.CourierAgency.ID,
				Code:        m.CourierAgency.Code,
				Name:        m.CourierAgency.Name,
				Description: m.CourierAgency.Description,
				Phone:       m.CourierAgency.Phone,
				Address:     m.CourierAgency.Address,
				TrackingURL: m.CourierAgency.TrackingURL,
			}
		}
	}

	if m.CreatedBy != nil {
		response.CreatedBy = m.CreatedBy
	}

	if m.ShippedBy != nil {
		response.ShippedBy = m.ShippedBy
		if m.ShippedAt != nil {
			shippedAt := m.ShippedAt.Format(time.RFC3339)
			response.ShippedAt = &shippedAt
		}
	}

	if m.DeliveredAt != nil {
		deliveredAt := m.DeliveredAt.Format(time.RFC3339)
		response.DeliveredAt = &deliveredAt
	}

	if m.CancelledBy != nil {
		response.CancelledBy = m.CancelledBy
		if m.CancelledAt != nil {
			cancelledAt := m.CancelledAt.Format(time.RFC3339)
			response.CancelledAt = &cancelledAt
		}
		response.CancellationReason = m.CancellationReason
	}

	if m.JournalEntryID != nil {
		response.JournalEntryID = m.JournalEntryID
	}

	// Map items
	if len(m.Items) > 0 {
		response.Items = make([]dto.DeliveryOrderItemResponse, len(m.Items))
		for i, item := range m.Items {
			response.Items[i] = ToDeliveryOrderItemResponse(&item)
		}
	}

	return response
}

// ToDeliveryOrderItemResponse converts a DeliveryOrderItem model to response DTO
func ToDeliveryOrderItemResponse(m *salesModels.DeliveryOrderItem) dto.DeliveryOrderItemResponse {
	response := dto.DeliveryOrderItemResponse{
		ID:                m.ID,
		DeliveryOrderID:   m.DeliveryOrderID,
		WarehouseID:       m.WarehouseID,
		ProductID:         m.ProductID,
		Quantity:          m.Quantity,
		Price:             m.Price,
		Subtotal:          m.Subtotal,
		AvgCostSnapshot:   m.AvgCostSnapshot,
		COGSAmount:        m.COGSAmount,
		IsEquipment:       m.IsEquipment,
		InstallationNotes: m.InstallationNotes,
		CreatedAt:         m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         m.UpdatedAt.Format(time.RFC3339),
	}

	if m.Warehouse != nil {
		response.Warehouse = &dto.WarehouseResponse{
			ID:   m.Warehouse.ID,
			Code: m.Warehouse.Code,
			Name: m.Warehouse.Name,
		}
	}

	if m.SalesOrderItemID != nil {
		response.SalesOrderItemID = m.SalesOrderItemID
		if m.SalesOrderItem != nil {
			salesOrderItemResp := ToSalesOrderItemResponse(m.SalesOrderItem, 0)
			response.SalesOrderItem = &salesOrderItemResp
		}
	}

	if m.InventoryBatchID != nil {
		response.InventoryBatchID = m.InventoryBatchID
		if m.InventoryBatch != nil {
			var expiryDate time.Time
			if m.InventoryBatch.ExpiryDate != nil {
				expiryDate = *m.InventoryBatch.ExpiryDate
			}

			response.InventoryBatch = &dto.BatchInfo{
				ID:          m.InventoryBatch.ID,
				BatchNumber: m.InventoryBatch.BatchNumber,
				ExpiryDate:  expiryDate,
				Quantity:    m.InventoryBatch.InitialQuantity,
				Available:   m.InventoryBatch.CurrentQuantity - m.InventoryBatch.ReservedQuantity,
				// ReceivedDate not strictly available in InventoryBatch model directly, usually from GR. Skipping or using CreatedAt as fallback if essential, but zero value is fine for now.
			}
		}
	}

	if m.Product != nil {
		response.Product = &dto.ProductResponse{
			ID:           m.Product.ID,
			Code:         m.Product.Code,
			Name:         m.Product.Name,
			SellingPrice: m.Product.SellingPrice,
			ImageURL:     m.Product.ImageURL,
		}
	}

	if m.InstallationStatus != nil {
		response.InstallationStatus = m.InstallationStatus
	}

	if m.FunctionTestStatus != nil {
		response.FunctionTestStatus = m.FunctionTestStatus
	}

	if m.InstallationDate != nil {
		installationDate := m.InstallationDate.Format(time.RFC3339)
		response.InstallationDate = &installationDate
	}

	if m.FunctionTestDate != nil {
		functionTestDate := m.FunctionTestDate.Format(time.RFC3339)
		response.FunctionTestDate = &functionTestDate
	}

	return response
}

// ToDeliveryOrderModel converts a CreateDeliveryOrderRequest to DeliveryOrder model
func ToDeliveryOrderModel(req *dto.CreateDeliveryOrderRequest, code string, createdBy *string) (*salesModels.DeliveryOrder, error) {
	deliveryDate, err := time.Parse("2006-01-02", req.DeliveryDate)
	if err != nil {
		return nil, err
	}

	var warehouseID *string
	if req.WarehouseID != "" {
		warehouseID = &req.WarehouseID
	}

	deliveryOrder := &salesModels.DeliveryOrder{
		Code:              code,
		DeliveryDate:      deliveryDate,
		WarehouseID:       warehouseID,
		SalesOrderID:      req.SalesOrderID,
		DeliveredByID:     req.DeliveredByID,
		CourierAgencyID:   req.CourierAgencyID,
		TrackingNumber:    req.TrackingNumber,
		ReceiverName:      req.ReceiverName,
		ReceiverPhone:     req.ReceiverPhone,
		DeliveryAddress:   req.DeliveryAddress,
		Notes:             req.Notes,
		Status:            salesModels.DeliveryOrderStatusDraft,
		IsPartialDelivery: false, // Will be determined based on quantities
		CreatedBy:         createdBy,
		CreatedAt:         apptime.Now(),
		UpdatedAt:         apptime.Now(),
	}

	// Map items
	if len(req.Items) > 0 {
		deliveryOrder.Items = make([]salesModels.DeliveryOrderItem, len(req.Items))
		for i, itemReq := range req.Items {
			deliveryOrder.Items[i] = salesModels.DeliveryOrderItem{
				WarehouseID:      itemReq.WarehouseID,
				SalesOrderItemID: itemReq.SalesOrderItemID,
				ProductID:        itemReq.ProductID,
				InventoryBatchID: itemReq.InventoryBatchID,
				Quantity:         itemReq.Quantity,
				Price:            itemReq.Price,
				IsEquipment:      itemReq.IsEquipment,
				CreatedAt:        apptime.Now(),
				UpdatedAt:        apptime.Now(),
			}
			deliveryOrder.Items[i].CalculateSubtotal()
		}
	}

	return deliveryOrder, nil
}

// UpdateDeliveryOrderModel updates a DeliveryOrder model from UpdateDeliveryOrderRequest
func UpdateDeliveryOrderModel(m *salesModels.DeliveryOrder, req *dto.UpdateDeliveryOrderRequest) error {
	if req.DeliveryDate != nil {
		deliveryDate, err := time.Parse("2006-01-02", *req.DeliveryDate)
		if err != nil {
			return err
		}
		m.DeliveryDate = deliveryDate
	}

	if req.WarehouseID != nil {
		m.WarehouseID = req.WarehouseID
	}

	if req.DeliveredByID != nil {
		m.DeliveredByID = req.DeliveredByID
	}

	if req.CourierAgencyID != nil {
		m.CourierAgencyID = req.CourierAgencyID
	}

	if req.TrackingNumber != nil {
		m.TrackingNumber = *req.TrackingNumber
	}

	if req.ReceiverName != nil {
		m.ReceiverName = *req.ReceiverName
	}

	if req.ReceiverPhone != nil {
		m.ReceiverPhone = *req.ReceiverPhone
	}

	if req.DeliveryAddress != nil {
		m.DeliveryAddress = *req.DeliveryAddress
	}

	if req.Notes != nil {
		m.Notes = *req.Notes
	}

	// Update items if provided
	if len(req.Items) > 0 {
		m.Items = make([]salesModels.DeliveryOrderItem, len(req.Items))
		for i, itemReq := range req.Items {
			m.Items[i] = salesModels.DeliveryOrderItem{
				WarehouseID:      itemReq.WarehouseID,
				SalesOrderItemID: itemReq.SalesOrderItemID,
				ProductID:        itemReq.ProductID,
				InventoryBatchID: itemReq.InventoryBatchID,
				Quantity:         itemReq.Quantity,
				Price:            itemReq.Price,
				IsEquipment:      itemReq.IsEquipment,
				UpdatedAt:        apptime.Now(),
			}
			m.Items[i].CalculateSubtotal()
		}
	}

	m.UpdatedAt = apptime.Now()
	return nil
}
