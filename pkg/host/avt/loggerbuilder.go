package avt

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
)

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
