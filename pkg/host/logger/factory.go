package logger

import (
	"fmt"
	"strings"
)

type LoggerFactory interface {
	Initialize(debug bool)
	GetDefaultLogger() Logger
	GetLogger(name string) Logger
}

type DefaultLoggerFactory struct {
	loggingInitializer func()
}

func NewDefaultLoggerFactory() *DefaultLoggerFactory {
	return &DefaultLoggerFactory{}
}

func (lf *DefaultLoggerFactory) SetLoggingInitializer(initializer func()) {
	lf.loggingInitializer = initializer
}

func (lf *DefaultLoggerFactory) Initialize(debug bool) {
	if lf.loggingInitializer == nil {
		lf.loggingInitializer = func() { InitializeDefaultLogging(GetDefaultLoggingConfiguration(debug)) }
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
	return GetLogger(loggerName)
}
