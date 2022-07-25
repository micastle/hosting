package main

import (
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/dep"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/hosting"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/logger"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

type MyProc hosting.LoopProcessor

type MyProcessor struct {
	logger     logger.Logger
	comp       Component
}

func NewMyProcessor(context dep.Context) *MyProcessor {
	var p *MyProcessor
	p = &MyProcessor{
		logger:     context.GetLogger(),
		comp:       dep.GetComponent[Component](context),
	}
	p.logger.Debugw("Processor created", "type", types.Of(p).Name())
	return p
}

func (p *MyProcessor) Run(ctxt hosting.ScopeContext) {
	p.logger.Debugw("Processor run start", "looper", ctxt.GetLoopRunContext().LooperName(), "type", types.Of(p).Name())

	input := "VALUE"
	if ctxt.HasVariable("Input", true) {
		input = ctxt.GetVariable("Input").(string)
		p.logger.Debugw("input value found", "value", input, "looper", ctxt.GetLoopRunContext().LooperName())
	} else {
		p.logger.Warnw("input value not found, use default value as input", "looper", ctxt.GetLoopRunContext().LooperName())
	}

	p.comp.Print()

	time.Sleep(1 * time.Second)

	p.logger.Debugw("Processor run complete", "type", types.Of(p).Name())

	ctxt.SetVariable("Input", input)

	// stop current run of the loop
	//runContext.SetStopped()
}
