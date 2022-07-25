package hosting

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func Test_HostBuilder_basic(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}
}

type MyConfig struct {
	value int
}

func Test_HostBuilder_HostConfiguration(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureHostConfiguration(func(configBuilder ConfigurationBuilder) {
		configBuilder.SetConfigurationFilePath("")
		configBuilder.SetConfigurationLoader(func(configFilePath string) interface{} {
			return &MyConfig{
				value: 123,
			}
		})
	})

	host := builder.Build()
	config := host.GetConfiguration().Get().(*MyConfig)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}

	provider := host.GetComponentProvider()
	config = dep.GetConfig[MyConfig](provider)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}
}

func Test_HostBuilder_HostConfigurationEx(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureHostConfigurationEx(func(hs HostSettings) interface{} {
		hs.SetName("Test")
		hs.SetRunningMode(Debug)
		return &MyConfig{
			value: 123,
		}
	})

	host := builder.Build()
	config := host.GetConfiguration().Get().(*MyConfig)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}

	provider := host.GetComponentProvider()
	config = dep.GetConfig[MyConfig](provider)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}
}

type MyAppConfig struct {
	value int
}

func Test_HostBuilder_AppConfiguration(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureAppConfiguration(func(context dep.Context, configBuilder ConfigurationBuilder) {
		configBuilder.SetConfigurationFilePath("")
		configBuilder.SetConfigurationLoader(func(configFilePath string) interface{} {
			return &MyAppConfig{
				value: 123,
			}
		})
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	config := dep.GetConfig[MyAppConfig](provider)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}
}

func Test_HostBuilder_AppConfigurationEx(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureAppConfigurationEx(func(hostCtxt dep.HostContext) interface{} {
		return &MyAppConfig{
			value: 123,
		}
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	config := dep.GetConfig[MyAppConfig](provider)
	if config == nil {
		t.Errorf("host config should not be nil")
	} else {
		if config.value != 123 {
			t.Errorf("host config value is not expected: %v", config.value)
		}
	}
}

func Test_HostBuilder_UseDefaultComponentProvider(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.UseDefaultComponentProvider()

	host := builder.Build()
	provider := host.GetComponentProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}

func Test_HostBuilder_UseComponentProvider(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureHostConfiguration(func(configBuilder ConfigurationBuilder) {
		configBuilder.SetConfigurationFilePath("")
		configBuilder.SetConfigurationLoader(func(configFilePath string) interface{} {
			return &MyConfig{
				value: 123,
			}
		})
	})
	builder.ConfigureAppConfiguration(func(context dep.Context, configBuilder ConfigurationBuilder) {
		configBuilder.SetConfigurationFilePath("")
		configBuilder.SetConfigurationLoader(func(configFilePath string) interface{} {
			return &MyAppConfig{
				value: 456,
			}
		})
	})
	builder.UseComponentProvider(func(context BuilderContext, options *dep.ComponentProviderOptions) {
		config := context.GetHostConfiguration().(*MyConfig)
		appConfig := context.GetAppConfiguration()
		if config.value != 123 || appConfig != nil {
			// no allowed types will cause panic during registering system components
			options.AllowedComponentTypes = make([]dep.TypeConstraint, 0)
		}
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}

type LogConfig struct{}

func Test_HostBuilder_ConfigureLoggingEx(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureLoggingEx(func(context BuilderContext, loggingBuilder LoggingBuilder) {
		hostname := context.GetHostName()
		loggingBuilder.AddConfiguration(&LogConfig{})
		loggingBuilder.SetLoggingInitializer(func(config interface{}) {
			//fmt.Printf("init logging with host name %v", hostname)
			fmt.Printf("init logging with config of type %s, host name %v", types.Of(config).FullName(), hostname)
			logger.InitializeDefaultLogging(logger.GetDefaultLoggingConfiguration(true))
		})
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}

type FakeLoggerFactory struct {
}

func NewFakeLoggerFactory() *FakeLoggerFactory {
	return &FakeLoggerFactory{}
}
func (flf *FakeLoggerFactory) Initialize(debug bool) {
	logger.InitializeDefaultLogging(&logger.DefaultLoggingConfig{
		Name:  "Test",
		Level: zap.InfoLevel,
	})
}
func (flf *FakeLoggerFactory) GetDefaultLogger() logger.Logger {
	return logger.GetLogger("")
}
func (flf *FakeLoggerFactory) GetLogger(name string) logger.Logger {
	return logger.GetLogger(name)
}

func Test_HostBuilder_ConfigureLogging(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureLogging(func(context BuilderContext, factoryBuilder LoggerFactoryBuilder) {
		factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory { return NewFakeLoggerFactory() })
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}

func Test_HostBuilder_ConfigureLogging_not_exist(t *testing.T) {
	defer test.AssertPanicContent(t, "logger factory not registered: logger.LoggerFactory", "panic content not expected")

	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureLogging(func(context BuilderContext, factoryBuilder LoggerFactoryBuilder) {
		//factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory { return NewFakeLoggerFactory() })
	})

	host := builder.Build()
	provider := host.GetComponentProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}

func Test_HostBuilder_UseService_basic(t *testing.T) {
	hostName := "Test"
	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	UseService[MyService](builder, NewMyService)
}

func Test_HostBuilder_UseService_TypeNotInterface(t *testing.T) {
	defer test.AssertPanicContent(t, "specified service type is not interface", "panic content is not expected")

	hostName := "Test"
	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	UseService[*MyService](builder, NewMyService)
}

func Test_HostBuilder_UseService_TypeDuplicated(t *testing.T) {
	defer test.AssertPanicContent(t, "specified service type already exist, duplicated", "panic content is not expected")

	hostName := "Test"
	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	UseService[MyService](builder, NewMyService)
	UseService[MyService](builder, NewMyService)
}

func Test_HostBuilder_ConfigureServices(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureServices(func(hb HostBuilder) {
		hb.UseLoop("TestLoop", func(context ServiceContext, looper ConfigureLoopContext) {
			looper.UseFuncProcessor(func() {
				// do nothing
			})
		})
	})

	host := builder.Build()
	services := host.GetServices()
	if _, exist := services["Looper:TestLoop"]; !exist {
		t.Error("service not configured successfully")
	}
}

func Test_HostBuilder_ConfigureAppRunner(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureAppRunner(func(hostContext dep.HostContext, components dep.ComponentCollection) {
		syncRunner := dep.GetComponent[SyncAppRunner](hostContext)
		dep.RegisterSingleton[AppRunner](components, func() AppRunner { return syncRunner })
	})

	host := builder.Build()
	rawCtxt := host.GetRawContext()
	if rawCtxt == nil {
		t.Error("raw context should not be nil")
	}

	hostCtxt := host.GetContext()
	if hostCtxt == nil {
		t.Error("host context should not be nil")
	}

	services := host.GetServices()
	if len(services) != 0 {
		t.Error("service count is not expected")
	}

	host.Run()
}

func Test_HostBuilder_UseBasicSyncAppRunner(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.UseBasicSyncAppRunner()

	host := builder.Build()
	rawCtxt := host.GetRawContext()
	if rawCtxt == nil {
		t.Error("raw context should not be nil")
	}

	hostCtxt := host.GetContext()
	if hostCtxt == nil {
		t.Error("host context should not be nil")
	}

	services := host.GetServices()
	if len(services) != 0 {
		t.Error("service count is not expected")
	}

	host.Run()
}

func Test_HostBuilder_ConfigureAppRunner_already_exist(t *testing.T) {
	defer test.AssertPanicContent(t, "don't call UseXXXAppRunner and register your own AppRunner both", "panic content not expected")

	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterSingleton[AppRunner](components, func() AppRunner { return nil })
	})
	builder.ConfigureAppRunner(func(hostContext dep.HostContext, components dep.ComponentCollection) {
		syncRunner := dep.GetComponent[SyncAppRunner](hostContext)
		dep.RegisterSingleton[AppRunner](components, func() AppRunner { return syncRunner })
	})

	builder.Build()
}

func Test_HostBuilder_ConfigureAppRunner_not_exist(t *testing.T) {
	defer test.AssertPanicContent(t, "AppRunner is not registered during ConfigAppRunner", "panic content not expected")

	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureAppRunner(func(hostContext dep.HostContext, components dep.ComponentCollection) {
		// syncRunner := dep.GetComponent[SyncAppRunner](hostContext)
		// dep.RegisterSingleton[AppRunner](components, func() AppRunner { return syncRunner })
	})

	builder.Build()
}