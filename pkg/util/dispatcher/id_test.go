package dispatcher

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func BenchmarkParallel(b *testing.B) {
	ring, err := NewDispatcher(1, 12)
	if err != nil {
		b.Fatal(err)
	}
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id, err := ring.Get()
			for errors.Is(err, QueueEmptyErr) {
				time.Sleep(time.Microsecond)
				id, err = ring.Get()
			}

			if _, ok := m.Load(id); ok {
				b.Fatal("id crash")
			}

			m.Store(id, struct{}{})
			time.Sleep(time.Millisecond * 2)
			m.Delete(id)
			ring.Release(id)
		}
	})
}
