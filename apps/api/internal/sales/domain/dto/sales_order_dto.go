package dto

// CreateSalesOrderRequest represents the request to create a sales order
type CreateSalesOrderRequest struct {
	OrderDate         string                        `json:"order_date" binding:"required"`
	SalesQuotationID  *string                       `json:"sales_quotation_id"`
	PaymentTermsID    *string                       `json:"payment_terms_id" binding:"required"`
	SalesRepID        *string                       `json:"sales_rep_id" binding:"required"`
	BusinessUnitID    *string                       `json:"business_unit_id" binding:"required"`
	BusinessTypeID    *string                       `json:"business_type_id"`
	DeliveryAreaID    *string                       `json:"delivery_area_id"`
	CustomerID        *string                       `json:"customer_id"`
	CustomerContactID *string                       `json:"customer_contact_id"`
	CustomerName      string                        `json:"customer_name"`
	CustomerContact   string                        `json:"customer_contact"`
	CustomerPhone     string                        `json:"customer_phone"`
	CustomerEmail     string                        `json:"customer_email"`
	TaxRate           float64                       `json:"tax_rate" binding:"gte=0,lte=100"`
	DeliveryCost      float64                       `json:"delivery_cost" binding:"gte=0"`
	OtherCost         float64                       `json:"other_cost" binding:"gte=0"`
	DiscountAmount    float64                       `json:"discount_amount" binding:"gte=0"`
	Notes             string                        `json:"notes"`
	Items             []CreateSalesOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateSalesOrderItemRequest represents an item in the order
type CreateSalesOrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Discount  float64 `json:"discount" binding:"gte=0"`
}

// UpdateSalesOrderRequest represents the request to update a sales order
type UpdateSalesOrderRequest struct {
	OrderDate         *string                       `json:"order_date"`
	PaymentTermsID    *string                       `json:"payment_terms_id"`
	SalesRepID        *string                       `json:"sales_rep_id"`
	BusinessUnitID    *string                       `json:"business_unit_id"`
	BusinessTypeID    *string                       `json:"business_type_id"`
	DeliveryAreaID    *string                       `json:"delivery_area_id"`
	CustomerID        *string                       `json:"customer_id"`
	CustomerContactID *string                       `json:"customer_contact_id"`
	CustomerName      *string                       `json:"customer_name"`
	CustomerContact   *string                       `json:"customer_contact"`
	CustomerPhone     *string                       `json:"customer_phone"`
	CustomerEmail     *string                       `json:"customer_email"`
	TaxRate           *float64                      `json:"tax_rate" binding:"omitempty,gte=0,lte=100"`
	DeliveryCost      *float64                      `json:"delivery_cost" binding:"omitempty,gte=0"`
	OtherCost         *float64                      `json:"other_cost" binding:"omitempty,gte=0"`
	DiscountAmount    *float64                      `json:"discount_amount" binding:"omitempty,gte=0"`
	Notes             *string                       `json:"notes"`
	Items             []CreateSalesOrderItemRequest `json:"items" binding:"omitempty,min=1,dive"`
}

