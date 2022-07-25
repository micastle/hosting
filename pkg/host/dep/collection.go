package dep

import (
	"fmt"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type FactoryMethod func(context Context, interfaceType types.DataType, props Properties) any

// required type: func(...) interface{} {}
type FreeStyleFactoryMethod interface{}

// required type: func(...) {}
type FreeStyleProcessorMethod interface{}

type Evaluator func(props Properties) interface{}
type ConfigureComponentType func(CompImplCollection)

type ComponentCollection interface {
	AddConfiguration(configuration interface{})

	RegisterSingletonForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType)
	RegisterSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType)
	RegisterScopedForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType)
	RegisterScopedForTypeEx(createInstance FreeStyleFactoryMethod, interfaceType types.DataType, scopeType types.DataType)
	RegisterTransientForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType)

	IsComponentRegistered(componentType types.DataType) bool
	Count() int
}

func AddConfig[T any](collection ComponentCollection, config *T) {
	collection.AddConfiguration(config)
}
func RegisterInstance[T any](collection ComponentCollection, inst T) {
	collection.RegisterSingletonForType(Getter[T](inst), types.Get[T]())
}
func RegisterSingleton[T any](collection ComponentCollection, createInstance FreeStyleFactoryMethod) {
	collection.RegisterSingletonForType(createInstance, types.Get[T]())
}
func RegisterScoped[T any, S any](collection ComponentCollection, createInstance FreeStyleFactoryMethod) {
	collection.RegisterScopedForTypeEx(createInstance, types.Get[T](), types.Get[S]())
}
func RegisterTransient[T any](collection ComponentCollection, createInstance FreeStyleFactoryMethod) {
	collection.RegisterTransientForType(createInstance, types.Get[T]())
}
func IsComponentRegistered[T any](collection ComponentCollection) bool {
	return collection.IsComponentRegistered(types.Get[T]())
}

type ComponentCollectionEx interface {
	ComponentCollection

	// register multi-impl component
	RegisterComponent(interfaceType types.DataType, propEval Evaluator, configure ConfigureComponentType)

}

func RegisterComponent[T any](collection ComponentCollectionEx, propEval Evaluator, configure ConfigureComponentType) {
	collection.RegisterComponent(types.Get[T](), propEval, configure)
}

type DefaultComponentCollection struct {
	context Context
	cm      ComponentManager
}

func NewComponentCollection(context Context, cm ComponentManager) *DefaultComponentCollection {
	return &DefaultComponentCollection{
		context: context,
		cm:      cm,
	}
}
func (cc *DefaultComponentCollection) AddConfiguration(configuration interface{}) {
	cc.cm.AddConfiguration(configuration)
}
func (cc *DefaultComponentCollection) RegisterSingletonForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cc.RegisterSingletonForTypes(createInstance, interfaceType)
}
func (cc *DefaultComponentCollection) RegisterSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType) {
	cc.cm.RegisterSingletonForTypes(createInstance, interfaceTypes...)
}

func (cc *DefaultComponentCollection) RegisterScopedForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cc.RegisterScopedForTypeEx(createInstance, interfaceType, ScopeType_Any)
}
func (cc *DefaultComponentCollection) RegisterScopedForTypeEx(createInstance FreeStyleFactoryMethod, interfaceType types.DataType, scopeType types.DataType) {
	cc.cm.RegisterScopedForTypeEx(createInstance, interfaceType, scopeType)
}
func (cc *DefaultComponentCollection) RegisterTransientForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cc.cm.RegisterTransientForType(createInstance, interfaceType)
}

func (cc *DefaultComponentCollection) RegisterComponent(interfaceType types.DataType, propsEval Evaluator, configure ConfigureComponentType) {
	implHub := GetComponent[ComponentHub](cc.context)
	implHub.SetComponentType(interfaceType)
	configure(implHub)
	cc.cm.AddComponent(func(context Context, interfaceType types.DataType, props Properties) interface{} {
		key := propsEval(props)
		if key == nil {
			panic(fmt.Errorf("evaluated component implementation key should never be nil"))
		}
		context.GetLogger().Debugf("Get component implementation for type: %s %s", interfaceType.FullName(), props.String())
		factoryMethod := implHub.GetImplementation(key)
		return factoryMethod(context, interfaceType, props)
	}, interfaceType)
}
func (cc *DefaultComponentCollection) IsComponentRegistered(componentType types.DataType) bool {
	return cc.cm.IsComponentRegistered(componentType)
}

func (cc *DefaultComponentCollection) Count() int {
	return cc.cm.Count()
}
