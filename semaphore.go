package semaphore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Semaphore interface {
	Acquire(ctx context.Context) error
	Release()
	SetLimit(limit int)
	GetLimit() int
	GetCount() int
}

type semaphore struct {
	state uint64
	lock  sync.Mutex
	cond  sync.Cond
}

func New(limit int) Semaphore {
	s := &semaphore{state: uint64(limit) << 32}
	s.cond.L = &s.lock
	return s
}

func (s *semaphore) Acquire(ctx context.Context) error {
	for {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return errors.New("ctx.Done()")
			default:
			}
		}

		state := atomic.LoadUint64(&s.state)
		value := state & 0xFFFFFFFF
		limit := state >> 32
		newValue := value + 1
		if newValue <= limit {
			if atomic.CompareAndSwapUint64(&s.state, state, limit<<32+newValue) {
				return nil
			} else {
				continue
			}
		} else {
			s.lock.Lock()
			s.cond.Wait()
			s.lock.Unlock()
		}
	}
	panic("unreachable")
}

func (s *semaphore) Release() {
	for {
		state := atomic.LoadUint64(&s.state)
		value := state & 0xFFFFFFFF
		if value == 0 {
			panic("Release without acquire")
		}
		newValue := value - 1
		if atomic.CompareAndSwapUint64(&s.state, state, state&0xFFFFFFFF00000000+newValue) {
			s.lock.Lock()
			s.cond.Signal()
			s.lock.Unlock()
			return
		}
	}
	panic("unreachable")
}

func (s *semaphore) SetLimit(limit int) {
	for {
		state := atomic.LoadUint64(&s.state)
		if atomic.CompareAndSwapUint64(&s.state, state, uint64(limit)<<32+state&0xFFFFFFFF) {
			return
		}
	}
	panic("unreachable")
}

func (s *semaphore) GetCount() int {
	state := atomic.LoadUint64(&s.state)
	return int(state & 0xFFFFFFFF)
}

func (s *semaphore) GetLimit() int {
	state := atomic.LoadUint64(&s.state)
	return int(state >> 32)
}
