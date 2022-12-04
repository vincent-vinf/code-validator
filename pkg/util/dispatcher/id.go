package dispatcher

import (
	"errors"
	"sync"
)

var (
	QueueFullErr  = errors.New("queue is full")
	QueueEmptyErr = errors.New("queue is empty")
	ReleaseIDErr  = errors.New("attempt to free an unused id")
)

type Dispatcher struct {
	idSet map[int]struct{}
	queue *RingQueue
	lock  sync.Mutex
}

// NewDispatcher new ring with id range [min,max)
func NewDispatcher(min, max int) (*Dispatcher, error) {
	if min >= max {
		return nil, errors.New("min must be less than max")
	}
	q := NewRingQueue(max - min)
	for i := min; i < max; i++ {
		if err := q.Push(i); err != nil {
			return nil, err
		}
	}

	return &Dispatcher{
		idSet: make(map[int]struct{}, max-min),
		queue: q,
	}, nil
}

func (r *Dispatcher) Get() (int, error) {
	id, err := r.queue.Pop()
	if err != nil {
		return 0, err
	}
	r.idSet[id] = struct{}{}

	return id, nil
}

func (r *Dispatcher) Release(id int) error {
	if _, ok := r.idSet[id]; ok {
		if err := r.queue.Push(id); err != nil {
			return err
		}
		delete(r.idSet, id)

		return nil
	} else {
		return ReleaseIDErr
	}
}

type RingQueue struct {
	data  []int
	start int
	end   int
	len   int
	cap   int
}

func NewRingQueue(cap int) *RingQueue {
	return &RingQueue{
		data: make([]int, cap),
		cap:  cap,
	}
}

func (q *RingQueue) Push(e int) error {
	if q.len == q.cap {
		return QueueFullErr
	}
	q.len++
	q.data[q.end] = e
	q.end = (q.end + 1) % q.cap

	return nil
}

func (q *RingQueue) Pop() (int, error) {
	if q.len == 0 {
		return 0, QueueEmptyErr
	}
	q.len--
	res := q.data[q.start]
	q.start = (q.start + 1) % q.cap

	return res, nil
}

func (q *RingQueue) Len() int {
	return q.len
}

func (q *RingQueue) Cap() int {
	return q.cap
}
