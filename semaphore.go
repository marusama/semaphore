package semaphore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Semaphore interface {
	Acquire(ctx context.Context, n int) error
	TryAcquire(n int) bool
	Release(n int)
	SetLimit(limit int)
	GetLimit() int
	GetCount() int
}

var (
	ErrCtxDone = errors.New("ctx.Done()")
)

type semaphore struct {
	state       uint64
	lock        sync.RWMutex
	broadcastCh *chan struct{}
}

func New(limit int) Semaphore {
	if limit <= 0 {
		panic("semaphore limit must be greater than 0")
	}
	broadcastCh := make(chan struct{})
	return &semaphore{
		state:       uint64(limit) << 32,
		broadcastCh: &broadcastCh,
	}
}

func (s *semaphore) Acquire(ctx context.Context, n int) error {
	if n <= 0 {
		panic("n must be positive number")
	}
	for {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ErrCtxDone
			default:
			}
		}

		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		newCount := count + uint64(n)
		if newCount <= limit {
			if atomic.CompareAndSwapUint64(&s.state, state, limit<<32+newCount) {
				return nil
			} else {
				continue
			}
		} else {
			s.lock.RLock()
			broadcastCh := *s.broadcastCh
			s.lock.RUnlock()

			if ctx != nil {
				select {
				case <-ctx.Done():
					return ErrCtxDone
				// wait for broadcast
				case <-broadcastCh:
				}
			} else {
				select {
				// wait for broadcast
				case <-broadcastCh:
				}
			}
		}
	}
}

func (s *semaphore) TryAcquire(n int) bool {
	if n <= 0 {
		panic("n must be positive number")
	}

	for {
		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		newCount := count + uint64(n)
		if newCount <= limit {
			if atomic.CompareAndSwapUint64(&s.state, state, limit<<32+newCount) {
				return true
			} else {
				continue
			}
		} else {
			return false
		}
	}
}

func (s *semaphore) Release(n int) {
	if n <= 0 {
		panic("n must be positive number")
	}
	for {
		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		if count == 0 {
			panic("semaphore release without acquire")
		}
		newCount := count - uint64(n)
		if atomic.CompareAndSwapUint64(&s.state, state, state&0xFFFFFFFF00000000+newCount) {
			if count >= limit {
				newBroadcastCh := make(chan struct{})
				s.lock.Lock()
				oldBroadcastCh := s.broadcastCh
				s.broadcastCh = &newBroadcastCh
				s.lock.Unlock()

				// send broadcast signal
				close(*oldBroadcastCh)
			}
			return
		}
	}
}

func (s *semaphore) SetLimit(limit int) {
	if limit <= 0 {
		panic("semaphore limit must be greater than 0")
	}
	for {
		state := atomic.LoadUint64(&s.state)
		if atomic.CompareAndSwapUint64(&s.state, state, uint64(limit)<<32+state&0xFFFFFFFF) {
			newBroadcastCh := make(chan struct{})
			s.lock.Lock()
			oldBroadcastCh := s.broadcastCh
			s.broadcastCh = &newBroadcastCh
			s.lock.Unlock()

			// send broadcast signal
			close(*oldBroadcastCh)
			return
		}
	}
}

func (s *semaphore) GetCount() int {
	state := atomic.LoadUint64(&s.state)
	return int(state & 0xFFFFFFFF)
}

func (s *semaphore) GetLimit() int {
	state := atomic.LoadUint64(&s.state)
	return int(state >> 32)
}
