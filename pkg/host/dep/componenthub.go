package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ComponentImpl struct {
	Key            interface{}
	Singleton      bool
	CreateInstance FreeStyleFactoryMethod
}

type ComponentImplProvider interface {
	GetComponentType() types.DataType
	GetImplementation(key interface{}) FactoryMethod
}

type CompImplCollection interface {
	GetComponentType() types.DataType
	AddImpl(key interface{}, createInstance FreeStyleFactoryMethod)

	AddSingletonImpl(key interface{}, createInstance FreeStyleFactoryMethod)
	AddTransientImpl(key interface{}, createInstance FreeStyleFactoryMethod)
}

type ComponentHub interface {
	SetComponentType(types.DataType)

	ComponentImplProvider
	CompImplCollection
}

type DefaultComponentImplHub struct {
	context             Context
	provider            ContextualProvider
	lifecycleController LifecycleController
	interfaceType       types.DataType
	impls               map[interface{}]FactoryMethod
}

func NewComponentImplHub(context Context, provider ContextualProvider) *DefaultComponentImplHub {
	return &DefaultComponentImplHub{
		context:             context,
		provider:            provider,
		lifecycleController: provider.GetOrCreateWithProperties(types.Get[LifecycleController](), context, nil).(LifecycleController),
		impls:               make(map[interface{}]FactoryMethod),
	}
}

func (ih *DefaultComponentImplHub) SetComponentType(interfaceType types.DataType) {
	ih.interfaceType = interfaceType
}
func (ih *DefaultComponentImplHub) GetComponentType() types.DataType {
	return ih.interfaceType
}

func (ih *DefaultComponentImplHub) AddSingletonImpl(key interface{}, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildSingletonFactoryMethod([]types.DataType{ih.interfaceType}, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub) AddTransientImpl(key interface{}, createInstance FreeStyleFactoryMethod) {
	ctxtFactory := GetComponentContextFactory(ih.provider, ih.interfaceType)
	factoryMethod := ih.lifecycleController.BuildTransientFactoryMethod(ih.interfaceType, createInstance, ctxtFactory)
	ih.impls[key] = factoryMethod
}
func (ih *DefaultComponentImplHub) AddImpl(key interface{}, createInstance FreeStyleFactoryMethod) {
	ih.AddTransientImpl(key, createInstance)
}
func (ih *DefaultComponentImplHub) GetImplementation(key interface{}) FactoryMethod {
	factoryMethod, exist := ih.impls[key]
	if !exist {
		panic(fmt.Errorf("component(%s) implementation not exist for key %v", ih.interfaceType.FullName(), key))
	}
	return factoryMethod
}
