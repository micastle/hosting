package dep

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ComponentScope interface {
}

type FuncProcFactoryMethod func(context Context, interfaceType types.DataType, actionFunc GenericActionMethod) interface{}

type ContextFactoryMethod func(scopeCtxt ScopeContextEx) ContextEx

type InternalFactoryMethod func(dependent ContextEx, scopeCtxt ScopeContextEx, interfaceType types.DataType, props Properties) (interface{}, ContextEx)

type ComponentManager interface {
	ContextualProvider
	ComponentCollection

	Initialize()
	GetOptions() *ComponentProviderOptions
	PrintDiagnostics()

	// test used only, wrapped call to ContextualProvider::GetOrCreateWithProperties
	GetComponent(interfaceType types.DataType, context Context) any

	// customized for component context creation
	AddSingletonWithContext(createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod, interfaceTypes ...types.DataType)

	// below API is for framwork internal usage
	// it does not validate output type of factory method compatible with interface type or not.
	AddSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType)
	AddTransientForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType)
	AddComponent(factoryMethod FactoryMethod, interfaceType types.DataType)

	//RegisterProcessorForType(processorFunc FreeStyleProcessorMethod, createProcessor FuncProcFactoryMethod, processorType types.DataType)
	//RegisterServiceForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType, createCtxt ServiceContextFactoryMethod)
}

func AddSingleton[T any](mgr ComponentManager, inst T) {
	mgr.AddComponent(Factorize(inst), types.Get[T]())
}
func AddComponent[T any](mgr ComponentManager, factoryMethod FactoryMethod) {
	mgr.AddComponent(factoryMethod, types.Get[T]())
}

// test only
func Test_GetComponent[T any](mgr ComponentManager, dependent Context) T {
	return mgr.GetComponent(types.Get[T](), dependent).(T)
}

type DefaultComponentManager struct {
	globalScope         ScopeContextEx
	context             Context
	options             *ComponentProviderOptions
	dependencies        DepDict[FactoryMethod]
	lifecycleController LifecycleController
}

func NewDefaultComponentManager(hostCtxt HostContextEx, options *ComponentProviderOptions) *DefaultComponentManager {
	cm := &DefaultComponentManager{
		globalScope: hostCtxt.GetGlobalScope(),
		options:     options,
	}
	// dependencies is pre-condition to register components
	cm.dependencies = NewDependencyDictionary[FactoryMethod]()

	hostCtxt.SetComponentManager(cm)

	AddSingleton[ContextualProvider](cm, cm)

	return cm
}

func (cm *DefaultComponentManager) Initialize() {
	options := cm.options

	// register internal dependencies
	cm.AddConfiguration(options)
	cm.AddConfiguration(&LifecycleOptions{
		EnableDiagnostics:          options.EnableDiagnostics,
		EnableSingletonConcurrency: options.EnableSingletonConcurrency,
		TrackTransientRecurrence:   options.TrackTransientRecurrence,
		MaxAllowedRecurrence:       options.MaxAllowedRecurrence,
		PropertiesPassOver:         options.PropertiesPassOver,
	})

	AddComponent[LifecycleController](cm, func(context Context, interfaceType types.DataType, props Properties) interface{} {
		compCtxt := NewComponentContext(cm.globalScope, cm, interfaceType)
		compCtxt.GetTracker().AddDependents(context.(ContextEx))
		lcOptions := GetConfig[LifecycleOptions](context)
		return LifecycleController(NewLifecycleController(compCtxt, lcOptions))
	})
	AddComponent[DepInjector](cm, func(context Context, interfaceType types.DataType, props Properties) interface{} {
			compCtxt := NewComponentContext(cm.globalScope, cm, interfaceType)
			compCtxt.GetTracker().AddDependents(context.(ContextEx))
			return DepInjector(NewDependencyInjector(compCtxt))
		})
	AddComponent[Scope](cm, func(depCtxt Context, interfaceType types.DataType, props Properties) interface{} {
		parentScope := depCtxt.(ContextEx).GetScopeContext()
		scopeCtxt := NewScopeContext(parentScope)
		compCtxt := NewComponentContext(scopeCtxt, cm, types.Of(new(Scope)))
		return ScopeEx(NewScope(compCtxt))
	})

	cm.context = NewComponentContext(cm.globalScope, cm, types.Of(new(ComponentManager)))
	cm.lifecycleController = GetComponent[LifecycleController](cm.context)

	RegisterTransient[ComponentHub](cm, NewComponentImplHub)
	RegisterSingleton[ScopeFactory](cm, NewScopeFactory)
}

