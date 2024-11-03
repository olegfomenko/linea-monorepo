package parallel

import (
	"runtime"
	"sync"
)

// ExecuteJobStealing parallelizes a workload specified by a function consuming
// a channel distributing the workload. It is appropriate when each iteration
// takes an order of magnitude more time than the other functions.
//
// This is as [ExecuteChunky] but gives more freedom to the caller to initialize
// its threads.
func ExecuteFromChan(nbIterations int, work func(wg *sync.WaitGroup, ch <-chan int), numcpus ...int) {

	numcpu := runtime.GOMAXPROCS(0)
	if len(numcpus) > 0 && numcpus[0] > 0 {
		numcpu = numcpus[0]
	}

	jobChan := make(chan int, nbIterations)
	go fillChan(jobChan, nbIterations)

	// The wait group ensures that all the children goroutine have terminated
	// before we close the
	wg := &sync.WaitGroup{}
	wg.Add(nbIterations)

	// Each goroutine consumes the jobChan to
	for p := 0; p < numcpu; p++ {
		go func() {
			work(wg, jobChan)
		}()
	}

	wg.Wait()
	close(jobChan)
}
