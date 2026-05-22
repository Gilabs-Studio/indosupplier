package dto

// CreateSalesQuotationRequest represents the request to create a sales quotation
type CreateSalesQuotationRequest struct {
	QuotationDate   string   `json:"quotation_date" binding:"required"`
	ValidUntil      *string  `json:"valid_until"`
	PaymentTermsID  string   `json:"payment_terms_id" binding:"required,uuid"`
	SalesRepID      string   `json:"sales_rep_id" binding:"required,uuid"`
	BusinessUnitID  string   `json:"business_unit_id" binding:"required,uuid"`
	BusinessTypeID  *string  `json:"business_type_id"`
	CustomerID      *string  `json:"customer_id"`
	CustomerContactID *string `json:"customer_contact_id"`
	CustomerName    string   `json:"customer_name"`
	CustomerContact string   `json:"customer_contact"`
	CustomerPhone   string   `json:"customer_phone"`
	CustomerEmail   string   `json:"customer_email"`
	TaxRate         float64  `json:"tax_rate" binding:"gte=0,lte=100"`
	DeliveryCost    float64  `json:"delivery_cost" binding:"gte=0"`
	OtherCost       float64  `json:"other_cost" binding:"gte=0"`
	DiscountAmount  float64  `json:"discount_amount" binding:"gte=0"`
	Notes           string   `json:"notes"`
	Items           []CreateSalesQuotationItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateSalesQuotationItemRequest represents an item in the quotation
type CreateSalesQuotationItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Discount  float64 `json:"discount" binding:"gte=0"`
}

// UpdateSalesQuotationRequest represents the request to update a sales quotation
type UpdateSalesQuotationRequest struct {
	QuotationDate   *string  `json:"quotation_date"`
	ValidUntil      *string  `json:"valid_until"`
	PaymentTermsID  *string  `json:"payment_terms_id"`
	SalesRepID      *string  `json:"sales_rep_id"`
	BusinessUnitID  *string  `json:"business_unit_id"`
	BusinessTypeID  *string  `json:"business_type_id"`
	CustomerID      *string  `json:"customer_id"`
	CustomerContactID *string `json:"customer_contact_id"`
	CustomerName    *string  `json:"customer_name"`
	CustomerContact *string  `json:"customer_contact"`
	CustomerPhone   *string  `json:"customer_phone"`
	CustomerEmail   *string  `json:"customer_email"`
	TaxRate         *float64 `json:"tax_rate" binding:"omitempty,gte=0,lte=100"`
	DeliveryCost    *float64 `json:"delivery_cost" binding:"omitempty,gte=0"`
	OtherCost       *float64 `json:"other_cost" binding:"omitempty,gte=0"`
	DiscountAmount  *float64 `json:"discount_amount" binding:"omitempty,gte=0"`
	Notes           *string  `json:"notes"`
	Items           *[]CreateSalesQuotationItemRequest `json:"items" binding:"omitempty,dive"`
}

// ListSalesQuotationsRequest represents the request to list sales quotations
type ListSalesQuotationsRequest struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	PerPage        int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search         string `form:"search"`
	Status         string `form:"status"`
	DateFrom       string `form:"date_from"`
	DateTo         string `form:"date_to"`
	SalesRepID     string `form:"sales_rep_id"`
	BusinessUnitID string `form:"business_unit_id"`
	SortBy         string `form:"sort_by"`
	SortDir        string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ListSalesQuotationItemsRequest represents the request to list quotation items with pagination
type ListSalesQuotationItemsRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// UpdateSalesQuotationStatusRequest represents the request to update quotation status
type UpdateSalesQuotationStatusRequest struct {
	Status         string  `json:"status" binding:"required,oneof=sent approved rejected converted"`
	RejectionReason *string `json:"rejection_reason"`
}