func (cm *DefaultComponentManager) GetOptions() *ComponentProviderOptions {
	return cm.options
}
func (cm *DefaultComponentManager) Count() int {
	return cm.dependencies.Count()
}
func (cm *DefaultComponentManager) PrintDiagnostics() {
	fmt.Printf("Registered dependencies(%d)\n", cm.Count())
	if cm.options.EnableDiagnostics {
		deps := cm.dependencies.GetAllDeps()
		for _, dep := range deps {
			fmt.Printf("\tDependency Type: %s\n", dep.FullName())
		}
	}
}

func (cm *DefaultComponentManager) addComponent(getInstance FactoryMethod, componentType types.DataType) {
	if cm.IsComponentRegistered(componentType) {
		panic(fmt.Errorf("specified component type already exist: %s", componentType.FullName()))
	}

	cm.dependencies.AddDependency(getInstance, componentType)
}
func (cm *DefaultComponentManager) IsComponentRegistered(componentType types.DataType) bool {
	return cm.dependencies.ExistDependency(componentType)
}

func (cm *DefaultComponentManager) AddConfiguration(configuration interface{}) {
	if configuration == nil {
		panic(fmt.Errorf("specified configuration is nil"))
	}
	configType := types.Of(configuration).ElementType()
	cm.options.ValidateConfigurationTypeAllowed(configType)
	cm.addComponent(func(Context, types.DataType, Properties) interface{} {
		return configuration
	}, configType)
}

func (cm *DefaultComponentManager) validateFreeStyleFactoryMethod(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType) {
	funcType := types.GetFuncType(createInstance)
	if funcType.GetNumOfOutput() > 2 {
		panic(fmt.Errorf("factory method should return no more than 2 outputs"))
	}
	if funcType.GetNumOfOutput() >= 2 {
		errType := funcType.GetOutput(1)
		if !errType.CheckCompatible(types.Get[error]()) {
			panic(fmt.Errorf("second output of factory method should of type error, actual - %s", errType.FullName()))
		}
	}
	if funcType.GetNumOfOutput() >= 1 {
		// validate return type is compatible with the interface type to be registered
		outputType := funcType.GetOutputType()
		if outputType.IsAny() {
			if cm.options.AllowTypeAnyFromFactoryMethod {
				fmt.Printf("fatory method returns interface{}, allowed but not recommended, skip checking type compatibility\n")
			} else {
				panic(fmt.Errorf("not allowed to register factory method with return type interface{}"))
			}
		} else {
			for _, interfaceType := range interfaceTypes {
				if !outputType.CheckCompatible(interfaceType) {
					panic(fmt.Errorf("func output type(%s) is not compatible with interface type(%s) to be registered", outputType.FullName(), interfaceType.FullName()))
				}
			}
		}
	} else {
		panic(fmt.Errorf("factory method should return at least one output"))
	}
}

func (cm *DefaultComponentManager) RegisterSingletonForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cm.RegisterSingletonForTypes(createInstance, interfaceType)
}
func (cm *DefaultComponentManager) RegisterSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType) {
	for _, interfaceType := range interfaceTypes {
		cm.options.ValidateComponentTypeAllowed(interfaceType)
	}
	cm.validateFreeStyleFactoryMethod(createInstance, interfaceTypes...)
	cm.AddSingletonForTypes(createInstance, interfaceTypes...)
}
func (cm *DefaultComponentManager) AddSingletonForTypes(createInstance FreeStyleFactoryMethod, interfaceTypes ...types.DataType) {
	globalScope := cm.globalScope
	// use the first interface type as the singleton's component context type
	createCtxt := func(scopeCtxt ScopeContextEx) ContextEx {
		return GetComponentContextFactory(cm, interfaceTypes[0])(globalScope)
	}
	cm.addSingletonWithContext(createInstance, createCtxt, interfaceTypes...)
}
func (cm *DefaultComponentManager)AddSingletonWithContext(createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod, interfaceTypes ...types.DataType) {
	for _, interfaceType := range interfaceTypes {
		cm.options.ValidateComponentTypeAllowed(interfaceType)
	}
	cm.validateFreeStyleFactoryMethod(createInstance, interfaceTypes...)
	cm.addSingletonWithContext(createInstance, createCtxt, interfaceTypes...)
}
func (cm *DefaultComponentManager) addSingletonWithContext(createInstance FreeStyleFactoryMethod, createCtxt ContextFactoryMethod, interfaceTypes ...types.DataType) {
	factoryMethod := cm.lifecycleController.BuildSingletonFactoryMethod(interfaceTypes, createInstance, createCtxt)

	for _, interfaceType := range interfaceTypes {
		cm.addComponent(factoryMethod, interfaceType)
	}
}

