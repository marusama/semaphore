package bench

import (
	"sync"
	"testing"

	"github.com/pivotal-golang/semaphore"
)

func BenchmarkPivotalGolangSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := semaphore.New(b.N, 10000)

	for i := 0; i < b.N; i++ {
		rf, _ := sem.Acquire()
		rf.Release()
	}
}

func BenchmarkPivotalGolangSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := semaphore.New(100, 10000)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				rf, _ := sem.Acquire()
				rf.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}

func BenchmarkPivotalGolangSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := semaphore.New(10, 10000)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				rf, _ := sem.Acquire()
				rf.Release()
			}
			wg.Done()
		}()
	}

	b.ResetTimer()
	close(c) // start
	wg.Wait()
}
