package queue

type ItemType interface {
	bool | int64 | float64 | string
}

type Base[T ItemType] struct {
	maxSize uint64
	items   []T
	pointer uint64
}

func (q *Base[T]) Size() uint64 {
	return q.pointer
}

func (q *Base[T]) Empty() bool {
	return q.pointer == 0
}

func (q *Base[T]) Full() bool {
	return q.pointer >= q.maxSize
}
