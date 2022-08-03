package avt

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

func prepareActivator(configureComponents ConfigureComponentsMethod) Activator {
	return CreateActivator(configureComponents)
}

func prepareActivatorInMode(debug bool, configureComponents ConfigureComponentsMethod, loggerFactory logger.LoggerFactory) Activator {
	return CreateActivatorEx(debug, "UnitTest", nil, configureComponents, loggerFactory)
}

func prepareActivatorWithProperties(debug bool, props dep.Properties, configureComponents ConfigureComponentsMethod) Activator {
	return CreateActivatorEx(debug, "UnitTest", props, configureComponents, nil)
}

func Test_Activator_basic(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	}
	avt := prepareActivator(registerComponents)

	ano := GetComponent[AnotherInterface](avt)
	ano.Another()
}

func runTest_Activator_extended_debug_defaultlogging(t *testing.T, debug bool, loggerFactory logger.LoggerFactory) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	}
	avt := prepareActivatorInMode(debug, registerComponents, loggerFactory)

	ano := GetComponent[AnotherInterface](avt)
	ano.Another()
}

func Test_Activator_extended(t *testing.T) {
	loggerFactory := logger.NewDefaultLoggerFactory()
	tests := []struct {
		name          string
		debug         bool
		loggerFactory logger.LoggerFactory
	}{
		{
			"debug_with_default_logger",
			true,
			nil,
		},
		{
			"debug_with_custom_logger",
			true,
			loggerFactory,
		},
		{
			"release_with_default_logger",
			true,
			nil,
		},
		{
			"release_with_custom_logger",
			true,
			loggerFactory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest_Activator_extended_debug_defaultlogging(t, tt.debug, tt.loggerFactory)
		})
	}
}

func Test_Activator_comp_not_registered(t *testing.T) {
	defer test.AssertPanicContent(t, "dependency not configured, type: avt.AnotherInterface", "panic content is not expected")

	avt := prepareActivator(nil)

	ano := GetComponent[AnotherInterface](avt)
	ano.Another()
}

func Test_Activator_sys_component(t *testing.T) {
	avt := prepareActivator(nil)

	// context
	ctxt := GetComponent[dep.Context](avt)
	ctxtType := ctxt.Type()
	ctxtName := ctxt.Name()
	if ctxtType != dep.ContextType_Host && ctxtName != types.Of(new(Activator)).Name() {
		t.Errorf("unexpected activator context type %s and name %s", ctxtType, ctxtName)
	}

	// logger
	logger := GetComponent[logger.Logger](avt)
	logger.Infof("logger from component: %s %s", ctxtType, ctxtName)

	//properties
	props := GetComponent[dep.Properties](avt)
	if props == nil {
		t.Errorf("unexpected nil properties, context type %s and name %s", ctxtType, ctxtName)
	} else {
		count := len(props.Keys())
		if count > 0 {
			t.Errorf("unexpected property count: %d", count)
		}
	}

	// scope
	scope := GetComponent[dep.ScopeContext](avt)
	scopeType := scope.Type()
	scopeName := scope.Name()
	scopeId := scope.ScopeId()
	if scopeType != dep.ContextType_Scope && scopeName != dep.ScopeType_Global.Name() && scopeId != dep.ScopeType_Global.Name() {
		t.Errorf("unexpected scope context, type %s, name %s, id %s", scopeType, scopeName, scopeId)
	}

	// activator
	inst := GetComponent[Activator](avt)
	if inst != avt {
		t.Errorf("unexpected activator instance, actual %p, expected %p", inst, avt)
	}
}

func Test_Activator_multi_implementations(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterComponent(
			components,
			func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
			func(comp dep.CompImplCollection[Downloader, string]) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	}
	avt := prepareActivatorWithProperties(true, nil, registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)

	for _, Type := range []string{"url", "blob"} {
		downloader := dep.CreateComponent[Downloader](ctxt, dep.Props(dep.Pair("type", Type)))
		downloader.Download()

		if downloader.GetType() != Type {
			t.Errorf("expected - %s, actual - %s", Type, downloader.GetType())
		}

		if Type == "url" {
			// check singleton
			singleton := dep.CreateComponent[Downloader](ctxt, dep.Props(dep.Pair("type", "url")))
			if singleton != downloader {
				t.Error("singleton implementation should not have multiple instances!")
			}
		}
	}
}

