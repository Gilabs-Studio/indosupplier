package dto

type SupplierProductDto struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`
	Currency     string   `json:"currency"`
	MinOrder     string   `json:"minOrder"`
	CapacityText string   `json:"capacityText"`
	CategoryName string   `json:"categoryName"`
	Photos       []string `json:"photos"`
}

type SupplierCertificationDto struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Institution string `json:"institution"`
	Year        int    `json:"year"`
}

type PublicSupplierDto struct {
	ID                string                     `json:"id"`
	Slug              string                     `json:"slug"`
	CompanyName       string                     `json:"companyName"`
	BusinessType      string                     `json:"businessType"`
	EstablishedYear   int                        `json:"establishedYear"`
	EmployeeCount     string                     `json:"employeeCount"`
	Location          string                     `json:"location"`
	Province          string                     `json:"province"`
	Address           string                     `json:"address"`
	Description       string                     `json:"description"`
	IsVerified        bool                       `json:"isVerified"`
	VerificationLevel int                        `json:"verificationLevel"`
	IsPremiumVerified bool                       `json:"isPremiumVerified"`
	TaxStatus         string                     `json:"taxStatus"`
	ResponseRate      float64                    `json:"responseRate"`
	ResponseTime      string                     `json:"responseTime"`
	Rating            float64                    `json:"rating"`
	ReviewCount       int                        `json:"reviewCount"`
	KeyProducts       []string                   `json:"keyProducts"`
	Certifications    []string                   `json:"certifications"`
	Products          []SupplierProductDto       `json:"products"`
	CertificationList []SupplierCertificationDto `json:"certificationList"`
	Reviews           []PublicReviewDto          `json:"reviews"`
	Phone             string                     `json:"phone,omitempty"`
	WhatsApp          string                     `json:"whatsApp,omitempty"`
	Email             string                     `json:"email,omitempty"`
	Website           string                     `json:"website,omitempty"`
}

type PublicProductDto struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Price               float64  `json:"price"`
	Currency            string   `json:"currency"`
	MinOrder            string   `json:"minOrder"`
	CapacityText        string   `json:"capacityText"`
	CategoryName        string   `json:"categoryName"`
	Photos              []string `json:"photos"`
	SupplierID          string   `json:"supplierId"`
	SupplierCompanyName string   `json:"supplierCompanyName"`
	SupplierSlug        string   `json:"supplierSlug"`
	SupplierLocation    string   `json:"supplierLocation"`
	SupplierVerified    bool     `json:"supplierVerified"`
	SupplierRating      float64  `json:"supplierRating"`
	SupplierReviewCount int      `json:"supplierReviewCount"`
}

type PublicReviewDto struct {
	ID                string `json:"id"`
	BuyerName         string `json:"buyerName"`
	Rating            int    `json:"rating"`
	ReviewText        string `json:"reviewText"`
	SupplierReply     string `json:"supplierReply"`
	SupplierRepliedAt string `json:"supplierRepliedAt,omitempty"`
	CreatedAt         string `json:"createdAt"`
}

type PublicProductDetailDto struct {
	Product         PublicProductDto   `json:"product"`
	Supplier        PublicSupplierDto  `json:"supplier"`
	Reviews         []PublicReviewDto  `json:"reviews"`
	RelatedProducts []PublicProductDto `json:"relatedProducts"`
}
