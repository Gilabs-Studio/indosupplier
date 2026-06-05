package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	CompanyName string `json:"company_name" binding:"omitempty,min=2"`
	Industry    string `json:"industry" binding:"omitempty,min=2"`
}

type SupplierOnboardingRequest struct {
	CompanyName       string `json:"company_name" binding:"required,min=2"`
	PrimaryCategory   string `json:"primary_category" binding:"required,min=2"`
	Subcategory       string `json:"subcategory" binding:"required,min=2"`
	ProvinceID        string `json:"province_id" binding:"required,min=2"`
	CityID            string `json:"city_id" binding:"required,min=2"`
	Phone             string `json:"phone" binding:"required,min=6"`
	WhatsApp          string `json:"whatsapp" binding:"required,min=6"`
	Email             string `json:"email" binding:"required,email"`
	Website           string `json:"website" binding:"omitempty"`
	CompanyType       string `json:"company_type" binding:"required,min=2"`
	TaxStatus         string `json:"tax_status" binding:"required,min=2"`
	NPWP              string `json:"npwp" binding:"omitempty,min=4"`
	NIB               string `json:"nib" binding:"omitempty,min=4"`
	Address           string `json:"address" binding:"omitempty"`
	BusinessHours     string `json:"business_hours" binding:"omitempty"`
	Timezone          string `json:"timezone" binding:"omitempty"`
	Description       string `json:"description" binding:"required,min=10"`
	FirstProductName  string `json:"first_product_name" binding:"omitempty,min=2"`
	FirstProductPrice string `json:"first_product_price" binding:"omitempty"`
}

type LoginResponse struct {
	User         *UserResponse `json:"user"`
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"`
}

type UserResponse struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Email           string                      `json:"email"`
	Capabilities    AccountCapabilitiesResponse `json:"capabilities"`
	BuyerProfile    *AccountProfileRefResponse  `json:"buyer_profile,omitempty"`
	SupplierProfile *AccountProfileRefResponse  `json:"supplier_profile,omitempty"`
}

type AccountCapabilitiesResponse struct {
	Buyer    bool `json:"buyer"`
	Supplier bool `json:"supplier"`
}

type AccountProfileRefResponse struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
}
