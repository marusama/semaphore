package semaphore

import (
	"fmt"
	"testing"
	"context"
	"sync"
	"time"
)

func TestSemaphore_Acquire_Release_over_limit(t *testing.T) {
	sem := New(10)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<- c
			for j := 0; j < 100000; j++ {
				sem.Acquire(nil)
				sem.Release()
			}
			wg.Done()
		}()
	}

	close(c)	// start
	wg.Wait()

	if sem.GetCount() != 0 {
		t.Error("semaphore must have count = 0")
	}
}

func TestSemaphore_Acquire_Release_over_limit_ctx_done(t *testing.T) {
	sem := New(10)
	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<- c
			for {
				sem.Acquire(ctx)
				sem.Release()
			}
			wg.Done()
		}()
	}

	close(c)	// start
	wg.Wait()

	if sem.GetCount() != 0 {
		t.Error("semaphore must have count = 0")
	}
}

func BenchmarkSemaphore_Acquire(b *testing.B) {
	sem := New(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx)
	}

	if sem.GetCount() != sem.GetLimit() {
		b.Error(fmt.Printf("semaphore must have count = %v, but has %v", sem.GetLimit(), sem.GetCount()))
	}
}

func BenchmarkSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := New(b.N)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx)
		sem.Release()
	}

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := New(20)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<- c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil)
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

func BenchmarkSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := New(10)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<- c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil)
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
