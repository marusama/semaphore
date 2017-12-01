package bench

import (
	"testing"
	"sync"

	"github.com/dropbox/godropbox/sync2"
)

func BenchmarkDropboxBoundedSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := sync2.NewBoundedSemaphore(uint(b.N))

	for i := 0; i < b.N; i++ {
		sem.Acquire()
		sem.Release()
	}
}

func BenchmarkDropboxBoundedSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := sync2.NewBoundedSemaphore(100)

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

func BenchmarkDropboxBoundedSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := sync2.NewBoundedSemaphore(10)

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

func BenchmarkDropboxUnboundedSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := sync2.NewUnboundedSemaphore(b.N)

	for i := 0; i < b.N; i++ {
		sem.Acquire()
		sem.Release()
	}
}

func BenchmarkDropboxUnboundedSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := sync2.NewUnboundedSemaphore(100)

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

func BenchmarkDropboxUnboundedSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := sync2.NewUnboundedSemaphore(10)

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
