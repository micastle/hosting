package dep

import (
	"fmt"
	"sort"
	"strings"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type LifecycleOptions struct {
	EnableDiagnostics          bool
	EnableSingletonConcurrency bool
	TrackTransientRecurrence   bool
	MaxAllowedRecurrence       uint32
	PropertiesPassOver         bool
}

type FreeStyleActionMethod interface{}
type FreeStyleScopeActionMethod interface{}
type NamedActionMethod func(context Context, actionName string)
type GenericActionMethod func(context Context, actionName string, deps ...*Dependency[ComponentGetter])

type LifecycleController interface {
	BuildSingletonFactoryMethod(compTypes []types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod
	BuildScopedFactoryMethod(compType types.DataType, scopeType types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod
	BuildTransientFactoryMethod(compType types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod

	BuildActionMethod(processorFunc FreeStyleActionMethod) GenericActionMethod
	BuildScopeActionMethod(scopeFunc FreeStyleScopeActionMethod, scope ScopeEx) NamedActionMethod

	//BuildProcessorFactoryMethod(procType types.DataType, processorFunc FreeStyleProcessorMethod, createProcessor FuncProcFactoryMethod) FactoryMethod
}

type DefaultLifecycleController struct {
	context Context
	options *LifecycleOptions
}

func NewLifecycleController(context Context, options *LifecycleOptions) *DefaultLifecycleController {
	return &DefaultLifecycleController{
		context: context,
		options: options,
	}
}

func (lc *DefaultLifecycleController) buildDepInjector(compProvider ComponentProviderEx, contextualDeps DepDictReader[ComponentGetter]) DepInjector {
	injector := GetComponent[DepInjector](lc.context)
	injector.Initialize(compProvider, contextualDeps)
	return injector
}
func (lc *DefaultLifecycleController) getComponentFactoryMethod(factoryMethod FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) InternalFactoryMethod {
	options := lc.options
	return func(dependent ContextEx, scopeCtxt ScopeContextEx, compType types.DataType, props Properties) (interface{}, ContextEx) {
		compCtxt := createCtxt(scopeCtxt)
		TrackDependent(compCtxt, dependent)
		if options.PropertiesPassOver {
			compCtxt.UpdateProperties(dependent.GetProperties())
		}
		if props != nil {
			compCtxt.UpdateProperties(props)
		}
		injector := lc.buildDepInjector(compCtxt, nil)
		return injector.BuildComponent(factoryMethod, compType), compCtxt
	}
}

type TransientFactoryMethod func(dependent ContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool)

func (lc *DefaultLifecycleController) createRecurrenceManager(compType types.DataType) RecurrenceManager {
	return NewRecurrenceManager(lc.options, compType)
}
func (lc *DefaultLifecycleController) getTransientFactoryMethod(compType types.DataType, createInstance InternalFactoryMethod) TransientFactoryMethod {
	recurMgr := lc.createRecurrenceManager(compType)
	return func(dependent ContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool) {
		tracker := recurMgr.GetTracker()
		return tracker.Execute(func() (interface{}, ContextEx) {
			return createInstance(dependent, dependent.GetScopeContext(), interfaceType, props)
		})
	}
}
func (lc *DefaultLifecycleController) BuildTransientFactoryMethod(compType types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod {
	createComponent := lc.getComponentFactoryMethod(createInstance, createCtxt)
	createTransient := lc.getTransientFactoryMethod(compType, createComponent)
	return func(depCtxt Context, interfaceType types.DataType, props Properties) interface{} {
		dependent := depCtxt.(ContextEx)
		instance, compCtxt, cycled_detected := createTransient(dependent, interfaceType, props)
		if cycled_detected {
			lc.raiseTransientCyclicDependencyFailure(dependent, interfaceType)
		}
		TrackDependent(compCtxt, dependent)
		return instance
	}
}

func typesToString(compTypes []types.DataType) string {
	if len(compTypes) == 0 {
		return "<None>"
	} else if len(compTypes) == 1 {
		return compTypes[0].FullName()
	}

	names := make([]string, 0, len(compTypes))
	for _, val := range compTypes {
		names = append(names, val.Name())
	}
	sort.Strings(names)

	return "<" + strings.Join(names, "|") + ">"
}

type SingletonFactoryMethod func(dependent ContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool, bool)

func (lc *DefaultLifecycleController) getSingletonFactoryMethod(compTypes []types.DataType, createInstance InternalFactoryMethod) SingletonFactoryMethod {
	// get from global scope data? eaiser by create it directly here as below
	scopedRecord := NewScopedCompRecord(typesToString(compTypes), lc.options.EnableSingletonConcurrency)
	return func(dependent ContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool, bool) {
		return scopedRecord.Execute(func() (interface{}, ContextEx) {
			return createInstance(dependent, dependent.GetScopeContext(), interfaceType, props)
		})
	}
}
func (lc *DefaultLifecycleController) BuildSingletonFactoryMethod(compTypes []types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod {
	createComponent := lc.getComponentFactoryMethod(createInstance, createCtxt)
	createSingleton := lc.getSingletonFactoryMethod(compTypes, createComponent)

	return func(depCtxt Context, interfaceType types.DataType, props Properties) interface{} {
		dependent := depCtxt.(ContextEx)
		// props should not be passed to factory method for singleton component
		instance, compCtxt, inst_exist, cycled_detected := createSingleton(dependent, interfaceType, nil)
		if cycled_detected {
			lc.raiseSingletonCyclicDependencyFailure(dependent, interfaceType)
		}
		if inst_exist {
			// track dependent context for existing instance
			TrackDependent(compCtxt, dependent)
		}
		return instance
	}
}

type ScopedFactoryMethod func(dependent ContextEx, scopeCtxt ScopeContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool, bool)

func (lc *DefaultLifecycleController) getScopedFactoryMethod(compType types.DataType, createInstance InternalFactoryMethod) ScopedFactoryMethod {
	return func(dependent ContextEx, scopeCtxt ScopeContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx, bool, bool) {
		scopedRecord := scopeCtxt.GetScope().GetCompRecord(interfaceType)
		return scopedRecord.Execute(func() (interface{}, ContextEx) {
			return createInstance(dependent, scopeCtxt, interfaceType, props)
		})
	}
}
func (lc *DefaultLifecycleController) BuildScopedFactoryMethod(compType types.DataType, scopeType types.DataType, createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod) FactoryMethod {
	createScoped := lc.getScopedFactoryMethod(compType, lc.getComponentFactoryMethod(createInstance, createCtxt))

	return func(depCtxt Context, interfaceType types.DataType, props Properties) interface{} {
		dependent := depCtxt.(ContextEx)
		//fmt.Printf("search scope type %s for component %s\n", scopeType.FullName(), interfaceType.Name())
		//PrintDependencyStack(depCtxt)
		//PrintAncestorStack(depCtxt, fmt.Sprintf("CreateScoped[%s]", interfaceType.Name()))
		// find target scope to get or create scoped component
		targetScope := lc.matchScopeForComponent(dependent.GetScopeContext(), interfaceType, scopeType)
		if targetScope == nil {
			lc.raiseScopeContextFailure(dependent, interfaceType, scopeType)
			panic(fmt.Errorf("scoped component %s must be used in scope of type %s", interfaceType.FullName(), scopeType.Name()))
		}
		//fmt.Printf("target scope for %s is %s\n", interfaceType.Name(), targetScope.ScopeId())

		// props should not be passed to factory method for singleton component
		instance, compCtxt, inst_exist, cycled_detected := createScoped(dependent, targetScope, interfaceType, nil)
		if cycled_detected {
			lc.raiseScopedCyclicDependencyFailure(dependent, targetScope, interfaceType)
		}
		if inst_exist {
			// track dependent context for existing instance
			TrackDependent(compCtxt, dependent)
		}
		return instance
	}
}
func (lc *DefaultLifecycleController) matchScopeForComponent(scopeCtxt ScopeContextEx, interfaceType types.DataType, scopeType types.DataType) ScopeContextEx {
	// global scope is a virtual scope which maps to Singleton, never match any requested Scoped type including "ScopeType_Any"
	if scopeCtxt.IsGlobal() {
		return nil
	}
	// match any scope
	if scopeType.Key() == ScopeType_Any.Key() {
		return scopeCtxt
	}

	for ; scopeCtxt != nil; scopeCtxt = scopeCtxt.GetParent() {
		//fmt.Printf("matching scope %s for scopeType %s requested by %s\n", scopeCtxt.ScopeId(), scopeType.Name(), interfaceType.Name())
		if scopeCtxt.GetScope().Match(scopeType) {
			return scopeCtxt
		}
	}
	return nil
}

func (lc *DefaultLifecycleController) buildActionDeps(context ContextEx, actionName string, deps ...*Dependency[ComponentGetter]) DepDict[ComponentGetter] {
	ctxtDeps := NewDependencyDictionary[ComponentGetter]()
	ctxtDeps.AddDependencies(
		DepInst[Context](context),
		DepFact[logger.Logger, ComponentGetter](func() any { return context.GetLoggerWithName(actionName) }),
	)
	ctxtDeps.AddDependencies(deps...)
	return ctxtDeps
}
func (lc *DefaultLifecycleController) getActionMethod(actionFunc FreeStyleProcessorMethod) GenericActionMethod {
	return func(owner Context, actionName string, deps ...*Dependency[ComponentGetter]) {
		context := owner.(ContextEx)
		ctxtDeps := lc.buildActionDeps(context, actionName, deps...)
		injector := lc.buildDepInjector(context, ctxtDeps)

		injector.ExecuteActionFunc(actionFunc, actionName)
	}
}

// func (lc *DefaultLifecycleController) BuildProcessorFactoryMethod(processorType types.DataType, processorFunc FreeStyleProcessorMethod, createProcessor FuncProcFactoryMethod) FactoryMethod {
// 	actionMethod := lc.getActionMethod(processorFunc)
// 	return func(context Context, procType types.DataType, props Properties) interface{} {
// 		return createProcessor(context, procType, actionMethod)
// 	}
// }

func (lc *DefaultLifecycleController) BuildActionMethod(processorFunc FreeStyleActionMethod) GenericActionMethod {
	return lc.getActionMethod(processorFunc)
}
func (lc *DefaultLifecycleController) BuildScopeActionMethod(scopeFunc FreeStyleScopeActionMethod, scope ScopeEx) NamedActionMethod {
	actionMethod := lc.getActionMethod(scopeFunc)
	scopeCtxt := scope.GetScopeContext()
	ctxtDeps := []*Dependency[ComponentGetter]{
		DepInst[ScopeContext](scopeCtxt),
		DepInst[Scope](scope),
	}

	return func(context Context, actionName string) {
		actionMethod(
			context,
			fmt.Sprintf("%s@Scope{%s}", actionName, scopeCtxt.ScopeId()),
			ctxtDeps...,
		)
	}
}

func (lc *DefaultLifecycleController) raiseScopeContextFailure(depCtxt ContextEx, interfaceType types.DataType, scopeType types.DataType) {
	fmt.Printf("scope mis-match detected: scoped component %s must be used in scope of type %s\n", interfaceType.FullName(), scopeType.Name())
	enableDiagnostics := lc.options.EnableDiagnostics
	if depCtxt.IsDebug() && enableDiagnostics {
		PrintAncestorStack(depCtxt, fmt.Sprintf("CreateScoped[%s]", interfaceType.Name()))
		panic(fmt.Errorf("scoped component %s must be used in scope of type %s", interfaceType.FullName(), scopeType.Name()))
	} else {
		panic(fmt.Errorf("scoped component %s must be used in scope of type %s, turn on EnableDiagnostics in Debug mode to show more details", interfaceType.FullName(), scopeType.Name()))
	}
}

func (lc *DefaultLifecycleController) raiseSingletonCyclicDependencyFailure(context ContextEx, interfaceType types.DataType) {
	lc.raiseCyclicDependencyFailure(context, interfaceType, "singleton")
}
func (lc *DefaultLifecycleController) raiseScopedCyclicDependencyFailure(context ContextEx, targetScope ScopeContextEx, interfaceType types.DataType) {
	lc.raiseCyclicDependencyFailure(context, interfaceType, fmt.Sprintf("scoped[%s]", targetScope.ScopeId()))
}
func (lc *DefaultLifecycleController) raiseCyclicDependencyFailure(context ContextEx, interfaceType types.DataType, lifeType string) {
	fmt.Printf("cyclic dependency detected on %s component: %s\n", lifeType, interfaceType.FullName())
	enableDiagnostics := lc.options.EnableDiagnostics
	if context.IsDebug() && enableDiagnostics {
		PrintDependencyStack(context)
		panic(fmt.Errorf("cyclic dependency detected on %s component %v", lifeType, interfaceType.FullName()))
	} else {
		panic(fmt.Errorf("cyclic dependency detected on %s component %v, turn on EnableDiagnostics in Debug mode to show more details", lifeType, interfaceType.FullName()))
	}
}

func (lc *DefaultLifecycleController) raiseTransientCyclicDependencyFailure(context ContextEx, interfaceType types.DataType) {
	fmt.Printf("recursive dependency overflow(MaxAllowedRecurrence=%d) on transient component: %s\n", lc.options.MaxAllowedRecurrence, interfaceType.FullName())
	if context.IsDebug() && lc.options.EnableDiagnostics {
		PrintDependencyStack(context)
		panic(fmt.Errorf("recursive dependency overflow on transient component %v", interfaceType.FullName()))
	} else {
		panic(fmt.Errorf("recursive dependency overflow on transient component %v, turn on EnableDiagnostics in Debug mode to show more details", interfaceType.FullName()))
	}
}
