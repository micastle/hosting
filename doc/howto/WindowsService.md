# Create Windows Service

Hosting framework can be used to create windows service with minimum code.

Windows Service has its own specific requirements, see details at official site: [service programs](https://docs.microsoft.com/en-us/windows/win32/services/service-programs). with hosting framework you can build your windows service with focus on your application logic while the framework help you take care of the windows service contract integration.



## Configure Your Windows Service

similar like other application types, you configure and build your host with below code for a windows service program:

```go
func main() {
	host := ConfigureHost().Build()
	host.Run()
}
```

the only additional thing you need to do is to register below configuration and component to your host during components configuration:

```go
	components.AddConfiguration(&hosting.WinSvcAppRunnerConfig{
		ServiceName:                "your-win-service-name",
		ShutdownTimeoutInSec:       15,
		ShutdownCheckIntervalInSec: 1,
	})
	components.RegisterSingletonForTypes(hosting.NewWinServiceRunner, types.Of(new(hosting.AppRunner)))
```

others remain unchanged, configure host configuration, logging, app configuration, registering your components and services, etc. You application code are organized in your components and services, just like when you build a console application in [examples](../samples/Console Application.md).



## Handle Stop Event

optionally you can handle lifecycle events dedicated for windows service. Windows Service have its own lifecycle callbacks which are different from other application types.

you can receive Stop Event callback from windows service when [service controller send request](https://docs.microsoft.com/en-us/windows/win32/services/service-control-requests) to your windows service if registered properly.

```go
appLifecycle.RegisterOnStopEvent(func(ctxt dep.Context, se *hosting.StopEvent) bool {...})
// Stop Event Type for windows service: EVENT_TYPE_WINSVC
// Event Data (StopReason):
// 	SIG_STOP StopReason = iota
//	USER_STOP
//	SC_STOP
//	SC_SHUTDOWN
//	SC_PRESHUTDOWN
```

