package dispatcher

import (
	"errors"

	"golang.design/x/lockfree"
)

var (
	SpaceFullErr = errors.New("no free space")
)

type Dispatcher struct {
	queue *lockfree.Queue
}

// NewDispatcher new ring with id range [min,max)
func NewDispatcher(min, max int) (*Dispatcher, error) {
	if min >= max {
		return nil, errors.New("min must be less than max")
	}
	q := lockfree.NewQueue()
	for i := min; i < max; i++ {
		q.Enqueue(i)
	}

	return &Dispatcher{
		queue: q,
	}, nil
}

func (r *Dispatcher) Get() (int, error) {
	id := r.queue.Dequeue()
	if id == nil {
		return 0, SpaceFullErr
	}

	return id.(int), nil
}

func (r *Dispatcher) Release(id int) {
	r.queue.Enqueue(id)
}
