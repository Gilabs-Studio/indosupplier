package dto

type CreateReviewRequest struct {
	PurchaseOrderID string `json:"purchaseOrderId" binding:"required,uuid"`
	Rating          int    `json:"rating" binding:"required,min=1,max=5"`
	ReviewText      string `json:"reviewText" binding:"required"`
}

type EligibleTransactionResponse struct {
	ID                string  `json:"id"`
	PONumber          string  `json:"poNumber"`
	SupplierProfileID string  `json:"supplierProfileId"`
	SupplierName      string  `json:"supplierName"`
	ProductName       string  `json:"productName"`
	QuantityValue     float64 `json:"quantityValue"`
	QuantityUnit      string  `json:"quantityUnit"`
	TotalAmount       float64 `json:"totalAmount"`
	Date              string  `json:"date"`
}

type ReviewHistoryResponse struct {
	ID                string  `json:"id"`
	PONumber          string  `json:"poNumber"`
	SupplierProfileID string  `json:"supplierProfileId"`
	SupplierName      string  `json:"supplierName"`
	ProductName       string  `json:"productName"`
	QuantityValue     float64 `json:"quantityValue"`
	QuantityUnit      string  `json:"quantityUnit"`
	TotalAmount       float64 `json:"totalAmount"`
	Rating            int     `json:"rating"`
	ReviewText        string  `json:"reviewText"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"date"`
}
