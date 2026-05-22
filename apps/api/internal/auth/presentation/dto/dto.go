package dto

// LoginRequestDTO represents login payload
type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequestDTO represents register payload
type RegisterRequestDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleCode string `json:"role_code" binding:"required"` // Can be optional if default role exists
}

// LoginResponseDTO represents login response
type LoginResponseDTO struct {
	User        UserDTO  `json:"user"`
	AccessToken string   `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
}

type UserDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	AvatarURL   string            `json:"avatar_url"`
	EmployeeID  string            `json:"employee_id,omitempty"`
	Role        RoleDTO           `json:"role"`
	Permissions map[string]string `json:"permissions"` // code -> scope
	TenantID    string            `json:"tenant_id,omitempty"`
	TenantName  string            `json:"tenant_name,omitempty"`
	SubscriptionPlan string       `json:"subscription_plan,omitempty"`
	SubscriptionAccess *SubscriptionAccessDTO `json:"subscription_access,omitempty"`
}

type SubscriptionAccessDTO struct {
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

type RoleDTO struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	DataScope string `json:"data_scope"`
	// IsOwner is true when this role is the unique, protected tenant-owner role
	// generated during tenant registration. Use this to gate Billing/Payment UI
	// instead of fragile string-matching on the role code.
	IsOwner bool `json:"is_owner"`
}
