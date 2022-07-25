# Loop Processor

Sample code for registering a Loop Processor with a func:

```go
components.RegisterProcessorForType(
    func(context dep.Context, scopeCtxt hosting.ScopeContext, membershipDetector membership.MembershipDetector) {
		_ := membershipDetector.DetectMembership()
	},
    types.Of(new(MembershipProcessor)),
)
```

