package avt

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ConfigureComponentsMethod func(context BuilderContext, components dep.ComponentCollection)

type Activator interface {
	GetProvider() dep.ComponentProvider
}

func GetComponent[T any](avt Activator) T {
	return dep.GetComponent[T](avt.GetProvider())
}

func CreateActivator(configureComponents ConfigureComponentsMethod) Activator {
	return buildActivator(true, "", nil, configureComponents, nil, nil)
}
func CreateActivatorEx(debug bool, name string, globalProps dep.Properties, configureComponents ConfigureComponentsMethod, loggerFactory logger.LoggerFactory) Activator {
	var configLoggerFactory ConfigureLoggerFactoryMethod
	if loggerFactory != nil {
		configLoggerFactory = func(context BuilderContext, factoryBuilder LoggerFactoryBuilder) {
			factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory { return loggerFactory })
		}
	}
	configCompProvider := func(context BuilderContext, options *dep.ComponentProviderOptions) {
		options.AllowTypeAnyFromFactoryMethod = false
		options.AllowedComponentTypes = []dep.TypeConstraint{dep.InterfaceType}
		options.EnableSingletonConcurrency = true
		options.TrackTransientRecurrence = false
		options.EnableDiagnostics = debug
	}
	return buildActivator(debug, name, globalProps, configureComponents, configLoggerFactory, configCompProvider)
}
func buildActivator(debug bool, name string, globalProps dep.Properties, configureComponents ConfigureComponentsMethod, configLoggerFactory ConfigureLoggerFactoryMethod, configCompProvider ConfigureComponentProviderMethod) Activator {
	hostName := types.Of(new(Activator)).Name()
	if name != "" {
		hostName = name
	}
	runningMode := Release
	if debug {
		runningMode = Debug
	}
	builder := newActivatorBuilder(hostName, runningMode, globalProps)

	builder.ConfigureLogging(configLoggerFactory)
	builder.ConfigureComponentProvider(configCompProvider)

	activator := builder.build()

	if configureComponents != nil {
		activator.configureComponents(configureComponents)
	}

	activator.Logger.Debugw("configured components complete", "count", activator.getRegisteredCount())

	return activator
}

type DefaultActivator struct {
	hostContext *DefaultHostContext
	LogFactory  logger.LoggerFactory
	Logger      logger.Logger
}

func newDefaultActivator(ctxt *DefaultHostContext) *DefaultActivator {
	logFactory := dep.GetComponent[logger.LoggerFactory](ctxt.ComponentProvider)
	host := &DefaultActivator{
		hostContext: ctxt,
		LogFactory:  logFactory,
	}
	host.Logger = logFactory.GetLogger(dep.GetDefaultLoggerNameForComponent(host))
	return host
}

func (da *DefaultActivator) configureComponents(configure ConfigureComponentsMethod) {
	builderContext := NewBuilderContext(da.hostContext.builderContext)
	configure(builderContext, da.hostContext.ComponentCollection)
}
func (da *DefaultActivator) getContext() dep.HostContextEx {
	return da.hostContext
}
func (da *DefaultActivator) getRegisteredCount() int {
	return da.hostContext.ComponentCollection.Count()
}

// public APIs
func (da *DefaultActivator) GetProvider() dep.ComponentProvider {
	return da.hostContext
}

// factory method to create activator from context
func newActivatorFromContext(context *DefaultHostContext) *DefaultActivator {
	return newDefaultActivator(context)
}
