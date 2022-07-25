# Component

Component is a very generic concept in hosting framework. Anything can be a component, so we don't have an explicit interface for component type. 

You can take interface of Component as below:

```go
type Component interface{}
```

You can define your own component type specifically, like:

```go
type MyComponent interface{
    Print()
}
```

then you can implement your own type of component and register to the host:

```go
type MyComponentImpl struct {}
func NewMyComponent() *MyComponentImpl { return &MyComponentImpl{} }
func (mc *MyComponentImpl) Print() { fmt.Println("Hi") }

func ConfigureComponents(context hosting.BuilderContext, components dep.ComponentCollection) {
	components.RegisterTransientForType(NewMyComponent, types.Of(new(MyComponent)))
}
```



## Registration

Your component can be registered as Singleton or Transient component:

- Singleton: the framework will manage your component type as singleton, only one instance will be created in the host lifecycle. You don't have to manage the instance number in your factory method "NewMyComponent" - and you shouldn't either.
- Transient: the framework will create new instance for the component type every time when it is requested. Instances are never re-used.



## Multi-interface Support

One object can implement multiple different interfaces in Golang. We support registering multiple interfaces for a singleton component within one registration.

```
components.RegisterSingletonForTypes(NewMyComponent, types.Of(new(TypeA)), types.Of(new(TypeB)), types.Of(new(TypeC)))
```

It is recommended to do it in this way. It may not result in single instance if you register for each type separately in three calls.