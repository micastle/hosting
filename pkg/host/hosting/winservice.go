// win service works for windows platform only

//go:build windows
// +build windows

package hosting

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

func registerPlatformComponents(components dep.ComponentCollection) {
	components.RegisterTransientForType(NewWinSVC, types.Of(new(WinSVC)))
}

//
// WinServiceRunner, a type of AsyncAppRunner
//
type WinServiceRunner interface {
	AsyncAppRunner
}

type DefaultWinServiceRunner struct {
	context      dep.ContextEx
	config       *WinSvcAppRunnerConfig
	logger       logger.Logger
	host         HostAsyncOperator
	winSvc       WinSVC
	shuttingDown bool
	stopped      bool
}

type WinSvcAppRunnerConfig struct {
	ServiceName                string
	ShutdownTimeoutInSec       int
	ShutdownCheckIntervalInSec int
}

type WinSvcRunnerCheckStoppingProcessor FunctionProcessor

func NewWinServiceRunner(context dep.Context, host HostAsyncOperator, winSvc WinSVC, config *WinSvcAppRunnerConfig, statusChecker FunctionProcessor) *DefaultWinServiceRunner {
	wsr := &DefaultWinServiceRunner{
		context:      context.(dep.ContextEx),
		config:       config,
		logger:       context.GetLogger(),
		host:         host,
		winSvc:       winSvc,
		shuttingDown: false,
		stopped:      false,
	}
	loopCtxt := NewLoopContextFromWinServiceRunner(wsr.context)
	wsr.winSvc.Initialize(loopCtxt, config.ServiceName)
	wsr.winSvc.SetStopSignalCallback(func(sig os.Signal) bool {
		return wsr.host.OnStopEvent(&StopEvent{Type: EVENT_TYPE_SIGNAL, Data: sig})
	})
	wsr.winSvc.SetStartStoppingCallback(func(reason StopReason) {
		wsr.host.OnStopEvent(&StopEvent{Type: EVENT_TYPE_WINSVC, Data: reason})
		wsr.onStartShuttingdown(reason)
	})

	statusChecker.Initialize(
		context,
		types.Of(new(WinSvcRunnerCheckStoppingProcessor)),
		func(context dep.Context, interfaceType types.DataType, scopeCtxt ScopeContext) {
			wsr.checkStoppingStatus(scopeCtxt)
		},
	)
	wsr.winSvc.SetStoppingStatusChecker(
		&StatusCheckerConfig{
			CheckIntervalInMS: uint32(config.ShutdownCheckIntervalInSec) * 1000,
			WaitTimeOutInMS:   uint32(config.ShutdownTimeoutInSec+1) * 1000,
		},
		statusChecker,
	)
	return wsr
}
func (wsr *DefaultWinServiceRunner) onStartShuttingdown(reason StopReason) {
	if !wsr.shuttingDown {
		wsr.shuttingDown = true
		go wsr.shutdownHost(wsr.config.ShutdownTimeoutInSec)
	} else {
		wsr.logger.Info("shutting down services ongoing...")
	}
}
func (wsr *DefaultWinServiceRunner) checkStoppingStatus(context ScopeContext) {
	if wsr.stopped {
		wsr.logger.Info("shutting down services completed")
		context.ExitScope(Global)
	} else if wsr.shuttingDown {
		wsr.logger.Info("shutting down services ongoing...")
	} else {
		wsr.logger.Warn("shutting down services not started!")
	}
}

func (wsr *DefaultWinServiceRunner) SendStopSignal() {
	wsr.winSvc.TriggerStop()
}
func (wsr *DefaultWinServiceRunner) Execute() {
	// start the host lifecycle
	wsr.host.Start()

	// wait for stop signal from console
	wsr.winSvc.ServiceMain()

	if !wsr.stopped && !wsr.shuttingDown {
		wsr.shutdownHost(wsr.config.ShutdownTimeoutInSec)
	}
}

func (wsr *DefaultWinServiceRunner) shutdownHost(timeoutInSec int) {
	wsr.logger.Infof("shutting down services with timeout(sec): %v", timeoutInSec)
	// shut down with timeout
	err := wsr.host.Shutdown(time.Duration(timeoutInSec) * time.Second)
	if err != nil {
		wsr.logger.Errorw("Host shut down with failure", "last error", err)
	}
	wsr.stopped = true
	wsr.shuttingDown = false
}

type LoopContextForWinSvcRunner struct {
	parent dep.ContextEx
}

