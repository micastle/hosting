# Looper

Looper is one type of Service which runs logic loops in application. Looper is widely used in background applications.

Below is an example of Looper service in C:

```go
func main() {
    for {
        fmt.Println("Hello, world!")
        sleep(5 * 1000)
    }
}
```



A typical Loop has below features:

- Run each loop iteration with time interval: usually a loop runs with interval to avoid occupying all CPU resources.
- Support Stopping the loop in certain conditions, external code can trigger the condition to stop the loop.
- Logic being executed in each iteration of the loop. We take the logic as a group of processors belongs to the Looper.
- Loop iteration logic is executed in the scope of the loop context, the context is created on iteration start, shared during the iteration and dropped at the end.



So we model the basic Looper with below components:

- LoopRunContext: the context of each Loop iteration, it is also a type of ScopeContext.
- LoopProcessor: a piece of logic running inside the Loop interation. Multiple processors can be executed with a shared LoopRunContext in one iteration.
- Looper can have a configurable interval and can have a service name.



## Features

In addition to the basic loop above, we also support conditions and processor groups inside the loop iteration.

- Conditional Processor: Processor that is executed when certain condition is satisfied.
- Processor Group: a group of Processors that are executed when certain condition is satisfied.
- Contextual Variables: Processors can share data in their scoped context through Get/Set variables.
- Scope Context initializer: prepare the scoped context with commonly referenced variables.

These components are helpful to address below loop with complexity:

```go
func main() {
    for {
        // prepare the context
        sm := GetStateMachine()
        
        // step 1: print something
        fmt.Println("Hello, world!")
        
        // step 2: sync status
        state := sm.SyncState()
        
        // step 3: logic group with condition
        if state == Running {
            // step 3.1: update state machine
            sm.Update()
            // step 3.1: send metrics
            client := GetMetricsClient()
            client.SendMetrics()
        }
        
        // step 4: notify stop state and quit loop
        if state == Stopped {
            NotifyStop()
            break
        }
        
        sleep(5 * 1000)
    }
}
```



## Constraints

Notice that we don't support running loops inside another loop right now.



