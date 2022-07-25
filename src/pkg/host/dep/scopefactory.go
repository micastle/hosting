package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type None interface{}

var ScopeType_None types.DataType = types.Get[None]()

var ScopeType_Any types.DataType = types.Get[any]()

type Global interface{}

var ScopeType_Global types.DataType = types.Get[Global]()

type Scope interface {
	GetType() types.DataType
	GetTypeName() string
	GetScopeId() string
	GetScopeContext() ScopeContext
	Provider() ComponentProvider
}

type ScopeInitializer func(scope ScopeEx)

type ActionMethod func()

type ScopeEx interface {
	Scope
	Scopable

	Dispose()
	Initialize(scopeType types.DataType, scopeInst Scopable)

	BuildActionMethod(string, FreeStyleScopeActionMethod) ActionMethod
	Execute(string, FreeStyleScopeActionMethod)
}

type Scopable interface {
	Context() Context
}

type ScopeFactory interface {
	CreateScope(context Context) ScopeEx
	CreateTypedScope(scopeType types.DataType) ScopeEx
	CreateScopeFrom(scopeInst Scopable, scopeType types.DataType) ScopeEx
}

func CreateTypedScope[T Scopable](factory ScopeFactory) ScopeEx {
	return factory.CreateTypedScope(types.Get[T]())
}
func CreateScopeFrom[T Scopable](factory ScopeFactory, scopeInst T) ScopeEx {
	return factory.CreateScopeFrom(scopeInst, types.Get[T]())
}

type DefaultScopeFactory struct {
	context ContextEx
}

func NewScopeFactory(context Context) *DefaultScopeFactory {
	return &DefaultScopeFactory{context: context.(ContextEx)}
}

func (sf *DefaultScopeFactory) createScope(parent Context, scopeInit ScopeInitializer) ScopeEx {
	// get rid of parent context but create scope directly from contextual provider
	// this is to avoid get parent scope object, as it is contextual available in scope context
	provider := GetComponent[ContextualProvider](sf.context)
	scope := provider.GetOrCreateWithProperties(types.Get[Scope](), parent, nil).(ScopeEx)
	// this line returns the parent scope object, no new scope created
	//scope := parent.GetComponent(types.Of(new(Scope))).(ScopeEx)

	scopeInit(scope)

	ctxt := scope.GetScopeContext().(ScopeContextEx)
	AddDependency[Scope](ctxt, Getter[Scope](scope))

	return scope
}

func (sf *DefaultScopeFactory) CreateScope(context Context) ScopeEx {
	scopeInit := func(scope ScopeEx) {
		scope.Initialize(ScopeType_None, scope)
	}
	return sf.createScope(context, scopeInit)
}

var ScopableType types.DataType = types.Of(new(Scopable))

func (sf *DefaultScopeFactory) CreateTypedScope(scopeType types.DataType) ScopeEx {
	inst := sf.context.GetComponent(scopeType)
	if !types.Of(inst).CheckCompatible(ScopableType) {
		panic(fmt.Errorf("scope type %s must implement interface %s", scopeType.FullName(), ScopableType.FullName()))
	}
	scopeInst := inst.(Scopable)
	scopeInit := func(scope ScopeEx) {
		scope.Initialize(scopeType, scopeInst)
	}
	return sf.createScope(scopeInst.Context(), scopeInit)
}
func (sf *DefaultScopeFactory) CreateScopeFrom(scopeInst Scopable, scopeType types.DataType) ScopeEx {
	context := scopeInst.Context()
	Type := ScopeType_None
	if scopeType != nil {
		Type = scopeType
	}

	scopeInit := func(scope ScopeEx) {
		scope.Initialize(Type, scopeInst)
	}

	return sf.createScope(context, scopeInit)
}

type DefaultScope struct {
	context      Context
	scopeCtxt    ScopeContextEx
}

func NewScope(context Context) *DefaultScope {
	scopeCtxt := context.(ContextEx).GetScopeContext()
	return &DefaultScope{
		context:      context,
		scopeCtxt:    scopeCtxt,
	}
}
func (ds *DefaultScope) Initialize(scopeType types.DataType, scopeInst Scopable) {
	ds.scopeCtxt.Initialize(scopeType, scopeInst)
}

func (ds *DefaultScope) GetType() types.DataType {
	return ds.scopeCtxt.GetScope().GetType()
}
func (ds *DefaultScope) GetTypeName() string {
	return ds.scopeCtxt.Name()
}
func (ds *DefaultScope) GetScopeId() string {
	return ds.scopeCtxt.ScopeId()
}

func (ds *DefaultScope) Context() Context {
	return ds.context
}
func (ds *DefaultScope) GetScopeContext() ScopeContext {
	return ds.scopeCtxt
}
func (ds *DefaultScope) Provider() ComponentProvider {
	return ds.context
}

func (ds *DefaultScope) BuildActionMethod(actionName string, actionFunc FreeStyleScopeActionMethod) ActionMethod {
	lcCtrl := GetComponent[LifecycleController](ds.context)
	action := lcCtrl.BuildScopeActionMethod(actionFunc, ds)
	context := ds.context
	return func() {
		action(context, actionName)
	}
}

func (ds *DefaultScope) Execute(actionName string, actionFunc FreeStyleScopeActionMethod) {
	actionMethod := ds.BuildActionMethod(actionName, actionFunc)
	actionMethod()
}

func (ds *DefaultScope) Dispose() {
	ds.scopeCtxt.GetScope().Clear()
}

// API to mimic using sytax using in dotnet
func Using[T Scopable](inst T, scopeFunc FreeStyleScopeActionMethod) {
	factory := GetComponent[ScopeFactory](inst.Context())

	scope := factory.CreateScopeFrom(inst, types.Get[T]())
	defer scope.Dispose()

	scope.Execute("ScopeFunc", scopeFunc)
}
