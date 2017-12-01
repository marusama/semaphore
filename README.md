semaphore
=========

Fast golang semaphore based on CAS

* allows weighted acquire/release;
* supports cancellation via context;
* allows change semaphore limit after creation;
* faster than channels based semaphores.

### Usage
Initiate
```go
import "github.com/marusama/semaphore"
...
sem := semaphore.New(5) // new semaphore with limit=5
```
Acquire
```go
sem.Acquire(ctx, n)     // acquire with context
sem.TryAcquire(n)       // try acquire without blocking 
```
Acquire with timeout
```go
ctx := context.WithTimeout(context.Background(), time.Second)
sem.Acquire(ctx, n)
``` 
Release
```go
sem.Release(n)          // release
```
Change semaphore limit
```go
sem.SetLimit(new_limit) // set new semaphore limit
```


### Some benchmarks
Benches run on MacBook Pro (early 2015) with 2,7GHz Core i5 cpu and 16GB DDR3 ram:
```text
// github.com/marusama/semaphore
BenchmarkSemaphore_Acquire_Release_under_limit_simple-4                   	50000000	        31.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkSemaphore_Acquire_Release_under_limit-4                          	 1000000	      1383 ns/op	       0 B/op	       0 allocs/op
BenchmarkSemaphore_Acquire_Release_over_limit-4                           	  100000	     13468 ns/op	      24 B/op	       0 allocs/op

// some other implementations
BenchmarkAbiosoftSemaphore_Acquire_Release_under_limit_simple-4           	 5000000	       269 ns/op	       0 B/op	       0 allocs/op
BenchmarkAbiosoftSemaphore_Acquire_Release_under_limit-4                  	  300000	      5602 ns/op	       0 B/op	       0 allocs/op
BenchmarkAbiosoftSemaphore_Acquire_Release_over_limit-4                   	   30000	     54090 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxBoundedSemaphore_Acquire_Release_under_limit_simple-4     	20000000	        99.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxBoundedSemaphore_Acquire_Release_under_limit-4            	 1000000	      1343 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxBoundedSemaphore_Acquire_Release_over_limit-4             	  100000	     35735 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxUnboundedSemaphore_Acquire_Release_under_limit_simple-4   	30000000	        56.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxUnboundedSemaphore_Acquire_Release_under_limit-4          	  500000	      2871 ns/op	       0 B/op	       0 allocs/op
BenchmarkDropboxUnboundedSemaphore_Acquire_Release_over_limit-4           	   30000	     41089 ns/op	       0 B/op	       0 allocs/op
BenchmarkKamilskSemaphore_Acquire_Release_under_limit_simple-4            	10000000	       170 ns/op	      16 B/op	       1 allocs/op
BenchmarkKamilskSemaphore_Acquire_Release_under_limit-4                   	  500000	      3023 ns/op	     160 B/op	      10 allocs/op
BenchmarkKamilskSemaphore_Acquire_Release_over_limit-4                    	   20000	     67687 ns/op	    1600 B/op	     100 allocs/op
BenchmarkPivotalGolangSemaphore_Acquire_Release_under_limit_simple-4      	 1000000	      1006 ns/op	     136 B/op	       2 allocs/op
BenchmarkPivotalGolangSemaphore_Acquire_Release_under_limit-4             	  100000	     11837 ns/op	    1280 B/op	      20 allocs/op
BenchmarkPivotalGolangSemaphore_Acquire_Release_over_limit-4              	   10000	    128890 ns/op	   12800 B/op	     200 allocs/op
BenchmarkXSyncSemaphore_Acquire_Release_under_limit_simple-4              	30000000	        51.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncSemaphore_Acquire_Release_under_limit-4                     	  500000	      2655 ns/op	       0 B/op	       0 allocs/op
BenchmarkXSyncSemaphore_Acquire_Release_over_limit-4                      	   20000	    100004 ns/op	   15991 B/op	     299 allocs/op
```
You can run your own benchmarks, just checkout `benchmarks` branch.
