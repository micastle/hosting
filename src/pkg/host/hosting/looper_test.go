package hosting

import (
	"fmt"
	"testing"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

type TestResultWriter interface {
	WriteResult(key string, result interface{})
}

type TestResultReader interface {
	HasResult(key string) bool
	GetResult(key string) interface{}
}

type SimpleTestResultStore struct {
	data map[string]interface{}
}

func NewTestResultStore() *SimpleTestResultStore {
	return &SimpleTestResultStore{
		data: make(map[string]interface{}),
	}
}

func (trs *SimpleTestResultStore) WriteResult(key string, result interface{}) {
	trs.data[key] = result
}
func (trs *SimpleTestResultStore) HasResult(key string) bool {
	_, exist := trs.data[key]
	return exist
}
func (trs *SimpleTestResultStore) GetResult(key string) interface{} {
	return trs.data[key]
}

type MyProc LoopProcessor

type MyProcessor struct {
	logger       logger.Logger
	resultWriter TestResultWriter
	value        int
}

var start_value int = 123
var exit_round int = 4

func NewMyProcessor(context dep.Context, writer TestResultWriter) *MyProcessor {
	p := &MyProcessor{
		logger:       context.GetLogger(),
		resultWriter: writer,
		value:        start_value,
	}
	p.logger.Infow("MyProcessor created", "type", types.Of(p).Name())
	return p
}

func (p *MyProcessor) Run(ctxt ScopeContext) {
	looperCtxt := ctxt.GetLooperContext()
	runCtxt := ctxt.GetLoopRunContext()
	p.logger.Infow("MyProcessor run start", "looper", runCtxt.LooperName(), "type", looperCtxt.Name(), "processor", types.Of(p).Name())

	intVal := 234
	SetVariable(ctxt, "MyValue", intVal)
	val := GetVariable[int](ctxt, "MyValue")

	if val != intVal {
		p.logger.Errorw("variable value is not the same as expected", "actual", val, "expected", intVal)
	}

	p.resultWriter.WriteResult("MyResult", p.value)
	fmt.Printf("MyResult: %d\n", p.value)
	p.logger.Infow("MyProcessor Write result", "value", p.value)
	p.value = p.value + 1
	if p.value >= start_value+exit_round {
		ctxt.ExitScope(TopLevel)
	}
}

func Test_looper_basic(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		looper.ConfigureLogger(func(context ServiceContext, log logger.Logger) logger.Logger {
			return logger.With(log, "Loop", "Test")
		})
		UseProcessor[MyProc](looper, nil)
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	expected_round := 1
	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != start_value+expected_round {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}

func Test_looper_processor_type_not_interface(t *testing.T) {
	defer test.AssertPanicContent(t, "specified processor type is not an interface: hosting.MyConfig", "panic content not expected")

	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		UseProcessor[MyConfig](looper, nil)
	})

	builder.Build()
}

type MyInterface interface {}
func Test_looper_processor_type_not_LoopProcessor(t *testing.T) {
	defer test.AssertPanicContent(t, "specified processor type does not implement LoopProcessor interface", "panic content not expected")

	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		UseProcessor[MyInterface](looper, nil)
	})

	builder.Build()
}

type MyFuncProc FunctionProcessor

func Test_looper_func_processor(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	expected := 789
	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		RegisterFuncProcessor[MyFuncProc](components, func(resultWriter TestResultWriter) {
			resultWriter.WriteResult("MyResult", expected)
		})
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		UseProcessor[MyFuncProc](looper, nil)
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != expected {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}

func Test_looper_anon_func_processor(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
	})

	expected := 357
	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		looper.UseFuncProcessor(func(resultWriter TestResultWriter) {
			resultWriter.WriteResult("MyResult", expected)
		})
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != expected {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}

func Test_looper_conditional_processor_false(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)

		UseProcessor[MyProc](looper, func(context ScopeContext) bool { return false })
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if resultReader.HasResult("MyResult") {
		t.Errorf("processor with false condition should not run and write result")
	}
}

func Test_looper_conditional_processor_true(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)

		UseProcessor[MyProc](looper, func(context ScopeContext) bool { return true })
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	// wait for stop signal from console
	host.Run()

	expected_round := 1
	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != start_value+expected_round {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}

func Test_looper_processor_group_true(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)

		looper.UseProcessorGroup(func(context dep.Context, group GroupContext) {
			group.SetGroupName("TestGroup")
			UseProcessor[MyProc](group, nil)
		}, func(context ScopeContext) bool { return true })
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	expected_round := 1
	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != start_value+expected_round {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}

