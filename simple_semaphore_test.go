package semaphore

import (
	"testing"
	"sync"
)

//func BenchmarkSimpleSemaphore_Acquire(b *testing.B) {
//	sem := NewSimple(b.N)
//
//	for i := 0; i < b.N; i++ {
//		sem.Acquire()
//	}
//
//	if sem.GetCount() != sem.GetLimit() {
//		b.Error(fmt.Printf("semaphore must have count = %v, but has %v", sem.GetLimit(), sem.GetCount()))
//	}
//}

func BenchmarkSimpleSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := NewSimple(b.N)

	for i := 0; i < b.N; i++ {
		sem.Acquire()
		sem.Release()
	}

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSimpleSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := NewSimple(100)

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

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

func BenchmarkSimpleSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := NewSimple(10)

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

	if sem.GetCount() != 0 {
		b.Error("semaphore must have count = 0")
	}
}

