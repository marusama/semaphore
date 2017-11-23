package semaphore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Semaphore2 interface {
	Acquire(ctx context.Context) error
	Release()
	SetLimit(limit int)
	GetLimit() int
	GetCount() int
}

type semaphore2 struct {
	state       uint64
	lock        sync.RWMutex
	broadcastCh *chan struct{}
}

func New2(limit int) Semaphore2 {
	broadcastCh := make(chan struct{})
	return &semaphore2{
		state:       uint64(limit) << 32,
		broadcastCh: &broadcastCh,
	}
}

func (s *semaphore2) Acquire(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("ctx.Done()")
		default:
		}

		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		newCount := count + 1
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
			select {
			case <-ctx.Done():
				return errors.New("ctx.Done()")
			// wait for broadcast
			case <-broadcastCh:
			}
		}
	}
}

func (s *semaphore2) Release() {
	for {
		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		if count == 0 {
			panic("Release without acquire")
		}
		newCount := count - 1
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

func (s *semaphore2) SetLimit(limit int) {
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
	panic("unreachable")
}

func (s *semaphore2) GetCount() int {
	state := atomic.LoadUint64(&s.state)
	return int(state & 0xFFFFFFFF)
}

func (s *semaphore2) GetLimit() int {
	state := atomic.LoadUint64(&s.state)
	return int(state >> 32)
}
