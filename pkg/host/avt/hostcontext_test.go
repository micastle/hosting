package avt

import (
	"testing"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
)

type TestConfig struct {
}

func Test_Host_context(t *testing.T) {
	hostName := "Test"
	runningMode := Debug

	builder := newActivatorBuilder(hostName, runningMode, nil)

	activator := builder.build()

	activator.configureComponents(func(context BuilderContext, components dep.ComponentCollection) {
		components.AddConfiguration(&TestConfig{})
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	})

	hostCtxt := activator.getContext()

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
	if result == nil {
		t.Error("unexpected nil properties")
	} else {
		count := len(result.Keys())
		if count > 0 {
			t.Errorf("unexpected property count: %d", count)
		}
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
