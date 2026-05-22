package dto

import "time"

type CreateDeliveryOrderRequest struct {
	DeliveryDate    string                           `json:"delivery_date" binding:"required"`
	WarehouseID     string                           `json:"warehouse_id" binding:"omitempty,uuid"`
	SalesOrderID    string                           `json:"sales_order_id" binding:"required,uuid"`
	DeliveredByID   *string                          `json:"delivered_by_id"`
	CourierAgencyID *string                          `json:"courier_agency_id"`
	TrackingNumber  string                           `json:"tracking_number"`
	ReceiverName    string                           `json:"receiver_name"`
	ReceiverPhone   string                           `json:"receiver_phone"`
	DeliveryAddress string                           `json:"delivery_address"`
	Notes           string                           `json:"notes"`
	Items           []CreateDeliveryOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateDeliveryOrderItemRequest represents an item in the delivery order
type CreateDeliveryOrderItemRequest struct {
	SalesOrderItemID *string `json:"sales_order_item_id"`
	ProductID        string  `json:"product_id" binding:"required,uuid"`
	WarehouseID      *string `json:"warehouse_id" binding:"omitempty,uuid"`
	InventoryBatchID *string `json:"inventory_batch_id" binding:"omitempty,uuid"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	Price            float64 `json:"price" binding:"omitempty,gte=0"`
	IsEquipment      bool    `json:"is_equipment"`
}

// UpdateDeliveryOrderItemRequest represents an item in the update delivery order
type UpdateDeliveryOrderItemRequest struct {
	SalesOrderItemID *string `json:"sales_order_item_id"`
	ProductID        string  `json:"product_id" binding:"required,uuid"`
	WarehouseID      *string `json:"warehouse_id" binding:"omitempty,uuid"`
	InventoryBatchID *string `json:"inventory_batch_id" binding:"omitempty,uuid"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	Price            float64 `json:"price" binding:"omitempty,gte=0"`
	IsEquipment      bool    `json:"is_equipment"`
}

// UpdateDeliveryOrderRequest represents the request to update a delivery order
type UpdateDeliveryOrderRequest struct {
	DeliveryDate    *string                          `json:"delivery_date"`
	WarehouseID     *string                          `json:"warehouse_id"`
	DeliveredByID   *string                          `json:"delivered_by_id"`
	CourierAgencyID *string                          `json:"courier_agency_id"`
	TrackingNumber  *string                          `json:"tracking_number"`
	ReceiverName    *string                          `json:"receiver_name"`
	ReceiverPhone   *string                          `json:"receiver_phone"`
	DeliveryAddress *string                          `json:"delivery_address"`
	Notes           *string                          `json:"notes"`
	Items           []UpdateDeliveryOrderItemRequest `json:"items" binding:"omitempty,min=1,dive"`
}

// ListDeliveryOrdersRequest represents the request to list delivery orders
type ListDeliveryOrdersRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PerPage      int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search       string `form:"search"`
	Status       string `form:"status"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
	SalesOrderID string `form:"sales_order_id"`
	SortBy       string `form:"sort_by"`
	SortDir      string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ListDeliveryOrderItemsRequest represents the request to list delivery order items with pagination
type ListDeliveryOrderItemsRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// UpdateDeliveryOrderStatusRequest represents the request to update delivery order status
type UpdateDeliveryOrderStatusRequest struct {
	Status             string  `json:"status" binding:"required,oneof=sent approved rejected prepared shipped delivered cancelled"`
	CancellationReason *string `json:"cancellation_reason"`
}

// ShipDeliveryOrderRequest represents the request to ship a delivery order
type ShipDeliveryOrderRequest struct {
	TrackingNumber string `json:"tracking_number" binding:"required"`
}

// DeliverDeliveryOrderRequest represents the request to mark delivery order as delivered
type DeliverDeliveryOrderRequest struct {
	ReceiverSignature string `json:"receiver_signature" binding:"required"`
	ReceiverName      string `json:"receiver_name" binding:"required"`
}

// BatchSelectionRequest represents the request for batch selection (FIFO/FEFO)
type BatchSelectionRequest struct {
	ProductID   string  `json:"product_id" binding:"required,uuid"`
	WarehouseID *string `json:"warehouse_id"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Method      string  `json:"method" binding:"required,oneof=FIFO FEFO"` // FIFO or FEFO
}

