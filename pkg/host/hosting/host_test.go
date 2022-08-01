package hosting

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

func createHostBuilder() HostBuilder {
	return NewDefaultHostBuilder()
}
func Test_Host_basic(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	config := host.GetConfiguration().Get()
	if config != nil {
		t.Errorf("config should be nil when it is not configured")
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	factory := dep.GetComponent[logger.LoggerFactory](provider)
	defaultLogger := factory.GetDefaultLogger()
	defaultLogger.Info("log one line from default logger")

	go func() {
		runner.SendStopSignal()
	}()

	host.Run()
}

type TestConfig struct {
}

func Test_Host_context(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.AddConfiguration(&TestConfig{})
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	})

	host := builder.Build()
	hostCtxt := host.GetContext().(dep.HostContextEx)

	scopeCtxt := hostCtxt.GetScopeContext()
	if !scopeCtxt.IsGlobal() {
		t.Error("host context must be in global context")
	}
	global := hostCtxt.GetGlobalScope()
	if global != scopeCtxt {
		t.Errorf("host context scope %p is not equal to global context %p", scopeCtxt, global)
	}

	parent := scopeCtxt.GetParent()
	if parent != nil {
		t.Errorf("scope of host context must has no parent scope: %p", parent)
	}
	scopeId := scopeCtxt.ScopeId()
	if scopeId != dep.ScopeType_Global.Name() {
		t.Errorf("host scope id not expected: %s", scopeId)
	}
	if !scopeCtxt.IsDebug() {
		t.Error("default mode of host context is debug for unit tests")
	}
	ty := scopeCtxt.Type()
	name := scopeCtxt.Name()
	if ty != dep.ContextType_Scope || name != dep.ScopeType_Global.Name() {
		t.Errorf("host scope type and name not expected: %s, %s", ty, name)
	}

	scope := scopeCtxt.GetScope()
	id := scope.GetScopeId()
	if id != scopeId {
		t.Errorf("host scope id not equal to id of scope data: %s", id)
	}
	tyName := scope.GetTypeName()
	if tyName != dep.ScopeType_Global.Name() {
		t.Errorf("host scope type not expected: %s", tyName)
	}

	rawCtxt := host.GetRawContext()
	if rawCtxt == nil {
		t.Error("host raw context should not be nil")
	}

	tracker := hostCtxt.GetTracker()
	parentCtxt := tracker.GetParent()
	if parentCtxt != nil {
		t.Errorf("host context should not have parent context: name -%s, type - %s, scopeId - %s", parentCtxt.Type(), parentCtxt.Name(), parentCtxt.ScopeId())
	}

	props := dep.Props(dep.Pair("key1", 1), dep.Pair("key2", 2))
	hostCtxt.UpdateProperties(props)
	props = dep.Props(dep.Pair("key2", 3), dep.Pair("key3", 4))
	hostCtxt.UpdateProperties(props)
	result := hostCtxt.GetProperties()
	if result != nil {
		t.Errorf("host properties is not supported, must be nil: %p", result)
	}
	// val := result.Get("key2").(int)
	// if val != 3 {
	// 	t.Errorf("property value is not expected: %d", val)
	// }

	config := dep.GetConfig[TestConfig](hostCtxt)
	if config == nil {
		t.Error("get config returns nil")
	}
	ano := dep.CreateComponent[AnotherInterface](hostCtxt, nil)
	ano.Another()
}

