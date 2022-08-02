package hosting

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ServiceContextFactoryMethod func(scopeCtxt dep.ScopeContextEx) ServiceContextEx

func RegisterServiceForType(compMgr dep.ComponentManager, createService dep.FreeStyleFactoryMethod, serviceType types.DataType, createServiceCtxt ServiceContextFactoryMethod) {
	createCtxt := func(scopeCtxt dep.ScopeContextEx) dep.ContextEx {
		return createServiceCtxt(scopeCtxt)
	}
	compMgr.AddSingletonWithContext(createService, createCtxt, serviceType)
}

func createFuncProcess[T FunctionProcessor](dependent dep.Context, processorFunc dep.FreeStyleProcessorMethod) T {
	lifecycleController := dep.GetComponent[dep.LifecycleController](dependent)
	actionMethod := lifecycleController.BuildActionMethod(processorFunc)
	return createFuncProcessor(dependent, types.Get[T](), actionMethod).(T)
}

func RegisterFuncProcessor[T FunctionProcessor](collection dep.ComponentCollection, processorFunc dep.FreeStyleProcessorMethod) {
	createProcessor := func(context dep.Context) T {
		return createFuncProcess[T](context, processorFunc)
	}
	collection.RegisterTransientForType(createProcessor, types.Get[T]())
}
