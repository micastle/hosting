package avt

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
)

type RunningMode uint8

const (
	Debug RunningMode = iota
	Release
)

type BuilderContext interface {
	GetHostName() string
	IsDebug() bool
}

type HostBuilderContext struct {
	HostName    string
	RunningMode RunningMode

	ComponentManager dep.ComponentManager
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
