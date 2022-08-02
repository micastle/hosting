package hosting

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ServiceContext interface {
	dep.ComponentContext
}
type ServiceContextEx interface {
	ServiceContext
	dep.ContextEx
}

type DefaultServiceContext struct {
	logger             logger.Logger
	parent             dep.ScopeContextEx
	depTracker         dep.DependencyTracker
	props              dep.Properties
	contextualProvider dep.ContextualProvider
	contextType        string
	name               string
	serviceType        types.DataType

	// local context dependencies
	localDeps dep.DepDict[dep.ComponentGetter]
}

func (sc *DefaultServiceContext) Type() string {
	return sc.contextType
}
func (sc *DefaultServiceContext) Name() string {
	return sc.name
}

func NewLoopContext(scopeCtxt dep.ScopeContextEx, contextualProvider dep.ContextualProvider, serviceType types.DataType) *DefaultServiceContext {
	return newDefaultServiceContext(dep.ContextType_Loop, scopeCtxt, contextualProvider, serviceType)
}

func NewServiceContext(scopeCtxt dep.ScopeContextEx, contextualProvider dep.ContextualProvider, serviceType types.DataType) *DefaultServiceContext {
	return newDefaultServiceContext(dep.ContextType_Service, scopeCtxt, contextualProvider, serviceType)
}

func newDefaultServiceContext(contextType string, scopeCtxt dep.ScopeContextEx, contextualProvider dep.ContextualProvider, serviceType types.DataType) *DefaultServiceContext {
	sc := &DefaultServiceContext{
		parent:             scopeCtxt,
		depTracker:         dep.NewDependencyTracker(scopeCtxt),
		contextualProvider: contextualProvider,
		contextType:        contextType,
		name:               serviceType.FullName(),
		serviceType:        serviceType,
		props:              scopeCtxt.GetScope().CopyProperties(),
	}
	sc.localDeps = dep.NewDependencyDictionary[dep.ComponentGetter]()

	sc.logger = sc.GetLoggerFactory().GetLogger(fmt.Sprintf("%sContext", contextType))
	sc.logger.Debugw("created context", "type", sc.Type(), "name", sc.Name())

	return sc
}

func (sc *DefaultServiceContext) IsDebug() bool {
	return sc.parent.IsDebug()
}
func (sc *DefaultServiceContext) GetScopeContext() dep.ScopeContextEx {
	return sc.parent
}

func (sc *DefaultServiceContext) GetTracker() dep.DependencyTracker {
	return sc.depTracker
}

func (sc *DefaultServiceContext) UpdateProperties(props dep.Properties) {
	if sc.props == nil {
		sc.props = props
	} else {
		sc.props.Update(props)
	}
}
func (sc *DefaultServiceContext) GetProperties() dep.Properties {
	return sc.props
}

func (sc *DefaultServiceContext) GetLoggerFactory() logger.LoggerFactory {
	componentType := types.Get[logger.LoggerFactory]()
	return sc.contextualProvider.GetOrCreateWithProperties(componentType, sc, nil).(logger.LoggerFactory)
}
func (sc *DefaultServiceContext) GetLogger() logger.Logger {
	return sc.GetLoggerWithName(dep.GetDefaultLoggerNameForComponentType(sc.serviceType))
}
func (sc *DefaultServiceContext) GetLoggerWithName(name string) logger.Logger {
	sc.logger.Debugw("inject logger", "type", sc.Type(), "service", sc.Name(), "name", name)
	return sc.GetLoggerFactory().GetLogger(name)
}

func (sc *DefaultServiceContext) AddDependency(depType types.DataType, depGetter dep.ComponentGetter) {
	sc.localDeps.AddDependency(depGetter, depType)
}
func (sc *DefaultServiceContext) GetConfiguration(configType types.DataType) any {
	sc.logger.Debugw("inject configuration", "type", configType.FullName(), "service", sc.Name())
	return sc.contextualProvider.GetConfiguration(configType, sc)
}

func (sc *DefaultServiceContext) GetComponent(interfaceType types.DataType) any {
	sc.logger.Debugw("inject component", "type", interfaceType.FullName(), "service", sc.Name())
	return sc.CreateWithProperties(interfaceType, nil)
}

func (sc *DefaultServiceContext) getContextDependency(depType types.DataType) any {
	// check local dict first
	if sc.localDeps.ExistDependency(depType) {
		return sc.localDeps.GetDependency(depType)()
	}
	// check ancester scopes recursively
	return sc.parent.GetDependency(depType)
}
func (sc *DefaultServiceContext) CreateWithProperties(interfaceType types.DataType, props dep.Properties) any {
	sc.logger.Debugw("inject component", "type", interfaceType.FullName(), "service", sc.Name())
	// match dependency of current context including ancestor scopes
	inst := sc.getContextDependency(interfaceType)
	if inst != nil {
		return inst
	}
	return sc.contextualProvider.GetOrCreateWithProperties(interfaceType, sc, props)
}

// Factory API for creating service context
func GetServiceContextFactory(contextualProvider dep.ContextualProvider, serviceType types.DataType) ServiceContextFactoryMethod {
	return func(scopeCtxt dep.ScopeContextEx) ServiceContextEx {
		serviceCtxt := NewServiceContext(scopeCtxt, contextualProvider, serviceType)
		dep.AddDependency[dep.Context](serviceCtxt, dep.Getter[dep.Context](serviceCtxt))
		dep.AddDependency[ServiceContext](serviceCtxt, dep.Getter[ServiceContext](serviceCtxt))
		dep.AddDependency[dep.ComponentProviderEx](serviceCtxt, dep.Getter[dep.ComponentProviderEx](serviceCtxt))
		dep.AddDependency[logger.Logger](serviceCtxt, func() logger.Logger { return serviceCtxt.GetLogger() })
		dep.AddDependency[dep.Properties](serviceCtxt, func() dep.Properties { return serviceCtxt.GetProperties() })
		dep.AddDependency[dep.ScopeContext](serviceCtxt, func() dep.ScopeContext { return serviceCtxt.GetScopeContext() })
		return serviceCtxt
	}
}
