package dep

import (
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type FactoryMethod func(context Context, interfaceType types.DataType, props Properties) any
type TypedFactoryMethod[T any] func(context Context, interfaceType types.DataType, props Properties) T

// required type: func(...) interface{} {}
type FreeStyleFactoryMethod interface{}

// required type: func(...) {}
type FreeStyleProcessorMethod interface{}

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

type Evaluator[K comparable] func(props Properties) K
type ConfigureImpls[T any, K comparable] func(CompImplCollection[T, K])

func RegisterComponent[T any, K comparable](components ComponentCollection, propsEval Evaluator[K], configure ConfigureImpls[T, K]) {
	collection := components.(ComponentCollectionEx)
	implHub := createComponentHub[T, K](collection)
	configure(implHub)
	addComponent(collection, func(dependent Context, interfaceType types.DataType, props Properties) T {
		context := dependent.(ContextEx)
		// build properties to evaluate, context of to be created component will do it again, dup?
		inheritProps := context.GetScopeContext().GetScope().CopyProperties()
		inheritProps.Update(props)
		key := propsEval(inheritProps)
		context.GetLogger().Debugf("Get implementation for type %s, key: %v from %s", interfaceType.FullName(), key, inheritProps.String())
		factoryMethod := implHub.GetImplementation(key)
		return factoryMethod(context, interfaceType, props).(T)
	})
}

type ComponentCollectionEx interface {
	ComponentCollection

	// register multi-impl component
	CreateComponentHub(func(context Context, provider ContextualProvider) interface{}) interface{}
	AddComponent(FactoryMethod, types.DataType)
}

func addComponent[T any](collection ComponentCollectionEx, createInstance TypedFactoryMethod[T]) {
	collection.AddComponent(func(context Context, interfaceType types.DataType, props Properties) any {
		return createInstance(context, interfaceType, props)
	}, types.Get[T]())
}

func createComponentHub[T any, K comparable](collection ComponentCollectionEx) ComponentHub[T, K] {
	implHub := collection.CreateComponentHub(func(context Context, provider ContextualProvider) any {
		return NewComponentImplHub[T, K](context, provider)
	}).(ComponentHub[T, K])
	return implHub
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
func (cc *DefaultComponentCollection) AddConfiguration(configuration any) {
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

func (cc *DefaultComponentCollection) CreateComponentHub(creator func(context Context, provider ContextualProvider) any) any {
	return creator(cc.context, cc.cm)
}
func (cc *DefaultComponentCollection) AddComponent(createInstance FactoryMethod, compType types.DataType) {
	cc.cm.AddComponent(createInstance, compType)
}

func (cc *DefaultComponentCollection) IsComponentRegistered(componentType types.DataType) bool {
	return cc.cm.IsComponentRegistered(componentType)
}

func (cc *DefaultComponentCollection) Count() int {
	return cc.cm.Count()
}
