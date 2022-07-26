package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type FreeStyleMethod any

type DepInjector interface {
	Initialize(compProvider ComponentProviderEx, contextualDeps DepDictReader[ComponentGetter])
	BuildComponent(factoryMethod FreeStyleFactoryMethod, compType types.DataType) any
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

func (di *DefaultDepInjector) getDependency(depType types.DataType) any {
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

func (di *DefaultDepInjector) callMethodWithDepInjection(method FreeStyleMethod) []any {
	// call method with dependency injection
	outputs := types.ToFunc(method).Call(func(index int, argType types.DataType) any {
		return di.getDependency(argType)
	})

	return outputs
}

func (di *DefaultDepInjector) handleActionMethodOutputs(actionName string, outputs []any) {
	if len(outputs) > 1 {
		panic(fmt.Errorf("action %s should return no more than 1 outputs: %d", actionName, len(outputs)))
	}
	if len(outputs) >= 1 {
		if outputs[0] != nil {
			fmt.Printf("output value is not nil\n")
			err, ok := outputs[0].(error)
			if !ok {
				panic(fmt.Errorf("action %s should only return error or nothing, actual type: %s", actionName, types.Of(outputs[0]).FullName()))
			}
			if err != nil {
				panic(fmt.Errorf("action %s error: %v", actionName, err))
			}
		}
	}
}
func (di *DefaultDepInjector) ExecuteActionFunc(processorMethod FreeStyleProcessorMethod, actionName string) {
	// call processor method with dependency injection
	outputs := di.callMethodWithDepInjection(processorMethod)
	di.handleActionMethodOutputs(actionName, outputs)
}
func (di *DefaultDepInjector) handleFactoryMethodOutputs(compType types.DataType, outputs []any) any {
	if len(outputs) > 2 {
		panic(fmt.Errorf("dependency(%s) factory method should return no more than 2 outputs: %d", compType.FullName(), len(outputs)))
	}
	if len(outputs) >= 2 {
		if outputs[1] != nil {
			err := outputs[1].(error)
			if err != nil {
				panic(fmt.Errorf("dependency(%s) factory method error: %v", compType.FullName(), err))
			}
		}
	}
	if len(outputs) >= 1 {
		return outputs[0]
	} else {
		panic(fmt.Errorf("dependency(%s) factory method should return at least one output: %d", compType.FullName(), len(outputs)))
	}
}
func (di *DefaultDepInjector) BuildComponent(factoryMethod FreeStyleFactoryMethod, compType types.DataType) any {
	// call factory method with dependency injection to create the component
	outputs := di.callMethodWithDepInjection(factoryMethod)
	return di.handleFactoryMethodOutputs(compType, outputs)
}
