# Configuration

Configuration is commonly used in modern applications. The framework wraps configurations similar to dotNet framework but not exactly the same.

Configuration in hosting framework is also one kind of dependency. you can load them from anywhere as you want by your self, registering them to the framework and then use they anywhere as you want with the dependency injection mechanism.



## Configuration Loading and Registration

Configuration is organized into two levels: Host level and application level. You can use only host configuration if the application is simply enough.

### Host Configuration

Host configuration is the root configuration of your process. It is flexible that you can use it as the overall configuration for simple application, or you can use it to cover only fundamental configurations like application execution environment is local/onebox/production, configuration for linux/windows, logging settings, etc. while leaving actual configuration of your application components in specific configuration files. In this way you can have different application configurations for specific environment/platforms.

Below example shows how to configure configuration loader for your host. the returned configuration will be automatically registered to the framework, thus you can use it anywhere as needed.

```go
hostBuilder.ConfigureHostConfiguration(func(configBuilder hosting.ConfigurationBuilder) {
    configBuilder.SetConfigurationLoader(func(string) interface{} { return &Configuration{} })
  })
```

There is only one Host level configuration in the framework. One recommended way to use host configuration is to parse the configuration from command line arguments, which may provide application configuration file path, logging settings, etc.



### Application Configuration

as mentioned in above section, application configuration is optional and it should be used to cover specific configuration of the application and hosted services and components.

```go
hostBuilder.ConfigureAppConfiguration(func(hostContext dep.Context, configBuilder hosting.ConfigurationBuilder) {
		args := hostContext.GetConfiguration(types.Of(new(Arguments))).(*Arguments)
    	configBuilder.SetConfigurationLoader(func(string) interface{} { return &Configuration{} })
	})
```

during above application configuration API, you can get the basic utilities from host context, like logger, host configuration. then you can load your app configuration based on host configuration.

application configuration is also automatically registered to the framework.



## Configuration Injection

configuration injection automatically applies where dependency injection applies. see more details in [Dependency Injection](./DependencyInjection.md).



