package domain

type PageRequest struct {
	Limit  int
	Offset int
}

type PageResponse[T any] struct {
	Items      []T
	TotalCount int64
	Limit      int
	Offset     int
}
