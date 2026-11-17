package domain

// PaginatedResponse is a generic paginated response wrapper
type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int   `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse[T any](data []T, total int, limit int, offset int) *PaginatedResponse[T] {
	if data == nil {
		data = make([]T, 0)
	}
	return &PaginatedResponse[T]{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
