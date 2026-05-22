package repositories

// ListParams defines common parameters for listing entities
type ListParams struct {
	Search     string
	SortBy     string
	SortDir    string
	Limit      int
	Offset     int
	ActiveOnly bool // when true, only active (is_active=true) records are returned
}