func Test_looper_processor_group_exitscope(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	exeCount := 0
	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)

		looper.UseProcessorGroup(func(context dep.Context, group GroupContext) {
			group.SetGroupName("TestGroup")
			UseProcessor[MyProc](group, nil)
		}, func(context ScopeContext) bool { return true })
		looper.UseFuncProcessor(func() {
			exeCount = exeCount + 1
		})
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(2300) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result-start_value < exit_round {
			// exit group didn't happen
			t.Errorf("executed round didn't achieve exist round(%d), longer sleep time? - %d, %d", exit_round, result-start_value, exeCount)
		}
	}

	if exeCount >= exit_round {
		t.Errorf("exeCount should less than %d due to scope exit: %d", exit_round, exeCount)
	}
}

func Test_looper_processor_group_vars(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
	})

	varError := false
	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(200) * time.Millisecond)

		looper.ConfigureLoopGlobalContext(func(context LoopGlobalContext) {
			SetVariable(context, "GlobalContext", 123)
		})
		looper.ConfigureScopeContext(func(context ScopeContext) {
			SetVariable(context, "ScopeContext", 456)
		})
		looper.UseFuncProcessor(func(context ScopeContext) {
			if context.HasVariable("GlobalContext", true) {
				fmt.Errorf("GlobalContext does not exist in local scope context")
				varError = true
			}
			if !context.HasVariable("GlobalContext", false) {
				fmt.Errorf("GlobalContext should exist in looper global context")
				varError = true
			} else {
				globalVal := GetVariable[int](context, "GlobalContext")
				if globalVal != 123 {
					fmt.Errorf("GlobalContext value not expected: %d", globalVal)
					varError = true
				}
			}
			if !context.HasVariable("ScopeContext", true) {
				fmt.Errorf("ScopeContext should exist in local scope context")
				varError = true
			}
			if !context.HasVariable("ScopeContext", false) {
				fmt.Errorf("ScopeContext should always exist in local scope context")
				varError = true
			} else {
				scopeVal := GetVariable[int](context, "ScopeContext")
				if scopeVal != 456 {
					fmt.Errorf("ScopeContext value not expected: %d", scopeVal)
					varError = true
				}
			}
		})
		looper.UseProcessorGroup(func(context dep.Context, group GroupContext) {
			group.SetGroupName("TestGroup")

			group.ConfigureScopeContext(func(context ScopeContext) {
				SetVariable(context, "GroupScopeContext", 789)
			})
			group.UseFuncProcessor(func(context ScopeContext) {
				if context.HasVariable("GlobalContext", true) {
					fmt.Errorf("GlobalContext does not exist in local group context")
					varError = true
				}
				if !context.HasVariable("GlobalContext", false) {
					fmt.Errorf("GlobalContext should exist in looper global context")
					varError = true
				} else {
					globalVal := context.GetVariable("GlobalContext").(int)
					if globalVal != 123 {
						fmt.Errorf("GlobalContext value from group not expected: %d", globalVal)
						varError = true
					}
				}
				if context.HasVariable("ScopeContext", true) {
					fmt.Errorf("ScopeContext does not exist in local group context")
					varError = true
				}
				if !context.HasVariable("ScopeContext", false) {
					fmt.Errorf("ScopeContext should exist in looper scope context")
					varError = true
				} else {
					scopeVal := context.GetVariable("ScopeContext").(int)
					if scopeVal != 456 {
						fmt.Errorf("ScopeContext value from group not expected: %d", scopeVal)
						varError = true
					}
				}
				if context.HasVariable("GroupScopeContext", true) {
					groupVal := context.GetVariable("GroupScopeContext").(int)
					if groupVal != 789 {
						fmt.Errorf("GroupScopeContext value not expected: %d", groupVal)
						varError = true
					}
				} else {
					fmt.Errorf("GroupScopeContext should exist in local group context")
					varError = true
				}
				if !context.HasVariable("GroupScopeContext", false) {
					fmt.Errorf("GroupScopeContext should always exist in local group context")
					varError = true
				}
			})
		}, func(context ScopeContext) bool { return true })
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(500) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if varError {
		t.Errorf("has context variable status error, see above error msg")
	}
}

