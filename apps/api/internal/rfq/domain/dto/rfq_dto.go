package dto

type CreateRFQRequest struct {
	ProductName      string   `json:"product_name" binding:"required,min=3,max=100"`
	Category         string   `json:"category" binding:"required"`
	Quantity         string   `json:"quantity" binding:"required"`
	Unit             string   `json:"unit" binding:"required"`
	TargetPort       string   `json:"target_port" binding:"required,min=3,max=100"`
	Description      string   `json:"description"`
	AttachmentURL    string   `json:"attachment_url"`
	AttachmentName   string   `json:"attachment_name"`
	AttachmentSize   int64    `json:"attachment_size"`
	TargetSupplierID string   `json:"target_supplier_id,omitempty"`
	ProductID        *string  `json:"product_id,omitempty"`
	Budget           float64  `json:"budget,omitempty"`
	DeliveryTimeline string   `json:"delivery_timeline,omitempty"`
}

type RFQResponse struct {
	ID             string `json:"id"`
	Product        string `json:"product"`
	Category       string `json:"category"`
	Quantity       string `json:"quantity"`
	TargetPort     string `json:"targetPort"`
	Date           string `json:"date"`
	Status         string `json:"status"`
	Replies        int    `json:"replies"`
	Description    string `json:"description,omitempty"`
	AttachmentURL  string `json:"attachmentUrl,omitempty"`
	AttachmentName string `json:"attachmentName,omitempty"`
	AttachmentSize string `json:"attachmentSize,omitempty"`
}

type RFQBidResponse struct {
	ID           string `json:"id"`
	SupplierName string `json:"supplierName"`
	Price        string `json:"price"`
	MOQ          string `json:"moq"`
	ResponseTime string `json:"responseTime"`
	Verified     bool   `json:"verified"`
}

type SupplierSubmittedOfferDto struct {
	Price        string `json:"price"`
	MOQ          string `json:"moq"`
	DeliveryTime string `json:"deliveryTime"`
	Notes        string `json:"notes"`
	RespondedAt  string `json:"respondedAt"`
}

type SupplierRFQResponse struct {
	ID             string                     `json:"id"`
	Product        string                     `json:"product"`
	Category       string                     `json:"category"`
	Quantity       string                     `json:"quantity"`
	Port           string                     `json:"port"`
	Date           string                     `json:"date"`
	Budget         string                     `json:"budget"`
	Status         string                     `json:"status"`
	Description    string                     `json:"description,omitempty"`
	ShippingTerm   string                     `json:"shippingTerm,omitempty"`
	TargetDelivery string                     `json:"targetDelivery,omitempty"`
	SubmittedOffer *SupplierSubmittedOfferDto `json:"submittedOffer,omitempty"`
	Buyer          struct {
		Name        string `json:"name"`
		Established string `json:"established"`
		Location    string `json:"location"`
		Rating      string `json:"rating"`
	} `json:"buyer"`
}

type SubmitRFQProposalRequest struct {
	Price        string `json:"price" binding:"required,min=1,max=120"`
	MOQ          string `json:"moq" binding:"required,min=1,max=120"`
	DeliveryTime string `json:"deliveryTime" binding:"required,min=1,max=120"`
	Notes        string `json:"notes" binding:"required,min=3,max=1000"`
}
