package dto

type CreateRFQRequest struct {
	ProductName   string `json:"product_name" binding:"required,min=3,max=100"`
	Category      string `json:"category" binding:"required"`
	Quantity      string `json:"quantity" binding:"required"`
	Unit          string `json:"unit" binding:"required"`
	TargetPort    string `json:"target_port" binding:"required,min=3,max=100"`
	Description   string `json:"description"`
	AttachmentURL string `json:"attachment_url"`
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
