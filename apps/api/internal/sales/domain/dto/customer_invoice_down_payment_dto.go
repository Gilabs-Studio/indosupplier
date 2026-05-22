package dto

import "time"

// CreateCustomerInvoiceDownPaymentRequest represents the request to create a down payment
type CreateCustomerInvoiceDownPaymentRequest struct {
	SalesOrderID  string  `json:"sales_order_id" binding:"required,uuid"`
	InvoiceDate   string  `json:"invoice_date" binding:"required"`
	DueDate       string  `json:"due_date" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	AttachmentURL *string `json:"attachment_url"`
	Notes         *string `json:"notes"`
}

// UpdateCustomerInvoiceDownPaymentRequest represents the request to update a down payment
type UpdateCustomerInvoiceDownPaymentRequest struct {
	SalesOrderID  string  `json:"sales_order_id" binding:"required,uuid"`
	InvoiceDate   string  `json:"invoice_date" binding:"required"`
	DueDate       string  `json:"due_date" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	AttachmentURL *string `json:"attachment_url"`
	Notes         *string `json:"notes"`
}

// CustomerInvoiceDownPaymentDetailResponse represents a detailed down payment response
type CustomerInvoiceDownPaymentDetailResponse struct {
	ID                 string                                `json:"id"`
	SalesOrderID       string                                `json:"sales_order_id"`
	SalesOrder         *CustomerInvoiceDownPaymentSalesOrder `json:"sales_order,omitempty"`
	CustomerID         string                                `json:"customer_id"`
	Code               string                                `json:"code"`
	RelatedInvoiceCode *string                               `json:"related_invoice_code,omitempty"`
	InvoiceNumber      *string                               `json:"invoice_number"`
	InvoiceDate        string                                `json:"invoice_date"`
	DueDate            *string                               `json:"due_date"`
	Amount             float64                               `json:"amount"`
	RemainingAmount    float64                               `json:"remaining_amount"`
	Status             string                                `json:"status"`
	AttachmentURL      *string                               `json:"attachment_url,omitempty"`
	Notes              *string                               `json:"notes"`
	CreatedBy          *string                               `json:"created_by"`
	CreatedAt          time.Time                             `json:"created_at"`
	UpdatedAt          time.Time                             `json:"updated_at"`
}

// CustomerInvoiceDownPaymentListResponse represents a summary down payment response
type CustomerInvoiceDownPaymentListResponse struct {
	ID                 string                                `json:"id"`
	SalesOrderID       string                                `json:"sales_order_id"`
	SalesOrder         *CustomerInvoiceDownPaymentSalesOrder `json:"sales_order,omitempty"`
	Code               string                                `json:"code"`
	RelatedInvoiceCode *string                               `json:"related_invoice_code,omitempty"`
	InvoiceNumber      *string                               `json:"invoice_number"`
	InvoiceDate        string                                `json:"invoice_date"`
	DueDate            *string                               `json:"due_date"`
	Amount             float64                               `json:"amount"`
	RemainingAmount    float64                               `json:"remaining_amount"`
	Status             string                                `json:"status"`
	AttachmentURL      *string                               `json:"attachment_url,omitempty"`
	CreatedAt          time.Time                             `json:"created_at"`
}

type CustomerInvoiceDownPaymentSalesOrder struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	CustomerID   *string `json:"customer_id,omitempty"`
	CustomerName *string `json:"customer_name,omitempty"`
}

// CustomerInvoiceDownPaymentAddResponse is used for the fetch dropdowns on create screen
type CustomerInvoiceDownPaymentAddResponse struct {
	SalesOrders []CustomerInvoiceAddSalesOrder `json:"sales_orders"`
}

type CustomerInvoiceAddSalesOrder struct {
	ID          string                             `json:"id"`
	Customer    *CustomerInvoiceAddCustomerMini    `json:"customer,omitempty"`
	Code        string                             `json:"code"`
	OrderDate   time.Time                          `json:"order_date"`
	Status      string                             `json:"status"`
	TotalAmount float64                            `json:"total_amount"`
	Items       []CustomerInvoiceAddSalesOrderItem `json:"items"`
}

type CustomerInvoiceAddCustomerMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomerInvoiceAddSalesOrderItem struct {
	ID       string                         `json:"id"`
	Product  *CustomerInvoiceAddProductMini `json:"product,omitempty"`
	Quantity float64                        `json:"quantity"`
	Price    float64                        `json:"price"`
	Subtotal float64                        `json:"subtotal"`
}

type CustomerInvoiceAddProductMini struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	ImageURL string `json:"image_url"`
}

// CustomerInvoiceAuditTrailEntry is similarly re-used for Down Payment Audit Trail
type CustomerInvoiceAuditTrailEntry struct {
	ID             string                 `json:"id"`
	Action         string                 `json:"action"`
	PermissionCode string                 `json:"permission_code"`
	TargetID       string                 `json:"target_id"`
	Metadata       map[string]interface{} `json:"metadata"`
	User           *AuditTrailUser        `json:"user"`
	CreatedAt      time.Time              `json:"created_at"`
}

type AuditTrailUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
