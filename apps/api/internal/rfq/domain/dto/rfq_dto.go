package dto

type CreateRFQRequest struct {
	ProductName      string   `json:"product_name" binding:"required,min=3,max=100"`
	Category         string   `json:"category" binding:"required"`
	Quantity         string   `json:"quantity" binding:"required"`
	Unit             string   `json:"unit" binding:"required"`
	TargetPort       string   `json:"target_port,omitempty" binding:"omitempty,max=100"`
	Description      string   `json:"description"`
	AttachmentURL    string   `json:"attachment_url"`
	AttachmentName   string   `json:"attachment_name"`
	AttachmentSize   int64    `json:"attachment_size"`
	TargetSupplierID string   `json:"target_supplier_id,omitempty"`
	ProductID        *string  `json:"product_id,omitempty"`
	ImageURL         string   `json:"image_url,omitempty"`
	Budget           float64  `json:"budget,omitempty"`
	DeliveryTimeline string   `json:"delivery_timeline,omitempty"`
}

type RFQResponse struct {
	ID             string `json:"id"`
	Product        string `json:"product"`
	Category       string `json:"category"`
	Quantity       string `json:"quantity"`
	TargetPort     string `json:"targetPort,omitempty"`
	Date           string `json:"date"`
	Status         string `json:"status"`
	Replies        int    `json:"replies"`
	ImageURL       string `json:"imageUrl,omitempty"`
	Description    string `json:"description,omitempty"`
	AttachmentURL  string `json:"attachmentUrl,omitempty"`
	AttachmentName string `json:"attachmentName,omitempty"`
	AttachmentSize string               `json:"attachmentSize,omitempty"`
	Suppliers      []RFQSupplierSummary `json:"suppliers,omitempty"`
}

type RFQSupplierSummary struct {
	ID                string `json:"id"`
	SupplierProfileID string `json:"supplierProfileId"`
	SupplierName      string `json:"supplierName"`
	Status            string `json:"status"`
	Price             string `json:"price,omitempty"`
	IsAccepted        bool   `json:"isAccepted"`
}

type RFQBidResponse struct {
	ID                string `json:"id"`
	SupplierName      string `json:"supplierName"`
	SupplierProfileID string `json:"supplierProfileId,omitempty"`
	Price             string `json:"price"`
	MOQ               string `json:"moq"`
	ResponseTime      string `json:"responseTime"`
	Verified          bool   `json:"verified"`
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
	Port           string                     `json:"port,omitempty"`
	Date           string                     `json:"date"`
	Budget         string                     `json:"budget"`
	Status         string                     `json:"status"`
	ImageURL       string                     `json:"imageUrl,omitempty"`
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

type RFQThreadSupplierDTO struct {
	SupplierProfileID string  `json:"supplierProfileId"`
	SupplierName      string  `json:"supplierName"`
	SupplierLogo      string  `json:"supplierLogo,omitempty"`
	Verified          bool    `json:"verified"`
	Rating            float64 `json:"rating"`
	City              string  `json:"city,omitempty"`
	LatestOffer       string  `json:"latestOffer,omitempty"`
	LatestOfferAt     string  `json:"latestOfferAt,omitempty"`
	Status            string  `json:"status"`
	MessageCount      int     `json:"messageCount"`
}

type RFQMessageDTO struct {
	ID                 string   `json:"id"`
	RFQID              string   `json:"rfqId"`
	SupplierProfileID  string   `json:"supplierProfileId"`
	SenderType         string   `json:"senderType"` // "buyer", "supplier", "system"
	SenderID           string   `json:"senderId"`
	SenderName         string   `json:"senderName"`
	SenderAvatar       string   `json:"senderAvatar,omitempty"`
	SenderRole         string   `json:"senderRole,omitempty"`
	SenderRating       float64  `json:"senderRating,omitempty"`
	MessageType        string   `json:"messageType"` // "message", "offer", "bid_accepted", "system"
	Body               string   `json:"body"`
	Price              *float64 `json:"price,omitempty"`
	PriceFormatted     string   `json:"priceFormatted,omitempty"`
	MOQ                string   `json:"moq,omitempty"`
	DeliveryTime       string   `json:"deliveryTime,omitempty"`
	CreatedAt          string   `json:"createdAt"`
	CreatedAtFormatted string   `json:"createdAtFormatted"` // "07/09/2026 23:08:03 WIB"
	IsMine             bool     `json:"isMine"`
}

type SendRFQMessageRequest struct {
	Body         string `json:"body" binding:"required,min=1,max=2000"`
	Price        string `json:"price" binding:"omitempty,max=120"`
	MOQ          string `json:"moq" binding:"omitempty,max=120"`
	DeliveryTime string `json:"deliveryTime" binding:"omitempty,max=120"`
}

