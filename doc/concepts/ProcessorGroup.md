# Processor Group

Sample:

```go
hostBuilder.UseLoop("Health",
		func(context dep.ServiceContext, looper hosting.ConfigureLoopContext) {
			looper.SetInterval(time.Duration(10) * time.Second)
			looper.RegisterProcessor(types.Of(new(MyProc)))
			looper.RegisterConditionalProcessor(
				types.Of(new(MyProc)),
				func(context hosting.ScopeContext) bool { return true },
			)
			looper.RegisterProcessorGroup(
				func(context dep.Context, group hosting.GroupContext) {
					group.SetGroupName("TestGroup")
					group.RegisterProcessor(types.Of(new(MyProc)))
				},
				func(context hosting.ScopeContext) bool { return true },
			)
		},
	)
```

