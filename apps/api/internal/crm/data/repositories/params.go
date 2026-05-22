package repositories

// ListParams defines common parameters for listing CRM entities
type ListParams struct {
	Search  string
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}
