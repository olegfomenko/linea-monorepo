package pool

import (
	"runtime"
	"sync"
)

var nbTasks = 2 * runtime.GOMAXPROCS(0)
var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{}, nbTasks)
var once sync.Once

func ExecutePool(nbIterations int, work func(start, stop int)) {
	once.Do(run)

	nbIterationsPerCpus := nbIterations / nbTasks

	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
	}

	start := 0
	wg := sync.WaitGroup{}

	for start < nbIterations {
		wg.Add(1)
		_start := start
		_stop := min(nbIterations, _start+nbIterationsPerCpus)

		queue <- func() {
			work(_start, _stop)
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
	for i := 0; i < nbTasks; i++ {
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
