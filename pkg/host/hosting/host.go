package hosting

import (
	"context"
	"fmt"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type RunningMode uint8

const (
	Debug RunningMode = iota
	Release
)

func (rm RunningMode) String() string {
	switch rm {
	case Debug:
		return "Debug"
	case Release:
		return "Release"
	default:
		return fmt.Sprintf("RunningMode(%d)", rm)
	}
}

type HostSettings interface {
	SetName(name string)
	SetRunningMode(RunningMode)
	EnableMemoryStatistics(enable bool)
}

type DefaultHostSettings struct {
	context *HostBuilderContext
}

func NewDefaultHostSettings(context *HostBuilderContext) *DefaultHostSettings {
	return &DefaultHostSettings{
		context: context,
	}
}

func (hs *DefaultHostSettings) SetName(name string) {
	hs.context.HostName = name
}
func (hs *DefaultHostSettings) SetRunningMode(mode RunningMode) {
	hs.context.RunningMode = mode
}
func (hs *DefaultHostSettings) EnableMemoryStatistics(enable bool) {
	hs.context.EnableMemoryStatistics = enable
}

type HostAsyncOperator interface {
	Start()
	OnStopEvent(*StopEvent) bool
	Shutdown(timeout time.Duration) error
}

type Host interface {
	GetRawContext() context.Context
	GetName() string
	GetLogger() logger.Logger
	GetConfiguration() Configuration
	GetComponentProvider() dep.ComponentProvider

	GetContext() dep.HostContext
	GetServices() map[string]Service

	Run()
}

type GenericHost interface {
	HostAsyncOperator
	SyncAppRunner
	Host
}

type DefaultGenericHost struct {
	hostContext *DefaultHostContext
	provider    dep.ComponentProvider
	LogFactory  logger.LoggerFactory
	Logger      logger.Logger
}

func NewDefaultGenericHost(ctxt *DefaultHostContext) *DefaultGenericHost {
	logFactory := dep.GetComponent[logger.LoggerFactory](ctxt.ComponentProvider)
	host := &DefaultGenericHost{
		hostContext: ctxt,
		provider:    ctxt.ComponentProvider,
		LogFactory:  logFactory,
	}
	host.Logger = logFactory.GetLogger(dep.GetDefaultLoggerNameForComponent(host))
	return host
}

func (h *DefaultGenericHost) GetRawContext() context.Context {
	return h.hostContext.RawContext
}
func (h *DefaultGenericHost) GetContext() dep.HostContext {
	return h.hostContext
}
func (h *DefaultGenericHost) GetName() string {
	return h.hostContext.Name()
}
func (h *DefaultGenericHost) GetLogger() logger.Logger {
	return dep.GetComponent[logger.LoggerFactory](h.provider).GetDefaultLogger()
}
func (h *DefaultGenericHost) GetConfiguration() Configuration {
	return h.hostContext.Configuration
}

func (h *DefaultGenericHost) GetComponentProvider() dep.ComponentProvider {
	return h.provider
}

func (h *DefaultGenericHost) GetServices() map[string]Service {
	return h.hostContext.Services
}

func (h *DefaultGenericHost) startMemoryMonitor() {
	if h.hostContext.builderContext.EnableMemoryStatistics {
		mon := GetMemoryMonitor()
		mon.Start()
	}
}
func (h *DefaultGenericHost) stopMemoryMonitor() {
	if h.hostContext.builderContext.EnableMemoryStatistics {
		mon := GetMemoryMonitor()
		mon.Stop()
	}
}

func (h *DefaultGenericHost) Run() {
	// start memory statistics
	h.startMemoryMonitor()
	defer h.stopMemoryMonitor()

	// execute registered app runner
	runner := dep.GetComponent[AppRunner](h.provider)
	runner.Execute()
}

func (h *DefaultGenericHost) Start() {
	h.Logger.Infow("Hosted services starting")

	for name, service := range h.hostContext.Services {
		h.Logger.Debug("starting service: ", name)
		go service.Run()
	}

	// after all services started
	h.hostContext.Lifecycle.OnAppStarted(h.hostContext)

	h.Logger.Debug("Hosted services started Successfully!")
}

func (h *DefaultGenericHost) OnStopEvent(event *StopEvent) bool {
	return h.hostContext.Lifecycle.OnStopEvent(h.hostContext, event)
}
func (h *DefaultGenericHost) StopService(name string, service Service, ctxt context.Context) error {
	panicErr := error(nil)
	err := error(nil)

	func() {
		defer func() {
			if r := recover(); r != nil {
				h.Logger.Errorw("Panic stopping service", "Name", name, "Panic", r)
				panicErr = fmt.Errorf("%v", r)
			}
		}()

		err = service.Stop(ctxt)
		if err != nil {
			h.Logger.Errorw("stopping service failed", "service", name, "error", err)
		} else {
			h.Logger.Infow("service stopped complete", "service", name)
		}
	}()

	if panicErr != nil {
		return panicErr
	}

	return err
}

func (h *DefaultGenericHost) StopServiceWithTimeout(name string, service Service, timeout time.Duration) error {
	ctxt, cancel := context.WithTimeout(h.hostContext.RawContext, timeout)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := h.StopService(name, service, ctxt)

	return err
}

func (h *DefaultGenericHost) Shutdown(timeout time.Duration) error {
	// before shuting down
	h.hostContext.Lifecycle.OnAppStopping(h.hostContext)

	// shut down services registered on the host
	lastError := error(nil)
	serviceCount := len(h.hostContext.Services)
	done := make(chan error, serviceCount)
	for name, service := range h.hostContext.Services {
		h.Logger.Debug("shutting down service: ", name)
		go func(name string, srv Service) {
			err := h.StopServiceWithTimeout(name, srv, timeout)
			done <- err
		}(name, service)
	}
	// wait for all complete
	for i := 0; i < serviceCount; i++ {
		err := <-done
		if err != nil {
			lastError = err
		}
	}

	// after shuting down
	h.hostContext.Lifecycle.OnAppStopped(h.hostContext)

	h.Logger.Info("Hosted services were shut down complete")
	return lastError
}

func (h *DefaultGenericHost) Execute() {
	logger := h.Logger
	logger.Debug("Application start running")

	if len(h.hostContext.Services) > 1 {
		h.Logger.Fatalf("Application run in sync mode will execute only one registered service at a time! registered: %v", len(h.hostContext.Services))
	}

	for name, service := range h.hostContext.Services {
		h.Logger.Debugf("Running service: %s", name)
		service.Run()
	}

	logger.Debug("Application run complete")
}

// factory method to create host from context
func NewHostFromContext(context *DefaultHostContext) *DefaultGenericHost {
	return NewDefaultGenericHost(context)
}
