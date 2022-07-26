# Shared Common APIs 

Common APIs are shared by both Standalone API set Activator and Integrated API set Hosting.

Common APIs are associated with several common concepts below:

## Component Collection

Component Collection is an abstraction of dependency registry, where you register all your components/configurations which will be used in your system.

```go
type ComponentCollection interface{}
```



Dependency Registration APIs on component collection:

```go
func AddConfig[T any](collection ComponentCollection, config *T)

func RegisterTransient[T any](collection ComponentCollection, createInstance FreeStyleFactoryMethod)
func RegisterScoped[T any, S any](collection ComponentCollection, createInstance FreeStyleFactoryMethod)
func RegisterSingleton[T any](collection ComponentCollection, createInstance FreeStyleFactoryMethod)

func RegisterInstance[T any](collection ComponentCollection, inst T)

func IsComponentRegistered[T any](collection ComponentCollection) bool
```

And an extra API to register component that has multiple implementations which are chosen dynamically at runtime.

```go
func RegisterComponent[T any](collection ComponentCollectionEx, propEval Evaluator, configure ConfigureComponentType)
```



## Component Provider

Component Provider is an abstraction of dependency provider, from which you get necessary dependencies injected wherever needed.

```go
type ComponentProvider interface{}
```

Component Provider is used in two approaches:

- Injection Automatically: during automatic dependency injection, component provider is used to resolve proper dependency instances and inject into your scenarios. in this approach, developers are not involved and it happens automatically through the injection framework.
- Resolve Dependency Dynamically: occasionally, developer may want to resolve dependencies dynamically in their logic. In this case, below APIs on component provider will help.



Dependency Resolving APIs on component provider:

```go
func GetConfig[T any](provider ComponentProvider) *T

func GetComponent[T any](provider ComponentProvider) T
```

And an extra API to create component with properties below. This API can be used to resolve component with multiple implementations. the returned implementation is decided by the passed in properties:

```go
func CreateComponent[T any](provider ComponentProviderEx, props Properties) T
```



## Context Utilities

Context is an abstraction of the component injection context. It is always available for all dependency injection scenarios. developers can get relevant information about the component itself, injection context info as well as a set of handful utilities which can be used to resolve dependencies dynamically or diagnose issues regarding to the dependency graph.

### Functionalities

```go
type Context interface {
	ContextBase
	LoggerProvider
	ComponentProviderEx

	GetProperties() Properties
}

type ContextBase interface {
	Type() string
	Name() string
}

type LoggerProvider interface {
	GetLogger() logger.Logger
	GetLoggerWithName(name string) logger.Logger
	GetLoggerFactory() logger.LoggerFactory
}
```

Context instance can be injected in any injection scenarios supported by the framework. Below is an example to inject it into a component during creation:

```go
type AnotherInterface interface {
	Another()
}

type AnotherStruct struct {
	context Context
	value int
}

func NewAnotherStruct(context Context) *AnotherStruct {
	return &AnotherStruct{
		context: context,
		value: 0,
	}
}
func (as *AnotherStruct) Another() {
	fmt.Println("Another", as.value)
}

// register the component on collection
RegisterTransient[AnotherInterface](collection, NewAnotherStruct)
```

Context instance can be used for two purpose below:

- It represents the component being created itself
- It is the entry point for developers to access dependency injection and management functionalities provided by the framework.

Functionalities provided by Context instance:

- Context Info: Type and Name of the dependency.
- Logging: Logger Factory and Logger creation APIs to get logger for the context.
- Component Provider: Context implements the ComponentProviderEx interface, so developer can use it as a provider and use provider APIs to resolve dependencies dynamically.
- Properties: resolve or inject properties from the context which can be used to read initialize properties during component creation.

### Diagnostic Utilities

Below APIs are provided to diagnose dependency graph:

```go
func PrintDependencyStack(context Context)

func PrintAncestorStack(context Context, scenario string)
```

#### Dependency Stack:

Dependency Stack tracks the dependency chain which triggers the creation of the dependency. Example as below:

-  Root component depends on component A;
- Component A depends on Component B;
- Component B depends on Component C.

Then develop can call above API with the context of component C, and it will print the stack with details thus developers can diagnose potential issues.

Below is an example :

> cyclic dependency detected on scoped[TestScope] component: dep.FirstInterface
>
> Dependency Stack of Component[dep.SecondInterface]:
>
> ​     <TestScope> Component: dep.SecondInterface {type=blob}
>
> ​     <TestScope> Component: dep.FirstInterface
>
> ​     <TestScope> Component: TestScope
>
> ​     <Global> Host: Test
>
> ​     <Root>

#### Ancestor Stack: 

Ancestor Stack tracks the scope chain where the dependency lives in.

Below is an example:

> Ancestor Stack of nestedScope, context Component[*dep.Scope]:
>
> ​     Scope:<ScopeObject@0xc0001da950>
>
> ​     Scope:<TestScope@0xc0001da630>
>
> ​     Scope:<Global>
>
> ​     <Root>

## Scope APIs

Scope are supported fully in this framework.

- Typed Scope: you can have non-typed scope as well as typed scope.
- Typed Dependency for Scope: you can register scoped dependency for either specific scope type or any scope type.
- Scope can be nested inside another.
- Scope can have a scope instance from parent scope or global, which represents the scope. it is available for injection in the whole scope.



### Scope Factory

Scope Factory is used to create new scopes by developers.

Scope Factory is pre-registered by the framework, you can get it either by automatic injection or resolve it dynamically from context instance. Its type is defined below:

```go
type ScopeFactory interface{}
```

Scope APIs:

```go
func CreateTypedScope[T Scopable](factory ScopeFactory) ScopeEx

func CreateScopeFrom[T Scopable](factory ScopeFactory, scopeInst T) ScopeEx
```



### Scope Utilities

Below utility API helps create scope and manage the lifecycle of the scope safely, developer can ignore the internal details.

```go
func Using[T Scopable](inst T, scopeFunc FreeStyleScopeActionMethod)
```

