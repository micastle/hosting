package main

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
)

type LogConfig struct{}
type Configuration struct{ Log LogConfig }

func ConfigureHost() hosting.HostBuilder {
	hostBuilder := hosting.NewDefaultHostBuilder()
	hostBuilder.SetHostName("Sample")

	hostBuilder.ConfigureHostConfigurationEx(func(host hosting.HostSettings) interface{} {
		host.SetName("Sample")
		host.SetRunningMode(hosting.Debug)
		return &Configuration{}
	}).ConfigureAppConfigurationEx(
		func(hostCtxt dep.HostContext) interface{} {
			//hostConfig := hostCtxt.GetConfiguration(types.Of(new(Configuration))).(*Configuration)
			hostConfig := dep.GetConfig[Configuration](hostCtxt)
			return &hostConfig.Log
		},
	).ConfigureComponents(
		ConfigureComponents,
	).ConfigureServices(
		ConfigureServices,
	).ConfigureLifecycle(func(hostContext dep.Context, appLifecycle hosting.ApplicationLifecycle) {
		appLifecycle.RegisterOnStopEvent(func(ctxt dep.Context, se *hosting.StopEvent) bool {
			if se.Type != hosting.EVENT_TYPE_SIGNAL {
				ctxt.GetLogger().Errorf("unexpected stop event type: %v", se.Type)
			}
			// accept the stop event
			return true
		})
	})

	// Register your own logging factory
	// hostBuilder.ConfigureLogging(func(context hosting.BuilderContext, factoryBuilder hosting.LoggerFactoryBuilder) {
	// 	factoryBuilder.RegisterLoggerFactory(NewLoggerFactory)
	// })

	return hostBuilder
}
