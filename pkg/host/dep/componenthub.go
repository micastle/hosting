package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ComponentImplProvider[K comparable] interface {
	GetComponentType() types.DataType
	GetImplementation(key K) FactoryMethod
}

type CompImplCollection[K comparable] interface {
	GetComponentType() types.DataType
	AddImpl(key K, createInstance FreeStyleFactoryMethod)

	AddSingletonImpl(key K, createInstance FreeStyleFactoryMethod)
	AddTransientImpl(key K, createInstance FreeStyleFactoryMethod)
}

type ComponentHub[K comparable] interface {
	ComponentImplProvider[K]
	CompImplCollection[K]
}

type DefaultComponentImplHub[K comparable] struct {
	context             Context
	provider            ContextualProvider
	lifecycleController LifecycleController
	interfaceType       types.DataType
	impls               map[K]FactoryMethod
}

func NewComponentImplHub[K comparable](context Context, provider ContextualProvider, interfaceType types.DataType) *DefaultComponentImplHub[K] {
	return &DefaultComponentImplHub[K]{
		context:             context,
		provider:            provider,
		lifecycleController: GetComponent[LifecycleController](context),
		interfaceType:       interfaceType,
		impls:               make(map[K]FactoryMethod),
	}
}

func (ih *DefaultComponentImplHub[K]) GetComponentType() types.DataType {
	return ih.interfaceType
}

func (ih *DefaultComponentImplHub[K]) AddSingletonImpl(key K, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildSingletonFactoryMethod([]types.DataType{ih.interfaceType}, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub[K]) AddTransientImpl(key K, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildTransientFactoryMethod(ih.interfaceType, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub[K]) AddImpl(key K, createInstance FreeStyleFactoryMethod) {
	ih.AddTransientImpl(key, createInstance)
}
func (ih *DefaultComponentImplHub[K]) GetImplementation(key K) FactoryMethod {
	factoryMethod, exist := ih.impls[key]
	if !exist {
		panic(fmt.Errorf("component(%s) implementation not exist for key %v", ih.interfaceType.FullName(), key))
	}
	return factoryMethod
}
