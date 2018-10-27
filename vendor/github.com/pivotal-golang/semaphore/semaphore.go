package semaphore

import (
	"errors"
	"fmt"
	"sync"
)

type Resource interface {
	Release() error
}

type Semaphore interface {
	Acquire() (Resource, error)
}

type request chan struct{}

type resource struct {
	inflightRequests chan request
	sync.Mutex
	released bool
}

type semaphore struct {
	inflightRequests chan request
	pendingRequests  chan request
	maxPending       int

	schedulerLock chan struct{}
}

func New(maxInflight, maxPending int) Semaphore {
	return &semaphore{
		inflightRequests: make(chan request, maxInflight),
		pendingRequests:  make(chan request, maxPending),
		maxPending:       maxPending,
		schedulerLock:    make(chan struct{}, 1),
	}
}

func (s *semaphore) Acquire() (Resource, error) {
	return s.testableAcquire(func() {})
}

func (s *semaphore) testableAcquire(beforeScheduleLock func()) (Resource, error) {
	var newRequest request = make(chan struct{})

	select {
	case s.pendingRequests <- newRequest:
	default:
		return nil, errors.New(fmt.Sprintf("Cannot queue request, maxPending reached: %d", s.maxPending))
	}

	beforeScheduleLock()

	select {
	case s.schedulerLock <- struct{}{}:
	case <-newRequest: // Prevent deadlock when request is scheduled by concurrect call
		return &resource{inflightRequests: s.inflightRequests}, nil
	}

	defer func() {
		<-s.schedulerLock
	}()

	for {
		select {
		case nextRequest := <-s.pendingRequests:
			s.inflightRequests <- nextRequest
			close(nextRequest)
		case <-newRequest:
			return &resource{inflightRequests: s.inflightRequests}, nil
		}
	}
}

func (r *resource) Release() error {
	err := r.markReleased()
	if err != nil {
		return err
	}

	<-r.inflightRequests
	return nil
}

func (r *resource) markReleased() error {
	r.Lock()
	defer r.Unlock()
	if r.released {
		return errors.New("Resource has already been released")
	}
	r.released = true
	return nil
}