func NewLoopContextFromWinServiceRunner(ctxt dep.ContextEx) dep.ServiceContext {
	return &LoopContextForWinSvcRunner{parent: ctxt}
}

func (cws *LoopContextForWinSvcRunner) Type() string                    { return "WinServiceRunner" }
func (cws *LoopContextForWinSvcRunner) Name() string                    { return types.Of(cws).FullName() }
func (cws *LoopContextForWinSvcRunner) GetHostContext() dep.HostContext { return nil }
func (cws *LoopContextForWinSvcRunner) GetLogger() logger.Logger        { return cws.parent.GetLogger() }
func (cws *LoopContextForWinSvcRunner) GetLoggerWithName(name string) logger.Logger {
	return cws.parent.GetLoggerWithName(name)
}
func (cws *LoopContextForWinSvcRunner) GetLoggerFactory() logger.LoggerFactory {
	return cws.parent.GetLoggerFactory()
}
func (cws *LoopContextForWinSvcRunner) GetConfiguration(configType types.DataType) interface{} {
	return cws.parent.GetConfiguration(configType)
}
func (cws *LoopContextForWinSvcRunner) GetComponent(interfaceType types.DataType) interface{} {
	return cws.parent.GetComponent(interfaceType)
}
func (cws *LoopContextForWinSvcRunner) CreateWithProperties(interfaceType types.DataType, props dep.Properties) interface{} {
	return cws.parent.CreateWithProperties(interfaceType, props)
}
func (cws *LoopContextForWinSvcRunner) GetProperties() dep.Properties {
	return nil
}

//
// WinSVC
//
type StatusCheckerConfig struct {
	CheckIntervalInMS uint32
	WaitTimeOutInMS   uint32
}
type WinSVCConfig struct {
	ServiceName    string
	UsePreShutdown bool
	StatusChecker  StatusCheckerConfig
}

type WinSVC interface {
	Initialize(ctxt dep.ServiceContext, svcName string) WinSVC
	SetStopSignalCallback(func(os.Signal) bool)
	SetStartStoppingCallback(callback func(StopReason))
	SetStoppingStatusChecker(config *StatusCheckerConfig, checker LoopProcessor)

	ServiceMain()

	// request for stop the service, trigger only without waiting
	TriggerStop()
}

type DefaultWinSVC struct {
	context               dep.ServiceContext
	logger                logger.Logger
	config                *WinSVCConfig
	done                  chan os.Signal
	triggerStopOnce       sync.Once
	stopChan              chan bool
	shutdownOnce          sync.Once
	shutdownChan          chan bool
	startShuttingDownOnce sync.Once
	checkLoopOnce         sync.Once
	stopped               bool
	acceptStopSignal      func(sig os.Signal) bool
	onStartStopping       func(StopReason)
	checkStatusProcessor  LoopProcessor
}

func NewWinSVC(context dep.Context) *DefaultWinSVC {
	return &DefaultWinSVC{
		config: &WinSVCConfig{
			ServiceName:    "windows-service",
			UsePreShutdown: false,
			StatusChecker:  StatusCheckerConfig{},
		},
		done:                 make(chan os.Signal),
		stopChan:             make(chan bool),
		shutdownChan:         make(chan bool),
		stopped:              false,
		checkStatusProcessor: nil,
	}
}

func (ws *DefaultWinSVC) Initialize(ctxt dep.ServiceContext, svcName string) WinSVC {
	ws.context = ctxt
	ws.config.ServiceName = svcName

	ws.logger = ctxt.GetLoggerWithName(fmt.Sprintf("WinSVC[%s]", ws.config.ServiceName))
	return ws
}
func (ws *DefaultWinSVC) SetStopSignalCallback(callback func(os.Signal) bool) {
	ws.acceptStopSignal = callback
}
func (ws *DefaultWinSVC) SetStartStoppingCallback(callback func(StopReason)) {
	ws.onStartStopping = callback
}
func (ws *DefaultWinSVC) SetStoppingStatusChecker(config *StatusCheckerConfig, processor LoopProcessor) {
	ws.config.UsePreShutdown = true
	ws.checkStatusProcessor = processor
	if processor != nil {
		ws.config.StatusChecker = *config
	} else {
		ws.config.StatusChecker = StatusCheckerConfig{
			CheckIntervalInMS: 1000,
			WaitTimeOutInMS:   3000,
		}
	}
}

var elog debug.Log

