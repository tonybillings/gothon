package queue

// Lifo NOT thread-safe!  Also, note the absence of any
// cleanup logic!  IO performance is given priority here.
type Lifo[T ItemType] struct {
	Base[T]
}

func (q *Lifo[T]) Put(value T) (ok bool) {
	if q.maxSize == 0 {
		q.items = append(q.items, value)
	} else {
		if q.pointer >= q.maxSize {
			return false
		}
		q.items[q.pointer] = value
	}
	q.pointer++
	return true
}

func (q *Lifo[T]) Get() (val T, ok bool) {
	if q.pointer == 0 {
		var empty T
		return empty, false
	}
	q.pointer--
	return q.items[q.pointer], true
}
