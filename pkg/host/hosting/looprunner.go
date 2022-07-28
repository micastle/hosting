package hosting

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type LoopRunnerSettings struct {
	EnableRecover   bool
	MinLoopInterval time.Duration
	MaxStopInterval time.Duration
}
type LoopRunner struct {
	settings   LoopRunnerSettings
	ctxtInitor func() any

	Done    chan bool
	stopped bool
}

func NewLoopRunner(settings LoopRunnerSettings) *LoopRunner {
	return &LoopRunner{
		Done:     make(chan bool, 1),
		stopped:  false,
		settings: settings,
	}
}
func (lr *LoopRunner) Initialize(ctxtInitor func() any) {
	lr.ctxtInitor = ctxtInitor
}
func (lr *LoopRunner) Run(interval time.Duration, loopAction func(any)) {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	var context any
	if lr.ctxtInitor != nil {
		context = lr.ctxtInitor()
	}

	for {
		func() {
			start := time.Now()
			defer func() {
				if lr.settings.EnableRecover {
					if r := recover(); r != nil {
						fmt.Printf("panic from memory monitor: %v\n", r)
					}
				}

				eclipse := time.Since(start)
				actual := interval - eclipse
				if actual < lr.settings.MinLoopInterval {
					actual = lr.settings.MinLoopInterval
				}
				timer.Reset(actual)
			}()

			loopAction(context)
		}()

		select {
		case <-lr.Done:
			lr.stopped = true
			return
		case <-timer.C:
			continue
		}
	}
}

func (lr *LoopRunner) Stop(ctx context.Context) error {
	// stop the loop
	lr.Done <- true

	// wait for loop to stop
	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > lr.settings.MaxStopInterval {
			pollIntervalBase = lr.settings.MaxStopInterval
		}
		return interval
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if lr.stopped {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
