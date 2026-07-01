package dto

type SupplierProfileDTO struct {
	ID           string `json:"id"`
	CompanyName  string `json:"companyName"`
	BusinessType string `json:"businessType"`
	Established  string `json:"established"`
	Employees    string `json:"employees"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	TaxID        string `json:"taxId"`
	NIB          string `json:"nib"`
	Overview     string `json:"overview"`
	Location     string `json:"location"`
	Status       string `json:"status"`
	Logo         string `json:"logo,omitempty"`
}

type UpdateProfileRequest struct {
	CompanyName  string `json:"companyName" binding:"required"`
	BusinessType string `json:"businessType" binding:"required"`
	Established  string `json:"established"`
	Employees    string `json:"employees"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	Website      string `json:"website"`
	TaxID        string `json:"taxId"`
	NIB          string `json:"nib"`
	Overview     string `json:"overview"`
	Location     string `json:"location" binding:"required"`
	Logo         string `json:"logo"`
}

type BillingInvoiceDTO struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Status      string `json:"status"` // paid, pending, failed
	ReceiptURL  string `json:"receiptUrl"`
}

type SubscriptionDetailDTO struct {
	PlanID      string `json:"planId"`
	PlanName    string `json:"planName"`
	Price       string `json:"price"`
	Period      string `json:"period"`
	Active      bool   `json:"active"`
	RenewalDate string `json:"renewalDate"`
}

type MeteredUsageStatsDTO struct {
	ProductUploadsUsed  int    `json:"productUploadsUsed"`
	ProductUploadsLimit string `json:"productUploadsLimit"`
	RfqBidsUsed         int    `json:"rfqBidsUsed"`
	RfqBidsLimit        string `json:"rfqBidsLimit"`
	AuctionSlotsUsed    int    `json:"auctionSlotsUsed"`
	AuctionSlotsLimit   string `json:"auctionSlotsLimit"`
}

type BillingOverviewResponse struct {
	CurrentMeteredUsage  string                  `json:"currentMeteredUsage"`
	CurrentIncludedUsage string                  `json:"currentIncludedUsage"`
	NextPaymentDue       string                  `json:"nextPaymentDue"`
	NextPaymentDate      string                  `json:"nextPaymentDate"`
	Subscriptions        []SubscriptionDetailDTO `json:"subscriptions"`
	MeteredUsage         MeteredUsageStatsDTO    `json:"meteredUsage"`
	Invoices             []BillingInvoiceDTO     `json:"invoices"`
}

type UpgradePlanRequest struct {
	PlanID string `json:"planId" binding:"required"`
}
