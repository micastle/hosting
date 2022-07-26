# Hosting: Integrated API for app hosting and dependency injection

Hosting is a set of integrated APIs that developer can used in their application to configure their application and components used in the app.

In this chapter we only discuss APIs available only in Hosting API. For shared APIs please refer to [Common APIs](./CommonAPI.md).



## Entry Point: Host Creation and Run

There are a few steps to create an app host and run it.

- Create a host builder and configure it;
- Build the host instance from host builder;
- Run the host instance.

Below are an example of how hosting framework is used to create app host and start it:

```go
func createHostBuilder() HostBuilder {
	builder := NewDefaultHostBuilder()
    builder.SetHostName("Test")
    return builder
}
func Test_Host_basic(t *testing.T) {
	builder := createHostBuilder()
	host := builder.Build()
	host.Run()
}
```



## Host Builder APIs 

Below are the APIs for configuring your host:

```go
type HostBuilder interface {
	SetHostName(name string) HostBuilder
	UseDefaultComponentProvider() HostBuilder
	UseComponentProvider(configure ConfigureComponentProviderMethod) HostBuilder

	UseDefaultAppRunner() HostBuilder
	UseBasicSyncAppRunner() HostBuilder

	ConfigureHostConfiguration(configure ConfigureHostMethod) HostBuilder
	ConfigureHostConfigurationEx(configLoader LoadHostConfigurationMethod) HostBuilder

	ConfigureLogging(configure ConfigureLoggerFactoryMethod) HostBuilder
	ConfigureLoggingEx(configure ConfigureLoggingMethod) HostBuilder

	ConfigureAppConfiguration(configure ConfigureAppMethod) HostBuilder
	ConfigureAppConfigurationEx(configLoader LoadAppConfigurationMethod) HostBuilder

	ConfigureComponents(configure ConfigureComponentsMethod) HostBuilder
	ConfigureServices(configure ConfigureServicesMethod) HostBuilder
	ConfigureLifecycle(configure ConfigureLifecycleMethod) HostBuilder

	UseService(serviceType types.DataType, createService FreeStyleServiceFactoryMethod) HostBuilder
	UseLoop(name string, configure ConfigureLoopMethod) HostBuilder

	Build() Host
}
```

below is the API for developers to configure their components:

```go
ConfigureComponents(configure ConfigureComponentsMethod) HostBuilder

// ConfigureComponentsMethod syntax
type ConfigureComponentsMethod func(context BuilderContext, components dep.ComponentCollectionEx)
```

In the passed-in configure method, developer should register their dependencies through the component collection interface with APIs described in [Common API](./CommonAPI.md).



Below is an example to run a basic loop in your app:

```go
// define your component
type Helloer interface{
    Hello()
}
type MyHello struct {}
func NewHelloer() *MyHello { return &MyHello{} }
func (h *MyHello) Hello() {
	fmt.Println("Hello, world!")
}

func Test_looper_basic(t *testing.T) {
    // configure host builder
    builder := NewDefaultHostBuilder()
    builder.SetHostName("Test")
    builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
        dep.RegisterTransient[Helloer](components, NewHelloer)
	})
    // use a loop "Hello" in this app
	builder.UseLoop("Hello", func(context ServiceContext, looper ConfigureLoopContext) {
        looper.SetInterval(time.Duration(500) * time.Millisecond)
        looper.UseFuncProcessor(func(hello Helloer) {
            hello.Hello()
        })
	})
	// build the host and run it
	host := builder.Build()
	host.Run()
}
```

## Extra APIs for Host Concepts

Hosting framework introduced a few new concepts, like Service, Looper, Looper Processor, etc. Below are the new APIs regarding to these concepts.

### Service APIs

```go
func UseService[T any](builder HostBuilder, createService FreeStyleServiceFactoryMethod)
```

### Processor APIs

```go
func RegisterFuncProcessor[T FunctionProcessor](collection dep.ComponentCollectionEx, processorFunc dep.FreeStyleProcessorMethod)

func UseProcessor[T any](group ConfigureGroupContext, condition ConditionMethod)
```



### Loop Variable APIs

```go
func GetVariable[T any](scope ScopeContextBase, key string) T

func SetVariable[T any](scope ScopeContextBase, key string, value T)
```



### Utility Component APIs

There are a group of components registered by the framework itself, developers can resolve them by type and use them as necessary:

​        Dependency Type: dep.LifecycleController
​        Dependency Type: dep.ScopeFactory
​        Dependency Type: hosting.HostAsyncOperator
​        Dependency Type: hosting.FunctionProcessor
​        Dependency Type: logger.LoggerFactory
​        Dependency Type: dep.Scope
​        Dependency Type: hosting.Host

### Platform Specific Components

#### WinServiceRunner

Only available on Windows.

```go
type WinServiceRunner interface {
	AsyncAppRunner
}
```

#### WinSVC

Only available on Windows.

```go
type WinSVC interface {
	Initialize(ctxt dep.ServiceContext, svcName string) WinSVC
	SetStopSignalCallback(func(os.Signal) bool)
	SetStartStoppingCallback(callback func(StopReason))
	SetStoppingStatusChecker(config *StatusCheckerConfig, checker LoopProcessor)

	ServiceMain()

	// request for stop the service, trigger only without waiting
	TriggerStop()
}
```

#### SystemdService

Only available on Linux.

**[TODO] Not implemented yet at this point.**
