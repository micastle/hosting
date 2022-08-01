package dep

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

//
// NOTE:
// 	ComponentContext should not depends on logger factory, or cyclic dependency occurs, as creating logger factory component depends on it.
//
type DefaultComponentContext struct {
	debug              bool
	componentType      types.DataType
	depTracker         DependencyTracker
	parentContext      ScopeContextEx
	contextualProvider ContextualProvider
	props              Properties

	// local context dependencies
	localDeps DepDict[ComponentGetter]
}

func NewComponentContext(scopeCtxt ScopeContextEx, contextualProvider ContextualProvider, componentType types.DataType) *DefaultComponentContext {
	compCtxt := &DefaultComponentContext{
		debug:              scopeCtxt.IsDebug(),
		parentContext:      scopeCtxt,
		depTracker:         NewDependencyTracker(scopeCtxt),
		contextualProvider: contextualProvider,
		componentType:      componentType,
		props:              scopeCtxt.GetScope().CopyProperties(),
	}
	compCtxt.localDeps = NewDependencyDictionary[ComponentGetter]()
	return compCtxt
}

func (cc *DefaultComponentContext) GetTracker() DependencyTracker {
	return cc.depTracker
}

func (cc *DefaultComponentContext) IsDebug() bool {
	return cc.debug
}

func (cc *DefaultComponentContext) UpdateProperties(props Properties) {
	cc.props.Update(props)
}
func (cc *DefaultComponentContext) GetProperties() Properties {
	return NewPropertiesFrom(cc.props)
}

func (cc *DefaultComponentContext) GetScopeContext() ScopeContextEx {
	return cc.parentContext
}

func (cc *DefaultComponentContext) Type() string {
	return ContextType_Component
}

func (cc *DefaultComponentContext) Name() string {
	return cc.componentType.FullName()
}

func (cc *DefaultComponentContext) GetLoggerFactory() logger.LoggerFactory {
	componentType := types.Get[logger.LoggerFactory]()
	return cc.contextualProvider.GetOrCreateWithProperties(componentType, cc, nil).(logger.LoggerFactory)
}

func (cc *DefaultComponentContext) GetLogger() logger.Logger {
	loggerName := GetDefaultLoggerNameForComponentType(cc.componentType)
	return cc.GetLoggerWithName(loggerName)
}

func (cc *DefaultComponentContext) GetLoggerWithName(name string) logger.Logger {
	return cc.GetLoggerFactory().GetLogger(name)
}

func (cc *DefaultComponentContext) AddDependency(depType types.DataType, depGetter ComponentGetter) {
	cc.localDeps.AddDependency(depGetter, depType)
}
func (cc *DefaultComponentContext) GetConfiguration(configType types.DataType) any {
	return cc.contextualProvider.GetConfiguration(configType, cc)
}

func (cc *DefaultComponentContext) GetComponent(interfaceType types.DataType) any {
	return cc.CreateWithProperties(interfaceType, nil)
}

func (cc *DefaultComponentContext) getContextDependency(depType types.DataType) any {
	// check local dict first
	if cc.localDeps.ExistDependency(depType) {
		return cc.localDeps.GetDependency(depType)()
	}
	// check ancester scopes recursively
	return cc.parentContext.GetDependency(depType)
}
func (cc *DefaultComponentContext) CreateWithProperties(interfaceType types.DataType, props Properties) any {
	// match dependency of current context including ancestor scopes
	inst := cc.getContextDependency(interfaceType)
	if inst != nil {
		return inst
	}
	return cc.contextualProvider.GetOrCreateWithProperties(interfaceType, cc, props)
}

// utility API
func GetComponentContextFactory(ctxtProvider ContextualProvider, compType types.DataType) ContextFactoryMethod {
	return func(scopeCtxt ScopeContextEx) ContextEx {
		//PrintScopeAncestorStack(scopeCtxt, fmt.Sprintf("CreateCompContext[%s]", compType.Name()))
		compCtxt := NewComponentContext(scopeCtxt, ctxtProvider, compType)
		AddDependency[Context](compCtxt, Getter[Context](compCtxt))
		AddDependency[ComponentProviderEx](compCtxt, Getter[ComponentProviderEx](compCtxt))
		AddDependency[logger.Logger](compCtxt, func() logger.Logger { return compCtxt.GetLogger() })
		AddDependency[Properties](compCtxt, func() Properties { return compCtxt.GetProperties() })
		AddDependency[ScopeContext](compCtxt, func() ScopeContext { return compCtxt.GetScopeContext() })
		return compCtxt
	}
}
