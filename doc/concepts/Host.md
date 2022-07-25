# Host

A *host* is an object that encapsulates an app's resources and lifetime functionality, such as:

- Dependency injection (DI)
- Logging
- Configuration
- Hosted services
- App shutdown



## Set up a host

The host is typically configured, built, and run by code in the main function.

```go
hostBuilder := hosting.NewDefaultHostBuilder()
hostBuilder.SetHostName("Sample")
hostBuilder.UseLoop("HelloLoop", ...)
hostBuilder.UseService(types.Of(new(WorkerService)), func(context dep.ServiceContext) WorkerService {...}
                       
host := hostBuilder.Build()
host.Run()
```



## Host Component

Host is also a kind of component in hosting framework. anything applies to components also applies to host.
