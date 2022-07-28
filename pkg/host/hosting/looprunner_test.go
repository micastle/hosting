package hosting

import (
	"context"
	"testing"
	"time"
)

func Test_looprunner_basic(t *testing.T) {
	runner := NewLoopRunner(LoopRunnerSettings{
		EnableRecover:   true,
		MinLoopInterval: 500 * time.Millisecond,
		MaxStopInterval: 500 * time.Millisecond,
	})
	value := 123
	runner.Initialize(func() any { return value })

	base := 456
	result := 0
	go runner.Run(time.Duration(1)*time.Second, func(ctxt any) {
		result = base + ctxt.(int)
	})

	time.Sleep(time.Duration(500) * time.Millisecond)
	actual := result
	if actual < base+value {
		t.Errorf("loop should execute at least one iteration, actual: %d", actual)
	}

	ctxt := context.Background()
	err := runner.Stop(ctxt)
	if err != nil {
		t.Errorf("stop runner error: %v", err)
	}
}

func Test_looprunner_stop_timeout(t *testing.T) {
	runner := NewLoopRunner(LoopRunnerSettings{
		EnableRecover:   true,
		MinLoopInterval: 500 * time.Millisecond,
		MaxStopInterval: 500 * time.Millisecond,
	})
	value := 123
	runner.Initialize(func() any { return value })

	base := 456
	result := 0
	go runner.Run(time.Duration(1)*time.Second, func(ctxt any) {
		result = base + ctxt.(int)
	})

	time.Sleep(time.Duration(500) * time.Millisecond)
	actual := result
	if actual < base+value {
		t.Errorf("loop should execute at least one iteration, actual: %d", actual)
	}

	ctxt := context.Background()
	timeoutCtxt, cancel := context.WithTimeout(ctxt, 0)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := runner.Stop(timeoutCtxt)
	if err != nil {
		t.Logf("stop runner error: %v", err)
	} else {
		t.Error("stop runner didn't reach timeout error, not expected")
	}
}
