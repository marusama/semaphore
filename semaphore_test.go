package semaphore

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	xsync "golang.org/x/sync/semaphore"
)

func checkLimit(t *testing.T, sem Semaphore, expected int) {
	limit := sem.GetLimit()
	if limit != expected {
		t.Error("semaphore must have limit = ", expected, ", but has ", limit)
	}
}

func checkCount(t *testing.T, sem Semaphore, expected int) {
	count := sem.GetCount()
	if count != expected {
		t.Error("semaphore must have count = ", expected, ", but has ", count)
	}
}

func checkLimitAndCount(t *testing.T, sem Semaphore, expectedLimit, expectedCount int) {
	checkLimit(t, sem, expectedLimit)
	checkCount(t, sem, expectedCount)
}

func TestNew(t *testing.T) {
	sem := New(1)
	checkLimitAndCount(t, sem, 1, 0)
}

func TestNew_zero_limit_panic_expected(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Panic expected")
		}
	}()
	_ = New(0)
}

func TestNew_negative_limit_panic_expected(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Panic expected")
		}
	}()
	_ = New(-1)
}

func TestSemaphore_Acquire(t *testing.T) {
	sem := New(1)

	err := sem.Acquire(nil)

	if err != nil {
		t.Error("Error returned:", err.Error())
	}
	checkLimitAndCount(t, sem, 1, 1)
}

func TestSemaphore_Acquire_with_ctx(t *testing.T) {
	sem := New(1)

	err := sem.Acquire(context.Background())

	if err != nil {
		t.Error("Error returned:", err.Error())
	}
	checkLimitAndCount(t, sem, 1, 1)
}

func TestSemaphore_Acquire_ctx_done(t *testing.T) {
	sem := New(1)
	ctx_done, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	cancel() // make ctx.Done()

	err := sem.Acquire(ctx_done)

	if err != ErrCtxDone {
		t.Error("Error is not ErrCtxDone")
	}
	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Release(t *testing.T) {
	sem := New(1)

	sem.Acquire(nil)
	sem.Release()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Release_with_ctx(t *testing.T) {
	sem := New(1)

	sem.Acquire(context.Background())
	sem.Release()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Release_without_Acquire(t *testing.T) {
	sem := New(1)

	defer func() {
		if recover() == nil {
			t.Error("Panic expected")
		}
	}()
	sem.Release()
}

func TestSemaphore_Acquire_Release_2_times(t *testing.T) {
	sem := New(2)
	checkLimitAndCount(t, sem, 2, 0)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 2, 1)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 2, 2)

	sem.Release()
	checkLimitAndCount(t, sem, 2, 1)

	sem.Release()
	checkLimitAndCount(t, sem, 2, 0)
}

func TestSemaphore_Acquire_Release_under_limit(t *testing.T) {
	sem := New(100)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			err := sem.Acquire(nil)
			if err != nil {
				panic(err)
			}
			sem.Release()
			wg.Done()
		}()
	}

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 100, 0)
}

func TestSemaphore_Acquire_Release_under_limit_ctx_done(t *testing.T) {
	sem := New(10)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for {
				err := sem.Acquire(ctx)
				if err != nil {
					if err == ErrCtxDone {
						break
					}
					panic(err)
				}
				sem.Release()
			}
			wg.Done()
		}()
	}

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 10, 0)
}

func TestSemaphore_Acquire_Release_over_limit(t *testing.T) {
	sem := New(1)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < 1000; j++ {
				err := sem.Acquire(nil)
				if err != nil {
					panic(err)
				}
				sem.Release()
			}
			wg.Done()
		}()
	}

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Acquire_Release_over_limit_ctx_done(t *testing.T) {
	sem := New(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for {
				err := sem.Acquire(ctx)
				if err != nil {
					if err == ErrCtxDone {
						break
					}
					panic(err)
				}
				sem.Release()
			}
			wg.Done()
		}()
	}

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_SetLimit(t *testing.T) {
	sem := New(1)
	checkLimitAndCount(t, sem, 1, 0)

	sem.SetLimit(2)
	checkLimitAndCount(t, sem, 2, 0)

	sem.SetLimit(1)
	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_SetLimit_zero_limit_panic_expected(t *testing.T) {
	sem := New(1)
	checkLimitAndCount(t, sem, 1, 0)

	defer func() {
		if recover() == nil {
			t.Error("Panic expected")
		}
	}()
	sem.SetLimit(0)
}

func TestSemaphore_SetLimit_negative_limit_panic_expected(t *testing.T) {
	sem := New(1)
	checkLimitAndCount(t, sem, 1, 0)

	defer func() {
		if recover() == nil {
			t.Error("Panic expected")
		}
	}()
	sem.SetLimit(-1)
}

func TestSemaphore_SetLimit_increase_limit(t *testing.T) {
	sem := New(1)
	checkLimitAndCount(t, sem, 1, 0)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 1, 1)

	sem.SetLimit(2)
	checkLimitAndCount(t, sem, 2, 1)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 2, 2)

	sem.Release()
	checkLimitAndCount(t, sem, 2, 1)

	sem.Release()
	checkLimitAndCount(t, sem, 2, 0)
}

