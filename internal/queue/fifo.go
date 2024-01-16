package queue

// Fifo NOT thread-safe!  Also, note the absence of any
// cleanup logic!  IO performance is given priority here.
type Fifo[T ItemType] struct {
	Base[T]
}

func (q *Fifo[T]) Put(value T) (ok bool) {
	if q.maxSize != 0 && q.pointer >= q.maxSize {
		return false
	}
	q.items = append(q.items, value)
	q.pointer++
	return true
}

func (q *Fifo[T]) Get() (val T, ok bool) {
	if q.pointer == 0 {
		var empty T
		return empty, false
	}
	q.pointer--
	result := q.items[0]
	q.items = q.items[1:]
	return result, true
}
