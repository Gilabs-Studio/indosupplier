package dto

// CreateCustomerInvoiceRequest represents the request to create a customer invoice
type CreateCustomerInvoiceRequest struct {
	InvoiceDate          string                             `json:"invoice_date" binding:"required"`
	DueDate              *string                            `json:"due_date" binding:"required"`
	Type                 string                             `json:"type" binding:"omitempty,oneof=regular proforma down_payment"`
	SalesOrderID         *string                            `json:"sales_order_id" binding:"omitempty,uuid"`
	DeliveryOrderID      *string                            `json:"delivery_order_id" binding:"omitempty,uuid"`
	PaymentTermsID       *string                            `json:"payment_terms_id" binding:"omitempty,uuid"`
	TaxRate              float64                            `json:"tax_rate" binding:"gte=0,lte=100"`
	DeliveryCost         float64                            `json:"delivery_cost" binding:"gte=0"`
	OtherCost            float64                            `json:"other_cost" binding:"gte=0"`
	DownPaymentInvoiceID *string                            `json:"down_payment_invoice_id,omitempty"`
	Notes                string                             `json:"notes"`
	Items                []CreateCustomerInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CustomerInvoiceJournalPreviewRequest reuses the create payload to simulate posting before submit.
type CustomerInvoiceJournalPreviewRequest = CreateCustomerInvoiceRequest

// CustomerInvoiceJournalPreviewLine represents a simulated journal line.
type CustomerInvoiceJournalPreviewLine struct {
	ChartOfAccountID   string  `json:"chart_of_account_id"`
	ChartOfAccountCode string  `json:"chart_of_account_code,omitempty"`
	ChartOfAccountName string  `json:"chart_of_account_name,omitempty"`
	Debit              float64 `json:"debit"`
	Credit             float64 `json:"credit"`
	Memo               string  `json:"memo"`
}

// CustomerInvoiceJournalPreviewResponse represents the accounting preview for a customer invoice.
type CustomerInvoiceJournalPreviewResponse struct {
	ReferenceType string                           `json:"reference_type"`
	ReferenceID   string                           `json:"reference_id"`
	InvoiceDate   string                           `json:"invoice_date"`
	InvoiceNumber string                           `json:"invoice_number,omitempty"`
	Subtotal      float64                          `json:"subtotal"`
	TaxAmount     float64                          `json:"tax_amount"`
	DownPayment   float64                          `json:"down_payment"`
	TotalAmount   float64                          `json:"total_amount"`
	IsBalanced    bool                             `json:"is_balanced"`
	Lines         []CustomerInvoiceJournalPreviewLine `json:"lines"`
}

// CreateCustomerInvoiceItemRequest represents an item in the invoice
type CreateCustomerInvoiceItemRequest struct {
	ProductID           string  `json:"product_id" binding:"required"`
	SalesOrderItemID    *string `json:"sales_order_item_id" binding:"omitempty,uuid"`
	DeliveryOrderItemID *string `json:"delivery_order_item_id" binding:"omitempty,uuid"`
	Quantity            float64 `json:"quantity" binding:"required,gt=0"`
	Price               float64 `json:"price" binding:"required,gt=0"`
	Discount            float64 `json:"discount" binding:"gte=0"`
	HPPAmount           float64 `json:"hpp_amount" binding:"gte=0"`
}

// UpdateCustomerInvoiceRequest represents the request to update a customer invoice
type UpdateCustomerInvoiceRequest struct {
	InvoiceDate    *string                             `json:"invoice_date"`
	DueDate        *string                             `json:"due_date"`
	Type           *string                             `json:"type" binding:"omitempty,oneof=regular proforma"`
	PaymentTermsID *string                             `json:"payment_terms_id"`
	TaxRate        *float64                            `json:"tax_rate" binding:"omitempty,gte=0,lte=100"`
	DeliveryCost   *float64                            `json:"delivery_cost" binding:"omitempty,gte=0"`
	OtherCost      *float64                            `json:"other_cost" binding:"omitempty,gte=0"`
	Notes          *string                             `json:"notes"`
	Items          *[]CreateCustomerInvoiceItemRequest `json:"items" binding:"omitempty,dive"`
}

// ListCustomerInvoicesRequest represents the request to list customer invoices
type ListCustomerInvoicesRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PerPage      int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search       string `form:"search"`
	Status       string `form:"status" binding:"omitempty,oneof=draft submitted approved rejected unpaid waiting_approval waiting_payment partial paid cancelled DRAFT SUBMITTED APPROVED REJECTED UNPAID WAITING_APPROVAL WAITING_PAYMENT PARTIAL PAID CANCELLED"`
	Type         string `form:"type" binding:"omitempty,oneof=regular proforma down_payment"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
	DueDateFrom  string `form:"due_date_from"`
	DueDateTo    string `form:"due_date_to"`
	SalesOrderID string `form:"sales_order_id"`
	SortBy       string `form:"sort_by"`
	SortDir      string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ListCustomerInvoiceItemsRequest represents the request to list invoice items with pagination
type ListCustomerInvoiceItemsRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// UpdateCustomerInvoiceStatusRequest represents the request to update invoice status
type UpdateCustomerInvoiceStatusRequest struct {
	Status     string   `json:"status" binding:"required,oneof=submitted approved rejected unpaid waiting_approval waiting_payment partial paid cancelled SUBMITTED APPROVED REJECTED UNPAID WAITING_APPROVAL WAITING_PAYMENT PARTIAL PAID CANCELLED"`
	PaidAmount *float64 `json:"paid_amount" binding:"omitempty,gte=0"`
	PaymentAt  *string  `json:"payment_at"`
}

// CustomerInvoiceResponse represents the response for a customer invoice
type CustomerInvoiceResponse struct {
	ID                     string                        `json:"id"`
	Code                   string                        `json:"code"`
	InvoiceNumber          *string                       `json:"invoice_number"`
	Type                   string                        `json:"type"`
	InvoiceDate            string                        `json:"invoice_date"`
	DueDate                *string                       `json:"due_date"`
	SalesOrderID           *string                       `json:"sales_order_id"`
	SalesOrder             *SalesOrderBriefResponse      `json:"sales_order,omitempty"`
	DeliveryOrderID        *string                       `json:"delivery_order_id"`
	DeliveryOrder          *DeliveryOrderBriefResponse   `json:"delivery_order,omitempty"`
	PaymentTermsID         *string                       `json:"payment_terms_id"`
	PaymentTerms           *PaymentTermsResponse         `json:"payment_terms,omitempty"`
	DownPaymentInvoiceID   *string                       `json:"down_payment_invoice_id,omitempty"`
	DownPaymentInvoiceCode *string                       `json:"down_payment_invoice_code,omitempty"`
	Subtotal               float64                       `json:"subtotal"`
	TaxRate                float64                       `json:"tax_rate"`
	TaxAmount              float64                       `json:"tax_amount"`
	DeliveryCost           float64                       `json:"delivery_cost"`
	OtherCost              float64                       `json:"other_cost"`
	DownPaymentAmount      float64                       `json:"down_payment_amount"`
	Amount                 float64                       `json:"amount"`
	PaidAmount             float64                       `json:"paid_amount"`
	RemainingAmount        float64                       `json:"remaining_amount"`
	Status                 string                        `json:"status"`
	Notes                  string                        `json:"notes"`
	IsPosted               bool                          `json:"is_posted"`
	JournalEntryID         *string                       `json:"journal_entry_id,omitempty"`
	PaymentAt              *string                       `json:"payment_at"`
	CreatedBy              *string                       `json:"created_by"`
	CancelledBy            *string                       `json:"cancelled_by"`
	CancelledAt            *string                       `json:"cancelled_at"`
	Items                  []CustomerInvoiceItemResponse `json:"items,omitempty"`
	CreatedAt              string                        `json:"created_at"`
	UpdatedAt              string                        `json:"updated_at"`
}

// DeliveryOrderBriefResponse represents a brief delivery order info in response
type DeliveryOrderBriefResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

// CustomerInvoiceItemResponse represents an item in the invoice response
type CustomerInvoiceItemResponse struct {
	ID                  string           `json:"id"`
	CustomerInvoiceID   string           `json:"customer_invoice_id"`
	ProductID           string           `json:"product_id"`
	Product             *ProductResponse `json:"product,omitempty"`
	SalesOrderItemID    *string          `json:"sales_order_item_id"`
	DeliveryOrderItemID *string          `json:"delivery_order_item_id"`
	Quantity            float64          `json:"quantity"`
	Price               float64          `json:"price"`
	Discount            float64          `json:"discount"`
	Subtotal            float64          `json:"subtotal"`
	HPPAmount           float64          `json:"hpp_amount"`
	CreatedAt           string           `json:"created_at"`
	UpdatedAt           string           `json:"updated_at"`
}

// SalesOrderBriefResponse represents a brief sales order info in response
type SalesOrderBriefResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	// Snapshot customer fields copied from sales order
	CustomerID    *string           `json:"customer_id,omitempty"`
	Customer      *CustomerResponse `json:"customer,omitempty"`
	CustomerName  string            `json:"customer_name,omitempty"`
	CustomerPhone string            `json:"customer_phone,omitempty"`
	CustomerEmail string            `json:"customer_email,omitempty"`
}