func TestSemaphore_SetLimit_decrease_limit(t *testing.T) {
	sem := New(2)
	checkLimitAndCount(t, sem, 2, 0)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 2, 1)

	sem.Acquire(nil)
	checkLimitAndCount(t, sem, 2, 2)

	sem.SetLimit(1)
	checkLimitAndCount(t, sem, 1, 2)

	sem.Release()
	checkLimitAndCount(t, sem, 1, 1)

	sem.Release()
	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_SetLimit_increase_broadcast(t *testing.T) {
	getWGs := func(cnt int) []*sync.WaitGroup {
		wgs := make([]*sync.WaitGroup, cnt)
		for i := range wgs {
			wgs[i] = &sync.WaitGroup{}
			wgs[i].Add(1)
		}
		return wgs
	}

	sem := New(1)
	sem.Acquire(nil)

	innerWGs := getWGs(2)
	outerWGs := getWGs(2)

	go func() {
		outerWGs[0].Wait()

		// here we a trying to acquire over limit
		checkLimitAndCount(t, sem, 1, 1)
		sem.Acquire(nil)

		innerWGs[0].Done()
		outerWGs[1].Wait()

		sem.Release()

		innerWGs[1].Done()
	}()

	checkLimitAndCount(t, sem, 1, 1)
	outerWGs[0].Done()

	time.Sleep(100 * time.Millisecond)

	// increase limit so inner goroutine can acquire semaphore
	sem.SetLimit(2)
	checkLimitAndCount(t, sem, 2, 1)

	innerWGs[0].Wait()

	checkLimitAndCount(t, sem, 2, 2)

	outerWGs[1].Done()
	innerWGs[1].Wait()

	checkLimitAndCount(t, sem, 2, 1)

	sem.Release()
	checkLimitAndCount(t, sem, 2, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_under_limit(t *testing.T) {
	sem := New(100)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < 10000; j++ {
				err := sem.Acquire(nil)
				if err != nil {
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	c2 := make(chan struct{})
	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-c2:
				sem.SetLimit(100)
				wg2.Done()
				return
			default:
			}
			newLimit := rand.Intn(50) + 50 // range [50, 99]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	close(c2) // stop 'set limit' goroutine
	wg2.Wait()

	checkLimitAndCount(t, sem, 100, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_under_limit_ctx_done(t *testing.T) {
	sem := New(100)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for {
				err := sem.Acquire(ctx)
				if err != nil {
					if err == ErrCtxDone {
						break
					}
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-ctx.Done():
				sem.SetLimit(100)
				wg.Done()
				return
			default:
			}
			newLimit := rand.Intn(50) + 50 // range [50, 99]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 100, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_over_limit(t *testing.T) {
	sem := New(1)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < 10000; j++ {
				err := sem.Acquire(nil)
				if err != nil {
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	c2 := make(chan struct{})
	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-c2:
				sem.SetLimit(1)
				wg2.Done()
				return
			default:
			}
			newLimit := rand.Intn(50) + 1 // range [1, 50]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	close(c2) // stop 'set limit' goroutine
	wg2.Wait()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_over_limit_ctx_done(t *testing.T) {
	sem := New(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for {
				err := sem.Acquire(ctx)
				if err != nil {
					if err == ErrCtxDone {
						break
					}
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-ctx.Done():
				sem.SetLimit(1)
				wg.Done()
				return
			default:
			}
			newLimit := rand.Intn(50) + 1 // range [1, 50]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_random_limit(t *testing.T) {
	sem := New(1)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < 10000; j++ {
				err := sem.Acquire(nil)
				if err != nil {
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	c2 := make(chan struct{})
	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-c2:
				sem.SetLimit(1)
				wg2.Done()
				return
			default:
			}
			newLimit := rand.Intn(200) + 1 // range [1, 200]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	close(c2) // stop 'set limit' goroutine
	wg2.Wait()

	checkLimitAndCount(t, sem, 1, 0)
}

func TestSemaphore_Acquire_Release_SetLimit_random_limit_ctx_done(t *testing.T) {
	sem := New(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for {
				err := sem.Acquire(ctx)
				if err != nil {
					if err == ErrCtxDone {
						break
					}
					panic(err)
				}
				runtime.Gosched()
				sem.Release()
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		<-c
		for {
			select {
			case <-ctx.Done():
				sem.SetLimit(1)
				wg.Done()
				return
			default:
			}
			newLimit := rand.Intn(200) + 1 // range [1, 200]
			sem.SetLimit(newLimit)
			runtime.Gosched()
		}

	}()

	close(c) // start
	wg.Wait()

	checkLimitAndCount(t, sem, 1, 0)
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
	sem := New(100)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil)
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

func BenchmarkSemaphore_Acquire_Release_over_limit(b *testing.B) {
	sem := New(10)

	c := make(chan struct{})
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-c
			for j := 0; j < b.N; j++ {
				sem.Acquire(nil)
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

func BenchmarkXSyncSemaphore_Acquire_Release_under_limit_simple(b *testing.B) {
	sem := xsync.NewWeighted(int64(b.N))
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		sem.Acquire(ctx, 1)
		sem.Release(1)
	}
}

func BenchmarkXSyncSemaphore_Acquire_Release_under_limit(b *testing.B) {
	sem := xsync.NewWeighted(100)
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
	sem := xsync.NewWeighted(10)
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