// ListSalesOrdersRequest represents the request to list sales orders
type ListSalesOrdersRequest struct {
	Page                  int    `form:"page" binding:"omitempty,min=1"`
	PerPage               int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search                string `form:"search"`
	Status                string `form:"status"`
	SourceType            string `form:"source_type"`
	DateFrom              string `form:"date_from"`
	DateTo                string `form:"date_to"`
	SalesRepID            string `form:"sales_rep_id"`
	BusinessUnitID        string `form:"business_unit_id"`
	SalesQuotationID      string `form:"sales_quotation_id"`
	CustomerID            string `form:"customer_id"`
	UnfulfilledOnly       bool   `form:"unfulfilled_only"`
	AvailableForInvoice   bool   `form:"available_for_invoice"`
	ExcludeWithActiveCIDP bool   `form:"exclude_with_active_cidp"`
	ExcludeWithPaidCI     bool   `form:"exclude_with_paid_ci"`
	SortBy                string `form:"sort_by"`
	SortDir               string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ListSalesOrderItemsRequest represents the request to list order items with pagination
type ListSalesOrderItemsRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// UpdateSalesOrderStatusRequest represents the request to update order status
type UpdateSalesOrderStatusRequest struct {
	Status             string  `json:"status" binding:"required,oneof=draft submitted approved rejected cancelled"`
	CancellationReason *string `json:"cancellation_reason"`
}

// ConvertFromQuotationRequest represents the request to convert quotation to order
type ConvertFromQuotationRequest struct {
	QuotationID       string  `json:"quotation_id" binding:"required,uuid"`
	DeliveryAreaID    *string `json:"delivery_area_id"`
	CustomerID        *string `json:"customer_id"`
	CustomerContactID *string `json:"customer_contact_id"`
	CustomerName      string  `json:"customer_name"`
	CustomerContact   string  `json:"customer_contact"`
	CustomerPhone     string  `json:"customer_phone"`
	CustomerEmail     string  `json:"customer_email"`
	Notes             string  `json:"notes"`
}

// FulfillmentSummary represents the delivery fulfillment progress of a sales order
type FulfillmentSummary struct {
	TotalOrdered   float64 `json:"total_ordered"`
	TotalDelivered float64 `json:"total_delivered"`
	TotalPending   float64 `json:"total_pending"`
	TotalRemaining float64 `json:"total_remaining"`
}

// SalesOrderResponse represents the response for a sales order
type SalesOrderResponse struct {
	ID                 string                   `json:"id"`
	Code               string                   `json:"code"`
	OrderDate          string                   `json:"order_date"`
	SalesQuotationID   *string                  `json:"sales_quotation_id"`
	SalesQuotation     *SalesQuotationResponse  `json:"sales_quotation,omitempty"`
	PaymentTermsID     *string                  `json:"payment_terms_id"`
	PaymentTerms       *PaymentTermsResponse    `json:"payment_terms,omitempty"`
	SalesRepID         *string                  `json:"sales_rep_id"`
	SalesRep           *EmployeeResponse        `json:"sales_rep,omitempty"`
	BusinessUnitID     *string                  `json:"business_unit_id"`
	BusinessUnit       *BusinessUnitResponse    `json:"business_unit,omitempty"`
	BusinessTypeID     *string                  `json:"business_type_id"`
	BusinessType       *BusinessTypeResponse    `json:"business_type,omitempty"`
	DeliveryAreaID     *string                  `json:"delivery_area_id"`
	DeliveryArea       *AreaResponse            `json:"delivery_area,omitempty"`
	CustomerID         *string                  `json:"customer_id"`
	CustomerContactID  *string                  `json:"customer_contact_id"`
	CustomerContactRef *CustomerContactResponse `json:"customer_contact_ref,omitempty"`
	Customer           *CustomerResponse        `json:"customer,omitempty"`
	CustomerName       string                   `json:"customer_name"`
	CustomerContact    string                   `json:"customer_contact"`
	CustomerPhone      string                   `json:"customer_phone"`
	CustomerEmail      string                   `json:"customer_email"`
	Subtotal           float64                  `json:"subtotal"`
	DiscountAmount     float64                  `json:"discount_amount"`
	TaxRate            float64                  `json:"tax_rate"`
	TaxAmount          float64                  `json:"tax_amount"`
	DeliveryCost       float64                  `json:"delivery_cost"`
	OtherCost          float64                  `json:"other_cost"`
	TotalAmount        float64                  `json:"total_amount"`
	ReservedStock      bool                     `json:"reserved_stock"`
	Status             string                   `json:"status"`
	Notes              string                   `json:"notes"`
	SourceType         string                   `json:"source_type"`
	SourcePOSOrderID   *string                  `json:"source_pos_order_id"`
	Fulfillment        *FulfillmentSummary      `json:"fulfillment,omitempty"`
	CreatedBy          *string                  `json:"created_by"`
	ConfirmedBy        *string                  `json:"confirmed_by"`
	ConfirmedAt        *string                  `json:"confirmed_at"`
	CancelledBy        *string                  `json:"cancelled_by"`
	CancelledAt        *string                  `json:"cancelled_at"`
	CancellationReason *string                  `json:"cancellation_reason"`
	Items              []SalesOrderItemResponse `json:"items,omitempty"`
	DeliveryOrders     []DeliveryOrderSummary   `json:"delivery_orders,omitempty"`
	CustomerInvoices   []CustomerInvoiceSummary `json:"customer_invoices,omitempty"`
	CreatedAt          string                   `json:"created_at"`
	UpdatedAt          string                   `json:"updated_at"`
}

// DeliveryOrderSummary represents a minimal delivery order for SO status display
type DeliveryOrderSummary struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	Status            string `json:"status"`
	DeliveryDate      string `json:"delivery_date"`
	IsPartialDelivery bool   `json:"is_partial_delivery"`
}

// CustomerInvoiceSummary represents a minimal customer invoice for SO status display
type CustomerInvoiceSummary struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Status      string  `json:"status"`
	InvoiceDate string  `json:"invoice_date"`
	DueDate     string  `json:"due_date"`
	Amount      float64 `json:"amount"`
	PaidAmount  float64 `json:"paid_amount"`
}

// SalesOrderItemResponse represents an item in the order response
type SalesOrderItemResponse struct {
	ID                      string           `json:"id"`
	SalesOrderID            string           `json:"sales_order_id"`
	ProductID               string           `json:"product_id"`
	Product                 *ProductResponse `json:"product,omitempty"`
	Quantity                float64          `json:"quantity"`
	Price                   float64          `json:"price"`
	Discount                float64          `json:"discount"`
	Subtotal                float64          `json:"subtotal"`
	ReservedQuantity        float64          `json:"reserved_quantity"`
	DeliveredQuantity       float64          `json:"delivered_quantity"`
	InvoicedQuantity        float64          `json:"invoiced_quantity"`
	PendingDeliveryQuantity float64          `json:"pending_delivery_quantity"`
	CreatedAt               string           `json:"created_at"`
	UpdatedAt               string           `json:"updated_at"`
}

// AreaResponse represents area in response
type AreaResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
