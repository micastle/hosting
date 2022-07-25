package hosting

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
)

type OnHostReady func(ctxt dep.Context)

type EventType uint8

const (
	EVENT_TYPE_SIGNAL EventType = iota
	EVENT_TYPE_WINSVC
)

func (et EventType) String() string {
	switch et {
	case EVENT_TYPE_SIGNAL:
		return "Signal"
	case EVENT_TYPE_WINSVC:
		return "WinSvc"
	default:
		return fmt.Sprintf("EventType(%d)", et)
	}
}

type StopEvent struct {
	Type EventType
	Data interface{}
}

// return stop event is accepted or not
// - in some cases application is not able to control the acceptance of stop event, thus the returned value may be ignored.
type OnStopEvent func(dep.Context, *StopEvent) bool

type OnApplicationStarted func(dep.Context)
type OnApplicationStopped func(dep.Context)
type OnApplicationStopping func(dep.Context)

type ApplicationLifecycle interface {
	RegisterOnHostReady(OnHostReady)
	RegisterOnStopEvent(OnStopEvent)
	RegisterOnAppStarted(OnApplicationStarted)
	RegisterOnAppStopped(OnApplicationStopped)
	RegisterOnAppStopping(OnApplicationStopping)
}

type LifecycleHandler interface {
	OnHostReady(context *DefaultHostContext)

	OnAppStarted(context *DefaultHostContext)
	OnStopEvent(context *DefaultHostContext, event *StopEvent) bool
	OnAppStopping(context *DefaultHostContext)
	OnAppStopped(context *DefaultHostContext)
}

type DefaultLifecycle struct {
	hostReadyHook OnHostReady
	stopEventHook OnStopEvent

	onStoppingHook OnApplicationStopping
	onStoppedHook  OnApplicationStopped
	onStartedHook  OnApplicationStarted
}

func NewDefaultLifecycle() *DefaultLifecycle {
	return &DefaultLifecycle{}
}

func (l *DefaultLifecycle) RegisterOnHostReady(onHostReady OnHostReady) {
	l.hostReadyHook = onHostReady
}

func (l *DefaultLifecycle) RegisterOnStopEvent(onStopEvent OnStopEvent) {
	l.stopEventHook = onStopEvent
}

func (l *DefaultLifecycle) RegisterOnAppStarted(onAppStarted OnApplicationStarted) {
	l.onStartedHook = onAppStarted
}

func (l *DefaultLifecycle) RegisterOnAppStopped(onAppStopped OnApplicationStopped) {
	l.onStoppedHook = onAppStopped
}

func (l *DefaultLifecycle) RegisterOnAppStopping(onAppStopping OnApplicationStopping) {
	l.onStoppingHook = onAppStopping
}

func (l *DefaultLifecycle) OnHostReady(context *DefaultHostContext) {
	if l.hostReadyHook != nil {
		l.hostReadyHook(context)
	}
}

// return true(default): signal is accepted, false: signal is ignored.
func (l *DefaultLifecycle) OnStopEvent(context *DefaultHostContext, event *StopEvent) bool {
	if l.stopEventHook != nil {
		return l.stopEventHook(context, event)
	}
	return true
}

func (l *DefaultLifecycle) OnAppStarted(context *DefaultHostContext) {
	if l.onStartedHook != nil {
		l.onStartedHook(context)
	}
}

func (l *DefaultLifecycle) OnAppStopping(context *DefaultHostContext) {
	if l.onStoppingHook != nil {
		l.onStoppingHook(context)
	}
}

func (l *DefaultLifecycle) OnAppStopped(context *DefaultHostContext) {
	if l.onStoppedHook != nil {
		l.onStoppedHook(context)
	}
}
