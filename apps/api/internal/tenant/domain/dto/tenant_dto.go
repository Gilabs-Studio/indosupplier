package dto

// TenantListResponse is returned when listing or fetching a single tenant (system admin only)
type TenantListResponse struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	Slug                   string  `json:"slug"`
	Status                 string  `json:"status"`
	Plan                   string  `json:"plan"`
	MaxUsers               int     `json:"max_users"`
	CurrentUsers           int     `json:"current_users"`
	CompanyCount           int     `json:"company_count"`
	OutletCount            int     `json:"outlet_count"`
	WarehouseCount         int     `json:"warehouse_count"`
	OwnerUserID            string  `json:"owner_user_id,omitempty"`
	OwnerName              string  `json:"owner_name,omitempty"`
	OwnerEmail             string  `json:"owner_email,omitempty"`
	DeletionRequestedAt    *string `json:"deletion_requested_at,omitempty"`
	DeletionScheduledAt    *string `json:"deletion_scheduled_at,omitempty"`
	DeletionRequestedBy    string  `json:"deletion_requested_by,omitempty"`
	DeletionPreviousStatus string  `json:"deletion_previous_status,omitempty"`
	CreatedAt              *string `json:"created_at,omitempty"`
}

// TenantListParams holds query params for listing tenants
type TenantListParams struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Search  string `form:"search"`
}
