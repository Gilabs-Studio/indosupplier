package dto

type AddCompareRequest struct {
	SupplierProfileID string `json:"supplierProfileId" binding:"required,uuid"`
}

type AddProductCompareRequest struct {
	SupplierProductID string `json:"supplierProductId" binding:"required,uuid"`
}

type ComparedSupplierResponse struct {
	ID              string   `json:"id"`
	CompanyName     string   `json:"companyName"`
	Slug            string   `json:"slug"`
	Location        string   `json:"location"`
	BusinessType    string   `json:"businessType"`
	EstablishedYear int      `json:"establishedYear"`
	Rating          float64  `json:"rating"`
	ReviewCount     int      `json:"reviewCount"`
	Verified        bool     `json:"verified"`
	MOQ             string   `json:"moq"`
	ResponseTime    string   `json:"responseTime"`
	Capacity        string   `json:"capacity"`
	Certifications  []string `json:"certifications"`
}

type ComparedProductResponse struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	ImageUrl             string  `json:"imageUrl"`
	Price                float64 `json:"price"`
	MOQ                  string  `json:"moq"`
	Capacity             string  `json:"capacity"`
	CategoryName         string  `json:"categoryName"`
	Description          string  `json:"description"`
	SupplierID           string  `json:"supplierId"`
	SupplierCompanyName  string  `json:"supplierCompanyName"`
	SupplierSlug         string  `json:"supplierSlug"`
	SupplierRating       float64 `json:"supplierRating"`
	SupplierReviewCount  int     `json:"supplierReviewCount"`
	SupplierVerified     bool    `json:"supplierVerified"`
	SupplierLocation     string  `json:"supplierLocation"`
	SupplierResponseTime string  `json:"supplierResponseTime"`
}

