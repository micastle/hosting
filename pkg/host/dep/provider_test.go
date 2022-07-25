package dep

import (
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func createProvider(ctxt Context, provider ContextualProvider) ComponentProviderEx {
	return NewDefaultComponentProvider(ctxt, provider)
}

func TestProvider_basics(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)
	provider := createProvider(ctxt, cm)

	expected := &MyConfig{ value: 123 }
	AddConfig[MyConfig](cm, expected)

	config := GetConfig[MyConfig](provider)
	if config != expected {
		t.Errorf("get config result %p is not expected: %p", config, expected)
	}

	singleton := NewActualStruct()
	RegisterInstance[FirstInterface](cm, singleton)

	comp := GetComponent[FirstInterface](provider)
	if comp != singleton {
		t.Errorf("get component result %p is not expected: %p", comp, singleton)
	}

	RegisterTransient[SecondInterface](cm, NewActualStruct)

	trans := CreateComponent[SecondInterface](provider, nil)
	if trans == nil {
		t.Errorf("create component result %p is nil", trans)
	}
}
