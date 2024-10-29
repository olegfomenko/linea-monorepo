package pool

import (
	"runtime"
	"sync"
)

var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{}, runtime.GOMAXPROCS(0))
var once sync.Once

func ExecutePool(nbIterations int, work func(start, stop int)) {
	once.Do(run)

	nbIterationsPerCpus := nbIterations / runtime.GOMAXPROCS(0)
	if nbIterationsPerCpus == 0 {
		nbIterationsPerCpus = 1
	}

	start := 0
	wg := sync.WaitGroup{}

	for start < nbIterations {
		wg.Add(1)
		stop := min(nbIterations, start+nbIterationsPerCpus)
		queue <- func() {
			work(start, stop)
			wg.Done()
		}
		start += nbIterationsPerCpus
	}
	wg.Wait()
}

func ExecutePoolChunky(nbIterations int, work func(k int)) {
	once.Do(run)

	wg := sync.WaitGroup{}
	wg.Add(nbIterations)

	for i := 0; i < nbIterations; i++ {
		k := i
		queue <- func() {
			work(k)
			wg.Done()
		}
	}

	wg.Wait()
}

func run() {
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		available <- struct{}{}
	}

	go scheduler()
}

func scheduler() {
	for {
		<-available
		task := <-queue
		go func() {
			task()
			available <- struct{}{}
		}()
	}
}
