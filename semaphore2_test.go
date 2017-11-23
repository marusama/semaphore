package semaphore

import (
	"fmt"
	"testing"
	"context"
	"sync"
)

func BenchmarkSemaphore2_Acquire(b *testing.B) {
	sem := New2(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx)
	}

	if sem.GetCount() != sem.GetLimit() {
		b.Error(fmt.Printf("semaphore must have count = %v, but has %v", sem.GetLimit(), sem.GetCount()))
	}
}

func BenchmarkSemaphore2_Acquire_Release_under_limit(b *testing.B) {
	sem := New2(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx)
		sem.Release()
	}

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSemaphore2_Acquire_Release_over_limit(b *testing.B) {
	b.ReportAllocs()
	sem := New2(10)
	ctx := context.Background()

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<- c
			for j := 0; j < b.N; j++ {
				sem.Acquire(ctx)
				sem.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c)	// start
	wg.Wait()

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}
