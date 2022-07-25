package hosting

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type InitializeLoggingMethod func(config interface{})

type LoggingBuilder interface {
	AddConfiguration(configuration interface{}) LoggingBuilder
	SetLoggingInitializer(loggingInitializer InitializeLoggingMethod)
}

type DefaultLoggingBuilder struct {
	configuration      interface{}
	loggingInitializer InitializeLoggingMethod
}

func NewDefaultLoggingBuilder() *DefaultLoggingBuilder {
	return &DefaultLoggingBuilder{}
}

func (lb *DefaultLoggingBuilder) AddConfiguration(configuration interface{}) LoggingBuilder {
	lb.configuration = configuration
	return lb
}
func (lb *DefaultLoggingBuilder) SetLoggingInitializer(loggingInitializer InitializeLoggingMethod) {
	lb.loggingInitializer = loggingInitializer
}

type LoggerFactoryBuilder interface {
	RegisterLoggerFactory(func() logger.LoggerFactory)
}

type DefaultLoggerFactoryBuilder struct {
	cm dep.ComponentManager
}

func NewDefaultLoggerFactoryBuilder(cm dep.ComponentManager) *DefaultLoggerFactoryBuilder {
	return &DefaultLoggerFactoryBuilder{cm: cm}
}

func (lfb *DefaultLoggerFactoryBuilder) RegisterLoggerFactory(createInstance func() logger.LoggerFactory) {
	loggerFactory := createInstance()
	dep.AddComponent[logger.LoggerFactory](lfb.cm, dep.Factorize(loggerFactory))
}
