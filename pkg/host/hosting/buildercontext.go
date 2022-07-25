package hosting

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
)

type ApplicationContext struct {
	Configuration Configuration
}

type HostBuilderContext struct {
	HostName    string
	RunningMode RunningMode

	Configuration    Configuration
	ComponentManager dep.ComponentManager

	Application ApplicationContext
}

type BuilderContext interface {
	GetHostName() string
	IsDebug() bool
	GetHostConfiguration() interface{}
	GetAppConfiguration() interface{}
}

type DefaultBuilderContext struct {
	hostBuilderContext *HostBuilderContext
}

func NewBuilderContext(context *HostBuilderContext) *DefaultBuilderContext {
	return &DefaultBuilderContext{
		hostBuilderContext: context,
	}
}

func (bc *DefaultBuilderContext) GetHostName() string {
	return bc.hostBuilderContext.HostName
}
func (bc *DefaultBuilderContext) IsDebug() bool {
	return bc.hostBuilderContext.RunningMode == Debug
}
func (bc *DefaultBuilderContext) GetHostConfiguration() interface{} {
	return bc.hostBuilderContext.Configuration.Get()
}

func (bc *DefaultBuilderContext) GetAppConfiguration() interface{} {
	return bc.hostBuilderContext.Application.Configuration.Get()
}
