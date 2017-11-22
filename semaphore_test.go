package semaphore

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
	"fmt"
)

func TestSemaphore_Acquire(t *testing.T) {

	println(runtime.GOMAXPROCS(0))

	s := New(1)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			s.Acquire(context.Background())
			time.Sleep(10 * time.Minute)
			wg.Done()
		}()
	}

	wg.Wait()

	println("Done")
}

func BenchmarkSemaphore_Acquire(b *testing.B) {
	sem := New(b.N)

	for i := 0; i < b.N; i++ {
		_ = sem.Acquire(nil)
	}

	if sem.GetCount() != sem.GetLimit() {
		b.Error(fmt.Printf("semaphore must have count = %v, but has %v", sem.GetLimit(), sem.GetCount()))
	}
}

func BenchmarkSemaphore_Acquire_Release(b *testing.B) {
	sem := New(b.N)

	for i := 0; i < b.N; i++ {
		_ = sem.Acquire(nil)
		sem.Release()
	}

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}
