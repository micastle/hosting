# Diagnose Issues



In order to diagnose issues when using the framework, a few tools are provided for developers.



In additional to the cyclic dependency detection and recurrence detection, below APIs are provided to verify the runtime component graph created by the framework. Developers can use these APIs to find out bugs in case dependency is created by mistakes.



## Dependency Stack

```go
func PrintDependencyStack(context Context)
```

This API helps developers to dig into their dependency stack. The input context represents the target component and this API will show developer who depends on this target component, and seek backward through the dependency chain until the root, which is usually the host object. you can see the component type name as well as dependency type, and details like scope and starting properties will also be rendered.



## Scope Stack

```go
func PrintAncestorStack(context Context, scenario string)
```

This API help to understand the scopes that target component lives in. as scopes can be nested in another scope, this API will seek from bottom to top recursively and show you the ancestor scopes of the target component. with such diagnose info, developers may find out mis-matched scopes that are introduced accidentally.

