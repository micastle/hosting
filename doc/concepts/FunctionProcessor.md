# Function Processor

Function Processor is a pre-define Loop Processor, which can take a func as the processor logic.

```go
type MyFuncProc hosting.FunctionProcessor

components.RegisterProcessorForType(
    func() {
        // DO anything you want here
    }, 
    types.Of(new(MyFuncProc)),
)
```

