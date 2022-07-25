package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type DependencyTracker interface {
	AddDependents(context ...ContextEx)

	GetParent() ScopeContextEx
	GetDependents() []ContextEx
}

type LoggerProvider interface {
	GetLogger() logger.Logger
	GetLoggerWithName(name string) logger.Logger
	GetLoggerFactory() logger.LoggerFactory
}

const ContextType_Scope = "Scope"
const ContextType_Component = "Component"
const ContextType_Host = "Host"
const ContextType_Service = "Service"
const ContextType_Loop = "Loop"

type ContextBase interface {
	Type() string
	Name() string
}
type ContextBaseEx interface {
	ContextBase

	IsDebug() bool
	// context dependencies
	AddDependency(types.DataType, ComponentGetter)
}

func AddDependency[T any](context ContextBaseEx, depGetter TypedGetter[T]) {
	context.AddDependency(types.Get[T](), func() any { return depGetter() })
}

type Context interface {
	ContextBase
	LoggerProvider
	// expose Ex interface below to enable customer creating component with properties
	ComponentProviderEx

	GetProperties() Properties
}

type ContextEx interface {
	Context
	ContextBaseEx

	GetTracker() DependencyTracker
	UpdateProperties(props Properties)

	GetScopeContext() ScopeContextEx
}

type ScopeContext interface {
	ContextBase

	ScopeId() string
	IsGlobal() bool
}

type ScopeContextEx interface {
	ScopeContext
	ContextBaseEx

	Initialize(scopeType types.DataType, scopeInst Scopable)

	IsConcurrencyEnabled() bool

	// get scope data
	GetScope() ScopeDataEx
	// get context of parent scope
	GetParent() ScopeContextEx
	// get contextual dependencies of ancester scopes recursively
	GetDependency(depType types.DataType) any
}

type ComponentContext interface {
	Context
}
type ComponentContextEx interface {
	ComponentContext
	ContextEx
}

type HostContext interface {
	ComponentContext

	SetComponentManager(compMgr ComponentManager)
}
type HostContextEx interface {
	HostContext
	ContextEx

	GetGlobalScope() ScopeContextEx
}

type FactoryAction func() (interface{}, ContextEx)

type ScopedCompRecord interface {
	Execute(createComponent FactoryAction) (interface{}, ContextEx, bool, bool)
}

type ScopeData interface {
	GetType() types.DataType
	GetTypeName() string
	GetScopeId() string

	// clear all entries in the scope
	Clear()
}

type ScopeDataEx interface {
	ScopeData

	Initialize(scopeType types.DataType, scopeInst Scopable)

	IsTypedScope() bool
	GetInstance() (types.DataType, Scopable)

	Match(types.DataType) bool

	// retrieve if entry exist, or insert new entry and return if not
	GetCompRecord(compType types.DataType) ScopedCompRecord
}

// Utility APIs for default logger name
func GetDefaultLoggerNameForComponentType(componentType types.DataType) string {
	loggerName := componentType.Name()
	if componentType.IsPtr() {
		loggerName = componentType.ElementType().Name()
	}
	return loggerName
}

func GetDefaultLoggerNameForComponent(instance interface{}) string {
	return GetDefaultLoggerNameForComponentType(types.Of(instance))
}

// stack by parent context
func PrintAncestorStack(context Context, scenario string) {
	scopeCtxt := context.(ContextEx).GetScopeContext()
	scenario = fmt.Sprintf("%s, context %s[%s]", scenario, context.Type(), context.Name())
	PrintScopeAncestorStack(scopeCtxt, scenario)
}
func PrintScopeAncestorStack(scopeCtxt ScopeContextEx, scenario string) {
	fmt.Printf("Ancestor Stack of %s:\n", scenario)
	for ; scopeCtxt != nil; scopeCtxt = scopeCtxt.GetParent() {
		fmt.Printf("\t%s:<%s>\n", scopeCtxt.Type(), scopeCtxt.ScopeId())
	}
	fmt.Println("\t<Root>")
}

// stack by first dependent(who triggers instance creating)
func PrintDependencyStack(context Context) {
	scenario := fmt.Sprintf("%s[%s]", context.Type(), context.Name())
	PrintDependencyStackForScenario(context, scenario)
}
func PrintDependencyStackForScenario(ctxt Context, scenario string) {
	context := ctxt.(ContextEx)
	fmt.Printf("Dependency Stack of %s:\n", scenario)
	for {
		props := context.GetProperties()
		scopeId := context.GetScopeContext().ScopeId()
		if props != nil {
			fmt.Printf("\t<%s> %s: %s %s\n", scopeId, context.Type(), context.Name(), props.String())
		} else {
			fmt.Printf("\t<%s> %s: %s\n", scopeId, context.Type(), context.Name())
		}

		// get the first dependenct which triggers the instantiation of the component
		tracker := context.GetTracker()
		ctxts := tracker.GetDependents()
		if len(ctxts) > 0 {
			context = ctxts[0]
		} else {
			break
		}
	}
	fmt.Println("\t<Root>")
}
