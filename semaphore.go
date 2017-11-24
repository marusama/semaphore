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

type semaphore struct {
	state       uint64
	broadcastCh unsafe.Pointer
}

func New(limit int) Semaphore {
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
				return errors.New("ctx.Done()")
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
			broadcastCh := *(*chan struct{})(atomic.LoadPointer(&s.broadcastCh))

			if ctx != nil {
				select {
				case <-ctx.Done():
					return errors.New("ctx.Done()")
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
		limit := state >> 32
		if count == 0 {
			panic("Release without acquire")
		}
		newCount := count - 1
		if atomic.CompareAndSwapUint64(&s.state, state, state&0xFFFFFFFF00000000+newCount) {
			if count >= limit {
				newBroadcastCh := make(chan struct{})
				oldPtr := atomic.LoadPointer(&s.broadcastCh)
				if atomic.CompareAndSwapPointer(&s.broadcastCh, oldPtr, unsafe.Pointer(&newBroadcastCh)) {
					oldBroadcastCh := *(*chan struct{})(oldPtr)
					close(oldBroadcastCh)
				}
			}
			return
		}
	}
}

func (s *semaphore) SetLimit(limit int) {
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
