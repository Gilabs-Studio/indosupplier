package response

// Response is a generic response wrapper
type Response[T any] struct {
	Success   bool      `json:"success"`
	Data      T         `json:"data"`
	Meta      *Meta     `json:"meta,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp string    `json:"timestamp"`
	RequestID string    `json:"request_id"`
}

// PaginatedResponse is a generic response wrapper for lists
type PaginatedResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      []T       `json:"data"`
	Meta      *Meta     `json:"meta,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp string    `json:"timestamp"`
	RequestID string    `json:"request_id"`
}
