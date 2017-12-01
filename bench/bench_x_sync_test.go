package bench

import (
	"testing"
	"sync"
	"context"

	"golang.org/x/sync/semaphore"
)

func BenchmarkXSyncSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := semaphore.NewWeighted(int64(b.N))
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx, 1)
		sem.Release(1)
	}
}

func BenchmarkXSyncSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := semaphore.NewWeighted(100)
	ctx := context.Background()

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(ctx, 1)
				sem.Release(1)
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}

func BenchmarkXSyncSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := semaphore.NewWeighted(10)
	ctx := context.Background()

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(ctx, 1)
				sem.Release(1)
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}

