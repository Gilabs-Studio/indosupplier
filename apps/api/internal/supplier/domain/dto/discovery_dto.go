package dto

type SupplierProductDto struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	MinOrder    string  `json:"minOrder"`
	Photos      []string `json:"photos"`
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
	Address           string                     `json:"address"`
	Description       string                     `json:"description"`
	IsVerified        bool                       `json:"isVerified"`
	Rating            float64                    `json:"rating"`
	ReviewCount       int                        `json:"reviewCount"`
	KeyProducts       []string                   `json:"keyProducts"`
	Certifications    []string                   `json:"certifications"`
	Products          []SupplierProductDto       `json:"products"`
	CertificationList []SupplierCertificationDto `json:"certificationList"`
	Phone             string                     `json:"phone,omitempty"`
	Email             string                     `json:"email,omitempty"`
	Website           string                     `json:"website,omitempty"`
}