// BatchSelectionResponse represents available batches for selection
type BatchSelectionResponse struct {
	Batches        []BatchInfo `json:"batches"`
	TotalAvailable float64     `json:"total_available"`
}

// BatchInfo represents information about an inventory batch
type BatchInfo struct {
	ID           string    `json:"id"`
	BatchNumber  string    `json:"batch_number"`
	ExpiryDate   time.Time `json:"expiry_date"`
	Quantity     float64   `json:"quantity"`
	Available    float64   `json:"available"`
	ReceivedDate time.Time `json:"received_date"`
}

// DeliveryOrderResponse represents the response for a delivery order
type DeliveryOrderResponse struct {
	ID                 string                      `json:"id"`
	Code               string                      `json:"code"`
	DeliveryDate       string                      `json:"delivery_date"`
	WarehouseID        string                      `json:"warehouse_id"`
	Warehouse          *WarehouseResponse          `json:"warehouse,omitempty"`
	SalesOrderID       string                      `json:"sales_order_id"`
	SalesOrder         *SalesOrderResponse         `json:"sales_order,omitempty"`
	DeliveredByID      *string                     `json:"delivered_by_id"`
	DeliveredBy        *EmployeeResponse           `json:"delivered_by,omitempty"`
	CourierAgencyID    *string                     `json:"courier_agency_id"`
	CourierAgency      *CourierAgencyResponse      `json:"courier_agency,omitempty"`
	TrackingNumber     string                      `json:"tracking_number"`
	ReceiverName       string                      `json:"receiver_name"`
	ReceiverPhone      string                      `json:"receiver_phone"`
	DeliveryAddress    string                      `json:"delivery_address"`
	ReceiverSignature  string                      `json:"receiver_signature"`
	Status             string                      `json:"status"`
	Notes              string                      `json:"notes"`
	IsPosted           bool                        `json:"is_posted"`
	IsPartialDelivery  bool                        `json:"is_partial_delivery"`
	CreatedBy          *string                     `json:"created_by"`
	ShippedBy          *string                     `json:"shipped_by"`
	ShippedAt          *string                     `json:"shipped_at"`
	DeliveredAt        *string                     `json:"delivered_at"`
	CancelledBy        *string                     `json:"cancelled_by"`
	CancelledAt        *string                     `json:"cancelled_at"`
	CancellationReason *string                     `json:"cancellation_reason"`
	JournalEntryID     *string                     `json:"journal_entry_id,omitempty"`
	Items              []DeliveryOrderItemResponse `json:"items,omitempty"`
	CreatedAt          string                      `json:"created_at"`
	UpdatedAt          string                      `json:"updated_at"`
}

// DeliveryOrderItemResponse represents an item in the delivery order response
type DeliveryOrderItemResponse struct {
	ID                 string                  `json:"id"`
	DeliveryOrderID    string                  `json:"delivery_order_id"`
	WarehouseID       *string                 `json:"warehouse_id"`
	Warehouse         *WarehouseResponse     `json:"warehouse,omitempty"`
	SalesOrderItemID   *string                 `json:"sales_order_item_id"`
	SalesOrderItem     *SalesOrderItemResponse `json:"sales_order_item,omitempty"`
	ProductID          string                  `json:"product_id"`
	Product            *ProductResponse        `json:"product,omitempty"`
	InventoryBatchID   *string                 `json:"inventory_batch_id"`
	InventoryBatch     *BatchInfo              `json:"inventory_batch,omitempty"`
	Quantity           float64                 `json:"quantity"`
	Price              float64                 `json:"price"`
	Subtotal           float64                 `json:"subtotal"`
	AvgCostSnapshot    float64                 `json:"avg_cost_snapshot"`
	COGSAmount         float64                 `json:"cogs_amount"`
	IsEquipment        bool                    `json:"is_equipment"`
	InstallationStatus *string                 `json:"installation_status"`
	FunctionTestStatus *string                 `json:"function_test_status"`
	InstallationDate   *string                 `json:"installation_date"`
	FunctionTestDate   *string                 `json:"function_test_date"`
	InstallationNotes  string                  `json:"installation_notes"`
	CreatedAt          string                  `json:"created_at"`
	UpdatedAt          string                  `json:"updated_at"`
}

// WarehouseResponse represents warehouse data in response
type WarehouseResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// CourierAgencyResponse represents courier agency in response
type CourierAgencyResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	TrackingURL string `json:"tracking_url"`
}
