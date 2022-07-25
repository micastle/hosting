package dep

import "goms.io/azureml/mir/mir-vmagent/pkg/host/types"

type ComponentProvider interface {
	GetConfiguration(configType types.DataType) interface{}
	GetComponent(interfaceType types.DataType) interface{}
}

func GetConfig[T any](provider ComponentProvider) *T {
	return provider.GetConfiguration(types.Get[T]()).(*T)
}
func GetComponent[T any](provider ComponentProvider) T {
	return provider.GetComponent(types.Get[T]()).(T)
}

type ComponentProviderEx interface {
	ComponentProvider

	CreateWithProperties(interfaceType types.DataType, props Properties) interface{}
}

func CreateComponent[T any](provider ComponentProviderEx, props Properties) T {
	return provider.CreateWithProperties(types.Get[T](), props).(T)
}


type DefaultComponentProvider struct {
	context  Context
	provider ContextualProvider
}

func NewDefaultComponentProvider(context Context, provider ContextualProvider) *DefaultComponentProvider {
	return &DefaultComponentProvider{
		context:  context,
		provider: provider,
	}
}

func (cp *DefaultComponentProvider) GetConfiguration(configType types.DataType) interface{} {
	return cp.provider.GetConfiguration(configType, cp.context)
}

func (cp *DefaultComponentProvider) GetComponent(interfaceType types.DataType) interface{} {
	return cp.provider.GetOrCreateWithProperties(interfaceType, cp.context, nil)
}

func (cp *DefaultComponentProvider) CreateWithProperties(interfaceType types.DataType, props Properties) interface{} {
	return cp.provider.GetOrCreateWithProperties(interfaceType, cp.context, props)
}