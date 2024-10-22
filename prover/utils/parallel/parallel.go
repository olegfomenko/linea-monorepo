package parallel

import (
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/consensys/linea-monorepo/prover/utils"
	"log"
	"fmt"
	"strings"
	"os"
)

// ParallelCallTraces what -> data size -> batch size -> how many times
var traceMu sync.Mutex
var ParallelCallTraces map[string]map[int]map[int]int

func AddParallelCallTrace(dataSize int, batchSize int, stackSkip int) {
	traceMu.Lock()
	defer traceMu.Unlock()

	pc, _, _, ok := runtime.Caller(stackSkip)
	if !ok {
		log.Println("Could not get caller information")
		return
	}
	caller, _ := strings.CutPrefix(runtime.FuncForPC(pc).Name(), "github.com/consensys/linea-monorepo/")

	if ParallelCallTraces == nil {
		ParallelCallTraces = make(map[string]map[int]map[int]int)
	}

	if ParallelCallTraces[caller] == nil {
		ParallelCallTraces[caller] = make(map[int]map[int]int)
	}

	if ParallelCallTraces[caller][dataSize] == nil {
		ParallelCallTraces[caller][dataSize] = make(map[int]int)
	}

	ParallelCallTraces[caller][dataSize][batchSize]++
}

func WriteParallelCallTraces() {
	homeDir := os.Getenv("HOME")
	file, err := os.Create(fmt.Sprintf("%s/parallel_call_traces.csv", homeDir))
	if err != nil {
		log.Println("Could not create file to trace parallel calls")
		return
	}
	defer file.Close()

	traceMu.Lock()
	defer traceMu.Unlock()

	_, err = file.WriteString("what data_size batch_size number_of_times\n")
	if err != nil {
		log.Println("Could not write to file to trace parallel calls")
		return
	}

	for caller, dataSizes := range ParallelCallTraces {
		for dataSize, batchSizes := range dataSizes {
			for batchSize, count := range batchSizes {
				_, err := file.WriteString(fmt.Sprintf("%s %d %d %d\n", caller, dataSize, batchSize, count))
				if err != nil {
					log.Println("Could not write to file to trace parallel calls")
					return
				}
			}
		}
	}
}

// Execute process in parallel the work function
func Execute(nbIterations int, work func(int, int), maxCpus ...int) {

	nbTasks := runtime.GOMAXPROCS(0)
	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
	}
	nbIterationsPerCpus := nbIterations / nbTasks

	// more CPUs than tasks: a CPU will work on exactly one iteration
	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
		nbTasks = nbIterations
	}

	var wg sync.WaitGroup

	extraTasks := nbIterations - (nbTasks * nbIterationsPerCpus)
	extraTasksOffset := 0

	var (
		panicTrace []byte
		panicMsg   any
		panicOnce  = &sync.Once{}
	)

	AddParallelCallTrace(nbIterations, nbIterationsPerCpus, 2)
	for i := 0; i < nbTasks; i++ {
		wg.Add(1)
		_start := i*nbIterationsPerCpus + extraTasksOffset
		_end := _start + nbIterationsPerCpus
		if extraTasks > 0 {
			_end++
			extraTasks--
			extraTasksOffset++
		}

		go func() {
			// In case the subtask panics, we recover so that we can repanic in
			// the main goroutine. Simplifying the process of tracing back the
			// error and allowing to test the panics.
			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() {
						panicMsg = r
						panicTrace = debug.Stack()
					})
				}

				wg.Done()
			}()

			work(_start, _end)
		}()
	}

	wg.Wait()

	if len(panicTrace) > 0 {
		utils.Panic("Had a panic: %v\nStack: %v\n", panicMsg, string(panicTrace))
	}
}