func Test_Host_lifecycle(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)

	orderMap := make(map[string]int)
	builder.ConfigureLifecycle(func(hostContext dep.Context, appLifecycle ApplicationLifecycle) {
		var index int = 0
		appLifecycle.RegisterOnHostReady(func(ctx dep.Context) {
			orderMap["OnHostReady"] = index
			index += 1
		})
		appLifecycle.RegisterOnAppStarted(func(ctx dep.Context) {
			orderMap["OnAppStarted"] = index
			index += 1
		})
		appLifecycle.RegisterOnStopEvent(func(dep.Context, *StopEvent) bool {
			orderMap["OnStopEvent"] = index
			index += 1
			return true
		})
		appLifecycle.RegisterOnAppStopping(func(ctx dep.Context) {
			orderMap["OnAppStopping"] = index
			index += 1
		})
		appLifecycle.RegisterOnAppStopped(func(ctx dep.Context) {
			orderMap["OnAppStopped"] = index
			index += 1
		})
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)

	go func() {
		runner.SendStopSignal()
	}()

	host.Run()

	if orderMap["OnHostReady"] != 0 {
		t.Errorf("bad order for lifecycle stage: %v", "OnHostReady")
	}
	if orderMap["OnAppStarted"] != 1 {
		t.Errorf("bad order for lifecycle stage: %v", "OnAppStarted")
	}
	if orderMap["OnStopEvent"] != 2 {
		t.Errorf("bad order for lifecycle stage: %v", "OnStopEvent")
	}
	if orderMap["OnAppStopping"] != 3 {
		t.Errorf("bad order for lifecycle stage: %v", "OnAppStopping")
	}
	if orderMap["OnAppStopped"] != 4 {
		t.Errorf("bad order for lifecycle stage: %v", "OnAppStopped")
	}
}

func Test_Host_components(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewActualStruct, types.Get[FirstInterface](), types.Get[SecondInterface]())
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
		dep.RegisterSingleton[Component](components, NewComponent)
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	config := host.GetConfiguration().Get()
	if config != nil {
		t.Errorf("config should be nil when it is not configured")
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider()
	factory := dep.GetComponent[logger.LoggerFactory](provider)
	defaultLogger := factory.GetDefaultLogger()
	defaultLogger.Info("log one line from default logger")

	comp := dep.GetComponent[Component](provider)
	comp.DoWork()
}

type TestFuncProc FunctionProcessor

func Test_Host_processor(t *testing.T) {
	hostName := "Test"

	result := int(0)
	procFunc := func(ctxt dep.Context, scope ScopeContext) {
		result = 123
	}

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.UseComponentProvider(func(context BuilderContext, options *dep.ComponentProviderOptions) {
		options.AllowTypeAnyFromFactoryMethod = true
	})
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		RegisterFuncProcessor[TestFuncProc](components, procFunc)
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider()

	proc := dep.GetComponent[TestFuncProc](provider)
	proc.Run(NewFakeScopeContext())

	if result != 123 {
		t.Errorf("expected - %d, actual - %d", 123, result)
	}
}

