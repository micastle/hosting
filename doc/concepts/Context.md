# Context

There are several kinds of context in hosting framework.



## Component Contexts

### Component Context

In addition to all registered components, there are below contextual dependencies available during component instance creation:

- Context: Component Context itself
- Logger: a logger instance specifically for corresponding component
- Properties: property set of the component, a copy of the original property set thus it is logically **read-only**.
- ScopeContext: context of the scope where the component lives in.
- ComponentProviderEx: a provider instance which can be used to resolve other dependents dynamically.

### Service Context

In addition to all registered components, there are below contextual dependencies available during service instance creation:

- Context: Service Context itself
- ServiceContext: Service Context itself
- Logger: a logger instance specifically for corresponding component
- Properties: property set of the component, a copy of the original property set thus it is logically **read-only**.
- ScopeContext: context of the scope where the component lives in.
- ComponentProviderEx: a provider instance which can be used to resolve other dependents dynamically.

### Host Context

In addition to all registered components, there are below contextual dependencies available on host context:

- Context: Component Context itself
- Logger: a logger instance specifically for corresponding component
- Properties: property set of the component, a copy of the original property set thus it is logically **read-only**.
- ScopeContext: context of the scope where the component lives in.



## Looper Contexts

- LoopGlobalContext
- LoopRunContext
- ScopeContext



## Other Contexts

- BuilderContext
- 