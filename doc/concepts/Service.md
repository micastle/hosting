# Service

Service is a special kind of component. Below is the interface for Service:

```go
type Service interface {
	Run()
	Stop(ctx context.Context) error
}
```

- Run(): execute the actual service code, until it is stopped.
- Stop(ctx context.Context) error: stop the service with the background context, return error if any. It should wait for service stop completion instead triggering for stop only.



## Running Mode

Service can be run in sync mode or async mode.

- Sync Mode: external code calls Run() directly, it will run the actual service code continuously until complete and return.
- Async Mode: external code does not call Run() directly, but call it with go routine.



## AsyncService

AsyncService runs in async mode, below is the interface:

```
type Service interface {
	Start()
	Run()
	Stop(ctx context.Context) error
}
```

- Start(): start the service and return once the actual service code has started running. It should trigger the service running and wait until it is actually running.

