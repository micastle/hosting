# Logging

Logging is the basic components in modern application. the framework support default logging components and enable you to register your own logger to replace the default one.



## Configure Logging

You can use default logging, configure your logging or register your own logger factory to customize logging behavior.

Below is an example of using a customized logger factory.

```go
hostBuilder.ConfigureLogging(func(context BuilderContext, factoryBuilder LoggerFactoryBuilder){
    factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory {...})
})
```

the default logging system uses a standard output console logger, which can be directly used for tools, unit test purpose.



## Use Logging

Using logging in your application is simple, you can get logger for your components in factory method, like below:

```go
func NewComponent(context dep.Context, config *Configuration) *DefaultComponent {
    comp := &DefaultComponent{
		logger:     context.GetLogger(),
		config:     config,
	}
	comp.logger.Info("created component: ", types.Of(comp).FullName())
	return comp
}
```

or you can directly inject logger as a dependency:

```go
func NewComponent(log logger.Logger, config *Configuration) *DefaultComponent {
    comp := &DefaultComponent{
		logger:     log,
		config:     config,
	}
	comp.logger.Info("created component: ", types.Of(comp).FullName())
	return comp
}
```

By  default, this logger use your component type name as the logger name - it is configuration during injection. 



## Customize your Logging

You can either customize the logging initialization of the default impl, or provide your own logger factory to customize logging completely.

Two APIs are provided:

```go
type HostBuilder interface {
	...
	ConfigureLogging(configure ConfigureLoggerFactoryMethod) HostBuilder
	ConfigureLoggingEx(configure ConfigureLoggingMethod) HostBuilder
	...
}
```

