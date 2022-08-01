package avt

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type DefaultHostContext struct {
	scopeCtxt      *dep.DefaultScopeContext
	depTracker     dep.DependencyTracker
	builderContext *HostBuilderContext
	logger         logger.Logger

	ComponentManager    dep.ComponentManager
	ComponentProvider   dep.ComponentProviderEx
	ComponentCollection dep.ComponentCollectionEx

	// local context dependencies
	localDeps dep.DepDict[dep.ComponentGetter]
}

func NewHostContext(builderContext *HostBuilderContext, props dep.Properties) *DefaultHostContext {
	debug := builderContext.RunningMode == Debug
	return &DefaultHostContext{
		scopeCtxt:      dep.NewGlobalScopeContext(debug, props),
		depTracker:     dep.NewDependencyTracker(nil),
		builderContext: builderContext,
	}
}

func (hc *DefaultHostContext) SetComponentManager(compMgr dep.ComponentManager) {
	hc.ComponentManager = compMgr
	hc.ComponentProvider = dep.NewDefaultComponentProvider(hc, hc.ComponentManager)
	hc.ComponentCollection = dep.NewComponentCollection(hc, hc.ComponentManager)
}

func (hc *DefaultHostContext) Initialize() {
	// init golbal scope
	options := hc.ComponentManager.GetOptions()
	hc.scopeCtxt.EnableConcurrency(options.EnableSingletonConcurrency)
	hc.scopeCtxt.Initialize(dep.ScopeType_Global, nil)

	// set host configuration
	//hc.Configuration = hc.builderContext.Configuration

	// This is the first time we get logger factory, should initialize it before using
	loggerFactory := hc.GetLoggerFactory()
	loggerFactory.Initialize(hc.Name(), hc.IsDebug())

	hc.logger = loggerFactory.GetLogger(dep.GetDefaultLoggerNameForComponent(hc))

	// initialize component manager
	hc.ComponentManager.Initialize()

	// init local deps for host context
	hc.localDeps = dep.NewDependencyDictionary[dep.ComponentGetter]()

	// done
	hc.logger.Debugw("host context initialized", "name", hc.Name())
}

func (hc *DefaultHostContext) GetScopeContext() dep.ScopeContextEx {
	return hc.scopeCtxt
}
func (hc *DefaultHostContext) GetGlobalScope() dep.ScopeContextEx {
	return hc.scopeCtxt
}

func (hc *DefaultHostContext) Type() string {
	return dep.ContextType_Host
}
func (hc *DefaultHostContext) Name() string {
	return hc.builderContext.HostName
}
func (hc *DefaultHostContext) IsDebug() bool {
	return hc.scopeCtxt.IsDebug()
}
func (hc *DefaultHostContext) GetRunningMode() RunningMode {
	return hc.builderContext.RunningMode
}
func (hc *DefaultHostContext) GetTracker() dep.DependencyTracker {
	return hc.depTracker
}
func (cc *DefaultHostContext) UpdateProperties(props dep.Properties) {
}
func (cc *DefaultHostContext) GetProperties() dep.Properties {
	return cc.scopeCtxt.GetScope().CopyProperties()
}

func (hc *DefaultHostContext) GetLoggerFactory() logger.LoggerFactory {
	return dep.GetComponent[logger.LoggerFactory](hc.ComponentProvider)
}
func (hc *DefaultHostContext) GetLogger() logger.Logger {
	return hc.GetLoggerWithName(hc.builderContext.HostName)
}
func (hc *DefaultHostContext) GetLoggerWithName(name string) logger.Logger {
	hc.logger.Debugw("inject logger", "type", hc.Type(), "host", hc.Name(), "name", name)
	return hc.GetLoggerFactory().GetLogger(name)
}

func (hc *DefaultHostContext) AddDependency(depType types.DataType, depFact dep.ComponentGetter) {
	hc.localDeps.AddDependency(depFact, depType)
}
func (hc *DefaultHostContext) GetConfiguration(configType types.DataType) interface{} {
	hc.logger.Debugw("inject configuration", "type", configType.FullName(), "host", hc.Name())
	return hc.ComponentProvider.GetConfiguration(configType)
}

func (hc *DefaultHostContext) GetComponent(interfaceType types.DataType) interface{} {
	hc.logger.Debugw("inject component", "type", interfaceType.FullName(), "host", hc.Name())
	return hc.CreateWithProperties(interfaceType, nil)
}
func (hc *DefaultHostContext) getContextDependency(depType types.DataType) interface{} {
	// check local dict first
	if hc.localDeps.ExistDependency(depType) {
		return hc.localDeps.GetDependency(depType)()
	}
	// check ancester scopes recursively
	return hc.scopeCtxt.GetDependency(depType)
}
func (hc *DefaultHostContext) CreateWithProperties(interfaceType types.DataType, props dep.Properties) interface{} {
	hc.logger.Debugw("inject component", "type", interfaceType.FullName(), "host", hc.Name())
	// match dependency of current context including ancestor scopes
	inst := hc.getContextDependency(interfaceType)
	if inst != nil {
		return inst
	}
	return hc.ComponentProvider.CreateWithProperties(interfaceType, props)
}