func (cm *DefaultComponentManager) RegisterScopedForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cm.RegisterScopedForTypeEx(createInstance, interfaceType, ScopeType_Any)
}
func (cm *DefaultComponentManager) RegisterScopedForTypeEx(createInstance FreeStyleFactoryMethod, interfaceType types.DataType, scopeType types.DataType) {
	cm.options.ValidateComponentTypeAllowed(interfaceType)
	cm.validateFreeStyleFactoryMethod(createInstance, interfaceType)

	cm.AddScopedForType(createInstance, interfaceType, scopeType)
}
func (cm *DefaultComponentManager) AddScopedForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType, scopeType types.DataType) {
	createCtxt := GetComponentContextFactory(cm, interfaceType)
	factoryMethod := cm.lifecycleController.BuildScopedFactoryMethod(interfaceType, scopeType, createInstance, createCtxt)

	cm.addComponent(factoryMethod, interfaceType)
}

func (cm *DefaultComponentManager) RegisterTransientForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	cm.options.ValidateComponentTypeAllowed(interfaceType)
	cm.validateFreeStyleFactoryMethod(createInstance, interfaceType)
	cm.AddTransientForType(createInstance, interfaceType)
}
func (cm *DefaultComponentManager) AddTransientForType(createInstance FreeStyleFactoryMethod, interfaceType types.DataType) {
	createCtxt := GetComponentContextFactory(cm, interfaceType)
	factoryMethod := cm.lifecycleController.BuildTransientFactoryMethod(interfaceType, createInstance, createCtxt)
	cm.addComponent(factoryMethod, interfaceType)
}
func (cm *DefaultComponentManager) AddComponent(factoryMethod FactoryMethod, interfaceType types.DataType) {
	cm.addComponent(factoryMethod, interfaceType)
}

func validateInstanceType(instance any, componentType types.DataType) {
	if instance == nil {
		panic(fmt.Errorf("created component instance is nil, type: %v, quit", componentType.FullName()))
	}

	instanceType := types.Of(instance)
	// check instance type is equal to requested type
	if instanceType.Key() == componentType.Key() {
		return
	}

	// check instance type is compatible to requested type
	if instanceType.CheckCompatible(componentType) {
		return
	}

	panic(fmt.Errorf("created component instance type does not match, instance type: %v, requested type: %v", instanceType.FullName(), componentType.FullName()))
}
func (cm *DefaultComponentManager) resolveInstance(interfaceType types.DataType, dependent Context, props Properties) any {
	createInstance := cm.dependencies.GetDependency(interfaceType)
	return createInstance(dependent, interfaceType, props)
}
func (cm *DefaultComponentManager) GetConfiguration(configType types.DataType, dependent Context) any {
	cm.options.ValidateConfigurationTypeAllowed(configType)

	instance := cm.resolveInstance(configType, dependent, nil)
	validateInstanceType(instance, configType.PointerType())
	return instance
}
func (cm *DefaultComponentManager) GetOrCreateWithProperties(interfaceType types.DataType, dependent Context, props Properties) any {
	cm.options.ValidateComponentTypeAllowed(interfaceType)

	instance := cm.resolveInstance(interfaceType, dependent, props)
	validateInstanceType(instance, interfaceType)
	return instance
}

// test used only, wrapped call to ContextualProvider::GetOrCreateWithProperties
func (cm *DefaultComponentManager) GetComponent(componentType types.DataType, dependent Context) any {
	return cm.GetOrCreateWithProperties(componentType, dependent, nil)
}