package dep

import "goms.io/azureml/mir/mir-vmagent/pkg/host/types"

type ContextualProvider interface {
	GetConfiguration(configType types.DataType, dependent Context) interface{}
	GetOrCreateWithProperties(interfaceType types.DataType, dependent Context, props Properties) interface{}
}

func GetConfigFrom[T any](ctxtProvider ContextualProvider, dependent Context) *T {
	return ctxtProvider.GetConfiguration(types.Get[T](), dependent).(*T)
}
func GetComponentFrom[T any](ctxtProvider ContextualProvider, dependent Context, props Properties) T {
	return ctxtProvider.GetOrCreateWithProperties(types.Get[T](), dependent, props).(T)
}
