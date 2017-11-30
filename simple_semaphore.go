package semaphore

import (
	"sync/atomic"
)

type SimpleSemaphore interface {
	Acquire() error
	Release()
	GetLimit() int
	GetCount() int
}

type simpleSemaphore struct {
	limit       uint32
	count       uint32
	broadcastCh chan struct{}
}

func NewSimple(limit int) SimpleSemaphore {
	if limit <= 0 {
		panic("semaphore limit must be greater than 0")
	}
	broadcastCh := make(chan struct{})
	return &simpleSemaphore{
		limit:       uint32(limit),
		broadcastCh: broadcastCh,
	}
}

func (s *simpleSemaphore) Acquire() error {
	for {
		count := atomic.LoadUint32(&s.count)
		newCount := count + 1
		if newCount <= s.limit {
			if atomic.CompareAndSwapUint32(&s.count, count, newCount) {
				return nil
			} else {
				continue
			}
		} else {
			select {
			// wait for broadcast
			case <-s.broadcastCh:
			}
		}
	}
}

func (s *simpleSemaphore) Release() {
	for {
		count := atomic.LoadUint32(&s.count)
		if count == 0 {
			panic("semaphore release without acquire")
		}
		if atomic.CompareAndSwapUint32(&s.count, count, count - 1) {
			select {
				case s.broadcastCh <- struct{}{}:
				default:
			}
			return
		}
	}
}

func (s *simpleSemaphore) GetCount() int {
	return int(atomic.LoadUint32(&s.count))
}

func (s *simpleSemaphore) GetLimit() int {
	return int(s.limit)
}
