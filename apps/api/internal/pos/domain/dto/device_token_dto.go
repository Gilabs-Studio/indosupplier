package dto

type RegisterPOSDeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=android ios windows linux"`
	OutletID  string `json:"outlet_id" binding:"required,uuid"`
	TenantID  string `json:"tenant_id" binding:"required,uuid"`
}

type POSDeviceTokenResponse struct {
	ID       string `json:"id"`
	OutletID string `json:"outlet_id"`
	TenantID string `json:"tenant_id"`
	Platform string `json:"platform"`
}
