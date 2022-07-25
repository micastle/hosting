# Factory Method

Factory Method has a important position in hosting framework. As Golang does not have similar concept like class constructor in C++/C#, we use factory method to construct component instance, and also use its argument list to find out component dependencies during dependency injection.



## Component Factory Method

Used to register component. Support free-style factory method.

```go
// prototype
type FreeStyleFactoryMethod interface{}
// validation type
type FreeStyleFactoryMethod func(...) interface{}
```

Requirements:

- Must be a func
- The return type of the func must be your component type
- The kind of your component type must be an interface, empty interface{} is not allowed (by default)
- The actual returned instance at run time must implement your component interface
- Types of func arguments should all be registered either by the framework or by your ConfigureComponents method, or available in the component creation context(see details in [Component Context](./Context.md#Component Contexts)).



## Service Factory Method

Used to register component. Support free-style factory method.

```go
// prototype
type FreeStyleServiceFactoryMethod interface{}
// validation type
type FreeStyleServiceFactoryMethod func(...) hosting.Service
```

Requirements:

- Must be a func
- The return type of the func must be your service type
- The kind of your service type must be an interface, which implements at least the hosting.Service interface
- The actual returned instance at run time must implement your service interface
- Types of func arguments should all be registered either by the framework or by your ConfigureComponents method, or available in the service creation context(see details in [Component Context](./Context.md#Component Contexts)).



## Other Functions

Other functions may not actually factory methods, but they can improve the convinience for certain coding tasks. 

### Processor Func Method

Used to register component. Support free-style factory method.

```go
// prototype
type FreeStyleProcessorMethod interface{}
// validation type
type FreeStyleProcessorMethod func(...)
```

Requirements:

- Must be a func
- The func must not have any return outputs
- Types of func arguments should all be registered either by the framework or by your ConfigureComponents method, or available in the processor creation context:
  - The Component Context of the Processor itself
  - The Scope Context to be Run against in the processor
  - [TODO] Logger



### Loop Initializer:



### Logging Initializer:



### Configuration Loader:



### AppLifecycle Initializer:

