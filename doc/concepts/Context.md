# Context

There are several kinds of context in hosting framework.



## Component Contexts

### Component Context

In addition to all registered components, there are below contextual dependencies available during component instance creation:

- Component Context itself
- [TODO] Logger

### Service Context

In addition to all registered components, there are below contextual dependencies available during service instance creation:

- Service Context itself
- Component Context: service is a kind of component, so component context is also available.
- [TODO] Logger

### Host Context





## Looper Contexts

- LoopGlobalContext
- LoopRunContext
- ScopeContext



## Other Contexts

- BuilderContext
- 