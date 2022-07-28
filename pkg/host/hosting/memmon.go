package hosting

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type MemoryMonitor interface {
	Start()
	Stop()
}

type DefaultMemoryMonitor struct {
	Interval time.Duration

	runner *LoopRunner
}

const memmon_MinLoopInterval = 500 * time.Millisecond
const memmon_MaxStopInterval = 500 * time.Millisecond

func newMemoryMonitor() *DefaultMemoryMonitor {
	return &DefaultMemoryMonitor{
		Interval: time.Duration(5) * time.Second,
		runner: NewLoopRunner(LoopRunnerSettings{
			EnableRecover:   true,
			MinLoopInterval: memmon_MinLoopInterval,
			MaxStopInterval: memmon_MaxStopInterval,
		}),
	}
}

func (mm *DefaultMemoryMonitor) Start() {
	mm.runner.Initialize(nil)
	go mm.runner.Run(mm.Interval, func(any) {
		mm.printMemoryStatistics()
	})
}
func (mm *DefaultMemoryMonitor) Stop() {
	context := context.Background()
	_ = mm.runner.Stop(context)
}

func (mm *DefaultMemoryMonitor) printMemoryStatistics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Memory Statistics: time - %v, liveobjs - %d, heapusage - %d, osusage - %d\n", time.Now(), ms.Mallocs-ms.Frees, ms.Alloc, ms.Sys)
}

// APIs
var singleton MemoryMonitor
var once sync.Once

func GetMemoryMonitor() MemoryMonitor {
	once.Do(func() {
		singleton = newMemoryMonitor()
	})
	return singleton
}
