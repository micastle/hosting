package avt

import (
	"fmt"
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func Test_Activator_basic(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	}
	avt := CreateActivator(registerComponents)

	ano := GetComponent[AnotherInterface](avt)
	ano.Another()
}

func Test_Activator_comp_not_registered(t *testing.T) {
	defer test.AssertPanicContent(t, "dependency not configured, type: avt.AnotherInterface", "panic content is not expected")

	avt := CreateActivator(nil)

	ano := GetComponent[AnotherInterface](avt)
	ano.Another()
}

func Test_Activator_sys_component(t *testing.T) {
	avt := CreateActivator(nil)

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

	// properties
	// props := GetComponent[dep.Properties](avt)
	// if props != nil {
	// 	t.Errorf("unexpected properties, context type %s and name %s", ctxtType, ctxtName)
	// }

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


// TODO: multi-impls support
func Test_Activator_multi_implementations(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	}
	avt := CreateActivator(registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)
	ano := dep.CreateComponent[AnotherInterface](ctxt, dep.Props())
	ano.Another()
}

func Test_Activator_scopefactory(t *testing.T) {
	registerComponents := func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterTransient[ScopeInterface](components, NewScopeStruct)
	}
	avt := CreateActivator(registerComponents)

	// context
	ctxt := GetComponent[dep.Context](avt)
	factory := GetComponent[dep.ScopeFactory](avt)

	// this is a bad example, scope is not disposed
	scope := factory.CreateScope(ctxt)
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
	registerComponents := func(context BuilderContext, components dep.ComponentCollectionEx) {
		dep.RegisterTransient[ScopeInterface](components, NewScopeStruct)
	}
	avt := CreateActivator(registerComponents)

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
