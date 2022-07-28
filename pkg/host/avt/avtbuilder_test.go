package avt

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func Test_HostBuilder_ConfigureLogging_not_exist(t *testing.T) {
	defer test.AssertPanicContent(t, "logger factory not registered: logger.LoggerFactory", "panic content not expected")

	hostName := "Test"
	runningMode := Debug

	builder := newActivatorBuilder(hostName, runningMode)
	builder.ConfigureLogging(func(context BuilderContext, factoryBuilder LoggerFactoryBuilder) {
		hostname := context.GetHostName()
		debug := context.IsDebug()
		fmt.Printf("activator %s run in mode debug=%v\n", hostname, debug)
		//factoryBuilder.RegisterLoggerFactory(func() logger.LoggerFactory { return NewFakeLoggerFactory() })
	})

	activator := builder.build()
	provider := activator.GetProvider()
	if provider == nil {
		t.Errorf("provider is nil")
	} else {
		factory := dep.GetComponent[logger.LoggerFactory](provider)
		factory.GetDefaultLogger().Info("log one line")
	}
}
