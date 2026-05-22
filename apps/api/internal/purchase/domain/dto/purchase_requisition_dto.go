package dto

type PurchaseRequisitionPaymentTermsMini struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Days *int   `json:"days,omitempty"`
}

type PurchaseRequisitionListResponse struct {
	ID             string  `json:"id"`
	Code           string  `json:"code"`
	SupplierID     *string `json:"supplier_id"`
	PaymentTermsID *string `json:"payment_terms_id"`
	BusinessUnitID *string `json:"business_unit_id"`
	CompanyID      *string `json:"company_id"`
	FiscalYearID   *string `json:"fiscal_year_id"`
	RequestedBy    *string `json:"requested_by"`
	RequestDate    string  `json:"request_date"`
	Status         string  `json:"status"`
	Subtotal       float64 `json:"subtotal"`
	TaxRate        float64 `json:"tax_rate"`
	TaxAmount      float64 `json:"tax_amount"`
	DeliveryCost   float64 `json:"delivery_cost"`
	OtherCost      float64 `json:"other_cost"`
	TotalAmount    float64 `json:"total_amount"`
	Notes          string  `json:"notes"`

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

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
