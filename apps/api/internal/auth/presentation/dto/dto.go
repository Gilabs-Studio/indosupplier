package dto

type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseDTO struct {
	User         UserDTO `json:"user"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
}

type UserDTO struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Email           string                 `json:"email"`
	Capabilities    AccountCapabilitiesDTO `json:"capabilities"`
	BuyerProfile    *AccountProfileRefDTO  `json:"buyer_profile,omitempty"`
	SupplierProfile *AccountProfileRefDTO  `json:"supplier_profile,omitempty"`
}

type AccountCapabilitiesDTO struct {
	Buyer    bool `json:"buyer"`
	Supplier bool `json:"supplier"`
}

type AccountProfileRefDTO struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
}
