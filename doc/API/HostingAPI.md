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
```

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

