package hosting

import (
	"context"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type DefaultHostContext struct {
	RawContext     context.Context
	scopeCtxt      *dep.DefaultScopeContext
	depTracker     dep.DependencyTracker
	builderContext *HostBuilderContext
	logger         logger.Logger

	Configuration       Configuration
	ComponentManager    dep.ComponentManager
	ComponentProvider   dep.ComponentProviderEx
	ComponentCollection dep.ComponentCollectionEx
	Application         ApplicationContext
	Lifecycle           LifecycleHandler
	Services            map[string]Service
}

func NewHostContext(builderContext *HostBuilderContext) *DefaultHostContext {
	debug := builderContext.RunningMode == Debug
	return &DefaultHostContext{
		RawContext:     context.Background(),
		scopeCtxt:      dep.NewGlobalScopeContext(debug),
		depTracker:     dep.NewDependencyTracker(nil),
		builderContext: builderContext,
		Services:       make(map[string]Service),
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
	hc.Configuration = hc.builderContext.Configuration

	// This is the first time we get logger factory, should initialize it before using
	loggerFactory := hc.GetLoggerFactory()
	loggerFactory.Initialize(hc.Name(), hc.IsDebug())

	hc.logger = loggerFactory.GetLogger(dep.GetDefaultLoggerNameForComponent(hc))

	// initialize component manager
	hc.ComponentManager.Initialize()

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
	return nil
}

func (hc *DefaultHostContext) GetRawContext() context.Context {
	return hc.RawContext
}

func (hc *DefaultHostContext) GetLoggerFactory() logger.LoggerFactory {
	return dep.GetComponent[logger.LoggerFactory](hc.ComponentProvider)
}
func (hc *DefaultHostContext) GetLogger() logger.Logger {
	return hc.GetLoggerFactory().GetDefaultLogger()
}
func (hc *DefaultHostContext) GetLoggerWithName(name string) logger.Logger {
	return hc.GetLoggerFactory().GetLogger(name)
}

func (hc *DefaultHostContext) AddDependency(types.DataType, dep.ComponentGetter) {

}
func (hc *DefaultHostContext) GetConfiguration(configType types.DataType) interface{} {
	return hc.ComponentProvider.GetConfiguration(configType)
}

func (hc *DefaultHostContext) GetComponent(interfaceType types.DataType) interface{} {
	return hc.ComponentProvider.GetComponent(interfaceType)
}

func (hc *DefaultHostContext) CreateWithProperties(interfaceType types.DataType, props dep.Properties) interface{} {
	return hc.ComponentProvider.CreateWithProperties(interfaceType, props)
}
