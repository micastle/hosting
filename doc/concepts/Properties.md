# Properties

Properties are supported in the framework as a set of read-only flexible environment named properties. Developers control what properties are set and the types of the properties during initialization stage of specific scope. Properties are supported on global and specific scopes. Sub-scope automatically inherits properties from parent scope, and can have local properties which overwrite properties from parent by property name. 

When components are created, they inherit properties from the scope where they live in, or from the global if no scope is involved.

- Developer can retrieve a copy of properties from the context of specific component, or by automatic injection with type "dep.Properties".
- Developer can use properties as a source to initialize the component, but remember it is read-only, you cannot change it internally to pass over the changes made on it.
- Expected usage on properties is using them as key flags or options, not passing your business data from one place to another.



## Set Properties

There are several places that developers can initialize the value of properties:

- Create Global Context: you can initialize the properties of global context, which is attached to the global "scope", and then inherited by all sub-scopes and components created from this context.
- Create Scope: when you create a customized scope, you have a chance to initialize the scope with scope level properties. These properties will overwrite those inherited from parent scope if property name is the same.
- Create Transient Component : you can resolve a transient component with a specific set of properties, which will overwrite properties inherited from parent scopes. As transient component is always resolved as a newly created instance, this ensures each instance will take the provided properties. Singleton and Scoped component does not support this - provided properties are always dropped and the component gets only properties inherited from scopes. 



## Use Properties

### Get Global/Scope flags/options

```go
type Context interface {
	GetProperties() Properties
}
// sample code
is_debug := dep.GetProp[bool](context.GetProperties(), "debug_enabled")
```



### Create Multi-Implementation Component

properties can be used to create component with multiple implementations:

```go
dep.RegisterComponent[Downloader](
		components,
		func(props Properties) interface{} { return props.Get("type") },
		func(comp CompImplCollection) {
			comp.AddImpl("url", NewUrlDownloader)
			comp.AddImpl("blob", NewBlobDownloader)
		},
	)
```

