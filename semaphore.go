package semaphore

import (
	"sync"
	"sync/atomic"
)

type Semaphore interface {
	Acquire()
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

func (s *semaphore) Acquire() {
	for {
		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		limit := state >> 32
		newCount := count + 1
		if newCount <= limit {
			if atomic.CompareAndSwapUint64(&s.state, state, limit<<32+newCount) {
				return
			} else {
				continue
			}
		} else {
			s.lock.Lock()
			s.cond.Wait()
			s.lock.Unlock()
		}
	}
}

func (s *semaphore) Release() {
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
				s.lock.Lock()
				s.cond.Broadcast()
				s.lock.Unlock()
			}
			return
		}
	}
}

func (s *semaphore) SetLimit(limit int) {
	for {
		state := atomic.LoadUint64(&s.state)
		if atomic.CompareAndSwapUint64(&s.state, state, uint64(limit)<<32+state&0xFFFFFFFF) {
			s.lock.Lock()
			s.cond.Broadcast()
			s.lock.Unlock()
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
