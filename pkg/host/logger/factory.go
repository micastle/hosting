package logger

import (
	"fmt"
	"strings"
)

type LoggerFactory interface {
	Initialize(root string, debug bool)
	GetDefaultLogger() Logger
	GetLogger(name string) Logger
}

type LoggerFactoryMethod func(loggerName string) Logger

type DefaultLoggerFactory struct {
	loggingInitializer func()
	createLogger LoggerFactoryMethod
}

func NewDefaultLoggerFactory() *DefaultLoggerFactory {
	return &DefaultLoggerFactory{}
}

func (lf *DefaultLoggerFactory) SetLoggingInitializer(initializer func()) {
	lf.loggingInitializer = initializer
}
func (lf *DefaultLoggerFactory) SetLoggingCreator(createLogger LoggerFactoryMethod) {
	lf.createLogger = createLogger
}

func (lf *DefaultLoggerFactory) Initialize(name string, debug bool) {
	if lf.loggingInitializer == nil {
		lf.loggingInitializer = func() {
			config := GetDefaultLoggingConfig(debug)
			config.Name = name
			InitializeDefaultLogging(config)
		}
	}
	if lf.createLogger == nil {
		lf.createLogger = func(name string) Logger {
			return GetLogger(name)
		}
	}

	lf.loggingInitializer()
}

func (lf *DefaultLoggerFactory) GetDefaultLogger() Logger {
	return lf.getLogger("")
}

func (lf *DefaultLoggerFactory) GetLogger(name string) Logger {
	if len(strings.TrimSpace(name)) == 0 {
		panic(fmt.Errorf("whitespace only or empty logger name is not allowed"))
	}
	return lf.getLogger(name)
}

func (lf *DefaultLoggerFactory) getLogger(loggerName string) Logger {
	return lf.createLogger(loggerName)
}