func Test_Activator_multi_implementations_from_inheritprops(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterComponent(
			components,
			func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
			func(comp dep.CompImplCollection[Downloader, string]) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	}
	avt := prepareActivatorWithProperties(true, dep.Props(dep.Pair("type", "blob")), registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)

	downloader := dep.GetComponent[Downloader](ctxt)
	downloader.Download()

	if downloader.GetType() != "blob" {
		t.Errorf("expected - %s, actual - %s", "blob", downloader.GetType())
	}
}

func Test_Activator_multi_implementations_overwrite_inheritprops(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterComponent(
			components,
			func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
			func(comp dep.CompImplCollection[Downloader, string]) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	}
	avt := prepareActivatorWithProperties(true, dep.Props(dep.Pair("type", "blob")), registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)

	downloader := dep.CreateComponent[Downloader](ctxt, dep.Props(dep.Pair("type", "url")))
	downloader.Download()

	if downloader.GetType() != "url" {
		t.Errorf("should overwrite scope properties, expected - %s, actual - %s", "url", downloader.GetType())
	}
}

func Test_Activator_multi_implementations_props_missing(t *testing.T) {
	defer test.AssertPanicContent(t, "property \"type\" not exist", "panic content not expected")

	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterComponent(
			components,
			func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
			func(comp dep.CompImplCollection[Downloader, string]) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	}
	avt := prepareActivatorWithProperties(true, nil, registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)

	downloader := dep.GetComponent[Downloader](ctxt)
	downloader.Download()

	if downloader.GetType() != "url" {
		t.Errorf("should overwrite scope properties, expected - %s, actual - %s", "url", downloader.GetType())
	}
}

func Test_Activator_multi_implementations_impl_missing(t *testing.T) {
	defer test.AssertPanicContent(t, "component(avt.Downloader) implementation not exist for key not_exist", "panic content not expected")

	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterComponent(
			components,
			func(props dep.Properties) string { return dep.GetProp[string](props, "type") },
			func(comp dep.CompImplCollection[Downloader, string]) {
				compType := comp.GetComponentType()
				fmt.Printf("multi-impl componnent type: %s\n", compType.FullName())
				comp.AddSingletonImpl("url", NewUrlDownloader)
				comp.AddImpl("blob", NewBlobDownloader)
			},
		)
	}
	avt := prepareActivatorWithProperties(true, nil, registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)

	downloader := dep.CreateComponent[Downloader](ctxt, dep.Props(dep.Pair("type", "not_exist")))
	downloader.Download()

	if downloader.GetType() != "url" {
		t.Errorf("should overwrite scope properties, expected - %s, actual - %s", "url", downloader.GetType())
	}
}

func Test_Activator_scopefactory(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterTransient[ScopeInterface](components, NewScopeStruct)
	}
	avt := prepareActivator(registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)
	factory := GetComponent[dep.ScopeFactory](avt)

	// this is a bad example, scope is not disposed
	scope := factory.CreateScope(ctxt, nil)
	scope.Execute("TestFunc1", func(ctxt dep.Context, logger logger.Logger) {
		logger.Infof("scope func start with context %s, %s", ctxt.Type(), ctxt.Name())
	})

	inst := GetComponent[ScopeInterface](avt)
	dep.Using(inst, func(ctxt dep.Context, scope dep.Scope, logger logger.Logger) {
		logger.Infof("scope func start with context %s, %s", ctxt.Type(), ctxt.Name())
		logger.Infof("scope: type - %s, id - %s", scope.GetTypeName(), scope.GetScopeId())
	})
}

func Test_Activator_scopectxt(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollection) {
		dep.RegisterTransient[ScopeInterface](components, NewScopeStruct)
	}
	avt := prepareActivator(registerComponents)

	// scope
	scopeCtxt := GetComponent[dep.ScopeContext](avt)
	if !scopeCtxt.IsGlobal() {
		t.Error("top scope must be global")
	}
}

// target types for test
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
	Another()
}

type AnotherStruct struct {
	value int
}

func NewAnotherStruct() *AnotherStruct {
	return &AnotherStruct{
		value: 0,
	}
}

func (as *AnotherStruct) Another() {
	fmt.Println("Another", as.value)
}

type ScopeInterface interface {
	dep.Scopable
	Doit()
}

type ScopeStruct struct {
	context dep.Context
	value   int
}

func NewScopeStruct(context dep.Context) *ScopeStruct {
	return &ScopeStruct{
		context: context,
		value:   0,
	}
}

func (ss *ScopeStruct) Context() dep.Context {
	return ss.context
}
func (ss *ScopeStruct) Doit() {
	fmt.Printf("ScopeStruct: %d\n", ss.value)
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
