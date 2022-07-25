package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type FreeStyleMethod interface{}

type DepInjector interface {
	Initialize(compProvider ComponentProviderEx, contextualDeps DepDictReader[ComponentGetter])
	BuildComponent(factoryMethod FreeStyleFactoryMethod, compType types.DataType) interface{}
	ExecuteActionFunc(processorMethod FreeStyleProcessorMethod, actionName string)
}

type DefaultDepInjector struct {
	context           Context
	componentProvider ComponentProviderEx
	contextualDeps    DepDictReader[ComponentGetter]
}

func NewDependencyInjector(context Context) *DefaultDepInjector {
	return &DefaultDepInjector{
		context: context,
	}
}

func (di *DefaultDepInjector) Initialize(compProvider ComponentProviderEx, contextualDeps DepDictReader[ComponentGetter]) {
	di.componentProvider = compProvider
	di.contextualDeps = contextualDeps
}

func (di *DefaultDepInjector) getDependency(depType types.DataType) interface{} {
	// config type is registered as struct type but used as pointer of struct
	if depType.IsPtr() {
		return di.componentProvider.GetConfiguration(depType.ElementType())
	}

	if di.contextualDeps != nil {
		if di.contextualDeps.ExistDependency(depType) {
			getter := di.contextualDeps.GetDependency(depType)
			return getter()
		}
	}

	return di.componentProvider.GetComponent(depType)
}

func (di *DefaultDepInjector) callMethodWithDepInjection(method FreeStyleMethod) []interface{} {
	// call method with dependency injection
	outputs := types.ToFunc(method).Call(func(index int, argType types.DataType) interface{} {
		return di.getDependency(argType)
	})

	return outputs
}

func (di *DefaultDepInjector) ExecuteActionFunc(processorMethod FreeStyleProcessorMethod, actionName string) {
	// call processor method with dependency injection
	outputs := di.callMethodWithDepInjection(processorMethod)
	if len(outputs) != 0 {
		panic(fmt.Errorf("action func for %s should not has any output: %v", actionName, len(outputs)))
	}
}

func (di *DefaultDepInjector) BuildComponent(factoryMethod FreeStyleFactoryMethod, compType types.DataType) interface{} {
	// call factory method with dependency injection to create the component
	outputs := di.callMethodWithDepInjection(factoryMethod)
	if len(outputs) != 1 {
		panic(fmt.Errorf("dependency(%s) factory returns not only one output: %v", compType.FullName(), len(outputs)))
	}

	return outputs[0]
}
