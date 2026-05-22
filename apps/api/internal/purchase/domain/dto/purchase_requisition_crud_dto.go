package dto

type PurchaseRequisitionItemRequest struct {
	ProductID     string  `json:"product_id" validate:"required,uuid"`
	Quantity      float64 `json:"quantity" validate:"required,gt=0"`
	PurchasePrice float64 `json:"purchase_price" validate:"required,gte=0"`
	Discount      float64 `json:"discount" validate:"omitempty,gte=0,lte=100"`
	Notes         *string `json:"notes"`
}

type CreatePurchaseRequisitionRequest struct {
	SupplierID     *string `json:"supplier_id" validate:"omitempty,uuid"`
	PaymentTermsID *string `json:"payment_terms_id" validate:"required,uuid"`
	BusinessUnitID *string `json:"business_unit_id" validate:"omitempty,uuid"`
	EmployeeID     *string `json:"employee_id" validate:"omitempty,uuid"`

	RequestDate string  `json:"request_date" validate:"required,datetime=2006-01-02"`
	Address     *string `json:"address"`
	Notes       string  `json:"notes"`

	TaxRate      float64 `json:"tax_rate" validate:"omitempty,gte=0,lte=100"`
	DeliveryCost float64 `json:"delivery_cost" validate:"omitempty,gte=0"`
	OtherCost    float64 `json:"other_cost" validate:"omitempty,gte=0"`

	Items []PurchaseRequisitionItemRequest `json:"items" validate:"required,dive"`
}

type UpdatePurchaseRequisitionRequest struct {
	SupplierID     *string `json:"supplier_id" validate:"omitempty,uuid"`
	PaymentTermsID *string `json:"payment_terms_id" validate:"required,uuid"`
	BusinessUnitID *string `json:"business_unit_id" validate:"omitempty,uuid"`
	EmployeeID     *string `json:"employee_id" validate:"omitempty,uuid"`

	RequestDate string  `json:"request_date" validate:"required,datetime=2006-01-02"`
	Address     *string `json:"address"`
	Notes       string  `json:"notes"`

	TaxRate      float64 `json:"tax_rate" validate:"omitempty,gte=0,lte=100"`
	DeliveryCost float64 `json:"delivery_cost" validate:"omitempty,gte=0"`
	OtherCost    float64 `json:"other_cost" validate:"omitempty,gte=0"`

	Items []PurchaseRequisitionItemRequest `json:"items" validate:"required,dive"`
}

type PurchaseRequisitionItemResponse struct {
	ID            string  `json:"id"`
	ProductID     string  `json:"product_id"`
	Quantity      float64 `json:"quantity"`
	PurchasePrice float64 `json:"purchase_price"`
	Discount      float64 `json:"discount"`
	Subtotal      float64 `json:"subtotal"`
	Notes         *string `json:"notes"`

	Product *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"product,omitempty"`
}

type PurchaseRequisitionDetailResponse struct {
	ID             string  `json:"id"`
	Code           string  `json:"code"`
	SupplierID     *string `json:"supplier_id"`
	PaymentTermsID *string `json:"payment_terms_id"`
	BusinessUnitID *string `json:"business_unit_id"`
	CompanyID      *string `json:"company_id"`
	FiscalYearID   *string `json:"fiscal_year_id"`
	EmployeeID     *string `json:"employee_id"`
	RequestDate    string  `json:"request_date"`
	Address        *string `json:"address"`
	Notes          string  `json:"notes"`
	Status         string  `json:"status"`

	Subtotal     float64 `json:"subtotal"`
	TaxRate      float64 `json:"tax_rate"`
	TaxAmount    float64 `json:"tax_amount"`
	DeliveryCost float64 `json:"delivery_cost"`
	OtherCost    float64 `json:"other_cost"`
	TotalAmount  float64 `json:"total_amount"`

	// Workflow timestamps
	SubmittedAt *string `json:"submitted_at"`
	ApprovedAt  *string `json:"approved_at"`
	RejectedAt  *string `json:"rejected_at"`
	ConvertedAt *string `json:"converted_at"`

	// ID of the Purchase Order created on conversion
	ConvertedToPurchaseOrderID *string `json:"converted_to_purchase_order_id"`

	Supplier *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"supplier,omitempty"`

	PaymentTerms *PurchaseRequisitionPaymentTermsMini `json:"payment_terms,omitempty"`

	BusinessUnit *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"business_unit,omitempty"`

	Employee *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"employee,omitempty"`

	User *struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user,omitempty"`

	Items []PurchaseRequisitionItemResponse `json:"items"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
