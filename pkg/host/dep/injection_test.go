package dep

import (
	"testing"
	"errors"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func createLifecycleController(ctxt Context, options *LifecycleOptions) (LifecycleController, ComponentProviderEx) {
	return NewLifecycleController(ctxt, options), ctxt//NewDefaultProvider(ctxt, cm)
}

func TestInjection_Action_basics(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	RegisterTransient[FirstInterface](cm, NewActualStruct)

	done := false
	action := ctrl.BuildActionMethod(func(context Context, log logger.Logger, first FirstInterface, another AnotherInterface, value int, name string){
		t.Logf("action injected with int %d, string %s", value, name)
		done = true
	})
	action(ctxt, "TestAction", 
		DepInst[AnotherInterface](NewAnotherStruct(ctxt)),
		DepInst[int](123),
		DepInst[string]("haha"),
	)
	if !done {
		t.Error("test action is not actually done")
	}
}

func TestInjection_Action_return_error_nil(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	_, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	action := ctrl.BuildActionMethod(func() error { return nil })
	action(ctxt, "TestAction")
}
func TestInjection_Action_return_error_nil2(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	_, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	action := ctrl.BuildActionMethod(func() error {
		var err error
		return err
	})
	action(ctxt, "TestAction")
}

func TestInjection_Action_return_error(t *testing.T) {
	defer test.AssertPanicContent(t, "action TestAction error: failed to execute action", "panic content not expected")

	_, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	action := ctrl.BuildActionMethod(func() error { return errors.New("failed to execute action") })
	action(ctxt, "TestAction")
}

func TestInjection_Action_return_non_error(t *testing.T) {
	defer test.AssertPanicContent(t, "action TestAction should only return error or nothing, actual type:", "panic content not expected")

	_, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	action := ctrl.BuildActionMethod(func() int { return 123 })
	action(ctxt, "TestAction")
}

func TestInjection_Action_return_multi_outputs(t *testing.T) {
	defer test.AssertPanicContent(t, "action TestAction should return no more than 1 outputs", "panic content not expected")

	_, ctxt := prepareComponentManager(true)
	options := &LifecycleOptions{
		EnableDiagnostics:          true,
		EnableSingletonConcurrency: true,
		TrackTransientRecurrence:   true,
		MaxAllowedRecurrence:       3,
	}
	ctrl, _ := createLifecycleController(ctxt, options)

	action := ctrl.BuildActionMethod(func() (int, error) { return 123, errors.New("failed to execute action") })
	action(ctxt, "TestAction")
}

func TestInjection_Factory_basics(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () any { return NewActualStruct() })

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}

func TestInjection_Factory_return_error_nil(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () (any, error) { return NewActualStruct(), nil })

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}

func TestInjection_Factory_return_error_nil2(t *testing.T) {
	defer test.AssertNoPanic(t, "no panic is expected in this test")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () (any, error) {
		var err error
		return NewActualStruct(), err 
	})

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}

func TestInjection_Factory_return_error(t *testing.T) {
	defer test.AssertPanicContent(t, "factory method error: failed to create instance", "panic content not expected")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () (any, error) { return NewActualStruct(), errors.New("failed to create instance") })

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}

func TestInjection_Factory_return_non_error(t *testing.T) {
	defer test.AssertPanicContent(t, "second output of factory method should of type error, actual - int", "panic content not expected")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () (any, int) { return NewActualStruct(), 123 })

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}

func TestInjection_Factory_return_multi_outputs(t *testing.T) {
	defer test.AssertPanicContent(t, "factory method should return no more than 2 outputs", "panic content not expected")

	cm, ctxt := prepareComponentManager(true)

	RegisterTransient[FirstInterface](cm, func () (any, int, error) { return NewActualStruct(), 123, errors.New("failed to execute action") })

	inst := GetComponent[FirstInterface](ctxt)
	inst.First()
}