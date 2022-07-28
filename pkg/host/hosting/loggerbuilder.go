package hosting

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

type InitializeLoggingMethod func(config interface{})

type LoggingBuilder interface {
	AddConfiguration(configuration interface{}) LoggingBuilder
	SetLoggingInitializer(loggingInitializer InitializeLoggingMethod)
	SetLoggerFactory(createLogger logger.LoggerFactoryMethod)
}

type DefaultLoggingBuilder struct {
	configuration      interface{}
	loggingInitializer InitializeLoggingMethod
	createLogger       logger.LoggerFactoryMethod
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
func (lb *DefaultLoggingBuilder) SetLoggerFactory(createLogger logger.LoggerFactoryMethod) {
	lb.createLogger = createLogger
}

type LoggerFactoryCreator func(provider dep.ComponentProvider) logger.LoggerFactory

type LoggerFactoryBuilder interface {
	RegisterLoggerFactory(LoggerFactoryCreator)
}

type DefaultLoggerFactoryBuilder struct {
	hostCtxt *DefaultHostContext
	cm dep.ComponentManager
}

func NewDefaultLoggerFactoryBuilder(hostCtxt *DefaultHostContext, cm dep.ComponentManager) *DefaultLoggerFactoryBuilder {
	return &DefaultLoggerFactoryBuilder{
		hostCtxt: hostCtxt,
		cm: cm,
	}
}

func (lfb *DefaultLoggerFactoryBuilder) RegisterLoggerFactory(createInstance LoggerFactoryCreator) {
	loggerFactory := createInstance(lfb.hostCtxt)
	dep.AddComponent[logger.LoggerFactory](lfb.cm, dep.Factorize(loggerFactory))
}
