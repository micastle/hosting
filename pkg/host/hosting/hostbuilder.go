package hosting

import (
	"fmt"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ConfigureComponentProviderMethod func(context BuilderContext, options *dep.ComponentProviderOptions)

type ConfigureHostMethod func(configBuilder ConfigurationBuilder)
type LoadHostConfigurationMethod func(HostSettings) interface{}

type ConfigureLoggerFactoryMethod func(context BuilderContext, factoryBuilder LoggerFactoryBuilder)
type ConfigureLoggingMethod func(context BuilderContext, loggingBuilder LoggingBuilder)

type ConfigureAppMethod func(context dep.Context, configBuilder ConfigurationBuilder)
type LoadAppConfigurationMethod func(hostCtxt dep.HostContext) interface{}

type ConfigureComponentsMethod func(context BuilderContext, components dep.ComponentCollection)
type ConfigureServicesMethod func(HostBuilder)
type ConfigureAppRunnerMethod func(hostContext dep.HostContext, components dep.ComponentCollection)
type ConfigureLifecycleMethod func(ctxt dep.Context, appLifecycle ApplicationLifecycle)

type ServiceFactoryMethod func(context ServiceContext) Service
type FreeStyleServiceFactoryMethod interface{}

type HostBuilder interface {
	SetHostName(name string) HostBuilder
	UseDefaultComponentProvider() HostBuilder
	UseComponentProvider(configure ConfigureComponentProviderMethod) HostBuilder

	UseDefaultAppRunner() HostBuilder
	UseBasicSyncAppRunner() HostBuilder

	ConfigureHostConfiguration(configure ConfigureHostMethod) HostBuilder
	ConfigureHostConfigurationEx(configLoader LoadHostConfigurationMethod) HostBuilder

	ConfigureLogging(configure ConfigureLoggerFactoryMethod) HostBuilder
	ConfigureLoggingEx(configure ConfigureLoggingMethod) HostBuilder

	ConfigureAppConfiguration(configure ConfigureAppMethod) HostBuilder
	ConfigureAppConfigurationEx(configLoader LoadAppConfigurationMethod) HostBuilder

	ConfigureComponents(configure ConfigureComponentsMethod) HostBuilder
	ConfigureServices(configure ConfigureServicesMethod) HostBuilder
	ConfigureLifecycle(configure ConfigureLifecycleMethod) HostBuilder

	UseService(serviceType types.DataType, createService FreeStyleServiceFactoryMethod) HostBuilder
	UseLoop(name string, configure ConfigureLoopMethod) HostBuilder

	Build() Host
}

func UseService[T any](builder HostBuilder, createService FreeStyleServiceFactoryMethod) {
	builder.UseService(types.Get[T](), createService)
}

type DefaultHostBuilder struct {
	Logger                  logger.Logger
	HostName                string
	GlobalProps             dep.Properties
	ConfigComponentProvider ConfigureComponentProviderMethod
	ConfigHostConfiguration ConfigureHostMethod
	HostConfigLoader        LoadHostConfigurationMethod
	ConfigLogging           ConfigureLoggerFactoryMethod
	ConfigAppConfiguration  ConfigureAppMethod
	AppConfigLoader         LoadAppConfigurationMethod
	ConfigComponents        ConfigureComponentsMethod
	ConfigLifecycle         ConfigureLifecycleMethod
	Loopers                 map[string]*LooperSettings
	ConfigServices          map[interface{}]FreeStyleServiceFactoryMethod
	ConfigAppRunner         ConfigureAppRunnerMethod
}

func NewDefaultHostBuilder() *DefaultHostBuilder {
	return &DefaultHostBuilder{
		HostName:       "Default",
		Loopers:        make(map[string]*LooperSettings),
		ConfigServices: make(map[interface{}]FreeStyleServiceFactoryMethod),
	}
}

func (hb *DefaultHostBuilder) SetHostName(name string) HostBuilder {
	hb.HostName = name
	return hb
}

func (hb *DefaultHostBuilder) UseDefaultComponentProvider() HostBuilder {
	hb.ConfigComponentProvider = nil
	return hb
}

func (hb *DefaultHostBuilder) UseComponentProvider(configure ConfigureComponentProviderMethod) HostBuilder {
	hb.ConfigComponentProvider = configure
	return hb
}

func (hb *DefaultHostBuilder) UseDefaultAppRunner() HostBuilder {
	return hb.ConfigureAppRunner(
		func(hostContext dep.HostContext, components dep.ComponentCollection) {
			components.RegisterSingletonForTypes(
				NewBasicAsyncRunner,
				types.Get[AppRunner](),
				types.Get[AsyncAppRunner](),
			)
		})
}
func (hb *DefaultHostBuilder) UseBasicSyncAppRunner() HostBuilder {
	return hb.ConfigureAppRunner(func(ctxt dep.HostContext, components dep.ComponentCollection) {
		dep.RegisterSingleton[AppRunner](components, func() SyncAppRunner {
			return dep.GetComponent[SyncAppRunner](ctxt)
		})
	})
}

func (hb *DefaultHostBuilder) ConfigureAppRunner(configure ConfigureAppRunnerMethod) HostBuilder {
	hb.ConfigAppRunner = configure
	return hb
}

func (hb *DefaultHostBuilder) ConfigureHostConfiguration(configure ConfigureHostMethod) HostBuilder {
	hb.ConfigHostConfiguration = configure
	return hb
}
func (hb *DefaultHostBuilder) ConfigureHostConfigurationEx(configLoader LoadHostConfigurationMethod) HostBuilder {
	hb.HostConfigLoader = configLoader
	return hb
}
func (hb *DefaultHostBuilder) ConfigureLogging(configure ConfigureLoggerFactoryMethod) HostBuilder {
	hb.ConfigLogging = configure
	return hb
}

func (hb *DefaultHostBuilder) ConfigureLoggingEx(configure ConfigureLoggingMethod) HostBuilder {
	hb.ConfigLogging = func(context BuilderContext, factoryBuilder LoggerFactoryBuilder) {
		loggingBuilder := NewDefaultLoggingBuilder()
		configure(context, loggingBuilder)
		if loggingBuilder.loggingInitializer == nil || loggingBuilder.createLogger == nil {
			panic(fmt.Errorf("logging initializer or creator not configured"))
		}

		factoryBuilder.RegisterLoggerFactory(func(provider dep.ComponentProvider) logger.LoggerFactory {
			lf := logger.NewDefaultLoggerFactory()
			lf.SetLoggingInitializer(func() { loggingBuilder.loggingInitializer(loggingBuilder.configuration) })
			lf.SetLoggingCreator(loggingBuilder.createLogger)
			return lf
		})
	}
	return hb
}
func (hb *DefaultHostBuilder) ConfigureAppConfiguration(configure ConfigureAppMethod) HostBuilder {
	hb.ConfigAppConfiguration = configure
	return hb
}
func (hb *DefaultHostBuilder) ConfigureAppConfigurationEx(configLoader LoadAppConfigurationMethod) HostBuilder {
	hb.AppConfigLoader = configLoader
	return hb
}
func (hb *DefaultHostBuilder) ConfigureComponents(configure ConfigureComponentsMethod) HostBuilder {
	hb.ConfigComponents = configure
	return hb
}

func (hb *DefaultHostBuilder) ConfigureLifecycle(configure ConfigureLifecycleMethod) HostBuilder {
	hb.ConfigLifecycle = configure
	return hb
}

func (hb *DefaultHostBuilder) ConfigureServices(configure ConfigureServicesMethod) HostBuilder {
	configure(hb)
	return hb
}
func (hb *DefaultHostBuilder) UseLoop(name string, configure ConfigureLoopMethod) HostBuilder {
	hb.Loopers[name] = &LooperSettings{
		Name:      name,
		Interval:  time.Duration(60) * time.Second,
		Recover:   true,
		Configure: configure,
	}
	return hb
}

func (hb *DefaultHostBuilder) addService(serviceType types.DataType, createService FreeStyleServiceFactoryMethod) {
	_, exist := hb.ConfigServices[serviceType.Key()]
	if exist {
		panic(fmt.Errorf("specified service type already exist, duplicated: %s", serviceType.FullName()))
	}

	hb.ConfigServices[serviceType.Key()] = createService
}
func (hb *DefaultHostBuilder) UseService(serviceType types.DataType, createService FreeStyleServiceFactoryMethod) HostBuilder {
	if !serviceType.IsInterface() {
		panic(fmt.Errorf("specified service type is not interface: %s", serviceType.FullName()))
	}

	// TODO: validate createService signature here before adding to service list

	hb.addService(serviceType, createService)

	return hb
}

func (hb *DefaultHostBuilder) buildHostConfiguration(context *HostBuilderContext) {
	if hb.ConfigHostConfiguration != nil {
		configBuilder := NewDefaultConfigurationBuilder()
		hb.ConfigHostConfiguration(configBuilder)

		if configBuilder.configLoader != nil {
			hb.HostConfigLoader = func(host HostSettings) interface{} {
				return configBuilder.configLoader(configBuilder.configFilePath)
			}
		}
	}

	var config interface{}
	if hb.HostConfigLoader != nil {
		config = hb.HostConfigLoader(NewDefaultHostSettings(context))
		if config != nil {
			configType := types.Of(config)
			if !configType.IsPtr() {
				panic(fmt.Errorf("type of configured host config is not pointer type: %s", configType.FullName()))
			}
		}
	}

	fmt.Printf("Host \"%s\": running mode - %v [mem stats: %v]\n", context.HostName, context.RunningMode, context.EnableMemoryStatistics)

	context.Configuration = NewDefaultConfiguration(config)
}
func (hb *DefaultHostBuilder) buildComponentManager(context *DefaultHostContext) {
	options := dep.NewComponentProviderOptions(dep.InterfaceType)
	if context.builderContext.RunningMode == Debug {
		options.EnableDiagnostics = true
	}

	if hb.ConfigComponentProvider != nil {
		builderContext := NewBuilderContext(context.builderContext)
		hb.ConfigComponentProvider(builderContext, options)
	}

	context.builderContext.ComponentManager = dep.NewDefaultComponentManager(context, options)
}
func (hb *DefaultHostBuilder) registerHostComponents(context *HostBuilderContext) {
	// register Host Configuration
	config := context.Configuration.Get()
	if config != nil {
		context.ComponentManager.AddConfiguration(config)
	}

	// register Component Manager as ContextualProvider
	//dep.AddSingleton[dep.ContextualProvider](context.ComponentManager, context.ComponentManager)
}
func (hb *DefaultHostBuilder) buildLoggerFactory(hostCtxt *DefaultHostContext, context *HostBuilderContext) {
	if hb.ConfigLogging == nil {
		hb.ConfigLogging = func(ctxt BuilderContext, factoryBuilder LoggerFactoryBuilder) {
			factoryBuilder.RegisterLoggerFactory(func(dep.ComponentProvider) logger.LoggerFactory { return logger.NewDefaultLoggerFactory() })
		}
	}

	factoryBuilder := NewDefaultLoggerFactoryBuilder(hostCtxt, context.ComponentManager)
	builderContext := NewBuilderContext(context)
	hb.ConfigLogging(builderContext, factoryBuilder)

	factoryType := types.Get[logger.LoggerFactory]()
	if !context.ComponentManager.IsComponentRegistered(factoryType) {
		panic(fmt.Errorf("logger factory not registered: %v", factoryType.FullName()))
	}
}
func (hb *DefaultHostBuilder) buildHostFromContext(context *DefaultHostContext) Host {
	host := NewHostFromContext(context)
	context.ComponentCollection.RegisterSingletonForTypes(
		func() *DefaultGenericHost { return host },
		types.Get[Host](),
		types.Get[SyncAppRunner](),
		types.Get[HostAsyncOperator](),
	)
	return host
}
func (hb *DefaultHostBuilder) buildAppConfiguration(context *DefaultHostContext) {
	var config interface{}
	if hb.ConfigAppConfiguration != nil {
		configBuilder := NewDefaultConfigurationBuilder()
		hb.ConfigAppConfiguration(context, configBuilder)

		if configBuilder.configLoader != nil {
			hb.AppConfigLoader = func(hostCtxt dep.HostContext) interface{} {
				return configBuilder.configLoader(configBuilder.configFilePath)
			}
		}
	}

	if hb.AppConfigLoader != nil {
		config = hb.AppConfigLoader(context)
	}

	context.Application.Configuration = NewDefaultConfiguration(config)
}

func (hb *DefaultHostBuilder) buildLifecycleConfiguration(context *DefaultHostContext) *DefaultLifecycle {
	appLifecycle := NewDefaultLifecycle()
	if hb.ConfigLifecycle != nil {
		hb.ConfigLifecycle(context, appLifecycle)
	}
	return appLifecycle
}

func (hb *DefaultHostBuilder) registerAppComponents(context *DefaultHostContext) {
	// register App Configuration
	if context.Application.Configuration.Get() != nil {
		context.ComponentCollection.AddConfiguration(context.Application.Configuration.Get())
	}

	// register other application components
	if hb.ConfigComponents != nil {
		builderContext := NewBuilderContext(context.builderContext)
		hb.ConfigComponents(builderContext, context.ComponentCollection)
	}

	hb.Logger.Debugw("configured components complete", "count", context.ComponentCollection.Count())
}

func (hb *DefaultHostBuilder) registerService(context *DefaultHostContext, serviceType types.DataType, createService FreeStyleServiceFactoryMethod) {
	hb.Logger.Debug("Register service type: " + serviceType.FullName())
	RegisterServiceForType(context.ComponentManager,
		createService,
		serviceType,
		GetServiceContextFactory(context.ComponentManager, serviceType),
	)
}
func (hb *DefaultHostBuilder) registerServiceComponents(context *DefaultHostContext) {
	// register all service types
	for typeKey, createService := range hb.ConfigServices {
		hb.registerService(context, types.FromKey(typeKey), createService)
	}

	// register looper type if used
	if len(hb.Loopers) > 0 {
		hb.Logger.Debug("Register looper type: " + types.Get[Looper]().FullName())
		context.ComponentManager.AddComponent(createLooper, types.Get[Looper]())
		dep.RegisterTransient[ProcessorGroup](context.ComponentCollection, func(ctxt dep.Context) *DefaultProcessorGroup {
			return NewDefaultProcessorGroup(ctxt)
		})
	}

	// register generic components
	dep.RegisterTransient[FunctionProcessor](context.ComponentCollection, NewFunctionProcessor)

	// register platform specifics
	registerPlatformComponents(context.ComponentCollection)
}

func (hb *DefaultHostBuilder) registerAppRunner(context *DefaultHostContext) {
	// check if user register customized app runner
	if context.ComponentCollection.IsComponentRegistered(types.Get[AppRunner]()) {
		if hb.ConfigAppRunner != nil {
			panic(fmt.Errorf("don't call UseXXXAppRunner and register your own AppRunner both, choose either of the two approaches."))
		}
	} else {
		// default to async runner if not specified
		if hb.ConfigAppRunner == nil {
			hb.UseDefaultAppRunner()
		}
		hb.ConfigAppRunner(context, context.ComponentCollection)
		//double check component AppRunner is configured
		if !context.ComponentCollection.IsComponentRegistered(types.Get[AppRunner]()) {
			panic(fmt.Errorf("AppRunner is not registered during ConfigAppRunner"))
		}
	}
}
func (hb *DefaultHostBuilder) buildHostedServices(context *DefaultHostContext) {
	for typeKey, _ := range hb.ConfigServices {
		serviceName := "Service:" + types.FromKey(typeKey).FullName()

		hb.Logger.Debug("Building service: " + serviceName)
		service := context.GetComponent(types.FromKey(typeKey)).(Service)

		context.Services[serviceName] = service
	}

	for _, settings := range hb.Loopers {
		serviceName := "Looper:" + settings.Name

		hb.Logger.Debugw("Building looper", "name", settings.Name)
		looper := dep.GetComponent[Looper](context).(*DefaultLooper)
		looper.Initialize(settings)

		context.Services[serviceName] = looper
	}
}

func (hb *DefaultHostBuilder) Build() Host {
	//
	// Stage 0: create builder context and prepare host configuration, create host context - not fully initialized
	//
	builderContext := &HostBuilderContext{
		HostName:               hb.HostName,
		RunningMode:            Debug,
		EnableMemoryStatistics: false,
	}
	builderContext.Application.Configuration = NewDefaultConfiguration(nil)
	hb.buildHostConfiguration(builderContext)
	hostContext := NewHostContext(builderContext, hb.GlobalProps)

	//
	// Stage 1: prepare host level components: component manager and logger factory
	//
	hb.buildComponentManager(hostContext)
	hb.registerHostComponents(builderContext)

	hb.buildLoggerFactory(hostContext, builderContext)
	// host context ready: logger Factory is initialized ready here
	hostContext.Initialize()

	hb.Logger = hostContext.GetLoggerWithName(dep.GetDefaultLoggerNameForComponent(hb))
	// host created from host context
	host := hb.buildHostFromContext(hostContext)

	//
	// Stage 2: prepare app level components: app configuration, register user's components and services
	//
	hb.buildAppConfiguration(hostContext)

	hb.registerAppComponents(hostContext)
	hb.registerServiceComponents(hostContext)
	hb.registerAppRunner(hostContext)

	//
	// Stage 4: build hosted services and their dependencies with DI
	//
	hb.buildHostedServices(hostContext)

	//
	// Stage 5: prepare lifecycle for running the host
	//
	hostContext.Lifecycle = hb.buildLifecycleConfiguration(hostContext)

	//
	// Stage 6: Host ready, execute hooks
	//
	hostContext.Lifecycle.OnHostReady(hostContext)

	// print diagnostic info for registered components
	if hostContext.IsDebug() {
		builderContext.ComponentManager.PrintDiagnostics()
	}

	return host
}
