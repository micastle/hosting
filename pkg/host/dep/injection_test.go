package dep

import (
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func createLifecycleController(ctxt Context, options *LifecycleOptions) (LifecycleController, ComponentProviderEx) {
	return NewLifecycleController(ctxt, options), ctxt//NewDefaultProvider(ctxt, cm)
}

func TestInjection_basics(t *testing.T) {
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
	//func(context Context, actionName string, deps ...*Dependency[ComponentGetter])
	action(ctxt, "TestAction", 
		DepInst[AnotherInterface](NewAnotherStruct(ctxt)),
		DepInst[int](123),
		DepInst[string]("haha"),
	)
	if !done {
		t.Error("test action is not actually done")
	}
}
