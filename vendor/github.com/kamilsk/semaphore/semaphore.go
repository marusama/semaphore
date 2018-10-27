// Copyright (c) 2017 OctoLab. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package semaphore provides an implementation of Semaphore pattern
// with timeout of lock/unlock operations based on channels.
package semaphore // import "github.com/kamilsk/semaphore"

import "errors"

// HealthChecker defines helpful methods related with semaphore status.
type HealthChecker interface {
	// Capacity returns a capacity of a semaphore.
	// It must be safe to call Capacity concurrently on a single semaphore.
	Capacity() int
	// Occupied returns a current number of occupied slots.
	// It must be safe to call Occupied concurrently on a single semaphore.
	Occupied() int
}

// Releaser defines a method to release the previously occupied semaphore.
type Releaser interface {
	// Release releases the previously occupied slot.
	// If no places were occupied, then it returns an appropriate error.
	// It must be safe to call Release concurrently on a single semaphore.
	Release() error
}

// ReleaseFunc tells a semaphore to release the previously occupied slot
// and ignore an error if it occurs.
type ReleaseFunc func()

// Release calls f().
func (f ReleaseFunc) Release() error {
	f()
	return nil
}

// Semaphore provides the functionality of the same named pattern.
type Semaphore interface {
	HealthChecker
	Releaser

	// Acquire tries to reduce the number of available slots for 1.
	// The operation can be canceled using context. In this case,
	// it returns an appropriate error.
	// It must be safe to call Acquire concurrently on a single semaphore.
	Acquire(deadline <-chan struct{}) (ReleaseFunc, error)
	// Signal returns a channel to send to it release function
	// only if Acquire is successful. In any case, the channel will be closed.
	Signal(deadline <-chan struct{}) <-chan ReleaseFunc
}

// New constructs a new thread-safe Semaphore with the given capacity.
func New(capacity int) Semaphore {
	return make(semaphore, capacity)
}

// IsEmpty checks if passed error is related to call Release on empty semaphore.
func IsEmpty(err error) bool {
	return err == errEmpty
}

// IsTimeout checks if passed error is related to call Acquire on full semaphore.
func IsTimeout(err error) bool {
	return err == errTimeout
}

var (
	nothing ReleaseFunc = func() {}

	errEmpty   = errors.New("semaphore is empty")
	errTimeout = errors.New("operation timeout")
)

type semaphore chan struct{}

func (sem semaphore) Acquire(deadline <-chan struct{}) (ReleaseFunc, error) {
	select {
	case sem <- struct{}{}:
		return func() { _ = sem.Release() }, nil //nolint: gas
	case <-deadline:
		return nothing, errTimeout
	}
}

func (sem semaphore) Capacity() int {
	return cap(sem)
}

func (sem semaphore) Occupied() int {
	return len(sem)
}

func (sem semaphore) Release() error {
	select {
	case <-sem:
		return nil
	default:
		return errEmpty
	}
}

func (sem semaphore) Signal(deadline <-chan struct{}) <-chan ReleaseFunc {
	ch := make(chan ReleaseFunc, 1)
	go func() {
		if release, err := sem.Acquire(deadline); err == nil {
			ch <- release
		}
		close(ch)
	}()
	return ch
}