func (ws *DefaultWinSVC) ServiceMain() {
	ws.logger.Infow("WinSVC running...")
	elog, _ = eventlog.Open(ws.config.ServiceName)
	elog.Info(1, "WinSvc main start")
	defer elog.Close()

	signal.Notify(ws.done, syscall.SIGINT, syscall.SIGTERM)

	err := svc.Run(ws.config.ServiceName, ws)
	if err != nil {
		inService, err_in := svc.IsWindowsService()
		if err_in != nil {
			elog.Error(1, "failed to determine if we are running in windows service")
			ws.logger.Fatalf("failed to determine if we are running in windows service: %v", err_in)
		}

		if !inService {
			elog.Error(1, fmt.Sprintf("not running inside windows service: %v", err))
			ws.logger.Errorf("not running inside windows service: %v", err)
		} else {
			elog.Error(1, fmt.Sprintf("windows service exit with error: %v", err))
			ws.logger.Errorf("windows service exit with error: %v", err)
		}
	}

	ws.stopped = true

	elog.Info(1, "WinSvc main complete")
	ws.logger.Infow("WinSVC run complete")
}
func (ws *DefaultWinSVC) TriggerStop() {
	ws.triggerStopOnce.Do(func() {
		close(ws.stopChan)
	})
}
func (ws *DefaultWinSVC) stopExecution() {
	// trigger the loop to stop, executed exactly once
	ws.shutdownOnce.Do(func() {
		close(ws.shutdownChan)
	})
}

func (ws *DefaultWinSVC) onStopSignal(sig os.Signal) bool {
	if ws.acceptStopSignal != nil {
		return ws.acceptStopSignal(sig)
	}
	return true
}

func (ws *DefaultWinSVC) startStopping(reason StopReason) {
	if ws.onStartStopping == nil {
		panic(fmt.Errorf("OnStoppingCallback not configured!"))
	}

	ws.onStartStopping(reason)
}

func (ws *DefaultWinSVC) checkStoppingStatus(globalCtxt LoopGlobalContext, checkpoint uint32) bool {
	if ws.checkStatusProcessor != nil {
		// create context for new iteration run of the loop
		runCtxt := NewLoopRunContext(globalCtxt)

		runCtxt.SetVariable("checkpoint", checkpoint)

		ws.logger.Info("Run pre-shutdown processor...")
		ws.checkStatusProcessor.Run(runCtxt)

		return runCtxt.IsStopped()
	}
	return true
}

type StopReason uint8

const (
	SIG_STOP StopReason = iota
	USER_STOP
	SC_STOP
	SC_SHUTDOWN
	SC_PRESHUTDOWN

	//_minType = SC_STOP
	//_maxType = SC_PRESHUTDOWN
)

func (sr StopReason) String() string {
	switch sr {
	case SIG_STOP:
		return "Signal_Stop"
	case USER_STOP:
		return "User_Stop"
	case SC_STOP:
		return "SC_STOP"
	case SC_SHUTDOWN:
		return "SC_SHUTDOWN"
	case SC_PRESHUTDOWN:
		return "SC_PRESHUTDOWN"
	default:
		return fmt.Sprintf("StopReason(%d)", sr)
	}
}
func (ws *DefaultWinSVC) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	ws.logger.Info("service loop pending start")

	elog.Info(1, "WinSVC pending start")
	changes <- svc.Status{State: svc.StartPending}

	cmdsAccepted := svc.AcceptStop | svc.AcceptPauseAndContinue
	if ws.config.UsePreShutdown {
		cmdsAccepted = cmdsAccepted | svc.AcceptPreShutdown
		err := ws.setPreShutdownTimeOutInMilliSec(ws.config.StatusChecker.WaitTimeOutInMS)
		if err != nil {
			ws.logger.Errorf("Error set preshutdown timeout, err: %v", err)
		}
	} else {
		cmdsAccepted = cmdsAccepted | svc.AcceptShutdown
	}

	elog.Info(1, "WinSVC running")
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	ws.logger.Info("WinSVC loop running")
	var reason StopReason
