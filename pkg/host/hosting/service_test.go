package hosting

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type MyService interface {
	Service

	Context() ServiceContext
	SetStopError(panic bool, stopErrMsg string)
}

type DefaultMyService struct {
	context      ServiceContext
	logger       logger.Logger
	resultWriter TestResultWriter
	value        int

	stopPanic  bool
	stopErrMsg string
}

func NewMyService(context ServiceContext, writer TestResultWriter) *DefaultMyService {
	s := &DefaultMyService{
		context:      context,
		logger:       context.GetLogger(),
		resultWriter: writer,
		value:        123,
	}
	s.logger.Infow("MyService created", "type", types.Of(s).Name())
	s.logger.Info("service context info", "type", context.Type(), "name", context.Name())

	testLogger := context.GetLoggerWithName("Test")
	testLogger.Infow("service context info", "type", context.Type(), "name", context.Name())

	return s
}
func (s *DefaultMyService) Context() ServiceContext {
	return s.context
}
func (s *DefaultMyService) SetStopError(panic bool, stopErrMsg string) {
	s.stopPanic = panic
	s.stopErrMsg = stopErrMsg
}

func (s *DefaultMyService) Start() {
	go s.Run()
}
func (s *DefaultMyService) Run() {
	s.runService()
}
func (s *DefaultMyService) Stop(ctx context.Context) error {
	if s.stopPanic {
		panic(fmt.Errorf("panic: %s", s.stopErrMsg))
	} else if s.stopErrMsg != "" {
		return errors.New(s.stopErrMsg)
	} else {
		return nil
	}

}

func (s *DefaultMyService) runService() {
	s.logger.Infow("MyService run start", "type", types.Of(s).Name())

	s.resultWriter.WriteResult("MyResult", s.value)
	s.logger.Infow("MyService Write result", "value", s.value)
}

func Test_service_basic(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollection) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		//components.RegisterSingletonForType(NewMyService, types.Of(new(MyService)))
	})

	UseService[MyService](builder, func(context ServiceContext) MyService {
		//service := context.GetComponent(types.Of(new(MyService))).(MyService)
		writer := dep.GetComponent[TestResultWriter](context)
		service := NewMyService(context, writer)
		return service
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(200) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("MyService didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != 123 {
			t.Errorf("MyService didn't write expected result: %v", result)
		}
	}
}

func Test_service_stop_error(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollection) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		//components.RegisterSingletonForType(NewMyService, types.Of(new(MyService)))
	})

	stopErrorMsg := "MyStopError"
	UseService[MyService](builder, func(context ServiceContext) MyService {
		//service := context.GetComponent(types.Of(new(MyService))).(MyService)
		writer := dep.GetComponent[TestResultWriter](context)
		service := NewMyService(context, writer)
		service.SetStopError(false, stopErrorMsg)
		return service
	})

	host := builder.Build().(GenericHost)

	provider := host.GetComponentProvider()
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	host.Execute()
	// service := host.GetContext().GetComponent(types.Of(new(MyService))).(MyService)
	// service.Run()

	err := host.Shutdown(time.Duration(5) * time.Second)
	if err == nil {
		t.Error("service stop error should be returned")
	} else {
		if err.Error() != stopErrorMsg {
			t.Errorf("shutdown returned err \"%v\" does not match service stop error msg: %s", err, stopErrorMsg)
		}
	}

	if !resultReader.HasResult("MyResult") {
		t.Errorf("MyService didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != 123 {
			t.Errorf("MyService didn't write expected result: %v", result)
		}
	}
}

func Test_service_stop_panic(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollection) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		//components.RegisterSingletonForType(NewMyService, types.Of(new(MyService)))
	})

	stopErrorMsg := "MyStopError"
	UseService[MyService](builder, func(context ServiceContext) MyService {
		//service := context.GetComponent(types.Of(new(MyService))).(MyService)
		writer := dep.GetComponent[TestResultWriter](context)
		service := NewMyService(context, writer)
		service.SetStopError(true, stopErrorMsg)
		return service
	})

	host := builder.Build().(GenericHost)

	provider := host.GetComponentProvider()
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	//host.Execute()
	service := dep.GetComponent[MyService](host.GetContext())
	service.Run()

	err := host.Shutdown(time.Duration(5) * time.Second)
	if err == nil {
		t.Error("service stop error should be returned")
	} else {
		expected := fmt.Sprintf("panic: %s", stopErrorMsg)
		if err.Error() != expected {
			t.Errorf("shutdown returned err \"%v\" does not match service stop error msg: %s", err, expected)
		}
	}

	if !resultReader.HasResult("MyResult") {
		t.Errorf("MyService didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != 123 {
			t.Errorf("MyService didn't write expected result: %v", result)
		}
	}
}

func Test_service_context(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollection) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		components.AddConfiguration(&TestConfig{})
		dep.RegisterTransient[AnotherInterface](components, NewAnotherStruct)
	})

	UseService[MyService](builder, func(context ServiceContext) MyService {
		//service := context.GetComponent(types.Of(new(MyService))).(MyService)
		writer := dep.GetComponent[TestResultWriter](context)
		service := NewMyService(context, writer)
		return service
	})

	host := builder.Build()

	service := dep.GetComponent[MyService](host.GetContext())
	serviceCtxt := service.Context().(ServiceContextEx)

	scopeCtxt := serviceCtxt.GetScopeContext()
	if !scopeCtxt.IsGlobal() {
		t.Error("service context must be in global context")
	}

	parent := scopeCtxt.GetParent()
	if parent != nil {
		t.Errorf("scope of service context must has no parent scope: %p", parent)
	}
	scopeId := scopeCtxt.ScopeId()
	if scopeId != dep.ScopeType_Global.Name() {
		t.Errorf("service scope id not expected: %s", scopeId)
	}
	if !scopeCtxt.IsDebug() {
		t.Error("default mode of service context is debug for unit tests")
	}
	ty := scopeCtxt.Type()
	name := scopeCtxt.Name()
	if ty != "Scope" || name != dep.ScopeType_Global.Name() {
		t.Errorf("service scope type and name not expected: %s, %s", ty, name)
	}

	scope := scopeCtxt.GetScope()
	id := scope.GetScopeId()
	if id != scopeId {
		t.Errorf("service scope id not equal to id of scope data: %s", id)
	}
	tyName := scope.GetTypeName()
	if tyName != dep.ScopeType_Global.Name() {
		t.Errorf("service scope type not expected: %s", tyName)
	}

	tracker := serviceCtxt.GetTracker()
	parentCtxt := tracker.GetParent()
	if parentCtxt != host.GetContext().(dep.HostContextEx).GetScopeContext() {
		t.Errorf("parent scope of service context must be the host global scope: name -%s, type - %s, scopeId - %s", parentCtxt.Type(), parentCtxt.Name(), parentCtxt.ScopeId())
	}

	props := dep.Props(dep.Pair("key1", 1), dep.Pair("key2", 2))
	serviceCtxt.UpdateProperties(props)
	props = dep.Props(dep.Pair("key2", 3), dep.Pair("key3", 4))
	serviceCtxt.UpdateProperties(props)
	result := serviceCtxt.GetProperties()
	val := result.Get("key2").(int)
	if val != 3 {
		t.Errorf("property value is not expected: %d", val)
	}

	config := dep.GetConfig[TestConfig](serviceCtxt)
	if config == nil {
		t.Error("get config returns nil")
	}
	ano := dep.CreateComponent[AnotherInterface](serviceCtxt, nil)
	ano.Another()
}
