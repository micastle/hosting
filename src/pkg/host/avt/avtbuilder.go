package avt

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ConfigureLoggerFactoryMethod func(context BuilderContext, factoryBuilder LoggerFactoryBuilder)
type ConfigureComponentProviderMethod func(context BuilderContext, options *dep.ComponentProviderOptions)

type DefaultActivatorBuilder struct {
	HostName    string
	runningMode RunningMode
	Logger      logger.Logger

	ConfigLogging           ConfigureLoggerFactoryMethod
	ConfigComponentProvider ConfigureComponentProviderMethod
}

func newActivatorBuilder(hostName string, runningMode RunningMode) *DefaultActivatorBuilder {
	return &DefaultActivatorBuilder{
		HostName:    hostName,
		runningMode: runningMode,
	}
}

func (ab *DefaultActivatorBuilder) ConfigureLogging(configure ConfigureLoggerFactoryMethod) {
	ab.ConfigLogging = configure
}
func (ab *DefaultActivatorBuilder) ConfigureComponentProvider(configure ConfigureComponentProviderMethod) {
	ab.ConfigComponentProvider = configure
}
func (ab *DefaultActivatorBuilder) build() *DefaultActivator {
	//
	// Stage 0: create host context and builder context, not fully initialized
	//
	builderContext := &HostBuilderContext{
		HostName:    ab.HostName,
		RunningMode: ab.runningMode,
	}
	hostContext := NewHostContext(builderContext)

	//
	// Stage 1: prepare host level components: host configuration, component manager and logger factory
	//
	//hb.buildHostConfiguration(builderContext)
	ab.buildComponentManager(hostContext)
	ab.registerHostComponents(builderContext)

	ab.buildLoggerFactory(builderContext)
	// host context ready: logger Factory is initialized ready here
	hostContext.Initialize()

	// init context deps
	dep.AddDependency[dep.Context](hostContext, dep.Getter[dep.Context](hostContext))
	dep.AddDependency[logger.Logger](hostContext, func() logger.Logger { return hostContext.GetLogger() })
	dep.AddDependency[dep.Properties](hostContext, func() dep.Properties { return hostContext.GetProperties() })
	dep.AddDependency[dep.ScopeContext](hostContext, func() dep.ScopeContext { return hostContext.GetScopeContext() })

	ab.Logger = hostContext.GetLoggerWithName(dep.GetDefaultLoggerNameForComponent(ab))
	// host created from host context
	return ab.buildHostFromContext(hostContext)
}

func (ab *DefaultActivatorBuilder) buildComponentManager(context *DefaultHostContext) {
	options := dep.NewComponentProviderOptions(dep.InterfaceType)
	if context.builderContext.RunningMode == Debug {
		options.EnableDiagnostics = true
	}

	if ab.ConfigComponentProvider != nil {
		builderContext := NewBuilderContext(context.builderContext)
		ab.ConfigComponentProvider(builderContext, options)
	}

	context.builderContext.ComponentManager = dep.NewDefaultComponentManager(context, options)
}

func (ab *DefaultActivatorBuilder) registerHostComponents(context *HostBuilderContext) {
	// register Component Manager as ContextualProvider
	context.ComponentManager.AddComponent(func(dep.Context, types.DataType, dep.Properties) interface{} {
		return context.ComponentManager
	}, types.Of(new(dep.ContextualProvider)))
}

func (ab *DefaultActivatorBuilder) buildLoggerFactory(context *HostBuilderContext) {
	if ab.ConfigLogging == nil {
		ab.ConfigLogging = func(ctxt BuilderContext, factoryBuilder LoggerFactoryBuilder) {
			factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory { return logger.NewDefaultLoggerFactory() })
		}
	}

	factoryBuilder := NewDefaultLoggerFactoryBuilder(context.ComponentManager)
	builderContext := NewBuilderContext(context)
	ab.ConfigLogging(builderContext, factoryBuilder)

	factoryType := types.Get[logger.LoggerFactory]()
	if !context.ComponentManager.IsComponentRegistered(factoryType) {
		panic(fmt.Errorf("logger factory not registered: %v", factoryType.FullName()))
	}
}

func (ab *DefaultActivatorBuilder) buildHostFromContext(context *DefaultHostContext) *DefaultActivator {
	avt := newActivatorFromContext(context)
	dep.RegisterInstance[Activator](context.ComponentCollection, avt)
	return avt
}
