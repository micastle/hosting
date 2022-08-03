package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ComponentImplProvider[T any, K comparable] interface {
	GetComponentType() types.DataType
	GetImplementation(key K) FactoryMethod
}

type CompImplCollection[T any, K comparable] interface {
	GetComponentType() types.DataType
	AddImpl(key K, createInstance FreeStyleFactoryMethod)

	AddSingletonImpl(key K, createInstance FreeStyleFactoryMethod)
	AddTransientImpl(key K, createInstance FreeStyleFactoryMethod)
}

type ComponentHub[T any, K comparable] interface {
	ComponentImplProvider[T, K]
	CompImplCollection[T, K]
}

type DefaultComponentImplHub[T any, K comparable] struct {
	context             Context
	provider            ContextualProvider
	lifecycleController LifecycleController
	interfaceType       types.DataType
	impls               map[K]FactoryMethod
}

func NewComponentImplHub[T any, K comparable](context Context, provider ContextualProvider) *DefaultComponentImplHub[T, K] {
	return &DefaultComponentImplHub[T, K]{
		context:             context,
		provider:            provider,
		lifecycleController: GetComponent[LifecycleController](context),
		interfaceType:       types.Get[T](),
		impls:               make(map[K]FactoryMethod),
	}
}

func (ih *DefaultComponentImplHub[T, K]) GetComponentType() types.DataType {
	return ih.interfaceType
}

func (ih *DefaultComponentImplHub[T, K]) AddSingletonImpl(key K, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildSingletonFactoryMethod([]types.DataType{ih.interfaceType}, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub[T, K]) AddTransientImpl(key K, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildTransientFactoryMethod(ih.interfaceType, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub[T, K]) AddImpl(key K, createInstance FreeStyleFactoryMethod) {
	ih.AddTransientImpl(key, createInstance)
}
func (ih *DefaultComponentImplHub[T, K]) GetImplementation(key K) FactoryMethod {
	factoryMethod, exist := ih.impls[key]
	if !exist {
		panic(fmt.Errorf("component(%s) implementation not exist for key %v", ih.interfaceType.FullName(), key))
	}
	return factoryMethod
}
