# Sample for Activator
Below is how developer starting with activator:

```go
// create the activator with func to register your components
avt := CreateActivator(func(context BuilderContext, components dep.ComponentCollectionEx) {
	dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
})

// resolve your necessary component and use it
ano := GetComponent[AnotherInterface](avt)
ano.Another()
```



Below is how dependency chain extends without using a global reference to the created activator instance:

```go
type AnotherStruct struct {
	context Context
    first FirstInterface
	value int
}

func NewAnotherStruct(context Context, first FirstInterface) *AnotherStruct {
	return &AnotherStruct{
		context: context,
        first： first，
		value: 0,
	}
}

func (as *AnotherStruct) Another() {
	fmt.Println("Another", as.value)
    as.first.First()
}

type FirstInterface interface {
	First()
}
```

Notice that in above code:

- Developer don't need to pass in the activator instance in the factory method NewAnotherStruct.
- The context object is also optional, you can keep it in argument list or remove it.
- The second argument of type FirstInterface can work if the component type is registered.
- you can add any argument of registered dependency type into the argument list of the factory method, they will be injected automatically.