loop:
	for {
		select {
		case sig := <-ws.done:
			ws.logger.Info("signal is triggered.")
			elog.Info(1, "WinSVC received TERM or INT signal")
			accept := ws.onStopSignal(sig)
			if accept {
				ws.logger.Infow("Stop signal is accepted", "Signal", sig.String())
				reason = SIG_STOP
				ws.onStopping(reason, changes)
			}
		case <-ws.stopChan:
			ws.logger.Info("user triggered to stop.")
			elog.Info(1, "user triggered to stop.")
			reason = USER_STOP
			ws.onStopping(reason, changes)
		case <-ws.shutdownChan:
			ws.logger.Info("WinSVC loop exit is triggered.")
			elog.Info(1, "WinSVC loop exiting")
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Pause:
				ws.logger.Info("This service is paused.")
				elog.Info(1, "WinSVC paused")
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				ws.logger.Info("This service is continued.")
				elog.Info(1, "WinSVC continued")
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			case svc.Stop:
				ws.logger.Info("This service is stopped.")
				elog.Info(1, "WinSVC stop triggered")
				reason = SC_STOP
				ws.onStopping(reason, changes)
			case svc.PreShutdown:
				ws.logger.Info("This service is (pre)shutting down.")
				elog.Info(1, "WinSVC preshutdown triggered")
				reason = SC_PRESHUTDOWN
				ws.onStopping(reason, changes)
			case svc.Shutdown:
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				ws.logger.Info("This service is shutting down.")
				elog.Info(1, "WinSVC shutdown triggered")
				reason = SC_SHUTDOWN
				ws.onStopping(reason, changes)
			default:
				ws.logger.Errorf("unexpected control request #%d", c)
			}
		}
	}

	ws.logger.Infof("WinSVC loop exit with reason: %v", reason)
	elog.Info(1, fmt.Sprintf("WinSVC loop exit with reason: %v", reason))

	return
}

func (ws *DefaultWinSVC) onStopping(reason StopReason, changes chan<- svc.Status) {
	// step 1: start stopping actual services
	ws.startShuttingDownOnce.Do(func() {
		ws.startStopping(reason)
	})

	// step 2: checking stopping status
	ws.checkLoopOnce.Do(func() {
		go ws.checkAndReportStoppingStatus(changes)
	})
}
func (ws *DefaultWinSVC) checkAndReportStoppingStatus(changes chan<- svc.Status) {
	defer ws.stopExecution()

	// create global context for preShutdown loop
	preShutdownCtxt := NewWinSvcLoopGlobalContext(ws.context, ws.config.ServiceName)

	checkpoint := uint32(0)
	for {
		waitHint := uint32(3000)
		if ws.config.StatusChecker.WaitTimeOutInMS > checkpoint*ws.config.StatusChecker.CheckIntervalInMS+waitHint {
			waitHint = ws.config.StatusChecker.WaitTimeOutInMS - checkpoint*ws.config.StatusChecker.CheckIntervalInMS
		}

		checkpoint++
		changes <- svc.Status{State: svc.StopPending, WaitHint: waitHint, CheckPoint: checkpoint}

		if ws.checkStoppingStatus(preShutdownCtxt, checkpoint) {
			break
		}

		time.Sleep(time.Duration(ws.config.StatusChecker.CheckIntervalInMS) * time.Millisecond)
	}

	checkpoint++
	changes <- svc.Status{State: svc.StopPending, CheckPoint: checkpoint}
}

type SERVICE_PRESHUTDOWN_INFO struct {
	PreshutdownTimeout uint32
}

func (ws *DefaultWinSVC) setPreShutdownTimeOutInMilliSec(timeout uint32) error {
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(h)
	s, err := windows.UTF16PtrFromString(ws.config.ServiceName)
	if err != nil {
		return err
	}

	hs, err := windows.OpenService(h, s, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(hs)
	// set preshutdown timeout
	info := SERVICE_PRESHUTDOWN_INFO{
		PreshutdownTimeout: timeout,
	}
	err = windows.ChangeServiceConfig2(hs, windows.SERVICE_CONFIG_PRESHUTDOWN_INFO, (*byte)(unsafe.Pointer(&info)))
	if err != nil {
		return err
	}
	return nil
}

type WinServiceLoopGlobalContext struct {
	context   dep.ServiceContext
	svcName   string
	Variables *VariableSet
}

func NewWinSvcLoopGlobalContext(context dep.ServiceContext, svcName string) LoopGlobalContext {
	return &WinServiceLoopGlobalContext{
		context:   context,
		svcName:   svcName,
		Variables: NewVariableSet(),
	}
}

func (wslc *WinServiceLoopGlobalContext) LooperName() string {
	return wslc.svcName
}
func (wslc *WinServiceLoopGlobalContext) GetLooperContext() dep.ServiceContext {
	return wslc.context
}

func (wslc *WinServiceLoopGlobalContext) HasVariable(key string) bool {
	return wslc.Variables.Exist(key)
}
func (wslc *WinServiceLoopGlobalContext) GetVariable(key string) interface{} {
	return wslc.Variables.Get(key)
}
func (wslc *WinServiceLoopGlobalContext) SetVariable(key string, value interface{}) {
	wslc.Variables.Set(key, value)
}