// SalesQuotationResponse represents the response for a sales quotation
type SalesQuotationResponse struct {
	ID                  string                        `json:"id"`
	Code                string                        `json:"code"`
	QuotationDate       string                        `json:"quotation_date"`
	ValidUntil          *string                       `json:"valid_until"`
	PaymentTermsID      *string                       `json:"payment_terms_id"`
	PaymentTerms        *PaymentTermsResponse         `json:"payment_terms,omitempty"`
	SalesRepID          *string                       `json:"sales_rep_id"`
	SalesRep            *EmployeeResponse             `json:"sales_rep,omitempty"`
	BusinessUnitID      *string                       `json:"business_unit_id"`
	BusinessUnit        *BusinessUnitResponse         `json:"business_unit,omitempty"`
	BusinessTypeID      *string                       `json:"business_type_id"`
	BusinessType        *BusinessTypeResponse         `json:"business_type,omitempty"`
	CustomerID          *string                       `json:"customer_id"`
	Customer            *CustomerResponse             `json:"customer,omitempty"`
	CustomerContactID   *string                       `json:"customer_contact_id"`
	CustomerContactRef  *CustomerContactResponse      `json:"customer_contact_ref,omitempty"`
	CustomerName        string                        `json:"customer_name"`
	CustomerContact     string                        `json:"customer_contact"`
	CustomerPhone       string                        `json:"customer_phone"`
	CustomerEmail       string                        `json:"customer_email"`
	Subtotal            float64                       `json:"subtotal"`
	DiscountAmount      float64                       `json:"discount_amount"`
	TaxRate             float64                       `json:"tax_rate"`
	TaxAmount           float64                       `json:"tax_amount"`
	DeliveryCost        float64                       `json:"delivery_cost"`
	OtherCost           float64                       `json:"other_cost"`
	TotalAmount         float64                       `json:"total_amount"`
	Status              string                        `json:"status"`
	Notes               string                        `json:"notes"`
	CreatedBy           *string                       `json:"created_by"`
	ApprovedBy          *string                       `json:"approved_by"`
	ApprovedAt          *string                       `json:"approved_at"`
	RejectedBy          *string                       `json:"rejected_by"`
	RejectedAt          *string                       `json:"rejected_at"`
	RejectionReason     *string                       `json:"rejection_reason"`
	SourceDealID        *string                       `json:"source_deal_id"`
	ConvertedToSalesOrderID *string                   `json:"converted_to_sales_order_id"`
	ConvertedAt         *string                       `json:"converted_at"`
	Items               []SalesQuotationItemResponse  `json:"items,omitempty"`
	CreatedAt           string                        `json:"created_at"`
	UpdatedAt           string                        `json:"updated_at"`
}

// SalesQuotationItemResponse represents an item in the quotation response
type SalesQuotationItemResponse struct {
	ID               string            `json:"id"`
	SalesQuotationID string            `json:"sales_quotation_id"`
	ProductID        string            `json:"product_id"`
	Product          *ProductResponse  `json:"product,omitempty"`
	Quantity         float64           `json:"quantity"`
	Price            float64           `json:"price"`
	Discount         float64           `json:"discount"`
	Subtotal         float64           `json:"subtotal"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

// PaymentTermsResponse represents payment terms in response
type PaymentTermsResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Days        int    `json:"days"`
}

// EmployeeResponse represents employee in response
type EmployeeResponse struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// BusinessUnitResponse represents business unit in response
type BusinessUnitResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// BusinessTypeResponse represents business type in response
type BusinessTypeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProductResponse represents product in response
type ProductResponse struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	SellingPrice float64 `json:"selling_price"`
	ImageURL     *string `json:"image_url"`
}

// CustomerResponse represents customer master data in response
type CustomerResponse struct {
	ID             string  `json:"id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	CustomerTypeID *string `json:"customer_type_id"`
	Address        string  `json:"address"`
	Email          string  `json:"email"`
	ContactPerson  string  `json:"contact_person"`
}

// CustomerContactResponse represents CRM contact in sales quotation response
type CustomerContactResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}
