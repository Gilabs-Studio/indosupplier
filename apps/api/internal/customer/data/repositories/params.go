package repositories

// ListParams contains common list parameters
type ListParams struct {
	Search     string
	Limit      int
	Offset     int
	SortBy     string
	SortDir    string
	ActiveOnly *bool
}

// CustomerListParams contains customer-specific list parameters
type CustomerListParams struct {
	ListParams
	CustomerTypeID  string
	IsLoyaltyMember *bool
}