func Test_looper_scope_context(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	start := 456
	scopeVal := start
	max_round := 3
	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(200) * time.Millisecond)

		looper.ConfigureScopeContext(func(context ScopeContext) {
			SetVariable(context, "ScopeContext", scopeVal)
			fmt.Printf("scopeVal: %d\n", scopeVal)
			scopeVal = scopeVal + 1
		})

		UseProcessor[MyProc](looper, func(context ScopeContext) bool {
			val := context.GetVariable("ScopeContext").(int)
			return val < start+max_round
		})
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(2500) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != start_value+max_round-1 {
			t.Errorf("conditional processor should not run when scope context value is increased: %v", result)
		}
	}
}

func Test_looper_global_context(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	start := 456
	globalVal := start
	max_count := 3

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(200) * time.Millisecond)

		looper.ConfigureLoopGlobalContext(func(context LoopGlobalContext) {
			SetVariable(context, "GlobalContext", globalVal)
			globalVal = globalVal + 1
		})

		UseProcessor[MyProc](looper, func(context ScopeContext) bool {
			val := context.GetVariable("GlobalContext").(int)
			return val < start+max_count
		})
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(2200) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result <= start_value+max_count-1 {
			t.Errorf("conditional processor should not stop when as global context will not be updated for each loop: %v", result)
		}
	}
}

func Test_looper_scope_context_variable_not_exist(t *testing.T) {
	hostName := "Test"
	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[MyProc](components, NewMyProcessor)
	})

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(200) * time.Millisecond)
		looper.SetRecover(true)

		looper.ConfigureScopeContext(func(context ScopeContext) {
			_ = context.GetVariable("MyVar") // raise panic, code below will never run
		})

		UseProcessor[MyProc](looper, nil)
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(800) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if resultReader.HasResult("MyResult") {
		t.Errorf("should not have result due to variable not exist in processor")
	}
}

type TestScope interface {
	dep.Scopable

	SetValue(val int)
	GetValue() int
}
type DefaultTestScope struct {
	context dep.Context
	value   int
}

func NewTestScope(context dep.Context) *DefaultTestScope {
	return &DefaultTestScope{context: context}
}
func (ts *DefaultTestScope) Context() dep.Context {
	return ts.context
}
func (ts *DefaultTestScope) SetValue(val int) {
	ts.value = val
}
func (ts *DefaultTestScope) GetValue() int {
	return ts.value
}

func Test_looper_with_scope(t *testing.T) {
	hostName := "Test"

	builder := NewDefaultHostBuilder()
	builder.SetHostName(hostName)

	builder.ConfigureComponents(func(context BuilderContext, components dep.ComponentCollectionEx) {
		components.RegisterSingletonForTypes(NewTestResultStore, types.Get[TestResultWriter](), types.Get[TestResultReader]())
		dep.RegisterTransient[TestScope](components, NewTestScope)
		dep.RegisterScoped[AnotherInterface, TestScope](components, NewAnotherStruct)
	})
	testFunc := func(context dep.Context, test TestScope, resultWriter TestResultWriter) {

		ano := dep.GetComponent[AnotherInterface](context)

		ano.Another()

		resultWriter.WriteResult("MyResult", test.GetValue())
	}
	procFunc := func(scopeFactory dep.ScopeFactory) {
		scope := dep.CreateTypedScope[TestScope](scopeFactory)
		defer func() { scope.Dispose() }()

		_, inst := scope.GetScopeContext().(dep.ScopeContextEx).GetScope().GetInstance()
		test := inst.(TestScope)
		test.SetValue(124)

		testMethod := scope.BuildActionMethod("RunTest", testFunc)
		testMethod()

	}

	builder.UseLoop("Test", func(context ServiceContext, looper ConfigureLoopContext) {
		looper.SetInterval(time.Duration(500) * time.Millisecond)
		looper.UseFuncProcessor(procFunc)
	})

	host := builder.Build()

	provider := host.GetComponentProvider()
	runner := dep.GetComponent[AsyncAppRunner](provider)
	resultReader := dep.GetComponent[TestResultReader](provider)

	logger := host.GetLogger()
	logger.Infow("Host ready, starting", "name", host.GetName())

	go func() {
		time.Sleep(time.Duration(750) * time.Millisecond)
		runner.SendStopSignal()
	}()

	host.Run()

	if !resultReader.HasResult("MyResult") {
		t.Errorf("processor didn't write result yet")
	} else {
		result := resultReader.GetResult("MyResult").(int)
		if result != 124 {
			t.Errorf("processor didn't write expected result: %v", result)
		}
	}
}
