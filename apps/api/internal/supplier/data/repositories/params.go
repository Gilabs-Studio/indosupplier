package repositories

// ListParams defines common parameters for listing entities
type ListParams struct {
	Search     string
	Limit      int
	Offset     int
	SortBy     string
	SortDir    string
	ActiveOnly bool // when true, only active (is_active=true) records are returned
}

// SupplierListParams extends ListParams with supplier-specific filters
type SupplierListParams struct {
	ListParams
	SupplierTypeID string
	Status         string
	IsApproved     *bool
}
