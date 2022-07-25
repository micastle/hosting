package dep

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"go.uber.org/zap/zapcore"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

var initLoggerOnce sync.Once

func InitLoggerForTest() {
	initLoggerOnce.Do(func() {
		logger.InitStdOutLogger("UnitTest", zapcore.DebugLevel)
	})
}

// create global scope context for top component: host
func createTopScopeContext(context Context, debug bool, concurrency bool) *DefaultScopeContext {
	scopeCtxt := NewGlobalScopeContext(debug)
	scopeCtxt.EnableConcurrency(concurrency)
	scopeCtxt.Initialize(ScopeType_Global, nil)
	return scopeCtxt
}
func createNewTestScopeFrom(ctxt Context, scopeCtxt ScopeContextEx, scopeTypes ...types.DataType) ScopeContextEx {
	parent := scopeCtxt
	for _, t := range scopeTypes {
		newScopeCtxt := NewScopeContext(parent)
		newScopeCtxt.Initialize(t, nil)
		parent = newScopeCtxt
	}
	return parent
}
func createNewTestScope(hostCtxt HostContextEx, concurrency bool, scopeTypes ...types.DataType) ScopeContextEx {
	globalScope := createTopScopeContext(hostCtxt, hostCtxt.IsDebug(), concurrency)
	return createNewTestScopeFrom(hostCtxt, globalScope, scopeTypes...)
}
func initComponentManager(hostCtxt HostContext, options *ComponentProviderOptions) ComponentManager {
	InitLoggerForTest()
	cm := NewDefaultComponentManager(hostCtxt.(HostContextEx), options)
	cm.Initialize()

	loggerFactory := logger.NewDefaultLoggerFactory()
	AddSingleton[logger.LoggerFactory](cm, loggerFactory)

	return cm
}

func prepareComponentManagerWithScope(options *ComponentProviderOptions, scopeTypes ...types.DataType) (ComponentManager, ContextEx) {
	hostCtxt := NewMockContext(ContextType_Host, "Test", nil)
	globalScope := createTopScopeContext(hostCtxt, true, options.EnableSingletonConcurrency)
	hostCtxt.SetScopeContext(globalScope)

	cm := initComponentManager(hostCtxt, options)

	ctxt := NewMockContext(ContextType_Component, "Starter", nil)
	scopeCtxt := createNewTestScopeFrom(ctxt, globalScope, scopeTypes...)
	ctxt.SetScopeContext(scopeCtxt)
	ctxt.SetComponentManager(cm)

	return cm, ctxt
}
func prepareComponentManagerWithOptions(options *ComponentProviderOptions) (ComponentManager, ContextEx) {
	ctxt := NewMockContext(ContextType_Host, "Test", nil)
	scope := createNewTestScope(ctxt, options.EnableSingletonConcurrency)
	ctxt.SetScopeContext(scope)
	cm := initComponentManager(ctxt, options)
	return cm, ctxt
}
func prepareComponentManager(allowFactoryReturnAny bool) (ComponentManager, Context) {
	options := NewComponentProviderOptions(InterfaceType, StructPtrType)
	options.AllowTypeAnyFromFactoryMethod = allowFactoryReturnAny

	return prepareComponentManagerWithOptions(options)
}
func createNewContext(ctxt ContextEx, scopeType types.DataType) ContextEx {
	scope := ctxt.GetScopeContext()
	scope = createNewTestScopeFrom(ctxt, scope, scopeType)
	newCtxt := NewMockContext(ContextType_Component, scopeType.Name(), scope)
	newCtxt.GetTracker().AddDependents(ctxt)
	return newCtxt
}

type MyConfig struct {
	value int
}

type TestScope interface {
	Scopable
}

var ScopeTest types.DataType = types.Get[TestScope]()

type DefaultTestScope struct {
	context Context
}

func NewTestScope(context Context) *DefaultTestScope {
	return &DefaultTestScope{context: context}
}
func (ts *DefaultTestScope) Context() Context {
	return ts.context
}

func TestReflectType(t *testing.T) {
	var test FirstInterface = &ActualStruct{}
	t2 := reflect.TypeOf((*ActualStruct)(nil))
	t3 := reflect.TypeOf(test)
	t4 := reflect.TypeOf(new(FirstInterface))
	fmt.Println(t4)
	if t2 != t3 {
		t.Errorf("Type not equal for t2(%v) and t3(%v)!", t2.String(), t3.String())
	}
	if t3 == t4 {
		t.Errorf("Type are equal for t3(%v) and t4(%v)!", t3.String(), t4.String())
	}
}

func TestTypeConstraint(t *testing.T) {
	constraints := []TypeConstraint{InterfaceType, StructType, PointerType, InterfacePtrType, StructPtrType}
	names := []string{"Interface", "Struct", "Pointer", "InterfacePointer", "StructPointer"}
	for index, constraint := range constraints {
		if constraint.String() != names[index] {
			t.Errorf("name of TypeConstraint %v does not match with %v", constraint, names[index])
		}
	}
}

func Test_ComponentProviderOptions_ToString(t *testing.T) {
	constraints := []TypeConstraint{InterfacePtrType, StructPtrType}
	options := NewComponentProviderOptions(InterfacePtrType, StructPtrType)
	optionsStr := options.ToString(constraints)

	expected := "InterfacePointer,StructPointer"
	if optionsStr != expected {
		t.Errorf("ComponentProvider options string %v does not match with expected: %v", optionsStr, expected)
	}
}

func Test_ComponentProviderOptions_ValidateType_Allowed(t *testing.T) {
	if matchTypeConstraint(types.Of(new(MyConfig)), InterfacePtrType) {
		t.Errorf("struct pointer should not match interface ptr")
	}
	if !matchTypeConstraint(types.Of(new(MyConfig)), StructPtrType) {
		t.Errorf("struct pointer should match struct ptr")
	}
	if !matchTypeConstraint(types.Of(new(FirstInterface)), InterfacePtrType) {
		t.Errorf("interface pointer should match interface ptr")
	}
	if matchTypeConstraint(types.Of(new(FirstInterface)), StructPtrType) {
		t.Errorf("interface pointer should not match struct ptr")
	}
	if matchTypeConstraint(types.Of(new(MyConfig)), InterfaceType) {
		t.Errorf("struct pointer should not match interface ptr")
	}
	if matchTypeConstraint(types.Of(new(MyConfig)), StructType) {
		t.Errorf("struct pointer should not match interface ptr")
	}
	if !matchTypeConstraint(types.Of(new(FirstInterface)), PointerType) {
		t.Errorf("interface pointer should match ptr")
	}
	if !matchTypeConstraint(types.Of(new(MyConfig)), PointerType) {
		t.Errorf("struct pointer should match ptr")
	}
}

func Test_ComponentProviderOptions_ValidateConfig_NotAllowed(t *testing.T) {
	defer test.AssertPanicContent(t, "configuration type not allowed: dep.FirstInterface, allowed types: Struct", "panic content not expected")

	options := NewComponentProviderOptions(InterfaceType)
	options.ValidateConfigurationTypeAllowed(types.Get[FirstInterface]())
}
func Test_ComponentProviderOptions_ValidateComponent_NotAllowed(t *testing.T) {
	defer test.AssertPanicContent(t, "component type not allowed: *dep.MyConfig, allowed types: Interface", "panic content not expected")

	options := NewComponentProviderOptions(InterfaceType)
	options.ValidateComponentTypeAllowed(types.Get[*MyConfig]())
}

type MockContext struct {
	debug        bool
	ctxtType     string
	ctxtName     string
	componentMgr ComponentManager
	depTracker   DependencyTracker
	scopeCtxt    ScopeContextEx
	props        Properties
}

func NewMockContext(ctxtType string, ctxtName string, scopeCtxt ScopeContextEx) *MockContext {
	var tracker DependencyTracker = nil
	if scopeCtxt != nil {
		tracker = NewDependencyTracker(scopeCtxt)
	}
	return &MockContext{
		debug:      true,
		ctxtType:   ctxtType,
		ctxtName:   ctxtName,
		depTracker: tracker,
		scopeCtxt:  scopeCtxt,
		props:      NewProperties(),
	}
}
func (mc *MockContext) SetScopeContext(scopeCtxt ScopeContextEx) {
	mc.scopeCtxt = scopeCtxt
	mc.depTracker = NewDependencyTracker(scopeCtxt)
}
func (mc *MockContext) SetComponentManager(compMgr ComponentManager) {
	mc.componentMgr = compMgr
}

