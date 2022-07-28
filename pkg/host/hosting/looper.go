package hosting

import (
	"context"
	"fmt"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type ConfigureLoopMethod func(context ServiceContext, looper ConfigureLoopContext)
type ConfigureLoopLoggerMethod func(context ServiceContext, logger logger.Logger) logger.Logger
type ConfigureLoopGroupMethod func(context dep.Context, group GroupContext)

type LoopGlobalContextInitMethod func(LoopGlobalContext)
type ScopeContextInitMethod func(ScopeContext)

type ConfigureGroupContext interface {
	ConfigureScopeContext(ScopeContextInitMethod)
	UseProcessor(types.DataType)
	UseConditionalProcessor(types.DataType, ConditionMethod)
	UseProcessorGroup(ConfigureLoopGroupMethod, ConditionMethod)
	UseFuncProcessor(procFunc dep.FreeStyleProcessorMethod)
}

func UseProcessor[T any](group ConfigureGroupContext, condition ConditionMethod) {
	if condition == nil {
		group.UseProcessor(types.Get[T]())
	} else {
		group.UseConditionalProcessor(types.Get[T](), condition)
	}
}

type ConfigureLoopContext interface {
	ConfigureGroupContext

	SetInterval(time.Duration)
	SetRecover(enabled bool)
	ConfigureLogger(ConfigureLoopLoggerMethod)
	ConfigureLoopGlobalContext(LoopGlobalContextInitMethod)
}

type GroupContext interface {
	ConfigureGroupContext

	SetGroupName(string)
}

type LooperSettings struct {
	Name      string
	Interval  time.Duration
	Recover   bool
	Configure ConfigureLoopMethod
}

type Looper interface {
	Name() string

	Run()
	Stop(ctx context.Context) error
}

type ConditionMethod func(context ScopeContext) bool

type LoopProcessor interface {
	Run(ctxt ScopeContext)
}

type CreateProcessorMethod func(context dep.Context, interfaceType types.DataType, props dep.Properties) LoopProcessor

type ProcessorGroupBuilder interface {
	AddProcessor(processorType types.DataType, createInstance CreateProcessorMethod, condition ConditionMethod)
}
type ProcessorGroup interface {
	SetName(string)
	SetScopeContextInitializer(ScopeContextInitMethod)
	AddProcessor(processorType types.DataType, createInstance CreateProcessorMethod, condition ConditionMethod)

	SetLooperContext(LooperContext)
	GetLooperContext() LooperContext

	Initialize()

	Run(parent ScopeContext)
	RunNewIteration(global LoopGlobalContext)
}

type DefaultProcessorGroup struct {
	context          dep.Context
	logger           logger.Logger
	looper           LooperContext
	name             string
	initGroupContext ScopeContextInitMethod
	processors       []*ProcessorRecord
}

func NewDefaultProcessorGroup(context dep.Context) *DefaultProcessorGroup {
	pg := &DefaultProcessorGroup{
		context:    context,
		processors: make([]*ProcessorRecord, 0),
	}

	return pg
}

func (pg *DefaultProcessorGroup) SetName(name string) {
	pg.name = name
}
func (pg *DefaultProcessorGroup) SetScopeContextInitializer(initScopeContext ScopeContextInitMethod) {
	pg.initGroupContext = initScopeContext
}
func (pg *DefaultProcessorGroup) AddProcessor(processorType types.DataType, createInstance CreateProcessorMethod, condition ConditionMethod) {
	record := &ProcessorRecord{
		Type:      processorType,
		Condition: condition,
		Instance:  createInstance(pg.context, processorType, nil),
	}
	pg.processors = append(pg.processors, record)
}

func (pg *DefaultProcessorGroup) SetLooperContext(looper LooperContext) {
	pg.looper = looper
}

func (pg *DefaultProcessorGroup) GetLooperContext() LooperContext {
	return pg.looper
}

func (pg *DefaultProcessorGroup) LooperName() string {
	return pg.looper.GetName()
}
func (pg *DefaultProcessorGroup) Name() string {
	return pg.name
}

// func (pg *DefaultProcessorGroup) FullName() string {
// 	if pg.Name() == "" {
// 		return pg.LooperName() // TODO: use full path to the group instead of looper name?
// 	}
// 	return fmt.Sprintf("%s.%s", pg.LooperName(), pg.Name())
// }

func (pg *DefaultProcessorGroup) getLooperLoggerName() string {
	return fmt.Sprintf("Loop[%s]", pg.LooperName())
}
func (pg *DefaultProcessorGroup) getLoggerName() string {
	groupName := pg.Name()
	if groupName == "" {
		groupName = dep.GetDefaultLoggerNameForComponent(pg)
	}
	return fmt.Sprintf("%s.%s", pg.getLooperLoggerName(), groupName)
}

// init happens after instance is created and all context configuration are done
func (pg *DefaultProcessorGroup) Initialize() {
	pg.logger = pg.context.GetLoggerWithName(pg.getLoggerName())
}

func (pg *DefaultProcessorGroup) Run(parent ScopeContext) {
	// create context for the group scope
	groupCtxt := NewGroupScopeContext(pg, parent)

	pg.runWithContext(groupCtxt)
}

func (pg *DefaultProcessorGroup) RunNewIteration(global LoopGlobalContext) {
	// create context for new iteration run of the loop
	runContext := NewLoopRunContext(global)

	pg.runWithContext(runContext)
}

func (pg *DefaultProcessorGroup) runWithContext(groupCtxt ScopeContext) {
	if pg.initGroupContext != nil {
		pg.initGroupContext(groupCtxt)
	}

	// start running loop and execute all processors
	for _, record := range pg.processors {
		// check processor condition if exist
		if record.Condition != nil {
			if !record.Condition(groupCtxt) {
				continue
			}
		}

		pg.logger.Debugw("run loop processor", "looper", pg.LooperName(), "processor", record.Type.Name())
		record.Instance.Run(groupCtxt)

		// check context complete or loop run stopped
		if groupCtxt.IsExit() || groupCtxt.GetLoopRunContext().IsStopped() {
			pg.logger.Debugw("the scope of current group is marked as complete by processor", "name", record.Type.Name(), "looper", pg.LooperName())
			break
		}
	}
}

type DefaultLoopContext struct {
	looper       *DefaultLooper
	groupContext *DefaultGroupContext
}

func NewDefaultLoopContext(looper *DefaultLooper) *DefaultLoopContext {
	lc := &DefaultLoopContext{
		looper:       looper,
		groupContext: NewDefaultGroupContext(looper.processorGroup),
	}
	return lc
}

func (lc *DefaultLoopContext) GetName() string {
	return lc.looper.name
}

// func (lc *DefaultLoopContext) GetInterval() time.Duration {
// 	return lc.looper.timerInterval
// }

func (lc *DefaultLoopContext) SetInterval(interval time.Duration) {
	lc.looper.timerInterval = interval
}
func (lc *DefaultLoopContext) SetRecover(enabled bool) {
	lc.looper.enableRecover = enabled
}
func (lc *DefaultLoopContext) ConfigureLogger(configLogger ConfigureLoopLoggerMethod) {
	lc.looper.configLogger(configLogger)
}
func (lc *DefaultLoopContext) ConfigureLoopGlobalContext(initLoopContext LoopGlobalContextInitMethod) {
	lc.looper.initLoopContext = initLoopContext
}
func (lc *DefaultLoopContext) ConfigureScopeContext(initScopeContext ScopeContextInitMethod) {
	lc.looper.processorGroup.SetScopeContextInitializer(initScopeContext)
}

func (lc *DefaultLoopContext) UseProcessor(processorType types.DataType) {
	lc.UseConditionalProcessor(processorType, nil)
}
func (lc *DefaultLoopContext) UseConditionalProcessor(processorType types.DataType, condition ConditionMethod) {
	lc.groupContext.UseConditionalProcessor(processorType, condition)
}
func (lc *DefaultLoopContext) UseProcessorGroup(configureGroup ConfigureLoopGroupMethod, condition ConditionMethod) {
	lc.groupContext.UseProcessorGroup(configureGroup, condition)
}
func (lc *DefaultLoopContext) UseFuncProcessor(procFunc dep.FreeStyleProcessorMethod) {
	lc.groupContext.UseFuncProcessor(procFunc)
}

type DefaultGroupContext struct {
	group ProcessorGroup
}

func NewDefaultGroupContext(group ProcessorGroup) *DefaultGroupContext {
	return &DefaultGroupContext{
		group: group,
	}
}

func (gc *DefaultGroupContext) validateProcessorType(processorType types.DataType) {
	if !processorType.IsInterface() {
		panic(fmt.Errorf("specified processor type is not an interface: %s", processorType.FullName()))
	}
	if !processorType.CheckCompatible(types.Get[LoopProcessor]()) {
		panic(fmt.Errorf("specified processor type does not implement LoopProcessor interface: %s", processorType.FullName()))
	}
}

func (gc *DefaultGroupContext) SetGroupName(name string) {
	gc.group.SetName(name)
}
func (gc *DefaultGroupContext) ConfigureScopeContext(initScopeContext ScopeContextInitMethod) {
	gc.group.SetScopeContextInitializer(initScopeContext)
}
func (gc *DefaultGroupContext) UseProcessor(processorType types.DataType) {
	gc.UseConditionalProcessor(processorType, nil)
}
func (gc *DefaultGroupContext) UseConditionalProcessor(processorType types.DataType, condition ConditionMethod) {
	gc.validateProcessorType(processorType)
	createInstance := func(context dep.Context, interfaceType types.DataType, props dep.Properties) LoopProcessor {
		return context.GetComponent(interfaceType).(LoopProcessor)
	}
	gc.group.AddProcessor(processorType, createInstance, condition)
}
func (gc *DefaultGroupContext) UseProcessorGroup(configureGroup ConfigureLoopGroupMethod, condition ConditionMethod) {
	processorType := types.Of(new(ProcessorGroup))

	createInstance := func(context dep.Context, interfaceType types.DataType, props dep.Properties) LoopProcessor {
		instance := dep.GetComponent[ProcessorGroup](context)
		instance.SetLooperContext(gc.group.GetLooperContext())

		groupContext := NewDefaultGroupContext(instance)
		configureGroup(context, groupContext)

		instance.Initialize()
		return instance
	}

	gc.group.AddProcessor(processorType, createInstance, condition)
}

type AnonFuncProcessor FunctionProcessor

func (gc *DefaultGroupContext) UseFuncProcessor(procFunc dep.FreeStyleProcessorMethod) {
	createProcessor := func(context dep.Context, interfaceType types.DataType, props dep.Properties) LoopProcessor {
		return createFuncProcess[AnonFuncProcessor](context, procFunc)
	}
	gc.group.AddProcessor(types.Get[AnonFuncProcessor](), createProcessor, nil)
}

type ProcessorRecord struct {
	Type      types.DataType
	Condition ConditionMethod
	Instance  LoopProcessor
}

type LooperContext interface {
	GetName() string
}

type DefaultLooper struct {
	context ServiceContext
	name    string
	logger  logger.Logger
	runner  *LoopRunner

	// settings
	timerInterval   time.Duration
	enableRecover   bool
	initLoopContext LoopGlobalContextInitMethod

	processorGroup ProcessorGroup
}

func NewDefaultLooper(context ServiceContext) *DefaultLooper {
	lp := &DefaultLooper{
		context: context,
	}
	return lp
}

const minLoopInterval = 500 * time.Millisecond
const maxStopInterval = 500 * time.Millisecond

func (lp *DefaultLooper) Initialize(settings *LooperSettings) {
	lp.name = settings.Name
	lp.timerInterval = settings.Interval
	lp.enableRecover = settings.Recover

	lp.logger = lp.context.GetLoggerWithName(lp.getLoggerName())
	lp.logger.Debugw("initializing Looper", "name", lp.name)

	lp.processorGroup = dep.GetComponent[ProcessorGroup](lp.context)

	loopContext := NewDefaultLoopContext(lp)
	lp.processorGroup.SetLooperContext(loopContext)
	settings.Configure(lp.context, loopContext)

	lp.runner = NewLoopRunner(LoopRunnerSettings{
		EnableRecover:   lp.enableRecover,
		MinLoopInterval: minLoopInterval,
		MaxStopInterval: maxStopInterval,
	})
	lp.runner.Initialize(func() any {
		// initialize looper context before loop start
		loopContext := NewLoopGlobalContext(lp)
		if lp.initLoopContext != nil {
			lp.initLoopContext(loopContext)
		}
		return LoopGlobalContext(loopContext)
	})

	lp.processorGroup.Initialize()
}

func (lp *DefaultLooper) configLogger(configLogger ConfigureLoopLoggerMethod) {
	lp.logger = configLogger(lp.context, lp.logger)
}

// implement interface LooperContext
func (lp *DefaultLooper) Name() string {
	return lp.name
}

// implement interface LoggerContract
func (lp *DefaultLooper) getLoggerName() string {
	return fmt.Sprintf("Loop[%s]", lp.Name())
}

func (lp *DefaultLooper) Run() {
	lp.logger.Debugw("Looper started to run", "name", lp.Name())

	lp.runner.Run(lp.timerInterval, func(ctxt any) {
		loopContext := ctxt.(LoopGlobalContext)
		lp.runIteration(loopContext)
	})
}
func (lp *DefaultLooper) runIteration(loopContext LoopGlobalContext) {
	lp.logger.Debugw("Looper start new iteration", "Name", lp.Name())
	start := time.Now()
	lp.processorGroup.RunNewIteration(loopContext)
	cost := float64(time.Since(start).Milliseconds())
	lp.logger.Debugw("Looper completed one iteration", "Name", lp.Name(), "Cost(ms)", cost)
}

func (lp *DefaultLooper) Stop(ctx context.Context) error {
	lp.logger.Debugw("shutting down Looper", "name", lp.Name())

	return lp.runner.Stop(ctx)
}

// utility API: looper factory method
func createLooper(depCtxt dep.Context, interfaceType types.DataType, props dep.Properties) any {
	dependent := depCtxt.(dep.ContextEx)
	scopeCtxt := dependent.GetScopeContext()
	ctxtProvider := dep.GetComponent[dep.ContextualProvider](dependent)
	serviceCtxt := NewLoopContext(scopeCtxt, ctxtProvider, interfaceType)
	dep.TrackDependent(serviceCtxt, dependent)
	return NewDefaultLooper(serviceCtxt)
}
