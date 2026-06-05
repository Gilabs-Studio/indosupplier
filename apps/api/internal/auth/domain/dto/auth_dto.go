package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
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
