package semaphore

import (
	"context"
	"errors"
	"sync/atomic"
	"unsafe"
)

type Semaphore interface {
	Acquire(ctx context.Context) error
	Release()
	SetLimit(limit int)
	GetLimit() int
	GetCount() int
}

var (
	ErrCtxDone = errors.New("ctx.Done()")
)

type semaphore struct {
	state       uint64
	waitCount   int64
	broadcastCh unsafe.Pointer
}

func New(limit int) Semaphore {
	if limit <= 0 {
		panic("semaphore limit must be greater than 0")
	}
	broadcastCh := make(chan struct{})
	return &semaphore{
		state:       uint64(limit) << 32,
		broadcastCh: unsafe.Pointer(&broadcastCh),
	}
}

func (s *semaphore) Acquire(ctx context.Context) error {
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
		newCount := count + 1
		if newCount <= limit {
			if atomic.CompareAndSwapUint64(&s.state, state, limit<<32+newCount) {
				return nil
			} else {
				continue
			}
		} else {
			atomic.AddInt64(&s.waitCount, 1)
			broadcastCh := *(*chan struct{})(atomic.LoadPointer(&s.broadcastCh))
			if ctx != nil {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&s.waitCount, -1)
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

func (s *semaphore) Release() {
	for {
		state := atomic.LoadUint64(&s.state)
		count := state & 0xFFFFFFFF
		if count == 0 {
			panic("semaphore release without acquire")
		}
		newCount := count - 1
		if atomic.CompareAndSwapUint64(&s.state, state, state&0xFFFFFFFF00000000+newCount) {

			for {
				waitCount := atomic.LoadInt64(&s.waitCount)
				if waitCount > 0 {
					if atomic.CompareAndSwapInt64(&s.waitCount, waitCount, waitCount - 1) {
						broadcastCh := *(*chan struct{})(atomic.LoadPointer(&s.broadcastCh))
						broadcastCh <- struct{}{}
					}
				} else {
					return
				}
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
			oldPtr := atomic.LoadPointer(&s.broadcastCh)
			if atomic.CompareAndSwapPointer(&s.broadcastCh, oldPtr, unsafe.Pointer(&newBroadcastCh)) {
				oldBroadcastCh := *(*chan struct{})(oldPtr)
				close(oldBroadcastCh)
			}
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
