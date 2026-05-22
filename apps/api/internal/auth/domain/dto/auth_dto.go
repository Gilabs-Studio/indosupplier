package dto

import "time"

// LoginRequest represents login request DTO
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterTenantRequest represents self-service tenant registration payload.
// Access is gated by one of:
//   - coupon: a valid promotional code granting a trial subscription
//   - plan + billing_period: a paid subscription handled via Xendit invoice
type RegisterTenantRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	// Coupon bypasses payment and grants a trial subscription.
	// Only alphanumeric chars accepted; input is normalised to uppercase.
	Coupon string `json:"coupon" binding:"omitempty,min=4,max=64"`
	// Plan is required when using paid registration (no coupon).
	// Valid values: pos_essential, pos_growth, pos_enterprise,
	//               erp_core, erp_pro, erp_enterprise,
	//               crm_basic, crm_growth, crm_enterprise,
	//               hr_basic, hr_growth, hr_enterprise, full_access
	Plan string `json:"plan" binding:"omitempty"`
	// BillingPeriod is required when Plan is set. Values: "monthly", "yearly".
	BillingPeriod string `json:"billing_period" binding:"omitempty,oneof=monthly yearly"`
	// UserCount is the number of users for per-user billing. Defaults to 1 when omitted.
	UserCount int `json:"user_count" binding:"omitempty,min=1,max=999"`
	// CompanyName is required because every registration must be tied to at least one company.
	CompanyName string `json:"company_name" binding:"required,min=2,max=150"`
}

// RegisterInitResponse is returned when a paid plan is selected during registration.
// The frontend must redirect to InvoiceURL to complete payment; the tenant is
// provisioned by the Xendit webhook after successful payment.
type RegisterInitResponse struct {
	InvoiceURL  string `json:"invoice_url"`
	InvoiceID   string `json:"invoice_id"`
	ExpiresAt   string `json:"expires_at"`
}

// ConfirmPendingRegistrationRequest confirms a paid registration token and
// provisions tenant/user when needed, then returns an authenticated session.
type ConfirmPendingRegistrationRequest struct {
	Token string `json:"token" binding:"required,uuid"`
}

// LoginResponse represents login response DTO
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"` // in seconds
}

// SubscriptionAccessResponse describes tenant subscription enforcement state.
// Frontend can use this to render banners and enforce billing redirects.
type SubscriptionAccessResponse struct {
	State                string `json:"state"`
	Enforcement          string `json:"enforcement"`
	DaysOverdue          int    `json:"days_overdue"`
	GracePeriodDays      int    `json:"grace_period_days"`
	ForceBillingRedirect bool   `json:"force_billing_redirect"`
	AllowRead            bool   `json:"allow_read"`
	AllowWrite           bool   `json:"allow_write"`
	Message              string `json:"message,omitempty"`
	BillingPath          string `json:"billing_path"`
}

// UserResponse represents user response DTO for auth
type UserResponse struct {
	ID               string            `json:"id"`
	Email            string            `json:"email"`
	Name             string            `json:"name"`
	AvatarURL        string            `json:"avatar_url"`
	EmployeeID       string            `json:"employee_id,omitempty"`
	Role             string            `json:"role"`
	RoleName         string            `json:"role_name"`
	RoleDataScope    string            `json:"role_data_scope"`
	// IsOwner is true when the logged-in user holds the protected tenant-owner role.
	// Consumers should use this flag to gate owner-only UI (Billing, Payment) instead of
	// pattern-matching on the role code string.
	IsOwner          bool              `json:"is_owner"`
	Permissions      map[string]string `json:"permissions"` // code -> scope (e.g., {"sales_order.read": "DIVISION"})
	Status           string            `json:"status"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	TenantID         string            `json:"tenant_id,omitempty"`
	TenantName       string            `json:"tenant_name,omitempty"`
	SubscriptionPlan string            `json:"subscription_plan,omitempty"`
	SubscriptionAccess *SubscriptionAccessResponse `json:"subscription_access,omitempty"`
}