func Test_Host_multi_implementations(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterComponent[Downloader](
			components,
			func(props dep.Properties) interface{} { return props.Get("type") },
			func(comp dep.CompImplCollection) {
				comp.AddImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider().(dep.ComponentProviderEx)

	for _, Type := range []string{"url", "blob"} {
		downloader := dep.CreateComponent[Downloader](provider, dep.Props(dep.Pair("type", Type)))
		downloader.Download()

		if downloader.GetType() != Type {
			t.Errorf("expected - %s, actual - %s", Type, downloader.GetType())
		}
	}
}

func Test_Host_multi_implementations_evaluator_negative(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterComponent[Downloader](
			components,
			func(props dep.Properties) interface{} { return nil },
			func(comp dep.CompImplCollection) {
				comp.AddImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider().(dep.ComponentProviderEx)

	for _, Type := range []string{"url", "blob"} {
		defer test.AssertPanicContent(t, "evaluated component implementation key should never be nil", "panic content is not expected")

		downloader := dep.CreateComponent[Downloader](provider, dep.Props(dep.Pair("type", Type)))
		downloader.Download()
	}
}

func Test_Host_multi_impl_singleton(t *testing.T) {
	hostName := "Test"

	builder := createHostBuilder()
	builder.SetHostName(hostName)
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterComponent[Downloader](
			components,
			func(props dep.Properties) interface{} { return props.Get("type") },
			func(comp dep.CompImplCollection) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	})

	host := builder.Build()
	if host.GetName() != hostName {
		t.Errorf("host name is not expected - %v", host.GetName())
	}

	log := host.GetLogger()
	log.Info("log one line using default logging")

	provider := host.GetComponentProvider().(dep.ComponentProviderEx)

	for _, Type := range []string{"url", "blob"} {
		downloader := dep.CreateComponent[Downloader](provider, dep.Props(dep.Pair("type", Type)))
		downloader.Download()

		if downloader.GetType() != Type {
			t.Errorf("expected - %s, actual - %s", Type, downloader.GetType())
		}

		if Type == "url" {
			// check singleton
			singleton := dep.CreateComponent[Downloader](provider, dep.Props(dep.Pair("type", "url")))
			if singleton != downloader {
				t.Error("singleton implementation should not have multiple instances!")
			}
		}
	}
}

type Downloader interface {
	GetType() string
	Download()
}

type UrlDownloader interface {
	Downloader

	Url() string
}
type DefaultUrlDownloader struct {
	props dep.Properties
}

func NewUrlDownloader() UrlDownloader {
	return &DefaultUrlDownloader{
		props: nil,
	}
}
func (ud *DefaultUrlDownloader) GetType() string {
	return "url"
}
func (ud *DefaultUrlDownloader) Download() {
	fmt.Printf("download for type: %v\n", ud.GetType())
}
func (ud *DefaultUrlDownloader) Url() string {
	return "url"
}

type BlobDownloader interface {
	Downloader

	Blob() string
}
type DefaultBlobDownloader struct {
	props dep.Properties
}

func NewBlobDownloader(props dep.Properties) BlobDownloader {
	return &DefaultBlobDownloader{
		props: props,
	}
}
func (bd *DefaultBlobDownloader) GetType() string {
	return bd.props.Get("type").(string)
}
func (bd *DefaultBlobDownloader) Download() {
	fmt.Printf("download for type: %v\n", bd.GetType())
}
func (bd *DefaultBlobDownloader) Blob() string {
	return "blob"
}

type FakeLoopRunContext struct {
}

func NewFakeLoopRunContext() *FakeLoopRunContext {
	return &FakeLoopRunContext{}
}

func (rc *FakeLoopRunContext) LooperName() string { return "TestLoop" }
func (rc *FakeLoopRunContext) IsStopped() bool    { return false }
func (rc *FakeLoopRunContext) SetStopped()        {}

type FakeScopeContext struct {
}

func NewFakeScopeContext() *FakeScopeContext {
	return &FakeScopeContext{}
}

func (sc *FakeScopeContext) GetLoopRunContext() LoopRunContext            { return NewFakeLoopRunContext() }
func (sc *FakeScopeContext) GetLooperContext() ServiceContext             { return nil }
func (sc *FakeScopeContext) HasVariable(key string, localScope bool) bool { return false }
func (sc *FakeScopeContext) GetVariable(key string) interface{}           { return nil }
func (sc *FakeScopeContext) SetVariable(key string, value interface{})    {}
func (sc *FakeScopeContext) ExitScope(ScopeOption)                        {}
func (sc *FakeScopeContext) IsExit() bool                                 { return false }

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

type Component interface {
	DoWork()
}
type DefaultComponent struct {
	context dep.Context
	first   FirstInterface
	second  SecondInterface
	another AnotherInterface
}

func (c *DefaultComponent) DoWork() {
	logger := c.context.GetLogger()
	logger.Info("DefaultComponent DoWork start")
	c.first.First()
	c.second.Second()
	c.another.Another()
	logger.Info("DefaultComponent DoWork done")
}

func NewComponent(ctxt dep.Context, first FirstInterface, second SecondInterface, another AnotherInterface) *DefaultComponent {
	return &DefaultComponent{
		context: ctxt,
		first:   first,
		second:  second,
		another: another,
	}
}
