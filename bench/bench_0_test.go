package bench

import (
	"context"
	"sync"
	"testing"

	"github.com/marusama/semaphore/v2"
)

func BenchmarkSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := semaphore.New(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx, 1)
		sem.Release(1)
	}

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := semaphore.New(100)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil, 1)
				sem.Release(1)
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := semaphore.New(10)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil, 1)
				sem.Release(1)
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}
