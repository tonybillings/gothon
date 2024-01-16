package queue

// Queue NOT thread-safe!
type Queue[T ItemType] interface {
	Size() uint64
	Empty() bool
	Full() bool
	Put(T) bool
	Get() (T, bool)
}

// New Returns either a FiFo queue or LiFo queue (stack).
// Pass 0 for maxSize for no capacity limit.
func New[T ItemType](maxSize uint64, lastInFirstOut ...bool) Queue[T] {
	if len(lastInFirstOut) > 0 && lastInFirstOut[0] {
		q := &Lifo[T]{}
		q.maxSize = maxSize
		return q
	}

	q := &Fifo[T]{}
	q.maxSize = maxSize
	return q
}
