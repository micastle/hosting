package hosting

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type AppRunner interface {
	Execute()
}

type SyncAppRunner interface {
	AppRunner
}

type AsyncAppRunner interface {
	AppRunner

	SendStopSignal()
}

type BasicAsyncAppRunner struct {
	context dep.Context
	logger  logger.Logger
	done    chan os.Signal
	host    HostAsyncOperator
}

func NewBasicAsyncRunner(context dep.Context, host HostAsyncOperator) *BasicAsyncAppRunner {
	ar := &BasicAsyncAppRunner{
		context: context,
		done:    make(chan os.Signal, 1),
		host:    host,
	}
	ar.logger = context.GetLogger()
	return ar
}

func (ar *BasicAsyncAppRunner) Execute() {
	// start the host lifecycle
	ar.host.Start()

	// wait for stop signal from console
	ar.WaitForStop()

	// shut down with timeout
	err := ar.host.Shutdown(8 * time.Second)
	if err != nil {
		ar.logger.Errorw("Host shut down with failure", "last error", err)
	}
}

func (ar *BasicAsyncAppRunner) SendStopSignal() {
	ar.done <- syscall.SIGINT
}

func (ar *BasicAsyncAppRunner) WaitForStop() {

	signal.Notify(ar.done, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ar.done
		ar.logger.Debugw("Receiving server stop signal!", "Signal", sig.String())

		accept := ar.host.OnStopEvent(&StopEvent{Type: EVENT_TYPE_SIGNAL, Data: sig})
		if accept {
			ar.logger.Infow("Stop signal is accepted", "Signal", sig.String())
			break
		}

		ar.logger.Debugw("Stop signal is ignored", "Signal", sig.String())
	}
}
