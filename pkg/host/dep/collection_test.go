package dep

import (
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func createCollection(ctxt Context, cm ComponentManager) (ComponentCollectionEx, ComponentProviderEx) {
	return NewComponentCollection(ctxt, cm), ctxt//NewDefaultProvider(ctxt, cm)
}

func TestCollection_basics(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)
	components, provider := createCollection(ctxt, cm)

	expected := &MyConfig{ value: 123 }
	AddConfig[MyConfig](components, expected)

	config := GetConfig[MyConfig](provider)
	if config != expected {
		t.Errorf("get config result %p is not expected: %p", config, expected)
	}

	singleton := NewActualStruct()
	RegisterInstance[FirstInterface](components, singleton)

	comp := GetComponent[FirstInterface](provider)
	if comp != singleton {
		t.Errorf("get component result %p is not expected: %p", comp, singleton)
	}

	RegisterSingleton[AnotherInterface](components, NewAnotherStruct)

	comp = GetComponent[FirstInterface](provider)
	if comp != singleton {
		t.Errorf("get component result %p is not expected: %p", comp, singleton)
	}

	RegisterTransient[SecondInterface](components, NewActualStruct)

	trans := CreateComponent[SecondInterface](provider, nil)
	if trans == nil {
		t.Errorf("create component result %p is nil", trans)
	}

	registered := IsComponentRegistered[SecondInterface](components)
	if !registered {
		t.Errorf("component is already registered, result is not expected: %v", registered)
	}

	expected_count := 13
	cnt := components.Count()
	if cnt != expected_count {
		t.Errorf("registered component count is not expected: %d", cnt)
	}
}

func TestCollection_scoped(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	options := NewComponentProviderOptions(InterfaceType, StructType)
	options.AllowTypeAnyFromFactoryMethod = true

	cm, ctxt := prepareComponentManagerWithScope(options, ScopeTest)
	components, provider := createCollection(ctxt, cm)

	components.RegisterScopedForType(NewActualStruct, types.Get[FirstInterface]())

	first := GetComponent[FirstInterface](provider)
	if first == nil {
		t.Errorf("create component FirstInterface result %p is nil", first)
	}

	RegisterScoped[SecondInterface, any](components, NewActualStruct)

	scoped := CreateComponent[SecondInterface](provider, nil)
	if scoped == nil {
		t.Errorf("create component SecondInterface result %p is nil", scoped)
	}
}