package hosting

import (
	"fmt"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

// utility API to create Func Processor
func createFuncProcessor(context dep.Context, procType types.DataType, processorFunc dep.GenericActionMethod) FunctionProcessor {
	provider := dep.GetComponent[dep.ContextualProvider](context)
	fpCtxt := dep.NewComponentContext(context.(dep.ContextEx).GetScopeContext(), provider, procType)
	processor := dep.GetComponent[FunctionProcessor](fpCtxt)
	processor.Initialize(fpCtxt, procType, func(ctxt dep.Context, interfaceType types.DataType, scopeCtxt ScopeContext) {
		processorFunc(ctxt, fmt.Sprintf("Processor(%s)", interfaceType.FullName()), dep.DepInst[ScopeContext](scopeCtxt))
	})
	return processor
}

type ProcessorMethod func(dep.Context, types.DataType, ScopeContext)

type FunctionProcessor interface {
	Initialize(fpCtxt dep.Context, dt types.DataType, proc ProcessorMethod) FunctionProcessor

	Run(ctxt ScopeContext)
}

type DefaultFuncProcessor struct {
	context  dep.Context
	fpCtxt   dep.Context
	procType types.DataType
	logger   logger.Logger
	procFunc ProcessorMethod
}

func NewFunctionProcessor(context dep.Context) *DefaultFuncProcessor {
	return &DefaultFuncProcessor{
		context: context,
		logger:  context.GetLogger(),
	}
}

func (fp *DefaultFuncProcessor) Initialize(fpCtxt dep.Context, procType types.DataType, procFunc ProcessorMethod) FunctionProcessor {
	fp.fpCtxt = fpCtxt
	fp.procType = procType
	fp.procFunc = procFunc

	return fp
}

func (fp *DefaultFuncProcessor) Run(scopeCtxt ScopeContext) {
	if fp.procFunc == nil {
		fp.logger.Fatalf("FunctionProcessor is not initialized before running")
	} else {
		fp.logger.Debugw("FunctionProcessor run start", "type", fp.procType.Name(), "looper", scopeCtxt.GetLoopRunContext().LooperName())
	}

	fp.procFunc(fp.fpCtxt, fp.procType, scopeCtxt)

	fp.logger.Debugw("FunctionProcessor run complete", "type", fp.procType.Name())
}
