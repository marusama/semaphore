package semaphore

import "runtime"

var def = New(runtime.GOMAXPROCS(0))

// Acquire tries to reduce the number of available slots of the default semaphore for 1.
// The operation can be canceled using deadline channel. In this case,
// it returns an appropriate error.
func Acquire(deadline <-chan struct{}) (ReleaseFunc, error) {
	return def.Acquire(deadline)
}

// Capacity returns a capacity of the default semaphore.
func Capacity() int {
	return def.Capacity()
}

// Occupied returns a current number of occupied slots of the default semaphore.
func Occupied() int {
	return def.Occupied()
}

// Release releases the previously occupied slot of the default semaphore.
func Release() error {
	return def.Release()
}

// Signal returns a channel to send to it release function only if Acquire is successful.
// In any case, the channel will be closed.
func Signal(deadline <-chan struct{}) <-chan ReleaseFunc {
	return def.Signal(deadline)
}