func (mc *MockContext) GetRawContext() context.Context {
	return nil
}

func (mc *MockContext) GetScopeContext() ScopeContextEx {
	return mc.scopeCtxt
}
func (mc *MockContext) GetGlobalScope() ScopeContextEx {
	return mc.scopeCtxt
}

func (mc *MockContext) Type() string {
	return mc.ctxtType
}
func (mc *MockContext) Name() string {
	return mc.ctxtName
}

func (mc *MockContext) GetLogger() logger.Logger {
	return mc.GetLoggerFactory().GetDefaultLogger()
}
func (mc *MockContext) GetLoggerFactory() logger.LoggerFactory {
	return GetComponentFrom[logger.LoggerFactory](mc.componentMgr, mc, nil)
}
func (mc *MockContext) GetLoggerWithName(name string) logger.Logger {
	return mc.GetLoggerFactory().GetLogger(name)
}

func (mc *MockContext) AddDependency(types.DataType, ComponentGetter) {

}
func (mc *MockContext) GetConfiguration(configType types.DataType) interface{} {
	return mc.componentMgr.GetConfiguration(configType, mc)
}

func (mc *MockContext) GetComponent(interfaceType types.DataType) interface{} {
	return mc.componentMgr.GetOrCreateWithProperties(interfaceType, mc, nil)
}

func (mc *MockContext) CreateWithProperties(interfaceType types.DataType, props Properties) interface{} {
	return mc.componentMgr.GetOrCreateWithProperties(interfaceType, mc, props)
}
func (mc *MockContext) GetProperties() Properties {
	return mc.props
}
func (mc *MockContext) UpdateProperties(props Properties) {
}
func (mc *MockContext) GetTracker() DependencyTracker {
	return mc.depTracker
}
func (mc *MockContext) IsDebug() bool {
	return mc.debug
}
func (mc *MockContext) GetScope() ScopeData {
	return mc.scopeCtxt.GetScope()
}

func TestComponentManager_Diagnostics(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.EnableDiagnostics = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	ops := cm.GetOptions()
	if ops.EnableDiagnostics != true {
		t.Error("diagnostics is not enabled")
	}

	cm.PrintDiagnostics()

	tracker := ctxt.GetTracker()
	scope := tracker.GetParent()
	if !scope.IsGlobal() {
		t.Error("tracked parent scope for host context should be global")
	}

	depCtxt := NewComponentContext(scope, cm, types.Of(new(AnotherInterface)))
	depCtxt.GetTracker().AddDependents(ctxt)

	dependents := depCtxt.GetTracker().GetDependents()
	for _, ddt := range dependents {
		fmt.Printf("Dependent: %s - %s\n", ddt.Type(), ddt.Name())
	}

	PrintDependencyStack(depCtxt)
}

func TestComponentManager_AddConfiguration(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	cm, ctxt := prepareComponentManagerWithOptions(options)
	config := &MyConfig{
		value: 123,
	}
	num := cm.Count()
	AddConfig[MyConfig](cm, config)

	configResult := GetConfigFrom[MyConfig](cm, ctxt)

	if config != configResult {
		t.Errorf("Get configuration result is different: %p - %p", config, configResult)
	}

	if config.value != configResult.value {
		t.Errorf("Get configuration's value is different: %v - %v", config.value, configResult.value)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered configuration type count is not 1: %v", num)
	}
}

func TestComponentManager_AddConfiguration_nil(t *testing.T) {
	defer test.AssertPanicContent(t, "specified configuration is nil", "panic content not expected")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	cm, _ := prepareComponentManagerWithOptions(options)

	var config *MyConfig = nil
	AddConfig[MyConfig](cm, config) // no panic

	var nilConfig interface{} = nil
	cm.AddConfiguration(nilConfig) // panic
}

func TestComponentManager_BuiltIns(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	num := cm.Count()

	options := GetConfigFrom[ComponentProviderOptions](cm, ctxt)
	if options == nil {
		t.Error("ComponentProviderOptions is not registered as builtin")
	} else {
		if !options.AllowTypeAnyFromFactoryMethod {
			t.Error("AllowTypeAnyFromFactoryMethod value not expected")
		}
	}

	lcCtrl := Test_GetComponent[LifecycleController](cm, ctxt)
	if lcCtrl == nil {
		t.Error("LifecycleController is not registered as a builtin component")
	}

	expected := 9
	if num != expected {
		t.Errorf("registered component type count is not %d: %d", expected, num)
	}
}

func TestComponentManager_RegisterSingleton(t *testing.T) {
	cm, ctxt := prepareComponentManager(true)
	num := cm.Count()

	op := &ActualStruct{}
	op1 := &ActualStruct{}
	RegisterInstance[FirstInterface](cm, op)
	RegisterInstance[*ActualStruct](cm, op1)

	interfaceResult := Test_GetComponent[FirstInterface](cm, ctxt)
	structResult := Test_GetComponent[*ActualStruct](cm, ctxt)

	if op != interfaceResult {
		t.Errorf("Get Interface instance failed! expected: %p, actual: %p", op, interfaceResult)
	}

	if op1 != structResult {
		t.Errorf("Get Struct instance failed! expected: %p, actual: %p", op1, structResult)
	}

	num = cm.Count() - num
	if num != 2 {
		t.Errorf("registered component type count is not 2: %v", num)
	}
}

