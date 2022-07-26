# Activator: Standalone API for Dependency Injection

Activator is a set of standalone APIs that developer can used in their application to manage component dependencies.

In this chapter we only discuss APIs available only in Activator. For shared APIs please refer to [Common APIs](./CommonAPI.md).



## Entry Point: Activator Creation

Two APIs are provided to create activator instance:

```go
func CreateActivator(configureComponents ConfigureComponentsMethod) Activator

func CreateActivatorEx(debug bool, configureComponents ConfigureComponentsMethod, loggerFactory logger.LoggerFactory) Activator
```

The first one create activator with default logger factory in debug mode. Default logger factory print logs to stdout with default format and debug level.

Developer can specify running mode as well as your own logger factory.



## Activator APIs:

Activator is has a very simple interface as below,even though internally it encapsulates all necessary resources for dependency management and injection:

```go
type Activator interface {
	GetProvider() dep.ComponentProvider
}

func GetComponent[T any](avt Activator) T
```

Usually developer don't use Activator APIs here to resolve their components. Instead, Activator APIs are used to resolve the "root" components of their system, then dependencies of these root components will be created automatically through dependency injection. This happens in a dependency chain and Activator APIs are the starter of this chain. Afterwards, developer leverage the automated injection or use the shared common APIs to interact with the framework.
