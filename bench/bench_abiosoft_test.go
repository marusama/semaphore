package bench

import (
	"sync"
	"testing"

	"github.com/abiosoft/semaphore"
)

func BenchmarkAbiosoftSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := semaphore.New(b.N)

	for i := 0; i < b.N; i++ {
		sem.Acquire()
		sem.Release()
	}
}

func BenchmarkAbiosoftSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := semaphore.New(100)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire()
				sem.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}

func BenchmarkAbiosoftSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := semaphore.New(10)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire()
				sem.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}