func TestComponentManager_Singleton_OneInstance(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableSingletonConcurrency = false
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	inst1 := Test_GetComponent[AnotherInterface](cm, ctxt1)
	inst2 := Test_GetComponent[AnotherInterface](cm, ctxt1)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("singleton instances are the same one, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("singleton instances are different! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Singleton_OneInstance_concurrent(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	inst1 := Test_GetComponent[AnotherInterface](cm, ctxt1)
	inst2 := Test_GetComponent[AnotherInterface](cm, ctxt1)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("singleton instances are the same one, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("singleton instances are different! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Singleton_custom_context(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableSingletonConcurrency = false
	cm, ctxt := prepareComponentManagerWithOptions(options)

	compType := types.Get[AnotherInterface]()
	createCtxt := func(scopeCtxt ScopeContextEx) ContextEx {
		serviceCtxt := NewFakeServiceContext(cm, scopeCtxt)
		AddDependency[Context](serviceCtxt, Getter[Context](serviceCtxt))
		AddDependency[ComponentContext](serviceCtxt, Getter[ComponentContext](serviceCtxt))
		AddDependency[logger.Logger](serviceCtxt, func() logger.Logger { return serviceCtxt.GetLogger() })
		AddDependency[Properties](serviceCtxt, func() Properties { return serviceCtxt.GetProperties() })
		AddDependency[ScopeContext](serviceCtxt, func() ScopeContext { return serviceCtxt.GetScopeContext() })
		return serviceCtxt
	}

	cm.AddSingletonWithContext(NewAnotherStruct, createCtxt, compType)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	inst1 := Test_GetComponent[AnotherInterface](cm, ctxt1)
	ctxtName := inst1.GetContext().Name()
	if ctxtName != "Fake_Name" {
		t.Errorf("customized context is not created for singleton: %s", ctxtName)
	}

	dep := GetComponent[ComponentContext](inst1.GetContext())
	if dep.Type() != "Service" {
		t.Errorf("dependency of customized context is not injected for singleton: %s", dep.Type())
	}
}

func TestComponentManager_RegisterSingleton_duplicate(t *testing.T) {
	defer test.AssertPanicContent(t, "specified component type already exist:", "panic content not expected")

	cm, _ := prepareComponentManager(true)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
}
func TestComponentManager_RegisterSingleton_return_nil(t *testing.T) {
	defer test.AssertPanicContent(t, "created component instance is nil, type: dep.AnotherInterface", "panic content not expected")

	cm, ctxt := prepareComponentManager(true)

	RegisterSingleton[AnotherInterface](cm, func() interface{} { return nil})
	_ = GetComponent[AnotherInterface](ctxt)
}

func TestComponentManager_RegisterSingleton_return_bad_type(t *testing.T) {
	defer test.AssertPanicContent(t, "created component instance type does not match", "panic content not expected")

	cm, ctxt := prepareComponentManager(true)

	RegisterSingleton[AnotherInterface](cm, func() interface{} { return NewActualStruct() })
	_ = GetComponent[AnotherInterface](ctxt)
}

func TestComponentManager_Context_getContext(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, func(ctxt Context) interface{} {
		context := GetComponent[Context](ctxt)
		if context.Type() != ContextType_Component {
			t.Errorf("context type not expected: %s", context.Type())
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = Test_GetComponent[AnotherInterface](cm, ctxt1)
}

func TestComponentManager_Context_getLogger(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, func(ctxt Context) interface{} {
		logger := GetComponent[logger.Logger](ctxt)
		logger.Infof("get a logger from context!")
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = Test_GetComponent[AnotherInterface](cm, ctxt1)
}

func TestComponentManager_Context_getScopeCtxt(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, func(ctxt Context) interface{} {
		scopeCtxt := GetComponent[ScopeContext](ctxt)
		if !scopeCtxt.IsGlobal() {
			t.Error("not expected, context scope should be global scope in this test")
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = Test_GetComponent[AnotherInterface](cm, ctxt1)
}

func TestComponentManager_Context_getProperties_transient(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[AnotherInterface](cm, func(ctxt Context) interface{} {
		props := GetComponent[Properties](ctxt)
		name := GetProp[string](props, "name")
		if name != "test" {
			t.Errorf("not expected prop name: %s", name)
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	props := NewProperties()
	SetProp(props, "name", "test")
	_ = GetComponentFrom[AnotherInterface](cm, ctxt1, props)
}

func TestComponentManager_Context_getProperties_scoped(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[AnotherInterface, any](cm, func(ctxt Context) interface{} {
		props := GetComponent[Properties](ctxt)
		if props.Has("name") {
			t.Error("not expected, prop name should not be passed in for scoped component")
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	props := NewProperties()
	SetProp(props, "name", "test")
	_ = GetComponentFrom[AnotherInterface](cm, ctxt1, props)
}

func TestComponentManager_Context_getProperties_singleton(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, func(ctxt Context) interface{} {
		props := GetComponent[Properties](ctxt)
		if props.Has("name") {
			t.Error("not expected, prop name should not be passed in for singleton component")
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = GetComponentFrom[AnotherInterface](cm, ctxt1, Props(Pair("name", "test")))
}

func TestComponentManager_Context_getProperties_transient_passover_default(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[FirstInterface](cm, func(ctxt Context, log logger.Logger) interface{} {
		props := GetComponent[Properties](ctxt)
		log.Infof("get a props from context: %p", props)
		if props.Has("name") {
			t.Error("not expected, prop name should not be inherited by default")
		}
		return NewActualStruct()
	})
	RegisterTransient[AnotherInterface](cm, func(ctxt Context, log logger.Logger, first FirstInterface) interface{} {
		props := GetComponent[Properties](ctxt)
		log.Infof("get a props from context: %p", props)
		name := props.Get("name").(string)
		if name != "test" {
			t.Errorf("not expected prop name: %s", name)
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = GetComponentFrom[AnotherInterface](cm, ctxt1, Props(Pair("name", "test")))
}

func TestComponentManager_Context_getProperties_transient_passover_true(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.PropertiesPassOver = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[FirstInterface](cm, func(ctxt Context, log logger.Logger) interface{} {
		props := GetComponent[Properties](ctxt)
		log.Infof("get a props from context: %p", props)
		if !props.Has("name") {
			t.Error("not expected, prop name should be passed over if PropertiesPassOver=true")
		}
		return NewActualStruct()
	})
	RegisterTransient[AnotherInterface](cm, func(ctxt Context, log logger.Logger, first FirstInterface) interface{} {
		props := GetComponent[Properties](ctxt)
		log.Infof("get a props from context: %p", props)
		name := props.Get("name").(string)
		if name != "test" {
			t.Errorf("not expected prop name: %s", name)
		}
		return NewAnotherStruct(ctxt)
	})

	ctxt1 := createNewContext(ctxt, ScopeTest)
	_ = GetComponentFrom[AnotherInterface](cm, ctxt1, Props(Pair("name", "test")))
}

func TestComponentManager_Scoped_SameInsance(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	cm.RegisterScopedForType(NewAnotherStruct, types.Get[AnotherInterface]())

	ctxt1 := createNewContext(ctxt, ScopeType_None)
	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("scoped instances are the same one in the same scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are different in the same scope! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Scoped_DifferentInstance(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	ctxt1 := createNewContext(ctxt, ScopeType_None)
	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)
	ctxt2 := createNewContext(ctxt, ScopeType_None)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt2, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 != inst2 {
		t.Logf("scoped instances are different in different scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are the same in different scope! %p == %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Scoped_OutofScope(t *testing.T) {
	defer test.AssertPanicContent(t, "must be used in scope of type", "panic content should indict using scoped component out of scope")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	another := GetComponentFrom[AnotherInterface](cm, ctxt, nil)
	another.Another()

	t.Errorf("scoped component should not be created out of a scope")
}

func TestComponentManager_Scoped_InScopeAction(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeFactory := Test_GetComponent[ScopeFactory](cm, ctxt)

	test := NewTestScope(ctxt)

	executed := int32(0)
	runTest := func(ctxt Context, scope Scope, t TestScope, scopeCtxt ScopeContext, logger logger.Logger, ano AnotherInterface) {
		logger.Infof("runTest execute once in scope: %s", scope.GetScopeId())

		executed++
	}

	scope := scopeFactory.CreateScopeFrom(test, ScopeTest)
	defer func() { scope.Dispose() }()

	scopeCtxt := scope.GetScopeContext()
	fmt.Printf("scope created, name - %s[%s], id - %s\n", scopeCtxt.Type(), scope.GetTypeName(), scope.GetScopeId())

	testMethod := scope.BuildActionMethod("RunTest", runTest)
	testMethod()

	if executed != 1 {
		t.Errorf("runtest is not executed as expected once: %d", executed)
	}
}

func TestComponentManager_Scoped_Using(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	test := NewTestScope(ctxt)

	executed := int32(0)

	Using[TestScope](test, func(ctxt Context, scope Scope, ts TestScope, scopeCtxt ScopeContext, logger logger.Logger, ano AnotherInterface) {
		logger.Infof("runTest execute once in scope: %s", scope.GetScopeId())

		executed++

		if ts != test {
			t.Errorf("injected scope inst %p is not expected: %p", ts, test)
		}
	})

	if executed != 1 {
		t.Errorf("runtest is not executed as expected once: %d", executed)
	}
}

func TestComponentManager_Scoped_NonTypedScope(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)

	executed := int32(0)
	runTest := func(ctxt Context, scope Scope, scopeCtxt ScopeContext, logger logger.Logger, ano AnotherInterface) {
		logger.Infof("runTest execute once in scope: %s", scope.GetScopeId())

		executed++
	}

	scope := scopeFactory.CreateScope(ctxt)
	defer func() { scope.Dispose() }()

	scopeCtxt := scope.GetScopeContext()
	fmt.Printf("scope created, name - %s[%s], id - %s\n", scopeCtxt.Type(), scope.GetTypeName(), scope.GetScopeId())

	data := scope.GetScopeContext().(ScopeContextEx).GetScope()
	if data.IsTypedScope() {
		t.Error("not epxected scope category, should be nontyped in this case")
	}

	scopeType := scope.GetType()
	if scopeType.Name() != ScopeType_None.Name() {
		t.Errorf("nontyped scope name not expected: %s", scopeType.Name())
	}

	match := data.Match(ScopeType_Any)
	if !match {
		t.Errorf("nontyped scope %s should match ScopeType_Any", scopeType.Name())
	}

	testMethod := scope.BuildActionMethod("RunTest", runTest)
	testMethod()

	if executed != 1 {
		t.Errorf("runtest is not executed as expected once: %d", executed)
	}
}

func TestComponentManager_Scoped_TypedScope(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)

	executed := int32(0)
	runTest := func(ctxt Context, scope Scope, scopeCtxt ScopeContext, logger logger.Logger, ano AnotherInterface) {
		logger.Infof("runTest execute once in scope: %s", scope.GetScopeId())

		executed++
	}

	scope := scopeFactory.CreateTypedScope(ScopeTest)
	defer func() { scope.Dispose() }()

	scopeCtxt := scope.GetScopeContext()
	fmt.Printf("scope created, name - %s[%s], id - %s\n", scopeCtxt.Type(), scope.GetTypeName(), scope.GetScopeId())

	data := scope.GetScopeContext().(ScopeContextEx).GetScope()
	if !data.IsTypedScope() {
		t.Error("not epxected scope category, should be typed scope in this case")
	}

	scopeType := scope.GetType()
	if scopeType.Name() != ScopeTest.Name() {
		t.Errorf("typed scope name not expected: %s, should be: %s", scopeType.Name(), ScopeTest.Name())
	}

	match := data.Match(ScopeType_Any)
	if !match {
		t.Errorf("typed scope %s should match ScopeType_Any", scopeType.Name())
	}
	match = data.Match(ScopeTest)
	if !match {
		t.Errorf("typed scope %s should match ScopeTest", scopeType.Name())
	}

	testMethod := scope.BuildActionMethod("RunTest", runTest)
	testMethod()

	if executed != 1 {
		t.Errorf("runtest is not executed as expected once: %d", executed)
	}
}

func TestComponentManager_Scoped_Typed_Scope_inject(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)

	scope := scopeFactory.CreateTypedScope(ScopeTest)
	defer func() { scope.Dispose() }()

	provider := scope.Provider()
	scopeObj := GetComponent[Scope](provider).(ScopeEx)
	if scopeObj != scope {
		t.Errorf("injected scope %p does not match with source scope %p", scopeObj, scope)
	}
}

type NestedScope interface{}

var ScopeType_Nested = types.Of(new(NestedScope))

type ScopeObject interface {
	Scopable
}
type DefaultScopeObject struct {
	context Context
}

func NewScopeObject(context Context) *DefaultScopeObject {
	return &DefaultScopeObject{context: context}
}
func (so *DefaultScopeObject) Context() Context {
	return so.context
}
func TestComponentManager_Scoped_Typed_Scope_Nested_inject(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterTransient[ScopeObject](cm, NewScopeObject)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)
	{
		scope := CreateTypedScope[TestScope](scopeFactory)
		defer func() { scope.Dispose() }()

		provider := scope.Provider()
		thescope := GetComponent[Scope](provider).(ScopeEx)
		if thescope != scope {
			t.Errorf("injected scope %p does not match with source scope %p", thescope, scope)
		}

		// check scope instance always the same one
		scopeInst := GetComponent[TestScope](provider)
		scopeInst2 := GetComponent[TestScope](provider)
		if scopeInst != scopeInst2 {
			t.Errorf("scope instance should be injected and always the same in the scope: %p != %p", scopeInst, scopeInst2)
		}

		scopeType := scope.GetTypeName()
		if scopeType != ScopeTest.Name() {
			t.Errorf("parent scope type not expected: %s", scopeType)
		}

		scopeObj := GetComponent[ScopeObject](provider)
		scopeCtxt := scopeObj.Context().(ContextEx).GetScopeContext()
		if scopeCtxt.Name() != ScopeTest.Name() {
			t.Errorf("scopeObj's scope type is not expected: %s", scopeCtxt.Name())
		}

		{
			nestedScope := CreateScopeFrom[ScopeObject](scopeFactory, scopeObj)
			defer func() { scope.Dispose() }()

			if nestedScope == scope {
				t.Error("nested scope should never the same as parent scope")
			}

			nestedprovider := nestedScope.Provider()
			thenestedscope := GetComponent[Scope](nestedprovider).(ScopeEx)
			if thenestedscope != nestedScope {
				t.Errorf("nested: injected scope %p does not match with source scope %p", thescope, scope)
			}

			// check nested scope instance always the same one
			scopeObj1 := GetComponent[ScopeObject](nestedprovider)
			scopeObj2 := GetComponent[ScopeObject](nestedprovider)
			if scopeObj1 != scopeObj {
				t.Errorf("scope instance should be injected and always the same in the scope: %p != %p", scopeObj1, scopeObj)
			}
			if scopeObj2 != scopeObj {
				t.Errorf("scope instance should be injected and always the same in the scope: %p != %p", scopeObj2, scopeObj)
			}

			nestedScopeType := nestedScope.GetTypeName()
			if nestedScopeType != types.Get[ScopeObject]().Name() {
				t.Errorf("nested: parent scope type not expected: %s", scopeType)
			}

			PrintAncestorStack(scopeObj.Context(), "scopeObj")
			PrintAncestorStack(nestedScope.Context(), "nestedScope")
		}
	}
}

func TestComponentManager_Scoped_Typed_ScopeInstance(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[TestScope](cm, NewTestScope)
	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)

	scope := CreateTypedScope[TestScope](scopeFactory)
	_, scopeObj := scope.GetScopeContext().(ScopeContextEx).GetScope().GetInstance()
	defer func() { scope.Dispose() }()

	provider := scope.Provider()
	scopeInst := provider.GetComponent(ScopeTest)
	if scopeInst != scopeObj {
		t.Errorf("scope object from scope provider %p does not match with object from scope %p", scopeInst, scopeObj)
	}
}

func TestComponentManager_Scoped_Typed_ScopeInstance_global(t *testing.T) {
	defer test.AssertPanicContent(t, "dependency not configured, type: dep.Global", "panic content not expected")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	cm.RegisterTransientForType(NewTestScope, ScopeTest)

	scopeFactory := GetComponentFrom[ScopeFactory](cm, ctxt, nil)

	scope := scopeFactory.CreateTypedScope(ScopeTest)
	defer func() { scope.Dispose() }()

	// never got the global inst as its nil is treated as not found, thus fail due to not registered in depDict
	provider := scope.Provider()
	globalInst := provider.GetComponent(ScopeType_Global)
	if globalInst != nil {
		t.Errorf("global scope object should be nil, actual - %p", globalInst)
	}
}

func TestComponentManager_Scoped_Typed_Component(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterScoped[AnotherInterface, TestScope](cm, NewAnotherStruct)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("scoped instances are the same one in the same scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are different in the same scope! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Scoped_Typed_not_match(t *testing.T) {
	defer test.AssertPanicContent(t, "must be used in scope of type", "panic content mis-match for Scoped_Typed_not_match")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterScoped[AnotherInterface, TestScope](cm, NewAnotherStruct)

	ctxt1 := createNewContext(ctxt, ScopeType_None)
	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("scoped instances are the same one in the same scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are different in the same scope! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Scoped_lifetime(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true

	cm, ctxt := prepareComponentManagerWithScope(options, ScopeTest)

	// singleton depends on scoped component
	RegisterSingleton[FirstInterface](cm, NewActualStruct)
	RegisterScoped[SecondInterface, TestScope](cm, NewSecondDepOnFirst)

	inst1 := GetComponentFrom[SecondInterface](cm, ctxt, nil)
	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
}

func TestComponentManager_Scoped_lifetime_exceed(t *testing.T) {
	defer test.AssertPanicContent(t, "must be used in scope of type", "panic content mis-match for Scoped_lifetime_exceed")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true

	cm, ctxt := prepareComponentManagerWithScope(options, ScopeTest)

	// singleton depends on scoped component
	RegisterSingleton[FirstInterface](cm, NewFirstDepOnSecond)
	RegisterScoped[SecondInterface, TestScope](cm, NewActualStruct)

	inst1 := GetComponentFrom[FirstInterface](cm, ctxt, nil)
	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
}

func TestComponentManager_Scoped_Nested(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithScope(options, ScopeTest, ScopeType_None)
	num := cm.Count()

	RegisterScoped[AnotherInterface, TestScope](cm, NewAnotherStruct)

	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt, nil)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("scoped instances are the same one in the same scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are different in the same scope! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

func TestComponentManager_Scoped_Nested_not_match(t *testing.T) {
	defer test.AssertPanicContent(t, "must be used in scope of type", "panic content mis-match for Scoped_Nested_not_match")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithScope(options, ScopeType_None, ScopeType_None)
	num := cm.Count()

	RegisterScoped[AnotherInterface, TestScope](cm, NewAnotherStruct)

	inst1 := GetComponentFrom[AnotherInterface](cm, ctxt, nil)
	inst2 := GetComponentFrom[AnotherInterface](cm, ctxt, nil)

	if inst1 == nil {
		t.Logf("inst1 is nil")
	}
	if inst1 == inst2 {
		t.Logf("scoped instances are the same one in the same scope, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("scoped instances are different in the same scope! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 1 {
		t.Errorf("registered component type count is not 1: %v", num)
	}
}

type BigScope interface{}

var ScopeBig types.DataType = types.Get[BigScope]()

type SmallScope interface{}

var ScopeSmall types.DataType = types.Get[SmallScope]()

func TestComponentManager_Scoped_Nested_lifetime(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true

	cm, ctxt := prepareComponentManagerWithScope(options, ScopeBig, ScopeSmall)

	RegisterScoped[FirstInterface, BigScope](cm, NewActualStruct)
	RegisterScoped[SecondInterface, SmallScope](cm, NewSecondDepOnFirst)

	inst := GetComponentFrom[SecondInterface](cm, ctxt, nil)

	fmt.Printf("created instance within nested scope: %p\n", inst)
}

func TestComponentManager_Scoped_Nested_lifetime_exceed(t *testing.T) {
	defer test.AssertPanicContent(t, "must be used in scope of type", "panic content mis-match for Scoped_Nested_lifetime_exceed")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true

	cm, ctxt := prepareComponentManagerWithScope(options, ScopeBig, ScopeSmall)

	// bigger scoped component depends on samller scoped component
	RegisterScoped[FirstInterface, BigScope](cm, NewFirstDepOnSecond)
	RegisterScoped[SecondInterface, SmallScope](cm, NewActualStruct)

	inst := GetComponentFrom[FirstInterface](cm, ctxt, nil)
	if inst == nil {
		t.Logf("inst1 is nil")
	}
}

func TestComponentManager_Singleton_ThreadSafe(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)

	size := 10
	start := make(chan bool)
	complete := make(chan interface{}, size)
	createInstance := func() {
		<-start
		inst := Test_GetComponent[AnotherInterface](cm, ctxt)
		complete <- inst
	}

	for i := 0; i < size; i++ {
		go createInstance()
	}

	time.Sleep(500 * time.Millisecond)
	close(start)

	var inst interface{}
	for i := 0; i < size; i++ {
		new_inst := <-complete
		if new_inst == nil {
			t.Logf("new_inst is nil")
		} else if inst == nil {
			inst = new_inst
		} else {
			if inst != new_inst {
				t.Errorf("singleton instances are different! %p != %p", inst, new_inst)
			}
		}
	}
}

func TestComponentManager_Scoped_ThreadSafe(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[AnotherInterface, any](cm, NewAnotherStruct)

	scopeCnt := 3
	size := 10

	start := make(chan bool)
	complete := make([]chan interface{}, 0, scopeCnt)
	for j := 0; j < scopeCnt; j++ {
		ctxt1 := createNewContext(ctxt, ScopeTest)

		complete = append(complete, make(chan interface{}, size))
		createInstance := func(index int) {
			<-start
			inst := GetComponentFrom[AnotherInterface](cm, ctxt1, nil)
			complete[index] <- inst
		}

		for i := 0; i < size; i++ {
			go createInstance(j)
		}
	}

	time.Sleep(500 * time.Millisecond)
	close(start)

	insts := make([]interface{}, 0, scopeCnt)
	for j := 0; j < scopeCnt; j++ {
		var inst interface{}
		for i := 0; i < size; i++ {
			new_inst := <-complete[j]
			if new_inst == nil {
				t.Logf("new_inst is nil")
			} else if inst == nil {
				inst = new_inst
			} else {
				if inst != new_inst {
					t.Errorf("scoped instances from same scope are different! %p != %p", inst, new_inst)
				}
			}
		}
		// check inst from different scopes are different
		insts = append(insts, inst)
		for i := 0; i < j; i++ {
			if insts[i] == inst {
				t.Errorf("scoped instances from different scopes are same! %p != %p", inst, insts[i])
			}
		}
	}
}

type Contextable interface{
	GetContext() Context
}
type FirstInterface interface {
	First()
}

type SecondInterface interface {
	Second()
}

type ActualStruct struct {
	value int
}

func NewActualStruct() *ActualStruct {
	return &ActualStruct{
		value: 1,
	}
}

func (as *ActualStruct) First() {
	fmt.Println("First", as.value)
}
func (as *ActualStruct) Second() {
	fmt.Println("Second", as.value)
}

type AnotherInterface interface {
	Contextable
	Another()
}

type AnotherStruct struct {
	context Context
	value int
}

func NewAnotherStruct(context Context) *AnotherStruct {
	return &AnotherStruct{
		context: context,
		value: 0,
	}
}
func (as *AnotherStruct) GetContext() Context {
	return as.context
}
func (as *AnotherStruct) Another() {
	fmt.Println("Another", as.value)
}

type rtype struct{}
type emptyInterface struct {
	typ  *rtype
	word unsafe.Pointer
}

func GetTypeAndValue(inst interface{}) (types.DataType, interface{}) {
	instType := types.Of(inst)
	if instType.IsPtr() {
		instVal := *(*emptyInterface)(unsafe.Pointer(&inst))
		valPtr := (*emptyInterface)(instVal.word)
		ptrVal := (*emptyInterface)(valPtr.word)
		return instType.ElementType(), ptrVal
	} else {
		return instType, inst
	}
}
func TestValuePointer(t *testing.T) {
	actual := NewActualStruct()
	fmt.Printf("%s: %p\n", types.Of(actual).Name(), actual)
	var first FirstInterface = actual
	fmt.Printf("%s: %p\n", types.Of(first).Name(), first)

	var pf interface{} = &first
	fmt.Printf("%s: %p %p\n", types.Of(pf).Name(), first, pf)
	T, V := GetTypeAndValue(pf)
	fmt.Printf("%s: %p\n", T.Name(), V)

	ty := types.Of(pf)

	eface := *(*emptyInterface)(unsafe.Pointer(&pf))
	ptr := (*emptyInterface)(eface.word)
	val := (*emptyInterface)(ptr.word)
	fmt.Printf("%s: %p\n", ty.Name(), val)
}
func TestComponentManager_AddSingleton_MultiTypes(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)
	num := cm.Count()

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	cm.RegisterSingletonForTypes(NewActualStruct, types.Get[FirstInterface](), types.Get[SecondInterface]())

	inst0 := Test_GetComponent[AnotherInterface](cm, ctxt)
	if inst0 == nil {
		t.Logf("inst0 is nil")
	}

	inst1 := Test_GetComponent[FirstInterface](cm, ctxt)
	inst2 := Test_GetComponent[SecondInterface](cm, ctxt)

	t.Logf("service instances, inst0 - %p, inst1 - %p, inst2 - %p", inst0, inst1, inst2)

	inst0.Another()
	inst1.First()
	inst2.Second()

	if inst0.(interface{}) == inst1.(interface{}) {
		t.Errorf("service instances should not be the same! inst0 - %p, inst1 - %p", inst0, inst1)
	} else {
		t.Logf("service instances are different %p != %p", inst0, inst1)
	}

	if inst1 == nil {
		t.Errorf("inst1 is nil")
	}
	if inst1.(interface{}) == inst2.(interface{}) {
		t.Logf("service instances are the same one, inst1 - %p, inst2 - %p", inst1, inst2)
	} else {
		t.Errorf("service instances are different! %p != %p", inst1, inst2)
	}

	num = cm.Count() - num
	if num != 3 {
		t.Errorf("registered component type count is not 3: %v", num)
	}
}

func TestComponentManager_AddSingleton_InCompatibleType(t *testing.T) {
	captured := false
	var failure string
	done := make(chan bool, 1)

	go func(){
		defer func() {
			if r := recover(); r != nil {
				captured = true
				failure = fmt.Sprintf("%v", r)
				fmt.Printf("captured panic: %s\n", failure)
			}
			done <- true
		}()

		options := NewComponentProviderOptions(InterfaceType, StructType)
		cm, ctxt := prepareComponentManagerWithOptions(options)
		// create a mismatch type register
		RegisterSingleton[AnotherInterface](cm, NewActualStruct)

		inst0 := Test_GetComponent[AnotherInterface](cm, ctxt)
		inst0.Another()
	}()

	<-done

	if !captured {
		t.Errorf("instance type mismatch not captured")
	} else {
		if !strings.Contains(failure, "is not compatible with interface type(dep.AnotherInterface)") {
			t.Errorf("panic content not expected: %s", failure)
		}
	}
}

func TestComponentManager_AddTransient(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructPtrType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	op := &ActualStruct{}
	RegisterTransient[*ActualStruct](cm, func(ctxt Context) interface{} { return op })
	RegisterTransient[FirstInterface](cm, func(ctxt Context) interface{} { return op })

	structResult := Test_GetComponent[*ActualStruct](cm, ctxt)

	if op != structResult {
		t.Errorf("Get Interface instance failed!")
	}
}

func TestComponentManager_Transient_ThreadSafe(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.TrackTransientRecurrence = false
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[AnotherInterface](cm, func(ctxt Context) interface{} {
		time.Sleep(20 * time.Millisecond)
		return NewAnotherStruct(ctxt)
	})

	size := 10
	start := make(chan bool)
	complete := make(chan interface{}, 10)
	createInstance := func() {
		<-start
		inst := Test_GetComponent[AnotherInterface](cm, ctxt)
		complete <- inst
	}

	for i := 0; i < size; i++ {
		go createInstance()
	}

	time.Sleep(500 * time.Millisecond)
	close(start)

	insts := make([]interface{}, 0, size)
	for i := 0; i < size; i++ {
		new_inst := <-complete
		if new_inst == nil {
			t.Logf("new_inst is nil")
		} else {
			for j := 0; j < i; j++ {
				if insts[j] == new_inst {
					t.Errorf("transient instances are the same! [%d]:%p != [%d]:%p", j, insts[j], i, new_inst)
				}
			}
			insts = append(insts, new_inst)
		}
	}
}

type Component interface {
	DoWork()
}
type DefaultComponent struct {
	context Context
	config  *MyConfig
	first   FirstInterface
	second  SecondInterface
	another AnotherInterface
}

func (c *DefaultComponent) DoWork() {
	fmt.Println("DefaultComponent DoWork start")
	c.first.First()
	c.second.Second()
	c.another.Another()
	fmt.Println("DefaultComponent DoWork done")
}

func NewComponent(first FirstInterface, second SecondInterface, another AnotherInterface) *DefaultComponent {
	return &DefaultComponent{
		first:   first,
		second:  second,
		another: another,
	}
}

func NewComponentWithContext(context Context, first FirstInterface, second SecondInterface, another AnotherInterface) *DefaultComponent {
	comp := &DefaultComponent{
		context: context,
		config:  GetConfig[MyConfig](context),
		first:   first,
		second:  second,
		another: another,
	}

	defaultLogger := context.GetLogger()
	defaultLogger.Info("DEFAULT: component created with context injected")

	logger := context.GetLoggerWithName(GetDefaultLoggerNameForComponent(comp))
	logger.Info("Component logger created from component context")

	return comp
}

func NewComponentWithConfig(config *MyConfig, first FirstInterface, second SecondInterface, another AnotherInterface) *DefaultComponent {
	return &DefaultComponent{
		config:  config,
		first:   first,
		second:  second,
		another: another,
	}
}

func TestComponentManager_DependencyInjection_Singleton(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	cm.RegisterSingletonForTypes(NewActualStruct, types.Get[FirstInterface](), types.Get[SecondInterface]())
	RegisterSingleton[Component](cm, NewComponent)

	comp := Test_GetComponent[Component](cm, ctxt)
	comp.DoWork()

	dc := comp.(*DefaultComponent)
	if dc.first.(interface{}) != dc.second.(interface{}) {
		t.Errorf("first instance is not the same as second instance!")
	}
}

func NewFirstDepOnSecond(second SecondInterface) FirstInterface {
	return NewActualStruct()
}
func NewSecondDepOnFirst(first FirstInterface) SecondInterface {
	return NewActualStruct()
}
func TestComponentManager_DependencyInjection_CyclicDependency(t *testing.T) {
	defer test.AssertPanicContent(t, "cyclic dependency detected on singleton component", fmt.Sprintf("panic error should happened for cyclic dependency on type %s", "FirstInterface"))

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = false
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[FirstInterface](cm, NewFirstDepOnSecond)
	RegisterSingleton[SecondInterface](cm, NewSecondDepOnFirst)

	first := Test_GetComponent[FirstInterface](cm, ctxt)
	first.First()

	t.Errorf("cyclic dependency should be detected on creating component of type FirstInterface")
}

func TestComponentManager_DependencyInjection_CyclicDependency_concurrent(t *testing.T) {
	defer test.AssertPanic(t, fmt.Sprintf("panic error should happened for cyclic dependency on type %s", "FirstInterface"))

	options := NewComponentProviderOptions(InterfacePtrType, StructPtrType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	cm.RegisterSingletonForType(NewFirstDepOnSecond, types.Of(new(FirstInterface)))
	cm.RegisterSingletonForType(NewSecondDepOnFirst, types.Of(new(SecondInterface)))

	first := cm.GetComponent(types.Of(new(FirstInterface)), ctxt).(FirstInterface)
	first.First()

	t.Errorf("cyclic dependency should be detected on creating component of type FirstInterface")
}

func TestComponentManager_DependencyInjection_CyclicDependency_Scoped(t *testing.T) {
	defer test.AssertPanicContent(t, "cyclic dependency detected on scoped", fmt.Sprintf("panic error should happened for cyclic dependency on scoped component type %s", "FirstInterface"))

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = false
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[FirstInterface, any](cm, NewFirstDepOnSecond)
	RegisterScoped[SecondInterface, any](cm, NewSecondDepOnFirst)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	first := GetComponentFrom[FirstInterface](cm, ctxt1, nil)
	first.First()

	t.Errorf("cyclic dependency should be detected on creating scoped component of type FirstInterface")
}

func TestComponentManager_DependencyInjection_CyclicDependency_Scoped_concurrent(t *testing.T) {
	defer test.AssertPanicContent(t, "cyclic dependency detected on scoped", fmt.Sprintf("panic error should happened for cyclic dependency on scoped component type %s", "FirstInterface"))

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.EnableSingletonConcurrency = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterScoped[FirstInterface, any](cm, NewFirstDepOnSecond)
	RegisterScoped[SecondInterface, any](cm, NewSecondDepOnFirst)

	ctxt1 := createNewContext(ctxt, ScopeTest)
	first := GetComponentFrom[FirstInterface](cm, ctxt1, nil)
	first.First()

	t.Errorf("cyclic dependency should be detected on creating scoped component of type FirstInterface")
}

func TestComponentManager_DependencyInjection_CyclicDependency_Transient(t *testing.T) {
	defer test.AssertPanicContent(t, "recursive dependency overflow on transient component", fmt.Sprintf("panic error should happened for cyclic dependency on transient component type %s", "FirstInterface"))

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.EnableDiagnostics = true
	options.TrackTransientRecurrence = true
	options.MaxAllowedRecurrence = 4
	cm, ctxt := prepareComponentManagerWithOptions(options)


	RegisterTransient[FirstInterface](cm, NewFirstDepOnSecond)
	RegisterTransient[SecondInterface](cm, NewSecondDepOnFirst)

	first := Test_GetComponent[FirstInterface](cm, ctxt)
	first.First()

	t.Errorf("cyclic dependency for transient component should be detected on creating component of type FirstInterface")
}

// type MyService interface {
// 	Start()
// 	Shutdown(ctx context.Context) error

// 	Run()
// }
// type DefaultMyService struct {
// 	logger logger.Logger
// 	value  int
// }

// func NewMyService(context Context) *DefaultMyService {
// 	s := &DefaultMyService{
// 		logger: context.GetLogger(),
// 		value:  123,
// 	}
// 	s.logger.Infow("MyService created", "type", types.Of(s).Name())
// 	s.logger.Info("service context info", "type", context.Type(), "name", context.Name())
// 	return s
// }

// func (s *DefaultMyService) Start() {
// 	go s.Run()
// }
// func (s *DefaultMyService) Run() {
// 	s.runService()
// }
// func (s *DefaultMyService) Shutdown(ctx context.Context) error {
// 	return nil
// }

// func (s *DefaultMyService) runService() {
// 	s.logger.Infow("MyService run start", "type", types.Of(s).Name())

// 	s.logger.Infow("MyService run complete", "value", s.value)
// }

type FakeServiceContext struct {
	scopeCtxt  ScopeContextEx
	cm         ComponentManager
	debug      bool
	depTracker DependencyTracker

	// local context dependencies
	localDeps DepDict[ComponentGetter]
}

func NewFakeServiceContext(cm ComponentManager, scopeCtxt ScopeContextEx) *FakeServiceContext {
	sc := &FakeServiceContext{
		debug:      true,
		scopeCtxt:  scopeCtxt,
		cm:         cm,
		depTracker: NewDependencyTracker(nil),
	}
	sc.localDeps = NewDependencyDictionary[ComponentGetter]()
	return sc
}

func (fsc *FakeServiceContext) Type() string                { return "Service" }
func (fsc *FakeServiceContext) Name() string                { return "Fake_Name" }
func (fsc *FakeServiceContext) GetHostContext() HostContext { return nil }
func (fsc *FakeServiceContext) GetLogger() logger.Logger    { return logger.GetLogger("") }
func (fsc *FakeServiceContext) GetLoggerWithName(name string) logger.Logger {
	return logger.GetLogger(name)
}
func (fsc *FakeServiceContext) GetLoggerFactory() logger.LoggerFactory {
	return logger.NewDefaultLoggerFactory()
}
func (fsc *FakeServiceContext) AddDependency(depType types.DataType, depFact ComponentGetter) {
	fsc.localDeps.AddDependency(depFact, depType)
}
func (fsc *FakeServiceContext) GetConfiguration(configType types.DataType) interface{} {
	return fsc.cm.GetConfiguration(configType, fsc)
}
func (fsc *FakeServiceContext) GetComponent(interfaceType types.DataType) interface{} {
	return fsc.CreateWithProperties(interfaceType, nil)
}
func (fsc *FakeServiceContext) getContextDependency(depType types.DataType) interface{} {
	// check local dict first
	if fsc.localDeps.ExistDependency(depType) {
		return fsc.localDeps.GetDependency(depType)()
	}
	// check ancester scopes recursively
	return fsc.scopeCtxt.GetDependency(depType)
}
func (fsc *FakeServiceContext) CreateWithProperties(interfaceType types.DataType, props Properties) interface{} {
	// match dependency of current context including ancestor scopes
	inst := fsc.getContextDependency(interfaceType)
	if inst != nil {
		return inst
	}
	return fsc.cm.GetOrCreateWithProperties(interfaceType, fsc, props)
}
func (fsc *FakeServiceContext) GetProperties() Properties {
	return nil
}
func (fsc *FakeServiceContext) UpdateProperties(props Properties) {
}
func (fsc *FakeServiceContext) GetTracker() DependencyTracker {
	return fsc.depTracker
}
func (fsc *FakeServiceContext) IsDebug() bool {
	return fsc.debug
}
func (fsc *FakeServiceContext) GetScopeContext() ScopeContextEx {
	return fsc.scopeCtxt
}
func (fsc *FakeServiceContext) GetScope() ScopeData {
	return nil
}

// func NewFirstDepOnMyService(service MyService) FirstInterface {
// 	return NewActualStruct()
// }

// func NewMyServiceDepOnContext(context Context) MyService {
// 	return NewMyService(context)
// }

// func TestComponentManager_DependencyInjection_Service(t *testing.T) {
// 	options := NewComponentProviderOptions(InterfaceType, StructType)
// 	options.AllowTypeAnyFromFactoryMethod = true
// 	options.EnableDiagnostics = true
// 	cm, ctxt := prepareComponentManagerWithOptions(options)

// 	serviceCtxt := NewFakeServiceContext(cm, ctxt.GetScopeContext())
// 	RegisterSingleton[FirstInterface](cm, NewFirstDepOnMyService)
// 	RegisterService[MyService](cm, NewMyServiceDepOnContext,
// 		func(scopeCtxt ScopeContextEx) ServiceContextEx {
// 			AddDependency[Context](serviceCtxt, Getter[Context](serviceCtxt))
// 			AddDependency[ServiceContext](serviceCtxt, Getter[ServiceContext](serviceCtxt))
// 			AddDependency[logger.Logger](serviceCtxt, func() logger.Logger { return serviceCtxt.GetLogger() })
// 			AddDependency[Properties](serviceCtxt, func() Properties { return serviceCtxt.GetProperties() })
// 			AddDependency[ScopeContext](serviceCtxt, func() ScopeContext { return serviceCtxt.GetScopeContext() })
// 			return serviceCtxt
// 		},
// 	)

// 	first := Test_GetComponent[FirstInterface](cm, ctxt)
// 	first.First()
// 	service := Test_GetComponent[MyService](cm, ctxt)
// 	service.Run()
// }

// func NewMyServiceDepOnFirst(context Context, first FirstInterface) MyService {
// 	return NewMyService(context)
// }
// func TestComponentManager_DependencyInjection_ServiceCycleDeps(t *testing.T) {
// 	defer test.AssertPanicContent(t, "cyclic dependency detected on singleton component dep.MyService", fmt.Sprintf("panic error should happened for cyclic dependency on type %s", "MyService"))

// 	options := NewComponentProviderOptions(InterfaceType, StructType)
// 	options.AllowTypeAnyFromFactoryMethod = true
// 	options.EnableDiagnostics = true
// 	options.EnableSingletonConcurrency = true
// 	cm, ctxt := prepareComponentManagerWithOptions(options)

// 	serviceCtxt := NewFakeServiceContext(cm, ctxt.GetScopeContext())
// 	AddDependency[Context](serviceCtxt, Getter[Context](serviceCtxt))
// 	AddDependency[ServiceContext](serviceCtxt, Getter[ServiceContext](serviceCtxt))

// 	RegisterSingleton[FirstInterface](cm, NewFirstDepOnMyService)
// 	RegisterService[MyService](cm, NewMyServiceDepOnFirst, func(scopeCtxt ScopeContextEx) ServiceContextEx { return serviceCtxt })

// 	service := Test_GetComponent[MyService](cm, ctxt)
// 	service.Run()
// }

func TestComponentManager_DependencyInjection_Transient(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	RegisterTransient[FirstInterface](cm, NewActualStruct)
	RegisterTransient[SecondInterface](cm, NewActualStruct)
	RegisterSingleton[Component](cm, NewComponent)

	comp := Test_GetComponent[Component](cm, ctxt)
	comp.DoWork()

	dc := comp.(*DefaultComponent)
	if dc.first.(interface{}) == dc.second.(interface{}) {
		t.Errorf("first instance should not be the same as second instance!")
	}
}

func TestComponentManager_DependencyInjection_Context(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	// Inject LoggerFactory
	//cm.RegisterSingletonForType(func(ctxt Context) logger.LoggerFactory { return logger.NewDefaultLoggerFactory() }, types.Of(new(logger.LoggerFactory)))

	config := &MyConfig{
		value: 123,
	}
	cm.AddConfiguration(config)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	cm.RegisterSingletonForTypes(func(ctxt Context) interface{} { return NewActualStruct() }, types.Get[FirstInterface](), types.Get[SecondInterface]())
	RegisterSingleton[Component](cm, NewComponentWithContext)

	comp := Test_GetComponent[Component](cm, ctxt)
	comp.DoWork()

	dc := comp.(*DefaultComponent)
	if dc.first.(interface{}) != dc.second.(interface{}) {
		t.Errorf("first instance is not the same as second instance!")
	}

	if dc.context == nil {
		t.Errorf("context is not injected")
	} else {
		if dc.context.Type() != "Component" || dc.context.Name() != types.Get[Component]().FullName() {
			t.Errorf("context type and name not expected: %v, %v", dc.context.Type(), dc.context.Name())
		}
	}
}

func TestComponentManager_DependencyInjection_Transient_Contextual(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterSingleton[AnotherInterface](cm, func(ctxt Context, log logger.Logger) interface{} {
		log.Info("singleton component with logger injected")
		return NewAnotherStruct(ctxt)
	})
	RegisterTransient[FirstInterface](cm, func(ctxt Context, log logger.Logger) interface{} {
		log.Info("transient component with logger injected")
		return NewActualStruct()
	})

	ano := Test_GetComponent[AnotherInterface](cm, ctxt)
	ano.Another()
	first := Test_GetComponent[FirstInterface](cm, ctxt)
	first.First()
}

func TestComponentManager_DependencyInjection_Config(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	config := &MyConfig{
		value: 123,
	}
	cm.AddConfiguration(config)

	RegisterSingleton[AnotherInterface](cm, NewAnotherStruct)
	cm.RegisterSingletonForTypes(NewActualStruct, types.Get[FirstInterface](), types.Get[SecondInterface]())
	RegisterSingleton[Component](cm, NewComponentWithConfig)

	comp := Test_GetComponent[Component](cm, ctxt)
	comp.DoWork()

	dc := comp.(*DefaultComponent)
	if dc.first.(interface{}) != dc.second.(interface{}) {
		t.Errorf("first instance is not the same as second instance!")
	}
}

// type FakeScopeContext interface {
// }
// type DefaultFakeScopeContext struct {
// }

// type FakeFuncProc interface {
// 	Run()
// }

// type ProcessorMethod func(context Context, interfaceType types.DataType, scopeCtxt FakeScopeContext)

// type TestFuncProc FakeFuncProc

// type DefaultFakeFuncProc struct {
// 	context  Context
// 	procFunc ProcessorMethod
// }

// func NewFakeFuncProc(context Context, processorFunc ProcessorMethod) *DefaultFakeFuncProc {
// 	return &DefaultFakeFuncProc{
// 		context:  context,
// 		procFunc: processorFunc,
// 	}
// }

// func (ffp *DefaultFakeFuncProc) Run() {
// 	ffp.procFunc(ffp.context, types.Of(new(TestFuncProc)), &DefaultFakeScopeContext{})
// }

// func TestComponentManager_DependencyInjection_Processor(t *testing.T) {
// 	options := NewComponentProviderOptions(InterfaceType, StructType)
// 	options.AllowTypeAnyFromFactoryMethod = true
// 	cm, ctxt := prepareComponentManagerWithOptions(options)

// 	val := int(1)
// 	procFunc := func(ctxt Context, scopeCtxt FakeScopeContext, log logger.Logger) {
// 		val = val + 1
// 		log.Infof("proc func set val to %d", val)
// 	}
// 	createProcessor := func(context Context, compType types.DataType, actionFunc GenericActionMethod) interface{} {
// 		processor := NewFakeFuncProc(context, func(context Context, interfaceType types.DataType, scopeCtxt FakeScopeContext) {
// 			actionFunc(context, "TestProcFunc", DepInst[FakeScopeContext](scopeCtxt))
// 		})
// 		return processor
// 	}

// 	RegisterProcessor[TestFuncProc](cm, procFunc, createProcessor)

// 	proc := Test_GetComponent[TestFuncProc](cm, ctxt)
// 	proc.Run()

// 	fmt.Printf("val: %d\n", val)
// 	if val != 2 {
// 		t.Errorf("proc func is not executed, expected - %d, actual - %d", 2, val)
// 	}
// }

// func TestComponentManager_DependencyInjection_FuncProcessor(t *testing.T) {
// 	options := NewComponentProviderOptions(InterfaceType, StructType)
// 	options.AllowTypeAnyFromFactoryMethod = true
// 	cm, ctxt := prepareComponentManagerWithOptions(options)

// 	val := int(1)
// 	procFunc := func(ctxt Context, scopeCtxt FakeScopeContext, log logger.Logger) {
// 		val = val + 1
// 		log.Infof("proc func set val to %d", val)
// 	}
// 	createProcessor := func(context Context, compType types.DataType, actionFunc GenericActionMethod) interface{} {
// 		processor := NewFakeFuncProc(context, func(context Context, interfaceType types.DataType, scopeCtxt FakeScopeContext) {
// 			actionFunc(context, "TestProcFunc", DepInst[FakeScopeContext](scopeCtxt))
// 		})
// 		return processor
// 	}

// 	// TODO: RegisterFuncProcessor instead
// 	RegisterProcessor[TestFuncProc](cm, procFunc, createProcessor)

// 	proc := Test_GetComponent[TestFuncProc](cm, ctxt)
// 	proc.Run()

// 	fmt.Printf("val: %d\n", val)
// 	if val != 2 {
// 		t.Errorf("proc func is not executed, expected - %d, actual - %d", 2, val)
// 	}
// }

type Dependent interface {
	GetValue() int
}
type DefaultDependent struct {
	props Properties
}

func NewDependent() *DefaultDependent {
	return &DefaultDependent{}
}
func NewDependentWithProps(props Properties) *DefaultDependent {
	return &DefaultDependent{props: props}
}
func (dd *DefaultDependent) GetValue() int {
	return dd.props.Get("value").(int)
}

type Value interface {
	Value() int
	GetDependent() Dependent
}

type PropValue struct {
	value int
	dep   Dependent
}

func NewPropValue(props Properties, dep Dependent) *PropValue {
	return &PropValue{
		value: props.Get("value").(int),
		dep:   dep,
	}
}
func (pv *PropValue) Value() int {
	return pv.value
}
func (pv *PropValue) GetDependent() Dependent {
	return pv.dep
}

func NewAnotherStructWithProps(props Properties) *AnotherStruct {
	return &AnotherStruct{
		value: 0,
	}
}

func TestComponentManager_DependencyInjection_Properties(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[Dependent](cm, NewDependent)
	RegisterTransient[Value](cm, NewPropValue)

	props := Props(Pair("value", 123))
	val := CreateComponent[Value](ctxt, NewPropertiesFrom(props))
	result := val.Value()
	if result != 123 {
		t.Errorf("expected - %d, actual - %d", 123, result)
	}

	dep := val.GetDependent()
	if dep == nil {
		t.Error("dependent is not injected")
	}
}

func TestComponentManager_DependencyInjection_Properties_inherit(t *testing.T) {
	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true
	options.PropertiesPassOver = true
	cm, ctxt := prepareComponentManagerWithOptions(options)

	RegisterTransient[Dependent](cm, NewDependentWithProps)
	RegisterTransient[Value](cm, NewPropValue)

	props := Props(Pair("value", 123))
	val := GetComponentFrom[Value](cm, ctxt, NewPropertiesFrom(props))
	result := val.Value()
	if result != 123 {
		t.Errorf("expected - %d, actual - %d", 123, result)
	}

	dep := val.GetDependent()
	if dep == nil {
		t.Error("dependent is not injected")
	}
	inheritVal := dep.GetValue()
	if inheritVal != 123 {
		t.Errorf("property value is not inherited: %d", inheritVal)
	}
}

func TestProperties_String(t *testing.T) {
	props := Props(Pair("age", 123), Pair("name", "jack"))
	strVal := props.String()
	fmtStr := fmt.Sprintf("%v", props)
	if fmtStr != strVal {
		t.Errorf("string value of props are not the same: %s != %s", strVal, fmtStr)
	}

	if strVal != "{age=123,name=jack}" {
		t.Errorf("string value of props is not expected: %s", strVal)
	}
}

func TestSingletonTypeName(t *testing.T) {
	compTypes := []types.DataType{}
	strVal := typesToString(compTypes)
	if strVal != "<None>" {
		t.Errorf("string value of empty compTypes is not expected: %s", strVal)
	}

	firstType := types.Of(new(FirstInterface))
	secondType := types.Of(new(SecondInterface))
	anotherType := types.Of(new(AnotherInterface))

	compTypes = []types.DataType{firstType}
	strVal = typesToString(compTypes)
	if strVal != firstType.FullName() {
		t.Errorf("string value of single compTypes is not expected: %s, actual - %s", firstType.FullName(), strVal)
	}

	compTypes = []types.DataType{firstType, secondType, anotherType}
	strVal = typesToString(compTypes)
	expected := "<AnotherInterface|FirstInterface|SecondInterface>"
	if strVal != expected {
		t.Errorf("string value of single compTypes is not expected: %s, actual - %s", expected, strVal)
	}
}

func Test_threadId(t *testing.T) {
	//stackAsBytes := debug.Stack()
	//stackAsString := string(stackAsBytes)
	//fmt.Printf("stack: %s\n", stackAsString)

	gid := goroutine_id()
	if gid <= 0 {
		t.Errorf("unexpected goroutine id: %d", gid)
	} else {
		t.Logf("goroutine id: %d\n", gid)
	}
}
