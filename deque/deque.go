package deque

// Deque represents a double-ended queue.
type Deque[T any] struct {
	data []T
}

// NewDeque creates a new Deque.
func NewDeque[T any]() *Deque[T] {
	return &Deque[T]{data: []T{}}
}

// PushFront adds an element to the front of the deque.
func (d *Deque[T]) PushFront(value T) {
	d.data = append([]T{value}, d.data...)
}

// PushBack adds an element to the back of the deque.
func (d *Deque[T]) PushBack(value T) {
	d.data = append(d.data, value)
}

// PopFront removes and returns the front element of the deque.
func (d *Deque[T]) PopFront() (T, bool) {
	if len(d.data) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	front := d.data[0]
	d.data = d.data[1:]
	return front, true
}

// PopBack removes and returns the back element of the deque.
func (d *Deque[T]) PopBack() (T, bool) {
	if len(d.data) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	back := d.data[len(d.data)-1]
	d.data = d.data[:len(d.data)-1]
	return back, true
}

// IsEmpty checks if the deque is empty.
func (d *Deque[T]) IsEmpty() bool {
	return len(d.data) == 0
}

// Size returns the number of elements in the deque.
func (d *Deque[T]) Size() int {
	return len(d.data)
}
