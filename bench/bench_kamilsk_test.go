package bench

import (
	"context"
	"sync"
	"testing"

	"github.com/kamilsk/semaphore"
)

func BenchmarkKamilskSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := semaphore.New(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		rf, _ := sem.Acquire(ctx.Done())
		rf.Release()
	}
}

func BenchmarkKamilskSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := semaphore.New(100)
	ctx := context.Background()

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				rf, _ := sem.Acquire(ctx.Done())
				rf.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}

func BenchmarkKamilskSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := semaphore.New(10)
	ctx := context.Background()

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				rf, _ := sem.Acquire(ctx.Done())
				rf.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}
