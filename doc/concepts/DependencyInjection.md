# Dependency Injection

Dependency Injection(DI) is a popular concept in developer world. You can search on internet to learn more about it. In this article we will discuss more about features and constraints in hosting framework.



DI in hosting framework happens in majorly two categories of scenarios:

1. Dependencies are injected automatically during factory methods.
2. Dependencies are injected manually through context.



## Dependency Registration

Before hosting framework is able to inject dependencies for you, they should be registered.

### Dependency Types

Dependencies are registered agaist dependency type, thus they can be matched by the type when dependency injection is requested. there are two basic categories of dependencies supported in the framework:

1. Configurations: configurations are struct pointers, you register against the struct pointer type with a pointer to your configuration struct, then injection can happen for the same pointer type to your configuration struct.
2. Components: other components are required to be registered against your component interface type. we explicitly forbidden struct types as component type.

There are several types of component dependencies:

1. Built-in dependencies: they are registered by the hosting framework, you can safely use they through injection without registering.
2. Contextual dependencies: they are injected within specific context if requried, no need to register them. One example is the dep.Context,  we will discuss more about them in here: [Context](./Context.md).
3. Registered dependencies: these are those component types you registered before using them.



### Multi-type Registration

We can implement a component that implements multiple interfaces in Golang. So the framework also support registering one component instance against multiple component types.

This is especially necessary for singleton component, as separate registration will lead to multiple instances which breaks the "singleton" concept.

```go
RegisterSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType)
```

In this way, the framework ensures there is only one instance created, even though multiple component types are registered. You can call GetComponent() or expect automatical DI with any one of the component types to create and inject the instance as necessary just like usual, not worrying about multiple instances being created.



### Constraints

A few constraints on dependency registration:

- Duplicated dependency type registration is not allowed and lead to panic. 
- Don't register dependency type which is the same as contextual dependencies, or your registration may get hidden by them as contextual ones have highest priority.
- Be careful to register component type that is a built-in type: you may replace the default implementation of that component type in this case. So make sure the registration is intended if you do register the type, or it will lead to unexpected behavior. 



## Built-In Dependencies

Built-in dependencies are registered by the framework itself. You can use these system dependencies directly in your code through injection.

Registered system dependencies(10)

​    Dependency Type: *logger.LoggerFactory

​    Dependency Type: *dep.ContextualProvider

​    Dependency Type: *hosting.Host

​    Dependency Type: *hosting.HostAsyncOperator

​    Dependency Type: *hosting.AppRunner

​    Dependency Type: *hosting.AsyncAppRunner

​    Dependency Type: *hosting.Looper

​    Dependency Type: *hosting.ProcessorGroup

​    Dependency Type: *hosting.FunctionProcessor

​    Dependency Type: *hosting.WinSVC



## Dependency Injection

Dependency injection can apply to both component category and configuration category, we don't differentiate a dependency is a configuration or a component. 

As mentioned above, there are several supported DI scenarios. We will have different contextual dependencies for specific DI scenarios, they can be different components of different types or event different actual components of the same type depending on the scenario (Context is one example for this).

### Factory Method Injection

As Golang does not have component constructor concept at programming language level, we usually create component instance with a function. This function is called "factory method" for that specific component type.

```go
type MyComponent interface{}
func NewMyComponent(dep1 Dep1Type, dep2 Dep2Type, ...) MyComponent {...}
```

Above sample code shows how a factory method for component type "MyComponent" looks like.

- A factory method should return a single output and the output type should be compatible with your component type, either your component interface type or pointer type of a struct that implements your component interface.
- all input arguments of the factory method will be injected automatically if below constraints are met, or panic will be raised.
  - The argument type are either registered configuration type or component type
  - The argument type can be built-in component type
  - The argument type can be contextual dependency type(highest priority for dependency type matching)



There are a few DI scenarios that have built-in the factory method injection capability and can apply automatically and apply through the whole dependency chain. They are not limited to below list:

- Component Registration: When you use a free-style factory method to register a component, DI will take place when other components depends on this component and request to create instance of this component type.  DI will find out the argument types of registered factory method and create corresponding instances through recursive DI with registered component types, then call the registered factory method using these instances as the input arguments, thus the requested component is created. Dependency tree is created in this recursive way.
- Use Service: When you use a free-style service factory method to register a service, DI will take place similar like above.
- [TODO]Use Loop: When you use a free-style loop configuration method to register a loop service, DI will take place similar like above.
- Register Processor: When you use a free-style processor method to register a func processor, DI will take place similar like above.



### Manual Context Injection 

The framework supports Manual Injection through contextual dependency "dep.Context". See below:

```go
type Dependency interface{}

type MyComponent interface{}
type DefaultMyComponent struct {
    context dep.Context
    dependency Dependency
}
func NewMyComponent(context dep.Context, dep1 Dep1Type, dep2 Dep2Type, ...) MyComponent {
    return &DefaultMyComponent{ context: context }
}
func (mc *DefaultMyComponent) Initialize() {
    mc.dependency = mc.context.GetComponent(types.Of(new(Dependency))).(Dependency)
}
```

The first argument of factory method "NewMyComponent" is of type "dep.Context", it is a contextual dependency supported by the framework, and will be injected automatically by the framework. 

With the approach above, you can have your component depends on the context, and then you can use this context to get components you wanted dynamically later, see "mc.context.GetComponent" in Initialize() method. In this way, you can control when the dependency is injected manually by calling the context API "GetComponent".



### Cyclic Dependency Detection

It may happen in practice that one component depends on another one while the other one depends on this one which brings us a cyclic dependency issue. The dependency cycle can be large than two elements, which makes it not easier to realize.

Hosting framework has built-in capability to detect such dependency cycle, print the cycle for diagnose purpose and raise panic when there is any.

Developers need to change their code to use manual context injection outside of the factory method, instead of using direct factory method dependency. See the "Initialize()" method in the example in previous section. In this way, we can fix the dependency cycle easily.