package httpx

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func NewPage[T any](items []T, nextCursor string) Page[T] {
	if len(items) == 0 {
		items = make([]T, 0)
	}
	return Page[T]{
		Items:      items,
		NextCursor: nextCursor,
	}
}
