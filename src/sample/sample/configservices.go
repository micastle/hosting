package main

import (
	"fmt"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

func ConfigureServices(hostBuilder hosting.HostBuilder) {
	hosting.UseService[WebAPIService](hostBuilder, func(context hosting.ServiceContext) WebAPIService {
		service := NewWebAPIService()
		err := service.Initialize(context)
		if err != nil {
			panic(fmt.Errorf("fail to initialize WebAPIService: %v", err.Error()))
		}
		return service
	})
	hosting.UseService[WorkerService](hostBuilder, func(context hosting.ServiceContext, scopeFactory dep.ScopeFactory) WorkerService {
		service := NewWorkerService(scopeFactory)
		err := service.Initialize(context, WorkerLoopRun)
		if err != nil {
			panic(fmt.Errorf("fail to initialize WorkerService: %v", err.Error()))
		}
		return service
	})
	hostBuilder.UseLoop("Main",
		func(context hosting.ServiceContext, looper hosting.ConfigureLoopContext) {
			looper.ConfigureLogger(func(context hosting.ServiceContext, log logger.Logger) logger.Logger {
				return logger.With(log, "Loop", "Main")
			})
			looper.SetInterval(time.Duration(5) * time.Second)
			hosting.UseProcessor[MyProc](looper, nil)
			hosting.UseProcessor[MyFuncProc](looper, nil)
			looper.UseFuncProcessor(func(context dep.Context, scopeCtxt hosting.ScopeContext, component Component) {
				context.GetLogger().Infof("UseFuncProcessor called, context of looper %s: %s-%s", scopeCtxt.GetLoopRunContext().LooperName(), context.Type(), context.Name())
				component.Print()
			})
		},
	)
	hostBuilder.UseLoop("Health",
		func(context hosting.ServiceContext, looper hosting.ConfigureLoopContext) {
			looper.SetInterval(time.Duration(10) * time.Second)
			//looper.UseProcessor(types.Of(new(MyProc)))
			hosting.UseProcessor[MyProc](looper, nil)
			hosting.UseProcessor[MyProc](looper, func(context hosting.ScopeContext) bool { return true })
			looper.UseProcessorGroup(
				func(context dep.Context, group hosting.GroupContext) {
					group.SetGroupName("TestGroup")
					//group.UseProcessor(types.Of(new(MyProc)))
					hosting.UseProcessor[MyProc](group, nil)
					group.UseFuncProcessor(func(context dep.Context, scopeCtxt hosting.ScopeContext, component Component) {
						context.GetLogger().Infof("UseFuncProcessor called in group, context of looper %s: %s-%s", scopeCtxt.GetLoopRunContext().LooperName(), context.Type(), context.Name())
						component.Print()
					})
				},
				func(context hosting.ScopeContext) bool {
					return true
				},
			)
		},
	)
}

func HandlerFunc(context dep.Context, scope dep.Scope, scopeCtxt dep.ScopeContext, logger logger.Logger, handler MyHandler, scopeComp ScopedComp, handlerComp HandleComp) {
	logger.Infof("HandlerFunc run once in scope: %s", scope.GetScopeId())
	logger.Infof("HandlerFunc got context: %s, %s", context.Type(), context.Name())

	dep.PrintAncestorStack(context, "HandlerFunc")

	fmt.Printf("scope id: %s\n", scope.GetScopeId())
	fmt.Printf("handler name: %s\n", handler.GetName())
	scopeComp.Print()
	handlerComp.Print()
}
func WorkerLoopRun(context dep.Context, scope dep.Scope, scopeCtxt dep.ScopeContext, iter LoopIteration, logger logger.Logger, scopeFactory dep.ScopeFactory, handler MyHandler, scopeComp ScopedComp, scopeComp2 ScopedComp) {
	logger.Infof("WorkerService loop iterate once in scope: %s", scope.GetScopeId())

	dep.PrintAncestorStack(context, "WorkerLoopRun")

	fmt.Printf("scope id: %s\n", scope.GetScopeId())
	fmt.Printf("iteration id: %s\n", iter.IterationId())
	scopeComp.Print()
	scopeComp2.Print()

	fmt.Printf("MyHandler scope instance created: %p\n", handler)
	//handlerScope := scopeFactory.CreateScopeFrom(handler, types.Of(new(MyHandler)))
	handlerScope := dep.CreateScopeFrom[MyHandler](scopeFactory, handler)
	//fmt.Printf("scope created, id: %s\n", handlerScope.GetScopeId())
	defer func() { handlerScope.Dispose() }()

	execute := handlerScope.BuildActionMethod("Handle", HandlerFunc)
	execute()
}
