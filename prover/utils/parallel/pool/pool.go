package pool

import (
	"runtime"
	"sync"
)

var queue chan func() = make(chan func(), runtime.GOMAXPROCS(0))
var once sync.Once

func ExecutePool(task func()) {
	once.Do(run)

	ch := make(chan struct{}, 1)
	queue <- func() {
		task()
		close(ch)
	}

	<-ch
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
		go func() {
			for {
				task := <-queue
				task()
			}
		}()
	}
}
